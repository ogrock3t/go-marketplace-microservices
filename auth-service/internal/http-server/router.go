package httpserver

import (
	"log/slog"
	"net/http"

	_ "github.com/ogrock3t/go-marketplace-microservices/authentication-service/docs"
	"github.com/ogrock3t/go-marketplace-microservices/authentication-service/internal/http-server/handler"
	"github.com/ogrock3t/go-marketplace-microservices/authentication-service/internal/http-server/middleware/logger"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(h *handler.AuthHandler, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", h.Register)
	mux.HandleFunc("POST /login", h.Login)
	mux.HandleFunc("POST /refresh-token", h.RefreshToken)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	return logger.New(log)(mux)
}
