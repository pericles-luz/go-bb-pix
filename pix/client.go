package pix

import (
	"net/http"

	httpclient "github.com/pericles-luz/go-bb-pix/internal/http"
)

// Client is the PIX API client
type Client struct {
	http *httpclient.Client
}

// NewClient creates a new PIX client
func NewClient(httpClient *http.Client, apiURL string) *Client {
	return &Client{
		http: httpclient.NewClient(httpClient, apiURL),
	}
}
