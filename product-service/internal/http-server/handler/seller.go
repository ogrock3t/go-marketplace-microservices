package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"product-service/internal/domain"
	"product-service/internal/dto"
	"product-service/internal/service"
)

type SellerHandler struct {
	sellerService *service.SellerService
}

func NewSellerHandler(sellerService *service.SellerService) *SellerHandler {
	return &SellerHandler{
		sellerService: sellerService,
	}
}

func (h *SellerHandler) CreateSeller(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	defer r.Body.Close()

	var req dto.CreateSellerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.sellerService.CreateSeller(r.Context(), &req)
	if err != nil {
		writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *SellerHandler) GetSellerByID(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeJSONError(w, "invalid seller id", http.StatusBadRequest)
		return
	}
	req := dto.GetSellerByIDRequest{ID: id}

	resp, err := h.sellerService.GetSellerByID(r.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrSellerNotFound) {
			writeJSONError(w, "seller not found", http.StatusNotFound)
			return
		}
		writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *SellerHandler) UpdateSeller(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	defer r.Body.Close()

	var req dto.UpdateSellerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, ErrInvalidJSON, http.StatusBadRequest)
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeJSONError(w, "invalid seller id", http.StatusBadRequest)
		return
	}
	req.ID = id

	if err := validate.Struct(req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.sellerService.UpdateSeller(r.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrSellerNotFound) {
			writeJSONError(w, "seller not found", http.StatusNotFound)
			return
		}
		writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "seller updated"})
}

func (h *SellerHandler) DeleteSeller(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeJSONError(w, "invalid seller id", http.StatusBadRequest)
		return
	}
	req := dto.DeleteSellerRequest{ID: id}

	err = h.sellerService.DeleteSeller(r.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrSellerNotFound) {
			writeJSONError(w, "seller not found", http.StatusNotFound)
			return
		}
		writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "seller deleted"})
}
