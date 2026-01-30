package pix

import (
	"context"
	"fmt"
	"net/http"
)

// CreateQRCode creates a new QR Code
func (c *Client) CreateQRCode(ctx context.Context, req CreateQRCodeRequest) (*QRCodeResponse, error) {
	// Validate request
	if req.TxID == "" {
		return nil, fmt.Errorf("txid is required")
	}

	// Build path
	path := fmt.Sprintf("/cob/%s", req.TxID)

	// Create HTTP request
	httpReq, err := c.http.NewRequest(ctx, http.MethodPut, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	var resp QRCodeResponse
	if err := c.http.Do(httpReq, &resp); err != nil {
		return nil, fmt.Errorf("failed to create qr code: %w", err)
	}

	return &resp, nil
}

// GetQRCode retrieves a QR Code by TxID
func (c *Client) GetQRCode(ctx context.Context, txID string) (*QRCodeResponse, error) {
	if txID == "" {
		return nil, fmt.Errorf("txid is required")
	}

	path := fmt.Sprintf("/cob/%s", txID)

	httpReq, err := c.http.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var resp QRCodeResponse
	if err := c.http.Do(httpReq, &resp); err != nil {
		return nil, fmt.Errorf("failed to get qr code: %w", err)
	}

	return &resp, nil
}

// UpdateQRCode updates an existing QR Code
func (c *Client) UpdateQRCode(ctx context.Context, txID string, req UpdateQRCodeRequest) (*QRCodeResponse, error) {
	if txID == "" {
		return nil, fmt.Errorf("txid is required")
	}

	path := fmt.Sprintf("/cob/%s", txID)

	httpReq, err := c.http.NewRequest(ctx, http.MethodPatch, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var resp QRCodeResponse
	if err := c.http.Do(httpReq, &resp); err != nil {
		return nil, fmt.Errorf("failed to update qr code: %w", err)
	}

	return &resp, nil
}

// ListQRCodes lists QR Codes with optional filters
func (c *Client) ListQRCodes(ctx context.Context, params ListQRCodesParams) (*QRCodeListResponse, error) {
	path := "/cob"

	httpReq, err := c.http.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := httpReq.URL.Query()
	q.Set("inicio", params.StartDate.Format("2006-01-02T15:04:05Z07:00"))
	q.Set("fim", params.EndDate.Format("2006-01-02T15:04:05Z07:00"))

	if params.CPF != "" {
		q.Set("cpf", params.CPF)
	}
	if params.CNPJ != "" {
		q.Set("cnpj", params.CNPJ)
	}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if params.Page > 0 {
		q.Set("paginaAtual", fmt.Sprintf("%d", params.Page))
	}
	if params.PageSize > 0 {
		q.Set("itensPorPagina", fmt.Sprintf("%d", params.PageSize))
	}

	httpReq.URL.RawQuery = q.Encode()

	var resp QRCodeListResponse
	if err := c.http.Do(httpReq, &resp); err != nil {
		return nil, fmt.Errorf("failed to list qr codes: %w", err)
	}

	return &resp, nil
}

// DeleteQRCode deletes a QR Code
func (c *Client) DeleteQRCode(ctx context.Context, txID string) error {
	if txID == "" {
		return fmt.Errorf("txid is required")
	}

	path := fmt.Sprintf("/cob/%s", txID)

	httpReq, err := c.http.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if err := c.http.Do(httpReq, nil); err != nil {
		return fmt.Errorf("failed to delete qr code: %w", err)
	}

	return nil
}
