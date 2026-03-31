package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nurtikaga/jun-1/internal/app"
	"github.com/nurtikaga/jun-1/internal/handler"
	"github.com/nurtikaga/jun-1/internal/infrastructure/postgres"
	"github.com/nurtikaga/jun-1/pkg/health"
	"github.com/nurtikaga/jun-1/pkg/logger"
)

func main() {
	log := logger.New()
	log.Info("service started", "version", "1.0.0")

	dsn := envOr("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable")
	migrationsPath := envOr("MIGRATIONS_PATH", "migrations")

	if err := postgres.RunMigrations(dsn, migrationsPath); err != nil {
		log.Error("migrations failed", "error", err)
		os.Exit(1)
	}
	log.Info("migrations applied")

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Error("db connect failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := postgres.NewProductRepo(pool)
	svc := app.NewProductService(repo)

	mux := http.NewServeMux()
	mux.Handle("GET /healthz", health.Handler())

	productHandler := handler.NewProductHandler(svc, log)
	productHandler.RegisterRoutes(mux)

	addr := envOr("SERVER_ADDR", ":8080")
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("server listening", "addr", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	log.Info("shutting down gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	log.Info("server stopped")
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
