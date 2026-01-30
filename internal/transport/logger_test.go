package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewLoggingTransport(t *testing.T) {
	base := &mockRoundTripper{}
	logger := slog.Default()

	transport := NewLoggingTransport(base, logger)

	if transport == nil {
		t.Fatal("NewLoggingTransport returned nil")
	}
	if transport.base != base {
		t.Error("base transport not set correctly")
	}
	if transport.logger != logger {
		t.Error("logger not set correctly")
	}
}

func TestLoggingTransport_LogsRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Header:     make(http.Header),
			}, nil
		},
	}

	transport := NewLoggingTransport(base, logger)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	transport.RoundTrip(req)

	// Parse log output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Verify request was logged
	if logEntry["method"] != "GET" {
		t.Errorf("method = %v, want GET", logEntry["method"])
	}
	if logEntry["url"] != "http://example.com/test" {
		t.Errorf("url = %v, want http://example.com/test", logEntry["url"])
	}
}

func TestLoggingTransport_LogsResponse(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Header:     make(http.Header),
			}, nil
		},
	}

	transport := NewLoggingTransport(base, logger)

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	transport.RoundTrip(req)

	// Parse log output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Verify response was logged
	if logEntry["status"] != float64(200) { // JSON numbers are float64
		t.Errorf("status = %v, want 200", logEntry["status"])
	}

	// Verify duration was logged
	if _, ok := logEntry["duration_ms"]; !ok {
		t.Error("duration_ms not logged")
	}
}

func TestLoggingTransport_LogsError(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	expectedErr := errors.New("network error")
	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return nil, expectedErr
		},
	}

	transport := NewLoggingTransport(base, logger)

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	_, err := transport.RoundTrip(req)

	if err != expectedErr {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}

	// Parse log output
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) < 1 {
		t.Fatal("Expected at least one log line")
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Verify error was logged
	if logEntry["error"] != "network error" {
		t.Errorf("error = %v, want 'network error'", logEntry["error"])
	}
}

func TestLoggingTransport_UsesRequestContext(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Header:     make(http.Header),
			}, nil
		},
	}

	transport := NewLoggingTransport(base, logger)

	// Create request with context value
	ctx := context.WithValue(context.Background(), "request_id", "test-123")
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req = req.WithContext(ctx)

	transport.RoundTrip(req)

	// The logger should use the request's context
	// This is mainly tested by not panicking or causing issues
	if buf.Len() == 0 {
		t.Error("Expected log output")
	}
}

func TestLoggingTransport_LogsDuration(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Simulate slow request
	delay := 100 * time.Millisecond
	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			time.Sleep(delay)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Header:     make(http.Header),
			}, nil
		},
	}

	transport := NewLoggingTransport(base, logger)

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	transport.RoundTrip(req)

	// Parse log output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Verify duration is reasonable (at least the delay)
	durationMS, ok := logEntry["duration_ms"].(float64)
	if !ok {
		t.Fatal("duration_ms not found or not a number")
	}

	if durationMS < float64(delay.Milliseconds()) {
		t.Errorf("duration_ms = %.2f, want at least %.2f", durationMS, float64(delay.Milliseconds()))
	}
}

func TestLoggingTransport_LogsDifferentMethods(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

			base := &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       http.NoBody,
						Header:     make(http.Header),
					}, nil
				},
			}

			transport := NewLoggingTransport(base, logger)

			req := httptest.NewRequest(method, "http://example.com", nil)
			transport.RoundTrip(req)

			// Parse log output
			var logEntry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
				t.Fatalf("Failed to parse log output: %v", err)
			}

			if logEntry["method"] != method {
				t.Errorf("method = %v, want %s", logEntry["method"], method)
			}
		})
	}
}

func TestLoggingTransport_LogsErrorStatus(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       http.NoBody,
				Header:     make(http.Header),
			}, nil
		},
	}

	transport := NewLoggingTransport(base, logger)

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	transport.RoundTrip(req)

	// Parse log output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if logEntry["status"] != float64(500) {
		t.Errorf("status = %v, want 500", logEntry["status"])
	}
}

func TestLoggingTransport_PropagatesResponse(t *testing.T) {
	logger := slog.Default()

	expectedResp := &http.Response{
		StatusCode: http.StatusCreated,
		Body:       http.NoBody,
		Header:     make(http.Header),
	}

	base := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return expectedResp, nil
		},
	}

	transport := NewLoggingTransport(base, logger)

	req := httptest.NewRequest(http.MethodPost, "http://example.com", nil)
	resp, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}

	if resp != expectedResp {
		t.Error("Response was not propagated correctly")
	}
}
