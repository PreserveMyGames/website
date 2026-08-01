package constants

import "time"

const (
	AppName        = "Preserve My Games Store"
	DefaultLocale  = "en"
	LangCookieName = "lang"
	CurrencyCookie = "currency"
	AssetVersion   = "dev"

	DefaultPort         = "8080"
	DefaultSiteURL      = "http://localhost:8080"
	DefaultDataDir      = "./data"
	DefaultBaseCurrency = "USD"

	EnvPort                   = "PORT"
	EnvAppEnv                 = "APP_ENV"
	EnvSiteURL                = "SITE_URL"
	EnvDataDir                = "DATA_DIR"
	EnvAccessLog              = "ACCESS_LOG"
	EnvBaseCurrency           = "BASE_CURRENCY"
	EnvStripeSecretKey        = "STRIPE_SECRET_KEY"
	EnvStripeWebhookSecret    = "STRIPE_WEBHOOK_SECRET"
	EnvMoneroRPCURL           = "MONERO_RPC_URL"
	EnvMoneroRPCUser          = "MONERO_RPC_USER"
	EnvMoneroRPCPass          = "MONERO_RPC_PASS"
	EnvMoneroMinConfirmations = "MONERO_MIN_CONFIRMATIONS"
	EnvAdminUser              = "ADMIN_USER"
	EnvAdminPassword          = "ADMIN_PASSWORD"
	EnvMainSiteURL            = "MAIN_SITE_URL"
	EnvProduction             = "production"
	EnvDevelopment            = "development"

	PathHealthz     = "/healthz"
	PathRobots      = "/robots.txt"
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

	CSPPolicy = "default-src 'self'; base-uri 'self'; form-action 'self' https://checkout.stripe.com; frame-ancestors 'none'; object-src 'none'; img-src 'self' data:"
	Nosniff   = "nosniff"
	Referrer  = "strict-origin-when-cross-origin"
	DenyFrame = "DENY"
	Perms     = "camera=(), microphone=(), geolocation=()"
	HSTS      = "max-age=31536000; includeSubDomains"

	StaticCacheControl = "public, max-age=31536000, immutable"
	PageCacheControl   = "no-cache"
	DevCacheControl    = "no-cache"
	FeedCacheControl   = "public, max-age=300"

	ReadHeaderTimeout = 5 * time.Second
	ReadTimeout       = 10 * time.Second
	WriteTimeout      = 30 * time.Second
	IdleTimeout       = 60 * time.Second
	ShutdownTimeout   = 15 * time.Second
	MaxHeaderBytes    = 1 << 20

	ProductDigital  = "digital"
	ProductPhysical = "physical"

	OrderPending        = "pending"
	OrderAwaitingStripe = "awaiting_stripe"
	OrderAwaitingMonero = "awaiting_monero"
	OrderPaid           = "paid"
	OrderAwaitingShip   = "awaiting_fulfillment"
	OrderShipped        = "shipped"
	OrderCancelled      = "cancelled"
	OrderExpired        = "expired"

	PaymentStripe = "stripe"
	PaymentMonero = "monero"

	ReservationTTL          = 30 * time.Minute
	DefaultMinConfirmations = 10
	FXRefreshInterval       = 6 * time.Hour
	MoneroPollInterval      = 30 * time.Second
)
