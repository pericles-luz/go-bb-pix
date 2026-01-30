package auth

import (
	"context"
	"time"
)

// Token represents an OAuth2 access token
type Token struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int       // seconds
	IssuedAt    time.Time
}

// IsExpired checks if the token is expired or about to expire
// Returns true if the token expires in less than 5 minutes
func (t *Token) IsExpired() bool {
	if t == nil {
		return true
	}

	// Consider token expired if it expires in less than 5 minutes
	// This gives us a buffer for clock skew and request time
	expirationThreshold := 5 * time.Minute
	expiresAt := t.IssuedAt.Add(time.Duration(t.ExpiresIn) * time.Second)
	timeUntilExpiration := time.Until(expiresAt)

	return timeUntilExpiration < expirationThreshold
}

// ExpiresAt returns the time when the token expires
func (t *Token) ExpiresAt() time.Time {
	return t.IssuedAt.Add(time.Duration(t.ExpiresIn) * time.Second)
}

// TokenProvider is an interface for obtaining OAuth2 access tokens
type TokenProvider interface {
	// GetToken returns a valid access token
	// If the current token is expired or doesn't exist, it will fetch a new one
	GetToken(ctx context.Context) (*Token, error)

	// Invalidate marks the current token as invalid, forcing a new token
	// to be fetched on the next GetToken call
	Invalidate()
}
