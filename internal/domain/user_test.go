package domain_test

import (
	"testing"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    domain.User
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid user",
			user: domain.User{
				Email: "john@example.com",
			},
			wantErr: false,
		},
		{
			name: "empty email",
			user: domain.User{
				Email: "",
			},
			wantErr: true,
			errMsg:  "user email is required",
		},
		{
			name: "invalid email format",
			user: domain.User{
				Email: "invalid-email",
			},
			wantErr: true,
			errMsg:  "user email is invalid",
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
