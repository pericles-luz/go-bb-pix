package transport

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// RetryTransport is an http.RoundTripper that implements retry logic with exponential backoff
type RetryTransport struct {
	base           http.RoundTripper
	maxRetries     int
	initialBackoff time.Duration
}

// NewRetryTransport creates a new RetryTransport
func NewRetryTransport(base http.RoundTripper, maxRetries int, initialBackoff time.Duration) *RetryTransport {
	if base == nil {
		base = http.DefaultTransport
	}

	return &RetryTransport{
		base:           base,
		maxRetries:     maxRetries,
		initialBackoff: initialBackoff,
	}
}

// RoundTrip implements http.RoundTripper with retry logic
func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var lastErr error
	var resp *http.Response

	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		// Check if context is cancelled
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		default:
		}

		// Execute request
		resp, lastErr = t.base.RoundTrip(req)

		// If successful or should not retry, return
		if lastErr == nil && !shouldRetry(resp, nil) {
			return resp, nil
		}

		// If error occurred or retryable status code
		if shouldRetry(resp, lastErr) && isIdempotent(req.Method) {
			// Close response body if we got one (to avoid leaks)
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}

			// If this is not the last attempt, sleep before retry
			if attempt < t.maxRetries {
				backoff := t.calculateBackoff(attempt)

				// Check context before sleeping
				select {
				case <-req.Context().Done():
					return nil, req.Context().Err()
				case <-time.After(backoff):
					// Continue to next attempt
				}
			}
		} else {
			// Not retryable or not idempotent, return immediately
			return resp, lastErr
		}
	}

	// All retries exhausted
	if lastErr != nil {
		return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
	}

	return resp, nil
}

// calculateBackoff calculates exponential backoff with jitter
func (t *RetryTransport) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: initialBackoff * 2^attempt
	backoff := float64(t.initialBackoff) * math.Pow(2, float64(attempt))

	// Add jitter (random ±25%)
	jitter := 0.75 + (rand.Float64() * 0.5) // 0.75 to 1.25
	backoff *= jitter

	return time.Duration(backoff)
}

// isIdempotent checks if an HTTP method is idempotent
// Only idempotent methods should be retried to avoid duplicating operations
func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

// shouldRetry determines if a request should be retried based on response and error
func shouldRetry(resp *http.Response, err error) bool {
	// Retry on network errors
	if err != nil {
		return true
	}

	// No response means network error (already handled above)
	if resp == nil {
		return true
	}

	// Retry on specific status codes
	switch resp.StatusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusBadGateway,           // 502
		http.StatusServiceUnavailable,   // 503
		http.StatusGatewayTimeout:       // 504
		return true
	default:
		return false
	}
}
