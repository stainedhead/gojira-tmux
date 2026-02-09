package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// Authenticate handles the authentication flow using Atlassian API tokens.
type Authenticate struct {
	authPort   domain.AuthPort
	tokenStore domain.TokenStorePort
}

// NewAuthenticate creates a new Authenticate use case.
func NewAuthenticate(authPort domain.AuthPort, tokenStore domain.TokenStorePort) *Authenticate {
	return &Authenticate{
		authPort:   authPort,
		tokenStore: tokenStore,
	}
}

// ValidateAndSaveToken validates credentials against the Jira API and saves the token.
func (a *Authenticate) ValidateAndSaveToken(ctx context.Context, email, token string) error {
	email = strings.TrimSpace(email)
	token = strings.TrimSpace(token)

	if email == "" || token == "" {
		return errors.New("email and token are required")
	}

	validatedEmail, err := a.authPort.ValidateToken(ctx, email, token)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	if !strings.EqualFold(validatedEmail, email) {
		return fmt.Errorf("token email mismatch: expected %s, got %s", email, validatedEmail)
	}

	if err := a.tokenStore.SetJiraToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}

// HasValidToken checks if a valid token exists.
func (a *Authenticate) HasValidToken(ctx context.Context) bool {
	return a.authPort.IsTokenValid(ctx)
}

// ClearToken removes the stored token.
func (a *Authenticate) ClearToken() error {
	return a.tokenStore.DeleteJiraToken()
}
