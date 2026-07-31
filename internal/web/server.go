package web

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PreserveMyGames/website/internal/blog"
	"github.com/PreserveMyGames/website/internal/config"
	"github.com/PreserveMyGames/website/internal/constants"
	"github.com/PreserveMyGames/website/internal/i18n"
	"github.com/PreserveMyGames/website/internal/render"
	"github.com/PreserveMyGames/website/internal/seo"
	"github.com/PreserveMyGames/website/internal/staticfiles"
	"github.com/PreserveMyGames/website/internal/validate"
)

const assetVersion = constants.AssetVersion

type Server struct {
	cfg     config.Config
	i18n    *i18n.Bundle
	blog    *blog.Index
	engines map[string]*render.Engine
	static  http.Handler
	cache   *precomputed
	handler http.Handler
}

type SiteNoticeView struct {
	Kind    string
	Message string
}

type PageView struct {
	Lang                string
	Meta                seo.Meta
	ContentTemplate     string
	PagePath            string
	Locales             []LocaleOption
	Body                template.HTML
	AssetVersion        string
	WikiURL             string
	ForumsURL           string
	Notice              *SiteNoticeView
	ContactEmail        string
	ContactBody         string
	ContactLXMF         string
	IncludeSearchJS     bool
	IncludeBlogSearchJS bool
	Posts               []blog.Post
	Post                blogPostView
	Query               string
	Results             []blog.SearchEntry
	ErrorStatus         int
	ErrorTitle          string
	ErrorLead           string
}

type LocaleOption struct {
	Code    string
	Label   string
	URL     string
	Current bool
}

type blogPostView struct {
	Title     string
	Author    string
	AuthorURL string
	Date      time.Time
	Updated   time.Time
	BodyHTML  template.HTML
}

func New(cfg config.Config, bundle *i18n.Bundle, index *blog.Index) (*Server, error) {
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
		blog:    index,
		engines: engines,
		static:  staticHandler(staticRoot),
	}
	if err := s.warmCaches(); err != nil {
		return nil, err
	}
	s.handler = s.buildMux()
	return s, nil
}

func (s *Server) Handler() http.Handler {
	return s.handler
}

func (s *Server) buildMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+constants.PathHealthz, s.healthz)
	mux.HandleFunc("GET "+constants.PathRobots, s.robots)
	mux.HandleFunc("GET "+constants.PathSitemap, s.sitemap)
	mux.Handle("GET "+constants.PathStatic, s.static)
	mux.HandleFunc("GET /{$}", s.rootRedirect)

	for _, lang := range s.i18n.SupportedCodes() {
		locale := lang
		mux.HandleFunc("GET /"+locale+"/{$}", func(w http.ResponseWriter, r *http.Request) {
			s.home(w, r, locale)
		})
		mux.HandleFunc("GET /"+locale+"/about", func(w http.ResponseWriter, r *http.Request) {
			s.about(w, r, locale)
		})
		mux.HandleFunc("GET /"+locale+"/privacy", func(w http.ResponseWriter, r *http.Request) {
			s.privacy(w, r, locale)
		})
		mux.HandleFunc("GET /"+locale+"/contact", func(w http.ResponseWriter, r *http.Request) {
			s.contact(w, r, locale)
		})
		mux.HandleFunc("GET /"+locale+"/blog", func(w http.ResponseWriter, r *http.Request) {
			s.blogIndex(w, r, locale)
		})
		mux.HandleFunc("GET /"+locale+"/blog/rss.xml", func(w http.ResponseWriter, r *http.Request) {
			s.blogRSS(w, r, locale)
		})
		mux.HandleFunc("GET /"+locale+"/blog/atom.xml", func(w http.ResponseWriter, r *http.Request) {
			s.blogAtom(w, r, locale)
		})
		mux.HandleFunc("GET /"+locale+"/blog/{slug}", func(w http.ResponseWriter, r *http.Request) {
			s.blogPost(w, r, locale)
		})
		mux.HandleFunc("GET /"+locale+"/search", func(w http.ResponseWriter, r *http.Request) {
			s.search(w, r, locale)
		})
		mux.HandleFunc("GET /"+locale+"/search-index.json", func(w http.ResponseWriter, r *http.Request) {
			s.searchIndex(w, r, locale)
		})
	}

	mux.HandleFunc("GET /{path...}", s.unmatched)

	h := accessLogMiddleware(s.cfg, mux)
	h = securityMiddleware(s.cfg, s, h)
	return recoverMiddleware(s, h)
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(constants.HeaderContentType, "text/plain; charset=utf-8")
	w.Header().Set(constants.HeaderContentLength, "2")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(constants.HealthzResponse))
}

func (s *Server) rootRedirect(w http.ResponseWriter, r *http.Request) {
	codes := s.i18n.SupportedCodes()
	preferred := s.preferredLang(r)
	idx := 0
	for i, code := range codes {
		if code == preferred {
			idx = i
			break
		}
	}
	http.Redirect(w, r, "/"+codes[idx]+"/", http.StatusFound)
}

func (s *Server) home(w http.ResponseWriter, r *http.Request, lang string) {
	title := s.i18n.T(lang, "site.name") + " | " + s.i18n.T(lang, "home.headline")
	desc := s.i18n.T(lang, "home.lead")
	s.renderPage(w, r, "home-content", PageView{
		Lang:         lang,
		Meta:         seo.Home(s.cfg.SiteURL, lang, title, desc, s.cachedAlternates("")),
		PagePath:     "",
		AssetVersion: assetVersion,
	})
}

func (s *Server) about(w http.ResponseWriter, r *http.Request, lang string) {
	title := s.i18n.T(lang, "about.title") + " | " + s.i18n.T(lang, "site.name")
	desc := s.i18n.T(lang, "about.body")
	s.renderPage(w, r, "about-content", PageView{
		Lang:         lang,
		Meta:         seo.Page(s.cfg.SiteURL, lang, "about", title, desc, s.cachedAlternates("about")),
		PagePath:     "about",
		AssetVersion: assetVersion,
	})
}

func (s *Server) privacy(w http.ResponseWriter, r *http.Request, lang string) {
	title := s.i18n.T(lang, "privacy.title") + " | " + s.i18n.T(lang, "site.name")
	desc := s.i18n.T(lang, "privacy.body")
	s.renderPage(w, r, "privacy-content", PageView{
		Lang:         lang,
		Meta:         seo.Page(s.cfg.SiteURL, lang, "privacy", title, desc, s.cachedAlternates("privacy")),
		PagePath:     "privacy",
		AssetVersion: assetVersion,
	})
}

func (s *Server) contact(w http.ResponseWriter, r *http.Request, lang string) {
	title := s.i18n.T(lang, "contact.title") + " | " + s.i18n.T(lang, "site.name")
	desc := s.i18n.TWith(lang, "contact.body", map[string]any{"Email": s.cfg.ContactEmail})
	s.renderPage(w, r, "contact-content", PageView{
		Lang:         lang,
		Meta:         seo.Page(s.cfg.SiteURL, lang, "contact", title, desc, s.cachedAlternates("contact")),
		PagePath:     "contact",
		AssetVersion: assetVersion,
		ContactEmail: s.cfg.ContactEmail,
		ContactBody:  desc,
		ContactLXMF:  constants.ContactLXMF,
	})
}

func (s *Server) populateLayout(lang string, view *PageView) {
	view.WikiURL = s.cfg.WikiURL()
	view.ForumsURL = s.cfg.ForumsURL()
	view.Notice = s.siteNotice(lang)
	view.Locales = s.cachedLocales(lang, view.PagePath)
	view.AssetVersion = assetVersion
}

func (s *Server) siteNotice(lang string) *SiteNoticeView {
	kind := s.cfg.NoticeKind()
	if kind == "" {
		return nil
	}
	msg := s.cfg.SiteNoticeMessage
	if msg == "" {
		msg = s.i18n.T(lang, "notice."+kind)
	}
	return &SiteNoticeView{Kind: kind, Message: msg}
}

func (s *Server) blogIndex(w http.ResponseWriter, r *http.Request, lang string) {
	title := s.i18n.T(lang, "blog.title") + " | " + s.i18n.T(lang, "site.name")
	desc := s.i18n.T(lang, "site.tagline")
	s.renderPage(w, r, "blog_index-content", PageView{
		Lang:                lang,
		Meta:                seo.Page(s.cfg.SiteURL, lang, "blog", title, desc, s.cachedAlternates("blog")),
		PagePath:            "blog",
		AssetVersion:        assetVersion,
		IncludeBlogSearchJS: true,
		Query:               strings.TrimSpace(r.URL.Query().Get("q")),
		Posts:               s.blog.Posts(lang),
	})
}

func (s *Server) blogPost(w http.ResponseWriter, r *http.Request, lang string) {
	slug := r.PathValue("slug")
	if !validate.Slug(slug) {
		s.writeNotFound(w, r, lang)
		return
	}
	post, ok := s.blog.Post(lang, slug)
	if !ok {
		s.writeNotFound(w, r, lang)
		return
	}
	pagePath := "blog/" + slug
	title := post.Title + " | " + s.i18n.T(lang, "site.name")
	s.renderPage(w, r, "blog_post-content", PageView{
		Lang:         lang,
		Meta:         seo.BlogPost(s.cfg.SiteURL, lang, slug, title, post.Description, post.Date, s.cachedAlternates(pagePath)),
		PagePath:     pagePath,
		AssetVersion: assetVersion,
		Post: blogPostView{
			Title:     post.Title,
			Author:    post.Author,
			AuthorURL: post.AuthorURL,
			Date:      post.Date,
			Updated:   post.Updated,
			BodyHTML:  template.HTML(post.BodyHTML),
		},
	})
}

func (s *Server) blogRSS(w http.ResponseWriter, r *http.Request, lang string) {
	data := s.cache.rssXML[lang]
	if len(data) == 0 {
		s.writeNotFound(w, r, lang)
		return
	}
	w.Header().Set(constants.HeaderContentType, "application/rss+xml; charset=utf-8")
	w.Header().Set(constants.HeaderContentLength, fmt.Sprintf("%d", len(data)))
	w.Header().Set(constants.HeaderCacheControl, constants.SearchCacheControl)
	_, _ = w.Write(data)
}

func (s *Server) blogAtom(w http.ResponseWriter, r *http.Request, lang string) {
	data := s.cache.atomXML[lang]
	if len(data) == 0 {
		s.writeNotFound(w, r, lang)
		return
	}
	w.Header().Set(constants.HeaderContentType, "application/atom+xml; charset=utf-8")
	w.Header().Set(constants.HeaderContentLength, fmt.Sprintf("%d", len(data)))
	w.Header().Set(constants.HeaderCacheControl, constants.SearchCacheControl)
	_, _ = w.Write(data)
}

func (s *Server) search(w http.ResponseWriter, r *http.Request, lang string) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	title := s.i18n.T(lang, "search.title") + " | " + s.i18n.T(lang, "site.name")
	desc := s.i18n.T(lang, "search.placeholder")
	s.renderPage(w, r, "search-content", PageView{
		Lang:            lang,
		Meta:            seo.Page(s.cfg.SiteURL, lang, "search", title, desc, s.cachedAlternates("search")),
		PagePath:        "search",
		AssetVersion:    assetVersion,
		IncludeSearchJS: true,
		Query:           query,
		Results:         s.filterSearch(lang, query),
	})
}

func (s *Server) searchIndex(w http.ResponseWriter, r *http.Request, lang string) {
	data := s.cache.searchJSON[lang]
	if len(data) == 0 {
		s.writeNotFound(w, r, lang)
		return
	}
	w.Header().Set(constants.HeaderContentType, "application/json; charset=utf-8")
	w.Header().Set(constants.HeaderContentLength, fmt.Sprintf("%d", len(data)))
	w.Header().Set(constants.HeaderCacheControl, constants.SearchCacheControl)
	_, _ = w.Write(data)
}

func (s *Server) robots(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(constants.HeaderContentType, "text/plain; charset=utf-8")
	w.Header().Set(constants.HeaderContentLength, fmt.Sprintf("%d", len(s.cache.robotsTXT)))
	_, _ = w.Write(s.cache.robotsTXT)
}

func (s *Server) sitemap(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(constants.HeaderContentType, "application/xml; charset=utf-8")
	w.Header().Set(constants.HeaderContentLength, fmt.Sprintf("%d", len(s.cache.sitemapXML)))
	w.Header().Set(constants.HeaderCacheControl, constants.SearchCacheControl)
	_, _ = w.Write(s.cache.sitemapXML)
}

func (s *Server) renderPage(w http.ResponseWriter, r *http.Request, contentTemplate string, view PageView) {
	lang := view.Lang
	if !s.i18n.Allowed(lang) {
		s.writeNotFound(w, r, s.preferredLang(r))
		return
	}

	engine := s.engines[lang]
	if engine == nil {
		s.writeError(w, r, http.StatusInternalServerError, lang, "error.server")
		return
	}

	s.populateLayout(lang, &view)

	w.Header().Set(constants.HeaderContentType, "text/html; charset=utf-8")

	err := engine.RenderPage(w, contentTemplate, "layout", &view, func(content []byte) {
		view.Body = template.HTML(content)
	})
	if err != nil {
		log.Printf("render page: %v", err)
		s.writeError(w, r, http.StatusInternalServerError, lang, "error.server")
		return
	}
}

func (s *Server) preferredLang(r *http.Request) string {
	if c, err := r.Cookie(constants.LangCookieName); err == nil && s.i18n.Allowed(c.Value) {
		return c.Value
	}
	return s.i18n.Match(r.Header.Get("Accept-Language"))
}

func (s *Server) filterSearch(lang, query string) []blog.SearchEntry {
	if query == "" {
		return nil
	}
	q := strings.ToLower(query)
	entries := s.cache.searchFolded[lang]
	out := make([]blog.SearchEntry, 0, 4)
	for _, e := range entries {
		if strings.Contains(e.fold, q) {
			out = append(out, e.entry)
		}
	}
	return out
}

func staticHandler(root fs.FS) http.Handler {
	inner := http.FileServer(http.FS(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "..") {
			http.NotFound(w, r)
			return
		}
		p := strings.TrimPrefix(r.URL.Path, constants.PathStatic)
		if !validate.StaticPath(p) {
			http.NotFound(w, r)
			return
		}
		w.Header().Set(constants.HeaderCacheControl, constants.StaticCacheControl)
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/" + p
		inner.ServeHTTP(w, r2)
	})
}
