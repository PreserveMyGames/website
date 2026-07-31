package web

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/PreserveMyGames/website/internal/blog"
	"github.com/PreserveMyGames/website/internal/seo"
)

type foldedSearchEntry struct {
	entry blog.SearchEntry
	fold  string
}

type precomputed struct {
	robotsTXT    []byte
	sitemapXML   []byte
	searchJSON   map[string][]byte
	searchFolded map[string][]foldedSearchEntry
	rssXML       map[string][]byte
	atomXML      map[string][]byte
	alternates   map[string][]seo.Alternate
	localeOpts   map[string][]LocaleOption
}

func (s *Server) warmCaches() error {
	s.cache = &precomputed{
		searchJSON:   make(map[string][]byte),
		searchFolded: make(map[string][]foldedSearchEntry),
		rssXML:       make(map[string][]byte),
		atomXML:      make(map[string][]byte),
		alternates:   make(map[string][]seo.Alternate),
		localeOpts:   make(map[string][]LocaleOption),
	}

	s.cache.robotsTXT = fmt.Appendf(nil, "User-agent: *\nAllow: /\nSitemap: %s/sitemap.xml\n", s.cfg.SiteURL)

	locales := s.i18n.SupportedCodes()
	slugs := make(map[string][]string)
	for _, lang := range locales {
		entries := s.buildSearchEntries(lang)
		folded := make([]foldedSearchEntry, len(entries))
		for i, e := range entries {
			folded[i] = foldedSearchEntry{
				entry: e,
				fold:  strings.ToLower(e.Title + "\n" + e.Description + "\n" + e.Body),
			}
		}
		s.cache.searchFolded[lang] = folded
		data, err := json.Marshal(entries)
		if err != nil {
			return err
		}
		s.cache.searchJSON[lang] = data

		posts := s.blog.Posts(lang)
		rss, err := blog.RSS(s.cfg.SiteURL, lang, posts)
		if err != nil {
			return err
		}
		s.cache.rssXML[lang] = append(append([]byte(nil), xml.Header...), rss...)

		atom, err := blog.Atom(s.cfg.SiteURL, lang, posts)
		if err != nil {
			return err
		}
		s.cache.atomXML[lang] = append(append([]byte(nil), xml.Header...), atom...)

		for _, p := range posts {
			slugs[lang] = append(slugs[lang], p.Slug)
		}
	}

	staticSlugs := []string{"", "about", "privacy", "contact", "blog", "search"}
	for _, slug := range staticSlugs {
		s.cache.alternates[slug] = s.buildAlternates(slug)
	}
	for _, lang := range locales {
		for _, slug := range staticSlugs {
			s.cache.localeOpts[localeCacheKey(lang, slug)] = s.buildLocaleOptions(lang, slug)
		}
		for _, slug := range slugs[lang] {
			pagePath := "blog/" + slug
			if _, ok := s.cache.alternates[pagePath]; !ok {
				s.cache.alternates[pagePath] = s.buildAlternates(pagePath)
			}
			s.cache.localeOpts[localeCacheKey(lang, pagePath)] = s.buildLocaleOptions(lang, pagePath)
		}
	}

	urls := seo.SitemapURLs(s.cfg.SiteURL, locales, slugs)
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	buf.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	for _, u := range urls {
		buf.WriteString("<url><loc>")
		buf.WriteString(u)
		buf.WriteString("</loc></url>")
	}
	buf.WriteString("</urlset>")
	s.cache.sitemapXML = buf.Bytes()

	return nil
}

func localeCacheKey(lang, pagePath string) string {
	return lang + "\x00" + pagePath
}

func (s *Server) cachedAlternates(slug string) []seo.Alternate {
	if alt, ok := s.cache.alternates[slug]; ok {
		return alt
	}
	return s.buildAlternates(slug)
}

func (s *Server) cachedLocales(lang, pagePath string) []LocaleOption {
	if opts, ok := s.cache.localeOpts[localeCacheKey(lang, pagePath)]; ok {
		return opts
	}
	return s.buildLocaleOptions(lang, pagePath)
}

func (s *Server) buildAlternates(slug string) []seo.Alternate {
	codes := s.i18n.SupportedCodes()
	out := make([]seo.Alternate, len(codes))
	for i, lang := range codes {
		out[i] = seo.Alternate{
			Lang: lang,
			URL:  s.i18n.PageURL(s.cfg.SiteURL, lang, slug),
		}
	}
	return out
}

func (s *Server) buildLocaleOptions(currentLang, pagePath string) []LocaleOption {
	codes := s.i18n.SupportedCodes()
	out := make([]LocaleOption, len(codes))
	for i, code := range codes {
		out[i] = LocaleOption{
			Code:    code,
			Label:   s.i18n.T(currentLang, "lang."+code),
			URL:     s.i18n.LocalPath(code, pagePath),
			Current: code == currentLang,
		}
	}
	return out
}

func (s *Server) buildSearchEntries(lang string) []blog.SearchEntry {
	entries := s.blog.PostSearchEntries(lang)
	entries = append(entries,
		blog.SearchEntry{
			Title:       s.i18n.T(lang, "about.title"),
			Description: s.i18n.T(lang, "about.body"),
			URL:         s.i18n.LocalPath(lang, "about"),
			Type:        "page",
		},
		blog.SearchEntry{
			Title:       s.i18n.T(lang, "privacy.title"),
			Description: s.i18n.T(lang, "privacy.body"),
			URL:         s.i18n.LocalPath(lang, "privacy"),
			Type:        "page",
		},
		blog.SearchEntry{
			Title:       s.i18n.T(lang, "contact.title"),
			Description: s.i18n.T(lang, "contact.body"),
			URL:         s.i18n.LocalPath(lang, "contact"),
			Type:        "page",
		},
	)
	return entries
}
