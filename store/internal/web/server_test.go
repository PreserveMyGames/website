package web_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PreserveMyGames/store/internal/catalog"
	"github.com/PreserveMyGames/store/internal/config"
	"github.com/PreserveMyGames/store/internal/db"
	"github.com/PreserveMyGames/store/internal/fx"
	"github.com/PreserveMyGames/store/internal/i18n"
	"github.com/PreserveMyGames/store/internal/web"
)

func testServer(t *testing.T) *web.Server {
	t.Helper()
	bundle, err := i18n.New()
	if err != nil {
		t.Fatalf("i18n: %v", err)
	}
	sqlDB, err := db.Open(t.TempDir())
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	store := catalog.New(sqlDB)
	if err := store.SyncCatalog(); err != nil {
		t.Fatalf("sync: %v", err)
	}
	cfg := config.Config{
		Port:         "0",
		AppEnv:       "development",
		SiteURL:      "http://example.com",
		DataDir:      t.TempDir(),
		BaseCurrency: "USD",
		MainSiteURL:  "https://preservemygames.org",
	}
	srv, err := web.New(cfg, bundle, store, fx.New(), nil)
	if err != nil {
		t.Fatalf("server: %v", err)
	}
	return srv
}

func TestHealthz(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Body.String() != "ok" {
		t.Fatalf("unexpected healthz response: %d %q", rec.Code, rec.Body.String())
	}
}

func TestEmptyCatalogPage(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/en/", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "empty-state") {
		t.Fatal("expected empty-state markup")
	}
	if !strings.Contains(body, "Nothing for sale yet") {
		t.Fatal("expected empty catalog title")
	}
}
