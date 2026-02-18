package domain

import (
	"context"
)

// IssueFilter represents the filter state for querying issues.
type IssueFilter struct {
	Project  string // Empty or "-All-" means all projects
	Assignee string // Empty or "-All-" means all assignees
	Status   string // "All" means no filter; otherwise the exact Jira status name
}

// FilterState holds persisted filter selections for config storage.
type FilterState struct {
	Assignee string `yaml:"assignee,omitempty"`
	Project  string `yaml:"project,omitempty"`
	Status   string `yaml:"status,omitempty"`
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

	// ListStatuses returns all available issue status names from the Jira instance.
	ListStatuses(ctx context.Context) ([]string, error)
}

// AuthPort defines the interface for authentication operations.
type AuthPort interface {
	// ValidateToken validates an Atlassian API token by calling the Jira API.
	// Returns the validated email address on success.
	ValidateToken(ctx context.Context, email, token string) (string, error)

	// IsTokenValid checks if a valid token exists in the store.
	IsTokenValid(ctx context.Context) bool
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

	// HasJiraToken returns true if a Jira token exists.
	HasJiraToken() bool
}

// ConfigPort defines the interface for configuration operations.
type ConfigPort interface {
	// Load reads and parses the configuration file.
	Load() (*Config, error)

	// Save writes the configuration back to the file.
	Save(config *Config) error

	// GetProjects returns configured projects.
	GetProjects() []Project

	// GetTeamMembers returns configured team members.
	GetTeamMembers() []TeamMember
}

// Config represents the application configuration.
type Config struct {
	Atlassian  AtlassianConfig `yaml:"atlassian"`
	Projects   []Project       `yaml:"projects"`
	Team       []TeamMember    `yaml:"team"`
	Statuses   []string        `yaml:"statuses,omitempty"`
	LastFilter FilterState     `yaml:"last_filter,omitempty"`
}

// AtlassianConfig holds Atlassian-specific configuration.
type AtlassianConfig struct {
	URL          string            `yaml:"url"`
	Email        string            `yaml:"email"`
	CustomFields CustomFieldConfig `yaml:"custom_fields,omitempty"`
}

// CustomFieldConfig maps custom field names to IDs.
type CustomFieldConfig struct {
	Sprint      string `yaml:"sprint,omitempty"`
	Epic        string `yaml:"epic,omitempty"`
	StoryPoints string `yaml:"story_points,omitempty"`
}
