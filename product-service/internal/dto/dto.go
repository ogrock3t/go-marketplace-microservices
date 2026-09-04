package dto

type CreateProductRequest struct {
	SellerID          int64  `json:"seller_id" validate:"required,gt=0"`
	CategoryID        int64  `json:"category_id" validate:"required,gt=0"`
	Name              string `json:"name" validate:"required"`
	Description       string `json:"description"`
	Price             int64  `json:"price" validate:"required,gt=0"`
	AvailableQuantity int64  `json:"available_quantity" validate:"gte=0"`
	Status            string `json:"status" validate:"omitempty,oneof=ACTIVE OUT_OF_STOCK ARCHIVED"`
}

type UpdateProductByIDRequest struct {
	ID                int64  `json:"-"`
	SellerID          int64  `json:"seller_id" validate:"required,gt=0"`
	CategoryID        int64  `json:"category_id" validate:"required,gt=0"`
	Name              string `json:"name" validate:"required"`
	Description       string `json:"description"`
	Price             int64  `json:"price" validate:"required,gt=0"`
	AvailableQuantity int64  `json:"available_quantity" validate:"gte=0"`
	Status            string `json:"status" validate:"required,oneof=ACTIVE OUT_OF_STOCK ARCHIVED"`
}

type ReserveProductRequest struct {
	Quantity int64 `json:"quantity" validate:"required,gt=0"`
}

type CreateSellerRequest struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
}

type GetSellerByIDRequest struct {
	ID int64 `json:"-"`
}

type UpdateSellerRequest struct {
	ID        int64  `json:"-"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
}

type DeleteSellerRequest struct {
	ID int64 `json:"-"`
}

type CreateCategoryRequest struct {
	ParentID    *int64 `json:"parent_id"` // nil = root category
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type UpdateCategoryRequest struct {
	ID          int64  `json:"-"`
	ParentID    *int64 `json:"parent_id"` // nil = root category
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type GetProductByIDResponse struct {
	ID                int64  `json:"id"`
	SellerID          int64  `json:"seller_id"`
	CategoryID        int64  `json:"category_id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Price             int64  `json:"price"`
	AvailableQuantity int64  `json:"available_quantity"`
	Status            string `json:"status"`
}

type ListProductsBySellerResponse struct {
	Products []*GetProductByIDResponse `json:"products"`
}

type ListProductsByCategoryResponse struct {
	Products []*GetProductByIDResponse `json:"products"`
}

type GetSellerByIDResponse struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type GetCategoryByIDResponse struct {
	ID          int64  `json:"id"`
	ParentID    *int64 `json:"parent_id"` // nil = root category
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ListCategoriesResponse struct {
	Categories []*GetCategoryByIDResponse `json:"categories"`
}

type ListSubcategoriesResponse struct {
	Categories []*GetCategoryByIDResponse `json:"categories"`
}
