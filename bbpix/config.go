package bbpix

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Environment represents the API environment
type Environment string

const (
	// EnvironmentSandbox is the sandbox environment for testing
	EnvironmentSandbox Environment = "sandbox"

	// EnvironmentHomologacao is the homologation environment
	EnvironmentHomologacao Environment = "homologacao"

	// EnvironmentProducao is the production environment
	EnvironmentProducao Environment = "producao"
)

// String returns the string representation of the environment
func (e Environment) String() string {
	return string(e)
}

// URLs returns the OAuth and API URLs for the environment
func (e Environment) URLs() (oauthURL, apiURL string) {
	switch e {
	case EnvironmentSandbox:
		return "https://oauth.sandbox.bb.com.br/oauth/token",
			"https://api.sandbox.bb.com.br/pix-bb/v1"
	case EnvironmentHomologacao:
		return "https://oauth.hm.bb.com.br/oauth/token",
			"https://api.hm.bb.com.br/pix-bb/v1"
	case EnvironmentProducao:
		return "https://oauth.bb.com.br/oauth/token",
			"https://api.bb.com.br/pix-bb/v1"
	default:
		return "", ""
	}
}

// ParseEnvironment parses a string into an Environment
func ParseEnvironment(s string) (Environment, error) {
	switch strings.ToLower(s) {
	case "sandbox":
		return EnvironmentSandbox, nil
	case "homologacao":
		return EnvironmentHomologacao, nil
	case "producao":
		return EnvironmentProducao, nil
	default:
		return "", fmt.Errorf("invalid environment: %s (must be sandbox, homologacao, or producao)", s)
	}
}

// Config holds the configuration for the Banco do Brasil PIX client
type Config struct {
	// Environment is the API environment (sandbox, homologacao, producao)
	Environment Environment

	// ClientID is the OAuth2 client ID
	ClientID string

	// ClientSecret is the OAuth2 client secret
	ClientSecret string

	// DeveloperAppKey is the developer application key
	// (gw-dev-app-key for sandbox, gw-app-key for production)
	DeveloperAppKey string
}

// Validate checks if the configuration is valid
func (c Config) Validate() error {
	if c.Environment == "" {
		return errors.New("environment is required")
	}

	if c.ClientID == "" {
		return errors.New("client_id is required")
	}

	if c.ClientSecret == "" {
		return errors.New("client_secret is required")
	}

	if c.DeveloperAppKey == "" {
		return errors.New("developer_application_key is required")
	}

	// Validate environment URLs
	oauthURL, apiURL := c.Environment.URLs()
	if oauthURL == "" || apiURL == "" {
		return fmt.Errorf("invalid environment: %s", c.Environment)
	}

	return nil
}

// NewConfig creates and validates a new Config
func NewConfig(cfg Config) error {
	return cfg.Validate()
}

// LoadConfigFromEnv loads configuration from environment variables
// Expected variables:
//   - BB_ENVIRONMENT: sandbox, homologacao, or producao
//   - BB_CLIENT_ID: OAuth2 client ID
//   - BB_CLIENT_SECRET: OAuth2 client secret
//   - BB_DEV_APP_KEY: Developer application key
func LoadConfigFromEnv() (Config, error) {
	envStr := os.Getenv("BB_ENVIRONMENT")
	if envStr == "" {
		return Config{}, errors.New("BB_ENVIRONMENT environment variable is required")
	}

	env, err := ParseEnvironment(envStr)
	if err != nil {
		return Config{}, fmt.Errorf("invalid BB_ENVIRONMENT: %w", err)
	}

	clientID := os.Getenv("BB_CLIENT_ID")
	if clientID == "" {
		return Config{}, errors.New("BB_CLIENT_ID environment variable is required")
	}

	clientSecret := os.Getenv("BB_CLIENT_SECRET")
	if clientSecret == "" {
		return Config{}, errors.New("BB_CLIENT_SECRET environment variable is required")
	}

	appKey := os.Getenv("BB_DEV_APP_KEY")
	if appKey == "" {
		return Config{}, errors.New("BB_DEV_APP_KEY environment variable is required")
	}

	cfg := Config{
		Environment:     env,
		ClientID:        clientID,
		ClientSecret:    clientSecret,
		DeveloperAppKey: appKey,
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}
