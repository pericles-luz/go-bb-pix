package pix

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCreateQRCodeRequest_Marshal(t *testing.T) {
	req := CreateQRCodeRequest{
		TxID:                   "txid123",
		Value:                  100.50,
		Expiration:             3600,
		PayerSolicitation:      "Pagamento de teste",
		AdditionalInformation:  "Info adicional",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Verify JSON structure
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded["calendario"].(map[string]interface{})["expiracao"] != float64(3600) {
		t.Error("expiracao not marshaled correctly")
	}
	if decoded["valor"].(map[string]interface{})["original"] != "100.50" {
		t.Error("valor not marshaled correctly")
	}
}

func TestQRCodeResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"calendario": {
			"criacao": "2024-01-15T10:00:00Z",
			"expiracao": 3600
		},
		"txid": "txid123",
		"revisao": 1,
		"loc": {
			"id": 123,
			"location": "pix.example.com/qr/v2/123",
			"tipoCob": "cob"
		},
		"location": "pix.example.com/qr/v2/123",
		"status": "ATIVA",
		"devedor": {
			"cpf": "12345678900",
			"nome": "João Silva"
		},
		"valor": {
			"original": "100.50"
		},
		"chave": "chave-pix-123",
		"solicitacaoPagador": "Pagamento de teste",
		"infoAdicionais": [
			{"nome": "info1", "valor": "value1"}
		],
		"pixCopiaECola": "00020126580014br.gov.bcb.pix..."
	}`

	var resp QRCodeResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if resp.TxID != "txid123" {
		t.Errorf("TxID = %s, want txid123", resp.TxID)
	}
	if resp.Status != "ATIVA" {
		t.Errorf("Status = %s, want ATIVA", resp.Status)
	}
	if resp.Value.Original != "100.50" {
		t.Errorf("Value.Original = %s, want 100.50", resp.Value.Original)
	}
	if resp.QRCode != "00020126580014br.gov.bcb.pix..." {
		t.Errorf("QRCode not unmarshaled correctly")
	}
}

func TestListQRCodesParams_BuildQueryString(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	params := ListQRCodesParams{
		StartDate: start,
		EndDate:   end,
		CPF:       "12345678900",
		CNPJ:      "",
		Status:    "ATIVA",
		Page:      2,
		PageSize:  50,
	}

	// Test that params can be marshaled
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
}

func TestUpdateQRCodeRequest_Marshal(t *testing.T) {
	req := UpdateQRCodeRequest{
		Value:       150.00,
		Expiration:  7200,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded["valor"].(map[string]interface{})["original"] != "150.00" {
		t.Error("valor not marshaled correctly")
	}
	if decoded["calendario"].(map[string]interface{})["expiracao"] != float64(7200) {
		t.Error("expiracao not marshaled correctly")
	}
}

func TestQRCodeListResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"parametros": {
			"inicio": "2024-01-01T00:00:00Z",
			"fim": "2024-01-31T23:59:59Z",
			"paginacao": {
				"paginaAtual": 0,
				"itensPorPagina": 100,
				"quantidadeDePaginas": 1,
				"quantidadeTotalDeItens": 2
			}
		},
		"cobs": [
			{
				"calendario": {"criacao": "2024-01-15T10:00:00Z", "expiracao": 3600},
				"txid": "txid1",
				"revisao": 1,
				"status": "ATIVA",
				"valor": {"original": "100.00"}
			},
			{
				"calendario": {"criacao": "2024-01-16T10:00:00Z", "expiracao": 3600},
				"txid": "txid2",
				"revisao": 1,
				"status": "CONCLUIDA",
				"valor": {"original": "200.00"}
			}
		]
	}`

	var resp QRCodeListResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if len(resp.QRCodes) != 2 {
		t.Errorf("len(QRCodes) = %d, want 2", len(resp.QRCodes))
	}

	if resp.Parameters.Pagination.CurrentPage != 0 {
		t.Errorf("CurrentPage = %d, want 0", resp.Parameters.Pagination.CurrentPage)
	}
	if resp.Parameters.Pagination.TotalItems != 2 {
		t.Errorf("TotalItems = %d, want 2", resp.Parameters.Pagination.TotalItems)
	}
}

func TestCalendar_Unmarshal(t *testing.T) {
	jsonData := `{
		"criacao": "2024-01-15T10:30:45Z",
		"expiracao": 3600
	}`

	var calendar Calendar
	if err := json.Unmarshal([]byte(jsonData), &calendar); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if calendar.Expiration != 3600 {
		t.Errorf("Expiration = %d, want 3600", calendar.Expiration)
	}

	// Check that creation time was parsed correctly
	expectedTime := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	if !calendar.Creation.Equal(expectedTime) {
		t.Errorf("Creation = %v, want %v", calendar.Creation, expectedTime)
	}
}

func TestLocation_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": 12345,
		"location": "pix.example.com/qr/v2/12345",
		"tipoCob": "cob"
	}`

	var loc Location
	if err := json.Unmarshal([]byte(jsonData), &loc); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if loc.ID != 12345 {
		t.Errorf("ID = %d, want 12345", loc.ID)
	}
	if loc.Location != "pix.example.com/qr/v2/12345" {
		t.Errorf("Location = %s, want pix.example.com/qr/v2/12345", loc.Location)
	}
	if loc.Type != "cob" {
		t.Errorf("Type = %s, want cob", loc.Type)
	}
}

func TestDebtor_Marshal(t *testing.T) {
	debtor := Debtor{
		CPF:  "12345678900",
		Name: "João Silva",
	}

	data, err := json.Marshal(debtor)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded["cpf"] != "12345678900" {
		t.Error("CPF not marshaled correctly")
	}
	if decoded["nome"] != "João Silva" {
		t.Error("Name not marshaled correctly")
	}
}

func TestDebtor_CNPJ(t *testing.T) {
	debtor := Debtor{
		CNPJ: "12345678000190",
		Name: "Empresa LTDA",
	}

	data, err := json.Marshal(debtor)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded["cnpj"] != "12345678000190" {
		t.Error("CNPJ not marshaled correctly")
	}
}

func TestValue_Marshal(t *testing.T) {
	value := Value{Original: "123.45"}

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded["original"] != "123.45" {
		t.Error("Original not marshaled correctly")
	}
}

func TestAdditionalInfo_Marshal(t *testing.T) {
	info := AdditionalInfo{
		Name:  "campo1",
		Value: "valor1",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded["nome"] != "campo1" {
		t.Error("Name not marshaled correctly")
	}
	if decoded["valor"] != "valor1" {
		t.Error("Value not marshaled correctly")
	}
}
