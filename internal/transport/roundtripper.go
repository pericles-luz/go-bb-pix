package transport

import (
	"fmt"
	"net/http"

	"github.com/pericles-luz/go-bb-pix/internal/auth"
)

// AuthTransport is an http.RoundTripper that injects OAuth2 authentication
type AuthTransport struct {
	base            http.RoundTripper
	tokenProvider   auth.TokenProvider
	developerAppKey string
}

// NewAuthTransport creates a new AuthTransport
func NewAuthTransport(base http.RoundTripper, provider auth.TokenProvider, developerAppKey string) *AuthTransport {
	if base == nil {
		base = http.DefaultTransport
	}

	return &AuthTransport{
		base:            base,
		tokenProvider:   provider,
		developerAppKey: developerAppKey,
	}
}

// RoundTrip implements http.RoundTripper
func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Get token
	token, err := t.tokenProvider.GetToken(req.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	// Clone request to avoid modifying the original
	req = cloneRequest(req)

	// Add Authorization header
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", token.TokenType, token.AccessToken))

	// Add Developer Application Key header
	req.Header.Set("gw-dev-app-key", t.developerAppKey)

	// Execute request
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// If we get 401, invalidate the token for next request
	if resp.StatusCode == http.StatusUnauthorized {
		t.tokenProvider.Invalidate()
	}

	return resp, nil
}

// cloneRequest creates a shallow copy of the request with a cloned header
func cloneRequest(req *http.Request) *http.Request {
	r := new(http.Request)
	*r = *req

	// Deep copy headers
	r.Header = make(http.Header, len(req.Header))
	for k, v := range req.Header {
		r.Header[k] = append([]string(nil), v...)
	}

	return r
}
