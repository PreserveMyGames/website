package web

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/PreserveMyGames/website/internal/config"
	"github.com/PreserveMyGames/website/internal/constants"
)

func securityMiddleware(cfg config.Config, s *Server, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			s.writeMethodNotAllowed(w, r)
			return
		}

		w.Header().Set(constants.HeaderCSP, constants.CSPPolicy)
		w.Header().Set(constants.HeaderXContentType, constants.Nosniff)
		w.Header().Set(constants.HeaderReferrerPolicy, constants.Referrer)
		w.Header().Set(constants.HeaderXFrameOptions, constants.DenyFrame)
		w.Header().Set(constants.HeaderPermissionsPolicy, constants.Perms)
		if cfg.Production() {
			w.Header().Set(constants.HeaderStrictTransport, constants.HSTS)
		}

		next.ServeHTTP(w, r)
	})
}

func accessLogMiddleware(cfg config.Config, next http.Handler) http.Handler {
	if !cfg.AccessLog {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func ListenAndServe(cfg config.Config, handler http.Handler) error {
	lc := net.ListenConfig{}
	ln, err := lc.Listen(context.Background(), "tcp", net.JoinHostPort("", cfg.Port))
	if err != nil {
		return err
	}

	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: constants.ReadHeaderTimeout,
		ReadTimeout:       constants.ReadTimeout,
		WriteTimeout:      constants.WriteTimeout,
		IdleTimeout:       constants.IdleTimeout,
		MaxHeaderBytes:    constants.MaxHeaderBytes,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case sig := <-stop:
		log.Printf("shutdown signal: %v", sig)
		ctx, cancel := context.WithTimeout(context.Background(), constants.ShutdownTimeout)
		defer cancel()
		return srv.Shutdown(ctx)
	}
}

func redirectTarget(path string) (string, bool) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", false
	}
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return "", false
	}
	return path, true
}
