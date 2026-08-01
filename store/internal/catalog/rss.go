package catalog

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
}

func RSS(siteURL, locale, feedTitle, feedDesc, pagePath string, products []Product) ([]byte, error) {
	base := strings.TrimRight(siteURL, "/")
	feed := rssFeed{Version: "2.0"}
	feed.Channel = rssChannel{
		Title:       feedTitle,
		Link:        base + pagePath,
		Description: feedDesc,
		Language:    locale,
		LastBuild:   time.Now().UTC().Format(time.RFC1123Z),
	}
	for _, p := range products {
		link := fmt.Sprintf("%s/%s/%s", base, locale, p.Slug)
		item := rssItem{
			Title:       p.Title,
			Link:        link,
			GUID:        link,
			Description: html.EscapeString(p.Description),
			PubDate:     time.Now().UTC().Format(time.RFC1123Z),
		}
		feed.Channel.Items = append(feed.Channel.Items, item)
	}
	return xml.MarshalIndent(feed, "", "  ")
}
