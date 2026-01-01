package auth_test

import (
	"testing"

	"github.com/stainedhead/gojira-tmux/internal/adapter/auth"
	"github.com/stainedhead/gojira-tmux/internal/domain"
)

func TestOktaAdapter_GenerateAuthURL(t *testing.T) {
	cfg := domain.OktaConfig{
		Issuer:       "https://example.okta.com/oauth2/default",
		ClientID:     "test-client-id",
		CallbackPort: 8080,
		Scopes:       []string{"openid", "profile", "email"},
	}

	adapter := auth.NewOktaAdapter(cfg, auth.NewMemoryTokenStore())

	authURL, state, err := adapter.GenerateAuthURL()
	if err != nil {
		t.Fatalf("GenerateAuthURL() error = %v", err)
	}

	if authURL == "" {
		t.Error("GenerateAuthURL() returned empty URL")
	}

	if state == "" {
		t.Error("GenerateAuthURL() returned empty state")
	}

	// Verify URL contains expected components
	expectedParts := []string{
		"https://example.okta.com",
		"client_id=test-client-id",
		"redirect_uri=",
		"scope=openid",
		"code_challenge=",
		"code_challenge_method=S256",
	}

	for _, part := range expectedParts {
		if !containsSubstring(authURL, part) {
			t.Errorf("GenerateAuthURL() URL missing %q", part)
		}
	}
}

func TestOktaAdapter_ValidateState(t *testing.T) {
	cfg := domain.OktaConfig{
		Issuer:       "https://example.okta.com/oauth2/default",
		ClientID:     "test-client-id",
		CallbackPort: 8080,
		Scopes:       []string{"openid", "profile", "email"},
	}

	adapter := auth.NewOktaAdapter(cfg, auth.NewMemoryTokenStore())

	// Generate auth URL to get state
	_, state, err := adapter.GenerateAuthURL()
	if err != nil {
		t.Fatalf("GenerateAuthURL() error = %v", err)
	}

	// Valid state should pass
	if !adapter.ValidateState(state) {
		t.Error("ValidateState() returned false for valid state")
	}

	// Invalid state should fail
	if adapter.ValidateState("invalid-state") {
		t.Error("ValidateState() returned true for invalid state")
	}

	// Empty state should fail
	if adapter.ValidateState("") {
		t.Error("ValidateState() returned true for empty state")
	}
}

func TestOktaAdapter_IsSessionValid(t *testing.T) {
	cfg := domain.OktaConfig{
		Issuer:       "https://example.okta.com/oauth2/default",
		ClientID:     "test-client-id",
		CallbackPort: 8080,
		Scopes:       []string{"openid", "profile", "email"},
	}

	adapter := auth.NewOktaAdapter(cfg, auth.NewMemoryTokenStore())

	// No session should be invalid
	if adapter.IsSessionValid() {
		t.Error("IsSessionValid() returned true for no session")
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
