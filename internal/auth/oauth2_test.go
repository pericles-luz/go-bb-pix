package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewOAuth2Provider(t *testing.T) {
	provider := NewOAuth2Provider("https://oauth.example.com/token", "client-id", "client-secret")

	if provider == nil {
		t.Fatal("NewOAuth2Provider returned nil")
	}

	if provider.tokenURL != "https://oauth.example.com/token" {
		t.Errorf("tokenURL = %q, want %q", provider.tokenURL, "https://oauth.example.com/token")
	}
	if provider.clientID != "client-id" {
		t.Errorf("clientID = %q, want %q", provider.clientID, "client-id")
	}
	if provider.clientSecret != "client-secret" {
		t.Errorf("clientSecret = %q, want %q", provider.clientSecret, "client-secret")
	}
}

func TestOAuth2Provider_GetToken_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("Method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/token" {
			t.Errorf("Path = %q, want /token", r.URL.Path)
		}

		// Verify basic auth
		username, password, ok := r.BasicAuth()
		if !ok {
			t.Error("Basic auth not present")
		}
		if username != "client-id" {
			t.Errorf("Basic auth username = %q, want client-id", username)
		}
		if password != "client-secret" {
			t.Errorf("Basic auth password = %q, want client-secret", password)
		}

		// Verify form data
		if err := r.ParseForm(); err != nil {
			t.Fatalf("Failed to parse form: %v", err)
		}
		if r.FormValue("grant_type") != "client_credentials" {
			t.Errorf("grant_type = %q, want client_credentials", r.FormValue("grant_type"))
		}

		// Return token
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "test-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer server.Close()

	provider := NewOAuth2Provider(server.URL+"/token", "client-id", "client-secret")
	token, err := provider.GetToken(context.Background())

	if err != nil {
		t.Fatalf("GetToken() error = %v", err)
	}
	if token.AccessToken != "test-access-token" {
		t.Errorf("AccessToken = %q, want test-access-token", token.AccessToken)
	}
	if token.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want Bearer", token.TokenType)
	}
	if token.ExpiresIn != 3600 {
		t.Errorf("ExpiresIn = %d, want 3600", token.ExpiresIn)
	}
}

func TestOAuth2Provider_GetToken_CachedToken(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "test-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer server.Close()

	provider := NewOAuth2Provider(server.URL+"/token", "client-id", "client-secret")

	// First call should fetch token
	token1, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("First GetToken() error = %v", err)
	}

	// Second call should use cached token
	token2, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("Second GetToken() error = %v", err)
	}

	if callCount != 1 {
		t.Errorf("Server called %d times, want 1 (token should be cached)", callCount)
	}

	if token1.AccessToken != token2.AccessToken {
		t.Error("Cached token should be the same as first token")
	}
}

func TestOAuth2Provider_GetToken_ExpiredToken(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": fmt.Sprintf("token-%d", callCount),
			"token_type":   "Bearer",
			"expires_in":   1, // 1 second
		})
	}))
	defer server.Close()

	provider := NewOAuth2Provider(server.URL+"/token", "client-id", "client-secret")

	// First call
	token1, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("First GetToken() error = %v", err)
	}

	// Manually expire the token
	provider.mu.Lock()
	provider.cachedToken.IssuedAt = time.Now().Add(-10 * time.Minute)
	provider.mu.Unlock()

	// Second call should fetch new token
	token2, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("Second GetToken() error = %v", err)
	}

	if callCount != 2 {
		t.Errorf("Server called %d times, want 2 (expired token should be refreshed)", callCount)
	}

	if token1.AccessToken == token2.AccessToken {
		t.Error("New token should be different from expired token")
	}
}

func TestOAuth2Provider_GetToken_NetworkError(t *testing.T) {
	// Use invalid URL to simulate network error
	provider := NewOAuth2Provider("http://invalid-host-that-does-not-exist.local/token", "client-id", "client-secret")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := provider.GetToken(ctx)
	if err == nil {
		t.Fatal("Expected error for network failure, got nil")
	}
}

func TestOAuth2Provider_GetToken_401Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid_client"}`))
	}))
	defer server.Close()

	provider := NewOAuth2Provider(server.URL+"/token", "wrong-id", "wrong-secret")

	_, err := provider.GetToken(context.Background())
	if err == nil {
		t.Fatal("Expected error for 401 response, got nil")
	}

	if !strings.Contains(err.Error(), "401") {
		t.Errorf("Error should mention 401 status: %v", err)
	}
}

func TestOAuth2Provider_GetToken_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	provider := NewOAuth2Provider(server.URL+"/token", "client-id", "client-secret")

	_, err := provider.GetToken(context.Background())
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

func TestOAuth2Provider_Invalidate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "test-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer server.Close()

	provider := NewOAuth2Provider(server.URL+"/token", "client-id", "client-secret")

	// Get initial token
	_, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken() error = %v", err)
	}

	// Verify token is cached
	if provider.cachedToken == nil {
		t.Fatal("Token should be cached")
	}

	// Invalidate
	provider.Invalidate()

	// Verify token is cleared
	provider.mu.RLock()
	cached := provider.cachedToken
	provider.mu.RUnlock()

	if cached != nil {
		t.Error("Token should be nil after Invalidate()")
	}
}

func TestOAuth2Provider_Concurrency(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		mu.Unlock()

		// Simulate slow response to increase chance of concurrent requests
		time.Sleep(50 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "test-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer server.Close()

	provider := NewOAuth2Provider(server.URL+"/token", "client-id", "client-secret")

	// Launch multiple concurrent GetToken calls
	numGoroutines := 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_, err := provider.GetToken(context.Background())
			if err != nil {
				t.Errorf("GetToken() error = %v", err)
			}
		}()
	}

	wg.Wait()

	// With proper mutex locking, only one request should be made
	// (first goroutine fetches, others wait for cache)
	if callCount > 1 {
		t.Errorf("Server called %d times, want 1 (concurrent requests should use mutex)", callCount)
	}
}

func TestOAuth2Provider_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(200 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "test-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer server.Close()

	provider := NewOAuth2Provider(server.URL+"/token", "client-id", "client-secret")

	// Create context that will be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := provider.GetToken(ctx)
	if err == nil {
		t.Fatal("Expected error for cancelled context, got nil")
	}

	if !strings.Contains(err.Error(), "context") {
		t.Errorf("Error should mention context: %v", err)
	}
}
