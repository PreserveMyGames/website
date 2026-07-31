package blog

import (
	"encoding/xml"
	"fmt"
	"html"
	"strings"
	"time"
)

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language"`
	LastBuild   string    `xml:"lastBuildDate"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Author      string `xml:"author,omitempty"`
}

type atomFeed struct {
	XMLName xml.Name    `xml:"http://www.w3.org/2005/Atom feed"`
	Title   string      `xml:"title"`
	ID      string      `xml:"id"`
	Link    atomLink    `xml:"link"`
	Updated string      `xml:"updated"`
	Entries []atomEntry `xml:"entry"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

type atomEntry struct {
	Title   string      `xml:"title"`
	ID      string      `xml:"id"`
	Link    atomLink    `xml:"link"`
	Updated string      `xml:"updated"`
	Summary string      `xml:"summary"`
	Author  *atomAuthor `xml:"author,omitempty"`
}

type atomAuthor struct {
	Name string `xml:"name"`
	URI  string `xml:"uri,omitempty"`
}

func RSS(siteURL, locale string, posts []Post) ([]byte, error) {
	feed := rssFeed{Version: "2.0"}
	feed.Channel = rssChannel{
		Title:       "PreserveMyGames Blog",
		Link:        fmt.Sprintf("%s/%s/blog", strings.TrimRight(siteURL, "/"), locale),
		Description: "Updates from PreserveMyGames",
		Language:    locale,
		LastBuild:   time.Now().UTC().Format(time.RFC1123Z),
	}
	for _, p := range posts {
		link := fmt.Sprintf("%s/%s/blog/%s", strings.TrimRight(siteURL, "/"), locale, p.Slug)
		item := rssItem{
			Title:       p.Title,
			Link:        link,
			GUID:        link,
			Description: html.EscapeString(p.Description),
			PubDate:     p.Date.Format(time.RFC1123Z),
		}
		if p.Author != "" {
			item.Author = p.Author
		}
		feed.Channel.Items = append(feed.Channel.Items, item)
	}
	return xml.MarshalIndent(feed, "", "  ")
}

func Atom(siteURL, locale string, posts []Post) ([]byte, error) {
	base := strings.TrimRight(siteURL, "/")
	feed := atomFeed{
		Title:   "PreserveMyGames Blog",
		ID:      fmt.Sprintf("%s/%s/blog", base, locale),
		Link:    atomLink{Href: fmt.Sprintf("%s/%s/blog", base, locale), Rel: "self"},
		Updated: time.Now().UTC().Format(time.RFC3339),
	}
	for _, p := range posts {
		link := fmt.Sprintf("%s/%s/blog/%s", base, locale, p.Slug)
		entry := atomEntry{
			Title:   p.Title,
			ID:      link,
			Link:    atomLink{Href: link, Rel: "alternate"},
			Updated: pubTime(p).Format(time.RFC3339),
			Summary: p.Description,
		}
		if p.Author != "" {
			entry.Author = &atomAuthor{Name: p.Author}
			if p.AuthorURL != "" {
				entry.Author.URI = p.AuthorURL
			}
		}
		feed.Entries = append(feed.Entries, entry)
	}
	return xml.MarshalIndent(feed, "", "  ")
}

func pubTime(p Post) time.Time {
	if !p.Updated.IsZero() {
		return p.Updated
	}
	return p.Date
}
