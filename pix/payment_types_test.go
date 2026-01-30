package pix

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPaymentResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"endToEndId": "E12345678202401151000000000001",
		"txid": "txid123",
		"valor": "100.50",
		"horario": "2024-01-15T10:30:45Z",
		"infoPagador": "Informação do pagador",
		"devolucoes": [
			{
				"id": "dev123",
				"rtrId": "D12345678202401151000000000001",
				"valor": "50.00",
				"horario": {
					"solicitacao": "2024-01-15T11:00:00Z",
					"liquidacao": "2024-01-15T11:01:00Z"
				},
				"status": "DEVOLVIDO"
			}
		]
	}`

	var resp PaymentResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if resp.EndToEndID != "E12345678202401151000000000001" {
		t.Errorf("EndToEndID = %s, want E12345678202401151000000000001", resp.EndToEndID)
	}
	if resp.TxID != "txid123" {
		t.Errorf("TxID = %s, want txid123", resp.TxID)
	}
	if resp.Value != "100.50" {
		t.Errorf("Value = %s, want 100.50", resp.Value)
	}
	if resp.PayerInfo != "Informação do pagador" {
		t.Errorf("PayerInfo = %s, want 'Informação do pagador'", resp.PayerInfo)
	}

	expectedTime := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	if !resp.Time.Equal(expectedTime) {
		t.Errorf("Time = %v, want %v", resp.Time, expectedTime)
	}

	if len(resp.Refunds) != 1 {
		t.Fatalf("len(Refunds) = %d, want 1", len(resp.Refunds))
	}

	refund := resp.Refunds[0]
	if refund.ID != "dev123" {
		t.Errorf("Refund.ID = %s, want dev123", refund.ID)
	}
	if refund.Value != "50.00" {
		t.Errorf("Refund.Value = %s, want 50.00", refund.Value)
	}
	if refund.Status != "DEVOLVIDO" {
		t.Errorf("Refund.Status = %s, want DEVOLVIDO", refund.Status)
	}
}

func TestListPaymentsParams_Marshal(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	params := ListPaymentsParams{
		StartDate: start,
		EndDate:   end,
		TxID:      "txid123",
		CPF:       "12345678900",
		Page:      0,
		PageSize:  100,
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Verify fields are present
	if decoded["inicio"] == nil {
		t.Error("inicio not marshaled")
	}
	if decoded["fim"] == nil {
		t.Error("fim not marshaled")
	}
}

func TestPaymentListResponse_Unmarshal(t *testing.T) {
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
		"pix": [
			{
				"endToEndId": "E12345678202401151000000000001",
				"txid": "txid1",
				"valor": "100.00",
				"horario": "2024-01-15T10:00:00Z"
			},
			{
				"endToEndId": "E12345678202401151000000000002",
				"txid": "txid2",
				"valor": "200.00",
				"horario": "2024-01-16T10:00:00Z"
			}
		]
	}`

	var resp PaymentListResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if len(resp.Payments) != 2 {
		t.Errorf("len(Payments) = %d, want 2", len(resp.Payments))
	}

	if resp.Parameters.Pagination.TotalItems != 2 {
		t.Errorf("TotalItems = %d, want 2", resp.Parameters.Pagination.TotalItems)
	}

	// Verify first payment
	if resp.Payments[0].EndToEndID != "E12345678202401151000000000001" {
		t.Errorf("Payment[0].EndToEndID = %s", resp.Payments[0].EndToEndID)
	}
	if resp.Payments[0].Value != "100.00" {
		t.Errorf("Payment[0].Value = %s, want 100.00", resp.Payments[0].Value)
	}
}

func TestRefundInfo_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": "refund123",
		"rtrId": "D12345678202401151000000000001",
		"valor": "75.50",
		"horario": {
			"solicitacao": "2024-01-15T11:00:00Z",
			"liquidacao": "2024-01-15T11:01:00Z"
		},
		"status": "DEVOLVIDO",
		"motivo": "Cliente solicitou"
	}`

	var refund RefundInfo
	if err := json.Unmarshal([]byte(jsonData), &refund); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if refund.ID != "refund123" {
		t.Errorf("ID = %s, want refund123", refund.ID)
	}
	if refund.RtrID != "D12345678202401151000000000001" {
		t.Errorf("RtrID = %s", refund.RtrID)
	}
	if refund.Value != "75.50" {
		t.Errorf("Value = %s, want 75.50", refund.Value)
	}
	if refund.Status != "DEVOLVIDO" {
		t.Errorf("Status = %s, want DEVOLVIDO", refund.Status)
	}

	expectedSolicitation := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)
	if !refund.Time.Solicitation.Equal(expectedSolicitation) {
		t.Errorf("Time.Solicitation = %v, want %v", refund.Time.Solicitation, expectedSolicitation)
	}

	expectedSettlement := time.Date(2024, 1, 15, 11, 1, 0, 0, time.UTC)
	if !refund.Time.Settlement.Equal(expectedSettlement) {
		t.Errorf("Time.Settlement = %v, want %v", refund.Time.Settlement, expectedSettlement)
	}
}

func TestRefundTime_Unmarshal(t *testing.T) {
	jsonData := `{
		"solicitacao": "2024-01-15T11:00:00Z",
		"liquidacao": "2024-01-15T11:01:00Z"
	}`

	var refundTime RefundTime
	if err := json.Unmarshal([]byte(jsonData), &refundTime); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	expectedSolicitation := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)
	if !refundTime.Solicitation.Equal(expectedSolicitation) {
		t.Errorf("Solicitation = %v, want %v", refundTime.Solicitation, expectedSolicitation)
	}

	expectedSettlement := time.Date(2024, 1, 15, 11, 1, 0, 0, time.UTC)
	if !refundTime.Settlement.Equal(expectedSettlement) {
		t.Errorf("Settlement = %v, want %v", refundTime.Settlement, expectedSettlement)
	}
}

func TestPaymentResponse_NoRefunds(t *testing.T) {
	jsonData := `{
		"endToEndId": "E12345678202401151000000000001",
		"txid": "txid123",
		"valor": "100.50",
		"horario": "2024-01-15T10:30:45Z"
	}`

	var resp PaymentResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if len(resp.Refunds) != 0 {
		t.Errorf("len(Refunds) = %d, want 0", len(resp.Refunds))
	}
}
