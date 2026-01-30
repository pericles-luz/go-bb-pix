package pix

import "time"

// PaymentResponse represents a PIX payment
type PaymentResponse struct {
	EndToEndID string       `json:"endToEndId"`
	TxID       string       `json:"txid"`
	Value      string       `json:"valor"`
	Time       time.Time    `json:"horario"`
	PayerInfo  string       `json:"infoPagador,omitempty"`
	Refunds    []RefundInfo `json:"devolucoes,omitempty"`
}

// RefundInfo represents information about a refund
type RefundInfo struct {
	ID     string     `json:"id"`
	RtrID  string     `json:"rtrId"`
	Value  string     `json:"valor"`
	Time   RefundTime `json:"horario"`
	Status string     `json:"status"`
	Reason string     `json:"motivo,omitempty"`
}

// RefundTime represents refund timing information
type RefundTime struct {
	Solicitation time.Time `json:"solicitacao"`
	Settlement   time.Time `json:"liquidacao,omitempty"`
}

// ListPaymentsParams represents parameters for listing payments
type ListPaymentsParams struct {
	StartDate time.Time `json:"inicio"`
	EndDate   time.Time `json:"fim"`
	TxID      string    `json:"txid,omitempty"`
	CPF       string    `json:"cpf,omitempty"`
	CNPJ      string    `json:"cnpj,omitempty"`
	Page      int       `json:"paginaAtual,omitempty"`
	PageSize  int       `json:"itensPorPagina,omitempty"`
}

// PaymentListResponse represents a list of payments
type PaymentListResponse struct {
	Parameters struct {
		Start      time.Time  `json:"inicio"`
		End        time.Time  `json:"fim"`
		Pagination Pagination `json:"paginacao"`
	} `json:"parametros"`
	Payments []PaymentResponse `json:"pix"`
}
