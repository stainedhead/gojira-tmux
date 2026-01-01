package usecase

import (
	"errors"
	"strings"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// SetupToken handles first-time API token setup.
type SetupToken struct {
	tokenStore domain.TokenStorePort
}

// NewSetupToken creates a new SetupToken use case.
func NewSetupToken(tokenStore domain.TokenStorePort) *SetupToken {
	return &SetupToken{
		tokenStore: tokenStore,
	}
}

// NeedsSetup returns true if no API token is configured.
func (s *SetupToken) NeedsSetup() bool {
	return !s.tokenStore.HasJiraToken()
}

// SaveToken validates and saves the API token.
func (s *SetupToken) SaveToken(token string) error {
	// Trim whitespace
	token = strings.TrimSpace(token)

	// Validate
	if token == "" {
		return errors.New("token cannot be empty")
	}

	// Store the token
	return s.tokenStore.SetJiraToken(token)
}
