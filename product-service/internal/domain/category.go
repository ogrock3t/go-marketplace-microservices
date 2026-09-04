package domain

import (
	"fmt"
	"time"
)

var (
	ErrCategoryNotFound = fmt.Errorf("category not found")
)

type Category struct {
	ID          int64     `db:"id"`
	ParentID    *int64    `db:"parent_id"` // nil = root category
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
