package pix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_CreateQRCode_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPut {
			t.Errorf("Method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/cob/txid123" {
			t.Errorf("Path = %s, want /cob/txid123", r.URL.Path)
		}

		// Return response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"calendario": map[string]interface{}{
				"criacao":   "2024-01-15T10:00:00Z",
				"expiracao": 3600,
			},
			"txid":    "txid123",
			"revisao": 1,
			"status":  "ATIVA",
			"valor": map[string]interface{}{
				"original": "100.50",
			},
			"pixCopiaECola": "00020126580014br.gov.bcb.pix...",
		})
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	req := CreateQRCodeRequest{
		TxID:       "txid123",
		Value:      100.50,
		Expiration: 3600,
	}

	resp, err := client.CreateQRCode(context.Background(), req)

	if err != nil {
		t.Fatalf("CreateQRCode() error = %v", err)
	}
	if resp.TxID != "txid123" {
		t.Errorf("TxID = %s, want txid123", resp.TxID)
	}
	if resp.Status != "ATIVA" {
		t.Errorf("Status = %s, want ATIVA", resp.Status)
	}
	if resp.QRCode == "" {
		t.Error("QRCode should not be empty")
	}
}

func TestClient_CreateQRCode_InvalidRequest(t *testing.T) {
	client := NewClient(&http.Client{}, "http://localhost")

	req := CreateQRCodeRequest{
		TxID: "", // Empty TxID should cause validation error
	}

	_, err := client.CreateQRCode(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error for invalid request, got nil")
	}
}

func TestClient_GetQRCode_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/cob/txid123" {
			t.Errorf("Path = %s, want /cob/txid123", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"calendario": map[string]interface{}{
				"criacao":   "2024-01-15T10:00:00Z",
				"expiracao": 3600,
			},
			"txid":    "txid123",
			"revisao": 1,
			"status":  "ATIVA",
			"valor": map[string]interface{}{
				"original": "100.50",
			},
		})
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	resp, err := client.GetQRCode(context.Background(), "txid123")

	if err != nil {
		t.Fatalf("GetQRCode() error = %v", err)
	}
	if resp.TxID != "txid123" {
		t.Errorf("TxID = %s, want txid123", resp.TxID)
	}
}

func TestClient_GetQRCode_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "QR Code not found",
		})
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	_, err := client.GetQRCode(context.Background(), "nonexistent")

	if err == nil {
		t.Fatal("Expected error for not found, got nil")
	}
}

func TestClient_UpdateQRCode_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/cob/txid123" {
			t.Errorf("Path = %s, want /cob/txid123", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"calendario": map[string]interface{}{
				"criacao":   "2024-01-15T10:00:00Z",
				"expiracao": 7200,
			},
			"txid":    "txid123",
			"revisao": 2,
			"status":  "ATIVA",
			"valor": map[string]interface{}{
				"original": "150.00",
			},
		})
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	req := UpdateQRCodeRequest{
		Value:      150.00,
		Expiration: 7200,
	}

	resp, err := client.UpdateQRCode(context.Background(), "txid123", req)

	if err != nil {
		t.Fatalf("UpdateQRCode() error = %v", err)
	}
	if resp.Revision != 2 {
		t.Errorf("Revision = %d, want 2", resp.Revision)
	}
	if resp.Value.Original != "150.00" {
		t.Errorf("Value.Original = %s, want 150.00", resp.Value.Original)
	}
}

func TestClient_ListQRCodes_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/cob" {
			t.Errorf("Path = %s, want /cob", r.URL.Path)
		}

		// Verify query parameters
		query := r.URL.Query()
		if query.Get("inicio") == "" {
			t.Error("inicio query parameter missing")
		}
		if query.Get("fim") == "" {
			t.Error("fim query parameter missing")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"parametros": map[string]interface{}{
				"inicio": "2024-01-01T00:00:00Z",
				"fim":    "2024-01-31T23:59:59Z",
				"paginacao": map[string]interface{}{
					"paginaAtual":           0,
					"itensPorPagina":        100,
					"quantidadeDePaginas":   1,
					"quantidadeTotalDeItens": 2,
				},
			},
			"cobs": []interface{}{
				map[string]interface{}{
					"calendario": map[string]interface{}{
						"criacao":   "2024-01-15T10:00:00Z",
						"expiracao": 3600,
					},
					"txid":    "txid1",
					"revisao": 1,
					"status":  "ATIVA",
					"valor": map[string]interface{}{
						"original": "100.00",
					},
				},
				map[string]interface{}{
					"calendario": map[string]interface{}{
						"criacao":   "2024-01-16T10:00:00Z",
						"expiracao": 3600,
					},
					"txid":    "txid2",
					"revisao": 1,
					"status":  "CONCLUIDA",
					"valor": map[string]interface{}{
						"original": "200.00",
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	params := ListQRCodesParams{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
	}

	resp, err := client.ListQRCodes(context.Background(), params)

	if err != nil {
		t.Fatalf("ListQRCodes() error = %v", err)
	}
	if len(resp.QRCodes) != 2 {
		t.Errorf("len(QRCodes) = %d, want 2", len(resp.QRCodes))
	}
	if resp.Parameters.Pagination.TotalItems != 2 {
		t.Errorf("TotalItems = %d, want 2", resp.Parameters.Pagination.TotalItems)
	}
}

func TestClient_ListQRCodes_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// Verify filters are sent as query parameters
		if query.Get("cpf") != "12345678900" {
			t.Errorf("cpf = %s, want 12345678900", query.Get("cpf"))
		}
		if query.Get("status") != "ATIVA" {
			t.Errorf("status = %s, want ATIVA", query.Get("status"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"parametros": map[string]interface{}{
				"inicio": "2024-01-01T00:00:00Z",
				"fim":    "2024-01-31T23:59:59Z",
				"paginacao": map[string]interface{}{
					"paginaAtual":           0,
					"itensPorPagina":        100,
					"quantidadeDePaginas":   1,
					"quantidadeTotalDeItens": 1,
				},
			},
			"cobs": []interface{}{},
		})
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	params := ListQRCodesParams{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		CPF:       "12345678900",
		Status:    "ATIVA",
	}

	_, err := client.ListQRCodes(context.Background(), params)

	if err != nil {
		t.Fatalf("ListQRCodes() error = %v", err)
	}
}

func TestClient_DeleteQRCode_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/cob/txid123" {
			t.Errorf("Path = %s, want /cob/txid123", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	err := client.DeleteQRCode(context.Background(), "txid123")

	if err != nil {
		t.Fatalf("DeleteQRCode() error = %v", err)
	}
}
