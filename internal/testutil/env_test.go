package testutil

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		want         string
	}{
		{
			name:         "returns environment value when set",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "custom",
			want:         "custom",
		},
		{
			name:         "returns default when not set",
			key:          "TEST_MISSING_KEY",
			defaultValue: "default",
			envValue:     "",
			want:         "default",
		},
		{
			name:         "returns empty string when default is empty",
			key:          "TEST_EMPTY_KEY",
			defaultValue: "",
			envValue:     "",
			want:         "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			// Test
			got := GetEnv(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetEnv() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetRequiredEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		envValue string
		want     string
	}{
		{
			name:     "returns value when set",
			key:      "REQUIRED_KEY",
			envValue: "value",
			want:     "value",
		},
		{
			name:     "returns empty when not set",
			key:      "MISSING_REQUIRED_KEY",
			envValue: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			// Test
			got := GetRequiredEnv(tt.key)
			if got != tt.want {
				t.Errorf("GetRequiredEnv() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHasCredentials(t *testing.T) {
	tests := []struct {
		name      string
		clientID  string
		secret    string
		appKey    string
		want      bool
	}{
		{
			name:     "all credentials set",
			clientID: "id",
			secret:   "secret",
			appKey:   "key",
			want:     true,
		},
		{
			name:     "missing client ID",
			clientID: "",
			secret:   "secret",
			appKey:   "key",
			want:     false,
		},
		{
			name:     "missing secret",
			clientID: "id",
			secret:   "",
			appKey:   "key",
			want:     false,
		},
		{
			name:     "missing app key",
			clientID: "id",
			secret:   "secret",
			appKey:   "",
			want:     false,
		},
		{
			name:     "all missing",
			clientID: "",
			secret:   "",
			appKey:   "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			os.Setenv("BB_CLIENT_ID", tt.clientID)
			os.Setenv("BB_CLIENT_SECRET", tt.secret)
			os.Setenv("BB_DEV_APP_KEY", tt.appKey)
			defer func() {
				os.Unsetenv("BB_CLIENT_ID")
				os.Unsetenv("BB_CLIENT_SECRET")
				os.Unsetenv("BB_DEV_APP_KEY")
			}()

			// Test
			got := HasCredentials()
			if got != tt.want {
				t.Errorf("HasCredentials() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBBConfig(t *testing.T) {
	// Setup
	os.Setenv("BB_ENVIRONMENT", "sandbox")
	os.Setenv("BB_CLIENT_ID", "test-id")
	os.Setenv("BB_CLIENT_SECRET", "test-secret")
	os.Setenv("BB_DEV_APP_KEY", "test-key")
	defer func() {
		os.Unsetenv("BB_ENVIRONMENT")
		os.Unsetenv("BB_CLIENT_ID")
		os.Unsetenv("BB_CLIENT_SECRET")
		os.Unsetenv("BB_DEV_APP_KEY")
	}()

	// Test
	config := GetBBConfig()

	tests := []struct {
		key  string
		want string
	}{
		{"environment", "sandbox"},
		{"client_id", "test-id"},
		{"client_secret", "test-secret"},
		{"dev_app_key", "test-key"},
		{"timeout", "30"}, // default value
		{"retry_count", "3"}, // default value
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := config[tt.key]; got != tt.want {
				t.Errorf("config[%q] = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}
