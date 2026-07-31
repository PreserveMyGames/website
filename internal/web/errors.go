package web

import (
	"html/template"
	"log"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/PreserveMyGames/website/internal/constants"
	"github.com/PreserveMyGames/website/internal/seo"
)

func recoverMiddleware(s *Server, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v\n%s", err, debug.Stack())
				s.writeError(w, r, http.StatusInternalServerError, s.langFromRequest(r), "error.server")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) unmatched(w http.ResponseWriter, r *http.Request) {
	lang := s.langFromPath(strings.Trim(r.PathValue("path"), "/"), r)
	s.writeNotFound(w, r, lang)
}

func (s *Server) writeNotFound(w http.ResponseWriter, r *http.Request, lang string) {
	if wantsPlainError(r) {
		http.NotFound(w, r)
		return
	}
	s.writeError(w, r, http.StatusNotFound, lang, "error.not_found")
}

func (s *Server) writeMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	if wantsPlainError(r) {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	s.writeError(w, r, http.StatusMethodNotAllowed, s.langFromRequest(r), "error.method")
}

func (s *Server) writeError(w http.ResponseWriter, r *http.Request, status int, lang, keyPrefix string) {
	if wantsPlainError(r) {
		http.Error(w, http.StatusText(status), status)
		return
	}

	if !s.i18n.Allowed(lang) {
		codes := s.i18n.SupportedCodes()
		if len(codes) > 0 {
			lang = codes[0]
		}
	}

	title := s.i18n.T(lang, keyPrefix+".title")
	lead := s.i18n.T(lang, keyPrefix+".lead")
	pageTitle := title + " | " + s.i18n.T(lang, "site.name")

	view := PageView{
		Lang:         lang,
		Meta:         seo.Page(s.cfg.SiteURL, lang, "", pageTitle, lead, s.cachedAlternates("")),
		PagePath:     "",
		AssetVersion: assetVersion,
		ErrorStatus:  status,
		ErrorTitle:   title,
		ErrorLead:    lead,
	}

	engine := s.engines[lang]
	if engine == nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	view.Locales = s.cachedLocales(lang, view.PagePath)
	s.populateLayout(lang, &view)
	w.Header().Set(constants.HeaderContentType, "text/html; charset=utf-8")
	w.Header().Set(constants.HeaderCacheControl, constants.PageCacheControl)
	w.WriteHeader(status)

	if err := engine.RenderPage(w, "error-content", "layout", &view, func(content []byte) {
		view.Body = template.HTML(content)
	}); err != nil {
		log.Printf("error page render: %v", err)
	}
}

func (s *Server) langFromRequest(r *http.Request) string {
	return s.langFromPath(strings.TrimPrefix(r.URL.Path, "/"), r)
}

func (s *Server) langFromPath(path string, r *http.Request) string {
	path = strings.Trim(path, "/")
	if path == "" {
		return s.preferredLang(r)
	}
	segment := path
	if before, _, ok := strings.Cut(path, "/"); ok {
		segment = before
	}
	if s.i18n.Allowed(segment) {
		return segment
	}
	return s.preferredLang(r)
}

func wantsPlainError(r *http.Request) bool {
	path := r.URL.Path
	if strings.HasPrefix(path, constants.PathStatic) ||
		path == constants.PathHealthz ||
		path == constants.PathRobots ||
		path == constants.PathSitemap {
		return true
	}
	if strings.HasSuffix(path, ".json") || strings.HasSuffix(path, ".xml") {
		return true
	}
	accept := r.Header.Get("Accept")
	if accept == "" {
		return false
	}
	if strings.Contains(accept, "text/html") {
		return false
	}
	if strings.Contains(accept, "application/json") || strings.Contains(accept, "application/xml") {
		return true
	}
	return strings.Contains(accept, "*/*") && !strings.Contains(accept, "text/html")
}
