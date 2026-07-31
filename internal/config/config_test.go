package config

import (
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("SITE_URL", "")
	t.Setenv("CONTACT_EMAIL", "")
	t.Setenv("ACCESS_LOG", "")
	t.Setenv("SITE_NOTICE", "")
	t.Setenv("SITE_NOTICE_MESSAGE", "")

	cfg := Load()
	if cfg.Port != "8080" {
		t.Fatalf("port: got %q", cfg.Port)
	}
	if cfg.AppEnv != "development" {
		t.Fatalf("app env: got %q", cfg.AppEnv)
	}
	if cfg.SiteURL != "http://localhost:8080" {
		t.Fatalf("site url: got %q", cfg.SiteURL)
	}
	if cfg.Production() {
		t.Fatal("expected non-production default")
	}
}

func TestLoadProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	cfg := Load()
	if !cfg.Production() {
		t.Fatal("expected production")
	}
}

func TestLoadAccessLogBool(t *testing.T) {
	t.Setenv("ACCESS_LOG", "true")
	cfg := Load()
	if !cfg.AccessLog {
		t.Fatal("expected access log enabled")
	}

	t.Setenv("ACCESS_LOG", "not-a-bool")
	cfg = Load()
	if cfg.AccessLog {
		t.Fatal("expected invalid bool to fall back to false")
	}
}

func TestListenPortUint16(t *testing.T) {
	cfg := Config{Port: "8080"}
	port, err := cfg.ListenPortUint16()
	if err != nil || port != 8080 {
		t.Fatalf("port: got %d err %v", port, err)
	}

	cfg.Port = "bad"
	if _, err := cfg.ListenPortUint16(); err == nil {
		t.Fatal("expected error for invalid port")
	}
}
