package bbpix

import (
	"github.com/pericles-luz/go-bb-pix/internal/apierror"
)

// Re-export types from internal/apierror for public API
type (
	// ErrorDetail represents a detailed error message for a specific field
	ErrorDetail = apierror.ErrorDetail

	// APIError represents an error returned by the Banco do Brasil API
	APIError = apierror.APIError
)

// NewAPIError creates a new APIError
func NewAPIError(statusCode int, message string, details ...ErrorDetail) *APIError {
	return apierror.New(statusCode, message, details...)
}

// IsAPIError checks if an error is an APIError or wraps one
func IsAPIError(err error) bool {
	return apierror.Is(err)
}

// GetAPIError extracts an APIError from an error chain
// Returns the APIError and nil if found, or nil and an error if not found
func GetAPIError(err error) (*APIError, error) {
	return apierror.As(err)
}
