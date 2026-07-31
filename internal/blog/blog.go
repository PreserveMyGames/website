package blog

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"
)

//go:embed all:content
var contentFS embed.FS

type Post struct {
	Author      string
	AuthorURL   string
	Updated     time.Time
	Locale      string
	Slug        string
	Title       string
	Description string
	Date        time.Time
	Draft       bool
	Tags        []string
	BodyHTML    string
	Excerpt     string
}

type Index struct {
	byLocale map[string][]Post
	bySlug   map[string]map[string]Post
}

type SearchEntry struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Tags        []string `json:"tags"`
	Body        string   `json:"body"`
	Type        string   `json:"type"`
}

var mdParser = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	goldmark.WithRendererOptions(html.WithHardWraps()),
)

func Load(production bool) (*Index, error) {
	idx := &Index{
		byLocale: make(map[string][]Post),
		bySlug:   make(map[string]map[string]Post),
	}

	md := mdParser

	err := fs.WalkDir(contentFS, "content/blog", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(filePath, ".md") {
			return nil
		}

		rel := strings.TrimPrefix(filePath, "content/blog/")
		parts := strings.Split(rel, "/")
		if len(parts) != 2 {
			return nil
		}
		locale := parts[0]
		slug := strings.TrimSuffix(parts[1], ".md")

		data, err := contentFS.ReadFile(filePath)
		if err != nil {
			return err
		}

		post, err := parsePost(locale, slug, data, md)
		if err != nil {
			return fmt.Errorf("parse %s: %w", filePath, err)
		}
		if production && post.Draft {
			return nil
		}

		idx.byLocale[locale] = append(idx.byLocale[locale], post)
		if idx.bySlug[locale] == nil {
			idx.bySlug[locale] = make(map[string]Post)
		}
		idx.bySlug[locale][slug] = post
		return nil
	})
	if err != nil {
		return nil, err
	}

	for locale := range idx.byLocale {
		sort.Slice(idx.byLocale[locale], func(i, j int) bool {
			return idx.byLocale[locale][i].Date.After(idx.byLocale[locale][j].Date)
		})
	}

	return idx, nil
}

func parsePost(locale, slug string, data []byte, md goldmark.Markdown) (Post, error) {
	var post Post
	post.Locale = locale
	post.Slug = slug

	body := data
	if fm, rest, ok := splitFrontMatter(data); ok {
		var meta struct {
			Title       string   `yaml:"title"`
			Description string   `yaml:"description"`
			Date        string   `yaml:"date"`
			Updated     string   `yaml:"updated"`
			Draft       bool     `yaml:"draft"`
			Tags        []string `yaml:"tags"`
			Author      string   `yaml:"author"`
			AuthorURL   string   `yaml:"author_url"`
		}
		if err := yaml.Unmarshal(fm, &meta); err != nil {
			return post, err
		}
		post.Title = meta.Title
		post.Description = meta.Description
		post.Draft = meta.Draft
		post.Tags = meta.Tags
		post.Author = strings.TrimSpace(meta.Author)
		post.AuthorURL = strings.TrimSpace(meta.AuthorURL)
		if meta.Date != "" {
			t, err := time.Parse("2006-01-02", meta.Date)
			if err != nil {
				return post, err
			}
			post.Date = t
		}
		if meta.Updated != "" {
			t, err := time.Parse("2006-01-02", meta.Updated)
			if err != nil {
				return post, err
			}
			post.Updated = t
		}
		body = rest
	}

	if post.Title == "" {
		post.Title = slug
	}
	if post.Date.IsZero() {
		post.Date = time.Now().UTC()
	}

	var htmlBuf bytes.Buffer
	if err := md.Convert(body, &htmlBuf); err != nil {
		return post, err
	}
	post.BodyHTML = htmlBuf.String()
	post.Excerpt = excerpt(post.Description, string(body))
	return post, nil
}

func (i *Index) Posts(locale string) []Post {
	posts := i.byLocale[locale]
	if posts == nil {
		return emptyPosts
	}
	return posts
}

var emptyPosts []Post

func (i *Index) Post(locale, slug string) (Post, bool) {
	if m, ok := i.bySlug[locale]; ok {
		p, ok := m[slug]
		return p, ok
	}
	return Post{}, false
}

func (i *Index) PostSearchEntries(locale string) []SearchEntry {
	var out []SearchEntry
	for _, p := range i.Posts(locale) {
		out = append(out, SearchEntry{
			Title:       p.Title,
			Description: p.Description,
			URL:         path.Join("/", locale, "blog", p.Slug),
			Tags:        p.Tags,
			Body:        p.Excerpt,
			Type:        "post",
		})
	}
	return out
}
