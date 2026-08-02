package config

import (
	"net/url"
	"strings"

	"github.com/PreserveMyGames/website/internal/constants"
)

func (c Config) SiteHost() string {
	u, err := url.Parse(c.SiteURL)
	if err != nil || u.Hostname() == "" {
		return "preservemygames.org"
	}
	return u.Hostname()
}

func (c Config) WikiURL() string {
	return "https://wiki." + c.SiteHost()
}

func (c Config) ForumsURL() string {
	return "https://forums." + c.SiteHost()
}

func (c Config) FilesURL() string {
	return "https://files." + c.SiteHost()
}

func (c Config) StoreURL() string {
	if c.StoreURLOverride != "" {
		return c.StoreURLOverride
	}
	return "https://store." + c.SiteHost()
}

func (c Config) NoticeKind() string {
	kind := strings.ToLower(strings.TrimSpace(c.SiteNotice))
	switch kind {
	case constants.NoticeConstruction,
		constants.NoticeMaintenance,
		constants.NoticeInfo,
		constants.NoticeWarning:
		return kind
	default:
		return ""
	}
}
