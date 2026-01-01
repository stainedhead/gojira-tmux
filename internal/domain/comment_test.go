package domain_test

import (
	"testing"
	"time"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

func TestComment_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		comment domain.Comment
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid comment",
			comment: domain.Comment{
				ID:      "12345",
				Author:  "John Doe",
				Body:    "This is a comment",
				Created: now,
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			comment: domain.Comment{
				ID:      "",
				Author:  "John Doe",
				Body:    "This is a comment",
				Created: now,
			},
			wantErr: true,
			errMsg:  "comment ID is required",
		},
		{
			name: "empty author",
			comment: domain.Comment{
				ID:      "12345",
				Author:  "",
				Body:    "This is a comment",
				Created: now,
			},
			wantErr: true,
			errMsg:  "comment author is required",
		},
		{
			name: "empty body is allowed",
			comment: domain.Comment{
				ID:      "12345",
				Author:  "John Doe",
				Body:    "",
				Created: now,
			},
			wantErr: false,
		},
		{
			name: "zero time",
			comment: domain.Comment{
				ID:      "12345",
				Author:  "John Doe",
				Body:    "Comment",
				Created: time.Time{},
			},
			wantErr: true,
			errMsg:  "comment created time is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.comment.Validate()
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

func TestComment_Age(t *testing.T) {
	tests := []struct {
		name     string
		created  time.Time
		wantDays int
	}{
		{
			name:     "comment from today",
			created:  time.Now(),
			wantDays: 0,
		},
		{
			name:     "comment from 7 days ago",
			created:  time.Now().Add(-7 * 24 * time.Hour),
			wantDays: 7,
		},
		{
			name:     "comment from 30 days ago",
			created:  time.Now().Add(-30 * 24 * time.Hour),
			wantDays: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comment := domain.Comment{
				ID:      "123",
				Author:  "Test",
				Body:    "Test",
				Created: tt.created,
			}
			gotDays := comment.AgeDays()
			// Allow 1 day tolerance for edge cases around midnight
			if gotDays < tt.wantDays || gotDays > tt.wantDays+1 {
				t.Errorf("AgeDays() = %d, want approximately %d", gotDays, tt.wantDays)
			}
		})
	}
}
