package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ogrock3t/go-marketplace-microservices/authentication-service/internal/config"
	httpserver "github.com/ogrock3t/go-marketplace-microservices/authentication-service/internal/http-server"
	"github.com/ogrock3t/go-marketplace-microservices/authentication-service/internal/http-server/handler"
	"github.com/ogrock3t/go-marketplace-microservices/authentication-service/internal/security"
	"github.com/ogrock3t/go-marketplace-microservices/authentication-service/internal/service"
	storage "github.com/ogrock3t/go-marketplace-microservices/authentication-service/internal/storage/postgres"
)

var bcryptCost = 5

// @title           Auth Service API
// @version         1.0
// @host            localhost:8080
// @BasePath        /
func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := config.Load()

	m, err := migrate.New("file://migrations", "pgx5://"+cfg.DatabaseDSN[len("postgres://"):])
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

	userStorage := storage.NewUserStorage(db)
	hasher := security.NewBcryptHasher(bcryptCost)
	jwtService, err := security.NewJWTService(cfg.RSAPrivateKeyPath, cfg.AccessTokenDuration)
	if err != nil {
		log.Error("failed to init jwt service", "error", err)
		os.Exit(1)
	}

	authService := service.NewAuthService(userStorage, hasher, jwtService)
	authHandler := handler.NewAuthHandler(authService)

	router := httpserver.NewRouter(authHandler, log)

	log.Info("starting server", "addr", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, router); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
