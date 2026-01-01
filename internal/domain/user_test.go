package domain_test

import (
	"testing"
	"time"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

func TestUser_Validate(t *testing.T) {
	validExpiry := time.Now().Add(8 * time.Hour)

	tests := []struct {
		name    string
		user    domain.User
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid user",
			user: domain.User{
				Email:         "john@example.com",
				SessionExpiry: validExpiry,
			},
			wantErr: false,
		},
		{
			name: "empty email",
			user: domain.User{
				Email:         "",
				SessionExpiry: validExpiry,
			},
			wantErr: true,
			errMsg:  "user email is required",
		},
		{
			name: "invalid email format",
			user: domain.User{
				Email:         "invalid-email",
				SessionExpiry: validExpiry,
			},
			wantErr: true,
			errMsg:  "user email is invalid",
		},
		{
			name: "zero session expiry",
			user: domain.User{
				Email:         "john@example.com",
				SessionExpiry: time.Time{},
			},
			wantErr: true,
			errMsg:  "session expiry is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
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

func TestUser_IsSessionValid(t *testing.T) {
	tests := []struct {
		name   string
		expiry time.Time
		want   bool
	}{
		{
			name:   "valid session - expires in 8 hours",
			expiry: time.Now().Add(8 * time.Hour),
			want:   true,
		},
		{
			name:   "valid session - expires in 1 minute",
			expiry: time.Now().Add(1 * time.Minute),
			want:   true,
		},
		{
			name:   "expired session - expired 1 hour ago",
			expiry: time.Now().Add(-1 * time.Hour),
			want:   false,
		},
		{
			name:   "expired session - expired 1 minute ago",
			expiry: time.Now().Add(-1 * time.Minute),
			want:   false,
		},
		{
			name:   "zero expiry is invalid",
			expiry: time.Time{},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := domain.User{
				Email:         "test@example.com",
				SessionExpiry: tt.expiry,
			}
			if got := user.IsSessionValid(); got != tt.want {
				t.Errorf("IsSessionValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_ValidateTeamMembership(t *testing.T) {
	team := []domain.TeamMember{
		{Name: "John Doe", Email: "john@example.com"},
		{Name: "Jane Smith", Email: "jane@example.com"},
	}

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "user is team member",
			email:   "john@example.com",
			wantErr: false,
		},
		{
			name:    "user is team member (case insensitive)",
			email:   "JOHN@example.com",
			wantErr: false,
		},
		{
			name:    "user is not team member",
			email:   "stranger@example.com",
			wantErr: true,
		},
		{
			name:    "empty team",
			email:   "john@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := domain.User{
				Email:         tt.email,
				SessionExpiry: time.Now().Add(1 * time.Hour),
			}

			teamToUse := team
			if tt.name == "empty team" {
				teamToUse = []domain.TeamMember{}
			}

			err := user.ValidateTeamMembership(teamToUse)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTeamMembership() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
