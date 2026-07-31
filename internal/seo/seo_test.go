package seo

import (
	"strings"
	"testing"
	"time"
)

func TestHomeJSONLD(t *testing.T) {
	meta := Home("https://example.com", "en", "Title", "Description", nil)
	if meta.JSONLDScript == "" {
		t.Fatal("expected json-ld script")
	}
	if !strings.Contains(meta.JSONLDScript, "WebSite") {
		t.Fatal("expected WebSite schema")
	}
	if meta.Canonical != "https://example.com/en/" {
		t.Fatalf("canonical: got %q", meta.Canonical)
	}
}

func TestBlogPostMeta(t *testing.T) {
	when := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	meta := BlogPost("https://example.com", "en", "welcome", "Welcome", "Desc", when, nil)
	if meta.OGType != "article" {
		t.Fatalf("og type: got %q", meta.OGType)
	}
	if !strings.Contains(meta.JSONLDScript, "BlogPosting") {
		t.Fatal("expected BlogPosting schema")
	}
}

func TestSitemapURLs(t *testing.T) {
	urls := SitemapURLs("https://example.com", []string{"en"}, map[string][]string{
		"en": {"welcome"},
	})
	if len(urls) < 7 {
		t.Fatalf("expected core routes plus post, got %d urls", len(urls))
	}
	found := false
	for _, u := range urls {
		if u == "https://example.com/en/blog/welcome" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected blog post in sitemap")
	}
}
