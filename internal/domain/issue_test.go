package domain_test

import (
	"testing"
	"time"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

func TestIssue_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		issue   domain.Issue
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid issue",
			issue: domain.Issue{
				Key:     "PROJ-123",
				Summary: "Test issue",
				Status:  "Open",
				Created: now,
				Updated: now,
			},
			wantErr: false,
		},
		{
			name: "empty key",
			issue: domain.Issue{
				Key:     "",
				Summary: "Test issue",
				Status:  "Open",
				Created: now,
				Updated: now,
			},
			wantErr: true,
			errMsg:  "issue key is required",
		},
		{
			name: "invalid key format - no hyphen",
			issue: domain.Issue{
				Key:     "PROJ123",
				Summary: "Test issue",
				Status:  "Open",
				Created: now,
				Updated: now,
			},
			wantErr: true,
			errMsg:  "issue key must match pattern PROJECT-123",
		},
		{
			name: "invalid key format - lowercase",
			issue: domain.Issue{
				Key:     "proj-123",
				Summary: "Test issue",
				Status:  "Open",
				Created: now,
				Updated: now,
			},
			wantErr: true,
			errMsg:  "issue key must match pattern PROJECT-123",
		},
		{
			name: "empty summary",
			issue: domain.Issue{
				Key:     "PROJ-123",
				Summary: "",
				Status:  "Open",
				Created: now,
				Updated: now,
			},
			wantErr: true,
			errMsg:  "issue summary is required",
		},
		{
			name: "empty status",
			issue: domain.Issue{
				Key:     "PROJ-123",
				Summary: "Test issue",
				Status:  "",
				Created: now,
				Updated: now,
			},
			wantErr: true,
			errMsg:  "issue status is required",
		},
		{
			name: "zero created time",
			issue: domain.Issue{
				Key:     "PROJ-123",
				Summary: "Test issue",
				Status:  "Open",
				Created: time.Time{},
				Updated: now,
			},
			wantErr: true,
			errMsg:  "issue created time is required",
		},
		{
			name: "zero updated time",
			issue: domain.Issue{
				Key:     "PROJ-123",
				Summary: "Test issue",
				Status:  "Open",
				Created: now,
				Updated: time.Time{},
			},
			wantErr: true,
			errMsg:  "issue updated time is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.issue.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Validate() error = %q, want %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestIssue_NeedsAttention(t *testing.T) {
	now := time.Now()
	dueDate := now.Add(7 * 24 * time.Hour)

	tests := []struct {
		name     string
		issue    domain.Issue
		expected domain.AttentionType
	}{
		{
			name: "no attention for non-open issue (Done)",
			issue: domain.Issue{
				Key:     "PROJ-1",
				Summary: "Test",
				Status:  "Done",
				Created: now,
				Updated: now,
			},
			expected: domain.AttentionNone,
		},
		{
			name: "yellow dot for active non-done status (In Test) without due date",
			issue: domain.Issue{
				Key:     "PROJ-1",
				Summary: "Test",
				Status:  "In Test",
				Created: now,
				Updated: now,
			},
			expected: domain.AttentionNoDueDate,
		},
		{
			name: "no attention for active status (To Do) with due date and recent update",
			issue: domain.Issue{
				Key:     "PROJ-1",
				Summary: "Test",
				Status:  "To Do",
				DueDate: &dueDate,
				Created: now,
				Updated: now,
			},
			expected: domain.AttentionNone,
		},
		{
			name: "yellow dot for In Progress without due date",
			issue: domain.Issue{
				Key:     "PROJ-1",
				Summary: "Test",
				Status:  "In Progress",
				Created: now,
				Updated: now,
			},
			expected: domain.AttentionNoDueDate,
		},
		{
			name: "yellow dot for open issue without due date",
			issue: domain.Issue{
				Key:     "PROJ-1",
				Summary: "Test",
				Status:  "Open",
				DueDate: nil,
				Created: now,
				Updated: now,
			},
			expected: domain.AttentionNoDueDate,
		},
		{
			name: "no attention for open issue with due date and recent activity",
			issue: domain.Issue{
				Key:     "PROJ-1",
				Summary: "Test",
				Status:  "Open",
				DueDate: &dueDate,
				Created: now,
				Updated: now,
			},
			expected: domain.AttentionNone,
		},
		{
			name: "red dot for stale open issue (assignee, no comments, 15+ days since last update)",
			issue: domain.Issue{
				Key:      "PROJ-1",
				Summary:  "Test",
				Status:   "Open",
				DueDate:  &dueDate,
				Assignee: &domain.TeamMember{Name: "John", Email: "john@test.com"},
				Created:  now.Add(-15 * 24 * time.Hour),
				Updated:  now.Add(-15 * 24 * time.Hour),
				Comments: []domain.Comment{},
			},
			expected: domain.AttentionStale,
		},
		{
			name: "red dot takes precedence over yellow (stale AND no due date)",
			issue: domain.Issue{
				Key:      "PROJ-1",
				Summary:  "Test",
				Status:   "Open",
				DueDate:  nil,
				Assignee: &domain.TeamMember{Name: "John", Email: "john@test.com"},
				Created:  now.Add(-15 * 24 * time.Hour),
				Updated:  now.Add(-15 * 24 * time.Hour),
				Comments: []domain.Comment{},
			},
			expected: domain.AttentionStale,
		},
		{
			name: "no red dot when Updated is recent even if Created is old (no assignee comment)",
			issue: domain.Issue{
				Key:      "PROJ-1",
				Summary:  "Test",
				Status:   "Open",
				DueDate:  &dueDate,
				Assignee: &domain.TeamMember{Name: "John", Email: "john@test.com"},
				Created:  now.Add(-30 * 24 * time.Hour),
				Updated:  now.Add(-2 * 24 * time.Hour),
				Comments: []domain.Comment{},
			},
			expected: domain.AttentionNone,
		},
		{
			name: "no attention for open issue with recent assignee comment",
			issue: domain.Issue{
				Key:      "PROJ-1",
				Summary:  "Test",
				Status:   "Open",
				DueDate:  &dueDate,
				Assignee: &domain.TeamMember{Name: "John Doe", Email: "john@test.com"},
				Created:  now.Add(-30 * 24 * time.Hour),
				Updated:  now,
				Comments: []domain.Comment{
					{
						ID:      "1",
						Author:  "John Doe",
						Body:    "Working on it",
						Created: now.Add(-5 * 24 * time.Hour),
					},
				},
			},
			expected: domain.AttentionNone,
		},
		{
			name: "red dot for stale assignee comment (14+ days ago)",
			issue: domain.Issue{
				Key:      "PROJ-1",
				Summary:  "Test",
				Status:   "Open",
				DueDate:  &dueDate,
				Assignee: &domain.TeamMember{Name: "John Doe", Email: "john@test.com"},
				Created:  now.Add(-30 * 24 * time.Hour),
				Updated:  now,
				Comments: []domain.Comment{
					{
						ID:      "1",
						Author:  "John Doe",
						Body:    "Old comment",
						Created: now.Add(-15 * 24 * time.Hour),
					},
				},
			},
			expected: domain.AttentionStale,
		},
		{
			name: "no stale check for unassigned issue",
			issue: domain.Issue{
				Key:      "PROJ-1",
				Summary:  "Test",
				Status:   "Open",
				DueDate:  &dueDate,
				Assignee: nil,
				Created:  now.Add(-30 * 24 * time.Hour),
				Updated:  now,
			},
			expected: domain.AttentionNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.issue.NeedsAttention()
			if got != tt.expected {
				t.Errorf("NeedsAttention() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIssue_IsStale(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		issue domain.Issue
		want  bool
	}{
		{
			name: "not stale - no assignee",
			issue: domain.Issue{
				Assignee: nil,
				Created:  now.Add(-30 * 24 * time.Hour),
			},
			want: false,
		},
		{
			name: "not stale - recent issue with assignee",
			issue: domain.Issue{
				Assignee: &domain.TeamMember{Name: "John", Email: "john@test.com"},
				Created:  now.Add(-5 * 24 * time.Hour),
			},
			want: false,
		},
		{
			name: "stale - old issue with assignee, no comments",
			issue: domain.Issue{
				Assignee: &domain.TeamMember{Name: "John", Email: "john@test.com"},
				Created:  now.Add(-15 * 24 * time.Hour),
				Comments: []domain.Comment{},
			},
			want: true,
		},
		{
			name: "not stale - old issue with recent assignee comment",
			issue: domain.Issue{
				Assignee: &domain.TeamMember{Name: "John Doe", Email: "john@test.com"},
				Created:  now.Add(-30 * 24 * time.Hour),
				Comments: []domain.Comment{
					{
						ID:      "1",
						Author:  "John Doe",
						Body:    "Recent update",
						Created: now.Add(-5 * 24 * time.Hour),
					},
				},
			},
			want: false,
		},
		{
			name: "stale - old issue with old assignee comment",
			issue: domain.Issue{
				Assignee: &domain.TeamMember{Name: "John Doe", Email: "john@test.com"},
				Created:  now.Add(-30 * 24 * time.Hour),
				Comments: []domain.Comment{
					{
						ID:      "1",
						Author:  "John Doe",
						Body:    "Old update",
						Created: now.Add(-20 * 24 * time.Hour),
					},
				},
			},
			want: true,
		},
		{
			name: "stale - comments exist but none from assignee",
			issue: domain.Issue{
				Assignee: &domain.TeamMember{Name: "John Doe", Email: "john@test.com"},
				Created:  now.Add(-30 * 24 * time.Hour),
				Comments: []domain.Comment{
					{
						ID:      "1",
						Author:  "Jane Smith",
						Body:    "Any updates?",
						Created: now.Add(-1 * 24 * time.Hour),
					},
				},
			},
			want: true,
		},
		{
			name: "not stale - recent Updated even though Created is old (no comments loaded)",
			issue: domain.Issue{
				Assignee: &domain.TeamMember{Name: "John", Email: "john@test.com"},
				Created:  now.Add(-30 * 24 * time.Hour),
				Updated:  now.Add(-2 * 24 * time.Hour),
			},
			want: false,
		},
		{
			name: "stale - both Created and Updated are old (no comments)",
			issue: domain.Issue{
				Assignee: &domain.TeamMember{Name: "John", Email: "john@test.com"},
				Created:  now.Add(-20 * 24 * time.Hour),
				Updated:  now.Add(-16 * 24 * time.Hour),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.issue.IsStale()
			if got != tt.want {
				t.Errorf("IsStale() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIssue_LastAssigneeComment(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		issue   domain.Issue
		wantNil bool
		wantID  string
	}{
		{
			name: "no assignee",
			issue: domain.Issue{
				Assignee: nil,
			},
			wantNil: true,
		},
		{
			name: "no comments",
			issue: domain.Issue{
				Assignee: &domain.TeamMember{Name: "John", Email: "john@test.com"},
				Comments: []domain.Comment{},
			},
			wantNil: true,
		},
		{
			name: "no comments from assignee",
			issue: domain.Issue{
				Assignee: &domain.TeamMember{Name: "John Doe", Email: "john@test.com"},
				Comments: []domain.Comment{
					{ID: "1", Author: "Jane Smith", Body: "Comment", Created: now},
				},
			},
			wantNil: true,
		},
		{
			name: "returns most recent assignee comment",
			issue: domain.Issue{
				Assignee: &domain.TeamMember{Name: "John Doe", Email: "john@test.com"},
				Comments: []domain.Comment{
					{ID: "1", Author: "John Doe", Body: "First", Created: now.Add(-10 * 24 * time.Hour)},
					{ID: "2", Author: "Jane Smith", Body: "Reply", Created: now.Add(-5 * 24 * time.Hour)},
					{ID: "3", Author: "John Doe", Body: "Latest", Created: now.Add(-1 * 24 * time.Hour)},
				},
			},
			wantNil: false,
			wantID:  "3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.issue.LastAssigneeComment()
			if tt.wantNil {
				if got != nil {
					t.Errorf("LastAssigneeComment() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Errorf("LastAssigneeComment() = nil, want comment with ID %s", tt.wantID)
				return
			}
			if got.ID != tt.wantID {
				t.Errorf("LastAssigneeComment().ID = %s, want %s", got.ID, tt.wantID)
			}
		})
	}
}

func TestIssue_LastCommentAt(t *testing.T) {
	now := time.Now()

	t.Run("returns nil when no comments", func(t *testing.T) {
		issue := domain.Issue{}
		got := issue.LastCommentAt()
		if got != nil {
			t.Errorf("LastCommentAt() = %v, want nil", got)
		}
	})

	t.Run("returns nil when comments slice is empty", func(t *testing.T) {
		issue := domain.Issue{Comments: []domain.Comment{}}
		got := issue.LastCommentAt()
		if got != nil {
			t.Errorf("LastCommentAt() = %v, want nil", got)
		}
	})

	t.Run("returns date of single comment", func(t *testing.T) {
		expected := now.Add(-3 * 24 * time.Hour)
		issue := domain.Issue{
			Comments: []domain.Comment{
				{ID: "1", Author: "Alice", Created: expected},
			},
		}
		got := issue.LastCommentAt()
		if got == nil {
			t.Fatal("LastCommentAt() = nil, want non-nil")
		}
		if !got.Equal(expected) {
			t.Errorf("LastCommentAt() = %v, want %v", got, expected)
		}
	})

	t.Run("returns date of last comment in slice", func(t *testing.T) {
		early := now.Add(-10 * 24 * time.Hour)
		latest := now.Add(-1 * 24 * time.Hour)
		issue := domain.Issue{
			Comments: []domain.Comment{
				{ID: "1", Author: "Alice", Created: early},
				{ID: "2", Author: "Bob", Created: latest},
			},
		}
		got := issue.LastCommentAt()
		if got == nil {
			t.Fatal("LastCommentAt() = nil, want non-nil")
		}
		if !got.Equal(latest) {
			t.Errorf("LastCommentAt() = %v, want %v", got, latest)
		}
	})
}
