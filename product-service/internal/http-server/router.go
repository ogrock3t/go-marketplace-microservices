package httpserver

import (
	"log/slog"
	"net/http"

	"product-service/internal/http-server/handler"
	"product-service/internal/http-server/middleware/logger"
)

func NewRouter(
	sellerHandler *handler.SellerHandler,
	categoryHandler *handler.CategoryHandler,
	productHandler *handler.ProductHandler,
	log *slog.Logger,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/sellers", sellerHandler.CreateSeller)
	mux.HandleFunc("GET /api/v1/sellers/{id}", sellerHandler.GetSellerByID)
	mux.HandleFunc("PUT /api/v1/sellers/{id}", sellerHandler.UpdateSeller)
	mux.HandleFunc("DELETE /api/v1/sellers/{id}", sellerHandler.DeleteSeller)

	mux.HandleFunc("POST /api/v1/categories", categoryHandler.CreateCategory)
	mux.HandleFunc("GET /api/v1/categories", categoryHandler.ListCategories)
	mux.HandleFunc("GET /api/v1/categories/{id}", categoryHandler.GetCategoryByID)
	mux.HandleFunc("PUT /api/v1/categories/{id}", categoryHandler.UpdateCategory)
	mux.HandleFunc("DELETE /api/v1/categories/{id}", categoryHandler.DeleteCategory)
	mux.HandleFunc("GET /api/v1/categories/{id}/subcategories", categoryHandler.ListSubcategories)

	mux.HandleFunc("POST /api/v1/products", productHandler.CreateProduct)
	mux.HandleFunc("GET /api/v1/products/{id}", productHandler.GetProductByID)
	mux.HandleFunc("PUT /api/v1/products/{id}", productHandler.UpdateProductByID)
	mux.HandleFunc("DELETE /api/v1/products/{id}", productHandler.DeleteProduct)
	mux.HandleFunc("POST /api/v1/products/{id}/reserve", productHandler.ReserveProduct)
	mux.HandleFunc("POST /api/v1/products/{id}/release", productHandler.ReleaseProduct)
	mux.HandleFunc("GET /api/v1/sellers/{id}/products", productHandler.ListProductsBySeller)
	mux.HandleFunc("GET /api/v1/categories/{id}/products", productHandler.ListProductsByCategory)

	log.Info("Router successfully initialized with REST endpoints", slog.String("version", "v1"))

	loggerMiddleware := logger.RequestLogger(log)

	wrappedMux := loggerMiddleware(mux)

	return wrappedMux
}
