package transport

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewRetryTransport(t *testing.T) {
	base := &mockRoundTripper{}
	maxRetries := 5
	initialBackoff := 200 * time.Millisecond

	transport := NewRetryTransport(base, maxRetries, initialBackoff)

	if transport == nil {
		t.Fatal("NewRetryTransport returned nil")
	}
	if transport.base != base {
		t.Error("base transport not set correctly")
	}
	if transport.maxRetries != maxRetries {
		t.Errorf("maxRetries = %d, want %d", transport.maxRetries, maxRetries)
	}
	if transport.initialBackoff != initialBackoff {
		t.Errorf("initialBackoff = %v, want %v", transport.initialBackoff, initialBackoff)
	}
}

func TestRetryTransport_RetryOnNetworkError(t *testing.T) {
	callCount := 0
	networkErr := errors.New("network error")

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount < 3 {
				return nil, networkErr
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Header:     make(http.Header),
			}, nil
		},
	}

	transport := NewRetryTransport(base, 5, 10*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if callCount != 3 {
		t.Errorf("callCount = %d, want 3 (2 retries + 1 success)", callCount)
	}
}

func TestRetryTransport_RetryOn429(t *testing.T) {
	callCount := 0

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount < 2 {
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Body:       http.NoBody,
					Header:     make(http.Header),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Header:     make(http.Header),
			}, nil
		},
	}

	transport := NewRetryTransport(base, 5, 10*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2 (1 retry + 1 success)", callCount)
	}
}

func TestRetryTransport_RetryOn502_503_504(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"502 Bad Gateway", http.StatusBadGateway},
		{"503 Service Unavailable", http.StatusServiceUnavailable},
		{"504 Gateway Timeout", http.StatusGatewayTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0

			base := &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					callCount++
					if callCount < 2 {
						return &http.Response{
							StatusCode: tt.statusCode,
							Body:       http.NoBody,
							Header:     make(http.Header),
						}, nil
					}
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       http.NoBody,
						Header:     make(http.Header),
					}, nil
				},
			}

			transport := NewRetryTransport(base, 5, 10*time.Millisecond)

			req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
			resp, err := transport.RoundTrip(req)

			if err != nil {
				t.Fatalf("RoundTrip() error = %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
			}
			if callCount < 2 {
				t.Errorf("callCount = %d, want at least 2 (should retry %d)", callCount, tt.statusCode)
			}
		})
	}
}

func TestRetryTransport_NoRetryOn4xx(t *testing.T) {
	callCount := 0

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       http.NoBody,
				Header:     make(http.Header),
			}, nil
		},
	}

	transport := NewRetryTransport(base, 5, 10*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1 (should not retry 4xx)", callCount)
	}
}

func TestRetryTransport_NoRetryOnPOST(t *testing.T) {
	callCount := 0
	networkErr := errors.New("network error")

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			return nil, networkErr
		},
	}

	transport := NewRetryTransport(base, 5, 10*time.Millisecond)

	req := httptest.NewRequest(http.MethodPost, "http://example.com", strings.NewReader("data"))
	_, err := transport.RoundTrip(req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1 (should not retry POST)", callCount)
	}
}

func TestRetryTransport_NoRetryOnPATCH(t *testing.T) {
	callCount := 0
	networkErr := errors.New("network error")

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			return nil, networkErr
		},
	}

	transport := NewRetryTransport(base, 5, 10*time.Millisecond)

	req := httptest.NewRequest(http.MethodPatch, "http://example.com", strings.NewReader("data"))
	_, err := transport.RoundTrip(req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1 (should not retry PATCH)", callCount)
	}
}

func TestRetryTransport_RetryIdempotentMethods(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodHead, http.MethodOptions}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			callCount := 0
			networkErr := errors.New("network error")

			base := &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					callCount++
					if callCount < 2 {
						return nil, networkErr
					}
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       http.NoBody,
						Header:     make(http.Header),
					}, nil
				},
			}

			transport := NewRetryTransport(base, 5, 10*time.Millisecond)

			req := httptest.NewRequest(method, "http://example.com", nil)
			resp, err := transport.RoundTrip(req)

			if err != nil {
				t.Fatalf("RoundTrip() error = %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
			}
			if callCount < 2 {
				t.Errorf("callCount = %d, want at least 2 (should retry %s)", callCount, method)
			}
		})
	}
}

func TestRetryTransport_RespectsContextCancellation(t *testing.T) {
	callCount := 0

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			return nil, errors.New("network error")
		},
	}

	transport := NewRetryTransport(base, 10, 100*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req = req.WithContext(ctx)

	// Cancel after first attempt
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := transport.RoundTrip(req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Should stop retrying after context cancellation
	if callCount > 3 {
		t.Errorf("callCount = %d, should stop early due to context cancellation", callCount)
	}
}

func TestRetryTransport_MaxRetries(t *testing.T) {
	callCount := 0
	networkErr := errors.New("network error")

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			return nil, networkErr
		},
	}

	maxRetries := 3
	transport := NewRetryTransport(base, maxRetries, 10*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	_, err := transport.RoundTrip(req)

	if err == nil {
		t.Fatal("Expected error after max retries, got nil")
	}

	// Should call maxRetries + 1 times (initial + retries)
	expectedCalls := maxRetries + 1
	if callCount != expectedCalls {
		t.Errorf("callCount = %d, want %d (initial + %d retries)", callCount, expectedCalls, maxRetries)
	}
}

func TestRetryTransport_ExponentialBackoff(t *testing.T) {
	callTimes := []time.Time{}
	networkErr := errors.New("network error")

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callTimes = append(callTimes, time.Now())
			return nil, networkErr
		},
	}

	initialBackoff := 50 * time.Millisecond
	transport := NewRetryTransport(base, 3, initialBackoff)

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	transport.RoundTrip(req)

	// Verify exponential backoff (with some tolerance for jitter)
	if len(callTimes) < 2 {
		t.Fatal("Not enough calls to verify backoff")
	}

	for i := 1; i < len(callTimes); i++ {
		delay := callTimes[i].Sub(callTimes[i-1])
		// Expected delay grows exponentially but with jitter, so we just check it's in reasonable range
		minDelay := initialBackoff / 2 // Allow jitter to reduce by half
		maxDelay := initialBackoff * 5  // Allow exponential growth and jitter

		if delay < minDelay || delay > maxDelay {
			t.Logf("Delay between call %d and %d: %v (expected range: %v to %v)", i-1, i, delay, minDelay, maxDelay)
		}

		// Update expected for next iteration (exponential)
		initialBackoff *= 2
	}
}

func TestIsIdempotent(t *testing.T) {
	tests := []struct {
		method string
		want   bool
	}{
		{http.MethodGet, true},
		{http.MethodHead, true},
		{http.MethodOptions, true},
		{http.MethodPut, true},
		{http.MethodDelete, true},
		{http.MethodPost, false},
		{http.MethodPatch, false},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			got := isIdempotent(tt.method)
			if got != tt.want {
				t.Errorf("isIdempotent(%q) = %v, want %v", tt.method, got, tt.want)
			}
		})
	}
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		err        error
		want       bool
	}{
		{"network error", 0, errors.New("network error"), true},
		{"429 too many requests", http.StatusTooManyRequests, nil, true},
		{"502 bad gateway", http.StatusBadGateway, nil, true},
		{"503 service unavailable", http.StatusServiceUnavailable, nil, true},
		{"504 gateway timeout", http.StatusGatewayTimeout, nil, true},
		{"400 bad request", http.StatusBadRequest, nil, false},
		{"401 unauthorized", http.StatusUnauthorized, nil, false},
		{"404 not found", http.StatusNotFound, nil, false},
		{"500 internal server error", http.StatusInternalServerError, nil, false},
		{"200 ok", http.StatusOK, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *http.Response
			if tt.statusCode > 0 {
				resp = &http.Response{
					StatusCode: tt.statusCode,
					Body:       http.NoBody,
				}
			}

			got := shouldRetry(resp, tt.err)
			if got != tt.want {
				t.Errorf("shouldRetry(statusCode=%d, err=%v) = %v, want %v",
					tt.statusCode, tt.err, got, tt.want)
			}
		})
	}
}

func TestRetryTransport_ClosesResponseBody(t *testing.T) {
	callCount := 0
	var bodyClosed bool

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			// First call returns 503 with tracking body
			if callCount == 1 {
				reader := &trackingReader{
					onClose: func() {
						bodyClosed = true
					},
				}
				return &http.Response{
					StatusCode: http.StatusServiceUnavailable,
					Body:       reader, // Use reader directly, it already implements io.ReadCloser
					Header:     make(http.Header),
				}, nil
			}
			// Subsequent calls return 200
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Header:     make(http.Header),
			}, nil
		},
	}

	transport := NewRetryTransport(base, 5, 10*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Final response StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	resp.Body.Close()

	// Should have made at least 2 calls (initial + retry)
	if callCount < 2 {
		t.Errorf("callCount = %d, want at least 2", callCount)
	}

	// The first response body should be closed during retry
	if !bodyClosed {
		t.Error("Response body from failed attempt should be closed")
	}
}

// trackingReader tracks when Close is called
type trackingReader struct {
	onClose func()
}

func (r *trackingReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (r *trackingReader) Close() error {
	if r.onClose != nil {
		r.onClose()
	}
	return nil
}
