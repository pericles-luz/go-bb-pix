package bbpix

import (
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestWithLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	opts := &clientOptions{}
	opt := WithLogger(logger)
	opt(opts)

	if opts.logger != logger {
		t.Error("WithLogger did not set the logger")
	}
}

func TestWithHTTPClient(t *testing.T) {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	opts := &clientOptions{}
	opt := WithHTTPClient(httpClient)
	opt(opts)

	if opts.httpClient != httpClient {
		t.Error("WithHTTPClient did not set the HTTP client")
	}
}

func TestWithTimeout(t *testing.T) {
	timeout := 30 * time.Second

	opts := &clientOptions{}
	opt := WithTimeout(timeout)
	opt(opts)

	if opts.timeout != timeout {
		t.Errorf("WithTimeout = %v, want %v", opts.timeout, timeout)
	}
}

func TestWithRetry(t *testing.T) {
	maxRetries := 5
	initialBackoff := 200 * time.Millisecond

	opts := &clientOptions{}
	opt := WithRetry(maxRetries, initialBackoff)
	opt(opts)

	if opts.maxRetries != maxRetries {
		t.Errorf("maxRetries = %d, want %d", opts.maxRetries, maxRetries)
	}
	if opts.initialBackoff != initialBackoff {
		t.Errorf("initialBackoff = %v, want %v", opts.initialBackoff, initialBackoff)
	}
}

func TestWithCircuitBreaker(t *testing.T) {
	maxFailures := 10
	resetTimeout := 5 * time.Second

	opts := &clientOptions{}
	opt := WithCircuitBreaker(maxFailures, resetTimeout)
	opt(opts)

	if opts.circuitBreakerMaxFailures != maxFailures {
		t.Errorf("circuitBreakerMaxFailures = %d, want %d", opts.circuitBreakerMaxFailures, maxFailures)
	}
	if opts.circuitBreakerResetTimeout != resetTimeout {
		t.Errorf("circuitBreakerResetTimeout = %v, want %v", opts.circuitBreakerResetTimeout, resetTimeout)
	}
}

func TestWithUserAgent(t *testing.T) {
	userAgent := "custom-agent/1.0"

	opts := &clientOptions{}
	opt := WithUserAgent(userAgent)
	opt(opts)

	if opts.userAgent != userAgent {
		t.Errorf("userAgent = %q, want %q", opts.userAgent, userAgent)
	}
}

func TestMultipleOptions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	timeout := 30 * time.Second
	maxRetries := 3

	opts := &clientOptions{}

	// Apply multiple options
	options := []Option{
		WithLogger(logger),
		WithTimeout(timeout),
		WithRetry(maxRetries, 100*time.Millisecond),
	}

	for _, opt := range options {
		opt(opts)
	}

	if opts.logger != logger {
		t.Error("logger was not set")
	}
	if opts.timeout != timeout {
		t.Errorf("timeout = %v, want %v", opts.timeout, timeout)
	}
	if opts.maxRetries != maxRetries {
		t.Errorf("maxRetries = %d, want %d", opts.maxRetries, maxRetries)
	}
}

func TestDefaultClientOptions(t *testing.T) {
	opts := defaultClientOptions()

	if opts.timeout != 30*time.Second {
		t.Errorf("default timeout = %v, want %v", opts.timeout, 30*time.Second)
	}
	if opts.maxRetries != 3 {
		t.Errorf("default maxRetries = %d, want %d", opts.maxRetries, 3)
	}
	if opts.initialBackoff != 100*time.Millisecond {
		t.Errorf("default initialBackoff = %v, want %v", opts.initialBackoff, 100*time.Millisecond)
	}
	if opts.circuitBreakerMaxFailures != 5 {
		t.Errorf("default circuitBreakerMaxFailures = %d, want %d", opts.circuitBreakerMaxFailures, 5)
	}
	if opts.circuitBreakerResetTimeout != 60*time.Second {
		t.Errorf("default circuitBreakerResetTimeout = %v, want %v", opts.circuitBreakerResetTimeout, 60*time.Second)
	}
	if opts.logger == nil {
		t.Error("default logger should not be nil")
	}
	if opts.httpClient != nil {
		t.Error("default httpClient should be nil (created later)")
	}
}

func TestOptionsOverrideDefaults(t *testing.T) {
	customTimeout := 60 * time.Second
	customRetries := 10

	opts := defaultClientOptions()

	WithTimeout(customTimeout)(opts)
	WithRetry(customRetries, 200*time.Millisecond)(opts)

	if opts.timeout != customTimeout {
		t.Errorf("timeout = %v, want %v (should override default)", opts.timeout, customTimeout)
	}
	if opts.maxRetries != customRetries {
		t.Errorf("maxRetries = %d, want %d (should override default)", opts.maxRetries, customRetries)
	}
}
