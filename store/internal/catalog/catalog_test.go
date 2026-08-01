package catalog_test

import (
	"path/filepath"
	"sync"
	"testing"

	"github.com/PreserveMyGames/store/internal/catalog"
	"github.com/PreserveMyGames/store/internal/constants"
	"github.com/PreserveMyGames/store/internal/db"
)

func TestEmptyCatalog(t *testing.T) {
	dir := t.TempDir()
	sqlDB, err := db.Open(dir)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer sqlDB.Close()

	store := catalog.New(sqlDB)
	if err := store.SyncCatalog(); err != nil {
		t.Fatalf("sync: %v", err)
	}
	products, err := store.ListProducts()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(products) != 0 {
		t.Fatalf("expected empty catalog, got %d", len(products))
	}
}

func TestStockReservationRace(t *testing.T) {
	dir := t.TempDir()
	sqlDB, err := db.Open(filepath.Join(dir))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer sqlDB.Close()

	store := catalog.New(sqlDB)
	stock := 1
	_, err = sqlDB.Exec(`
		INSERT INTO products (slug, type, title, description, price_cents, currency, stock, category, tags, sku, active)
		VALUES ('limited', 'physical', 'Limited', '', 1000, 'USD', 1, '', '[]', '', 1)`)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	p := catalog.Product{
		Slug:       "limited",
		Type:       constants.ProductPhysical,
		Title:      "Limited",
		PriceCents: 1000,
		Currency:   "USD",
		Stock:      &stock,
	}

	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.CreateOrder("en", constants.PaymentStripe, p, 1)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	ok := 0
	for err := range errs {
		if err == nil {
			ok++
		}
	}
	if ok != 1 {
		t.Fatalf("expected exactly 1 successful reservation, got %d", ok)
	}
	avail, err := store.Available("limited")
	if err != nil {
		t.Fatalf("available: %v", err)
	}
	if avail != 0 {
		t.Fatalf("expected 0 available, got %d", avail)
	}
}
