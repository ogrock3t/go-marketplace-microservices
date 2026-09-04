package domain

import (
	"fmt"
	"time"
)

type ProductStatus string

const (
	StatusActive     ProductStatus = "ACTIVE"
	StatusOutOfStock ProductStatus = "OUT_OF_STOCK"
	StatusArchived   ProductStatus = "ARCHIVED"
)

var (
	ErrProductNotFound   = fmt.Errorf("product not found")
	ErrInsufficientStock = fmt.Errorf("insufficient stock")
)

type Product struct {
	ID                int64         `db:"id"`
	SellerID          int64         `db:"seller_id"`
	CategoryID        int64         `db:"category_id"`
	Name              string        `db:"name"`
	Description       string        `db:"description"`
	Price             int64         `db:"price"`
	AvailableQuantity int64         `db:"available_quantity"`
	Status            ProductStatus `db:"status"`
	CreatedAt         time.Time     `db:"created_at"`
	UpdatedAt         time.Time     `db:"updated_at"`
}

func (p *Product) IsAvailable() bool {
	return p.Status == StatusActive && p.AvailableQuantity > 0
}

func (p *Product) SetStatus(status string) {
	if status == string(StatusActive) || status == string(StatusOutOfStock) || status == string(StatusArchived) {
		p.Status = ProductStatus(status)
	}
}

func (p *Product) GetStatus() string {
	return string(p.Status)
}
