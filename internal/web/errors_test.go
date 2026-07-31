package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNotFoundPageHTML(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/en/blog/not-real", nil)
	req.Header.Set("Accept", "text/html")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want 404", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Page not found") {
		t.Fatal("expected localized 404 title in HTML body")
	}
	if !strings.Contains(body, "<html") {
		t.Fatal("expected full HTML error page")
	}
}

func TestNotFoundUnknownLocaleUsesPreferredLanguage(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/fr/missing", nil)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want 404", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Seite nicht gefunden") {
		t.Fatal("expected german 404 page for unsupported locale path")
	}
}

func TestUnknownLocalePathNotFound(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/en/no-such-page", nil)
	req.Header.Set("Accept", "text/html")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want 404", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "Games deserve a lasting home") {
		t.Fatal("unknown locale path should not render the home page")
	}
}

func TestNotFoundJSONStaysPlain(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/xyzabc", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want 404", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "<html") {
		t.Fatal("expected plain 404 for JSON accept header")
	}
}

func TestRecoverMiddleware(t *testing.T) {
	srv := testServer(t)
	inner := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("test panic")
	})
	handler := recoverMiddleware(srv, inner)

	req := httptest.NewRequest(http.MethodGet, "/en/", nil)
	req.Header.Set("Accept", "text/html")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Something went wrong") {
		t.Fatal("expected localized 500 page")
	}
}

func TestMethodNotAllowedHTML(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodPost, "/en/about", nil)
	req.Header.Set("Accept", "text/html")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status: got %d want 405", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Method not allowed") {
		t.Fatal("expected localized 405 page")
	}
}
