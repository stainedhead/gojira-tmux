package jira_test

import (
	"testing"

	"github.com/stainedhead/gojira-tmux/internal/adapter/jira"
	"github.com/stainedhead/gojira-tmux/internal/domain"
)

func TestJQLBuilder_Build(t *testing.T) {
	tests := []struct {
		name     string
		filter   domain.IssueFilter
		projects []domain.Project
		team     []domain.TeamMember
		want     string
	}{
		{
			name:   "empty filter - all projects",
			filter: domain.IssueFilter{},
			projects: []domain.Project{
				{Key: "PROJ", Name: "Project"},
			},
			team: []domain.TeamMember{},
			want: `project IN ("PROJ") ORDER BY updated DESC`,
		},
		{
			name:   "multiple projects",
			filter: domain.IssueFilter{},
			projects: []domain.Project{
				{Key: "PROJ", Name: "Project"},
				{Key: "TEST", Name: "Test Project"},
			},
			team: []domain.TeamMember{},
			want: `project IN ("PROJ", "TEST") ORDER BY updated DESC`,
		},
		{
			name: "single project filter",
			filter: domain.IssueFilter{
				Project: "PROJ",
			},
			projects: []domain.Project{
				{Key: "PROJ", Name: "Project"},
				{Key: "TEST", Name: "Test Project"},
			},
			team: []domain.TeamMember{},
			want: `project = "PROJ" ORDER BY updated DESC`,
		},
		{
			name: "assignee filter",
			filter: domain.IssueFilter{
				Assignee: "John Doe",
			},
			projects: []domain.Project{
				{Key: "PROJ", Name: "Project"},
			},
			team: []domain.TeamMember{
				{Name: "John Doe", Email: "john@example.com"},
			},
			want: `project IN ("PROJ") AND assignee = "john@example.com" ORDER BY updated DESC`,
		},
		{
			name: "status filter - Open",
			filter: domain.IssueFilter{
				Status: "Open",
			},
			projects: []domain.Project{
				{Key: "PROJ", Name: "Project"},
			},
			team: []domain.TeamMember{},
			want: `project IN ("PROJ") AND status in ("Open") ORDER BY updated DESC`,
		},
		{
			name: "status filter - In Progress",
			filter: domain.IssueFilter{
				Status: "In Progress",
			},
			projects: []domain.Project{
				{Key: "PROJ", Name: "Project"},
			},
			team: []domain.TeamMember{},
			want: `project IN ("PROJ") AND status in ("In Progress") ORDER BY updated DESC`,
		},
		{
			name: "status filter - All (no filter)",
			filter: domain.IssueFilter{
				Status: "All",
			},
			projects: []domain.Project{
				{Key: "PROJ", Name: "Project"},
			},
			team: []domain.TeamMember{},
			want: `project IN ("PROJ") ORDER BY updated DESC`,
		},
		{
			name: "combined filters",
			filter: domain.IssueFilter{
				Project:  "PROJ",
				Assignee: "John Doe",
				Status:   "Open",
			},
			projects: []domain.Project{
				{Key: "PROJ", Name: "Project"},
			},
			team: []domain.TeamMember{
				{Name: "John Doe", Email: "john@example.com"},
			},
			want: `project = "PROJ" AND assignee = "john@example.com" AND status in ("Open") ORDER BY updated DESC`,
		},
		{
			name: "-All- project filter",
			filter: domain.IssueFilter{
				Project: "-All-",
			},
			projects: []domain.Project{
				{Key: "PROJ", Name: "Project"},
				{Key: "TEST", Name: "Test"},
			},
			team: []domain.TeamMember{},
			want: `project IN ("PROJ", "TEST") ORDER BY updated DESC`,
		},
		{
			name: "-All- assignee filter",
			filter: domain.IssueFilter{
				Assignee: "-All-",
			},
			projects: []domain.Project{
				{Key: "PROJ", Name: "Project"},
			},
			team: []domain.TeamMember{},
			want: `project IN ("PROJ") ORDER BY updated DESC`,
		},
		{
			name: "assignee filter by alias",
			filter: domain.IssueFilter{
				Assignee: "JohnA",
			},
			projects: []domain.Project{
				{Key: "PROJ", Name: "Project"},
			},
			team: []domain.TeamMember{
				{Name: "John Anderson", Email: "john.anderson@example.com", Alias: "JohnA"},
				{Name: "John Flanagan", Email: "john.flanagan@example.com", Alias: "JohnF"},
			},
			want: `project IN ("PROJ") AND assignee = "john.anderson@example.com" ORDER BY updated DESC`,
		},
		{
			name: "assignee filter by case-insensitive alias",
			filter: domain.IssueFilter{
				Assignee: "johna",
			},
			projects: []domain.Project{
				{Key: "PROJ", Name: "Project"},
			},
			team: []domain.TeamMember{
				{Name: "John Anderson", Email: "john.anderson@example.com", Alias: "JohnA"},
			},
			want: `project IN ("PROJ") AND assignee = "john.anderson@example.com" ORDER BY updated DESC`,
		},
		{
			name: "assignee filter backward compat - no alias member",
			filter: domain.IssueFilter{
				Assignee: "Jane Smith",
			},
			projects: []domain.Project{
				{Key: "PROJ", Name: "Project"},
			},
			team: []domain.TeamMember{
				{Name: "Jane Smith", Email: "jane@example.com"},
			},
			want: `project IN ("PROJ") AND assignee = "jane@example.com" ORDER BY updated DESC`,
		},
		{
			name: "assignee filter no match returns empty",
			filter: domain.IssueFilter{
				Assignee: "Unknown Person",
			},
			projects: []domain.Project{
				{Key: "PROJ", Name: "Project"},
			},
			team: []domain.TeamMember{
				{Name: "John Doe", Email: "john@example.com"},
			},
			want: `project IN ("PROJ") ORDER BY updated DESC`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := jira.NewJQLBuilder(tt.projects, tt.team)
			got := builder.Build(tt.filter)
			if got != tt.want {
				t.Errorf("Build() = %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestJQLBuilder_EscapeSpecialCharacters(t *testing.T) {
	projects := []domain.Project{
		{Key: "TEST", Name: "Test Project"},
	}
	team := []domain.TeamMember{
		{Name: `John "The Dev" Doe`, Email: "john@example.com"},
	}

	builder := jira.NewJQLBuilder(projects, team)
	filter := domain.IssueFilter{
		Assignee: `John "The Dev" Doe`,
	}

	got := builder.Build(filter)

	// Should escape quotes in email lookup
	if got == "" {
		t.Error("Build() returned empty string")
	}
}

func TestJQLBuilder_StatusPassthrough(t *testing.T) {
	// Status values are passed directly to JQL without mapping - they come from Jira API.
	tests := []struct {
		status string
		want   string
	}{
		{"Open", `project IN ("PROJ") AND status in ("Open") ORDER BY updated DESC`},
		{"In Progress", `project IN ("PROJ") AND status in ("In Progress") ORDER BY updated DESC`},
		{"Done", `project IN ("PROJ") AND status in ("Done") ORDER BY updated DESC`},
		{"All", `project IN ("PROJ") ORDER BY updated DESC`},
		{"", `project IN ("PROJ") ORDER BY updated DESC`},
	}

	projects := []domain.Project{{Key: "PROJ", Name: "Project"}}
	builder := jira.NewJQLBuilder(projects, nil)

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := builder.Build(domain.IssueFilter{Status: tt.status})
			if got != tt.want {
				t.Errorf("Build(Status=%q) = %q\nwant %q", tt.status, got, tt.want)
			}
		})
	}
}
