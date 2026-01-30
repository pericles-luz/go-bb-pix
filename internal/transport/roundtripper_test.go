package transport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pericles-luz/go-bb-pix/internal/auth"
)

// mockTokenProvider implements auth.TokenProvider for testing
type mockTokenProvider struct {
	token           *auth.Token
	err             error
	invalidateCount int
}

func (m *mockTokenProvider) GetToken(ctx context.Context) (*auth.Token, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.token, nil
}

func (m *mockTokenProvider) Invalidate() {
	m.invalidateCount++
}

// mockRoundTripper implements http.RoundTripper for testing
type mockRoundTripper struct {
	roundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.roundTripFunc != nil {
		return m.roundTripFunc(req)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       http.NoBody,
		Header:     make(http.Header),
	}, nil
}

func TestNewAuthTransport(t *testing.T) {
	provider := &mockTokenProvider{
		token: &auth.Token{AccessToken: "test-token"},
	}
	base := &mockRoundTripper{}
	appKey := "test-app-key"

	transport := NewAuthTransport(base, provider, appKey)

	if transport == nil {
		t.Fatal("NewAuthTransport returned nil")
	}
	if transport.base != base {
		t.Error("base transport not set correctly")
	}
	if transport.tokenProvider != provider {
		t.Error("token provider not set correctly")
	}
	if transport.developerAppKey != appKey {
		t.Error("developer app key not set correctly")
	}
}

func TestAuthTransport_RoundTrip_AddsAuthHeaders(t *testing.T) {
	provider := &mockTokenProvider{
		token: &auth.Token{
			AccessToken: "test-access-token",
			TokenType:   "Bearer",
		},
	}

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Verify Authorization header
			authHeader := req.Header.Get("Authorization")
			expectedAuth := "Bearer test-access-token"
			if authHeader != expectedAuth {
				t.Errorf("Authorization header = %q, want %q", authHeader, expectedAuth)
			}

			// Verify Developer App Key header
			appKeyHeader := req.Header.Get("gw-dev-app-key")
			expectedAppKey := "test-app-key"
			if appKeyHeader != expectedAppKey {
				t.Errorf("gw-dev-app-key header = %q, want %q", appKeyHeader, expectedAppKey)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Header:     make(http.Header),
			}, nil
		},
	}

	transport := NewAuthTransport(base, provider, "test-app-key")

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	_, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
}

func TestAuthTransport_RoundTrip_TokenError(t *testing.T) {
	provider := &mockTokenProvider{
		err: errors.New("token fetch failed"),
	}

	base := &mockRoundTripper{}
	transport := NewAuthTransport(base, provider, "test-app-key")

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	_, err := transport.RoundTrip(req)

	if err == nil {
		t.Fatal("Expected error when token fetch fails, got nil")
	}
}

func TestAuthTransport_RoundTrip_Invalidates401(t *testing.T) {
	provider := &mockTokenProvider{
		token: &auth.Token{AccessToken: "test-token"},
	}

	callCount := 0
	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			// Return 401 on first call
			if callCount == 1 {
				return &http.Response{
					StatusCode: http.StatusUnauthorized,
					Body:       http.NoBody,
					Header:     make(http.Header),
				}, nil
			}
			// Success on retry
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Header:     make(http.Header),
			}, nil
		},
	}

	transport := NewAuthTransport(base, provider, "test-app-key")

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}

	// Should invalidate token on 401
	if provider.invalidateCount != 1 {
		t.Errorf("Invalidate called %d times, want 1", provider.invalidateCount)
	}

	// Should return the 401 response (not retry)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestAuthTransport_RoundTrip_PropagatesErrors(t *testing.T) {
	provider := &mockTokenProvider{
		token: &auth.Token{AccessToken: "test-token"},
	}

	expectedErr := errors.New("network error")
	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return nil, expectedErr
		},
	}

	transport := NewAuthTransport(base, provider, "test-app-key")

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	_, err := transport.RoundTrip(req)

	if err == nil {
		t.Fatal("Expected error to be propagated, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}

func TestAuthTransport_RoundTrip_PreservesExistingHeaders(t *testing.T) {
	provider := &mockTokenProvider{
		token: &auth.Token{AccessToken: "test-token"},
	}

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Verify existing headers are preserved
			if req.Header.Get("User-Agent") != "custom-agent" {
				t.Error("User-Agent header was not preserved")
			}
			if req.Header.Get("Content-Type") != "application/json" {
				t.Error("Content-Type header was not preserved")
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Header:     make(http.Header),
			}, nil
		},
	}

	transport := NewAuthTransport(base, provider, "test-app-key")

	req := httptest.NewRequest(http.MethodPost, "http://example.com", nil)
	req.Header.Set("User-Agent", "custom-agent")
	req.Header.Set("Content-Type", "application/json")

	_, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
}

func TestAuthTransport_RoundTrip_UsesRequestContext(t *testing.T) {
	provider := &mockTokenProvider{
		token: &auth.Token{AccessToken: "test-token"},
	}

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Verify context is from request
			if req.Context() == context.Background() {
				t.Error("Expected custom context, got Background")
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Header:     make(http.Header),
			}, nil
		},
	}

	transport := NewAuthTransport(base, provider, "test-app-key")

	// Create request with custom context
	ctx := context.WithValue(context.Background(), "test-key", "test-value")
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req = req.WithContext(ctx)

	_, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
}
