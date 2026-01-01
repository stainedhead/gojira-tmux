package domain

import (
	"context"
)

// IssueFilter represents the filter state for querying issues.
type IssueFilter struct {
	Project  string // Empty or "-All-" means all projects
	Assignee string // Empty or "-All-" means all assignees
	Status   string // "All", "Open", "Ready", "In Test", "Done"
}

// IsEmpty returns true if no filters are applied.
func (f *IssueFilter) IsEmpty() bool {
	return (f.Project == "" || f.Project == "-All-") &&
		(f.Assignee == "" || f.Assignee == "-All-") &&
		(f.Status == "" || f.Status == "All")
}

// JiraPort defines the interface for Jira API operations.
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
	Jira     JiraConfig     `yaml:"jira"`
	Okta     OktaConfig     `yaml:"okta"`
	Projects []Project      `yaml:"projects"`
	Team     []TeamMember   `yaml:"team"`
}

// JiraConfig holds Jira-specific configuration.
type JiraConfig struct {
	URL          string            `yaml:"url"`
	Username     string            `yaml:"username"`
	CustomFields CustomFieldConfig `yaml:"custom_fields,omitempty"`
}

// CustomFieldConfig maps custom field names to IDs.
type CustomFieldConfig struct {
	Sprint      string `yaml:"sprint,omitempty"`
	Epic        string `yaml:"epic,omitempty"`
	StoryPoints string `yaml:"story_points,omitempty"`
}

// OktaConfig holds Okta-specific configuration.
type OktaConfig struct {
	Issuer       string   `yaml:"issuer"`
	ClientID     string   `yaml:"client_id"`
	CallbackPort int      `yaml:"callback_port"`
	Scopes       []string `yaml:"scopes"`
}
