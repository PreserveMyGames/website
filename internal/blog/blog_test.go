package blog

import (
	"strings"
	"testing"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func TestDraftHiddenInProduction(t *testing.T) {
	idx, err := Load(true)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, post := range idx.Posts("en") {
		if post.Draft {
			t.Fatalf("draft post %q leaked in production index", post.Slug)
		}
	}
}

func TestMarkdownStripsRawHTML(t *testing.T) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithHardWraps()),
	)

	raw := []byte(`---
title: XSS probe
description: test
date: 2026-07-02
draft: true
tags: [security]
---
<script>alert("xss")</script>
# Safe heading
`)
	post, err := parsePost("en", "xss-probe", raw, md)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if strings.Contains(post.BodyHTML, "<script") {
		t.Fatalf("raw script tag rendered: %q", post.BodyHTML)
	}
	if !strings.Contains(post.BodyHTML, "Safe heading") {
		t.Fatal("expected markdown heading to render")
	}
}

func TestParsePostDefaults(t *testing.T) {
	md := goldmark.New()
	post, err := parsePost("en", "untitled", []byte("Body only"), md)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if post.Title != "untitled" {
		t.Fatalf("title: got %q", post.Title)
	}
	if post.Date.IsZero() {
		t.Fatal("expected default date")
	}
	if post.Date.Location() != time.UTC {
		t.Fatal("expected UTC date")
	}
}

func TestPostSearchEntries(t *testing.T) {
	idx, err := Load(false)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	entries := idx.PostSearchEntries("en")
	if len(entries) == 0 {
		t.Fatal("expected at least one post entry")
	}
	if entries[0].Type != "post" {
		t.Fatalf("expected post type, got %q", entries[0].Type)
	}
}
