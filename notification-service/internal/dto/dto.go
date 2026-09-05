package dto

type Envelope struct {
	Type       string `json:"type"`
	Payload    any    `json:"payload"`
	OccurredAt string `json:"occurred_at"`
}

type OrderCreatedPayload struct {
	ID          int64           `json:"id"`
	UserID      int64           `json:"user_id"`
	Status      string          `json:"status"`
	TotalAmount int64           `json:"total_amount"`
	Items       []OrderItemData `json:"items"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

type OrderItemData struct {
	ID        int64 `json:"id"`
	OrderID   int64 `json:"order_id"`
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
	Price     int64 `json:"price"`
}

type OrderStatusChangedPayload struct {
	OrderID    int64  `json:"order_id"`
	UserID     int64  `json:"user_id"`
	OldStatus  string `json:"old_status"`
	NewStatus  string `json:"new_status"`
	OccurredAt string `json:"occurred_at"`
}

type PaymentProcessedPayload struct {
	PaymentID     int64  `json:"payment_id"`
	OrderID       int64  `json:"order_id"`
	UserID        int64  `json:"user_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
	Success       bool   `json:"success"`
	FailureReason string `json:"failure_reason,omitempty"`
	ProcessedAt   string `json:"processed_at"`
}
