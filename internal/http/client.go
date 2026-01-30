package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pericles-luz/go-bb-pix/internal/apierror"
)

// Client is an HTTP client for making API requests
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new HTTP client
func NewClient(httpClient *http.Client, baseURL string) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
	}
}

// NewRequest creates a new HTTP request
func (c *Client) NewRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	// Build URL
	u, err := c.buildURL(path)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	// Encode body if present
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// Do executes the HTTP request and decodes the response into target
// If target is nil, the response body is discarded
func (c *Client) Do(req *http.Request, target interface{}) error {
	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for error status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseErrorResponse(resp.StatusCode, resp.Body)
	}

	// If target is nil, just discard the body
	if target == nil {
		io.Copy(io.Discard, resp.Body)
		return nil
	}

	// Decode response
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// buildURL builds the full URL from base URL and path
func (c *Client) buildURL(path string) (string, error) {
	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Parse base URL
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}

	// Parse path
	ref, err := url.Parse(path)
	if err != nil {
		return "", err
	}

	// Resolve reference
	u := base.ResolveReference(ref)
	return u.String(), nil
}

// errorResponse represents an error response from the API
type errorResponse struct {
	Message string `json:"message"`
	Errors  []struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	} `json:"errors"`
}

// parseErrorResponse parses an error response into an APIError
func parseErrorResponse(statusCode int, body io.Reader) error {
	// Read body
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return apierror.New(statusCode, fmt.Sprintf("HTTP %d", statusCode))
	}

	// Try to parse as JSON
	var errResp errorResponse
	if err := json.Unmarshal(bodyBytes, &errResp); err != nil {
		// Not JSON, use status code
		return apierror.New(statusCode, fmt.Sprintf("HTTP %d", statusCode))
	}

	// Build error details
	var details []apierror.ErrorDetail
	for _, e := range errResp.Errors {
		details = append(details, apierror.ErrorDetail{
			Field:   e.Field,
			Message: e.Message,
		})
	}

	// Use message from response or default to status code
	message := errResp.Message
	if message == "" {
		message = fmt.Sprintf("HTTP %d", statusCode)
	}

	return apierror.New(statusCode, message, details...)
}
