package pix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_CreateRefund_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/pix/E12345678202401151000000000001/devolucao/refund123" {
			t.Errorf("Path = %s", r.URL.Path)
		}

		// Verify request body
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode body: %v", err)
		}
		if body["valor"] != "50.25" {
			t.Errorf("valor = %v, want 50.25", body["valor"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "refund123",
			"rtrId":  "D12345678202401151000000000001",
			"valor":  "50.25",
			"status": "EM_PROCESSAMENTO",
			"horario": map[string]interface{}{
				"solicitacao": "2024-01-15T11:00:00Z",
			},
		})
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	req := CreateRefundRequest{
		Value: 50.25,
	}

	resp, err := client.CreateRefund(context.Background(), "E12345678202401151000000000001", "refund123", req)

	if err != nil {
		t.Fatalf("CreateRefund() error = %v", err)
	}
	if resp.ID != "refund123" {
		t.Errorf("ID = %s, want refund123", resp.ID)
	}
	if resp.Value != "50.25" {
		t.Errorf("Value = %s, want 50.25", resp.Value)
	}
	if resp.Status != "EM_PROCESSAMENTO" {
		t.Errorf("Status = %s, want EM_PROCESSAMENTO", resp.Status)
	}
}

func TestClient_CreateRefund_WithReason(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode body: %v", err)
		}

		if body["motivo"] != "Cliente solicitou" {
			t.Errorf("motivo = %v, want 'Cliente solicitou'", body["motivo"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "refund456",
			"rtrId":  "D12345678202401151000000000002",
			"valor":  "100.00",
			"status": "EM_PROCESSAMENTO",
			"motivo": "Cliente solicitou",
			"horario": map[string]interface{}{
				"solicitacao": "2024-01-15T11:00:00Z",
			},
		})
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	req := CreateRefundRequest{
		Value:  100.00,
		Reason: "Cliente solicitou",
	}

	resp, err := client.CreateRefund(context.Background(), "E12345678202401151000000000001", "refund456", req)

	if err != nil {
		t.Fatalf("CreateRefund() error = %v", err)
	}
	if resp.Reason != "Cliente solicitou" {
		t.Errorf("Reason = %s, want 'Cliente solicitou'", resp.Reason)
	}
}

func TestClient_CreateRefund_InvalidParams(t *testing.T) {
	client := NewClient(&http.Client{}, "http://localhost")

	tests := []struct {
		name     string
		e2eid    string
		refundID string
		wantErr  bool
	}{
		{"empty e2eid", "", "refund123", true},
		{"empty refundID", "E12345678202401151000000000001", "", true},
		{"both empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateRefundRequest{Value: 50.00}
			_, err := client.CreateRefund(context.Background(), tt.e2eid, tt.refundID, req)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateRefund() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_GetRefund_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/pix/E12345678202401151000000000001/devolucao/refund123" {
			t.Errorf("Path = %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "refund123",
			"rtrId":  "D12345678202401151000000000001",
			"valor":  "50.25",
			"status": "DEVOLVIDO",
			"horario": map[string]interface{}{
				"solicitacao": "2024-01-15T11:00:00Z",
				"liquidacao":  "2024-01-15T11:01:00Z",
			},
		})
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	resp, err := client.GetRefund(context.Background(), "E12345678202401151000000000001", "refund123")

	if err != nil {
		t.Fatalf("GetRefund() error = %v", err)
	}
	if resp.ID != "refund123" {
		t.Errorf("ID = %s, want refund123", resp.ID)
	}
	if resp.Status != "DEVOLVIDO" {
		t.Errorf("Status = %s, want DEVOLVIDO", resp.Status)
	}
	if resp.Time.Settlement.IsZero() {
		t.Error("Settlement time should not be zero for completed refund")
	}
}

func TestClient_GetRefund_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Refund not found",
		})
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	_, err := client.GetRefund(context.Background(), "E12345678202401151000000000001", "nonexistent")

	if err == nil {
		t.Fatal("Expected error for not found, got nil")
	}
}

func TestClient_GetRefund_InvalidParams(t *testing.T) {
	client := NewClient(&http.Client{}, "http://localhost")

	tests := []struct {
		name     string
		e2eid    string
		refundID string
		wantErr  bool
	}{
		{"empty e2eid", "", "refund123", true},
		{"empty refundID", "E12345678202401151000000000001", "", true},
		{"both empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.GetRefund(context.Background(), tt.e2eid, tt.refundID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRefund() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
