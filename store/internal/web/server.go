package web

import (
	"context"
	"crypto/subtle"
	"encoding/xml"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/skip2/go-qrcode"

	"github.com/PreserveMyGames/store/internal/catalog"
	"github.com/PreserveMyGames/store/internal/config"
	"github.com/PreserveMyGames/store/internal/constants"
	"github.com/PreserveMyGames/store/internal/fx"
	"github.com/PreserveMyGames/store/internal/i18n"
	"github.com/PreserveMyGames/store/internal/payments/monero"
	stripepay "github.com/PreserveMyGames/store/internal/payments/stripe"
	"github.com/PreserveMyGames/store/internal/render"
	"github.com/PreserveMyGames/store/internal/staticfiles"
)

type Server struct {
	cfg     config.Config
	i18n    *i18n.Bundle
	store   *catalog.Store
	fx      *fx.Cache
	stripe  *stripepay.Client
	monero  monero.Wallet
	quoter  *monero.Quoter
	engines map[string]*render.Engine
	static  http.Handler
	handler http.Handler
}

type LocaleOption struct {
	Code    string
	Label   string
	URL     string
	Current bool
}

type EmptyStateView struct {
	Title string
	Body  string
}

type ProductView struct {
	catalog.Product
	DisplayPrice   string
	ReferencePrice string
	OutOfStock     bool
	StockLabel     string
	StockValue     string
}

type PageView struct {
	Lang                   string
	Title                  string
	Description            string
	Body                   template.HTML
	AssetVersion           string
	MainSiteURL            string
	Locales                []LocaleOption
	Currency               string
	Currencies             []string
	Products               []ProductView
	Product                catalog.Product
	ProductTypeLabel       string
	DisplayPrice           string
	ReferencePrice         string
	StockLabel             string
	OutOfStock             bool
	StripeEnabled          bool
	MoneroEnabled          bool
	EmptyState             *EmptyStateView
	Order                  catalog.Order
	StatusLabel            string
	MoneroAddress          string
	MoneroAmount           string
	QRDataURI              template.URL
	ExpiresAt              string
	DigitalReady           bool
	DigitalLinks           []string
	Orders                 []catalog.Order
	RefreshSeconds         int
	ErrorStatus            int
	ErrorTitle             string
	ErrorLead              string
	CatalogTitle           string
	CatalogLead            string
	Query                  string
	ResultsCount           string
	Categories             []string
	Tags                   []string
	ActiveCategory         string
	ActiveTag              string
	FeedPath               string
	ShowCatalogSearch      bool
	IncludeCatalogSearchJS bool
}

var assetVersion = constants.AssetVersion

func New(cfg config.Config, bundle *i18n.Bundle, store *catalog.Store, rates *fx.Cache, wallet monero.Wallet) (*Server, error) {
	engines := make(map[string]*render.Engine, len(bundle.SupportedCodes()))
	for _, lang := range bundle.SupportedCodes() {
		locale := lang
		funcs := template.FuncMap{
			"t": func(id string) string {
				return bundle.T(locale, id)
			},
		}
		engine, err := render.New(funcs)
		if err != nil {
			return nil, err
		}
		engines[locale] = engine
	}

	staticRoot, err := fs.Sub(staticfiles.FS, "static")
	if err != nil {
		return nil, err
	}

	s := &Server{
		cfg:     cfg,
		i18n:    bundle,
		store:   store,
		fx:      rates,
		monero:  wallet,
		quoter:  monero.NewQuoter(),
		engines: engines,
		static:  http.StripPrefix("/static/", http.FileServer(http.FS(staticRoot))),
	}
	if cfg.StripeEnabled() {
		s.stripe = stripepay.New(cfg.StripeSecretKey, cfg.StripeWebhookSecret)
	}
	s.handler = s.buildMux()
	return s, nil
}

func (s *Server) Handler() http.Handler { return s.handler }

func (s *Server) buildMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+constants.PathHealthz, s.healthz)
	mux.HandleFunc("GET "+constants.PathRobots, s.robots)
	mux.Handle("GET "+constants.PathStatic, s.static)
	mux.HandleFunc("GET /{$}", s.rootRedirect)
	mux.HandleFunc("POST /webhooks/stripe", s.stripeWebhook)

	if s.cfg.AdminEnabled() {
		mux.HandleFunc("GET /internal/orders", s.basicAuth(s.adminOrders))
		mux.HandleFunc("POST /internal/ship", s.basicAuth(s.adminShip))
		mux.HandleFunc("POST /internal/stock", s.basicAuth(s.adminStock))
	}

	for _, lang := range s.i18n.SupportedCodes() {
		locale := lang
		mux.HandleFunc("GET /"+locale+"/{$}", func(w http.ResponseWriter, r *http.Request) {
			s.catalog(w, r, locale)
		})
		mux.HandleFunc("GET /"+locale+"/rss.xml", func(w http.ResponseWriter, r *http.Request) {
			s.catalogRSS(w, r, locale, "", "")
		})
		mux.HandleFunc("GET /"+locale+"/search", func(w http.ResponseWriter, r *http.Request) {
			s.search(w, r, locale)
		})
		mux.HandleFunc("GET /"+locale+"/category/{category}", func(w http.ResponseWriter, r *http.Request) {
			s.category(w, r, locale)
		})
		mux.HandleFunc("GET /"+locale+"/category/{category}/rss.xml", func(w http.ResponseWriter, r *http.Request) {
			s.catalogRSS(w, r, locale, "category", r.PathValue("category"))
		})
		mux.HandleFunc("GET /"+locale+"/tag/{tag}", func(w http.ResponseWriter, r *http.Request) {
			s.tag(w, r, locale)
		})
		mux.HandleFunc("GET /"+locale+"/tag/{tag}/rss.xml", func(w http.ResponseWriter, r *http.Request) {
			s.catalogRSS(w, r, locale, "tag", r.PathValue("tag"))
		})
		mux.HandleFunc("GET /"+locale+"/order/{id}", func(w http.ResponseWriter, r *http.Request) {
			s.orderStatus(w, r, locale)
		})
		mux.HandleFunc("POST /"+locale+"/checkout/stripe", func(w http.ResponseWriter, r *http.Request) {
			s.checkoutStripe(w, r, locale)
		})
		mux.HandleFunc("POST /"+locale+"/checkout/monero", func(w http.ResponseWriter, r *http.Request) {
			s.checkoutMonero(w, r, locale)
		})
		mux.HandleFunc("GET /"+locale+"/{slug}", func(w http.ResponseWriter, r *http.Request) {
			s.product(w, r, locale)
		})
	}

	mux.HandleFunc("GET /{path...}", s.notFound)
	mux.HandleFunc("/", s.methodNotAllowed)

	h := s.security(mux)
	return recoverMiddleware(s, h)
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(constants.HeaderContentType, "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(constants.HealthzResponse))
}

func (s *Server) robots(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(constants.HeaderContentType, "text/plain; charset=utf-8")
	_, _ = fmt.Fprintf(w, "User-agent: *\nAllow: /\n")
}

func (s *Server) rootRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/"+constants.DefaultLocale+"/", http.StatusFound)
}

func (s *Server) currency(r *http.Request) string {
	if q := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("currency"))); q != "" {
		for _, c := range s.fx.Supported() {
			if c == q {
				return q
			}
		}
	}
	if c, err := r.Cookie(constants.CurrencyCookie); err == nil {
		v := strings.ToUpper(c.Value)
		for _, x := range s.fx.Supported() {
			if x == v {
				return v
			}
		}
	}
	return s.cfg.BaseCurrency
}

func (s *Server) setCurrencyCookie(w http.ResponseWriter, currency string) {
	http.SetCookie(w, &http.Cookie{
		Name:     constants.CurrencyCookie,
		Value:    currency,
		Path:     "/",
		MaxAge:   365 * 24 * 3600,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: false,
	})
}

func (s *Server) locales(current, pagePath string) []LocaleOption {
	codes := s.i18n.SupportedCodes()
	out := make([]LocaleOption, len(codes))
	for i, code := range codes {
		out[i] = LocaleOption{
			Code:    code,
			Label:   s.i18n.T(current, "lang."+code),
			URL:     s.i18n.LocalPath(code, pagePath),
			Current: code == current,
		}
	}
	return out
}

func (s *Server) catalog(w http.ResponseWriter, r *http.Request, lang string) {
	s.renderCatalog(w, r, lang, catalogPageOpts{
		title:      s.i18n.T(lang, "catalog.title"),
		lead:       s.i18n.T(lang, "catalog.lead"),
		pagePath:   "",
		feedPath:   "/" + lang + "/rss.xml",
		showSearch: true,
		listAll:    true,
	})
}

func (s *Server) category(w http.ResponseWriter, r *http.Request, lang string) {
	cat := r.PathValue("category")
	if !validTaxonomy(cat) {
		s.writeError(w, r, lang, http.StatusNotFound)
		return
	}
	s.renderCatalog(w, r, lang, catalogPageOpts{
		title:    cat,
		lead:     s.i18n.T(lang, "catalog.category_lead"),
		pagePath: "category/" + cat,
		feedPath: "/" + lang + "/category/" + cat + "/rss.xml",
		category: cat,
	})
}

func (s *Server) tag(w http.ResponseWriter, r *http.Request, lang string) {
	tag := r.PathValue("tag")
	if !validTaxonomy(tag) {
		s.writeError(w, r, lang, http.StatusNotFound)
		return
	}
	s.renderCatalog(w, r, lang, catalogPageOpts{
		title:    "#" + tag,
		lead:     s.i18n.T(lang, "catalog.tag_lead"),
		pagePath: "tag/" + tag,
		feedPath: "/" + lang + "/tag/" + tag + "/rss.xml",
		tag:      tag,
	})
}

type catalogPageOpts struct {
	title      string
	lead       string
	pagePath   string
	feedPath   string
	category   string
	tag        string
	showSearch bool
	listAll    bool
}

func (s *Server) renderCatalog(w http.ResponseWriter, r *http.Request, lang string, opts catalogPageOpts) {
	currency := s.currency(r)
	s.setCurrencyCookie(w, currency)

	var products []catalog.Product
	var err error
	switch {
	case opts.category != "":
		products, err = s.store.ListByCategory(opts.category)
	case opts.tag != "":
		products, err = s.store.ListByTag(opts.tag)
	default:
		products, err = s.store.ListProducts()
	}
	if err != nil {
		s.writeError(w, r, lang, http.StatusInternalServerError)
		return
	}

	all, _ := s.store.ListProducts()
	views := make([]ProductView, 0, len(products))
	for _, p := range products {
		views = append(views, s.productView(lang, p, currency))
	}

	view := PageView{
		Lang:                   lang,
		Title:                  opts.title + " | " + s.i18n.T(lang, "site.name"),
		Description:            opts.lead,
		Currency:               currency,
		Currencies:             s.fx.Supported(),
		CatalogTitle:           opts.title,
		CatalogLead:            opts.lead,
		Products:               views,
		Categories:             catalog.CollectCategories(all),
		Tags:                   catalog.CollectTags(all),
		ActiveCategory:         opts.category,
		ActiveTag:              opts.tag,
		FeedPath:               opts.feedPath,
		ShowCatalogSearch:      opts.showSearch,
		IncludeCatalogSearchJS: opts.showSearch && len(views) > 0,
		Locales:                s.locales(lang, opts.pagePath),
	}
	if len(views) == 0 {
		view.EmptyState = &EmptyStateView{
			Title: s.i18n.T(lang, "catalog.empty.title"),
			Body:  s.i18n.T(lang, "catalog.empty.body"),
		}
	}
	s.render(w, r, "catalog-content", view)
}

func (s *Server) search(w http.ResponseWriter, r *http.Request, lang string) {
	currency := s.currency(r)
	s.setCurrencyCookie(w, currency)
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	var products []catalog.Product
	var err error
	if query != "" {
		products, err = s.store.SearchProducts(query)
		if err != nil {
			s.writeError(w, r, lang, http.StatusInternalServerError)
			return
		}
	}
	views := make([]ProductView, 0, len(products))
	for _, p := range products {
		views = append(views, s.productView(lang, p, currency))
	}
	view := PageView{
		Lang:         lang,
		Title:        s.i18n.T(lang, "search.title") + " | " + s.i18n.T(lang, "site.name"),
		Description:  s.i18n.T(lang, "search.lead"),
		Currency:     currency,
		Currencies:   s.fx.Supported(),
		Query:        query,
		Products:     views,
		Locales:      s.locales(lang, "search"),
		ResultsCount: s.i18n.TWith(lang, "search.results", map[string]any{"Count": len(views)}),
	}
	if query != "" && len(views) == 0 {
		view.EmptyState = &EmptyStateView{
			Title: s.i18n.T(lang, "search.no_results"),
			Body:  s.i18n.T(lang, "search.no_results_lead"),
		}
	}
	s.render(w, r, "search-content", view)
}

func (s *Server) catalogRSS(w http.ResponseWriter, r *http.Request, lang, kind, value string) {
	var products []catalog.Product
	var err error
	var title, desc, pagePath string
	switch kind {
	case "category":
		if !validTaxonomy(value) {
			http.NotFound(w, r)
			return
		}
		products, err = s.store.ListByCategory(value)
		title = value + " | " + s.i18n.T(lang, "site.name")
		desc = s.i18n.T(lang, "catalog.category_lead")
		pagePath = "/" + lang + "/category/" + value
	case "tag":
		if !validTaxonomy(value) {
			http.NotFound(w, r)
			return
		}
		products, err = s.store.ListByTag(value)
		title = "#" + value + " | " + s.i18n.T(lang, "site.name")
		desc = s.i18n.T(lang, "catalog.tag_lead")
		pagePath = "/" + lang + "/tag/" + value
	default:
		products, err = s.store.ListProducts()
		title = s.i18n.T(lang, "catalog.title") + " | " + s.i18n.T(lang, "site.name")
		desc = s.i18n.T(lang, "catalog.lead")
		pagePath = "/" + lang + "/"
	}
	if err != nil {
		http.Error(w, "feed error", http.StatusInternalServerError)
		return
	}
	data, err := catalog.RSS(s.cfg.SiteURL, lang, title, desc, pagePath, products)
	if err != nil {
		http.Error(w, "feed error", http.StatusInternalServerError)
		return
	}
	w.Header().Set(constants.HeaderContentType, "application/rss+xml; charset=utf-8")
	w.Header().Set(constants.HeaderCacheControl, constants.FeedCacheControl)
	_, _ = w.Write(append(append([]byte(nil), xml.Header...), data...))
}

func (s *Server) product(w http.ResponseWriter, r *http.Request, lang string) {
	slug := r.PathValue("slug")
	if !validSlug(slug) {
		s.writeError(w, r, lang, http.StatusNotFound)
		return
	}
	currency := s.currency(r)
	s.setCurrencyCookie(w, currency)
	p, ok, err := s.store.GetProduct(slug)
	if err != nil {
		s.writeError(w, r, lang, http.StatusInternalServerError)
		return
	}
	if !ok {
		s.writeError(w, r, lang, http.StatusNotFound)
		return
	}
	pv := s.productView(lang, p, currency)
	view := PageView{
		Lang:             lang,
		Title:            p.Title + " | " + s.i18n.T(lang, "site.name"),
		Description:      p.Description,
		Currency:         currency,
		Currencies:       s.fx.Supported(),
		Product:          p,
		ProductTypeLabel: s.i18n.T(lang, "product.type."+p.Type),
		DisplayPrice:     pv.DisplayPrice,
		ReferencePrice:   pv.ReferencePrice,
		StockLabel:       pv.StockLabel,
		OutOfStock:       pv.OutOfStock,
		StripeEnabled:    s.cfg.StripeEnabled(),
		MoneroEnabled:    s.cfg.MoneroEnabled() && s.monero != nil,
		Locales:          s.locales(lang, slug),
	}
	if pv.OutOfStock {
		view.EmptyState = &EmptyStateView{
			Title: s.i18n.T(lang, "product.out_of_stock"),
			Body:  s.i18n.T(lang, "catalog.empty.body"),
		}
	}
	s.render(w, r, "product-content", view)
}

func (s *Server) productView(lang string, p catalog.Product, displayCurrency string) ProductView {
	outOfStock := p.Stock != nil && p.Available < 1
	display := catalog.FormatMoney(p.PriceCents, p.Currency)
	ref := ""
	if displayCurrency != "" && !strings.EqualFold(displayCurrency, p.Currency) {
		ref = s.i18n.TWith(lang, "product.reference_price", map[string]any{
			"Price": s.fx.FormatConverted(p.PriceCents, p.Currency, displayCurrency),
		})
	}
	stockLabel := s.i18n.T(lang, "product.unlimited")
	stockValue := ""
	if p.Stock != nil {
		stockLabel = fmt.Sprintf("%d", p.Available)
		stockValue = strconv.Itoa(*p.Stock)
	}
	return ProductView{
		Product:        p,
		DisplayPrice:   display,
		ReferencePrice: ref,
		OutOfStock:     outOfStock,
		StockLabel:     stockLabel,
		StockValue:     stockValue,
	}
}

func (s *Server) checkoutStripe(w http.ResponseWriter, r *http.Request, lang string) {
	if s.stripe == nil {
		s.writeError(w, r, lang, http.StatusServiceUnavailable)
		return
	}
	slug := r.FormValue("slug")
	p, ok, err := s.store.GetProduct(slug)
	if err != nil || !ok {
		s.writeError(w, r, lang, http.StatusNotFound)
		return
	}
	if p.Stock != nil && p.Available < 1 {
		http.Redirect(w, r, "/"+lang+"/"+slug, http.StatusSeeOther)
		return
	}
	order, err := s.store.CreateOrder(lang, constants.PaymentStripe, p, 1)
	if err != nil {
		s.writeError(w, r, lang, http.StatusConflict)
		return
	}
	success := s.cfg.SiteURL + "/" + lang + "/order/" + order.ID
	cancel := s.cfg.SiteURL + "/" + lang + "/" + slug
	sessionID, url, err := s.stripe.CreateCheckout(order, p, success, cancel)
	if err != nil {
		log.Printf("stripe checkout: %v", err)
		s.writeError(w, r, lang, http.StatusBadGateway)
		return
	}
	_ = s.store.SetStripeSession(order.ID, sessionID)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (s *Server) checkoutMonero(w http.ResponseWriter, r *http.Request, lang string) {
	if s.monero == nil {
		s.writeError(w, r, lang, http.StatusServiceUnavailable)
		return
	}
	slug := r.FormValue("slug")
	p, ok, err := s.store.GetProduct(slug)
	if err != nil || !ok {
		s.writeError(w, r, lang, http.StatusNotFound)
		return
	}
	if p.Stock != nil && p.Available < 1 {
		http.Redirect(w, r, "/"+lang+"/"+slug, http.StatusSeeOther)
		return
	}
	order, err := s.store.CreateOrder(lang, constants.PaymentMonero, p, 1)
	if err != nil {
		s.writeError(w, r, lang, http.StatusConflict)
		return
	}
	addr, idx, err := s.monero.CreateAddress("order-" + order.ID)
	if err != nil {
		log.Printf("monero address: %v", err)
		s.writeError(w, r, lang, http.StatusBadGateway)
		return
	}
	atomic, err := s.quoter.QuoteAtomic(order.TotalCents, strings.ToLower(p.Currency))
	if err != nil {
		log.Printf("monero quote: %v", err)
		s.writeError(w, r, lang, http.StatusBadGateway)
		return
	}
	_ = s.store.SetMoneroInvoice(order.ID, addr, idx, atomic)
	http.Redirect(w, r, "/"+lang+"/order/"+order.ID, http.StatusSeeOther)
}

func (s *Server) orderStatus(w http.ResponseWriter, r *http.Request, lang string) {
	id := r.PathValue("id")
	order, err := s.store.GetOrder(id)
	if err != nil {
		s.writeError(w, r, lang, http.StatusNotFound)
		return
	}
	currency := s.currency(r)
	view := PageView{
		Lang:        lang,
		Title:       s.i18n.T(lang, "order.title") + " | " + s.i18n.T(lang, "site.name"),
		Description: s.i18n.T(lang, "order.title"),
		Currency:    currency,
		Currencies:  s.fx.Supported(),
		Order:       order,
		StatusLabel: s.orderStatusLabel(lang, order.Status),
		Locales:     s.locales(lang, "order/"+id),
	}
	if order.Status == constants.OrderAwaitingMonero && order.MoneroAddress != "" {
		view.MoneroAddress = order.MoneroAddress
		view.MoneroAmount = monero.FormatXMR(order.MoneroAmountAtomic)
		view.RefreshSeconds = 15
		if order.ExpiresAt != nil {
			view.ExpiresAt = order.ExpiresAt.Format(time.RFC3339)
		}
		uri := "monero:" + order.MoneroAddress + "?tx_amount=" + view.MoneroAmount
		if png, err := qrcode.Encode(uri, qrcode.Medium, 220); err == nil {
			view.QRDataURI = template.URL("data:image/png;base64," + encodeBase64(png))
		}
	}
	if order.Status == constants.OrderPaid {
		var links []string
		for _, it := range order.Items {
			if it.ProductType == constants.ProductDigital && it.DigitalPayload != "" {
				links = append(links, it.DigitalPayload)
			}
		}
		if len(links) > 0 {
			view.DigitalReady = true
			view.DigitalLinks = links
		}
	}
	s.render(w, r, "order-content", view)
}

func (s *Server) orderStatusLabel(lang, status string) string {
	switch status {
	case constants.OrderPaid:
		return s.i18n.T(lang, "order.paid")
	case constants.OrderShipped:
		return s.i18n.T(lang, "order.shipped")
	case constants.OrderCancelled:
		return s.i18n.T(lang, "order.cancelled")
	case constants.OrderExpired:
		return s.i18n.T(lang, "order.expired")
	case constants.OrderAwaitingShip:
		return s.i18n.T(lang, "order.awaiting_fulfillment")
	default:
		return s.i18n.T(lang, "order.pending")
	}
}

func (s *Server) stripeWebhook(w http.ResponseWriter, r *http.Request) {
	if s.stripe == nil {
		http.Error(w, "stripe disabled", http.StatusServiceUnavailable)
		return
	}
	payload, err := stripepay.ReadBody(r, 1<<20)
	if err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}
	orderID, err := s.stripe.HandleWebhook(payload, r.Header.Get("Stripe-Signature"))
	if err != nil {
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}
	if orderID != "" {
		if err := s.store.MarkOrderPaid(orderID, constants.PaymentStripe, orderID); err != nil {
			log.Printf("mark paid: %v", err)
			http.Error(w, "store error", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) adminOrders(w http.ResponseWriter, r *http.Request) {
	lang := constants.DefaultLocale
	products, err := s.store.ListProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	orders, err := s.store.ListOrders()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	views := make([]ProductView, 0, len(products))
	for _, p := range products {
		views = append(views, s.productView(lang, p, s.cfg.BaseCurrency))
	}
	s.render(w, r, "admin-content", PageView{
		Lang:        lang,
		Title:       s.i18n.T(lang, "admin.title"),
		Description: s.i18n.T(lang, "admin.title"),
		Currency:    s.cfg.BaseCurrency,
		Currencies:  s.fx.Supported(),
		Products:    views,
		Orders:      orders,
		Locales:     s.locales(lang, ""),
		MainSiteURL: s.cfg.MainSiteURL,
	})
}

func (s *Server) adminShip(w http.ResponseWriter, r *http.Request) {
	_ = s.store.MarkShipped(r.FormValue("order_id"))
	http.Redirect(w, r, "/internal/orders", http.StatusSeeOther)
}

func (s *Server) adminStock(w http.ResponseWriter, r *http.Request) {
	slug := r.FormValue("slug")
	raw := strings.TrimSpace(r.FormValue("stock"))
	if raw == "" {
		_ = s.store.SetStock(slug, nil)
	} else if n, err := strconv.Atoi(raw); err == nil {
		_ = s.store.SetStock(slug, &n)
	}
	http.Redirect(w, r, "/internal/orders", http.StatusSeeOther)
}

func (s *Server) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(user), []byte(s.cfg.AdminUser)) != 1 ||
			subtle.ConstantTimeCompare([]byte(pass), []byte(s.cfg.AdminPassword)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="store-admin"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (s *Server) render(w http.ResponseWriter, r *http.Request, content string, view PageView) {
	view.AssetVersion = assetVersion
	view.MainSiteURL = s.cfg.MainSiteURL
	if view.Currencies == nil {
		view.Currencies = s.fx.Supported()
	}
	engine := s.engines[view.Lang]
	if engine == nil {
		engine = s.engines[constants.DefaultLocale]
	}
	w.Header().Set(constants.HeaderContentType, "text/html; charset=utf-8")
	w.Header().Set(constants.HeaderCacheControl, constants.PageCacheControl)
	err := engine.RenderPage(w, content, &view, func(b []byte) {
		view.Body = template.HTML(b)
	})
	if err != nil {
		log.Printf("render %s: %v", content, err)
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

func (s *Server) writeError(w http.ResponseWriter, r *http.Request, lang string, status int) {
	title := s.i18n.T(lang, "error.server")
	lead := s.i18n.T(lang, "error.server")
	switch status {
	case http.StatusNotFound:
		title = s.i18n.T(lang, "error.not_found")
		lead = s.i18n.T(lang, "error.not_found_lead")
	case http.StatusMethodNotAllowed:
		title = s.i18n.T(lang, "error.method")
		lead = s.i18n.T(lang, "error.method")
	}
	w.WriteHeader(status)
	s.render(w, r, "error-content", PageView{
		Lang:        lang,
		Title:       title,
		Description: lead,
		Currency:    s.cfg.BaseCurrency,
		Currencies:  s.fx.Supported(),
		Locales:     s.locales(lang, ""),
		ErrorStatus: status,
		ErrorTitle:  title,
		ErrorLead:   lead,
	})
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	s.writeError(w, r, constants.DefaultLocale, http.StatusNotFound)
}

func (s *Server) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		s.notFound(w, r)
		return
	}
	s.writeError(w, r, constants.DefaultLocale, http.StatusMethodNotAllowed)
}

func (s *Server) security(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(constants.HeaderCSP, constants.CSPPolicy)
		w.Header().Set(constants.HeaderXContentType, constants.Nosniff)
		w.Header().Set(constants.HeaderReferrerPolicy, constants.Referrer)
		w.Header().Set(constants.HeaderXFrameOptions, constants.DenyFrame)
		w.Header().Set(constants.HeaderPermissionsPolicy, constants.Perms)
		if s.cfg.Production() {
			w.Header().Set(constants.HeaderStrictTransport, constants.HSTS)
		}
		next.ServeHTTP(w, r)
	})
}

func recoverMiddleware(s *Server, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				s.writeError(w, r, constants.DefaultLocale, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func validSlug(slug string) bool {
	if slug == "" || len(slug) > 80 {
		return false
	}
	if reservedSlug(slug) {
		return false
	}
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			continue
		}
		return false
	}
	return true
}

func validTaxonomy(value string) bool {
	if value == "" || len(value) > 80 {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			continue
		}
		return false
	}
	return true
}

func reservedSlug(slug string) bool {
	switch slug {
	case "search", "rss.xml", "tag", "category", "order", "checkout", "internal":
		return true
	default:
		return false
	}
}

func encodeBase64(b []byte) string {
	const table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	out := make([]byte, 0, (len(b)+2)/3*4)
	for i := 0; i < len(b); i += 3 {
		var n uint32
		remain := len(b) - i
		switch {
		case remain >= 3:
			n = uint32(b[i])<<16 | uint32(b[i+1])<<8 | uint32(b[i+2])
			out = append(out, table[n>>18&63], table[n>>12&63], table[n>>6&63], table[n&63])
		case remain == 2:
			n = uint32(b[i])<<16 | uint32(b[i+1])<<8
			out = append(out, table[n>>18&63], table[n>>12&63], table[n>>6&63], '=')
		default:
			n = uint32(b[i]) << 16
			out = append(out, table[n>>18&63], table[n>>12&63], '=', '=')
		}
	}
	return string(out)
}

func ListenAndServe(cfg config.Config, handler http.Handler) error {
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: constants.ReadHeaderTimeout,
		ReadTimeout:       constants.ReadTimeout,
		WriteTimeout:      constants.WriteTimeout,
		IdleTimeout:       constants.IdleTimeout,
		MaxHeaderBytes:    constants.MaxHeaderBytes,
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errCh:
		return err
	case <-sigCh:
		ctx, cancel := context.WithTimeout(context.Background(), constants.ShutdownTimeout)
		defer cancel()
		return srv.Shutdown(ctx)
	}
}
