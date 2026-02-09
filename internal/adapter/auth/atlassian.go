package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// AtlassianAdapter implements the AuthPort interface for Atlassian API token authentication.
type AtlassianAdapter struct {
	tokenStore domain.TokenStorePort
	httpClient *http.Client
	jiraURL    string
}

// NewAtlassianAdapter creates a new Atlassian adapter.
func NewAtlassianAdapter(tokenStore domain.TokenStorePort, jiraURL string) *AtlassianAdapter {
	return &AtlassianAdapter{
		tokenStore: tokenStore,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		jiraURL:    jiraURL,
	}
}

// ValidateToken validates an Atlassian API token by calling the Jira API.
// Returns the validated email address on success.
func (a *AtlassianAdapter) ValidateToken(ctx context.Context, email, token string) (string, error) {
	url := a.jiraURL + "/rest/api/2/myself"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set Basic Auth: base64(email:token)
	auth := base64.StdEncoding.EncodeToString([]byte(email + ":" + token))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	// Check for CAPTCHA lockout
	if resp.Header.Get("X-Seraph-LoginReason") == "AUTHENTICATION_DENIED" {
		return "", errors.New("too many failed login attempts, account temporarily locked")
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return "", errors.New("invalid token or email")
	}
	if resp.StatusCode == http.StatusForbidden {
		return "", errors.New("insufficient permissions")
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		EmailAddress string `json:"emailAddress"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.EmailAddress, nil
}

// IsTokenValid checks if a valid token exists in the store.
func (a *AtlassianAdapter) IsTokenValid(_ context.Context) bool {
	token, err := a.tokenStore.GetJiraToken()
	return err == nil && token != ""
}

// Ensure AtlassianAdapter implements domain.AuthPort.
var _ domain.AuthPort = (*AtlassianAdapter)(nil)
