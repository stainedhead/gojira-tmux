// Package contracts defines the port interfaces for the gojira-tmux application.
// These interfaces are defined here for documentation and planning purposes.
// The actual implementation will be in internal/domain/ports.go.
//
// This file serves as the contract specification for Phase 1 of /speckit.plan.
package contracts

import (
	"context"
	"time"
)

// =============================================================================
// Domain Entities (simplified for contract definition)
// =============================================================================

// Issue represents a Jira ticket.
type Issue struct {
	Key         string
	Summary     string
	Description string
	Status      string
	Priority    string
	Assignee    *TeamMember
	Reporter    *TeamMember
	DueDate     *time.Time
	Created     time.Time
	Updated     time.Time
	Sprint      string
	Epic        string
	Labels      []string
	StoryPoints int
	Comments    []Comment
}

// TeamMember represents a team member.
type TeamMember struct {
	Name  string
	Email string
}

// Project represents a Jira project.
type Project struct {
	Key  string
	Name string
}

// Comment represents an issue comment.
type Comment struct {
	ID      string
	Author  string
	Body    string
	Created time.Time
}

// User represents the authenticated user.
type User struct {
	Email         string
	SessionExpiry time.Time
}

// IssueFilter represents the filter state.
type IssueFilter struct {
	Project  string
	Assignee string
	Status   string
}

// =============================================================================
// Port Interfaces (Domain Layer)
// =============================================================================

// JiraPort defines the interface for Jira API operations.
// Implemented by: internal/adapter/jira/client.go
type JiraPort interface {
	// SearchIssues searches for issues matching the filter criteria.
	// Returns up to 100 issues sorted by last updated.
	SearchIssues(ctx context.Context, filter IssueFilter) ([]Issue, error)

	// GetIssue retrieves a single issue with all details including comments.
	GetIssue(ctx context.Context, key string) (*Issue, error)

	// GetIssueComments retrieves comments for an issue.
	GetIssueComments(ctx context.Context, key string) ([]Comment, error)
}

// AuthPort defines the interface for authentication operations.
// Implemented by: internal/adapter/auth/okta.go
type AuthPort interface {
	// StartAuthFlow initiates the Okta OIDC flow.
	// Returns the authorization URL to open in browser.
	StartAuthFlow(ctx context.Context) (authURL string, err error)

	// WaitForCallback waits for the OAuth callback.
	// Returns the user info after successful authentication.
	// ctx should have timeout set (e.g., 5 minutes).
	WaitForCallback(ctx context.Context) (*User, error)

	// CancelAuthFlow cancels an in-progress authentication.
	CancelAuthFlow()

	// RefreshSession refreshes the user session if refresh token exists.
	// Returns error if refresh fails or no refresh token.
	RefreshSession(ctx context.Context) (*User, error)

	// IsSessionValid checks if current session is still valid.
	IsSessionValid() bool

	// Logout clears the current session.
	Logout() error
}

// TokenStorePort defines the interface for secure credential storage.
// Implemented by: internal/adapter/auth/token_store.go
type TokenStorePort interface {
	// GetJiraToken retrieves the stored Jira API token.
	// Returns empty string if not found.
	GetJiraToken() (string, error)

	// SetJiraToken stores the Jira API token securely.
	SetJiraToken(token string) error

	// DeleteJiraToken removes the stored Jira API token.
	DeleteJiraToken() error

	// GetRefreshToken retrieves the stored Okta refresh token.
	GetRefreshToken() (string, error)

	// SetRefreshToken stores the Okta refresh token.
	SetRefreshToken(token string) error

	// DeleteRefreshToken removes the stored refresh token.
	DeleteRefreshToken() error

	// HasJiraToken returns true if a Jira token exists.
	HasJiraToken() bool
}

// ConfigPort defines the interface for configuration operations.
// Implemented by: internal/adapter/config/config.go
type ConfigPort interface {
	// Load reads and parses the configuration file.
	Load() (*Config, error)

	// GetProjects returns configured projects.
	GetProjects() []Project

	// GetTeamMembers returns configured team members.
	GetTeamMembers() []TeamMember

	// ValidateUserAccess checks if email is in team list.
	ValidateUserAccess(email string) error
}

// Config represents the application configuration.
type Config struct {
	Jira     JiraConfig
	Okta     OktaConfig
	Projects []Project
	Team     []TeamMember
}

// JiraConfig holds Jira-specific configuration.
type JiraConfig struct {
	URL          string
	Username     string
	CustomFields CustomFieldConfig
}

// CustomFieldConfig maps custom field names to IDs.
type CustomFieldConfig struct {
	Sprint      string
	Epic        string
	StoryPoints string
}

// OktaConfig holds Okta-specific configuration.
type OktaConfig struct {
	Issuer       string
	ClientID     string
	CallbackPort int
	Scopes       []string
}

// =============================================================================
// Use Case Interfaces (Application Layer)
// =============================================================================

// ListIssuesUseCase handles listing and filtering issues.
// Implemented by: internal/usecase/list_issues.go
type ListIssuesUseCase interface {
	// Execute lists issues based on filter criteria.
	Execute(ctx context.Context, filter IssueFilter) ([]Issue, error)
}

// GetIssueDetailsUseCase handles fetching issue details.
// Implemented by: internal/usecase/get_issue_details.go
type GetIssueDetailsUseCase interface {
	// Execute retrieves full issue details including comments.
	Execute(ctx context.Context, key string) (*Issue, error)
}

// AuthenticateUseCase handles the authentication flow.
// Implemented by: internal/usecase/authenticate.go
type AuthenticateUseCase interface {
	// StartLogin initiates the login flow.
	// Returns auth URL for browser.
	StartLogin(ctx context.Context) (string, error)

	// CompleteLogin waits for and processes the OAuth callback.
	CompleteLogin(ctx context.Context) (*User, error)

	// CancelLogin cancels an in-progress login.
	CancelLogin()

	// CheckSession checks if user has valid session.
	CheckSession(ctx context.Context) (*User, error)

	// Logout logs out the current user.
	Logout() error
}

// SetupTokenUseCase handles first-time API token setup.
// Implemented by: internal/usecase/setup_token.go
type SetupTokenUseCase interface {
	// NeedsSetup returns true if no API token is configured.
	NeedsSetup() bool

	// SaveToken validates and saves the API token.
	SaveToken(token string) error
}

// =============================================================================
// TUI Message Types
// =============================================================================

// Messages for screen transitions and state updates.
// These are used with BubbleTea's Update function.

// TokenStoredMsg indicates API token was successfully stored.
type TokenStoredMsg struct{}

// AuthStartedMsg indicates browser was opened for auth.
type AuthStartedMsg struct {
	URL string
}

// AuthSuccessMsg indicates successful authentication.
type AuthSuccessMsg struct {
	User *User
}

// AuthErrorMsg indicates authentication failed.
type AuthErrorMsg struct {
	Err error
}

// AuthCancelledMsg indicates user cancelled authentication.
type AuthCancelledMsg struct{}

// IssuesLoadedMsg indicates issues were loaded.
type IssuesLoadedMsg struct {
	Issues []Issue
}

// IssueSelectedMsg indicates an issue was selected.
type IssueSelectedMsg struct {
	Issue *Issue
}

// IssueDetailsLoadedMsg indicates issue details were loaded.
type IssueDetailsLoadedMsg struct {
	Issue *Issue
}

// ErrorMsg indicates an error occurred.
type ErrorMsg struct {
	Err error
}

// RefreshMsg triggers a data refresh.
type RefreshMsg struct{}

// FilterChangedMsg indicates filter state changed.
type FilterChangedMsg struct {
	Filter IssueFilter
}
