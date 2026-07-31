package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"

	"github.com/PreserveMyGames/website/internal/constants"
)

//go:embed locales/*.json
var localeFS embed.FS

type Bundle struct {
	bundle     *i18n.Bundle
	supported  []language.Tag
	codes      []string
	localizers map[string]*i18n.Localizer
}

func New() (*Bundle, error) {
	b := i18n.NewBundle(language.English)
	b.RegisterUnmarshalFunc("json", json.Unmarshal)

	entries, err := localeFS.ReadDir("locales")
	if err != nil {
		return nil, err
	}

	supported, err := discoverLocales(entries)
	if err != nil {
		return nil, err
	}
	if len(supported) == 0 {
		return nil, fmt.Errorf("no locale files found")
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := localeFS.ReadFile("locales/" + entry.Name())
		if err != nil {
			return nil, err
		}
		if _, err := b.ParseMessageFileBytes(data, entry.Name()); err != nil {
			return nil, err
		}
	}

	codes := make([]string, len(supported))
	localizers := make(map[string]*i18n.Localizer, len(supported))
	for i, tag := range supported {
		code := tag.String()
		codes[i] = code
		localizers[code] = i18n.NewLocalizer(b, code)
	}

	return &Bundle{
		bundle:     b,
		supported:  supported,
		codes:      codes,
		localizers: localizers,
	}, nil
}

func discoverLocales(entries []fs.DirEntry) ([]language.Tag, error) {
	var tags []language.Tag
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		code := strings.TrimSuffix(entry.Name(), ".json")
		tag, err := language.Parse(code)
		if err != nil {
			return nil, fmt.Errorf("locale %s: %w", code, err)
		}
		tags = append(tags, tag)
	}

	sort.Slice(tags, func(i, j int) bool {
		if tags[i].String() == constants.DefaultLocale {
			return true
		}
		if tags[j].String() == constants.DefaultLocale {
			return false
		}
		return tags[i].String() < tags[j].String()
	})

	return tags, nil
}

func (b *Bundle) Supported() []language.Tag {
	return b.supported
}

func (b *Bundle) SupportedCodes() []string {
	return b.codes
}

func (b *Bundle) Match(accept string) string {
	if strings.TrimSpace(accept) == "" {
		return constants.DefaultLocale
	}
	tags, _, _ := language.ParseAcceptLanguage(accept)
	matcher := language.NewMatcher(b.supported)
	tag, _, _ := matcher.Match(tags...)
	for _, supported := range b.supported {
		if tag == supported {
			return supported.String()
		}
		base, _ := tag.Base()
		sbase, _ := supported.Base()
		if base == sbase {
			return supported.String()
		}
	}
	return constants.DefaultLocale
}

func (b *Bundle) Localizer(lang string) *i18n.Localizer {
	if l, ok := b.localizers[lang]; ok {
		return l
	}
	return i18n.NewLocalizer(b.bundle, lang)
}

func (b *Bundle) T(lang, id string) string {
	l := b.localizers[lang]
	if l == nil {
		l = b.Localizer(lang)
	}
	msg, err := l.Localize(&i18n.LocalizeConfig{MessageID: id})
	if err != nil {
		return id
	}
	return msg
}

func (b *Bundle) TWith(lang, id string, data map[string]any) string {
	l := b.localizers[lang]
	if l == nil {
		l = b.Localizer(lang)
	}
	msg, err := l.Localize(&i18n.LocalizeConfig{
		MessageID:    id,
		TemplateData: data,
	})
	if err != nil {
		return id
	}
	return msg
}

func (b *Bundle) Allowed(code string) bool {
	code = strings.ToLower(strings.TrimSpace(code))
	for _, tag := range b.supported {
		if tag.String() == code {
			return true
		}
	}
	return false
}

func (b *Bundle) PathPrefix(lang string) string {
	return "/" + lang
}

func (b *Bundle) PageURL(siteURL, lang, path string) string {
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return fmt.Sprintf("%s/%s/", strings.TrimRight(siteURL, "/"), lang)
	}
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(siteURL, "/"), lang, path)
}

func (b *Bundle) LocalPath(lang, pagePath string) string {
	pagePath = strings.TrimPrefix(pagePath, "/")
	if pagePath == "" {
		return "/" + lang + "/"
	}
	return "/" + lang + "/" + pagePath
}
