package blog

import (
	"strings"
	"testing"
	"time"
)

func TestRSSAndAtomFeeds(t *testing.T) {
	posts := []Post{
		{
			Slug:        "welcome",
			Title:       "Welcome",
			Description: "Intro",
			Author:      "Ivan",
			AuthorURL:   "https://github.com/Sudo-Ivan",
			Date:        time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			Updated:     time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	rss, err := RSS("https://example.com", "en", posts)
	if err != nil {
		t.Fatalf("rss: %v", err)
	}
	rssStr := string(rss)
	if !strings.Contains(rssStr, "<title>Welcome</title>") {
		t.Fatal("expected rss item title")
	}
	if !strings.Contains(rssStr, "<author>Ivan</author>") {
		t.Fatal("expected rss author")
	}

	atom, err := Atom("https://example.com", "en", posts)
	if err != nil {
		t.Fatalf("atom: %v", err)
	}
	atomStr := string(atom)
	if !strings.Contains(atomStr, "<title>Welcome</title>") {
		t.Fatal("expected atom entry title")
	}
	if !strings.Contains(atomStr, "<name>Ivan</name>") {
		t.Fatal("expected atom author name")
	}
	if !strings.Contains(atomStr, "2026-07-15") {
		t.Fatal("expected atom updated date from updated field")
	}
}
