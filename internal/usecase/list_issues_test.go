package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stainedhead/gojira-tmux/internal/domain"
	"github.com/stainedhead/gojira-tmux/internal/usecase"
)

// MockJiraPort is a mock implementation of domain.JiraPort.
type MockJiraPort struct {
	SearchIssuesFunc     func(ctx context.Context, filter domain.IssueFilter) ([]domain.Issue, error)
	GetIssueFunc         func(ctx context.Context, key string) (*domain.Issue, error)
	GetIssueCommentsFunc func(ctx context.Context, key string) ([]domain.Comment, error)
	ListStatusesFunc     func(ctx context.Context) ([]string, error)
}

func (m *MockJiraPort) SearchIssues(ctx context.Context, filter domain.IssueFilter) ([]domain.Issue, error) {
	if m.SearchIssuesFunc != nil {
		return m.SearchIssuesFunc(ctx, filter)
	}
	return nil, nil
}

func (m *MockJiraPort) GetIssue(ctx context.Context, key string) (*domain.Issue, error) {
	if m.GetIssueFunc != nil {
		return m.GetIssueFunc(ctx, key)
	}
	return nil, nil
}

func (m *MockJiraPort) GetIssueComments(ctx context.Context, key string) ([]domain.Comment, error) {
	if m.GetIssueCommentsFunc != nil {
		return m.GetIssueCommentsFunc(ctx, key)
	}
	return nil, nil
}

func (m *MockJiraPort) ListStatuses(ctx context.Context) ([]string, error) {
	if m.ListStatusesFunc != nil {
		return m.ListStatusesFunc(ctx)
	}
	return []string{"To Do", "In Progress", "Done"}, nil
}

func TestListIssues_Execute_Success(t *testing.T) {
	now := time.Now()
	mockIssues := []domain.Issue{
		{
			Key:     "PROJ-1",
			Summary: "First issue",
			Status:  "Open",
			Created: now,
			Updated: now,
		},
		{
			Key:     "PROJ-2",
			Summary: "Second issue",
			Status:  "Done",
			Created: now,
			Updated: now,
		},
	}

	jiraPort := &MockJiraPort{
		SearchIssuesFunc: func(ctx context.Context, filter domain.IssueFilter) ([]domain.Issue, error) {
			return mockIssues, nil
		},
	}

	uc := usecase.NewListIssues(jiraPort)

	ctx := context.Background()
	filter := domain.IssueFilter{}
	issues, err := uc.Execute(ctx, filter)

	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if len(issues) != 2 {
		t.Errorf("Execute() returned %d issues, want 2", len(issues))
	}
}

func TestListIssues_Execute_WithFilter(t *testing.T) {
	var capturedFilter domain.IssueFilter

	jiraPort := &MockJiraPort{
		SearchIssuesFunc: func(ctx context.Context, filter domain.IssueFilter) ([]domain.Issue, error) {
			capturedFilter = filter
			return []domain.Issue{}, nil
		},
	}

	uc := usecase.NewListIssues(jiraPort)

	ctx := context.Background()
	filter := domain.IssueFilter{
		Project:  "PROJ",
		Assignee: "John Doe",
		Status:   "Open",
	}
	_, err := uc.Execute(ctx, filter)

	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if capturedFilter.Project != "PROJ" {
		t.Errorf("filter.Project = %q, want %q", capturedFilter.Project, "PROJ")
	}

	if capturedFilter.Assignee != "John Doe" {
		t.Errorf("filter.Assignee = %q, want %q", capturedFilter.Assignee, "John Doe")
	}

	if capturedFilter.Status != "Open" {
		t.Errorf("filter.Status = %q, want %q", capturedFilter.Status, "Open")
	}
}

func TestListIssues_Execute_Error(t *testing.T) {
	jiraPort := &MockJiraPort{
		SearchIssuesFunc: func(ctx context.Context, filter domain.IssueFilter) ([]domain.Issue, error) {
			return nil, errors.New("jira API error")
		},
	}

	uc := usecase.NewListIssues(jiraPort)

	ctx := context.Background()
	_, err := uc.Execute(ctx, domain.IssueFilter{})

	if err == nil {
		t.Error("Execute() expected error, got nil")
	}
}

func TestListIssues_Execute_EmptyResult(t *testing.T) {
	jiraPort := &MockJiraPort{
		SearchIssuesFunc: func(ctx context.Context, filter domain.IssueFilter) ([]domain.Issue, error) {
			return []domain.Issue{}, nil
		},
	}

	uc := usecase.NewListIssues(jiraPort)

	ctx := context.Background()
	issues, err := uc.Execute(ctx, domain.IssueFilter{})

	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("Execute() returned %d issues, want 0", len(issues))
	}
}

func TestListIssues_Execute_ContextCancellation(t *testing.T) {
	jiraPort := &MockJiraPort{
		SearchIssuesFunc: func(ctx context.Context, filter domain.IssueFilter) ([]domain.Issue, error) {
			return nil, ctx.Err()
		},
	}

	uc := usecase.NewListIssues(jiraPort)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := uc.Execute(ctx, domain.IssueFilter{})

	if err == nil {
		t.Error("Execute() expected error for cancelled context, got nil")
	}
}
