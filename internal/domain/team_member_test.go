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
			name: "valid team member with alias",
			member: domain.TeamMember{
				Name:  "John Doe",
				Email: "john@example.com",
				Alias: "JohnD",
			},
			wantErr: false,
		},
		{
			name: "valid alias alphanumeric",
			member: domain.TeamMember{
				Name:  "John Doe",
				Email: "john@example.com",
				Alias: "John123",
			},
			wantErr: false,
		},
		{
			name: "valid alias two letters",
			member: domain.TeamMember{
				Name:  "Sarah Anderson",
				Email: "sa@example.com",
				Alias: "SA",
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
			name: "invalid alias with space",
			member: domain.TeamMember{
				Name:  "John Doe",
				Email: "john@example.com",
				Alias: "John D",
			},
			wantErr: true,
			errMsg:  "team member alias must be alphanumeric (no spaces)",
		},
		{
			name: "invalid alias with hyphen",
			member: domain.TeamMember{
				Name:  "John Doe",
				Email: "john@example.com",
				Alias: "John-D",
			},
			wantErr: true,
			errMsg:  "team member alias must be alphanumeric (no spaces)",
		},
		{
			name: "invalid alias with dot",
			member: domain.TeamMember{
				Name:  "John Doe",
				Email: "john@example.com",
				Alias: "John.D",
			},
			wantErr: true,
			errMsg:  "team member alias must be alphanumeric (no spaces)",
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

func TestTeamMember_MatchesIdentifier(t *testing.T) {
	member := domain.TeamMember{
		Name:  "John Anderson",
		Email: "john.anderson@example.com",
		Alias: "JohnA",
	}

	memberNoAlias := domain.TeamMember{
		Name:  "Jane Smith",
		Email: "jane@example.com",
	}

	tests := []struct {
		name       string
		member     domain.TeamMember
		identifier string
		want       bool
	}{
		{
			name:       "exact alias match",
			member:     member,
			identifier: "JohnA",
			want:       true,
		},
		{
			name:       "exact name match",
			member:     member,
			identifier: "John Anderson",
			want:       true,
		},
		{
			name:       "case-insensitive alias match",
			member:     member,
			identifier: "johna",
			want:       true,
		},
		{
			name:       "case-insensitive name match",
			member:     member,
			identifier: "john anderson",
			want:       true,
		},
		{
			name:       "no match",
			member:     member,
			identifier: "Unknown",
			want:       false,
		},
		{
			name:       "partial name no match",
			member:     member,
			identifier: "John",
			want:       false,
		},
		{
			name:       "member without alias - name match",
			member:     memberNoAlias,
			identifier: "Jane Smith",
			want:       true,
		},
		{
			name:       "member without alias - case insensitive",
			member:     memberNoAlias,
			identifier: "jane smith",
			want:       true,
		},
		{
			name:       "member without alias - no match",
			member:     memberNoAlias,
			identifier: "JohnA",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.member.MatchesIdentifier(tt.identifier); got != tt.want {
				t.Errorf("MatchesIdentifier(%q) = %v, want %v", tt.identifier, got, tt.want)
			}
		})
	}
}

func TestTeamMember_DisplayName(t *testing.T) {
	tests := []struct {
		name   string
		member domain.TeamMember
		want   string
	}{
		{
			name: "with alias",
			member: domain.TeamMember{
				Name:  "John Anderson",
				Email: "john@example.com",
				Alias: "JohnA",
			},
			want: "John Anderson (JohnA)",
		},
		{
			name: "without alias",
			member: domain.TeamMember{
				Name:  "Jane Smith",
				Email: "jane@example.com",
			},
			want: "Jane Smith",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.member.DisplayName(); got != tt.want {
				t.Errorf("DisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}
