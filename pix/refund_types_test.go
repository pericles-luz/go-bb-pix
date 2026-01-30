package pix

import (
	"encoding/json"
	"testing"
)

func TestCreateRefundRequest_Marshal(t *testing.T) {
	req := CreateRefundRequest{
		Value: 50.25,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded["valor"] != "50.25" {
		t.Errorf("valor = %v, want 50.25", decoded["valor"])
	}
}

func TestCreateRefundRequest_MarshalWithReason(t *testing.T) {
	req := CreateRefundRequest{
		Value:  100.00,
		Reason: "Cliente solicitou devolução",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded["valor"] != "100.00" {
		t.Errorf("valor = %v, want 100.00", decoded["valor"])
	}
	if decoded["motivo"] != "Cliente solicitou devolução" {
		t.Errorf("motivo = %v, want 'Cliente solicitou devolução'", decoded["motivo"])
	}
}

func TestRefundResponse_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": "refund123",
		"rtrId": "D12345678202401151000000000001",
		"valor": "75.50",
		"horario": {
			"solicitacao": "2024-01-15T11:00:00Z",
			"liquidacao": "2024-01-15T11:01:00Z"
		},
		"status": "DEVOLVIDO",
		"motivo": "Devolução solicitada pelo cliente"
	}`

	var resp RefundResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if resp.ID != "refund123" {
		t.Errorf("ID = %s, want refund123", resp.ID)
	}
	if resp.RtrID != "D12345678202401151000000000001" {
		t.Errorf("RtrID = %s", resp.RtrID)
	}
	if resp.Value != "75.50" {
		t.Errorf("Value = %s, want 75.50", resp.Value)
	}
	if resp.Status != "DEVOLVIDO" {
		t.Errorf("Status = %s, want DEVOLVIDO", resp.Status)
	}
	if resp.Reason != "Devolução solicitada pelo cliente" {
		t.Errorf("Reason = %s", resp.Reason)
	}
}

func TestRefundResponse_UnmarshalPending(t *testing.T) {
	jsonData := `{
		"id": "refund456",
		"rtrId": "D12345678202401151000000000002",
		"valor": "50.00",
		"horario": {
			"solicitacao": "2024-01-15T12:00:00Z"
		},
		"status": "EM_PROCESSAMENTO"
	}`

	var resp RefundResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if resp.Status != "EM_PROCESSAMENTO" {
		t.Errorf("Status = %s, want EM_PROCESSAMENTO", resp.Status)
	}

	// Settlement time should be zero value when not present
	if !resp.Time.Settlement.IsZero() {
		t.Error("Settlement should be zero when not present")
	}
}

func TestCreateRefundRequest_ZeroValue(t *testing.T) {
	req := CreateRefundRequest{
		Value: 0,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded["valor"] != "0.00" {
		t.Errorf("valor = %v, want 0.00", decoded["valor"])
	}
}
