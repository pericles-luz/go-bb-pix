package transport

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewCircuitBreakerTransport(t *testing.T) {
	base := &mockRoundTripper{}
	maxFailures := 5
	resetTimeout := 10 * time.Second

	transport := NewCircuitBreakerTransport(base, maxFailures, resetTimeout)

	if transport == nil {
		t.Fatal("NewCircuitBreakerTransport returned nil")
	}
	if transport.base != base {
		t.Error("base transport not set correctly")
	}
	if transport.breaker == nil {
		t.Error("circuit breaker not initialized")
	}
}

func TestCircuitBreaker_ClosedState_AllowsRequests(t *testing.T) {
	callCount := 0
	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Header:     make(http.Header),
			}, nil
		},
	}

	transport := NewCircuitBreakerTransport(base, 3, 1*time.Second)

	// Make multiple successful requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
		resp, err := transport.RoundTrip(req)

		if err != nil {
			t.Fatalf("Request %d: unexpected error: %v", i, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Request %d: StatusCode = %d, want %d", i, resp.StatusCode, http.StatusOK)
		}
	}

	if callCount != 5 {
		t.Errorf("callCount = %d, want 5", callCount)
	}
}

func TestCircuitBreaker_OpensAfterMaxFailures(t *testing.T) {
	callCount := 0
	failErr := errors.New("service unavailable")

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			return nil, failErr
		},
	}

	maxFailures := 3
	transport := NewCircuitBreakerTransport(base, maxFailures, 1*time.Second)

	// Make requests until circuit opens
	for i := 0; i < maxFailures; i++ {
		req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
		transport.RoundTrip(req)
	}

	// Next request should fail immediately without calling base transport
	initialCallCount := callCount
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	_, err := transport.RoundTrip(req)

	if err == nil {
		t.Fatal("Expected error when circuit is open, got nil")
	}

	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("Expected ErrCircuitOpen, got %v", err)
	}

	if callCount != initialCallCount {
		t.Error("Circuit breaker should not call base transport when open")
	}
}

func TestCircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	callCount := 0
	failErr := errors.New("service unavailable")

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			// Start succeeding after circuit opens
			if callCount > 3 {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       http.NoBody,
					Header:     make(http.Header),
				}, nil
			}
			return nil, failErr
		},
	}

	maxFailures := 3
	resetTimeout := 100 * time.Millisecond
	transport := NewCircuitBreakerTransport(base, maxFailures, resetTimeout)

	// Trigger circuit opening
	for i := 0; i < maxFailures; i++ {
		req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
		transport.RoundTrip(req)
	}

	// Verify circuit is open
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	_, err := transport.RoundTrip(req)
	if !errors.Is(err, ErrCircuitOpen) {
		t.Error("Circuit should be open")
	}

	// Wait for reset timeout
	time.Sleep(resetTimeout + 50*time.Millisecond)

	// Circuit should be half-open and allow one request
	req = httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("Request in half-open state failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Should have made 4 calls: 3 to open + 1 in half-open
	if callCount != 4 {
		t.Errorf("callCount = %d, want 4", callCount)
	}
}

func TestCircuitBreaker_ClosesAfterSuccessInHalfOpen(t *testing.T) {
	callCount := 0
	failErr := errors.New("service unavailable")

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			// Fail first 3, then succeed
			if callCount <= 3 {
				return nil, failErr
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Header:     make(http.Header),
			}, nil
		},
	}

	maxFailures := 3
	resetTimeout := 100 * time.Millisecond
	transport := NewCircuitBreakerTransport(base, maxFailures, resetTimeout)

	// Open circuit
	for i := 0; i < maxFailures; i++ {
		req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
		transport.RoundTrip(req)
	}

	// Wait for half-open
	time.Sleep(resetTimeout + 50*time.Millisecond)

	// Make successful request in half-open
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("Request in half-open failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("Request should succeed")
	}

	// Circuit should be closed now, make another request immediately
	req = httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err = transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("Request after closing failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("Request should succeed in closed state")
	}

	// Should have made 5 calls total
	if callCount != 5 {
		t.Errorf("callCount = %d, want 5 (3 failures + 1 half-open success + 1 closed)", callCount)
	}
}

func TestCircuitBreaker_ReopensOnFailureInHalfOpen(t *testing.T) {
	callCount := 0
	failErr := errors.New("service unavailable")

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			return nil, failErr
		},
	}

	maxFailures := 3
	resetTimeout := 100 * time.Millisecond
	transport := NewCircuitBreakerTransport(base, maxFailures, resetTimeout)

	// Open circuit
	for i := 0; i < maxFailures; i++ {
		req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
		transport.RoundTrip(req)
	}

	// Wait for half-open
	time.Sleep(resetTimeout + 50*time.Millisecond)

	// Make failing request in half-open
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	_, err := transport.RoundTrip(req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Circuit should be open again
	req = httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	_, err = transport.RoundTrip(req)

	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("Circuit should be open again after failure in half-open, got error: %v", err)
	}

	// Should have made 4 calls: 3 to open + 1 failed in half-open
	if callCount != 4 {
		t.Errorf("callCount = %d, want 4", callCount)
	}
}

func TestCircuitBreaker_SuccessResetsFailureCount(t *testing.T) {
	callCount := 0
	failErr := errors.New("service error")

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			// Fail on calls 1, 2, 4, 5
			// Success on call 3
			if callCount == 3 {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       http.NoBody,
					Header:     make(http.Header),
				}, nil
			}
			return nil, failErr
		},
	}

	maxFailures := 3
	transport := NewCircuitBreakerTransport(base, maxFailures, 1*time.Second)

	// Fail twice
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
		transport.RoundTrip(req)
	}

	// Succeed once (should reset failure count)
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatal("Expected success on third call")
	}

	// Fail twice more (shouldn't open circuit yet because counter was reset)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
		transport.RoundTrip(req)
	}

	// Circuit should still be closed (only 2 consecutive failures)
	req = httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	_, err = transport.RoundTrip(req)

	// Should get the actual error, not ErrCircuitOpen
	if errors.Is(err, ErrCircuitOpen) {
		t.Error("Circuit should not be open (success should have reset counter)")
	}
}

func TestCircuitBreaker_5xxErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		shouldFail bool
	}{
		{"500 Internal Server Error", http.StatusInternalServerError, true},
		{"502 Bad Gateway", http.StatusBadGateway, true},
		{"503 Service Unavailable", http.StatusServiceUnavailable, true},
		{"504 Gateway Timeout", http.StatusGatewayTimeout, true},
		{"200 OK", http.StatusOK, false},
		{"400 Bad Request", http.StatusBadRequest, false},
		{"404 Not Found", http.StatusNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: tt.statusCode,
						Body:       http.NoBody,
						Header:     make(http.Header),
					}, nil
				},
			}

			maxFailures := 2
			transport := NewCircuitBreakerTransport(base, maxFailures, 1*time.Second)

			// Make maxFailures requests
			for i := 0; i < maxFailures; i++ {
				req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
				transport.RoundTrip(req)
			}

			// Check if circuit opened
			req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
			_, err := transport.RoundTrip(req)

			isOpen := errors.Is(err, ErrCircuitOpen)

			if tt.shouldFail && !isOpen {
				t.Errorf("Circuit should be open for %d status", tt.statusCode)
			}
			if !tt.shouldFail && isOpen {
				t.Errorf("Circuit should not be open for %d status", tt.statusCode)
			}
		})
	}
}
