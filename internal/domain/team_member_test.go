package domain_test

import (
	"testing"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

func TestTeamMember_Validate(t *testing.T) {
	tests := []struct {
		name    string
		member  domain.TeamMember
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid team member",
			member: domain.TeamMember{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			member: domain.TeamMember{
				Name:  "",
				Email: "john@example.com",
			},
			wantErr: true,
			errMsg:  "team member name is required",
		},
		{
			name: "empty email",
			member: domain.TeamMember{
				Name:  "John Doe",
				Email: "",
			},
			wantErr: true,
			errMsg:  "team member email is required",
		},
		{
			name: "invalid email format - no @",
			member: domain.TeamMember{
				Name:  "John Doe",
				Email: "johnexample.com",
			},
			wantErr: true,
			errMsg:  "team member email is invalid",
		},
		{
			name: "invalid email format - no domain",
			member: domain.TeamMember{
				Name:  "John Doe",
				Email: "john@",
			},
			wantErr: true,
			errMsg:  "team member email is invalid",
		},
		{
			name: "invalid email format - no local part",
			member: domain.TeamMember{
				Name:  "John Doe",
				Email: "@example.com",
			},
			wantErr: true,
			errMsg:  "team member email is invalid",
		},
		{
			name: "valid email with subdomain",
			member: domain.TeamMember{
				Name:  "Jane Smith",
				Email: "jane@mail.example.com",
			},
			wantErr: false,
		},
		{
			name: "valid email with plus sign",
			member: domain.TeamMember{
				Name:  "Test User",
				Email: "test+filter@example.com",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.member.Validate()
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

func TestTeamMember_String(t *testing.T) {
	member := domain.TeamMember{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	expected := "John Doe <john@example.com>"
	if got := member.String(); got != expected {
		t.Errorf("String() = %q, want %q", got, expected)
	}
}
