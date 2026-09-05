package dto

type CreateOrderRequest struct {
	UserID int64                    `json:"user_id" validate:"required,gt=0"`
	Items  []CreateOrderItemRequest `json:"items" validate:"required,min=1,dive"`
}

type CreateOrderItemRequest struct {
	ProductID int64 `json:"product_id" validate:"required,gt=0"`
	Quantity  int64 `json:"quantity" validate:"required,gt=0"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=CREATED PAID CANCELED COMPLETED"`
}

type OrderResponse struct {
	ID          int64               `json:"id"`
	UserID      int64               `json:"user_id"`
	Status      string              `json:"status"`
	TotalAmount int64               `json:"total_amount"`
	Items       []OrderItemResponse `json:"items"`
	CreatedAt   string              `json:"created_at"`
	UpdatedAt   string              `json:"updated_at"`
}

type OrderItemResponse struct {
	ID        int64 `json:"id"`
	OrderID   int64 `json:"order_id"`
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
	Price     int64 `json:"price"`
}

type ListOrdersResponse struct {
	Orders []OrderResponse `json:"orders"`
}
