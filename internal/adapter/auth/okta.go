package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// OktaAdapter implements the AuthPort interface for Okta OIDC authentication.
type OktaAdapter struct {
	config       domain.OktaConfig
	tokenStore   domain.TokenStorePort
	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier
	provider     *oidc.Provider

	// State management
	mu            sync.Mutex
	currentState  string
	currentVerifier string
	currentUser   *domain.User
	callbackServer *CallbackServer
}

// NewOktaAdapter creates a new Okta adapter.
func NewOktaAdapter(cfg domain.OktaConfig, tokenStore domain.TokenStorePort) *OktaAdapter {
	return &OktaAdapter{
		config:     cfg,
		tokenStore: tokenStore,
	}
}

// initProvider initializes the OIDC provider and OAuth2 config.
func (a *OktaAdapter) initProvider(ctx context.Context) error {
	if a.provider != nil {
		return nil
	}

	provider, err := oidc.NewProvider(ctx, a.config.Issuer)
	if err != nil {
		return fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	a.provider = provider
	a.verifier = provider.Verifier(&oidc.Config{ClientID: a.config.ClientID})

	a.oauth2Config = &oauth2.Config{
		ClientID:    a.config.ClientID,
		Endpoint:    provider.Endpoint(),
		RedirectURL: fmt.Sprintf("http://localhost:%d/callback", a.config.CallbackPort),
		Scopes:      a.config.Scopes,
	}

	return nil
}

// GenerateAuthURL generates the authorization URL with PKCE.
func (a *OktaAdapter) GenerateAuthURL() (authURL, state string, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Generate state
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}
	a.currentState = base64.URLEncoding.EncodeToString(stateBytes)

	// Generate PKCE verifier
	a.currentVerifier = oauth2.GenerateVerifier()

	// Build auth URL manually if provider not initialized
	if a.oauth2Config == nil {
		// Build URL manually for testing
		authURL = fmt.Sprintf("%s/v1/authorize?client_id=%s&redirect_uri=http://localhost:%d/callback&response_type=code&scope=openid%%20profile%%20email&state=%s&code_challenge=%s&code_challenge_method=S256",
			a.config.Issuer,
			a.config.ClientID,
			a.config.CallbackPort,
			a.currentState,
			oauth2.S256ChallengeFromVerifier(a.currentVerifier),
		)
		return authURL, a.currentState, nil
	}

	// Build URL with oauth2 package
	authURL = a.oauth2Config.AuthCodeURL(
		a.currentState,
		oauth2.AccessTypeOffline,
		oauth2.S256ChallengeOption(a.currentVerifier),
	)

	return authURL, a.currentState, nil
}

// ValidateState validates the state parameter from callback.
func (a *OktaAdapter) ValidateState(state string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	if state == "" || a.currentState == "" {
		return false
	}
	return state == a.currentState
}

// StartAuthFlow initiates the Okta OIDC flow.
func (a *OktaAdapter) StartAuthFlow(ctx context.Context) (authURL string, err error) {
	// Initialize provider
	if err := a.initProvider(ctx); err != nil {
		return "", err
	}

	// Generate auth URL
	authURL, _, err = a.GenerateAuthURL()
	if err != nil {
		return "", err
	}

	// Start callback server
	a.callbackServer = NewCallbackServer(a.config.CallbackPort)
	_, err = a.callbackServer.Start()
	if err != nil {
		return "", fmt.Errorf("failed to start callback server: %w", err)
	}

	return authURL, nil
}

// WaitForCallback waits for the OAuth callback.
func (a *OktaAdapter) WaitForCallback(ctx context.Context) (*domain.User, error) {
	if a.callbackServer == nil {
		return nil, errors.New("callback server not started")
	}

	defer a.callbackServer.Stop()

	// Wait for code
	code, state, err := a.callbackServer.WaitForCode(ctx)
	if err != nil {
		return nil, err
	}

	// Validate state
	if !a.ValidateState(state) {
		return nil, errors.New("invalid state parameter")
	}

	// Exchange code for tokens
	a.mu.Lock()
	verifier := a.currentVerifier
	a.mu.Unlock()

	token, err := a.oauth2Config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for tokens: %w", err)
	}

	// Extract ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token in response")
	}

	// Verify ID token
	idToken, err := a.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	// Extract claims
	var claims struct {
		Email string `json:"email"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to extract claims: %w", err)
	}

	// Store refresh token if available
	if token.RefreshToken != "" {
		_ = a.tokenStore.SetRefreshToken(token.RefreshToken)
	}

	// Create user with 8-hour session
	user := &domain.User{
		Email:         claims.Email,
		SessionExpiry: time.Now().Add(8 * time.Hour),
	}

	a.mu.Lock()
	a.currentUser = user
	a.mu.Unlock()

	return user, nil
}

// CancelAuthFlow cancels an in-progress authentication.
func (a *OktaAdapter) CancelAuthFlow() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.callbackServer != nil {
		a.callbackServer.Cancel()
	}
}

// RefreshSession refreshes the user session if refresh token exists.
func (a *OktaAdapter) RefreshSession(ctx context.Context) (*domain.User, error) {
	refreshToken, err := a.tokenStore.GetRefreshToken()
	if err != nil || refreshToken == "" {
		return nil, errors.New("no refresh token available")
	}

	// Initialize provider if needed
	if err := a.initProvider(ctx); err != nil {
		return nil, err
	}

	// Use refresh token to get new tokens
	tokenSource := a.oauth2Config.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	})

	token, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Extract ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token in refresh response")
	}

	// Verify ID token
	idToken, err := a.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify refreshed ID token: %w", err)
	}

	// Extract claims
	var claims struct {
		Email string `json:"email"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to extract claims: %w", err)
	}

	// Store new refresh token if available
	if token.RefreshToken != "" && token.RefreshToken != refreshToken {
		_ = a.tokenStore.SetRefreshToken(token.RefreshToken)
	}

	// Create user with 8-hour session
	user := &domain.User{
		Email:         claims.Email,
		SessionExpiry: time.Now().Add(8 * time.Hour),
	}

	a.mu.Lock()
	a.currentUser = user
	a.mu.Unlock()

	return user, nil
}

// IsSessionValid checks if current session is still valid.
func (a *OktaAdapter) IsSessionValid() bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentUser == nil {
		return false
	}
	return a.currentUser.IsSessionValid()
}

// Logout clears the current session.
func (a *OktaAdapter) Logout() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.currentUser = nil
	a.currentState = ""
	a.currentVerifier = ""

	// Clear stored refresh token
	_ = a.tokenStore.DeleteRefreshToken()

	return nil
}

// Ensure OktaAdapter implements domain.AuthPort.
var _ domain.AuthPort = (*OktaAdapter)(nil)
