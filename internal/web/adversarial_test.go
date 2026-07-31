package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdversarialPaths(t *testing.T) {
	srv := testServer(t)
	handler := srv.Handler()

	cases := []struct {
		path       string
		wantStatus int
	}{
		{"/en/blog/%2e%2e%2fetc%2fpasswd", 404},
		{"/en/blog/foo.json", 404},
		{"/en/blog/-invalid-", 404},
		{"/static/%00.css", 404},
		{"/en/blog/2026-07-01-welcome%00", 404},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set("Accept", "text/html")
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status: got %d want %d body=%q", rec.Code, tc.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestBlogPostShowsAuthor(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/en/blog/2026-07-01-welcome", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	body := rec.Body.String()
	if !strings.Contains(body, "Ivan") {
		t.Fatal("expected author name on blog post page")
	}
	if !strings.Contains(body, `rel="author"`) {
		t.Fatal("expected author link with rel=author")
	}
}
