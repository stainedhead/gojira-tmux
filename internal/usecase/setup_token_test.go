package usecase_test

import (
	"testing"

	"github.com/stainedhead/gojira-tmux/internal/adapter/auth"
	"github.com/stainedhead/gojira-tmux/internal/usecase"
)

func TestSetupTokenUseCase_NeedsSetup(t *testing.T) {
	tests := []struct {
		name       string
		setupStore func() *auth.MemoryTokenStore
		want       bool
	}{
		{
			name: "needs setup when no token",
			setupStore: func() *auth.MemoryTokenStore {
				return auth.NewMemoryTokenStore()
			},
			want: true,
		},
		{
			name: "no setup needed when token exists",
			setupStore: func() *auth.MemoryTokenStore {
				store := auth.NewMemoryTokenStore()
				_ = store.SetJiraToken("existing-token")
				return store
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := tt.setupStore()
			uc := usecase.NewSetupToken(store)

			got := uc.NeedsSetup()
			if got != tt.want {
				t.Errorf("NeedsSetup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupTokenUseCase_SaveToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid token",
			token:   "valid-api-token-123",
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
			errMsg:  "token cannot be empty",
		},
		{
			name:    "whitespace only token",
			token:   "   ",
			wantErr: true,
			errMsg:  "token cannot be empty",
		},
		{
			name:    "token with leading/trailing whitespace is trimmed",
			token:   "  valid-token  ",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := auth.NewMemoryTokenStore()
			uc := usecase.NewSetupToken(store)

			err := uc.SaveToken(tt.token)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SaveToken() = nil, want error containing %q", tt.errMsg)
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("SaveToken() error = %q, want %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("SaveToken() unexpected error: %v", err)
				return
			}

			// Verify token was stored
			if !store.HasJiraToken() {
				t.Error("token was not stored")
			}
		})
	}
}

func TestSetupTokenUseCase_SaveToken_StoresCorrectValue(t *testing.T) {
	store := auth.NewMemoryTokenStore()
	uc := usecase.NewSetupToken(store)

	token := "my-api-token"
	err := uc.SaveToken(token)
	if err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}

	got, err := store.GetJiraToken()
	if err != nil {
		t.Fatalf("GetJiraToken() error = %v", err)
	}
	if got != token {
		t.Errorf("stored token = %q, want %q", got, token)
	}
}

func TestSetupTokenUseCase_SaveToken_TrimsWhitespace(t *testing.T) {
	store := auth.NewMemoryTokenStore()
	uc := usecase.NewSetupToken(store)

	err := uc.SaveToken("  my-token  ")
	if err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}

	got, err := store.GetJiraToken()
	if err != nil {
		t.Fatalf("GetJiraToken() error = %v", err)
	}
	if got != "my-token" {
		t.Errorf("stored token = %q, want %q", got, "my-token")
	}
}
