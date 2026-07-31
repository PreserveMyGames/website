package i18n

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"
)

func TestSupportedLocales(t *testing.T) {
	bundle, err := New()
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	codes := bundle.SupportedCodes()
	want := []string{"en", "de", "ru"}
	if len(codes) != len(want) {
		t.Fatalf("supported codes: got %v want %v", codes, want)
	}
	for i, code := range want {
		if codes[i] != code {
			t.Fatalf("codes[%d]: got %q want %q", i, codes[i], code)
		}
	}
}

func TestMatchGermanAndRussian(t *testing.T) {
	bundle, err := New()
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	if got := bundle.Match("de-DE,de;q=0.9"); got != "de" {
		t.Fatalf("de accept: got %q", got)
	}
	if got := bundle.Match("ru-RU,ru;q=0.9"); got != "ru" {
		t.Fatalf("ru accept: got %q", got)
	}
}

func TestLocaleKeyParity(t *testing.T) {
	entries, err := localeFS.ReadDir("locales")
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}

	byFile := make(map[string][]string)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := localeFS.ReadFile("locales/" + entry.Name())
		if err != nil {
			t.Fatalf("read %s: %v", entry.Name(), err)
		}
		var messages []struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(data, &messages); err != nil {
			t.Fatalf("parse %s: %v", entry.Name(), err)
		}
		ids := make([]string, 0, len(messages))
		for _, msg := range messages {
			ids = append(ids, msg.ID)
		}
		sort.Strings(ids)
		byFile[entry.Name()] = ids
	}

	var reference []string
	for file, ids := range byFile {
		if reference == nil {
			reference = ids
			continue
		}
		if len(ids) != len(reference) {
			t.Fatalf("%s has %d keys, reference has %d", file, len(ids), len(reference))
		}
		for i := range reference {
			if ids[i] != reference[i] {
				t.Fatalf("%s key mismatch at %d: got %q want %q", file, i, ids[i], reference[i])
			}
		}
	}
}

func TestGermanTranslation(t *testing.T) {
	bundle, err := New()
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	got := bundle.T("de", "home.headline")
	if got != "Spiele verdienen ein dauerhaftes Zuhause." {
		t.Fatalf("de headline: got %q", got)
	}
}

func TestRussianTranslation(t *testing.T) {
	bundle, err := New()
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	got := bundle.T("ru", "home.headline")
	if got != "Игры заслуживают постоянного дома." {
		t.Fatalf("ru headline: got %q", got)
	}
}

func TestLocalPath(t *testing.T) {
	bundle, err := New()
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if got := bundle.LocalPath("de", "about"); got != "/de/about" {
		t.Fatalf("local path: got %q", got)
	}
	if got := bundle.LocalPath("ru", ""); got != "/ru/" {
		t.Fatalf("home path: got %q", got)
	}
}
