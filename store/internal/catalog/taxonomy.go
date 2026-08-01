package catalog

import (
	"encoding/json"
	"strings"
)

func encodeTags(tags []string) string {
	if len(tags) == 0 {
		return "[]"
	}
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		t = normalizeTag(t)
		if t != "" {
			out = append(out, t)
		}
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func decodeTags(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}

func normalizeTag(tag string) string {
	tag = strings.ToLower(strings.TrimSpace(tag))
	var b strings.Builder
	for _, r := range tag {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func normalizeCategory(cat string) string {
	return normalizeTag(cat)
}

func productMatchesTag(p Product, tag string) bool {
	want := normalizeTag(tag)
	if want == "" {
		return false
	}
	for _, t := range p.Tags {
		if t == want {
			return true
		}
	}
	return false
}

func productMatchesQuery(p Product, query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return true
	}
	hay := strings.ToLower(p.Title + " " + p.Description + " " + p.Category + " " + p.SKU + " " + strings.Join(p.Tags, " "))
	for _, term := range strings.Fields(query) {
		if !strings.Contains(hay, term) {
			return false
		}
	}
	return true
}

func CollectCategories(products []Product) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, p := range products {
		c := normalizeCategory(p.Category)
		if c == "" {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	return out
}

func CollectTags(products []Product) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, p := range products {
		for _, t := range p.Tags {
			if t == "" {
				continue
			}
			if _, ok := seen[t]; ok {
				continue
			}
			seen[t] = struct{}{}
			out = append(out, t)
		}
	}
	return out
}
