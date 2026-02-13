package pix

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// TestQRCodeResponseSchema validates QRCode response against API specification
func TestQRCodeResponseSchema(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "testdata", "cob", "create_response.json"))
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	var response QRCodeResponse
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Validate txid format: 26-35 alphanumeric characters
	txidPattern := regexp.MustCompile(`^[a-zA-Z0-9]{26,35}$`)
	if !txidPattern.MatchString(response.TxID) {
		t.Errorf("TxID %q does not match pattern ^[a-zA-Z0-9]{26,35}$", response.TxID)
	}

	// Validate value format: ^\d{1,10}\.\d{2}$
	valuePattern := regexp.MustCompile(`^\d{1,10}\.\d{2}$`)
	if !valuePattern.MatchString(response.Value.Original) {
		t.Errorf("Value %q does not match pattern ^\\d{1,10}\\.\\d{2}$", response.Value.Original)
	}

	// Validate status
	validStatuses := map[string]bool{
		"ATIVA":      true,
		"CONCLUIDA":  true,
		"REMOVIDA_PELO_USUARIO_RECEBEDOR": true,
		"REMOVIDA_PELO_PSP": true,
	}
	if !validStatuses[response.Status] {
		t.Errorf("Invalid status %q", response.Status)
	}

	// Validate chave (PIX key) max length: 77 characters
	if len(response.Key) > 77 {
		t.Errorf("Chave length %d exceeds maximum of 77", len(response.Key))
	}

	// Validate calendar expiration
	if response.Calendar.Expiration < 0 || response.Calendar.Expiration > 86400 {
		t.Errorf("Invalid expiration %d (should be 0-86400 seconds)", response.Calendar.Expiration)
	}
}

// TestDebtorValidation validates debtor CPF/CNPJ format
func TestDebtorValidation(t *testing.T) {
	tests := []struct {
		name    string
		debtor  Debtor
		wantErr bool
	}{
		{
			name: "valid CPF",
			debtor: Debtor{
				CPF:  "12345678909",
				Name: "Francisco da Silva",
			},
			wantErr: false,
		},
		{
			name: "invalid CPF length",
			debtor: Debtor{
				CPF:  "123456789",
				Name: "Francisco da Silva",
			},
			wantErr: true,
		},
		{
			name: "valid CNPJ",
			debtor: Debtor{
				CNPJ: "12345678000195",
				Name: "Empresa SA",
			},
			wantErr: false,
		},
		{
			name: "invalid CNPJ length",
			debtor: Debtor{
				CNPJ: "123456780001",
				Name: "Empresa SA",
			},
			wantErr: true,
		},
		{
			name: "both CPF and CNPJ",
			debtor: Debtor{
				CPF:  "12345678909",
				CNPJ: "12345678000195",
				Name: "Invalid",
			},
			wantErr: true,
		},
		{
			name: "neither CPF nor CNPJ",
			debtor: Debtor{
				Name: "Invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDebtor(&tt.debtor)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDebtor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// validateDebtor validates debtor information
func validateDebtor(d *Debtor) error {
	if d == nil {
		return nil
	}

	cpfPattern := regexp.MustCompile(`^\d{11}$`)
	cnpjPattern := regexp.MustCompile(`^\d{14}$`)

	hasCPF := d.CPF != ""
	hasCNPJ := d.CNPJ != ""

	if hasCPF && hasCNPJ {
		return &ValidationError{Field: "devedor", Message: "cannot have both CPF and CNPJ"}
	}

	if !hasCPF && !hasCNPJ {
		return &ValidationError{Field: "devedor", Message: "must have either CPF or CNPJ"}
	}

	if hasCPF && !cpfPattern.MatchString(d.CPF) {
		return &ValidationError{Field: "devedor.cpf", Message: "must be 11 digits"}
	}

	if hasCNPJ && !cnpjPattern.MatchString(d.CNPJ) {
		return &ValidationError{Field: "devedor.cnpj", Message: "must be 14 digits"}
	}

	if d.Name == "" {
		return &ValidationError{Field: "devedor.nome", Message: "name is required"}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// TestValueValidation validates monetary value format
func TestValueValidation(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid small", "0.01", false},
		{"valid medium", "123.45", false},
		{"valid large", "9999999999.99", false},
		{"invalid no decimals", "100", true},
		{"invalid one decimal", "100.5", true},
		{"invalid three decimals", "100.555", true},
		{"invalid too large", "12345678901.00", true},
		{"invalid negative", "-10.00", true},
		{"invalid letters", "abc.de", true},
		{"zero valid", "0.00", false},
	}

	valuePattern := regexp.MustCompile(`^\d{1,10}\.\d{2}$`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := valuePattern.MatchString(tt.value)
			if valid == tt.wantErr {
				t.Errorf("Value %q validation = %v, wantErr %v", tt.value, !valid, tt.wantErr)
			}
		})
	}
}

// TestTxIDValidation validates transaction ID format
func TestTxIDValidation(t *testing.T) {
	tests := []struct {
		name    string
		txid    string
		wantErr bool
	}{
		{"valid 26 chars", "fb2761260e554ad593c7226beb", false},
		{"valid 35 chars", "fb2761260e554ad593c7226beb5cb650abc", false},
		{"invalid too short", "fb2761260e554ad593c7226b", true},
		{"invalid too long", "fb2761260e554ad593c7226beb5cb650abcd", true},
		{"invalid special chars", "fb2761260e554ad593c7226@eb", true},
		{"invalid spaces", "fb2761260e554ad593c7 226beb", true},
		{"empty", "", true},
	}

	txidPattern := regexp.MustCompile(`^[a-zA-Z0-9]{26,35}$`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := txidPattern.MatchString(tt.txid)
			if valid == tt.wantErr {
				t.Errorf("TxID %q validation = %v, wantErr %v", tt.txid, !valid, tt.wantErr)
			}
		})
	}
}

// TestEndToEndIDValidation validates E2E ID format
func TestEndToEndIDValidation(t *testing.T) {
	tests := []struct {
		name    string
		e2eID   string
		wantErr bool
	}{
		{"valid", "E12345678202406201221abcdef12345", false},
		{"valid long", "E98765432202406201330fedcba987654321", false},
		{"invalid no E prefix", "12345678202406201221abcdef12345", true},
		{"invalid too short", "E123456", true},
		{"empty", "", true},
	}

	e2ePattern := regexp.MustCompile(`^E[a-zA-Z0-9]+$`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := e2ePattern.MatchString(tt.e2eID) && len(tt.e2eID) >= 32
			if valid == tt.wantErr {
				t.Errorf("E2E ID %q validation = %v, wantErr %v", tt.e2eID, !valid, tt.wantErr)
			}
		})
	}
}

// TestPaymentListResponseSchema validates payment list response schema
func TestPaymentListResponseSchema(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "testdata", "pix", "list_response.json"))
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	var response PaymentListResponse
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Validate pagination
	if response.Parameters.Pagination.CurrentPage < 0 {
		t.Error("Current page cannot be negative")
	}
	if response.Parameters.Pagination.ItemsPerPage < 1 || response.Parameters.Pagination.ItemsPerPage > 500 {
		t.Errorf("Items per page %d should be 1-500", response.Parameters.Pagination.ItemsPerPage)
	}

	// Validate payments
	for i, pmt := range response.Payments {
		// Validate E2E ID
		e2ePattern := regexp.MustCompile(`^E[a-zA-Z0-9]+$`)
		if !e2ePattern.MatchString(pmt.EndToEndID) || len(pmt.EndToEndID) < 32 {
			t.Errorf("Payment[%d] invalid EndToEndID: %s", i, pmt.EndToEndID)
		}

		// Validate value format
		valuePattern := regexp.MustCompile(`^\d{1,10}\.\d{2}$`)
		if !valuePattern.MatchString(pmt.Value) {
			t.Errorf("Payment[%d] invalid value format: %s", i, pmt.Value)
		}

		// Validate refund statuses
		for j, refund := range pmt.Refunds {
			validRefundStatuses := map[string]bool{
				"DEVOLVIDO":        true,
				"EM_PROCESSAMENTO": true,
				"NAO_REALIZADO":    true,
			}
			if !validRefundStatuses[refund.Status] {
				t.Errorf("Payment[%d] Refund[%d] invalid status: %s", i, j, refund.Status)
			}
		}
	}
}

// TestDateTimeRFC3339 validates that all timestamps follow RFC 3339
func TestDateTimeRFC3339(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "testdata", "cob", "create_response.json"))
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	var response QRCodeResponse
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Time should parse without error
	if response.Calendar.Creation.IsZero() {
		t.Error("Calendar creation time is zero")
	}

	// Verify it's in RFC 3339 format
	formatted := response.Calendar.Creation.Format("2006-01-02T15:04:05Z07:00")
	if formatted == "" {
		t.Error("Failed to format time as RFC 3339")
	}
}
