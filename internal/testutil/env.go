package testutil

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/joho/godotenv"
)

var (
	envLoadOnce sync.Once
	envLoaded   bool
)

// LoadEnv loads environment variables from .env file
// It searches for .env in the current directory and parent directories
// This function is safe to call multiple times - it only loads once
func LoadEnv() error {
	var err error

	envLoadOnce.Do(func() {
		// Try to find .env file in current and parent directories
		paths := []string{
			".env",
			"../.env",
			"../../.env",
			"../../../.env",
		}

		for _, path := range paths {
			if _, statErr := os.Stat(path); statErr == nil {
				err = godotenv.Load(path)
				if err == nil {
					envLoaded = true
					log.Printf("Loaded environment from: %s", path)
					return
				}
			}
		}

		// Also try to find .env in the repository root
		// Go up until we find go.mod or reach the filesystem root
		dir, _ := os.Getwd()
		for {
			envPath := filepath.Join(dir, ".env")
			if _, statErr := os.Stat(envPath); statErr == nil {
				err = godotenv.Load(envPath)
				if err == nil {
					envLoaded = true
					log.Printf("Loaded environment from: %s", envPath)
					return
				}
			}

			// Check if we found go.mod (repository root)
			goModPath := filepath.Join(dir, "go.mod")
			if _, statErr := os.Stat(goModPath); statErr == nil {
				// We're at the root, stop searching
				break
			}

			// Go up one directory
			parent := filepath.Dir(dir)
			if parent == dir {
				// Reached filesystem root
				break
			}
			dir = parent
		}

		// If we get here, .env was not found - this is OK
		// Environment variables might be set directly
		log.Println("No .env file found, using system environment variables")
	})

	return err
}

// IsEnvLoaded returns true if .env file was successfully loaded
func IsEnvLoaded() bool {
	return envLoaded
}

// GetEnv gets an environment variable with a default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetRequiredEnv gets a required environment variable
// Returns empty string if not found
func GetRequiredEnv(key string) string {
	return os.Getenv(key)
}

// MustGetEnv gets a required environment variable
// Panics if not found
func MustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Required environment variable not set: %s", key)
	}
	return value
}

// HasCredentials checks if all required BB API credentials are set
func HasCredentials() bool {
	return os.Getenv("BB_CLIENT_ID") != "" &&
		os.Getenv("BB_CLIENT_SECRET") != "" &&
		os.Getenv("BB_DEV_APP_KEY") != ""
}

// GetBBConfig returns a map with BB configuration from environment
func GetBBConfig() map[string]string {
	return map[string]string{
		"environment":     GetEnv("BB_ENVIRONMENT", "sandbox"),
		"client_id":       GetRequiredEnv("BB_CLIENT_ID"),
		"client_secret":   GetRequiredEnv("BB_CLIENT_SECRET"),
		"dev_app_key":     GetRequiredEnv("BB_DEV_APP_KEY"),
		"timeout":         GetEnv("BB_TIMEOUT_SECONDS", "30"),
		"retry_count":     GetEnv("BB_RETRY_COUNT", "3"),
		"retry_delay_ms":  GetEnv("BB_RETRY_DELAY_MS", "100"),
		"log_level":       GetEnv("BB_LOG_LEVEL", "info"),
		"log_format":      GetEnv("BB_LOG_FORMAT", "text"),
		"pix_key":         GetEnv("BB_PIX_KEY", ""),
	}
}
