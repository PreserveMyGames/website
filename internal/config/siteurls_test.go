package config

import "testing"

func TestSiteHostAndSubdomains(t *testing.T) {
	cfg := Config{SiteURL: "https://preservemygames.org"}
	if cfg.SiteHost() != "preservemygames.org" {
		t.Fatalf("host: got %q", cfg.SiteHost())
	}
	if cfg.WikiURL() != "https://wiki.preservemygames.org" {
		t.Fatalf("wiki: got %q", cfg.WikiURL())
	}
	if cfg.ForumsURL() != "https://forums.preservemygames.org" {
		t.Fatalf("forums: got %q", cfg.ForumsURL())
	}
}

func TestNoticeKind(t *testing.T) {
	cfg := Config{SiteNotice: "construction"}
	if cfg.NoticeKind() != "construction" {
		t.Fatalf("kind: got %q", cfg.NoticeKind())
	}
	cfg.SiteNotice = "invalid"
	if cfg.NoticeKind() != "" {
		t.Fatal("expected empty kind for invalid value")
	}
}
