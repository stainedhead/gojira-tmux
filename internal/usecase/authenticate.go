package usecase

import (
	"context"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// Authenticate handles the authentication flow.
type Authenticate struct {
	authPort   domain.AuthPort
	configPort domain.ConfigPort
}

// NewAuthenticate creates a new Authenticate use case.
func NewAuthenticate(authPort domain.AuthPort, configPort domain.ConfigPort) *Authenticate {
	return &Authenticate{
		authPort:   authPort,
		configPort: configPort,
	}
}

// StartLogin initiates the login flow.
// Returns the auth URL to open in the browser.
func (a *Authenticate) StartLogin(ctx context.Context) (string, error) {
	return a.authPort.StartAuthFlow(ctx)
}

// CompleteLogin waits for and processes the OAuth callback.
// Returns the authenticated user after validating team membership.
func (a *Authenticate) CompleteLogin(ctx context.Context) (*domain.User, error) {
	// Wait for callback to complete
	user, err := a.authPort.WaitForCallback(ctx)
	if err != nil {
		return nil, err
	}

	// Validate team membership
	if err := a.configPort.ValidateUserAccess(user.Email); err != nil {
		// User authenticated but not in team
		_ = a.authPort.Logout()
		return nil, err
	}

	return user, nil
}

// CancelLogin cancels an in-progress login.
func (a *Authenticate) CancelLogin() {
	a.authPort.CancelAuthFlow()
}

// CheckSession checks if user has a valid session.
// Returns the user if session is valid, nil otherwise.
func (a *Authenticate) CheckSession(ctx context.Context) (*domain.User, error) {
	if !a.authPort.IsSessionValid() {
		return nil, nil
	}

	// Try to refresh the session
	user, err := a.authPort.RefreshSession(ctx)
	if err != nil {
		// Refresh failed, session is invalid
		return nil, nil
	}

	return user, nil
}

// Logout logs out the current user.
func (a *Authenticate) Logout() error {
	return a.authPort.Logout()
}
