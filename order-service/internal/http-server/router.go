package httpserver

import (
	"log/slog"
	"net/http"

	"order-service/internal/http-server/handler"
	"order-service/internal/http-server/middleware/logger"
)

func NewRouter(orderHandler *handler.OrderHandler, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/orders", orderHandler.CreateOrder)
	mux.HandleFunc("GET /api/v1/orders/{id}", orderHandler.GetOrderByID)
	mux.HandleFunc("PATCH /api/v1/orders/{id}/status", orderHandler.UpdateOrderStatus)
	mux.HandleFunc("GET /api/v1/users/{id}/orders", orderHandler.ListOrdersByUser)

	log.Info("Router successfully initialized with REST endpoints", slog.String("version", "v1"))

	loggerMiddleware := logger.RequestLogger(log)
	return loggerMiddleware(mux)
}
