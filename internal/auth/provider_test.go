package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockTokenProvider implements TokenProvider for testing
type mockTokenProvider struct {
	token *Token
	err   error
}

func (m *mockTokenProvider) GetToken(ctx context.Context) (*Token, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.token, nil
}

func (m *mockTokenProvider) Invalidate() {
	m.token = nil
}

func TestTokenProvider_Interface(t *testing.T) {
	// Test that mockTokenProvider implements TokenProvider
	var _ TokenProvider = (*mockTokenProvider)(nil)
}

func TestMockTokenProvider_GetToken(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expectedToken := &Token{
			AccessToken: "test-token",
			ExpiresIn:   3600,
			IssuedAt:    time.Now(),
		}

		mock := &mockTokenProvider{token: expectedToken}
		token, err := mock.GetToken(ctx)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token.AccessToken != expectedToken.AccessToken {
			t.Errorf("AccessToken = %q, want %q", token.AccessToken, expectedToken.AccessToken)
		}
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("auth error")
		mock := &mockTokenProvider{err: expectedErr}

		token, err := mock.GetToken(ctx)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != expectedErr {
			t.Errorf("error = %v, want %v", err, expectedErr)
		}
		if token != nil {
			t.Errorf("token should be nil on error")
		}
	})
}

func TestMockTokenProvider_Invalidate(t *testing.T) {
	mock := &mockTokenProvider{
		token: &Token{AccessToken: "test-token"},
	}

	if mock.token == nil {
		t.Fatal("token should be set before invalidate")
	}

	mock.Invalidate()

	if mock.token != nil {
		t.Error("token should be nil after invalidate")
	}
}

func TestToken_IsExpired(t *testing.T) {
	tests := []struct {
		name    string
		token   *Token
		want    bool
	}{
		{
			name: "not expired",
			token: &Token{
				AccessToken: "token",
				ExpiresIn:   3600,
				IssuedAt:    time.Now(),
			},
			want: false,
		},
		{
			name: "expired",
			token: &Token{
				AccessToken: "token",
				ExpiresIn:   1,
				IssuedAt:    time.Now().Add(-2 * time.Second),
			},
			want: true,
		},
		{
			name: "expires soon (within 5 minutes)",
			token: &Token{
				AccessToken: "token",
				ExpiresIn:   300, // 5 minutes
				IssuedAt:    time.Now().Add(-60 * time.Second), // issued 1 minute ago
			},
			want: true, // Should be considered expired (4 minutes remaining < 5 minute threshold)
		},
		{
			name: "just issued",
			token: &Token{
				AccessToken: "token",
				ExpiresIn:   600, // 10 minutes
				IssuedAt:    time.Now(),
			},
			want: false, // 10 minutes - 5 minute threshold = 5 minutes remaining
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.token.IsExpired()
			if got != tt.want {
				expiresAt := tt.token.IssuedAt.Add(time.Duration(tt.token.ExpiresIn) * time.Second)
				remaining := time.Until(expiresAt)
				t.Errorf("IsExpired() = %v, want %v (remaining: %v)", got, tt.want, remaining)
			}
		})
	}
}

func TestToken_ExpiresAt(t *testing.T) {
	issuedAt := time.Now()
	expiresIn := 3600 // 1 hour

	token := &Token{
		AccessToken: "token",
		ExpiresIn:   expiresIn,
		IssuedAt:    issuedAt,
	}

	expectedExpiresAt := issuedAt.Add(time.Duration(expiresIn) * time.Second)
	gotExpiresAt := token.ExpiresAt()

	// Allow 1 second tolerance for test execution time
	diff := gotExpiresAt.Sub(expectedExpiresAt)
	if diff > time.Second || diff < -time.Second {
		t.Errorf("ExpiresAt() = %v, want %v (diff: %v)", gotExpiresAt, expectedExpiresAt, diff)
	}
}
