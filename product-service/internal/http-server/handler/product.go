package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"product-service/internal/domain"
	"product-service/internal/dto"
	"product-service/internal/service"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

const ErrorProductNotFound = "product not found"

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // Limit request body to 1MB
	defer r.Body.Close()

	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdProduct, err := h.productService.CreateProduct(r.Context(), &req)
	if err != nil {
		writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": createdProduct})
}

func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeJSONError(w, "invalid product id", http.StatusBadRequest)
		return
	}

	resp, err := h.productService.GetProductByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			writeJSONError(w, ErrorProductNotFound, http.StatusNotFound)
			return
		}
		writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *ProductHandler) UpdateProductByID(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // Limit request body to 1MB
	defer r.Body.Close()

	var req dto.UpdateProductByIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	id, err := parsePathID(r, "id")
	if err != nil {
		writeJSONError(w, "invalid product id", http.StatusBadRequest)
		return
	}
	req.ID = id

	if err := validate.Struct(req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.productService.UpdateProductByID(r.Context(), &req); err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			writeJSONError(w, ErrorProductNotFound, http.StatusNotFound)
			return
		}
		writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "product updated successfully"}); err != nil {
		writeJSONError(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeJSONError(w, "invalid product id", http.StatusBadRequest)
		return
	}

	if err := h.productService.DeleteProduct(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			writeJSONError(w, ErrorProductNotFound, http.StatusNotFound)
			return
		}
		writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "product deleted successfully"}); err != nil {
		writeJSONError(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *ProductHandler) ListProductsBySeller(w http.ResponseWriter, r *http.Request) {
	sellerID, err := parsePathID(r, "id")
	if err != nil {
		writeJSONError(w, "invalid seller id", http.StatusBadRequest)
		return
	}

	resp, err := h.productService.ListProductsBySeller(r.Context(), sellerID)
	if err != nil {
		writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		writeJSONError(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *ProductHandler) ListProductsByCategory(w http.ResponseWriter, r *http.Request) {
	categoryID, err := parsePathID(r, "id")
	if err != nil {
		writeJSONError(w, "invalid category id", http.StatusBadRequest)
		return
	}

	resp, err := h.productService.ListProductsByCategory(r.Context(), categoryID)
	if err != nil {
		writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		writeJSONError(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *ProductHandler) ReserveProduct(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	defer r.Body.Close()

	id, err := parsePathID(r, "id")
	if err != nil {
		writeJSONError(w, "invalid product id", http.StatusBadRequest)
		return
	}

	var req dto.ReserveProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.productService.ReserveProduct(r.Context(), id, &req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProductNotFound):
			writeJSONError(w, ErrorProductNotFound, http.StatusNotFound)
		case errors.Is(err, domain.ErrInsufficientStock):
			writeJSONError(w, "insufficient stock", http.StatusConflict)
		default:
			writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
