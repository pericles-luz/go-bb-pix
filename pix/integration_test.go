package pix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestCreateQRCodeIntegration tests QR Code creation with real API format
func TestCreateQRCodeIntegration(t *testing.T) {
	responseData, err := os.ReadFile(filepath.Join("..", "testdata", "cob", "create_response.json"))
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	tests := []struct {
		name           string
		request        CreateQRCodeRequest
		serverResponse string
		serverStatus   int
		wantErr        bool
		validate       func(t *testing.T, resp *QRCodeResponse)
	}{
		{
			name: "successful creation with all fields",
			request: CreateQRCodeRequest{
				TxID:                  "fb2761260e554ad593c7226beb5cb650",
				Value:                 37.00,
				Expiration:            3600,
				PayerSolicitation:     "Serviço realizado.",
				AdditionalInformation: "Informação Adicional 1",
				Debtor: &Debtor{
					CPF:  "12345678909",
					Name: "Francisco da Silva",
				},
			},
			serverResponse: string(responseData),
			serverStatus:   http.StatusCreated,
			wantErr:        false,
			validate: func(t *testing.T, resp *QRCodeResponse) {
				if resp.TxID != "fb2761260e554ad593c7226beb5cb650" {
					t.Errorf("TxID = %s, want fb2761260e554ad593c7226beb5cb650", resp.TxID)
				}
				if resp.Status != "ATIVA" {
					t.Errorf("Status = %s, want ATIVA", resp.Status)
				}
				if resp.QRCode == "" {
					t.Error("QRCode should not be empty")
				}
				if resp.Value.Original != "37.00" {
					t.Errorf("Value = %s, want 37.00", resp.Value.Original)
				}
				if resp.Loc == nil {
					t.Error("Location should not be nil")
				}
			},
		},
		{
			name: "creation with minimal fields",
			request: CreateQRCodeRequest{
				TxID:       "minimal123456789012345678",
				Value:      100.00,
				Expiration: 3600,
			},
			serverResponse: string(responseData),
			serverStatus:   http.StatusCreated,
			wantErr:        false,
			validate: func(t *testing.T, resp *QRCodeResponse) {
				if resp.Status != "ATIVA" {
					t.Errorf("Status = %s, want ATIVA", resp.Status)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify HTTP method
				if r.Method != http.MethodPut {
					t.Errorf("Method = %s, want PUT", r.Method)
				}

				// Verify path includes txid
				expectedPath := "/cob/" + tt.request.TxID
				if r.URL.Path != expectedPath {
					t.Errorf("Path = %s, want %s", r.URL.Path, expectedPath)
				}

				// Verify Content-Type
				if ct := r.Header.Get("Content-Type"); ct != "application/json" {
					t.Errorf("Content-Type = %s, want application/json", ct)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			client := NewClient(&http.Client{}, server.URL)
			resp, err := client.CreateQRCode(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateQRCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, resp)
			}
		})
	}
}

// TestListQRCodesWithPagination tests pagination in list operations
func TestListQRCodesWithPagination(t *testing.T) {
	responseData, err := os.ReadFile(filepath.Join("..", "testdata", "cob", "list_response.json"))
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// Verify pagination parameters
		page := query.Get("paginacao.paginaAtual")
		pageSize := query.Get("paginacao.itensPorPagina")

		if page != "0" && page != "" {
			t.Errorf("Page = %s, expected 0 or empty", page)
		}
		if pageSize != "100" && pageSize != "" {
			t.Errorf("PageSize = %s, expected 100 or empty", pageSize)
		}

		// Verify date range (max 5 days for PIX payments)
		inicio := query.Get("inicio")
		fim := query.Get("fim")
		if inicio == "" || fim == "" {
			t.Error("Date range parameters missing")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(responseData)
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	params := ListQRCodesParams{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		Page:      0,
		PageSize:  100,
	}

	resp, err := client.ListQRCodes(context.Background(), params)
	if err != nil {
		t.Fatalf("ListQRCodes() error = %v", err)
	}

	if resp.Parameters.Pagination.TotalItems != 2 {
		t.Errorf("TotalItems = %d, want 2", resp.Parameters.Pagination.TotalItems)
	}

	if len(resp.QRCodes) != 2 {
		t.Errorf("len(QRCodes) = %d, want 2", len(resp.QRCodes))
	}
}

// TestListPaymentsDateRange tests payment listing with date range validation
func TestListPaymentsDateRange(t *testing.T) {
	responseData, err := os.ReadFile(filepath.Join("..", "testdata", "pix", "list_response.json"))
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	tests := []struct {
		name      string
		params    ListPaymentsParams
		wantErr   bool
		errString string
	}{
		{
			name: "valid 5 day window",
			params: ListPaymentsParams{
				StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 1, 5, 23, 59, 59, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "with txid filter",
			params: ListPaymentsParams{
				StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 1, 5, 23, 59, 59, 0, time.UTC),
				TxID:      "fb2761260e554ad593c7226beb5cb650",
			},
			wantErr: false,
		},
		{
			name: "with CPF filter",
			params: ListPaymentsParams{
				StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 1, 5, 23, 59, 59, 0, time.UTC),
				CPF:       "12345678909",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				query := r.URL.Query()

				// Verify required parameters
				if query.Get("inicio") == "" {
					t.Error("inicio parameter missing")
				}
				if query.Get("fim") == "" {
					t.Error("fim parameter missing")
				}

				// Verify optional filters
				if tt.params.TxID != "" && query.Get("txid") != tt.params.TxID {
					t.Errorf("txid = %s, want %s", query.Get("txid"), tt.params.TxID)
				}
				if tt.params.CPF != "" && query.Get("cpf") != tt.params.CPF {
					t.Errorf("cpf = %s, want %s", query.Get("cpf"), tt.params.CPF)
				}

				w.Header().Set("Content-Type", "application/json")
				w.Write(responseData)
			}))
			defer server.Close()

			client := NewClient(&http.Client{}, server.URL)
			resp, err := client.ListPayments(context.Background(), tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListPayments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(resp.Payments) != 2 {
					t.Errorf("len(Payments) = %d, want 2", len(resp.Payments))
				}
			}
		})
	}
}

// TestRefundCreation tests refund creation flow
func TestRefundCreation(t *testing.T) {
	responseData, err := os.ReadFile(filepath.Join("..", "testdata", "pix", "refund_response.json"))
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	tests := []struct {
		name        string
		e2eID       string
		refundID    string
		request     CreateRefundRequest
		wantStatus  string
		wantValue   string
	}{
		{
			name:     "partial refund",
			e2eID:    "E12345678202406201221abcdef12345",
			refundID: "dev1",
			request: CreateRefundRequest{
				Value: 7.89,
			},
			wantStatus: "EM_PROCESSAMENTO",
			wantValue:  "7.89",
		},
		{
			name:     "refund with reason",
			e2eID:    "E12345678202406201221abcdef12345",
			refundID: "dev2",
			request: CreateRefundRequest{
				Value:  10.00,
				Reason: "Produto não entregue",
			},
			wantStatus: "EM_PROCESSAMENTO",
			wantValue:  "7.89",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify method
				if r.Method != http.MethodPut {
					t.Errorf("Method = %s, want PUT", r.Method)
				}

				// Verify path
				expectedPath := "/pix/" + tt.e2eID + "/devolucao/" + tt.refundID
				if r.URL.Path != expectedPath {
					t.Errorf("Path = %s, want %s", r.URL.Path, expectedPath)
				}

				// Verify request body
				var reqBody CreateRefundRequest
				if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
					t.Errorf("Failed to decode request: %v", err)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				w.Write(responseData)
			}))
			defer server.Close()

			client := NewClient(&http.Client{}, server.URL)
			resp, err := client.CreateRefund(context.Background(), tt.e2eID, tt.refundID, tt.request)

			if err != nil {
				t.Fatalf("CreateRefund() error = %v", err)
			}

			if resp.Status != tt.wantStatus {
				t.Errorf("Status = %s, want %s", resp.Status, tt.wantStatus)
			}

			if resp.Value != tt.wantValue {
				t.Errorf("Value = %s, want %s", resp.Value, tt.wantValue)
			}
		})
	}
}

// TestContextCancellation tests context cancellation handling
func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.GetQRCode(ctx, "txid123")
	if err == nil {
		t.Fatal("Expected context deadline exceeded error")
	}

	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("Context error = %v, want DeadlineExceeded", ctx.Err())
	}
}

// TestConcurrentRequests tests thread safety
func TestConcurrentRequests(t *testing.T) {
	responseData, err := os.ReadFile(filepath.Join("..", "testdata", "cob", "create_response.json"))
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.Write(responseData)
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	// Launch 10 concurrent requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			req := CreateQRCodeRequest{
				TxID:       "concurrent" + string(rune('0'+id)) + "12345678901234567890",
				Value:      100.00,
				Expiration: 3600,
			}
			_, err := client.CreateQRCode(context.Background(), req)
			if err != nil {
				t.Errorf("Request %d failed: %v", id, err)
			}
			done <- true
		}(i)
	}

	// Wait for all requests
	for i := 0; i < 10; i++ {
		<-done
	}

	if requestCount != 10 {
		t.Errorf("Request count = %d, want 10", requestCount)
	}
}
