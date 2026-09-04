package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"product-service/internal/config"
	httpserver "product-service/internal/http-server"
	"product-service/internal/http-server/handler"
	"product-service/internal/service"
	storage "product-service/internal/storage/postgres"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	log.Info("Starting Product Service")

	cfg := config.Load()
	if cfg.DatabaseDSN == "" {
		log.Error("DATABASE_DSN is required")
		os.Exit(1)
	}

	m, err := migrate.New("file://migrations", migrationDSN(cfg.DatabaseDSN))
	if err != nil {
		log.Error("failed to init migrations", "error", err)
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	db, err := storage.NewConnection(ctx, cfg.DatabaseDSN)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	sellerStorage := storage.NewSellerStorage(db)
	categoryStorage := storage.NewCategoryStorage(db)
	productStorage := storage.NewProductStorage(db)

	sellerService := service.NewSellerService(sellerStorage)
	categoryService := service.NewCategoryService(categoryStorage)
	productService := service.NewProductService(productStorage)

	sellerHandler := handler.NewSellerHandler(sellerService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	productHandler := handler.NewProductHandler(productService)

	router := httpserver.NewRouter(sellerHandler, categoryHandler, productHandler, log)

	log.Info("starting server", "addr", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, router); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}

func migrationDSN(dsn string) string {
	switch {
	case strings.HasPrefix(dsn, "postgres://"):
		return "pgx5://" + strings.TrimPrefix(dsn, "postgres://")
	case strings.HasPrefix(dsn, "postgresql://"):
		return "pgx5://" + strings.TrimPrefix(dsn, "postgresql://")
	default:
		return dsn
	}
}
