package main

import (
	"log"

	"github.com/PreserveMyGames/website/internal/blog"
	"github.com/PreserveMyGames/website/internal/config"
	"github.com/PreserveMyGames/website/internal/constants"
	"github.com/PreserveMyGames/website/internal/i18n"
	"github.com/PreserveMyGames/website/internal/sandbox"
	"github.com/PreserveMyGames/website/internal/web"
)

func main() {
	cfg := config.Load()

	bundle, err := i18n.New()
	if err != nil {
		log.Fatalf("i18n: %v", err)
	}

	index, err := blog.Load(cfg.Production())
	if err != nil {
		log.Fatalf("blog: %v", err)
	}

	srv, err := web.New(cfg, bundle, index)
	if err != nil {
		log.Fatalf("server: %v", err)
	}

	if err := sandbox.Apply(cfg.Port); err != nil {
		log.Fatalf("sandbox: %v", err)
	}

	log.Printf("%s listening on :%s", constants.AppName, cfg.Port)
	if err := web.ListenAndServe(cfg, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
