package bbpix

import (
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestNew_ValidConfig(t *testing.T) {
	config := Config{
		Environment:     EnvironmentSandbox,
		ClientID:        "test-client-id",
		ClientSecret:    "test-client-secret",
		DeveloperAppKey: "test-app-key",
	}

	client, err := New(config)

	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if client == nil {
		t.Fatal("New() returned nil client")
	}
}

func TestNew_InvalidConfig(t *testing.T) {
	config := Config{
		Environment: EnvironmentSandbox,
		// Missing required fields
	}

	_, err := New(config)

	if err == nil {
		t.Fatal("Expected error for invalid config, got nil")
	}
}

func TestNew_WithOptions(t *testing.T) {
	config := Config{
		Environment:     EnvironmentSandbox,
		ClientID:        "test-client-id",
		ClientSecret:    "test-client-secret",
		DeveloperAppKey: "test-app-key",
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	timeout := 60 * time.Second

	client, err := New(config,
		WithLogger(logger),
		WithTimeout(timeout),
		WithRetry(5, 200*time.Millisecond),
	)

	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if client == nil {
		t.Fatal("New() returned nil client")
	}
}

func TestNew_BuildsHTTPClientWithTransportChain(t *testing.T) {
	config := Config{
		Environment:     EnvironmentSandbox,
		ClientID:        "test-client-id",
		ClientSecret:    "test-client-secret",
		DeveloperAppKey: "test-app-key",
	}

	client, err := New(config)

	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Verify HTTP client is configured
	if client.httpClient == nil {
		t.Error("HTTP client not initialized")
	}

	// Verify timeout is set
	if client.httpClient.Timeout == 0 {
		t.Error("HTTP client timeout not set")
	}
}

func TestNew_CustomHTTPClient(t *testing.T) {
	config := Config{
		Environment:     EnvironmentSandbox,
		ClientID:        "test-client-id",
		ClientSecret:    "test-client-secret",
		DeveloperAppKey: "test-app-key",
	}

	customClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	client, err := New(config, WithHTTPClient(customClient))

	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// When custom HTTP client is provided, transport chain should still be applied
	if client.httpClient == customClient {
		t.Error("Custom HTTP client should have transport chain applied")
	}
}

func TestClient_PIX(t *testing.T) {
	config := Config{
		Environment:     EnvironmentSandbox,
		ClientID:        "test-client-id",
		ClientSecret:    "test-client-secret",
		DeveloperAppKey: "test-app-key",
	}

	client, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	pixClient := client.PIX()

	if pixClient == nil {
		t.Fatal("PIX() returned nil")
	}
}

func TestClient_PIXAuto(t *testing.T) {
	config := Config{
		Environment:     EnvironmentSandbox,
		ClientID:        "test-client-id",
		ClientSecret:    "test-client-secret",
		DeveloperAppKey: "test-app-key",
	}

	client, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	pixAutoClient := client.PIXAuto()

	if pixAutoClient == nil {
		t.Fatal("PIXAuto() returned nil")
	}
}

func TestNew_DifferentEnvironments(t *testing.T) {
	environments := []Environment{
		EnvironmentSandbox,
		EnvironmentHomologacao,
		EnvironmentProducao,
	}

	for _, env := range environments {
		t.Run(env.String(), func(t *testing.T) {
			config := Config{
				Environment:     env,
				ClientID:        "test-client-id",
				ClientSecret:    "test-client-secret",
				DeveloperAppKey: "test-app-key",
			}

			client, err := New(config)

			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			if client == nil {
				t.Fatal("New() returned nil client")
			}

			// Verify environment URLs are set correctly
			_, apiURL := env.URLs()
			if client.apiURL != apiURL {
				t.Errorf("apiURL = %s, want %s", client.apiURL, apiURL)
			}
		})
	}
}

func TestClient_Singleton_PIX(t *testing.T) {
	config := Config{
		Environment:     EnvironmentSandbox,
		ClientID:        "test-client-id",
		ClientSecret:    "test-client-secret",
		DeveloperAppKey: "test-app-key",
	}

	client, _ := New(config)

	pix1 := client.PIX()
	pix2 := client.PIX()

	// PIX() should return the same instance
	if pix1 != pix2 {
		t.Error("PIX() should return singleton instance")
	}
}

func TestClient_Singleton_PIXAuto(t *testing.T) {
	config := Config{
		Environment:     EnvironmentSandbox,
		ClientID:        "test-client-id",
		ClientSecret:    "test-client-secret",
		DeveloperAppKey: "test-app-key",
	}

	client, _ := New(config)

	pixAuto1 := client.PIXAuto()
	pixAuto2 := client.PIXAuto()

	// PIXAuto() should return the same instance
	if pixAuto1 != pixAuto2 {
		t.Error("PIXAuto() should return singleton instance")
	}
}
