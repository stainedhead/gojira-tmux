package tui

import (
	"testing"
	"time"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

func TestFormatDueDate(t *testing.T) {
	tests := []struct {
		name    string
		dueDate *time.Time
		want    string
	}{
		{
			name:    "nil due date returns none",
			dueDate: nil,
			want:    "none",
		},
		{
			name: "formats date as YYYY-MM-DD",
			dueDate: func() *time.Time {
				d := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
				return &d
			}(),
			want: "2025-03-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDueDate(tt.dueDate)
			if got != tt.want {
				t.Errorf("formatDueDate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		t    *time.Time
		want string
	}{
		{
			name: "nil returns never",
			t:    nil,
			want: "never",
		},
		{
			name: "today returns today",
			t: func() *time.Time {
				d := now.Add(-2 * time.Hour)
				return &d
			}(),
			want: "today",
		},
		{
			name: "1 day ago returns 1d ago",
			t: func() *time.Time {
				d := now.Add(-25 * time.Hour)
				return &d
			}(),
			want: "1d ago",
		},
		{
			name: "30 days ago returns 30d ago",
			t: func() *time.Time {
				d := now.Add(-30 * 24 * time.Hour)
				return &d
			}(),
			want: "30d ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatRelativeTime(tt.t)
			if got != tt.want {
				t.Errorf("formatRelativeTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCommentsPanel_View_NoPanicWithSmallWidth(t *testing.T) {
	panel := NewCommentsPanel()
	panel.SetSize(2, 10) // width < 4 — separator would be negative
	panel.SetComments([]domain.Comment{
		{Author: "Alice", Body: "First", Created: time.Now().Add(-2 * time.Hour)},
		{Author: "Bob", Body: "Second", Created: time.Now().Add(-time.Hour)},
	})
	// Must not panic
	_ = panel.View()
}

func TestCommentsPanel_View_NoPanicWithZeroWidth(t *testing.T) {
	panel := NewCommentsPanel()
	panel.SetSize(0, 0)
	panel.SetComments([]domain.Comment{
		{Author: "Alice", Body: "Comment", Created: time.Now()},
		{Author: "Bob", Body: "Reply", Created: time.Now().Add(-time.Hour)},
	})
	_ = panel.View()
}

func TestPropertiesPanel_View_NoPanicWithZeroWidth(t *testing.T) {
	panel := NewPropertiesPanel()
	panel.SetSize(0, 0)
	now := time.Now()
	panel.SetIssue(&domain.Issue{
		Key:         "PROJ-1",
		Summary:     "Test issue",
		Status:      "In Progress",
		Description: "A long description that would need wrapping",
		Created:     now,
		Updated:     now,
	})
	_ = panel.View()
}

func TestFormatLabels(t *testing.T) {
	tests := []struct {
		name   string
		labels []string
		maxLen int
		want   string
	}{
		{
			name:   "empty labels returns none",
			labels: nil,
			maxLen: 20,
			want:   "none",
		},
		{
			name:   "single label",
			labels: []string{"bug"},
			maxLen: 20,
			want:   "bug",
		},
		{
			name:   "multiple labels joined with comma",
			labels: []string{"bug", "backend", "p1"},
			maxLen: 30,
			want:   "bug, backend, p1",
		},
		{
			name:   "truncates when too long",
			labels: []string{"very-long-label-one", "very-long-label-two"},
			maxLen: 20,
			want:   "very-long-label-o...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatLabels(tt.labels, tt.maxLen)
			if got != tt.want {
				t.Errorf("formatLabels() = %q, want %q", got, tt.want)
			}
		})
	}
}
