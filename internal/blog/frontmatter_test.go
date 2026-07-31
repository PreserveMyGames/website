package blog

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/PreserveMyGames/website/internal/constants"
)

func TestSplitFrontMatterRejectsAmbiguousOpen(t *testing.T) {
	_, _, ok := splitFrontMatter([]byte("------\n---\nbody"))
	if ok {
		t.Fatal("expected ambiguous opening delimiter to be rejected")
	}
}

func TestSplitFrontMatterRejectsInlineDelimiter(t *testing.T) {
	data := []byte(`---
title: Broken
date: 2026-07-01
---
Intro line
---
This should stay in the body.
`)
	fm, body, ok := splitFrontMatter(data)
	if !ok {
		t.Fatal("expected valid front matter block")
	}
	if !strings.Contains(string(fm), "title: Broken") {
		t.Fatalf("unexpected front matter: %q", fm)
	}
	if !strings.Contains(string(body), "Intro line") || !strings.Contains(string(body), "This should stay in the body.") {
		t.Fatalf("body should include markdown after first closing delimiter, got %q", body)
	}
}

func TestSplitFrontMatterNoFalsePositive(t *testing.T) {
	data := []byte("---\nnot actually front matter")
	_, body, ok := splitFrontMatter(data)
	if ok {
		t.Fatal("expected front matter parse to fail without closing delimiter")
	}
	if !bytes.Equal(body, data) {
		t.Fatal("expected original data when front matter is invalid")
	}
}

func TestExcerptPreservesUTF8(t *testing.T) {
	body := strings.Repeat("保存", 120)
	got := excerpt("", body)
	if !utf8.ValidString(got) {
		t.Fatal("excerpt must remain valid UTF-8")
	}
	if strings.HasSuffix(got, "...") {
		if len([]rune(got)) > constants.ExcerptMaxLen+3 {
			t.Fatalf("excerpt rune length %d exceeds limit", len([]rune(got)))
		}
	}
}

func TestParsePostAuthorMetadata(t *testing.T) {
	raw := []byte(`---
title: Meta post
description: desc
date: 2026-07-01
updated: 2026-07-15
author: Ivan
author_url: https://github.com/Sudo-Ivan
draft: false
---
Body
`)
	post, err := parsePost("en", "meta-post", raw, mdParser)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if post.Author != "Ivan" {
		t.Fatalf("author: got %q", post.Author)
	}
	if post.AuthorURL != "https://github.com/Sudo-Ivan" {
		t.Fatalf("author_url: got %q", post.AuthorURL)
	}
	if post.Updated.Format("2006-01-02") != "2026-07-15" {
		t.Fatalf("updated: got %s", post.Updated)
	}
}

func TestParsePostInvalidDateFails(t *testing.T) {
	raw := []byte(`---
title: Bad date
date: not-a-date
---
`)
	_, err := parsePost("en", "bad-date", raw, mdParser)
	if err == nil {
		t.Fatal("expected invalid date to fail parsing")
	}
}

func FuzzSplitFrontMatter(f *testing.F) {
	f.Add([]byte("---\ntitle: x\ndate: 2026-01-01\n---\nbody"))
	f.Add([]byte("no front matter"))
	f.Add([]byte("---\npartial"))
	f.Fuzz(func(t *testing.T, data []byte) {
		fm, body, ok := splitFrontMatter(data)
		if ok {
			if len(fm) == 0 {
				t.Fatal("front matter must not be empty when ok")
			}
			_ = body
		}
	})
}

func FuzzParsePostNoPanic(f *testing.F) {
	f.Add([]byte("plain body"))
	f.Add([]byte("---\ntitle: t\ndate: 2026-07-01\n---\n# Hi"))
	f.Fuzz(func(t *testing.T, data []byte) {
		post, err := parsePost("en", "fuzz-slug", data, mdParser)
		if err != nil {
			return
		}
		if post.Title == "" {
			t.Fatal("title must be set")
		}
		if strings.Contains(post.BodyHTML, "<script") {
			t.Fatal("script tags must not appear in rendered HTML")
		}
	})
}

func FuzzExcerptValidUTF8(f *testing.F) {
	f.Add("desc", "body")
	f.Add("", strings.Repeat("游", 300))
	f.Fuzz(func(t *testing.T, desc, body string) {
		got := excerpt(desc, body)
		if !utf8.ValidString(got) {
			t.Fatalf("invalid UTF-8 in excerpt: %q", got)
		}
	})
}
