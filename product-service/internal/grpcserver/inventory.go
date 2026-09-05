package grpcserver

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"product-service/internal/domain"
	"product-service/internal/dto"
	"product-service/internal/service"
)

const InventoryServiceName = "product.v1.InventoryService"

type ReserveProductRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
}

type ReleaseProductRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
}

type ProductResponse struct {
	ID                int64  `json:"id"`
	SellerID          int64  `json:"seller_id"`
	CategoryID        int64  `json:"category_id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Price             int64  `json:"price"`
	AvailableQuantity int64  `json:"available_quantity"`
	Status            string `json:"status"`
}

type InventoryServiceServer interface {
	ReserveProduct(context.Context, *ReserveProductRequest) (*ProductResponse, error)
	ReleaseProduct(context.Context, *ReleaseProductRequest) (*ProductResponse, error)
}

type InventoryServer struct {
	productService *service.ProductService
}

func NewInventoryServer(productService *service.ProductService) *InventoryServer {
	return &InventoryServer{productService: productService}
}

func RegisterInventoryServiceServer(server *grpc.Server, service InventoryServiceServer) {
	server.RegisterService(&inventoryServiceDesc, service)
}

func (s *InventoryServer) ReserveProduct(ctx context.Context, req *ReserveProductRequest) (*ProductResponse, error) {
	if req.ProductID <= 0 || req.Quantity <= 0 {
		return nil, status.Error(codes.InvalidArgument, "product_id and quantity must be positive")
	}

	product, err := s.productService.ReserveProduct(ctx, req.ProductID, &dto.ReserveProductRequest{
		Quantity: req.Quantity,
	})
	if err != nil {
		return nil, productError(err)
	}

	return productToResponse(product), nil
}

func (s *InventoryServer) ReleaseProduct(ctx context.Context, req *ReleaseProductRequest) (*ProductResponse, error) {
	if req.ProductID <= 0 || req.Quantity <= 0 {
		return nil, status.Error(codes.InvalidArgument, "product_id and quantity must be positive")
	}

	product, err := s.productService.ReleaseProduct(ctx, req.ProductID, &dto.ReleaseProductRequest{
		Quantity: req.Quantity,
	})
	if err != nil {
		return nil, productError(err)
	}

	return productToResponse(product), nil
}

func productError(err error) error {
	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		return status.Error(codes.NotFound, "product not found")
	case errors.Is(err, domain.ErrInsufficientStock):
		return status.Error(codes.FailedPrecondition, "insufficient stock")
	default:
		return status.Error(codes.Internal, "product service error")
	}
}

func productToResponse(product *dto.GetProductByIDResponse) *ProductResponse {
	return &ProductResponse{
		ID:                product.ID,
		SellerID:          product.SellerID,
		CategoryID:        product.CategoryID,
		Name:              product.Name,
		Description:       product.Description,
		Price:             product.Price,
		AvailableQuantity: product.AvailableQuantity,
		Status:            product.Status,
	}
}

var inventoryServiceDesc = grpc.ServiceDesc{
	ServiceName: InventoryServiceName,
	HandlerType: (*InventoryServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ReserveProduct",
			Handler:    reserveProductHandler,
		},
		{
			MethodName: "ReleaseProduct",
			Handler:    releaseProductHandler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "product/v1/inventory.proto",
}

func reserveProductHandler(
	srv any,
	ctx context.Context,
	dec func(any) error,
	interceptor grpc.UnaryServerInterceptor,
) (any, error) {
	req := new(ReserveProductRequest)
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).ReserveProduct(ctx, req)
	}

	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/" + InventoryServiceName + "/ReserveProduct",
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(InventoryServiceServer).ReserveProduct(ctx, req.(*ReserveProductRequest))
	}

	return interceptor(ctx, req, info, handler)
}

func releaseProductHandler(
	srv any,
	ctx context.Context,
	dec func(any) error,
	interceptor grpc.UnaryServerInterceptor,
) (any, error) {
	req := new(ReleaseProductRequest)
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InventoryServiceServer).ReleaseProduct(ctx, req)
	}

	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/" + InventoryServiceName + "/ReleaseProduct",
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(InventoryServiceServer).ReleaseProduct(ctx, req.(*ReleaseProductRequest))
	}

	return interceptor(ctx, req, info, handler)
}
