package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/PreserveMyGames/store/internal/constants"
)

type Config struct {
	Port                   string
	AppEnv                 string
	SiteURL                string
	DataDir                string
	AccessLog              bool
	BaseCurrency           string
	StripeSecretKey        string
	StripeWebhookSecret    string
	MoneroRPCURL           string
	MoneroRPCUser          string
	MoneroRPCPass          string
	MoneroMinConfirmations int
	AdminUser              string
	AdminPassword          string
	MainSiteURL            string
}

func Load() Config {
	return Config{
		Port:                   envOr(constants.EnvPort, constants.DefaultPort),
		AppEnv:                 envOr(constants.EnvAppEnv, constants.EnvDevelopment),
		SiteURL:                strings.TrimRight(envOr(constants.EnvSiteURL, constants.DefaultSiteURL), "/"),
		DataDir:                envOr(constants.EnvDataDir, constants.DefaultDataDir),
		AccessLog:              envBool(constants.EnvAccessLog, false),
		BaseCurrency:           strings.ToUpper(envOr(constants.EnvBaseCurrency, constants.DefaultBaseCurrency)),
		StripeSecretKey:        strings.TrimSpace(os.Getenv(constants.EnvStripeSecretKey)),
		StripeWebhookSecret:    strings.TrimSpace(os.Getenv(constants.EnvStripeWebhookSecret)),
		MoneroRPCURL:           strings.TrimSpace(os.Getenv(constants.EnvMoneroRPCURL)),
		MoneroRPCUser:          strings.TrimSpace(os.Getenv(constants.EnvMoneroRPCUser)),
		MoneroRPCPass:          strings.TrimSpace(os.Getenv(constants.EnvMoneroRPCPass)),
		MoneroMinConfirmations: envInt(constants.EnvMoneroMinConfirmations, constants.DefaultMinConfirmations),
		AdminUser:              strings.TrimSpace(os.Getenv(constants.EnvAdminUser)),
		AdminPassword:          strings.TrimSpace(os.Getenv(constants.EnvAdminPassword)),
		MainSiteURL:            strings.TrimRight(envOr(constants.EnvMainSiteURL, "https://preservemygames.org"), "/"),
	}
}

func (c Config) Production() bool {
	return c.AppEnv == constants.EnvProduction
}

func (c Config) StripeEnabled() bool {
	return c.StripeSecretKey != ""
}

func (c Config) MoneroEnabled() bool {
	return c.MoneroRPCURL != ""
}

func (c Config) AdminEnabled() bool {
	return c.AdminUser != "" && c.AdminPassword != ""
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

func envInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
