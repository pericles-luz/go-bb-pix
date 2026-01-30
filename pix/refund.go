package pix

import (
	"context"
	"fmt"
	"net/http"
)

// CreateRefund creates a refund for a payment
func (c *Client) CreateRefund(ctx context.Context, e2eid, refundID string, req CreateRefundRequest) (*RefundResponse, error) {
	if e2eid == "" {
		return nil, fmt.Errorf("e2eid is required")
	}
	if refundID == "" {
		return nil, fmt.Errorf("refundID is required")
	}

	path := fmt.Sprintf("/pix/%s/devolucao/%s", e2eid, refundID)

	httpReq, err := c.http.NewRequest(ctx, http.MethodPut, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var resp RefundResponse
	if err := c.http.Do(httpReq, &resp); err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	return &resp, nil
}

// GetRefund retrieves a refund by EndToEndID and refund ID
func (c *Client) GetRefund(ctx context.Context, e2eid, refundID string) (*RefundResponse, error) {
	if e2eid == "" {
		return nil, fmt.Errorf("e2eid is required")
	}
	if refundID == "" {
		return nil, fmt.Errorf("refundID is required")
	}

	path := fmt.Sprintf("/pix/%s/devolucao/%s", e2eid, refundID)

	httpReq, err := c.http.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var resp RefundResponse
	if err := c.http.Do(httpReq, &resp); err != nil {
		return nil, fmt.Errorf("failed to get refund: %w", err)
	}

	return &resp, nil
}
