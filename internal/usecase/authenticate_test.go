package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/stainedhead/gojira-tmux/internal/domain"
	"github.com/stainedhead/gojira-tmux/internal/usecase"
)

// MockAuthPort is a mock implementation of domain.AuthPort.
type MockAuthPort struct {
	StartAuthFlowFunc  func(ctx context.Context) (string, error)
	WaitForCallbackFunc func(ctx context.Context) (*domain.User, error)
	CancelAuthFlowFunc func()
	RefreshSessionFunc func(ctx context.Context) (*domain.User, error)
	IsSessionValidFunc func() bool
	LogoutFunc         func() error
}

func (m *MockAuthPort) StartAuthFlow(ctx context.Context) (string, error) {
	if m.StartAuthFlowFunc != nil {
		return m.StartAuthFlowFunc(ctx)
	}
	return "https://auth.example.com/authorize", nil
}

func (m *MockAuthPort) WaitForCallback(ctx context.Context) (*domain.User, error) {
	if m.WaitForCallbackFunc != nil {
		return m.WaitForCallbackFunc(ctx)
	}
	return &domain.User{
		Email:         "test@example.com",
		SessionExpiry: time.Now().Add(8 * time.Hour),
	}, nil
}

func (m *MockAuthPort) CancelAuthFlow() {
	if m.CancelAuthFlowFunc != nil {
		m.CancelAuthFlowFunc()
	}
}

func (m *MockAuthPort) RefreshSession(ctx context.Context) (*domain.User, error) {
	if m.RefreshSessionFunc != nil {
		return m.RefreshSessionFunc(ctx)
	}
	return nil, nil
}

func (m *MockAuthPort) IsSessionValid() bool {
	if m.IsSessionValidFunc != nil {
		return m.IsSessionValidFunc()
	}
	return false
}

func (m *MockAuthPort) Logout() error {
	if m.LogoutFunc != nil {
		return m.LogoutFunc()
	}
	return nil
}

// MockConfigPort is a mock implementation of domain.ConfigPort.
type MockConfigPort struct {
	LoadFunc            func() (*domain.Config, error)
	GetProjectsFunc     func() []domain.Project
	GetTeamMembersFunc  func() []domain.TeamMember
	ValidateUserAccessFunc func(email string) error
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

func (m *MockConfigPort) ValidateUserAccess(email string) error {
	if m.ValidateUserAccessFunc != nil {
		return m.ValidateUserAccessFunc(email)
	}
	return nil
}

func TestAuthenticate_StartLogin(t *testing.T) {
	authPort := &MockAuthPort{}
	configPort := &MockConfigPort{}

	uc := usecase.NewAuthenticate(authPort, configPort)

	ctx := context.Background()
	url, err := uc.StartLogin(ctx)

	if err != nil {
		t.Errorf("StartLogin() error = %v", err)
	}
	if url == "" {
		t.Error("StartLogin() returned empty URL")
	}
}

func TestAuthenticate_CompleteLogin_Success(t *testing.T) {
	authPort := &MockAuthPort{
		WaitForCallbackFunc: func(ctx context.Context) (*domain.User, error) {
			return &domain.User{
				Email:         "test@example.com",
				SessionExpiry: time.Now().Add(8 * time.Hour),
			}, nil
		},
	}
	configPort := &MockConfigPort{
		ValidateUserAccessFunc: func(email string) error {
			return nil // User is in team
		},
	}

	uc := usecase.NewAuthenticate(authPort, configPort)

	ctx := context.Background()
	user, err := uc.CompleteLogin(ctx)

	if err != nil {
		t.Errorf("CompleteLogin() error = %v", err)
	}
	if user == nil {
		t.Error("CompleteLogin() returned nil user")
	}
	if user != nil && user.Email != "test@example.com" {
		t.Errorf("CompleteLogin() user.Email = %q, want %q", user.Email, "test@example.com")
	}
}

func TestAuthenticate_CompleteLogin_UserNotInTeam(t *testing.T) {
	authPort := &MockAuthPort{
		WaitForCallbackFunc: func(ctx context.Context) (*domain.User, error) {
			return &domain.User{
				Email:         "stranger@example.com",
				SessionExpiry: time.Now().Add(8 * time.Hour),
			}, nil
		},
	}
	configPort := &MockConfigPort{
		ValidateUserAccessFunc: func(email string) error {
			user := domain.User{Email: email}
			return user.ValidateTeamMembership([]domain.TeamMember{
				{Name: "Test User", Email: "test@example.com"},
			})
		},
	}

	uc := usecase.NewAuthenticate(authPort, configPort)

	ctx := context.Background()
	_, err := uc.CompleteLogin(ctx)

	if err == nil {
		t.Error("CompleteLogin() expected error for user not in team, got nil")
	}
}

func TestAuthenticate_CheckSession_Valid(t *testing.T) {
	authPort := &MockAuthPort{
		IsSessionValidFunc: func() bool {
			return true
		},
		RefreshSessionFunc: func(ctx context.Context) (*domain.User, error) {
			return &domain.User{
				Email:         "test@example.com",
				SessionExpiry: time.Now().Add(8 * time.Hour),
			}, nil
		},
	}
	configPort := &MockConfigPort{}

	uc := usecase.NewAuthenticate(authPort, configPort)

	ctx := context.Background()
	user, err := uc.CheckSession(ctx)

	if err != nil {
		t.Errorf("CheckSession() error = %v", err)
	}
	if user == nil {
		t.Error("CheckSession() returned nil for valid session")
	}
}

func TestAuthenticate_CheckSession_Expired(t *testing.T) {
	authPort := &MockAuthPort{
		IsSessionValidFunc: func() bool {
			return false
		},
	}
	configPort := &MockConfigPort{}

	uc := usecase.NewAuthenticate(authPort, configPort)

	ctx := context.Background()
	user, err := uc.CheckSession(ctx)

	if err != nil {
		t.Errorf("CheckSession() error = %v", err)
	}
	if user != nil {
		t.Error("CheckSession() returned user for expired session")
	}
}

func TestAuthenticate_CancelLogin(t *testing.T) {
	cancelled := false
	authPort := &MockAuthPort{
		CancelAuthFlowFunc: func() {
			cancelled = true
		},
	}
	configPort := &MockConfigPort{}

	uc := usecase.NewAuthenticate(authPort, configPort)
	uc.CancelLogin()

	if !cancelled {
		t.Error("CancelLogin() did not cancel auth flow")
	}
}

func TestAuthenticate_Logout(t *testing.T) {
	loggedOut := false
	authPort := &MockAuthPort{
		LogoutFunc: func() error {
			loggedOut = true
			return nil
		},
	}
	configPort := &MockConfigPort{}

	uc := usecase.NewAuthenticate(authPort, configPort)
	err := uc.Logout()

	if err != nil {
		t.Errorf("Logout() error = %v", err)
	}
	if !loggedOut {
		t.Error("Logout() did not call authPort.Logout()")
	}
}
