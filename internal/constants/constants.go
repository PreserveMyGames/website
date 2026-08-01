package constants

import "time"

const (
	AppName        = "Preserve My Games"
	DefaultLocale  = "en"
	LangCookieName = "lang"
	AssetVersion   = "dev"
	ContactEmail   = "contact@preservemygames.org"
	ContactLXMF    = "f489752fbef161c64d65e385a4e9fc74"
	DonateMonero   = "87wT8aCdaY2JLkVzkcwc8G8LqSy23PsFgPWnjKybVqg5ce6j2pLdYP1d6nm7qtpJgzfKqaQAiCfDZRqasJMnuCNN99jeduq"
	DonateKoFiURL  = "https://ko-fi.com/preservemygames"

	NoticeConstruction = "construction"
	NoticeMaintenance  = "maintenance"
	NoticeInfo         = "info"
	NoticeWarning      = "warning"

	DefaultPort    = "8080"
	DefaultSiteURL = "http://localhost:8080"

	EnvPort              = "PORT"
	EnvAppEnv            = "APP_ENV"
	EnvSiteURL           = "SITE_URL"
	EnvContactEmail      = "CONTACT_EMAIL"
	EnvAccessLog         = "ACCESS_LOG"
	EnvSiteNotice        = "SITE_NOTICE"
	EnvSiteNoticeMessage = "SITE_NOTICE_MESSAGE"
	EnvProduction        = "production"
	EnvDevelopment       = "development"

	PathHealthz     = "/healthz"
	PathRobots      = "/robots.txt"
	PathSitemap     = "/sitemap.xml"
	PathStatic      = "/static/"
	HealthzResponse = "ok"

	HeaderCSP               = "Content-Security-Policy"
	HeaderXContentType      = "X-Content-Type-Options"
	HeaderReferrerPolicy    = "Referrer-Policy"
	HeaderXFrameOptions     = "X-Frame-Options"
	HeaderPermissionsPolicy = "Permissions-Policy"
	HeaderStrictTransport   = "Strict-Transport-Security"
	HeaderContentType       = "Content-Type"
	HeaderContentLength     = "Content-Length"
	HeaderCacheControl      = "Cache-Control"

	CSPPolicy = "default-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'; object-src 'none'"
	Nosniff   = "nosniff"
	Referrer  = "strict-origin-when-cross-origin"
	DenyFrame = "DENY"
	Perms     = "camera=(), microphone=(), geolocation=()"
	HSTS      = "max-age=31536000; includeSubDomains"

	StaticCacheControl = "public, max-age=31536000, immutable"
	PageCacheControl   = "no-cache"
	DevCacheControl    = "no-cache"
	SearchCacheControl = "public, max-age=300"

	ReadHeaderTimeout = 5 * time.Second
	ReadTimeout       = 10 * time.Second
	WriteTimeout      = 15 * time.Second
	IdleTimeout       = 60 * time.Second
	ShutdownTimeout   = 15 * time.Second
	MaxHeaderBytes    = 1 << 20

	ExcerptMaxLen = 200
)
