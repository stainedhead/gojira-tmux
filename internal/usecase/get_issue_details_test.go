package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stainedhead/gojira-tmux/internal/domain"
	"github.com/stainedhead/gojira-tmux/internal/usecase"
)

func TestGetIssueDetails_Execute_Success(t *testing.T) {
	now := time.Now()
	expectedIssue := &domain.Issue{
		Key:         "PROJ-123",
		Summary:     "Test issue",
		Description: "Detailed description",
		Status:      "Open",
		Priority:    "High",
		Created:     now,
		Updated:     now,
		Comments: []domain.Comment{
			{
				ID:      "1",
				Body:    "First comment",
				Author:  "John Doe",
				Created: now.Add(-1 * time.Hour),
			},
			{
				ID:      "2",
				Body:    "Second comment",
				Author:  "Jane Doe",
				Created: now,
			},
		},
	}

	jiraPort := &MockJiraPort{
		GetIssueFunc: func(ctx context.Context, key string) (*domain.Issue, error) {
			if key != "PROJ-123" {
				t.Errorf("GetIssue called with key = %q, want %q", key, "PROJ-123")
			}
			return expectedIssue, nil
		},
	}

	uc := usecase.NewGetIssueDetails(jiraPort)

	ctx := context.Background()
	issue, err := uc.Execute(ctx, "PROJ-123")

	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if issue == nil {
		t.Fatal("Execute() returned nil issue")
	}

	if issue.Key != expectedIssue.Key {
		t.Errorf("issue.Key = %q, want %q", issue.Key, expectedIssue.Key)
	}

	if issue.Description != expectedIssue.Description {
		t.Errorf("issue.Description = %q, want %q", issue.Description, expectedIssue.Description)
	}

	if len(issue.Comments) != 2 {
		t.Errorf("len(issue.Comments) = %d, want 2", len(issue.Comments))
	}
}

func TestGetIssueDetails_Execute_WithCommentsFetch(t *testing.T) {
	now := time.Now()

	// Issue without comments (comments fetched separately)
	issueWithoutComments := &domain.Issue{
		Key:         "PROJ-456",
		Summary:     "Another issue",
		Status:      "Open",
		Created:     now,
		Updated:     now,
		Comments:    nil,
	}

	comments := []domain.Comment{
		{ID: "1", Body: "Comment 1", Author: "Alice", Created: now},
		{ID: "2", Body: "Comment 2", Author: "Bob", Created: now},
	}

	jiraPort := &MockJiraPort{
		GetIssueFunc: func(ctx context.Context, key string) (*domain.Issue, error) {
			return issueWithoutComments, nil
		},
		GetIssueCommentsFunc: func(ctx context.Context, key string) ([]domain.Comment, error) {
			return comments, nil
		},
	}

	uc := usecase.NewGetIssueDetails(jiraPort)

	ctx := context.Background()
	issue, err := uc.Execute(ctx, "PROJ-456")

	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	if issue == nil {
		t.Fatal("Execute() returned nil issue")
	}

	// Comments should be fetched and attached
	if len(issue.Comments) != 2 {
		t.Errorf("len(issue.Comments) = %d, want 2", len(issue.Comments))
	}
}

func TestGetIssueDetails_Execute_IssueNotFound(t *testing.T) {
	jiraPort := &MockJiraPort{
		GetIssueFunc: func(ctx context.Context, key string) (*domain.Issue, error) {
			return nil, errors.New("issue not found")
		},
	}

	uc := usecase.NewGetIssueDetails(jiraPort)

	ctx := context.Background()
	issue, err := uc.Execute(ctx, "INVALID-999")

	if err == nil {
		t.Error("Execute() expected error, got nil")
	}

	if issue != nil {
		t.Errorf("Execute() returned issue, want nil")
	}
}

func TestGetIssueDetails_Execute_EmptyKey(t *testing.T) {
	jiraPort := &MockJiraPort{}

	uc := usecase.NewGetIssueDetails(jiraPort)

	ctx := context.Background()
	issue, err := uc.Execute(ctx, "")

	if err == nil {
		t.Error("Execute() expected error for empty key, got nil")
	}

	if issue != nil {
		t.Errorf("Execute() returned issue, want nil")
	}
}

func TestGetIssueDetails_Execute_ContextCancellation(t *testing.T) {
	jiraPort := &MockJiraPort{
		GetIssueFunc: func(ctx context.Context, key string) (*domain.Issue, error) {
			return nil, ctx.Err()
		},
	}

	uc := usecase.NewGetIssueDetails(jiraPort)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := uc.Execute(ctx, "PROJ-123")

	if err == nil {
		t.Error("Execute() expected error for cancelled context, got nil")
	}
}

func TestGetIssueDetails_Execute_CommentsFetchError(t *testing.T) {
	now := time.Now()
	issueWithoutComments := &domain.Issue{
		Key:      "PROJ-789",
		Summary:  "Issue",
		Status:   "Open",
		Created:  now,
		Updated:  now,
		Comments: nil,
	}

	jiraPort := &MockJiraPort{
		GetIssueFunc: func(ctx context.Context, key string) (*domain.Issue, error) {
			return issueWithoutComments, nil
		},
		GetIssueCommentsFunc: func(ctx context.Context, key string) ([]domain.Comment, error) {
			return nil, errors.New("comments fetch failed")
		},
	}

	uc := usecase.NewGetIssueDetails(jiraPort)

	ctx := context.Background()
	issue, err := uc.Execute(ctx, "PROJ-789")

	// Should still return issue even if comments fail
	if err != nil {
		t.Errorf("Execute() error = %v, should succeed even if comments fail", err)
	}

	if issue == nil {
		t.Fatal("Execute() returned nil issue")
	}

	// Comments should be empty
	if len(issue.Comments) != 0 {
		t.Errorf("len(issue.Comments) = %d, want 0", len(issue.Comments))
	}
}
