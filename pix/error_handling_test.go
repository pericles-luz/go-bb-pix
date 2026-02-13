package pix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/pericles-luz/go-bb-pix/internal/apierror"
)

// TestAPIErrorResponses tests handling of various API error responses
func TestAPIErrorResponses(t *testing.T) {
	tests := []struct {
		name           string
		fixtureFile    string
		statusCode     int
		wantStatusCode int
		wantErrType    string
	}{
		{
			name:           "400 bad request - schema validation",
			fixtureFile:    filepath.Join("..", "testdata", "errors", "400_bad_request.json"),
			statusCode:     http.StatusBadRequest,
			wantStatusCode: 400,
			wantErrType:    "CobOperacaoInvalida",
		},
		{
			name:           "403 forbidden - access denied",
			fixtureFile:    filepath.Join("..", "testdata", "errors", "403_forbidden.json"),
			statusCode:     http.StatusForbidden,
			wantStatusCode: 403,
			wantErrType:    "AcessoNegado",
		},
		{
			name:           "404 not found",
			fixtureFile:    filepath.Join("..", "testdata", "errors", "404_not_found.json"),
			statusCode:     http.StatusNotFound,
			wantStatusCode: 404,
			wantErrType:    "NaoEncontrado",
		},
		{
			name:           "422 unprocessable - business rule",
			fixtureFile:    filepath.Join("..", "testdata", "errors", "422_unprocessable.json"),
			statusCode:     http.StatusUnprocessableEntity,
			wantStatusCode: 422,
			wantErrType:    "RegraDeNegocio",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorData, err := os.ReadFile(tt.fixtureFile)
			if err != nil {
				t.Fatalf("Failed to read fixture: %v", err)
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write(errorData)
			}))
			defer server.Close()

			client := NewClient(&http.Client{}, server.URL)
			_, err = client.GetQRCode(context.Background(), "test123")

			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			// Check if it's an API error
			if !apierror.Is(err) {
				t.Fatalf("Expected API error, got: %v", err)
			}

			apiErr, extractErr := apierror.As(err)
			if extractErr != nil {
				t.Fatalf("Failed to extract API error: %v", extractErr)
			}

			if apiErr.StatusCode != tt.wantStatusCode {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.wantStatusCode)
			}
		})
	}
}

// TestValidationErrors tests client-side validation errors
// NOTE: These tests document expected validation behavior.
// Currently the client doesn't implement client-side validation,
// so these tests are skipped until validation is added.
func TestValidationErrors(t *testing.T) {
	t.Skip("Client-side validation not yet implemented")

	tests := []struct {
		name    string
		request interface{}
		method  string
		wantErr bool
	}{
		{
			name: "empty txid",
			request: CreateQRCodeRequest{
				TxID:       "",
				Value:      100.00,
				Expiration: 3600,
			},
			method:  "CreateQRCode",
			wantErr: true,
		},
		{
			name: "negative value",
			request: CreateQRCodeRequest{
				TxID:       "valid123456789012345678901234",
				Value:      -100.00,
				Expiration: 3600,
			},
			method:  "CreateQRCode",
			wantErr: true,
		},
		{
			name: "zero expiration",
			request: CreateQRCodeRequest{
				TxID:       "valid123456789012345678901234",
				Value:      100.00,
				Expiration: 0,
			},
			method:  "CreateQRCode",
			wantErr: true,
		},
		{
			name: "expiration too large",
			request: CreateQRCodeRequest{
				TxID:       "valid123456789012345678901234",
				Value:      100.00,
				Expiration: 86401, // Max is 86400 (24 hours)
			},
			method:  "CreateQRCode",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a server that should never be called due to validation
			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				t.Error("Server should not be called when validation fails")
			}))
			defer server.Close()

			client := NewClient(&http.Client{}, server.URL)

			var err error
			switch tt.method {
			case "CreateQRCode":
				req := tt.request.(CreateQRCodeRequest)
				_, err = client.CreateQRCode(context.Background(), req)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Validation error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && callCount > 0 {
				t.Error("Server was called despite validation error")
			}
		})
	}
}

// TestHTTPStatusCodes tests various HTTP status code handling
func TestHTTPStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    bool
	}{
		{
			name:       "200 OK",
			statusCode: http.StatusOK,
			body:       `{"txid":"test","status":"ATIVA","calendario":{"criacao":"2024-01-01T00:00:00Z","expiracao":3600},"valor":{"original":"100.00"},"revisao":0}`,
			wantErr:    false,
		},
		{
			name:       "201 Created",
			statusCode: http.StatusCreated,
			body:       `{"txid":"test","status":"ATIVA","calendario":{"criacao":"2024-01-01T00:00:00Z","expiracao":3600},"valor":{"original":"100.00"},"revisao":0}`,
			wantErr:    false,
		},
		{
			name:       "400 Bad Request",
			statusCode: http.StatusBadRequest,
			body:       `{"type":"error","status":400,"detail":"Invalid request"}`,
			wantErr:    true,
		},
		{
			name:       "401 Unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       `{"type":"error","status":401,"detail":"Unauthorized"}`,
			wantErr:    true,
		},
		{
			name:       "403 Forbidden",
			statusCode: http.StatusForbidden,
			body:       `{"type":"error","status":403,"detail":"Forbidden"}`,
			wantErr:    true,
		},
		{
			name:       "404 Not Found",
			statusCode: http.StatusNotFound,
			body:       `{"type":"error","status":404,"detail":"Not found"}`,
			wantErr:    true,
		},
		{
			name:       "429 Too Many Requests",
			statusCode: http.StatusTooManyRequests,
			body:       `{"type":"error","status":429,"detail":"Rate limit exceeded"}`,
			wantErr:    true,
		},
		{
			name:       "500 Internal Server Error",
			statusCode: http.StatusInternalServerError,
			body:       `{"type":"error","status":500,"detail":"Internal error"}`,
			wantErr:    true,
		},
		{
			name:       "502 Bad Gateway",
			statusCode: http.StatusBadGateway,
			body:       `{"type":"error","status":502,"detail":"Bad gateway"}`,
			wantErr:    true,
		},
		{
			name:       "503 Service Unavailable",
			statusCode: http.StatusServiceUnavailable,
			body:       `{"type":"error","status":503,"detail":"Service unavailable"}`,
			wantErr:    true,
		},
		{
			name:       "504 Gateway Timeout",
			statusCode: http.StatusGatewayTimeout,
			body:       `{"type":"error","status":504,"detail":"Gateway timeout"}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.body != "" {
					w.Write([]byte(tt.body))
				}
			}))
			defer server.Close()

			client := NewClient(&http.Client{}, server.URL)
			_, err := client.GetQRCode(context.Background(), "test123")

			if (err != nil) != tt.wantErr {
				t.Errorf("Error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				// Verify it's an API error
				if !apierror.Is(err) {
					t.Errorf("Expected API error, got: %T", err)
				}
			}
		})
	}
}

// TestMalformedJSONResponse tests handling of malformed JSON responses
func TestMalformedJSONResponse(t *testing.T) {
	tests := []struct {
		name     string
		response string
		wantErr  bool
	}{
		{
			name:     "valid JSON",
			response: `{"txid":"test","status":"ATIVA","calendario":{"criacao":"2024-01-01T00:00:00Z","expiracao":3600},"valor":{"original":"100.00"},"revisao":0}`,
			wantErr:  false,
		},
		{
			name:     "malformed JSON - missing quote",
			response: `{"txid":"test,"status":"ATIVA"}`,
			wantErr:  true,
		},
		{
			name:     "malformed JSON - invalid syntax",
			response: `{invalid json}`,
			wantErr:  true,
		},
		{
			name:     "empty response",
			response: ``,
			wantErr:  true,
		},
		{
			name:     "not JSON at all",
			response: `This is not JSON`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := NewClient(&http.Client{}, server.URL)
			_, err := client.GetQRCode(context.Background(), "test123")

			if (err != nil) != tt.wantErr {
				t.Errorf("Error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDuplicateTxID tests handling of duplicate transaction ID error
func TestDuplicateTxID(t *testing.T) {
	errorResponse := map[string]interface{}{
		"type":   "https://api.bb.com.br/api/v2/error/RegraDeNegocio",
		"title":  "Entidade não processável",
		"status": 422,
		"detail": "Violação de regra de negócio",
		"violacoes": []map[string]string{
			{
				"razao":      "O txid já existe para este CPF/CNPJ",
				"propriedade": "txid",
			},
		},
	}

	errorData, _ := json.Marshal(errorResponse)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(errorData)
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	req := CreateQRCodeRequest{
		TxID:       "duplicate123456789012345678901",
		Value:      100.00,
		Expiration: 3600,
	}

	_, err := client.CreateQRCode(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for duplicate txid")
	}

	if !apierror.Is(err) {
		t.Fatalf("Expected API error, got: %T", err)
	}

	apiErr, _ := apierror.As(err)
	if apiErr.StatusCode != 422 {
		t.Errorf("StatusCode = %d, want 422", apiErr.StatusCode)
	}
}

// TestNetworkErrors tests handling of network-level errors
func TestNetworkErrors(t *testing.T) {
	// Use an invalid URL to trigger network error
	client := NewClient(&http.Client{}, "http://invalid.localhost:99999")

	_, err := client.GetQRCode(context.Background(), "test123")

	if err == nil {
		t.Fatal("Expected network error, got nil")
	}

	// Network errors should not be API errors
	if apierror.Is(err) {
		t.Error("Network error should not be wrapped as API error")
	}
}
