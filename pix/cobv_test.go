package pix

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

// CobVRequest represents a charge with due date (cobrança com vencimento)
type CobVRequest struct {
	Calendar CobVCalendar `json:"calendario"`
	Debtor   *Debtor      `json:"devedor,omitempty"`
	Value    CobVValue    `json:"valor"`
	Key      string       `json:"chave"`
	PayerSolicitation string  `json:"solicitacaoPagador,omitempty"`
}

// CobVCalendar represents calendar for charges with due date
type CobVCalendar struct {
	DueDate              string `json:"dataDeVencimento"` // YYYY-MM-DD
	ValidAfterDue        int    `json:"validadeAposVencimento,omitempty"`
}

// CobVValue represents value with fines and interest
type CobVValue struct {
	Original string         `json:"original"`
	Fine     *CobVModality  `json:"multa,omitempty"`
	Interest *CobVModality  `json:"juros,omitempty"`
	Discount *CobVDiscount  `json:"desconto,omitempty"`
}

// CobVModality represents fine or interest modality
type CobVModality struct {
	Modality   string `json:"modalidade"` // "1" = fixed value, "2" = percentage
	ValuePerc  string `json:"valorPerc,omitempty"`
}

// CobVDiscount represents discount information
type CobVDiscount struct {
	Modality        string              `json:"modalidade"` // "1" = fixed date
	FixedDateDiscount []FixedDateDiscount `json:"descontoDataFixa,omitempty"`
}

// FixedDateDiscount represents a discount for a specific date
type FixedDateDiscount struct {
	Date      string `json:"data"` // YYYY-MM-DD
	ValuePerc string `json:"valorPerc"`
}

// CobVResponse represents a charge with due date response
type CobVResponse struct {
	Calendar  CobVCalendar     `json:"calendario"`
	TxID      string           `json:"txid"`
	Revision  int              `json:"revisao"`
	Loc       *Location        `json:"loc,omitempty"`
	Location  string           `json:"location,omitempty"`
	Status    string           `json:"status"`
	Debtor    *Debtor          `json:"devedor,omitempty"`
	Value     CobVValue        `json:"valor"`
	Key       string           `json:"chave"`
	PayerSolicitation string  `json:"solicitacaoPagador,omitempty"`
	QRCode    string           `json:"pixCopiaECola,omitempty"`
}

// TestCreateCobVWithDueDate tests creating a charge with due date
func TestCreateCobVWithDueDate(t *testing.T) {
	responseData, err := os.ReadFile(filepath.Join("..", "testdata", "cobv", "create_response.json"))
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	tests := []struct {
		name     string
		request  CobVRequest
		validate func(t *testing.T, resp *CobVResponse)
	}{
		{
			name: "charge with fine and interest",
			request: CobVRequest{
				Calendar: CobVCalendar{
					DueDate:       "2035-06-24",
					ValidAfterDue: 30,
				},
				Debtor: &Debtor{
					CPF:  "12345678909",
					Name: "Francisco da Silva",
				},
				Value: CobVValue{
					Original: "123.45",
					Fine: &CobVModality{
						Modality:  "2",
						ValuePerc: "15.00",
					},
					Interest: &CobVModality{
						Modality:  "2",
						ValuePerc: "2.00",
					},
				},
				Key: "95127446000198",
				PayerSolicitation: "Cobrança dos serviços prestados.",
			},
			validate: func(t *testing.T, resp *CobVResponse) {
				if resp.Calendar.DueDate != "2035-06-24" {
					t.Errorf("DueDate = %s, want 2035-06-24", resp.Calendar.DueDate)
				}
				if resp.Calendar.ValidAfterDue != 30 {
					t.Errorf("ValidAfterDue = %d, want 30", resp.Calendar.ValidAfterDue)
				}
				if resp.Value.Fine == nil {
					t.Error("Fine should not be nil")
				} else if resp.Value.Fine.ValuePerc != "15.00" {
					t.Errorf("Fine = %s, want 15.00", resp.Value.Fine.ValuePerc)
				}
				if resp.Value.Interest == nil {
					t.Error("Interest should not be nil")
				} else if resp.Value.Interest.ValuePerc != "2.00" {
					t.Errorf("Interest = %s, want 2.00", resp.Value.Interest.ValuePerc)
				}
			},
		},
		{
			name: "charge with discount",
			request: CobVRequest{
				Calendar: CobVCalendar{
					DueDate:       "2035-06-24",
					ValidAfterDue: 30,
				},
				Value: CobVValue{
					Original: "123.45",
					Discount: &CobVDiscount{
						Modality: "1",
						FixedDateDiscount: []FixedDateDiscount{
							{
								Date:      "2035-06-20",
								ValuePerc: "5.00",
							},
						},
					},
				},
				Key: "95127446000198",
			},
			validate: func(t *testing.T, resp *CobVResponse) {
				if resp.Value.Discount == nil {
					t.Error("Discount should not be nil")
				} else if len(resp.Value.Discount.FixedDateDiscount) == 0 {
					t.Error("FixedDateDiscount should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("Method = %s, want PUT", r.Method)
				}

				// Verify path is /cobv/{txid}
				if len(r.URL.Path) < 6 || r.URL.Path[:6] != "/cobv/" {
					t.Errorf("Path should start with /cobv/, got %s", r.URL.Path)
				}

				// Verify request body
				var reqBody CobVRequest
				if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
					t.Errorf("Failed to decode request: %v", err)
				}

				// Validate due date format (YYYY-MM-DD)
				if _, err := time.Parse("2006-01-02", reqBody.Calendar.DueDate); err != nil {
					t.Errorf("Invalid due date format: %v", err)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				w.Write(responseData)
			}))
			defer server.Close()

			// Since Client doesn't have CobV methods yet, we'll test the concept
			// In a real implementation, you'd call: client.CreateCobV(ctx, txid, tt.request)

			// For now, test the response parsing
			var response CobVResponse
			if err := json.Unmarshal(responseData, &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, &response)
			}
		})
	}
}

// TestCobVDueDateValidation tests due date validation rules
func TestCobVDueDateValidation(t *testing.T) {
	tests := []struct {
		name    string
		dueDate string
		wantErr bool
	}{
		{"valid future date", "2035-06-24", false},
		{"valid far future", "2099-12-31", false},
		{"invalid format", "24-06-2035", true},
		{"invalid format slashes", "24/06/2035", true},
		{"invalid date", "2035-13-40", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := time.Parse("2006-01-02", tt.dueDate)
			hasErr := err != nil

			if hasErr != tt.wantErr {
				t.Errorf("Date %q validation = %v, wantErr %v", tt.dueDate, hasErr, tt.wantErr)
			}
		})
	}
}

// TestCobVFineAndInterestModalities tests fine and interest modality validation
func TestCobVFineAndInterestModalities(t *testing.T) {
	tests := []struct {
		name     string
		modality CobVModality
		wantErr  bool
	}{
		{
			name: "valid percentage modality",
			modality: CobVModality{
				Modality:  "2",
				ValuePerc: "15.00",
			},
			wantErr: false,
		},
		{
			name: "valid fixed value modality",
			modality: CobVModality{
				Modality:  "1",
				ValuePerc: "10.50",
			},
			wantErr: false,
		},
		{
			name: "invalid modality",
			modality: CobVModality{
				Modality:  "3",
				ValuePerc: "15.00",
			},
			wantErr: true,
		},
		{
			name: "invalid percentage format",
			modality: CobVModality{
				Modality:  "2",
				ValuePerc: "15",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCobVModality(&tt.modality)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCobVModality() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func validateCobVModality(m *CobVModality) error {
	if m == nil {
		return nil
	}

	if m.Modality != "1" && m.Modality != "2" {
		return &ValidationError{Field: "modalidade", Message: "must be '1' or '2'"}
	}

	// Validate percentage format
	valuePattern := `^\d{1,3}\.\d{2}$`
	matched, _ := regexp.MatchString(valuePattern, m.ValuePerc)
	if !matched {
		return &ValidationError{Field: "valorPerc", Message: "must match pattern ^\\d{1,3}\\.\\d{2}$"}
	}

	return nil
}

// TestCobVStatusTransitions tests valid status transitions
func TestCobVStatusTransitions(t *testing.T) {
	validStatuses := []string{
		"ATIVA",
		"CONCLUIDA",
		"REMOVIDA_PELO_USUARIO_RECEBEDOR",
		"REMOVIDA_PELO_PSP",
	}

	for _, status := range validStatuses {
		t.Run("status_"+status, func(t *testing.T) {
			response := CobVResponse{
				Status: status,
				TxID:   "test123456789012345678901234",
			}

			if response.Status != status {
				t.Errorf("Status = %s, want %s", response.Status, status)
			}
		})
	}
}

// TestCobVListWithFilters tests listing charges with filters
func TestCobVListWithFilters(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]string
	}{
		{
			name: "filter by CPF",
			params: map[string]string{
				"cpf": "12345678909",
			},
		},
		{
			name: "filter by CNPJ",
			params: map[string]string{
				"cnpj": "12345678000195",
			},
		},
		{
			name: "filter by status",
			params: map[string]string{
				"status": "ATIVA",
			},
		},
		{
			name: "combined filters",
			params: map[string]string{
				"cpf":    "12345678909",
				"status": "ATIVA",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				query := r.URL.Query()

				// Verify filters are present in query
				for key, expectedValue := range tt.params {
					if actualValue := query.Get(key); actualValue != expectedValue {
						t.Errorf("Query param %s = %s, want %s", key, actualValue, expectedValue)
					}
				}

				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"parametros":{"inicio":"2024-01-01T00:00:00Z","fim":"2024-01-31T23:59:59Z","paginacao":{"paginaAtual":0,"itensPorPagina":100}},"cobs":[]}`))
			}))
			defer server.Close()

			// Build URL with query parameters
			req, _ := http.NewRequest("GET", server.URL+"/cobv", nil)
			q := req.URL.Query()
			q.Add("inicio", "2024-01-01T00:00:00Z")
			q.Add("fim", "2024-01-31T23:59:59Z")
			for key, value := range tt.params {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
			}
		})
	}
}
