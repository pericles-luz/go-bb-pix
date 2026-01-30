package bbpix

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

// Option is a functional option for configuring the client
type Option func(*clientOptions)

// clientOptions holds all configurable options for the client
type clientOptions struct {
	logger                       *slog.Logger
	httpClient                   *http.Client
	timeout                      time.Duration
	maxRetries                   int
	initialBackoff               time.Duration
	circuitBreakerMaxFailures    int
	circuitBreakerResetTimeout   time.Duration
	userAgent                    string
}

// defaultClientOptions returns the default client options
func defaultClientOptions() *clientOptions {
	return &clientOptions{
		logger:                     slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})),
		timeout:                    30 * time.Second,
		maxRetries:                 3,
		initialBackoff:             100 * time.Millisecond,
		circuitBreakerMaxFailures:  5,
		circuitBreakerResetTimeout: 60 * time.Second,
		userAgent:                  "go-bb-pix/1.0.0",
	}
}

// WithLogger sets a custom logger for the client
func WithLogger(logger *slog.Logger) Option {
	return func(opts *clientOptions) {
		opts.logger = logger
	}
}

// WithHTTPClient sets a custom HTTP client
// If not set, a default client with appropriate timeouts will be created
func WithHTTPClient(client *http.Client) Option {
	return func(opts *clientOptions) {
		opts.httpClient = client
	}
}

// WithTimeout sets the timeout for HTTP requests
// Default: 30 seconds
func WithTimeout(timeout time.Duration) Option {
	return func(opts *clientOptions) {
		opts.timeout = timeout
	}
}

// WithRetry configures the retry behavior
// maxRetries: maximum number of retry attempts (default: 3)
// initialBackoff: initial backoff duration (default: 100ms)
// Backoff will increase exponentially with jitter
func WithRetry(maxRetries int, initialBackoff time.Duration) Option {
	return func(opts *clientOptions) {
		opts.maxRetries = maxRetries
		opts.initialBackoff = initialBackoff
	}
}

// WithCircuitBreaker configures the circuit breaker
// maxFailures: number of consecutive failures before opening circuit (default: 5)
// resetTimeout: time to wait before attempting to close circuit (default: 60s)
func WithCircuitBreaker(maxFailures int, resetTimeout time.Duration) Option {
	return func(opts *clientOptions) {
		opts.circuitBreakerMaxFailures = maxFailures
		opts.circuitBreakerResetTimeout = resetTimeout
	}
}

// WithUserAgent sets a custom User-Agent header
// Default: "go-bb-pix/1.0.0"
func WithUserAgent(userAgent string) Option {
	return func(opts *clientOptions) {
		opts.userAgent = userAgent
	}
}
