package catalog

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/PreserveMyGames/store/internal/constants"
)

//go:embed content
var contentFS embed.FS

type ProductFile struct {
	Slug                string   `yaml:"slug"`
	Type                string   `yaml:"type"`
	Title               string   `yaml:"title"`
	Description         string   `yaml:"description"`
	PriceCents          int64    `yaml:"price_cents"`
	Currency            string   `yaml:"currency"`
	Stock               *int     `yaml:"stock"`
	Category            string   `yaml:"category"`
	Tags                []string `yaml:"tags"`
	SKU                 string   `yaml:"sku"`
	DigitalDeliveryKind string   `yaml:"digital_delivery_kind"`
	DigitalPayload      string   `yaml:"digital_payload"`
	ShippingClass       string   `yaml:"shipping_class"`
	ImagePath           string   `yaml:"image_path"`
	Active              *bool    `yaml:"active"`
}

type Product struct {
	Slug                string
	Type                string
	Title               string
	Description         string
	PriceCents          int64
	Currency            string
	Stock               *int
	Available           int
	Category            string
	Tags                []string
	SKU                 string
	DigitalDeliveryKind string
	DigitalPayload      string
	ShippingClass       string
	ImagePath           string
	Active              bool
}

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) SyncCatalog() error {
	entries, err := fs.ReadDir(contentFS, "content")
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		data, err := fs.ReadFile(contentFS, "content/"+e.Name())
		if err != nil {
			return err
		}
		var p ProductFile
		if err := yaml.Unmarshal(data, &p); err != nil {
			return fmt.Errorf("parse %s: %w", e.Name(), err)
		}
		if p.Slug == "" || p.Title == "" || p.PriceCents < 0 {
			return fmt.Errorf("invalid product file %s", e.Name())
		}
		if p.Type != constants.ProductDigital && p.Type != constants.ProductPhysical {
			return fmt.Errorf("invalid product type in %s", e.Name())
		}
		if p.Currency == "" {
			p.Currency = constants.DefaultBaseCurrency
		}
		p.Category = normalizeCategory(p.Category)
		tagsJSON := encodeTags(p.Tags)
		active := 1
		if p.Active != nil && !*p.Active {
			active = 0
		}
		var stock any
		if p.Stock != nil {
			stock = *p.Stock
		}
		_, err = s.db.Exec(`
			INSERT INTO products (
				slug, type, title, description, price_cents, currency, stock,
				category, tags, sku,
				digital_delivery_kind, digital_payload, shipping_class, image_path, active, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
			ON CONFLICT(slug) DO UPDATE SET
				type = excluded.type,
				title = excluded.title,
				description = excluded.description,
				price_cents = excluded.price_cents,
				currency = excluded.currency,
				category = excluded.category,
				tags = excluded.tags,
				sku = excluded.sku,
				digital_delivery_kind = excluded.digital_delivery_kind,
				digital_payload = excluded.digital_payload,
				shipping_class = excluded.shipping_class,
				image_path = excluded.image_path,
				active = excluded.active,
				updated_at = datetime('now')
		`, p.Slug, p.Type, p.Title, p.Description, p.PriceCents, strings.ToUpper(p.Currency), stock,
			p.Category, tagsJSON, strings.TrimSpace(p.SKU),
			p.DigitalDeliveryKind, p.DigitalPayload, p.ShippingClass, p.ImagePath, active)
		if err != nil {
			return err
		}
		if p.Stock != nil {
			_, err = s.db.Exec(`UPDATE products SET stock = ? WHERE slug = ? AND stock IS NULL`, *p.Stock, p.Slug)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) ListProducts() ([]Product, error) {
	return s.queryProducts(`
		SELECT slug, type, title, description, price_cents, currency, stock,
			category, tags, sku,
			digital_delivery_kind, digital_payload, shipping_class, image_path, active
		FROM products WHERE active = 1 ORDER BY title ASC`)
}

func (s *Store) ListByCategory(category string) ([]Product, error) {
	category = normalizeCategory(category)
	return s.queryProducts(`
		SELECT slug, type, title, description, price_cents, currency, stock,
			category, tags, sku,
			digital_delivery_kind, digital_payload, shipping_class, image_path, active
		FROM products WHERE active = 1 AND category = ? ORDER BY title ASC`, category)
}

func (s *Store) ListByTag(tag string) ([]Product, error) {
	all, err := s.ListProducts()
	if err != nil {
		return nil, err
	}
	var out []Product
	for _, p := range all {
		if productMatchesTag(p, tag) {
			out = append(out, p)
		}
	}
	return out, nil
}

func (s *Store) SearchProducts(query string) ([]Product, error) {
	all, err := s.ListProducts()
	if err != nil {
		return nil, err
	}
	var out []Product
	for _, p := range all {
		if productMatchesQuery(p, query) {
			out = append(out, p)
		}
	}
	return out, nil
}

func (s *Store) queryProducts(query string, args ...any) ([]Product, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	var out []Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	for i := range out {
		avail, err := s.Available(out[i].Slug)
		if err != nil {
			return nil, err
		}
		out[i].Available = avail
	}
	return out, nil
}

func (s *Store) GetProduct(slug string) (Product, bool, error) {
	row := s.db.QueryRow(`
		SELECT slug, type, title, description, price_cents, currency, stock,
			category, tags, sku,
			digital_delivery_kind, digital_payload, shipping_class, image_path, active
		FROM products WHERE slug = ? AND active = 1`, slug)
	p, err := scanProduct(row)
	if err == sql.ErrNoRows {
		return Product{}, false, nil
	}
	if err != nil {
		return Product{}, false, err
	}
	avail, err := s.Available(slug)
	if err != nil {
		return Product{}, false, err
	}
	p.Available = avail
	return p, true, nil
}

func (s *Store) Available(slug string) (int, error) {
	var stock sql.NullInt64
	err := s.db.QueryRow(`SELECT stock FROM products WHERE slug = ?`, slug).Scan(&stock)
	if err != nil {
		return 0, err
	}
	if !stock.Valid {
		return 999999, nil
	}
	var reserved int
	err = s.db.QueryRow(`
		SELECT COALESCE(SUM(qty), 0) FROM stock_reservations
		WHERE product_slug = ? AND released = 0 AND expires_at > datetime('now')
	`, slug).Scan(&reserved)
	if err != nil {
		return 0, err
	}
	avail := int(stock.Int64) - reserved
	if avail < 0 {
		avail = 0
	}
	return avail, nil
}

func (s *Store) SetStock(slug string, stock *int) error {
	var v any
	if stock != nil {
		v = *stock
	}
	_, err := s.db.Exec(`UPDATE products SET stock = ?, updated_at = datetime('now') WHERE slug = ?`, v, slug)
	return err
}

func (s *Store) ReleaseExpiredReservations() error {
	_, err := s.db.Exec(`
		UPDATE stock_reservations SET released = 1
		WHERE released = 0 AND expires_at <= datetime('now')`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
		UPDATE orders SET status = ?, updated_at = datetime('now')
		WHERE status IN (?, ?, ?) AND expires_at IS NOT NULL AND expires_at <= datetime('now')
	`, constants.OrderExpired, constants.OrderPending, constants.OrderAwaitingStripe, constants.OrderAwaitingMonero)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProduct(row scanner) (Product, error) {
	var p Product
	var stock sql.NullInt64
	var active int
	var tagsRaw string
	err := row.Scan(
		&p.Slug, &p.Type, &p.Title, &p.Description, &p.PriceCents, &p.Currency, &stock,
		&p.Category, &tagsRaw, &p.SKU,
		&p.DigitalDeliveryKind, &p.DigitalPayload, &p.ShippingClass, &p.ImagePath, &active,
	)
	if err != nil {
		return Product{}, err
	}
	if stock.Valid {
		v := int(stock.Int64)
		p.Stock = &v
	}
	p.Tags = decodeTags(tagsRaw)
	p.Active = active == 1
	return p, nil
}

func FormatMoney(cents int64, currency string) string {
	return fmt.Sprintf("%s %.2f", strings.ToUpper(currency), float64(cents)/100)
}

func NowUTC() time.Time {
	return time.Now().UTC()
}
