package usecase

import (
	"context"
	"errors"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// ErrEmptyIssueKey indicates an empty issue key was provided.
var ErrEmptyIssueKey = errors.New("issue key is required")

// GetIssueDetails handles fetching full issue details including comments.
type GetIssueDetails struct {
	jiraPort domain.JiraPort
}

// NewGetIssueDetails creates a new GetIssueDetails use case.
func NewGetIssueDetails(jiraPort domain.JiraPort) *GetIssueDetails {
	return &GetIssueDetails{
		jiraPort: jiraPort,
	}
}

// Execute fetches the full details for an issue including comments.
func (g *GetIssueDetails) Execute(ctx context.Context, key string) (*domain.Issue, error) {
	if key == "" {
		return nil, ErrEmptyIssueKey
	}

	// Fetch the issue
	issue, err := g.jiraPort.GetIssue(ctx, key)
	if err != nil {
		return nil, err
	}

	// If issue doesn't have comments, fetch them separately
	if len(issue.Comments) == 0 {
		comments, err := g.jiraPort.GetIssueComments(ctx, key)
		if err == nil {
			issue.Comments = comments
		}
		// Ignore comments fetch error - return issue without comments
	}

	return issue, nil
}
