package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"

	"order-service/internal/domain"
	"order-service/internal/dto"
	"order-service/internal/service"
)

var validate = validator.New()

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	defer r.Body.Close()

	var req dto.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.orderService.CreateOrder(r.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrProductReservationFailed) {
			writeJSONError(w, "failed to reserve product", http.StatusConflict)
			return
		}
		writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathID(r, "id")
	if err != nil {
		writeJSONError(w, "invalid order id", http.StatusBadRequest)
		return
	}

	resp, err := h.orderService.GetOrderByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			writeJSONError(w, "order not found", http.StatusNotFound)
			return
		}
		writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *OrderHandler) ListOrdersByUser(w http.ResponseWriter, r *http.Request) {
	userID, err := parsePathID(r, "id")
	if err != nil {
		writeJSONError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	resp, err := h.orderService.ListOrdersByUser(r.Context(), userID)
	if err != nil {
		writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	defer r.Body.Close()

	id, err := parsePathID(r, "id")
	if err != nil {
		writeJSONError(w, "invalid order id", http.StatusBadRequest)
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.orderService.UpdateOrderStatus(r.Context(), id, &req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderNotFound):
			writeJSONError(w, "order not found", http.StatusNotFound)
		case errors.Is(err, domain.ErrInvalidOrderStatus), errors.Is(err, domain.ErrInvalidStatusTransition):
			writeJSONError(w, err.Error(), http.StatusBadRequest)
		default:
			writeJSONError(w, ErrInternalServer, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
