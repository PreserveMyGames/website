package render

import (
	"html/template"
	"strings"
	"testing"

	"github.com/PreserveMyGames/website/internal/seo"
)

func TestEngineExecuteLayout(t *testing.T) {
	funcs := template.FuncMap{
		"t": func(id string) string {
			return id
		},
	}
	engine, err := New(funcs)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	content, err := engine.ExecuteBytes("home-content", map[string]string{
		"Lang": "en",
	})
	if err != nil {
		t.Fatalf("content: %v", err)
	}
	if !strings.Contains(string(content), "home.headline") {
		t.Fatal("expected home content template to render")
	}

	view := struct {
		Lang            string
		Body            template.HTML
		AssetVersion    string
		WikiURL         string
		ForumsURL       string
		Notice          any
		IncludeSearchJS bool
		Locales         []struct {
			Code, Label, URL string
			Current          bool
		}
		Meta seo.Meta
	}{
		Lang:         "en",
		Body:         template.HTML(content),
		AssetVersion: "1",
		WikiURL:      "https://wiki.example.com",
		ForumsURL:    "https://forums.example.com",
		Locales: []struct {
			Code, Label, URL string
			Current          bool
		}{
			{Code: "en", Label: "English", URL: "/en/", Current: true},
		},
		Meta: seo.Page("https://example.com", "en", "", "Test", "Desc", nil),
	}

	body, err := engine.ExecuteBytes("layout", view)
	if err != nil {
		t.Fatalf("layout: %v", err)
	}
	if !strings.Contains(string(body), "<html lang=\"en\"") {
		t.Fatal("expected html shell")
	}
}
