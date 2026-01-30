package transport

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when the circuit breaker is open
var ErrCircuitOpen = errors.New("circuit breaker is open")

// circuitState represents the state of the circuit breaker
type circuitState int

const (
	stateClosed circuitState = iota
	stateOpen
	stateHalfOpen
)

// circuitBreaker implements the circuit breaker pattern
type circuitBreaker struct {
	mu            sync.RWMutex
	state         circuitState
	failureCount  int
	maxFailures   int
	resetTimeout  time.Duration
	lastFailTime  time.Time
}

// newCircuitBreaker creates a new circuit breaker
func newCircuitBreaker(maxFailures int, resetTimeout time.Duration) *circuitBreaker {
	return &circuitBreaker{
		state:        stateClosed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

// canExecute checks if a request can be executed
func (cb *circuitBreaker) canExecute() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case stateClosed:
		return nil

	case stateOpen:
		// Check if it's time to transition to half-open
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = stateHalfOpen
			return nil
		}
		return ErrCircuitOpen

	case stateHalfOpen:
		// Allow one request in half-open state
		return nil

	default:
		return ErrCircuitOpen
	}
}

// recordSuccess records a successful request
func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0

	// If we were half-open and succeeded, close the circuit
	if cb.state == stateHalfOpen {
		cb.state = stateClosed
	}
}

// recordFailure records a failed request
func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailTime = time.Now()

	// If we're half-open and failed, reopen the circuit
	if cb.state == stateHalfOpen {
		cb.state = stateOpen
		return
	}

	// Open circuit if we've hit max failures
	if cb.failureCount >= cb.maxFailures {
		cb.state = stateOpen
	}
}

// CircuitBreakerTransport is an http.RoundTripper that implements circuit breaker pattern
type CircuitBreakerTransport struct {
	base    http.RoundTripper
	breaker *circuitBreaker
}

// NewCircuitBreakerTransport creates a new CircuitBreakerTransport
func NewCircuitBreakerTransport(base http.RoundTripper, maxFailures int, resetTimeout time.Duration) *CircuitBreakerTransport {
	if base == nil {
		base = http.DefaultTransport
	}

	return &CircuitBreakerTransport{
		base:    base,
		breaker: newCircuitBreaker(maxFailures, resetTimeout),
	}
}

// RoundTrip implements http.RoundTripper with circuit breaker logic
func (t *CircuitBreakerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Check if we can execute the request
	if err := t.breaker.canExecute(); err != nil {
		return nil, err
	}

	// Execute request
	resp, err := t.base.RoundTrip(req)

	// Check if request failed
	if isCircuitBreakerFailure(resp, err) {
		t.breaker.recordFailure()
		return resp, err
	}

	// Request succeeded
	t.breaker.recordSuccess()
	return resp, err
}

// isCircuitBreakerFailure determines if a response/error should be counted as a failure
func isCircuitBreakerFailure(resp *http.Response, err error) bool {
	// Network errors are failures
	if err != nil {
		return true
	}

	// No response is a failure
	if resp == nil {
		return true
	}

	// 5xx errors are failures
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		return true
	}

	return false
}
