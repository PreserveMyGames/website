package i18n

import (
	"testing"
)

func TestBundleTranslations(t *testing.T) {
	bundle, err := New()
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	got := bundle.T("en", "site.name")
	if got != "Preserve My Games" {
		t.Fatalf("site.name: got %q", got)
	}
}

func TestMatchAcceptLanguage(t *testing.T) {
	bundle, err := New()
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	if got := bundle.Match(""); got != "en" {
		t.Fatalf("empty accept: got %q", got)
	}
	if got := bundle.Match("en-US,en;q=0.9"); got != "en" {
		t.Fatalf("en accept: got %q", got)
	}
	if got := bundle.Match("fr-FR,fr;q=0.9"); got != "en" {
		t.Fatalf("unsupported accept should fallback to en, got %q", got)
	}
}

func TestAllowedLocales(t *testing.T) {
	bundle, err := New()
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	if !bundle.Allowed("en") {
		t.Fatal("en should be allowed")
	}
	if bundle.Allowed("fr") {
		t.Fatal("fr should not be allowed")
	}
	if !bundle.Allowed("EN") {
		t.Fatal("locale codes should be normalized to lowercase")
	}
}

func TestPageURL(t *testing.T) {
	bundle, err := New()
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	if got := bundle.PageURL("https://example.com", "en", ""); got != "https://example.com/en/" {
		t.Fatalf("home url: got %q", got)
	}
	if got := bundle.PageURL("https://example.com/", "en", "blog"); got != "https://example.com/en/blog" {
		t.Fatalf("blog url: got %q", got)
	}
}
