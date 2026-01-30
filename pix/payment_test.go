package pix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetPayment_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/pix/E12345678202401151000000000001" {
			t.Errorf("Path = %s, want /pix/E12345678202401151000000000001", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"endToEndId": "E12345678202401151000000000001",
			"txid":       "txid123",
			"valor":      "100.50",
			"horario":    "2024-01-15T10:30:45Z",
			"infoPagador": "Info do pagador",
		})
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	resp, err := client.GetPayment(context.Background(), "E12345678202401151000000000001")

	if err != nil {
		t.Fatalf("GetPayment() error = %v", err)
	}
	if resp.EndToEndID != "E12345678202401151000000000001" {
		t.Errorf("EndToEndID = %s", resp.EndToEndID)
	}
	if resp.TxID != "txid123" {
		t.Errorf("TxID = %s, want txid123", resp.TxID)
	}
	if resp.Value != "100.50" {
		t.Errorf("Value = %s, want 100.50", resp.Value)
	}
}

func TestClient_GetPayment_InvalidE2EID(t *testing.T) {
	client := NewClient(&http.Client{}, "http://localhost")

	_, err := client.GetPayment(context.Background(), "")

	if err == nil {
		t.Fatal("Expected error for empty e2eid, got nil")
	}
}

func TestClient_GetPayment_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Payment not found",
		})
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	_, err := client.GetPayment(context.Background(), "nonexistent")

	if err == nil {
		t.Fatal("Expected error for not found, got nil")
	}
}

func TestClient_ListPayments_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/pix" {
			t.Errorf("Path = %s, want /pix", r.URL.Path)
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
			"pix": []interface{}{
				map[string]interface{}{
					"endToEndId": "E12345678202401151000000000001",
					"txid":       "txid1",
					"valor":      "100.00",
					"horario":    "2024-01-15T10:00:00Z",
				},
				map[string]interface{}{
					"endToEndId": "E12345678202401151000000000002",
					"txid":       "txid2",
					"valor":      "200.00",
					"horario":    "2024-01-16T10:00:00Z",
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	params := ListPaymentsParams{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
	}

	resp, err := client.ListPayments(context.Background(), params)

	if err != nil {
		t.Fatalf("ListPayments() error = %v", err)
	}
	if len(resp.Payments) != 2 {
		t.Errorf("len(Payments) = %d, want 2", len(resp.Payments))
	}
	if resp.Parameters.Pagination.TotalItems != 2 {
		t.Errorf("TotalItems = %d, want 2", resp.Parameters.Pagination.TotalItems)
	}
}

func TestClient_ListPayments_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// Verify filters
		if query.Get("txid") != "txid123" {
			t.Errorf("txid = %s, want txid123", query.Get("txid"))
		}
		if query.Get("cpf") != "12345678900" {
			t.Errorf("cpf = %s, want 12345678900", query.Get("cpf"))
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
					"quantidadeTotalDeItens": 0,
				},
			},
			"pix": []interface{}{},
		})
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	params := ListPaymentsParams{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		TxID:      "txid123",
		CPF:       "12345678900",
	}

	_, err := client.ListPayments(context.Background(), params)

	if err != nil {
		t.Fatalf("ListPayments() error = %v", err)
	}
}

func TestClient_ListPayments_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// Verify pagination parameters
		if query.Get("paginaAtual") != "2" {
			t.Errorf("paginaAtual = %s, want 2", query.Get("paginaAtual"))
		}
		if query.Get("itensPorPagina") != "50" {
			t.Errorf("itensPorPagina = %s, want 50", query.Get("itensPorPagina"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"parametros": map[string]interface{}{
				"inicio": "2024-01-01T00:00:00Z",
				"fim":    "2024-01-31T23:59:59Z",
				"paginacao": map[string]interface{}{
					"paginaAtual":           2,
					"itensPorPagina":        50,
					"quantidadeDePaginas":   5,
					"quantidadeTotalDeItens": 250,
				},
			},
			"pix": []interface{}{},
		})
	}))
	defer server.Close()

	client := NewClient(&http.Client{}, server.URL)

	params := ListPaymentsParams{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		Page:      2,
		PageSize:  50,
	}

	resp, err := client.ListPayments(context.Background(), params)

	if err != nil {
		t.Fatalf("ListPayments() error = %v", err)
	}

	if resp.Parameters.Pagination.CurrentPage != 2 {
		t.Errorf("CurrentPage = %d, want 2", resp.Parameters.Pagination.CurrentPage)
	}
	if resp.Parameters.Pagination.ItemsPerPage != 50 {
		t.Errorf("ItemsPerPage = %d, want 50", resp.Parameters.Pagination.ItemsPerPage)
	}
}
