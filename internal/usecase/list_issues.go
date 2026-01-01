package usecase

import (
	"context"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// ListIssues handles listing and filtering issues.
type ListIssues struct {
	jiraPort domain.JiraPort
}

// NewListIssues creates a new ListIssues use case.
func NewListIssues(jiraPort domain.JiraPort) *ListIssues {
	return &ListIssues{
		jiraPort: jiraPort,
	}
}

// Execute lists issues based on filter criteria.
func (l *ListIssues) Execute(ctx context.Context, filter domain.IssueFilter) ([]domain.Issue, error) {
	return l.jiraPort.SearchIssues(ctx, filter)
}
