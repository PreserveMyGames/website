package validate

import "testing"

func TestSlug(t *testing.T) {
	tests := map[string]bool{
		"welcome":            true,
		"2026-07-01-welcome": true,
		"2026-07-01-post":    true,
		"":                   false,
		"../secrets":         false,
		"a/b":                false,
		`a\b`:                false,
		"feed.json":          false,
		"UPPER":              false,
		"-bad":               false,
		"bad-":               false,
	}

	for slug, want := range tests {
		if got := Slug(slug); got != want {
			t.Fatalf("Slug(%q) = %v, want %v", slug, got, want)
		}
	}
}

func TestStaticPath(t *testing.T) {
	if !StaticPath("css/site.css") {
		t.Fatal("expected valid static path")
	}
	if StaticPath("../secret") {
		t.Fatal("expected traversal path to be rejected")
	}
	if StaticPath("") {
		t.Fatal("expected empty path to be rejected")
	}
}
