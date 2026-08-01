package main

import (
	"log"

	"github.com/PreserveMyGames/store/internal/catalog"
	"github.com/PreserveMyGames/store/internal/config"
	"github.com/PreserveMyGames/store/internal/constants"
	"github.com/PreserveMyGames/store/internal/db"
	"github.com/PreserveMyGames/store/internal/fx"
	"github.com/PreserveMyGames/store/internal/i18n"
	"github.com/PreserveMyGames/store/internal/payments/monero"
	"github.com/PreserveMyGames/store/internal/sandbox"
	"github.com/PreserveMyGames/store/internal/web"
)

func main() {
	cfg := config.Load()

	bundle, err := i18n.New()
	if err != nil {
		log.Fatalf("i18n: %v", err)
	}

	sqlDB, err := db.Open(cfg.DataDir)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer sqlDB.Close()

	store := catalog.New(sqlDB)
	if err := store.SyncCatalog(); err != nil {
		log.Fatalf("catalog: %v", err)
	}

	rates := fx.New()
	rates.Start()

	var wallet monero.Wallet
	var poller *monero.Poller
	if cfg.MoneroEnabled() {
		rpc := monero.NewRPC(cfg.MoneroRPCURL, cfg.MoneroRPCUser, cfg.MoneroRPCPass)
		wallet = rpc
		poller = monero.NewPoller(store, wallet, cfg.MoneroMinConfirmations)
		poller.Start()
		defer poller.Stop()
	}

	go func() {
		_ = store.ReleaseExpiredReservations()
	}()

	srv, err := web.New(cfg, bundle, store, rates, wallet)
	if err != nil {
		log.Fatalf("server: %v", err)
	}

	if err := sandbox.Apply(cfg.Port, cfg.DataDir); err != nil {
		log.Fatalf("sandbox: %v", err)
	}

	log.Printf("%s listening on :%s", constants.AppName, cfg.Port)
	if err := web.ListenAndServe(cfg, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
