package web

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestConcurrentHandlerRequests(t *testing.T) {
	srv := testServer(t)
	handler := srv.Handler()

	paths := []string{
		"/en/",
		"/en/about",
		"/en/blog",
		"/en/blog/2026-07-01-welcome",
		"/en/search-index.json",
		"/healthz",
		"/en/no-such-page",
	}

	var wg sync.WaitGroup
	errCh := make(chan string, 64)

	for i := range 80 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			path := paths[i%len(paths)]
			req := httptest.NewRequest(http.MethodGet, path, nil)
			if path == "/en/search-index.json" {
				req.Header.Set("Accept", "application/json")
			} else {
				req.Header.Set("Accept", "text/html")
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code >= 500 {
				errCh <- path
			}
		}(i)
	}
	wg.Wait()
	close(errCh)

	for path := range errCh {
		t.Errorf("path %s returned 5xx under concurrency", path)
	}
}

func FuzzRedirectTarget(f *testing.F) {
	f.Add("/en/")
	f.Add("//evil.test")
	f.Add("https://evil.test")
	f.Add("")
	f.Fuzz(func(t *testing.T, path string) {
		got, ok := redirectTarget(path)
		if ok {
			if !stringsHasPrefix(got, "/") || stringsHasPrefix(got, "//") {
				t.Fatalf("redirectTarget(%q) unsafe target %q", path, got)
			}
		}
	})
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
