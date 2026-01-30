package bbpix

import (
	"os"
	"testing"
)

func TestEnvironment_URLs(t *testing.T) {
	tests := []struct {
		name        string
		env         Environment
		wantOAuthURL string
		wantAPIURL  string
	}{
		{
			name:        "sandbox",
			env:         EnvironmentSandbox,
			wantOAuthURL: "https://oauth.sandbox.bb.com.br/oauth/token",
			wantAPIURL:  "https://api.sandbox.bb.com.br/pix-bb/v1",
		},
		{
			name:        "homologacao",
			env:         EnvironmentHomologacao,
			wantOAuthURL: "https://oauth.hm.bb.com.br/oauth/token",
			wantAPIURL:  "https://api.hm.bb.com.br/pix-bb/v1",
		},
		{
			name:        "producao",
			env:         EnvironmentProducao,
			wantOAuthURL: "https://oauth.bb.com.br/oauth/token",
			wantAPIURL:  "https://api.bb.com.br/pix-bb/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOAuth, gotAPI := tt.env.URLs()
			if gotOAuth != tt.wantOAuthURL {
				t.Errorf("OAuth URL = %q, want %q", gotOAuth, tt.wantOAuthURL)
			}
			if gotAPI != tt.wantAPIURL {
				t.Errorf("API URL = %q, want %q", gotAPI, tt.wantAPIURL)
			}
		})
	}
}

func TestEnvironment_String(t *testing.T) {
	tests := []struct {
		name string
		env  Environment
		want string
	}{
		{
			name: "sandbox",
			env:  EnvironmentSandbox,
			want: "sandbox",
		},
		{
			name: "homologacao",
			env:  EnvironmentHomologacao,
			want: "homologacao",
		},
		{
			name: "producao",
			env:  EnvironmentProducao,
			want: "producao",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.env.String()
			if got != tt.want {
				t.Errorf("Environment.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseEnvironment(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Environment
		wantErr bool
	}{
		{
			name:    "sandbox",
			input:   "sandbox",
			want:    EnvironmentSandbox,
			wantErr: false,
		},
		{
			name:    "homologacao",
			input:   "homologacao",
			want:    EnvironmentHomologacao,
			wantErr: false,
		},
		{
			name:    "producao",
			input:   "producao",
			want:    EnvironmentProducao,
			wantErr: false,
		},
		{
			name:    "uppercase sandbox",
			input:   "SANDBOX",
			want:    EnvironmentSandbox,
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   "invalid",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty",
			input:   "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEnvironment(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEnvironment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseEnvironment() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				Environment:     EnvironmentSandbox,
				ClientID:        "client-id",
				ClientSecret:    "client-secret",
				DeveloperAppKey: "app-key",
			},
			wantErr: false,
		},
		{
			name: "missing client id",
			config: Config{
				Environment:     EnvironmentSandbox,
				ClientSecret:    "client-secret",
				DeveloperAppKey: "app-key",
			},
			wantErr: true,
			errMsg:  "client_id is required",
		},
		{
			name: "missing client secret",
			config: Config{
				Environment:     EnvironmentSandbox,
				ClientID:        "client-id",
				DeveloperAppKey: "app-key",
			},
			wantErr: true,
			errMsg:  "client_secret is required",
		},
		{
			name: "missing developer app key",
			config: Config{
				Environment:  EnvironmentSandbox,
				ClientID:     "client-id",
				ClientSecret: "client-secret",
			},
			wantErr: true,
			errMsg:  "developer_application_key is required",
		},
		{
			name: "missing environment",
			config: Config{
				ClientID:        "client-id",
				ClientSecret:    "client-secret",
				DeveloperAppKey: "app-key",
			},
			wantErr: true,
			errMsg:  "environment is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("NewConfig() error = %q, want %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid",
			config: Config{
				Environment:     EnvironmentSandbox,
				ClientID:        "id",
				ClientSecret:    "secret",
				DeveloperAppKey: "key",
			},
			wantErr: false,
		},
		{
			name: "empty client id",
			config: Config{
				Environment:     EnvironmentSandbox,
				ClientSecret:    "secret",
				DeveloperAppKey: "key",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Save original env vars
	origEnv := os.Getenv("BB_ENVIRONMENT")
	origClientID := os.Getenv("BB_CLIENT_ID")
	origClientSecret := os.Getenv("BB_CLIENT_SECRET")
	origAppKey := os.Getenv("BB_DEV_APP_KEY")

	// Restore after test
	defer func() {
		os.Setenv("BB_ENVIRONMENT", origEnv)
		os.Setenv("BB_CLIENT_ID", origClientID)
		os.Setenv("BB_CLIENT_SECRET", origClientSecret)
		os.Setenv("BB_DEV_APP_KEY", origAppKey)
	}()

	tests := []struct {
		name    string
		setup   func()
		want    Config
		wantErr bool
	}{
		{
			name: "valid environment variables",
			setup: func() {
				os.Setenv("BB_ENVIRONMENT", "sandbox")
				os.Setenv("BB_CLIENT_ID", "test-client-id")
				os.Setenv("BB_CLIENT_SECRET", "test-client-secret")
				os.Setenv("BB_DEV_APP_KEY", "test-app-key")
			},
			want: Config{
				Environment:     EnvironmentSandbox,
				ClientID:        "test-client-id",
				ClientSecret:    "test-client-secret",
				DeveloperAppKey: "test-app-key",
			},
			wantErr: false,
		},
		{
			name: "missing environment",
			setup: func() {
				os.Unsetenv("BB_ENVIRONMENT")
				os.Setenv("BB_CLIENT_ID", "test-client-id")
				os.Setenv("BB_CLIENT_SECRET", "test-client-secret")
				os.Setenv("BB_DEV_APP_KEY", "test-app-key")
			},
			wantErr: true,
		},
		{
			name: "invalid environment",
			setup: func() {
				os.Setenv("BB_ENVIRONMENT", "invalid")
				os.Setenv("BB_CLIENT_ID", "test-client-id")
				os.Setenv("BB_CLIENT_SECRET", "test-client-secret")
				os.Setenv("BB_DEV_APP_KEY", "test-app-key")
			},
			wantErr: true,
		},
		{
			name: "missing credentials",
			setup: func() {
				os.Setenv("BB_ENVIRONMENT", "sandbox")
				os.Unsetenv("BB_CLIENT_ID")
				os.Unsetenv("BB_CLIENT_SECRET")
				os.Unsetenv("BB_DEV_APP_KEY")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			got, err := LoadConfigFromEnv()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfigFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.Environment != tt.want.Environment {
					t.Errorf("Environment = %q, want %q", got.Environment, tt.want.Environment)
				}
				if got.ClientID != tt.want.ClientID {
					t.Errorf("ClientID = %q, want %q", got.ClientID, tt.want.ClientID)
				}
				if got.ClientSecret != tt.want.ClientSecret {
					t.Errorf("ClientSecret = %q, want %q", got.ClientSecret, tt.want.ClientSecret)
				}
				if got.DeveloperAppKey != tt.want.DeveloperAppKey {
					t.Errorf("DeveloperAppKey = %q, want %q", got.DeveloperAppKey, tt.want.DeveloperAppKey)
				}
			}
		})
	}
}
