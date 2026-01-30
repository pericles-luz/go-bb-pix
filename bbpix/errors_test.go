package bbpix

import (
	"errors"
	"fmt"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiError *APIError
		want     string
	}{
		{
			name: "basic error",
			apiError: &APIError{
				StatusCode: 400,
				Message:    "Invalid request",
			},
			want: "API error (400): Invalid request",
		},
		{
			name: "error with details",
			apiError: &APIError{
				StatusCode: 422,
				Message:    "Validation failed",
				Details: []ErrorDetail{
					{Field: "txid", Message: "required field"},
					{Field: "value", Message: "must be positive"},
				},
			},
			want: "API error (422): Validation failed [txid: required field; value: must be positive]",
		},
		{
			name: "error with single detail",
			apiError: &APIError{
				StatusCode: 404,
				Message:    "Not found",
				Details: []ErrorDetail{
					{Field: "id", Message: "resource not found"},
				},
			},
			want: "API error (404): Not found [id: resource not found]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.apiError.Error()
			if got != tt.want {
				t.Errorf("APIError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewAPIError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
		details    []ErrorDetail
		want       *APIError
	}{
		{
			name:       "simple error",
			statusCode: 500,
			message:    "Internal server error",
			details:    nil,
			want: &APIError{
				StatusCode: 500,
				Message:    "Internal server error",
				Details:    nil,
			},
		},
		{
			name:       "error with details",
			statusCode: 400,
			message:    "Bad request",
			details: []ErrorDetail{
				{Field: "amount", Message: "invalid"},
			},
			want: &APIError{
				StatusCode: 400,
				Message:    "Bad request",
				Details: []ErrorDetail{
					{Field: "amount", Message: "invalid"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAPIError(tt.statusCode, tt.message, tt.details...)
			if got.StatusCode != tt.want.StatusCode {
				t.Errorf("StatusCode = %d, want %d", got.StatusCode, tt.want.StatusCode)
			}
			if got.Message != tt.want.Message {
				t.Errorf("Message = %q, want %q", got.Message, tt.want.Message)
			}
			if len(got.Details) != len(tt.want.Details) {
				t.Errorf("len(Details) = %d, want %d", len(got.Details), len(tt.want.Details))
			}
		})
	}
}

func TestIsAPIError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "is api error",
			err:  &APIError{StatusCode: 400, Message: "bad request"},
			want: true,
		},
		{
			name: "wrapped api error",
			err:  fmt.Errorf("operation failed: %w", &APIError{StatusCode: 500, Message: "server error"}),
			want: true,
		},
		{
			name: "not api error",
			err:  errors.New("generic error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAPIError(tt.err)
			if got != tt.want {
				t.Errorf("IsAPIError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAPIError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		want    *APIError
		wantErr bool
	}{
		{
			name:    "direct api error",
			err:     &APIError{StatusCode: 400, Message: "bad request"},
			want:    &APIError{StatusCode: 400, Message: "bad request"},
			wantErr: false,
		},
		{
			name:    "wrapped api error",
			err:     fmt.Errorf("failed: %w", &APIError{StatusCode: 404, Message: "not found"}),
			want:    &APIError{StatusCode: 404, Message: "not found"},
			wantErr: false,
		},
		{
			name:    "not api error",
			err:     errors.New("generic error"),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "nil error",
			err:     nil,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAPIError(tt.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAPIError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.StatusCode != tt.want.StatusCode {
					t.Errorf("StatusCode = %d, want %d", got.StatusCode, tt.want.StatusCode)
				}
				if got.Message != tt.want.Message {
					t.Errorf("Message = %q, want %q", got.Message, tt.want.Message)
				}
			}
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	// Test that APIError works correctly with errors.Is and errors.As
	baseErr := &APIError{StatusCode: 400, Message: "bad request"}
	wrappedErr := fmt.Errorf("operation failed: %w", baseErr)

	// Test errors.Is
	if !errors.Is(wrappedErr, baseErr) {
		t.Error("errors.Is should find the wrapped APIError")
	}

	// Test errors.As
	var apiErr *APIError
	if !errors.As(wrappedErr, &apiErr) {
		t.Error("errors.As should extract the APIError")
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("extracted error StatusCode = %d, want 400", apiErr.StatusCode)
	}
}

func TestErrorDetail_String(t *testing.T) {
	tests := []struct {
		name   string
		detail ErrorDetail
		want   string
	}{
		{
			name:   "basic detail",
			detail: ErrorDetail{Field: "amount", Message: "must be positive"},
			want:   "amount: must be positive",
		},
		{
			name:   "empty field",
			detail: ErrorDetail{Field: "", Message: "generic error"},
			want:   ": generic error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.detail.String()
			if got != tt.want {
				t.Errorf("ErrorDetail.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
