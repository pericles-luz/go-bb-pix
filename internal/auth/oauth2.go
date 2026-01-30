package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// OAuth2Provider implements TokenProvider using OAuth2 Client Credentials flow
type OAuth2Provider struct {
	tokenURL     string
	clientID     string
	clientSecret string

	mu           sync.RWMutex
	cachedToken  *Token
	httpClient   *http.Client
}

// tokenResponse represents the OAuth2 token response
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// NewOAuth2Provider creates a new OAuth2Provider
func NewOAuth2Provider(tokenURL, clientID, clientSecret string) *OAuth2Provider {
	return &OAuth2Provider{
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetToken returns a valid access token
// If the current token is expired or doesn't exist, it will fetch a new one
func (p *OAuth2Provider) GetToken(ctx context.Context) (*Token, error) {
	// Check if we have a valid cached token (read lock)
	p.mu.RLock()
	if p.cachedToken != nil && !p.cachedToken.IsExpired() {
		token := p.cachedToken
		p.mu.RUnlock()
		return token, nil
	}
	p.mu.RUnlock()

	// Need to fetch new token (write lock)
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine might have fetched it)
	if p.cachedToken != nil && !p.cachedToken.IsExpired() {
		return p.cachedToken, nil
	}

	// Fetch new token
	token, err := p.fetchToken(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the token
	p.cachedToken = token
	return token, nil
}

// Invalidate marks the current token as invalid
func (p *OAuth2Provider) Invalidate() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cachedToken = nil
}

// fetchToken fetches a new token from the OAuth2 server
func (p *OAuth2Provider) fetchToken(ctx context.Context) (*Token, error) {
	// Prepare request body
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(p.clientID, p.clientSecret)

	// Execute request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Create token
	token := &Token{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		ExpiresIn:   tokenResp.ExpiresIn,
		IssuedAt:    time.Now(),
	}

	return token, nil
}
