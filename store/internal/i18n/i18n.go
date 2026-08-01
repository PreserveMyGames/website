package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localeFS embed.FS

// Bundle holds translations for supported locales.
type Bundle struct {
	bundle *i18n.Bundle
	codes  []string
}

func New() (*Bundle, error) {
	b := i18n.NewBundle(language.English)
	b.RegisterUnmarshalFunc("json", json.Unmarshal)

	entries, err := fs.ReadDir(localeFS, "locales")
	if err != nil {
		return nil, err
	}

	var codes []string
	for _, e := range entries {
		if e.IsDir() || len(e.Name()) < 6 || e.Name()[len(e.Name())-5:] != ".json" {
			continue
		}
		code := e.Name()[:len(e.Name())-5]
		path := "locales/" + e.Name()
		data, err := fs.ReadFile(localeFS, path)
		if err != nil {
			return nil, err
		}
		if _, err := b.ParseMessageFileBytes(data, path); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		codes = append(codes, code)
	}
	if len(codes) == 0 {
		return nil, fmt.Errorf("no locale files found")
	}
	return &Bundle{bundle: b, codes: codes}, nil
}

func (b *Bundle) SupportedCodes() []string {
	out := make([]string, len(b.codes))
	copy(out, b.codes)
	return out
}

func (b *Bundle) Has(code string) bool {
	for _, c := range b.codes {
		if c == code {
			return true
		}
	}
	return false
}

func (b *Bundle) T(locale, id string) string {
	localizer := i18n.NewLocalizer(b.bundle, locale)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: id})
	if err != nil {
		return id
	}
	return msg
}

func (b *Bundle) TWith(locale, id string, data map[string]any) string {
	localizer := i18n.NewLocalizer(b.bundle, locale)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    id,
		TemplateData: data,
	})
	if err != nil {
		return id
	}
	return msg
}

func (b *Bundle) LocalPath(lang, pagePath string) string {
	if pagePath == "" {
		return "/" + lang + "/"
	}
	return "/" + lang + "/" + pagePath
}

func (b *Bundle) PageURL(siteURL, lang, pagePath string) string {
	return siteURL + b.LocalPath(lang, pagePath)
}
