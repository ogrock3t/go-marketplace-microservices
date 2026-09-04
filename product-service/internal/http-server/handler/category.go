package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"product-service/internal/domain"
	"product-service/internal/dto"
	"product-service/internal/service"
)

const ErrorCategoryNotFound = "category not found"

type CategoryHandler struct {
	categoryService *service.CategoryService
}

func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // Limit request body to 1MB
	defer r.Body.Close()

	var req dto.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.categoryService.CreateCategory(r.Context(), &req)
	if err != nil {
		writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *CategoryHandler) GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeJSONError(w, "invalid category id", http.StatusBadRequest)
		return
	}

	resp, err := h.categoryService.GetCategoryByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			writeJSONError(w, ErrorCategoryNotFound, http.StatusNotFound)
			return
		}

		writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // Limit request body to 1MB
	defer r.Body.Close()

	var req dto.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	id, err := parsePathID(r, "id")
	if err != nil {
		writeJSONError(w, "invalid category id", http.StatusBadRequest)
		return
	}
	req.ID = id

	if err := validate.Struct(req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.categoryService.UpdateCategory(r.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			writeJSONError(w, ErrorCategoryNotFound, http.StatusNotFound)
			return
		}
		writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "category updated"})
}

func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeJSONError(w, "invalid category id", http.StatusBadRequest)
		return
	}

	err = h.categoryService.DeleteCategory(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			writeJSONError(w, ErrorCategoryNotFound, http.StatusNotFound)
			return
		}
		writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	resp, err := h.categoryService.ListCategories(r.Context())
	if err != nil {
		writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *CategoryHandler) ListSubcategories(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeJSONError(w, "invalid category id", http.StatusBadRequest)
		return
	}

	resp, err := h.categoryService.ListSubcategories(r.Context(), id)
	if err != nil {
		writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
