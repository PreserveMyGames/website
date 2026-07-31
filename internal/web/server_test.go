package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PreserveMyGames/website/internal/blog"
	"github.com/PreserveMyGames/website/internal/config"
	"github.com/PreserveMyGames/website/internal/constants"
	"github.com/PreserveMyGames/website/internal/i18n"
)

func testServer(tb testing.TB) *Server {
	tb.Helper()
	bundle, err := i18n.New()
	if err != nil {
		tb.Fatalf("i18n: %v", err)
	}
	index, err := blog.Load(false)
	if err != nil {
		tb.Fatalf("blog: %v", err)
	}
	cfg := config.Config{
		Port:         constants.DefaultPort,
		AppEnv:       "test",
		SiteURL:      "http://example.com",
		ContactEmail: constants.ContactEmail,
	}
	srv, err := New(cfg, bundle, index)
	if err != nil {
		tb.Fatalf("server: %v", err)
	}
	return srv
}

func TestRoutes(t *testing.T) {
	srv := testServer(t)
	handler := srv.Handler()

	cases := []struct {
		path       string
		wantStatus int
		contains   string
	}{
		{constants.PathHealthz, http.StatusOK, constants.HealthzResponse},
		{"/en/", http.StatusOK, constants.AppName},
		{"/de/", http.StatusOK, "Spiele verdienen"},
		{"/ru/", http.StatusOK, "Игры заслуживают"},
		{"/de/about", http.StatusOK, "Über uns"},
		{"/ru/about", http.StatusOK, "О проекте"},
		{"/en/about", http.StatusOK, "About"},
		{"/en/privacy", http.StatusOK, "Privacy"},
		{"/en/contact", http.StatusOK, constants.ContactEmail},
		{"/en/blog", http.StatusOK, "Welcome to Preserve My Games"},
		{"/en/blog/2026-07-01-welcome", http.StatusOK, "Welcome to Preserve My Games"},
		{"/en/blog?q=preservation", http.StatusOK, `value="preservation"`},
		{"/en/blog/rss.xml", http.StatusOK, "<rss"},
		{"/en/blog/atom.xml", http.StatusOK, "<feed"},
		{"/en/search?q=preservation", http.StatusOK, "preservation"},
		{"/en/blog/not-real", http.StatusNotFound, ""},
		{"/fr/", http.StatusNotFound, ""},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status: got %d want %d", rec.Code, tc.wantStatus)
			}
			if tc.contains != "" && !strings.Contains(rec.Body.String(), tc.contains) {
				t.Fatalf("body missing %q", tc.contains)
			}
		})
	}
}

func TestLocaleSwitcherRendered(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/en/about", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	body := rec.Body.String()
	if !strings.Contains(body, `hreflang="de"`) || !strings.Contains(body, `hreflang="ru"`) {
		t.Fatal("expected locale switcher links")
	}
	if !strings.Contains(body, `href="/de/about"`) {
		t.Fatal("expected german about link")
	}
}

func TestSearchIndexIncludesLocalizedPages(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/de/search-index.json", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	var entries []blog.SearchEntry
	if err := json.Unmarshal(rec.Body.Bytes(), &entries); err != nil {
		t.Fatalf("json: %v", err)
	}
	found := false
	for _, e := range entries {
		if e.URL == "/de/about" && e.Title == "Über uns" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected localized about page in german search index")
	}
}

func TestRootRedirectUsesAcceptLanguage(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("status: got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/en/" {
		t.Fatalf("location: got %q", loc)
	}
}

func TestSecurityHeaders(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, constants.PathHealthz, nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	for _, header := range []string{
		constants.HeaderCSP,
		constants.HeaderXContentType,
		constants.HeaderReferrerPolicy,
		constants.HeaderXFrameOptions,
		constants.HeaderPermissionsPolicy,
	} {
		if rec.Header().Get(header) == "" {
			t.Fatalf("missing header %s", header)
		}
	}
}

func TestMethodNotAllowed(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodPost, constants.PathHealthz, nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestBlogSlugTraversalBlocked(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/en/blog/..%2Fsecrets", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestStaticPathTraversalBlocked(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/static/%2e%2e/secret", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestSearchIndexJSONShape(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/en/search-index.json", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var entries []blog.SearchEntry
	if err := json.Unmarshal(rec.Body.Bytes(), &entries); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected search entries")
	}
	if entries[0].Title == "" || entries[0].URL == "" {
		t.Fatal("expected populated search entry fields")
	}
}

func TestRedirectTargetValidation(t *testing.T) {
	cases := []struct {
		in   string
		ok   bool
		want string
	}{
		{"/en/", true, "/en/"},
		{"//evil.test", false, ""},
		{"https://evil.test", false, ""},
		{"", false, ""},
	}
	for _, tc := range cases {
		got, ok := redirectTarget(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("redirectTarget(%q) = (%q, %v), want (%q, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func BenchmarkSearchIndex(b *testing.B) {
	srv := testServer(b)
	handler := srv.Handler()
	req := httptest.NewRequest(http.MethodGet, "/en/search-index.json", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkBlogRSS(b *testing.B) {
	srv := testServer(b)
	handler := srv.Handler()
	req := httptest.NewRequest(http.MethodGet, "/en/blog/rss.xml", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkHomepage(b *testing.B) {
	srv := testServer(b)
	handler := srv.Handler()
	req := httptest.NewRequest(http.MethodGet, "/en/", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
