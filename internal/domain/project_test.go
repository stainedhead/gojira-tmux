package domain_test

import (
	"testing"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

func TestProject_Validate(t *testing.T) {
	tests := []struct {
		name    string
		project domain.Project
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid project",
			project: domain.Project{
				Key:  "PROJ",
				Name: "My Project",
			},
			wantErr: false,
		},
		{
			name: "valid project with long key",
			project: domain.Project{
				Key:  "MYPROJECT",
				Name: "My Long Project",
			},
			wantErr: false,
		},
		{
			name: "empty key",
			project: domain.Project{
				Key:  "",
				Name: "My Project",
			},
			wantErr: true,
			errMsg:  "project key is required",
		},
		{
			name: "empty name",
			project: domain.Project{
				Key:  "PROJ",
				Name: "",
			},
			wantErr: true,
			errMsg:  "project name is required",
		},
		{
			name: "lowercase key",
			project: domain.Project{
				Key:  "proj",
				Name: "My Project",
			},
			wantErr: true,
			errMsg:  "project key must be uppercase letters only",
		},
		{
			name: "key with numbers",
			project: domain.Project{
				Key:  "PROJ123",
				Name: "My Project",
			},
			wantErr: true,
			errMsg:  "project key must be uppercase letters only",
		},
		{
			name: "key with special characters",
			project: domain.Project{
				Key:  "PROJ-A",
				Name: "My Project",
			},
			wantErr: true,
			errMsg:  "project key must be uppercase letters only",
		},
		{
			name: "mixed case key",
			project: domain.Project{
				Key:  "ProJ",
				Name: "My Project",
			},
			wantErr: true,
			errMsg:  "project key must be uppercase letters only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.project.Validate()
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

func TestProject_String(t *testing.T) {
	project := domain.Project{
		Key:  "PROJ",
		Name: "My Project",
	}

	expected := "PROJ - My Project"
	if got := project.String(); got != expected {
		t.Errorf("String() = %q, want %q", got, expected)
	}
}
