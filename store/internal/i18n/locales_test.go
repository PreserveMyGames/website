package i18n

import (
	"encoding/json"
	"io/fs"
	"testing"
)

func TestLocaleKeyParity(t *testing.T) {
	entries, err := fs.ReadDir(localeFS, "locales")
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	type msg struct {
		ID string `json:"id"`
	}
	keys := map[string]map[string]struct{}{}
	ref := ""
	for _, e := range entries {
		if e.IsDir() || len(e.Name()) < 6 || e.Name()[len(e.Name())-5:] != ".json" {
			continue
		}
		code := e.Name()[:len(e.Name())-5]
		data, err := fs.ReadFile(localeFS, "locales/"+e.Name())
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		var msgs []msg
		if err := json.Unmarshal(data, &msgs); err != nil {
			t.Fatalf("parse %s: %v", e.Name(), err)
		}
		set := make(map[string]struct{}, len(msgs))
		for _, m := range msgs {
			set[m.ID] = struct{}{}
		}
		keys[code] = set
		if code == "en" || ref == "" {
			ref = code
		}
	}
	if ref == "" {
		t.Fatal("no locales")
	}
	for code, set := range keys {
		if len(set) != len(keys[ref]) {
			t.Fatalf("%s.json has %d keys, reference has %d", code, len(set), len(keys[ref]))
		}
		for id := range keys[ref] {
			if _, ok := set[id]; !ok {
				t.Fatalf("%s missing key %s", code, id)
			}
		}
	}
}
