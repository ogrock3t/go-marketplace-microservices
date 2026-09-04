package product

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"order-service/internal/domain"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type Product struct {
	ID                int64  `json:"id"`
	SellerID          int64  `json:"seller_id"`
	CategoryID        int64  `json:"category_id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Price             int64  `json:"price"`
	AvailableQuantity int64  `json:"available_quantity"`
	Status            string `json:"status"`
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

func (c *Client) ReserveProduct(ctx context.Context, productID int64, quantity int64) (*domain.ReservedProduct, error) {
	product, err := c.sendQuantityRequest(ctx, "reserve", productID, quantity)
	if err != nil {
		return nil, err
	}

	return &domain.ReservedProduct{
		ID:    product.ID,
		Price: product.Price,
	}, nil
}

func (c *Client) ReleaseProduct(ctx context.Context, productID int64, quantity int64) error {
	_, err := c.sendQuantityRequest(ctx, "release", productID, quantity)
	return err
}

func (c *Client) sendQuantityRequest(ctx context.Context, action string, productID int64, quantity int64) (*Product, error) {
	body, err := json.Marshal(map[string]int64{"quantity": quantity})
	if err != nil {
		return nil, fmt.Errorf("failed to encode %s request: %w", action, err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/products/%d/%s", c.baseURL, productID, action),
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s request: %w", action, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to %s product %d: %w", action, productID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("product service returned status %d for product %d", resp.StatusCode, productID)
	}

	var product Product
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, fmt.Errorf("failed to decode %s response: %w", action, err)
	}

	return &product, nil
}
