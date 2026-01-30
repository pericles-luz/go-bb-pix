package pix

import "fmt"

// CreateRefundRequest represents a request to create a refund
type CreateRefundRequest struct {
	Value  float64 `json:"-"`
	Reason string  `json:"motivo,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for CreateRefundRequest
func (r CreateRefundRequest) MarshalJSON() ([]byte, error) {
	if r.Reason != "" {
		return []byte(fmt.Sprintf(`{"valor":"%.2f","motivo":%q}`, r.Value, r.Reason)), nil
	}
	return []byte(fmt.Sprintf(`{"valor":"%.2f"}`, r.Value)), nil
}

// RefundResponse represents a refund response
type RefundResponse struct {
	ID     string     `json:"id"`
	RtrID  string     `json:"rtrId"`
	Value  string     `json:"valor"`
	Time   RefundTime `json:"horario"`
	Status string     `json:"status"`
	Reason string     `json:"motivo,omitempty"`
}
