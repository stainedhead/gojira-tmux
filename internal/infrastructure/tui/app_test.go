package tui_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/stainedhead/gojira-tmux/internal/adapter/auth"
	"github.com/stainedhead/gojira-tmux/internal/domain"
	"github.com/stainedhead/gojira-tmux/internal/infrastructure/tui"
)

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// mockAuthPort implements domain.AuthPort for testing.
type mockAuthPort struct {
	validateTokenFunc func(ctx context.Context, email, token string) (string, error)
	isTokenValidFunc  func(ctx context.Context) bool
}

func (m *mockAuthPort) ValidateToken(ctx context.Context, email, token string) (string, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(ctx, email, token)
	}
	return email, nil
}

func (m *mockAuthPort) IsTokenValid(ctx context.Context) bool {
	if m.isTokenValidFunc != nil {
		return m.isTokenValidFunc(ctx)
	}
	return false
}

// mockJiraPort implements domain.JiraPort for testing.
type mockJiraPort struct{}

func (m *mockJiraPort) SearchIssues(_ context.Context, _ domain.IssueFilter) ([]domain.Issue, error) {
	return nil, nil
}

func (m *mockJiraPort) GetIssue(_ context.Context, _ string) (*domain.Issue, error) {
	return nil, nil
}

func (m *mockJiraPort) GetIssueComments(_ context.Context, _ string) ([]domain.Comment, error) {
	return nil, nil
}

// mockConfigPort implements domain.ConfigPort for testing.
type mockConfigPort struct {
	team     []domain.TeamMember
	projects []domain.Project
}

func (m *mockConfigPort) Load() (*domain.Config, error) {
	return &domain.Config{
		Team:     m.team,
		Projects: m.projects,
	}, nil
}

func (m *mockConfigPort) GetProjects() []domain.Project {
	return m.projects
}

func (m *mockConfigPort) GetTeamMembers() []domain.TeamMember {
	return m.team
}

func TestApp_InitialState_NoToken_ShowsSetup(t *testing.T) {
	tokenStore := auth.NewMemoryTokenStore()

	app := tui.NewApp(
		tui.WithTokenStore(tokenStore),
		tui.WithAuthPort(&mockAuthPort{}),
	)

	// Init should return a command
	cmd := app.Init()
	if cmd == nil {
		t.Fatal("Init() returned nil command")
	}

	// Execute the command - should get screenChangeMsg for setup
	msg := cmd()
	if msg == nil {
		t.Fatal("Init command returned nil message")
	}

	// Update with the message - should show setup screen
	model, _ := app.Update(msg)
	view := model.View()

	// Setup screen should contain token-related instructions
	if view == "Unknown screen" {
		t.Error("App shows unknown screen instead of setup")
	}
}

func TestApp_InitialState_WithToken_ShowsMain(t *testing.T) {
	tokenStore := auth.NewMemoryTokenStore()
	_ = tokenStore.SetJiraToken("existing-token")

	app := tui.NewApp(
		tui.WithTokenStore(tokenStore),
		tui.WithAuthPort(&mockAuthPort{}),
		tui.WithJiraPort(&mockJiraPort{}),
		tui.WithConfigPort(&mockConfigPort{
			team:     []domain.TeamMember{{Name: "Test", Email: "test@test.com"}},
			projects: []domain.Project{{Key: "TEST", Name: "Test"}},
		}),
	)

	cmd := app.Init()
	if cmd == nil {
		t.Fatal("Init() returned nil command")
	}

	msg := cmd()
	model, _ := app.Update(msg)
	view := model.View()

	// Should not show setup screen
	if view == "Loading setup..." {
		t.Error("App should not show setup screen when token exists")
	}
}

func TestApp_TokenStoredMsg_TransitionsToMain(t *testing.T) {
	tokenStore := auth.NewMemoryTokenStore()

	app := tui.NewApp(
		tui.WithTokenStore(tokenStore),
		tui.WithAuthPort(&mockAuthPort{}),
		tui.WithJiraPort(&mockJiraPort{}),
		tui.WithConfigPort(&mockConfigPort{
			team:     []domain.TeamMember{{Name: "Test", Email: "test@test.com"}},
			projects: []domain.Project{{Key: "TEST", Name: "Test"}},
		}),
	)

	// First init to setup screen
	cmd := app.Init()
	msg := cmd()
	app.Update(msg)

	// Simulate token stored message
	model, _ := app.Update(tui.TokenStoredMsg{})
	view := model.View()

	// Should not still be on setup
	if view == "Loading setup..." {
		t.Error("App should transition away from setup after TokenStoredMsg")
	}
}

func TestApp_LogoutMsg_TransitionsToSetup(t *testing.T) {
	tokenStore := auth.NewMemoryTokenStore()
	_ = tokenStore.SetJiraToken("existing-token")

	app := tui.NewApp(
		tui.WithTokenStore(tokenStore),
		tui.WithAuthPort(&mockAuthPort{}),
		tui.WithJiraPort(&mockJiraPort{}),
		tui.WithConfigPort(&mockConfigPort{
			team:     []domain.TeamMember{{Name: "Test", Email: "test@test.com"}},
			projects: []domain.Project{{Key: "TEST", Name: "Test"}},
		}),
	)

	// Init to main screen
	cmd := app.Init()
	msg := cmd()
	app.Update(msg)

	// Send logout
	model, _ := app.Update(tui.LogoutMsg{})
	view := model.View()

	// Should show setup or loading setup
	if view == "Unknown screen" {
		t.Error("App should show setup screen after logout")
	}
}

func TestApp_GlobalQuit(t *testing.T) {
	app := tui.NewApp(
		tui.WithTokenStore(auth.NewMemoryTokenStore()),
	)

	// Send ctrl+c
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("ctrl+c should return quit command")
	}
}

func TestApp_ErrorMsg_DisplaysError(t *testing.T) {
	app := tui.NewApp(
		tui.WithTokenStore(auth.NewMemoryTokenStore()),
	)

	// Send error message
	model, _ := app.Update(tui.ErrorMsg{Err: fmt.Errorf("test error")})
	view := model.View()

	if !contains(view, "test error") {
		t.Errorf("Error view should contain error message, got: %s", view)
	}
}

func TestApp_WindowSizeMsg(t *testing.T) {
	app := tui.NewApp(
		tui.WithTokenStore(auth.NewMemoryTokenStore()),
	)

	// Should not crash
	_, cmd := app.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	if cmd != nil {
		t.Error("WindowSizeMsg should not return a command")
	}
}
