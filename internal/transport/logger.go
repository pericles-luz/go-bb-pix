package transport

import (
	"log/slog"
	"net/http"
	"time"
)

// LoggingTransport is an http.RoundTripper that logs requests and responses
type LoggingTransport struct {
	base   http.RoundTripper
	logger *slog.Logger
}

// NewLoggingTransport creates a new LoggingTransport
func NewLoggingTransport(base http.RoundTripper, logger *slog.Logger) *LoggingTransport {
	if base == nil {
		base = http.DefaultTransport
	}

	return &LoggingTransport{
		base:   base,
		logger: logger,
	}
}

// RoundTrip implements http.RoundTripper with logging
func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Execute request
	resp, err := t.base.RoundTrip(req)

	// Calculate duration
	duration := time.Since(start)

	// Log the request/response
	if err != nil {
		t.logger.InfoContext(req.Context(), "HTTP request failed",
			slog.String("method", req.Method),
			slog.String("url", req.URL.String()),
			slog.Float64("duration_ms", float64(duration.Milliseconds())),
			slog.String("error", err.Error()),
		)
	} else {
		t.logger.InfoContext(req.Context(), "HTTP request completed",
			slog.String("method", req.Method),
			slog.String("url", req.URL.String()),
			slog.Int("status", resp.StatusCode),
			slog.Float64("duration_ms", float64(duration.Milliseconds())),
		)
	}

	return resp, err
}
