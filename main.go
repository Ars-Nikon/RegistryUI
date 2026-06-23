package main

import (
	"RegistryUI/internal/auth"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"RegistryUI/internal/api"
	"RegistryUI/internal/config"
)

func main() {
	cfg := config.Load()
	authSvc, err := auth.NewService(auth.Config{
		Secret: []byte(cfg.JwtSecret),
		TTL:    cfg.JwtTTL,
		Issuer: cfg.JwtIssuer,
	})
	if err != nil {
		log.Fatal(err)
	}

	srv := api.NewServer(cfg, authSvc)

	engine := srv.Engine()
	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if cfg.TLSEnabled() {
			log.Printf("RegistryUI API listening on %s (HTTPS, registry: %s)", cfg.Addr, cfg.RegistryURL)
			if err := httpServer.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("server error: %v", err)
			}
			return
		}
		log.Printf("RegistryUI API listening on %s (HTTP, registry: %s)", cfg.Addr, cfg.RegistryURL)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
