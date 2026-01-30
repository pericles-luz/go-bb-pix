package apierror

import (
	"errors"
	"fmt"
	"strings"
)

// ErrorDetail represents a detailed error message for a specific field
type ErrorDetail struct {
	Field   string
	Message string
}

// String returns a string representation of the error detail
func (e ErrorDetail) String() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// APIError represents an error returned by the Banco do Brasil API
type APIError struct {
	StatusCode int
	Message    string
	Details    []ErrorDetail
}

// Error implements the error interface
func (e *APIError) Error() string {
	if len(e.Details) == 0 {
		return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Message)
	}

	detailStrs := make([]string, len(e.Details))
	for i, detail := range e.Details {
		detailStrs[i] = detail.String()
	}

	return fmt.Sprintf("API error (%d): %s [%s]", e.StatusCode, e.Message, strings.Join(detailStrs, "; "))
}

// New creates a new APIError
func New(statusCode int, message string, details ...ErrorDetail) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Details:    details,
	}
}

// Is checks if an error is an APIError or wraps one
func Is(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *APIError
	return errors.As(err, &apiErr)
}

// As extracts an APIError from an error chain
// Returns the APIError and nil if found, or nil and an error if not found
func As(err error) (*APIError, error) {
	if err == nil {
		return nil, errors.New("nil error")
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, nil
	}

	return nil, errors.New("not an API error")
}
