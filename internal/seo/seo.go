package seo

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/PreserveMyGames/website/internal/constants"
)

type Meta struct {
	Title        string
	Description  string
	Canonical    string
	OGType       string
	OGImage      string
	Locale       string
	Alternates   []Alternate
	JSONLD       []map[string]any
	JSONLDScript string
}

type Alternate struct {
	Lang string
	URL  string
}

func Home(siteURL, lang, title, description string, alternates []Alternate) Meta {
	canonical := fmt.Sprintf("%s/%s/", strings.TrimRight(siteURL, "/"), lang)
	m := Meta{
		Title:       title,
		Description: description,
		Canonical:   canonical,
		OGType:      "website",
		Locale:      lang,
		Alternates:  alternates,
		JSONLD: []map[string]any{
			{
				"@context":    "https://schema.org",
				"@type":       "WebSite",
				"name":        "PreserveMyGames",
				"url":         canonical,
				"description": description,
			},
			{
				"@context": "https://schema.org",
				"@type":    "Organization",
				"name":     "PreserveMyGames",
				"url":      strings.TrimRight(siteURL, "/"),
				"email":    constants.ContactEmail,
			},
		},
	}
	m.JSONLDScript = m.jsonLDScript()
	return m
}

func Page(siteURL, lang, slug, title, description string, alternates []Alternate) Meta {
	canonical := pageURL(siteURL, lang, slug)
	m := Meta{
		Title:       title,
		Description: description,
		Canonical:   canonical,
		OGType:      "website",
		Locale:      lang,
		Alternates:  alternates,
		JSONLD: []map[string]any{
			{
				"@context":    "https://schema.org",
				"@type":       "WebPage",
				"name":        title,
				"url":         canonical,
				"description": description,
			},
		},
	}
	m.JSONLDScript = m.jsonLDScript()
	return m
}

func BlogPost(siteURL, lang, slug, title, description string, published time.Time, alternates []Alternate) Meta {
	canonical := fmt.Sprintf("%s/%s/blog/%s", strings.TrimRight(siteURL, "/"), lang, slug)
	m := Meta{
		Title:       title,
		Description: description,
		Canonical:   canonical,
		OGType:      "article",
		Locale:      lang,
		Alternates:  alternates,
		JSONLD: []map[string]any{
			{
				"@context":      "https://schema.org",
				"@type":         "BlogPosting",
				"headline":      title,
				"description":   description,
				"url":           canonical,
				"datePublished": published.Format(time.RFC3339),
			},
		},
	}
	m.JSONLDScript = m.jsonLDScript()
	return m
}

func (m Meta) jsonLDScript() string {
	if len(m.JSONLD) == 0 {
		return ""
	}
	b, err := json.Marshal(m.JSONLD)
	if err != nil {
		return ""
	}
	return string(b)
}

func pageURL(siteURL, lang, slug string) string {
	if slug == "" {
		return fmt.Sprintf("%s/%s/", strings.TrimRight(siteURL, "/"), lang)
	}
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(siteURL, "/"), lang, slug)
}

func SitemapURLs(siteURL string, locales []string, blogSlugs map[string][]string) []string {
	base := strings.TrimRight(siteURL, "/")
	var urls []string
	for _, lang := range locales {
		urls = append(urls,
			fmt.Sprintf("%s/%s/", base, lang),
			fmt.Sprintf("%s/%s/about", base, lang),
			fmt.Sprintf("%s/%s/privacy", base, lang),
			fmt.Sprintf("%s/%s/contact", base, lang),
			fmt.Sprintf("%s/%s/blog", base, lang),
			fmt.Sprintf("%s/%s/search", base, lang),
		)
		for _, slug := range blogSlugs[lang] {
			urls = append(urls, fmt.Sprintf("%s/%s/blog/%s", base, lang, slug))
		}
	}
	return urls
}
