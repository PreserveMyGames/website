package monero_test

import (
	"path/filepath"
	"testing"

	"github.com/PreserveMyGames/store/internal/catalog"
	"github.com/PreserveMyGames/store/internal/constants"
	"github.com/PreserveMyGames/store/internal/db"
	"github.com/PreserveMyGames/store/internal/payments/monero"
)

type fakeWallet struct {
	transfers []monero.Transfer
}

func (f *fakeWallet) CreateAddress(label string) (string, int, error) {
	return "addr", 7, nil
}

func (f *fakeWallet) GetTransfers(accountIndex int) ([]monero.Transfer, error) {
	return f.transfers, nil
}

func TestPollerMarksPaid(t *testing.T) {
	sqlDB, err := db.Open(filepath.Join(t.TempDir()))
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer sqlDB.Close()

	store := catalog.New(sqlDB)
	stock := 5
	_, err = sqlDB.Exec(`
		INSERT INTO products (slug, type, title, description, price_cents, currency, stock, category, tags, sku, active)
		VALUES ('item', 'digital', 'Item', '', 1000, 'USD', 5, '', '[]', '', 1)`)
	if err != nil {
		t.Fatalf("insert product: %v", err)
	}
	p := catalog.Product{
		Slug: "item", Type: constants.ProductDigital, Title: "Item",
		PriceCents: 1000, Currency: "USD", Stock: &stock, DigitalPayload: "https://example.com/file",
	}
	order, err := store.CreateOrder("en", constants.PaymentMonero, p, 1)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if err := store.SetMoneroInvoice(order.ID, "addr", 7, 1_000_000_000_000); err != nil {
		t.Fatalf("invoice: %v", err)
	}

	wallet := &fakeWallet{transfers: []monero.Transfer{{
		Amount: 1_000_000_000_000, Confirmations: 10, TxID: "abc", SubaddrIndex: 7, Unlocked: true,
	}}}
	poller := monero.NewPoller(store, wallet, 10)
	if err := poller.Tick(); err != nil {
		t.Fatalf("tick: %v", err)
	}
	got, err := store.GetOrder(order.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != constants.OrderPaid {
		t.Fatalf("expected paid, got %s", got.Status)
	}
}

func TestFormatXMR(t *testing.T) {
	if monero.FormatXMR(1_500_000_000_000) != "1.500000000000" {
		t.Fatalf("unexpected format: %s", monero.FormatXMR(1_500_000_000_000))
	}
}
