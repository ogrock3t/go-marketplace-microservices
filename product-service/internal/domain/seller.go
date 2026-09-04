package domain

import (
	"fmt"
	"time"
)

var (
	ErrSellerNotFound = fmt.Errorf("seller not found")
)

type Seller struct {
	ID        int64     `db:"id"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
