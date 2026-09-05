package product

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"

	"order-service/internal/domain"
)

const (
	codecName            = "json"
	inventoryServiceName = "product.v1.InventoryService"
	reserveProductMethod = "/" + inventoryServiceName + "/ReserveProduct"
	releaseProductMethod = "/" + inventoryServiceName + "/ReleaseProduct"
)

type jsonCodec struct{}

func (jsonCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (jsonCodec) Name() string {
	return codecName
}

func init() {
	encoding.RegisterCodec(jsonCodec{})
}

type Client struct {
	conn *grpc.ClientConn
}

type quantityRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
}

type productResponse struct {
	ID                int64  `json:"id"`
	SellerID          int64  `json:"seller_id"`
	CategoryID        int64  `json:"category_id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Price             int64  `json:"price"`
	AvailableQuantity int64  `json:"available_quantity"`
	Status            string `json:"status"`
}

func NewClient(target string) (*Client, error) {
	conn, err := grpc.Dial(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(jsonCodec{})),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create product grpc client: %w", err)
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) ReserveProduct(ctx context.Context, productID int64, quantity int64) (*domain.ReservedProduct, error) {
	var product productResponse
	if err := c.invokeQuantityMethod(ctx, reserveProductMethod, productID, quantity, &product); err != nil {
		return nil, err
	}

	return &domain.ReservedProduct{
		ID:    product.ID,
		Price: product.Price,
	}, nil
}

func (c *Client) ReleaseProduct(ctx context.Context, productID int64, quantity int64) error {
	var product productResponse
	return c.invokeQuantityMethod(ctx, releaseProductMethod, productID, quantity, &product)
}

func (c *Client) invokeQuantityMethod(
	ctx context.Context,
	method string,
	productID int64,
	quantity int64,
	out *productResponse,
) error {
	req := quantityRequest{
		ProductID: productID,
		Quantity:  quantity,
	}

	if err := c.conn.Invoke(ctx, method, &req, out); err != nil {
		return fmt.Errorf("product grpc call failed: %w", err)
	}

	return nil
}
