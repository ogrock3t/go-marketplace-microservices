package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"google.golang.org/grpc"

	"product-service/internal/config"
	"product-service/internal/grpcserver"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
	httpServer := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	grpcListener, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Error("failed to listen grpc", "error", err)
		os.Exit(1)
	}
	grpcServer := grpc.NewServer()
	grpcserver.RegisterInventoryServiceServer(grpcServer, grpcserver.NewInventoryServer(productService))

	errCh := make(chan error, 2)
	go func() {
		log.Info("starting http server", "addr", cfg.HTTPAddr)
		errCh <- httpServer.ListenAndServe()
	}()
	go func() {
		log.Info("starting grpc server", "addr", cfg.GRPCAddr)
		errCh <- grpcServer.Serve(grpcListener)
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("failed to shutdown http server", "error", err)
		os.Exit(1)
	}
	grpcServer.GracefulStop()
	log.Info("Product Service stopped")
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
