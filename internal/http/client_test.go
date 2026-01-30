package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pericles-luz/go-bb-pix/internal/apierror"
)

func TestNewClient(t *testing.T) {
	httpClient := &http.Client{}
	baseURL := "https://api.example.com"

	client := NewClient(httpClient, baseURL)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.httpClient != httpClient {
		t.Error("httpClient not set correctly")
	}
	if client.baseURL != baseURL {
		t.Error("baseURL not set correctly")
	}
}

func TestClient_NewRequest_GET(t *testing.T) {
	client := NewClient(&http.Client{}, "https://api.example.com")

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/path", nil)

	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	if req.Method != http.MethodGet {
		t.Errorf("Method = %s, want GET", req.Method)
	}
	if req.URL.String() != "https://api.example.com/path" {
		t.Errorf("URL = %s, want https://api.example.com/path", req.URL.String())
	}
}

func TestClient_NewRequest_WithBody(t *testing.T) {
	client := NewClient(&http.Client{}, "https://api.example.com")

	body := map[string]string{"key": "value"}
	req, err := client.NewRequest(context.Background(), http.MethodPost, "/path", body)

	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	// Verify Content-Type header
	if ct := req.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", ct)
	}

	// Verify body
	bodyBytes, _ := io.ReadAll(req.Body)
	var decoded map[string]string
	if err := json.Unmarshal(bodyBytes, &decoded); err != nil {
		t.Fatalf("Failed to decode body: %v", err)
	}
	if decoded["key"] != "value" {
		t.Errorf("Body key = %s, want value", decoded["key"])
	}
}

func TestClient_NewRequest_NilBody(t *testing.T) {
	client := NewClient(&http.Client{}, "https://api.example.com")

	req, err := client.NewRequest(context.Background(), http.MethodGet, "/path", nil)

	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	if req.Body != nil {
		t.Error("Body should be nil for nil input")
	}
}

func TestClient_Do_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)
	req, _ := client.NewRequest(context.Background(), http.MethodGet, "/", nil)

	var result map[string]string
	err := client.Do(req, &result)

	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if result["result"] != "success" {
		t.Errorf("result = %s, want success", result["result"])
	}
}

func TestClient_Do_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Invalid request",
			"errors": []map[string]string{
				{"field": "amount", "message": "required"},
			},
		})
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)
	req, _ := client.NewRequest(context.Background(), http.MethodPost, "/", nil)

	var result map[string]string
	err := client.Do(req, &result)

	if err == nil {
		t.Fatal("Expected error for 400 response, got nil")
	}

	apiErr, getErr := apierror.As(err)
	if getErr != nil {
		t.Fatalf("Expected APIError, got %v", getErr)
	}

	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusBadRequest)
	}
	if apiErr.Message != "Invalid request" {
		t.Errorf("Message = %s, want 'Invalid request'", apiErr.Message)
	}
	if len(apiErr.Details) != 1 {
		t.Errorf("len(Details) = %d, want 1", len(apiErr.Details))
	}
}

func TestClient_Do_NilTarget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)
	req, _ := client.NewRequest(context.Background(), http.MethodDelete, "/", nil)

	err := client.Do(req, nil)

	if err != nil {
		t.Fatalf("Do() with nil target error = %v", err)
	}
}

func TestClient_Do_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)
	req, _ := client.NewRequest(context.Background(), http.MethodGet, "/", nil)

	var result map[string]string
	err := client.Do(req, &result)

	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

func TestParseErrorResponse(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		body           string
		wantMessage    string
		wantDetailsLen int
	}{
		{
			name:       "error with message and details",
			statusCode: 400,
			body: `{
				"message": "Validation failed",
				"errors": [
					{"field": "txid", "message": "required"},
					{"field": "value", "message": "must be positive"}
				]
			}`,
			wantMessage:    "Validation failed",
			wantDetailsLen: 2,
		},
		{
			name:           "error with only message",
			statusCode:     404,
			body:           `{"message": "Not found"}`,
			wantMessage:    "Not found",
			wantDetailsLen: 0,
		},
		{
			name:           "error without message",
			statusCode:     500,
			body:           `{}`,
			wantMessage:    "HTTP 500",
			wantDetailsLen: 0,
		},
		{
			name:           "invalid JSON",
			statusCode:     500,
			body:           `invalid json`,
			wantMessage:    "HTTP 500",
			wantDetailsLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := io.NopCloser(strings.NewReader(tt.body))

			err := parseErrorResponse(tt.statusCode, body)

			apiErr, getErr := apierror.As(err)
			if getErr != nil {
				t.Fatalf("Expected APIError, got: %v", getErr)
			}

			if apiErr.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.statusCode)
			}
			if apiErr.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", apiErr.Message, tt.wantMessage)
			}
			if len(apiErr.Details) != tt.wantDetailsLen {
				t.Errorf("len(Details) = %d, want %d", len(apiErr.Details), tt.wantDetailsLen)
			}
		})
	}
}

func TestClient_BuildURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		path     string
		wantURL  string
		wantErr  bool
	}{
		{
			name:    "simple path",
			baseURL: "https://api.example.com",
			path:    "/path",
			wantURL: "https://api.example.com/path",
			wantErr: false,
		},
		{
			name:    "path without leading slash",
			baseURL: "https://api.example.com",
			path:    "path",
			wantURL: "https://api.example.com/path",
			wantErr: false,
		},
		{
			name:    "base URL with trailing slash",
			baseURL: "https://api.example.com/",
			path:    "/path",
			wantURL: "https://api.example.com/path",
			wantErr: false,
		},
		{
			name:    "base URL with path",
			baseURL: "https://api.example.com/api/v1",
			path:    "/resource",
			wantURL: "https://api.example.com/resource",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(&http.Client{}, tt.baseURL)
			req, err := client.NewRequest(context.Background(), http.MethodGet, tt.path, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && req.URL.String() != tt.wantURL {
				t.Errorf("URL = %s, want %s", req.URL.String(), tt.wantURL)
			}
		})
	}
}
