package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"order-service/internal/domain"
	"order-service/internal/dto"
)

type orderRepoMock struct {
	createErr error
	created   *domain.Order
	current   *domain.Order
	updated   *domain.Order
}

func (m *orderRepoMock) CreateOrder(_ context.Context, order *domain.Order) error {
	if m.createErr != nil {
		return m.createErr
	}

	order.ID = 100
	order.CreatedAt = time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)
	order.UpdatedAt = order.CreatedAt
	for i := range order.Items {
		order.Items[i].ID = int64(i + 1)
		order.Items[i].OrderID = order.ID
	}
	m.created = order
	return nil
}

func (m *orderRepoMock) GetOrderByID(_ context.Context, id int64) (*domain.Order, error) {
	if m.current == nil {
		return nil, domain.ErrOrderNotFound
	}
	return m.current, nil
}

func (m *orderRepoMock) ListOrdersByUser(_ context.Context, userID int64) ([]domain.Order, error) {
	return []domain.Order{*m.current}, nil
}

func (m *orderRepoMock) UpdateOrderStatus(_ context.Context, id int64, status domain.OrderStatus) (*domain.Order, error) {
	m.updated.Status = status
	return m.updated, nil
}

type inventoryMock struct {
	reserved []dto.CreateOrderItemRequest
	released []dto.CreateOrderItemRequest
	failOnID int64
}

func (m *inventoryMock) ReserveProduct(_ context.Context, productID int64, quantity int64) (*domain.ReservedProduct, error) {
	if productID == m.failOnID {
		return nil, errors.New("reservation failed")
	}

	m.reserved = append(m.reserved, dto.CreateOrderItemRequest{ProductID: productID, Quantity: quantity})
	return &domain.ReservedProduct{ID: productID, Price: productID * 100}, nil
}

func (m *inventoryMock) ReleaseProduct(_ context.Context, productID int64, quantity int64) error {
	m.released = append(m.released, dto.CreateOrderItemRequest{ProductID: productID, Quantity: quantity})
	return nil
}

type publisherMock struct {
	eventType string
	payload   any
}

func (m *publisherMock) Publish(_ context.Context, eventType string, payload any) error {
	m.eventType = eventType
	m.payload = payload
	return nil
}

func TestCreateOrderReservesProductsAndPublishesEvent(t *testing.T) {
	repo := &orderRepoMock{}
	inventory := &inventoryMock{}
	publisher := &publisherMock{}
	service := NewOrderService(repo, inventory, publisher)

	resp, err := service.CreateOrder(context.Background(), &dto.CreateOrderRequest{
		UserID: 7,
		Items: []dto.CreateOrderItemRequest{
			{ProductID: 10, Quantity: 2},
			{ProductID: 20, Quantity: 1},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder returned error: %v", err)
	}

	if resp.ID != 100 {
		t.Fatalf("expected order id 100, got %d", resp.ID)
	}
	if resp.Status != string(domain.StatusCreated) {
		t.Fatalf("expected status %s, got %s", domain.StatusCreated, resp.Status)
	}
	if resp.TotalAmount != 4000 {
		t.Fatalf("expected total 4000, got %d", resp.TotalAmount)
	}
	if len(inventory.reserved) != 2 {
		t.Fatalf("expected 2 reserve calls, got %d", len(inventory.reserved))
	}
	if publisher.eventType != "OrderCreatedEvent" {
		t.Fatalf("expected OrderCreatedEvent, got %s", publisher.eventType)
	}
}

func TestCreateOrderReleasesReservedProductsWhenRepositoryFails(t *testing.T) {
	repo := &orderRepoMock{createErr: errors.New("db failed")}
	inventory := &inventoryMock{}
	publisher := &publisherMock{}
	service := NewOrderService(repo, inventory, publisher)

	_, err := service.CreateOrder(context.Background(), &dto.CreateOrderRequest{
		UserID: 7,
		Items: []dto.CreateOrderItemRequest{
			{ProductID: 10, Quantity: 2},
			{ProductID: 20, Quantity: 1},
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}

	if len(inventory.released) != 2 {
		t.Fatalf("expected 2 release calls, got %d", len(inventory.released))
	}
	if inventory.released[0].ProductID != 20 || inventory.released[1].ProductID != 10 {
		t.Fatalf("expected releases in reverse order, got %+v", inventory.released)
	}
}

func TestUpdateOrderStatusRejectsInvalidTransition(t *testing.T) {
	repo := &orderRepoMock{
		current: &domain.Order{ID: 100, UserID: 7, Status: domain.StatusCompleted},
		updated: &domain.Order{ID: 100, UserID: 7, Status: domain.StatusCompleted},
	}
	service := NewOrderService(repo, &inventoryMock{}, &publisherMock{})

	_, err := service.UpdateOrderStatus(context.Background(), 100, &dto.UpdateOrderStatusRequest{Status: string(domain.StatusPaid)})
	if !errors.Is(err, domain.ErrInvalidStatusTransition) {
		t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
	}
}
