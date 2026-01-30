package pix

import (
	"context"
	"fmt"
	"net/http"
)

// GetPayment retrieves a payment by EndToEndID
func (c *Client) GetPayment(ctx context.Context, e2eid string) (*PaymentResponse, error) {
	if e2eid == "" {
		return nil, fmt.Errorf("e2eid is required")
	}

	path := fmt.Sprintf("/pix/%s", e2eid)

	httpReq, err := c.http.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var resp PaymentResponse
	if err := c.http.Do(httpReq, &resp); err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return &resp, nil
}

// ListPayments lists payments with optional filters
func (c *Client) ListPayments(ctx context.Context, params ListPaymentsParams) (*PaymentListResponse, error) {
	path := "/pix"

	httpReq, err := c.http.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := httpReq.URL.Query()
	q.Set("inicio", params.StartDate.Format("2006-01-02T15:04:05Z07:00"))
	q.Set("fim", params.EndDate.Format("2006-01-02T15:04:05Z07:00"))

	if params.TxID != "" {
		q.Set("txid", params.TxID)
	}
	if params.CPF != "" {
		q.Set("cpf", params.CPF)
	}
	if params.CNPJ != "" {
		q.Set("cnpj", params.CNPJ)
	}
	if params.Page > 0 {
		q.Set("paginaAtual", fmt.Sprintf("%d", params.Page))
	}
	if params.PageSize > 0 {
		q.Set("itensPorPagina", fmt.Sprintf("%d", params.PageSize))
	}

	httpReq.URL.RawQuery = q.Encode()

	var resp PaymentListResponse
	if err := c.http.Do(httpReq, &resp); err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}

	return &resp, nil
}
