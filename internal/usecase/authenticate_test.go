package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stainedhead/gojira-tmux/internal/adapter/auth"
	"github.com/stainedhead/gojira-tmux/internal/domain"
	"github.com/stainedhead/gojira-tmux/internal/usecase"
)

// MockAuthPort is a mock implementation of domain.AuthPort.
type MockAuthPort struct {
	ValidateTokenFunc func(ctx context.Context, email, token string) (string, error)
	IsTokenValidFunc  func(ctx context.Context) bool
}

func (m *MockAuthPort) ValidateToken(ctx context.Context, email, token string) (string, error) {
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(ctx, email, token)
	}
	return email, nil
}

func (m *MockAuthPort) IsTokenValid(ctx context.Context) bool {
	if m.IsTokenValidFunc != nil {
		return m.IsTokenValidFunc(ctx)
	}
	return false
}

// MockConfigPort is a mock implementation of domain.ConfigPort.
type MockConfigPort struct {
	LoadFunc           func() (*domain.Config, error)
	GetProjectsFunc    func() []domain.Project
	GetTeamMembersFunc func() []domain.TeamMember
}

func (m *MockConfigPort) Load() (*domain.Config, error) {
	if m.LoadFunc != nil {
		return m.LoadFunc()
	}
	return &domain.Config{
		Team: []domain.TeamMember{
			{Name: "Test User", Email: "test@example.com"},
		},
	}, nil
}

func (m *MockConfigPort) GetProjects() []domain.Project {
	if m.GetProjectsFunc != nil {
		return m.GetProjectsFunc()
	}
	return []domain.Project{{Key: "PROJ", Name: "Test Project"}}
}

func (m *MockConfigPort) GetTeamMembers() []domain.TeamMember {
	if m.GetTeamMembersFunc != nil {
		return m.GetTeamMembersFunc()
	}
	return []domain.TeamMember{{Name: "Test User", Email: "test@example.com"}}
}

func TestAuthenticate_ValidateAndSaveToken_Success(t *testing.T) {
	authPort := &MockAuthPort{
		ValidateTokenFunc: func(_ context.Context, email, _ string) (string, error) {
			return email, nil
		},
	}
	tokenStore := auth.NewMemoryTokenStore()

	uc := usecase.NewAuthenticate(authPort, tokenStore)

	ctx := context.Background()
	err := uc.ValidateAndSaveToken(ctx, "user@example.com", "valid-token")

	if err != nil {
		t.Errorf("ValidateAndSaveToken() error = %v", err)
	}

	// Verify token was saved
	got, _ := tokenStore.GetJiraToken()
	if got != "valid-token" {
		t.Errorf("stored token = %q, want %q", got, "valid-token")
	}
}

func TestAuthenticate_ValidateAndSaveToken_InvalidCredentials(t *testing.T) {
	authPort := &MockAuthPort{
		ValidateTokenFunc: func(_ context.Context, _, _ string) (string, error) {
			return "", errors.New("invalid token or email")
		},
	}
	tokenStore := auth.NewMemoryTokenStore()

	uc := usecase.NewAuthenticate(authPort, tokenStore)

	ctx := context.Background()
	err := uc.ValidateAndSaveToken(ctx, "user@example.com", "bad-token")

	if err == nil {
		t.Error("ValidateAndSaveToken() = nil, want error")
	}

	// Verify token was NOT saved
	if tokenStore.HasJiraToken() {
		t.Error("token should not have been saved on validation failure")
	}
}

func TestAuthenticate_ValidateAndSaveToken_EmailMismatch(t *testing.T) {
	authPort := &MockAuthPort{
		ValidateTokenFunc: func(_ context.Context, _, _ string) (string, error) {
			return "other@example.com", nil
		},
	}
	tokenStore := auth.NewMemoryTokenStore()

	uc := usecase.NewAuthenticate(authPort, tokenStore)

	ctx := context.Background()
	err := uc.ValidateAndSaveToken(ctx, "user@example.com", "token")

	if err == nil {
		t.Error("ValidateAndSaveToken() = nil, want email mismatch error")
	}
}

func TestAuthenticate_ValidateAndSaveToken_EmptyInputs(t *testing.T) {
	authPort := &MockAuthPort{}
	tokenStore := auth.NewMemoryTokenStore()

	uc := usecase.NewAuthenticate(authPort, tokenStore)
	ctx := context.Background()

	tests := []struct {
		name  string
		email string
		token string
	}{
		{"empty email", "", "token"},
		{"empty token", "user@example.com", ""},
		{"both empty", "", ""},
		{"whitespace email", "   ", "token"},
		{"whitespace token", "user@example.com", "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := uc.ValidateAndSaveToken(ctx, tt.email, tt.token)
			if err == nil {
				t.Error("ValidateAndSaveToken() = nil, want error for empty inputs")
			}
		})
	}
}

func TestAuthenticate_HasValidToken(t *testing.T) {
	ctx := context.Background()

	t.Run("returns true when token exists", func(t *testing.T) {
		authPort := &MockAuthPort{
			IsTokenValidFunc: func(_ context.Context) bool {
				return true
			},
		}
		tokenStore := auth.NewMemoryTokenStore()
		uc := usecase.NewAuthenticate(authPort, tokenStore)

		if !uc.HasValidToken(ctx) {
			t.Error("HasValidToken() = false, want true")
		}
	})

	t.Run("returns false when no token", func(t *testing.T) {
		authPort := &MockAuthPort{
			IsTokenValidFunc: func(_ context.Context) bool {
				return false
			},
		}
		tokenStore := auth.NewMemoryTokenStore()
		uc := usecase.NewAuthenticate(authPort, tokenStore)

		if uc.HasValidToken(ctx) {
			t.Error("HasValidToken() = true, want false")
		}
	})
}

func TestAuthenticate_ClearToken(t *testing.T) {
	authPort := &MockAuthPort{}
	tokenStore := auth.NewMemoryTokenStore()
	_ = tokenStore.SetJiraToken("existing-token")

	uc := usecase.NewAuthenticate(authPort, tokenStore)

	err := uc.ClearToken()
	if err != nil {
		t.Errorf("ClearToken() error = %v", err)
	}

	if tokenStore.HasJiraToken() {
		t.Error("token should have been cleared")
	}
}
