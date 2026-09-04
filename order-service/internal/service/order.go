package service

import (
	"context"
	"fmt"
	"time"

	"order-service/internal/domain"
	"order-service/internal/dto"
	"order-service/internal/events"
	"order-service/internal/repository"
)

type ProductInventory interface {
	ReserveProduct(ctx context.Context, productID int64, quantity int64) (*domain.ReservedProduct, error)
	ReleaseProduct(ctx context.Context, productID int64, quantity int64) error
}

type OrderService struct {
	orderRepo repository.OrderRepository
	inventory ProductInventory
	publisher events.Publisher
}

func NewOrderService(orderRepo repository.OrderRepository, inventory ProductInventory, publisher events.Publisher) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		inventory: inventory,
		publisher: publisher,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	items := make([]domain.OrderItem, 0, len(req.Items))
	reservedItems := make([]dto.CreateOrderItemRequest, 0, len(req.Items))
	var total int64

	for _, item := range req.Items {
		product, err := s.inventory.ReserveProduct(ctx, item.ProductID, item.Quantity)
		if err != nil {
			s.releaseReservedProducts(ctx, reservedItems)
			return nil, fmt.Errorf("%w: product_id=%d: %v", domain.ErrProductReservationFailed, item.ProductID, err)
		}

		reservedItems = append(reservedItems, item)
		lineTotal := product.Price * item.Quantity
		total += lineTotal
		items = append(items, domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
		})
	}

	order := &domain.Order{
		UserID:      req.UserID,
		Status:      domain.StatusCreated,
		TotalAmount: total,
		Items:       items,
	}

	if err := s.orderRepo.CreateOrder(ctx, order); err != nil {
		s.releaseReservedProducts(ctx, reservedItems)
		return nil, err
	}

	resp := orderToResponse(order)
	if err := s.publisher.Publish(ctx, events.OrderCreatedEvent, resp); err != nil {
		return nil, fmt.Errorf("failed to publish order created event: %w", err)
	}

	return resp, nil
}

func (s *OrderService) releaseReservedProducts(ctx context.Context, items []dto.CreateOrderItemRequest) {
	for i := len(items) - 1; i >= 0; i-- {
		_ = s.inventory.ReleaseProduct(ctx, items[i].ProductID, items[i].Quantity)
	}
}

func (s *OrderService) GetOrderByID(ctx context.Context, id int64) (*dto.OrderResponse, error) {
	order, err := s.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return orderToResponse(order), nil
}

func (s *OrderService) ListOrdersByUser(ctx context.Context, userID int64) (*dto.ListOrdersResponse, error) {
	orders, err := s.orderRepo.ListOrdersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := dto.ListOrdersResponse{
		Orders: make([]dto.OrderResponse, 0, len(orders)),
	}
	for i := range orders {
		resp.Orders = append(resp.Orders, *orderToResponse(&orders[i]))
	}

	return &resp, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, id int64, req *dto.UpdateOrderStatusRequest) (*dto.OrderResponse, error) {
	nextStatus, err := domain.ParseOrderStatus(req.Status)
	if err != nil {
		return nil, err
	}

	current, err := s.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if current.Status == nextStatus {
		return orderToResponse(current), nil
	}
	if !domain.CanTransition(current.Status, nextStatus) {
		return nil, domain.ErrInvalidStatusTransition
	}

	updated, err := s.orderRepo.UpdateOrderStatus(ctx, id, nextStatus)
	if err != nil {
		return nil, err
	}

	resp := orderToResponse(updated)
	if err := s.publisher.Publish(ctx, events.OrderStatusChangedEvent, map[string]any{
		"order_id":    updated.ID,
		"user_id":     updated.UserID,
		"old_status":  current.Status,
		"new_status":  updated.Status,
		"occurred_at": time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		return nil, fmt.Errorf("failed to publish order status changed event: %w", err)
	}

	return resp, nil
}

func orderToResponse(order *domain.Order) *dto.OrderResponse {
	items := make([]dto.OrderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, dto.OrderItemResponse{
			ID:        item.ID,
			OrderID:   item.OrderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	return &dto.OrderResponse{
		ID:          order.ID,
		UserID:      order.UserID,
		Status:      string(order.Status),
		TotalAmount: order.TotalAmount,
		Items:       items,
		CreatedAt:   order.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   order.UpdatedAt.Format(time.RFC3339),
	}
}
