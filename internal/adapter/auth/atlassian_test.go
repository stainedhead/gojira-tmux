package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stainedhead/gojira-tmux/internal/adapter/auth"
)

func TestAtlassianAdapter_ValidateToken(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		token      string
		handler    http.HandlerFunc
		wantEmail  string
		wantErr    bool
		errContains string
	}{
		{
			name:  "valid token",
			email: "user@example.com",
			token: "valid-token",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"emailAddress": "user@example.com",
				})
			},
			wantEmail: "user@example.com",
			wantErr:   false,
		},
		{
			name:  "unauthorized - invalid token",
			email: "user@example.com",
			token: "bad-token",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			},
			wantErr:     true,
			errContains: "invalid token or email",
		},
		{
			name:  "forbidden - insufficient permissions",
			email: "user@example.com",
			token: "token",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusForbidden)
			},
			wantErr:     true,
			errContains: "insufficient permissions",
		},
		{
			name:  "captcha lockout",
			email: "user@example.com",
			token: "token",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("X-Seraph-LoginReason", "AUTHENTICATION_DENIED")
				w.WriteHeader(http.StatusUnauthorized)
			},
			wantErr:     true,
			errContains: "too many failed login attempts",
		},
		{
			name:  "unexpected status code",
			email: "user@example.com",
			token: "token",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr:     true,
			errContains: "unexpected status: 500",
		},
		{
			name:  "invalid json response",
			email: "user@example.com",
			token: "token",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("not json"))
			},
			wantErr:     true,
			errContains: "failed to decode response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			store := auth.NewMemoryTokenStore()
			adapter := auth.NewAtlassianAdapter(store, server.URL)

			ctx := context.Background()
			email, err := adapter.ValidateToken(ctx, tt.email, tt.token)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateToken() = nil, want error containing %q", tt.errContains)
					return
				}
				if tt.errContains != "" && !containsStr(err.Error(), tt.errContains) {
					t.Errorf("ValidateToken() error = %q, want containing %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateToken() unexpected error: %v", err)
				return
			}
			if email != tt.wantEmail {
				t.Errorf("ValidateToken() email = %q, want %q", email, tt.wantEmail)
			}
		})
	}
}

func TestAtlassianAdapter_ValidateToken_BasicAuth(t *testing.T) {
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"emailAddress": "user@example.com",
		})
	}))
	defer server.Close()

	store := auth.NewMemoryTokenStore()
	adapter := auth.NewAtlassianAdapter(store, server.URL)

	ctx := context.Background()
	_, err := adapter.ValidateToken(ctx, "user@example.com", "my-token")
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if receivedAuth == "" {
		t.Error("no Authorization header sent")
	}
	// Should be "Basic base64(user@example.com:my-token)"
	if len(receivedAuth) < 6 || receivedAuth[:6] != "Basic " {
		t.Errorf("Authorization header = %q, want Basic scheme", receivedAuth)
	}
}

func TestAtlassianAdapter_ValidateToken_NetworkError(t *testing.T) {
	store := auth.NewMemoryTokenStore()
	// Use invalid URL to trigger network error
	adapter := auth.NewAtlassianAdapter(store, "http://invalid.local.invalid:1")

	ctx := context.Background()
	_, err := adapter.ValidateToken(ctx, "user@example.com", "token")

	if err == nil {
		t.Error("ValidateToken() = nil, want network error")
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

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
