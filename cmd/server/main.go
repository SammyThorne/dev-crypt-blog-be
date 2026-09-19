// Command server runs the dev-crypt blog API.
//
//	@title						Dev Crypt Blog API
//	@version					1.0
//	@description				Blog posts, comments and role management for the dev-crypt blog.
//	@BasePath					/
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				A Firebase ID token, formatted as "Bearer <token>".
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/EmiraBlight/dev-crypt-blog-be/internal/config"
	"github.com/EmiraBlight/dev-crypt-blog-be/internal/db"
	"github.com/EmiraBlight/dev-crypt-blog-be/internal/router"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	pool, err := db.Connect(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router.New(cfg, pool),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if cfg.UseTLS {
			log.Printf("Running on port %d with CORS and HTTPS enabled", cfg.Port)
			errCh <- srv.ListenAndServeTLS(cfg.TLSCertPath, cfg.TLSKeyPath)
		} else {
			log.Printf("Running on port %d with CORS enabled (HTTP, TLS disabled)", cfg.Port)
			errCh <- srv.ListenAndServe()
		}
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		log.Println("Shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
