package validate

import (
	"strings"
	"testing"
)

func slugOracle(slug string) bool {
	if slug == "" || len(slug) > 128 {
		return false
	}
	if strings.Contains(slug, "..") || strings.Contains(slug, "/") || strings.Contains(slug, "\\") {
		return false
	}
	if strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
		return false
	}
	for _, r := range slug {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return false
		}
	}
	return true
}

func TestSlugOracleAgreement(t *testing.T) {
	cases := []string{
		"",
		"welcome",
		"2026-07-01-welcome",
		"../secrets",
		"a/b",
		"post.json",
		"UPPER",
		"-leading",
		"trailing-",
		"a--b",
		strings.Repeat("a", 129),
	}
	for _, slug := range cases {
		if Slug(slug) != slugOracle(slug) {
			t.Fatalf("oracle mismatch for %q: Slug=%v oracle=%v", slug, Slug(slug), slugOracle(slug))
		}
	}
}

func TestSlugRejectsExtensionLikeNames(t *testing.T) {
	if Slug("feed.json") {
		t.Fatal("slug must not look like a file extension")
	}
}

func TestStaticPathRejectsControlsAndAbsolute(t *testing.T) {
	cases := []string{
		"/css/site.css",
		"css\x00/site.css",
		"css\\site.css",
		"css/\n/site.css",
	}
	for _, p := range cases {
		if StaticPath(p) {
			t.Fatalf("StaticPath(%q) should be rejected", p)
		}
	}
}

func FuzzSlug(f *testing.F) {
	f.Add("welcome")
	f.Add("../x")
	f.Add("bad.json")
	f.Fuzz(func(t *testing.T, slug string) {
		got := Slug(slug)
		want := slugOracle(slug)
		if got != want {
			t.Fatalf("Slug(%q)=%v oracle=%v", slug, got, want)
		}
	})
}

func FuzzStaticPath(f *testing.F) {
	f.Add("css/site.css")
	f.Add("../secret")
	f.Add("")
	f.Fuzz(func(t *testing.T, p string) {
		ok := StaticPath(p)
		if ok && (p == "" || strings.Contains(p, "..") || strings.HasPrefix(p, "/") || strings.Contains(p, "\\") || strings.ContainsRune(p, '\x00')) {
			t.Fatalf("StaticPath(%q) must be false", p)
		}
	})
}
