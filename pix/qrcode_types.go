package pix

import (
	"fmt"
	"time"
)

// CreateQRCodeRequest represents a request to create a QR Code
type CreateQRCodeRequest struct {
	TxID                  string  `json:"txid,omitempty"`
	Value                 float64 `json:"-"`
	Expiration            int     `json:"-"`
	PayerSolicitation     string  `json:"-"`
	AdditionalInformation string  `json:"-"`
	Debtor                *Debtor `json:"devedor,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for CreateQRCodeRequest
func (r CreateQRCodeRequest) MarshalJSON() ([]byte, error) {
	type Alias CreateQRCodeRequest
	return []byte(fmt.Sprintf(`{
		"calendario": {"expiracao": %d},
		"valor": {"original": "%.2f"},
		"chave": "",
		"solicitacaoPagador": %q,
		"infoAdicionais": [{"nome": "info", "valor": %q}]
	}`, r.Expiration, r.Value, r.PayerSolicitation, r.AdditionalInformation)), nil
}

// UpdateQRCodeRequest represents a request to update a QR Code
type UpdateQRCodeRequest struct {
	Value      float64 `json:"-"`
	Expiration int     `json:"-"`
}

// MarshalJSON implements custom JSON marshaling for UpdateQRCodeRequest
func (r UpdateQRCodeRequest) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{
		"calendario": {"expiracao": %d},
		"valor": {"original": "%.2f"}
	}`, r.Expiration, r.Value)), nil
}

// QRCodeResponse represents a QR Code response from the API
type QRCodeResponse struct {
	Calendar              Calendar         `json:"calendario"`
	TxID                  string           `json:"txid"`
	Revision              int              `json:"revisao"`
	Loc                   *Location        `json:"loc,omitempty"`
	Location              string           `json:"location,omitempty"`
	Status                string           `json:"status"`
	Debtor                *Debtor          `json:"devedor,omitempty"`
	Value                 Value            `json:"valor"`
	Key                   string           `json:"chave,omitempty"`
	PayerSolicitation     string           `json:"solicitacaoPagador,omitempty"`
	AdditionalInformation []AdditionalInfo `json:"infoAdicionais,omitempty"`
	QRCode                string           `json:"pixCopiaECola,omitempty"`
}

// Calendar represents the calendar information of a QR Code
type Calendar struct {
	Creation   time.Time `json:"criacao"`
	Expiration int       `json:"expiracao"`
}

// Location represents the location information of a QR Code
type Location struct {
	ID       int    `json:"id"`
	Location string `json:"location"`
	Type     string `json:"tipoCob"`
}

// Debtor represents debtor information
type Debtor struct {
	CPF  string `json:"cpf,omitempty"`
	CNPJ string `json:"cnpj,omitempty"`
	Name string `json:"nome"`
}

// Value represents monetary value
type Value struct {
	Original string `json:"original"`
}

// AdditionalInfo represents additional information
type AdditionalInfo struct {
	Name  string `json:"nome"`
	Value string `json:"valor"`
}

// ListQRCodesParams represents parameters for listing QR Codes
type ListQRCodesParams struct {
	StartDate time.Time `json:"inicio"`
	EndDate   time.Time `json:"fim"`
	CPF       string    `json:"cpf,omitempty"`
	CNPJ      string    `json:"cnpj,omitempty"`
	Status    string    `json:"status,omitempty"`
	Page      int       `json:"paginaAtual,omitempty"`
	PageSize  int       `json:"itensPorPagina,omitempty"`
}

// QRCodeListResponse represents a list of QR Codes
type QRCodeListResponse struct {
	Parameters struct {
		Start      time.Time  `json:"inicio"`
		End        time.Time  `json:"fim"`
		Pagination Pagination `json:"paginacao"`
	} `json:"parametros"`
	QRCodes []QRCodeResponse `json:"cobs"`
}

// Pagination represents pagination information
type Pagination struct {
	CurrentPage  int `json:"paginaAtual"`
	ItemsPerPage int `json:"itensPorPagina"`
	TotalPages   int `json:"quantidadeDePaginas"`
	TotalItems   int `json:"quantidadeTotalDeItens"`
}
