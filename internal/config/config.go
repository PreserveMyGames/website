package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/PreserveMyGames/website/internal/constants"
)

type Config struct {
	Port              string
	AppEnv            string
	SiteURL           string
	ContactEmail      string
	AccessLog         bool
	SiteNotice        string
	SiteNoticeMessage string
	StoreURLOverride  string
}

func Load() Config {
	return Config{
		Port:              envOr(constants.EnvPort, constants.DefaultPort),
		AppEnv:            envOr(constants.EnvAppEnv, constants.EnvDevelopment),
		SiteURL:           strings.TrimRight(envOr(constants.EnvSiteURL, constants.DefaultSiteURL), "/"),
		ContactEmail:      envOr(constants.EnvContactEmail, constants.ContactEmail),
		AccessLog:         envBool(constants.EnvAccessLog, false),
		SiteNotice:        strings.ToLower(strings.TrimSpace(envOr(constants.EnvSiteNotice, ""))),
		SiteNoticeMessage: strings.TrimSpace(envOr(constants.EnvSiteNoticeMessage, "")),
		StoreURLOverride:  strings.TrimRight(strings.TrimSpace(envOr(constants.EnvStoreURL, "")), "/"),
	}
}

func (c Config) Production() bool {
	return c.AppEnv == constants.EnvProduction
}

func (c Config) ListenPortUint16() (uint16, error) {
	v, err := strconv.ParseUint(c.Port, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(v), nil
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
