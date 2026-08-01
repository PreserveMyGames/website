package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/PreserveMyGames/website/internal/blog"
	"github.com/PreserveMyGames/website/internal/config"
	"github.com/PreserveMyGames/website/internal/constants"
	"github.com/PreserveMyGames/website/internal/i18n"
)

func TestNavExternalLinks(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/en/", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	body := rec.Body.String()
	if !strings.Contains(body, `href="https://wiki.example.com"`) {
		t.Fatal("expected wiki link in navbar")
	}
	if !strings.Contains(body, `href="https://forums.example.com"`) {
		t.Fatal("expected forums link in navbar")
	}
	if !strings.Contains(body, `href="https://store.example.com"`) {
		t.Fatal("expected store link in navbar")
	}
	if !strings.Contains(body, ">Preserve My Games</span>") {
		t.Fatal("expected brand title Preserve My Games")
	}

	start := strings.Index(body, `<nav class="site-nav"`)
	if start < 0 {
		t.Fatal("missing primary nav")
	}
	end := strings.Index(body[start:], "</nav>")
	if end < 0 {
		t.Fatal("missing primary nav end")
	}
	nav := body[start : start+end]
	if !strings.Contains(nav, "/donate") {
		t.Fatal("expected donate link in primary navigation")
	}
	if strings.Contains(nav, "/privacy") {
		t.Fatal("privacy should not appear in primary navigation")
	}
}

func TestContactReticulumBlock(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/en/contact", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	body := rec.Body.String()
	if !strings.Contains(body, "Reticulum Network") {
		t.Fatal("expected reticulum section")
	}
	if !strings.Contains(body, constants.ContactLXMF) {
		t.Fatal("expected LXMF address on contact page")
	}
}

func TestSiteNoticeBanner(t *testing.T) {
	os.Setenv(constants.EnvSiteNotice, constants.NoticeConstruction)
	t.Cleanup(func() {
		os.Unsetenv(constants.EnvSiteNotice)
		os.Unsetenv(constants.EnvSiteNoticeMessage)
	})

	bundle, err := i18n.New()
	if err != nil {
		t.Fatalf("i18n: %v", err)
	}
	index, err := blog.Load(false)
	if err != nil {
		t.Fatalf("blog: %v", err)
	}
	cfg := config.Load()
	srv, err := New(cfg, bundle, index)
	if err != nil {
		t.Fatalf("server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/en/", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	body := rec.Body.String()
	if !strings.Contains(body, `site-notice-construction`) {
		t.Fatal("expected construction notice banner")
	}
	if !strings.Contains(body, "under construction") {
		t.Fatal("expected construction notice message")
	}
}

func TestSiteNoticeCustomMessage(t *testing.T) {
	os.Setenv(constants.EnvSiteNotice, constants.NoticeMaintenance)
	os.Setenv(constants.EnvSiteNoticeMessage, "Custom maintenance text.")
	t.Cleanup(func() {
		os.Unsetenv(constants.EnvSiteNotice)
		os.Unsetenv(constants.EnvSiteNoticeMessage)
	})

	bundle, err := i18n.New()
	if err != nil {
		t.Fatalf("i18n: %v", err)
	}
	index, err := blog.Load(false)
	if err != nil {
		t.Fatalf("blog: %v", err)
	}
	cfg := config.Load()
	srv, err := New(cfg, bundle, index)
	if err != nil {
		t.Fatalf("server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/en/", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if !strings.Contains(rec.Body.String(), "Custom maintenance text.") {
		t.Fatal("expected custom notice message")
	}
}
