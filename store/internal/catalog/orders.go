package catalog

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/PreserveMyGames/store/internal/constants"
)

type Order struct {
	ID                  string
	Status              string
	PaymentMethod       string
	Lang                string
	Currency            string
	TotalCents          int64
	Email               string
	ShippingName        string
	ShippingAddress     string
	StripeSessionID     string
	MoneroAddress       string
	MoneroAddressIndex  int
	MoneroAmountAtomic  int64
	MoneroTxHash        string
	MoneroConfirmations int
	ExpiresAt           *time.Time
	PaidAt              *time.Time
	ShippedAt           *time.Time
	CreatedAt           time.Time
	Items               []OrderItem
}

type OrderItem struct {
	ProductSlug    string
	Title          string
	Qty            int
	UnitPriceCents int64
	ProductType    string
	DigitalPayload string
}

func newOrderID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func (s *Store) CreateOrder(lang, paymentMethod string, product Product, qty int) (Order, error) {
	if qty < 1 {
		qty = 1
	}
	id, err := newOrderID()
	if err != nil {
		return Order{}, err
	}
	expires := time.Now().UTC().Add(constants.ReservationTTL)
	status := constants.OrderPending
	switch paymentMethod {
	case constants.PaymentStripe:
		status = constants.OrderAwaitingStripe
	case constants.PaymentMonero:
		status = constants.OrderAwaitingMonero
	}

	tx, err := s.db.Begin()
	if err != nil {
		return Order{}, err
	}
	defer func() { _ = tx.Rollback() }()

	avail, err := availableTx(tx, product.Slug)
	if err != nil {
		return Order{}, err
	}
	if product.Stock != nil && avail < qty {
		return Order{}, fmt.Errorf("insufficient stock")
	}

	total := product.PriceCents * int64(qty)
	_, err = tx.Exec(`
		INSERT INTO orders (
			id, status, payment_method, lang, currency, total_cents, expires_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))
	`, id, status, paymentMethod, lang, product.Currency, total, expires.Format(time.RFC3339))
	if err != nil {
		return Order{}, err
	}
	_, err = tx.Exec(`
		INSERT INTO order_items (order_id, product_slug, title, qty, unit_price_cents, product_type, digital_payload)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, product.Slug, product.Title, qty, product.PriceCents, product.Type, product.DigitalPayload)
	if err != nil {
		return Order{}, err
	}
	if product.Stock != nil {
		_, err = tx.Exec(`
			INSERT INTO stock_reservations (order_id, product_slug, qty, expires_at)
			VALUES (?, ?, ?, ?)
		`, id, product.Slug, qty, expires.Format(time.RFC3339))
		if err != nil {
			return Order{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return Order{}, err
	}
	return s.GetOrder(id)
}

func availableTx(tx *sql.Tx, slug string) (int, error) {
	var stock sql.NullInt64
	err := tx.QueryRow(`SELECT stock FROM products WHERE slug = ?`, slug).Scan(&stock)
	if err != nil {
		return 0, err
	}
	if !stock.Valid {
		return 999999, nil
	}
	var reserved int
	err = tx.QueryRow(`
		SELECT COALESCE(SUM(qty), 0) FROM stock_reservations
		WHERE product_slug = ? AND released = 0 AND expires_at > datetime('now')
	`, slug).Scan(&reserved)
	if err != nil {
		return 0, err
	}
	avail := int(stock.Int64) - reserved
	if avail < 0 {
		return 0, nil
	}
	return avail, nil
}

func (s *Store) GetOrder(id string) (Order, error) {
	row := s.db.QueryRow(`
		SELECT id, status, payment_method, lang, currency, total_cents, email,
			shipping_name, shipping_address, stripe_session_id, monero_address,
			monero_address_index, monero_amount_atomic, monero_tx_hash, monero_confirmations,
			expires_at, paid_at, shipped_at, created_at
		FROM orders WHERE id = ?`, id)
	var o Order
	var expires, paid, shipped sql.NullString
	var created string
	err := row.Scan(
		&o.ID, &o.Status, &o.PaymentMethod, &o.Lang, &o.Currency, &o.TotalCents, &o.Email,
		&o.ShippingName, &o.ShippingAddress, &o.StripeSessionID, &o.MoneroAddress,
		&o.MoneroAddressIndex, &o.MoneroAmountAtomic, &o.MoneroTxHash, &o.MoneroConfirmations,
		&expires, &paid, &shipped, &created,
	)
	if err != nil {
		return Order{}, err
	}
	o.ExpiresAt = parseNullTime(expires)
	o.PaidAt = parseNullTime(paid)
	o.ShippedAt = parseNullTime(shipped)
	if t, err := time.Parse(time.RFC3339, created); err == nil {
		o.CreatedAt = t
	} else if t, err := time.Parse("2006-01-02 15:04:05", created); err == nil {
		o.CreatedAt = t
	}
	items, err := s.orderItems(id)
	if err != nil {
		return Order{}, err
	}
	o.Items = items
	return o, nil
}

func (s *Store) orderItems(orderID string) ([]OrderItem, error) {
	rows, err := s.db.Query(`
		SELECT product_slug, title, qty, unit_price_cents, product_type, digital_payload
		FROM order_items WHERE order_id = ?`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []OrderItem
	for rows.Next() {
		var it OrderItem
		if err := rows.Scan(&it.ProductSlug, &it.Title, &it.Qty, &it.UnitPriceCents, &it.ProductType, &it.DigitalPayload); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (s *Store) ListOrders() ([]Order, error) {
	rows, err := s.db.Query(`SELECT id FROM orders ORDER BY created_at DESC LIMIT 200`)
	if err != nil {
		return nil, err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	out := make([]Order, 0, len(ids))
	for _, id := range ids {
		o, err := s.GetOrder(id)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, nil
}

func (s *Store) ListAwaitingMonero() ([]Order, error) {
	rows, err := s.db.Query(`SELECT id FROM orders WHERE status = ?`, constants.OrderAwaitingMonero)
	if err != nil {
		return nil, err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	out := make([]Order, 0, len(ids))
	for _, id := range ids {
		o, err := s.GetOrder(id)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, nil
}

func (s *Store) SetStripeSession(orderID, sessionID string) error {
	_, err := s.db.Exec(`UPDATE orders SET stripe_session_id = ?, updated_at = datetime('now') WHERE id = ?`, sessionID, orderID)
	return err
}

func (s *Store) SetMoneroInvoice(orderID, address string, addressIndex int, amountAtomic int64) error {
	_, err := s.db.Exec(`
		UPDATE orders SET monero_address = ?, monero_address_index = ?, monero_amount_atomic = ?, updated_at = datetime('now')
		WHERE id = ?`, address, addressIndex, amountAtomic, orderID)
	return err
}

func (s *Store) UpdateMoneroProgress(orderID, txHash string, confirmations int) error {
	_, err := s.db.Exec(`
		UPDATE orders SET monero_tx_hash = ?, monero_confirmations = ?, updated_at = datetime('now')
		WHERE id = ?`, txHash, confirmations, orderID)
	return err
}

func (s *Store) MarkOrderPaid(orderID, provider, providerRef string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var status, productType string
	err = tx.QueryRow(`SELECT status FROM orders WHERE id = ?`, orderID).Scan(&status)
	if err != nil {
		return err
	}
	if status == constants.OrderPaid || status == constants.OrderAwaitingShip || status == constants.OrderShipped {
		return nil
	}

	rows, err := tx.Query(`SELECT product_slug, qty, product_type FROM order_items WHERE order_id = ?`, orderID)
	if err != nil {
		return err
	}
	type line struct {
		slug  string
		qty   int
		ptype string
	}
	var lines []line
	for rows.Next() {
		var l line
		if err := rows.Scan(&l.slug, &l.qty, &l.ptype); err != nil {
			rows.Close()
			return err
		}
		lines = append(lines, l)
		productType = l.ptype
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, l := range lines {
		_, err = tx.Exec(`
			UPDATE products SET stock = stock - ?
			WHERE slug = ? AND stock IS NOT NULL`, l.qty, l.slug)
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(`UPDATE stock_reservations SET released = 1 WHERE order_id = ?`, orderID)
	if err != nil {
		return err
	}

	newStatus := constants.OrderPaid
	if productType == constants.ProductPhysical {
		newStatus = constants.OrderAwaitingShip
	}
	_, err = tx.Exec(`
		UPDATE orders SET status = ?, paid_at = datetime('now'), updated_at = datetime('now') WHERE id = ?`,
		newStatus, orderID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO payments (order_id, provider, provider_ref, status)
		VALUES (?, ?, ?, 'paid')`, orderID, provider, providerRef)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) MarkShipped(orderID string) error {
	_, err := s.db.Exec(`
		UPDATE orders SET status = ?, shipped_at = datetime('now'), updated_at = datetime('now')
		WHERE id = ?`, constants.OrderShipped, orderID)
	return err
}

func parseNullTime(v sql.NullString) *time.Time {
	if !v.Valid || v.String == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, v.String); err == nil {
		return &t
	}
	if t, err := time.Parse("2006-01-02 15:04:05", v.String); err == nil {
		return &t
	}
	return nil
}
