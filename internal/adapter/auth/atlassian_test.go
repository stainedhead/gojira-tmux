package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stainedhead/gojira-tmux/internal/adapter/auth"
	"github.com/stainedhead/gojira-tmux/internal/adapter/jira/testutil"
)

func TestAtlassianAdapter_ValidateToken_Success(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetMyselfResponse(testutil.MyselfResponse("Test User", "user@example.com", "abc123"))

	store := auth.NewMemoryTokenStore()
	adapter := auth.NewAtlassianAdapter(store, srv.BaseURL())

	email, err := adapter.ValidateToken(context.Background(), "test@example.com", "test-token")
	if err != nil {
		t.Fatalf("ValidateToken() unexpected error: %v", err)
	}
	if email != "user@example.com" {
		t.Errorf("ValidateToken() email = %q, want %q", email, "user@example.com")
	}

	// Assert v3 endpoint was called
	if srv.MyselfRequestCount() != 1 {
		t.Errorf("expected 1 request to /rest/api/3/myself, got %d", srv.MyselfRequestCount())
	}

	// Verify the request went to the correct path
	req := srv.LastMyselfRequest()
	if req == nil {
		t.Fatal("no myself request recorded")
	}
	if req.Path != "/rest/api/3/myself" {
		t.Errorf("request path = %q, want %q", req.Path, "/rest/api/3/myself")
	}
}

func TestAtlassianAdapter_ValidateToken_BasicAuth(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetMyselfResponse(testutil.MyselfResponse("Test User", "user@example.com", "abc123"))

	store := auth.NewMemoryTokenStore()
	adapter := auth.NewAtlassianAdapter(store, srv.BaseURL())

	_, err := adapter.ValidateToken(context.Background(), "test@example.com", "test-token")
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	// Verify Authorization header was sent with Basic scheme
	req := srv.LastMyselfRequest()
	if req == nil {
		t.Fatal("no myself request recorded")
	}
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		t.Error("no Authorization header sent")
	}
	if !strings.HasPrefix(authHeader, "Basic ") {
		t.Errorf("Authorization header = %q, want Basic scheme", authHeader)
	}
}

func TestAtlassianAdapter_ValidateToken_Unauthorized(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	// Mock server expects test@example.com / test-token; send wrong credentials
	store := auth.NewMemoryTokenStore()
	adapter := auth.NewAtlassianAdapter(store, srv.BaseURL())

	_, err := adapter.ValidateToken(context.Background(), "wrong@example.com", "bad-token")
	if err == nil {
		t.Fatal("ValidateToken() = nil, want error")
	}
	if !strings.Contains(err.Error(), "invalid token or email") {
		t.Errorf("ValidateToken() error = %q, want containing %q", err.Error(), "invalid token or email")
	}
}

func TestAtlassianAdapter_ValidateToken_Forbidden(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetAuth("", "") // disable auth check so we can test 403 directly
	srv.SetMyselfError(http.StatusForbidden, testutil.ErrorResponse("Forbidden"))

	store := auth.NewMemoryTokenStore()
	adapter := auth.NewAtlassianAdapter(store, srv.BaseURL())

	_, err := adapter.ValidateToken(context.Background(), "user@example.com", "token")
	if err == nil {
		t.Fatal("ValidateToken() = nil, want error")
	}
	if !strings.Contains(err.Error(), "insufficient permissions") {
		t.Errorf("ValidateToken() error = %q, want containing %q", err.Error(), "insufficient permissions")
	}
}

func TestAtlassianAdapter_ValidateToken_CAPTCHA(t *testing.T) {
	// CAPTCHA lockout requires X-Seraph-LoginReason header - use custom handler
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Seraph-LoginReason", "AUTHENTICATION_DENIED")
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	store := auth.NewMemoryTokenStore()
	adapter := auth.NewAtlassianAdapter(store, server.URL)

	_, err := adapter.ValidateToken(context.Background(), "user@example.com", "token")
	if err == nil {
		t.Fatal("ValidateToken() = nil, want error")
	}
	if !strings.Contains(err.Error(), "too many failed login attempts") {
		t.Errorf("ValidateToken() error = %q, want containing %q", err.Error(), "too many failed login attempts")
	}
}

func TestAtlassianAdapter_ValidateToken_UnexpectedStatus(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	srv.SetAuth("", "") // disable auth check
	srv.SetMyselfError(http.StatusInternalServerError, testutil.ErrorResponse("Internal error"))

	store := auth.NewMemoryTokenStore()
	adapter := auth.NewAtlassianAdapter(store, srv.BaseURL())

	_, err := adapter.ValidateToken(context.Background(), "user@example.com", "token")
	if err == nil {
		t.Fatal("ValidateToken() = nil, want error")
	}
	if !strings.Contains(err.Error(), "unexpected status: 500") {
		t.Errorf("ValidateToken() error = %q, want containing %q", err.Error(), "unexpected status: 500")
	}
}

func TestAtlassianAdapter_ValidateToken_InvalidJSON(t *testing.T) {
	// Invalid JSON requires custom handler
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	store := auth.NewMemoryTokenStore()
	adapter := auth.NewAtlassianAdapter(store, server.URL)

	_, err := adapter.ValidateToken(context.Background(), "user@example.com", "token")
	if err == nil {
		t.Fatal("ValidateToken() = nil, want error")
	}
	if !strings.Contains(err.Error(), "failed to decode response") {
		t.Errorf("ValidateToken() error = %q, want containing %q", err.Error(), "failed to decode response")
	}
}

func TestAtlassianAdapter_ValidateToken_NetworkError(t *testing.T) {
	store := auth.NewMemoryTokenStore()
	// Use invalid URL to trigger network error
	adapter := auth.NewAtlassianAdapter(store, "http://invalid.local.invalid:1")

	_, err := adapter.ValidateToken(context.Background(), "user@example.com", "token")
	if err == nil {
		t.Fatal("ValidateToken() = nil, want network error")
	}
	if !strings.Contains(err.Error(), "failed to validate token") {
		t.Errorf("ValidateToken() error = %q, want containing %q", err.Error(), "failed to validate token")
	}
}

func TestAtlassianAdapter_ValidateToken_AcceptHeader(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	store := auth.NewMemoryTokenStore()
	adapter := auth.NewAtlassianAdapter(store, srv.BaseURL())

	_, _ = adapter.ValidateToken(context.Background(), "test@example.com", "test-token")

	req := srv.LastMyselfRequest()
	if req == nil {
		t.Fatal("no myself request recorded")
	}
	if req.Header.Get("Accept") != "application/json" {
		t.Errorf("Accept header = %q, want %q", req.Header.Get("Accept"), "application/json")
	}
}

func TestAtlassianAdapter_ValidateToken_ContextCancelled(t *testing.T) {
	srv := testutil.NewMockServer(t)
	defer srv.Close()

	store := auth.NewMemoryTokenStore()
	adapter := auth.NewAtlassianAdapter(store, srv.BaseURL())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := adapter.ValidateToken(ctx, "test@example.com", "test-token")
	if err == nil {
		t.Fatal("ValidateToken() = nil, want context cancelled error")
	}
}

func TestAtlassianAdapter_ValidateToken_V3ResponseFormat(t *testing.T) {
	// Verify we correctly parse a realistic v3 /myself response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"self":         "https://example.atlassian.net/rest/api/3/user?accountId=abc123",
			"accountId":    "abc123",
			"accountType":  "atlassian",
			"emailAddress": "v3user@example.com",
			"displayName":  "V3 User",
			"active":       true,
			"timeZone":     "America/New_York",
			"locale":       "en_US",
		})
	}))
	defer server.Close()

	store := auth.NewMemoryTokenStore()
	adapter := auth.NewAtlassianAdapter(store, server.URL)

	email, err := adapter.ValidateToken(context.Background(), "v3user@example.com", "token")
	if err != nil {
		t.Fatalf("ValidateToken() unexpected error: %v", err)
	}
	if email != "v3user@example.com" {
		t.Errorf("ValidateToken() email = %q, want %q", email, "v3user@example.com")
	}
}

func TestAtlassianAdapter_IsTokenValid(t *testing.T) {
	store := auth.NewMemoryTokenStore()
	adapter := auth.NewAtlassianAdapter(store, "https://example.atlassian.net")

	ctx := context.Background()

	// No token stored
	if adapter.IsTokenValid(ctx) {
		t.Error("IsTokenValid() = true, want false when no token stored")
	}

	// Set token
	_ = store.SetJiraToken("my-token")
	if !adapter.IsTokenValid(ctx) {
		t.Error("IsTokenValid() = false, want true when token stored")
	}

	// Delete token
	_ = store.DeleteJiraToken()
	if adapter.IsTokenValid(ctx) {
		t.Error("IsTokenValid() = true, want false after deleting token")
	}
}
