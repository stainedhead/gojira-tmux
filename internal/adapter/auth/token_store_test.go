package auth_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stainedhead/gojira-tmux/internal/adapter/auth"
)

func TestMemoryTokenStore(t *testing.T) {
	store := auth.NewMemoryTokenStore()

	t.Run("initial state has no token", func(t *testing.T) {
		if store.HasJiraToken() {
			t.Error("HasJiraToken() = true, want false for new store")
		}

		token, err := store.GetJiraToken()
		if err != nil {
			t.Errorf("GetJiraToken() error = %v, want nil", err)
		}
		if token != "" {
			t.Errorf("GetJiraToken() = %q, want empty string", token)
		}
	})

	t.Run("set and get jira token", func(t *testing.T) {
		testToken := "test-api-token-123"

		err := store.SetJiraToken(testToken)
		if err != nil {
			t.Errorf("SetJiraToken() error = %v, want nil", err)
		}

		if !store.HasJiraToken() {
			t.Error("HasJiraToken() = false, want true after setting token")
		}

		got, err := store.GetJiraToken()
		if err != nil {
			t.Errorf("GetJiraToken() error = %v, want nil", err)
		}
		if got != testToken {
			t.Errorf("GetJiraToken() = %q, want %q", got, testToken)
		}
	})

	t.Run("delete jira token", func(t *testing.T) {
		err := store.DeleteJiraToken()
		if err != nil {
			t.Errorf("DeleteJiraToken() error = %v, want nil", err)
		}

		if store.HasJiraToken() {
			t.Error("HasJiraToken() = true, want false after delete")
		}

		token, err := store.GetJiraToken()
		if err != nil {
			t.Errorf("GetJiraToken() error = %v, want nil", err)
		}
		if token != "" {
			t.Errorf("GetJiraToken() = %q, want empty string", token)
		}
	})
}

func TestFileTokenStore(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "credentials")

	store := auth.NewFileTokenStore(storePath)

	t.Run("initial state has no token", func(t *testing.T) {
		if store.HasJiraToken() {
			t.Error("HasJiraToken() = true, want false for new store")
		}

		token, err := store.GetJiraToken()
		if err != nil {
			t.Errorf("GetJiraToken() error = %v, want nil", err)
		}
		if token != "" {
			t.Errorf("GetJiraToken() = %q, want empty string", token)
		}
	})

	t.Run("set and get jira token", func(t *testing.T) {
		testToken := "test-api-token-123"

		err := store.SetJiraToken(testToken)
		if err != nil {
			t.Errorf("SetJiraToken() error = %v, want nil", err)
		}

		if !store.HasJiraToken() {
			t.Error("HasJiraToken() = false, want true after setting token")
		}

		got, err := store.GetJiraToken()
		if err != nil {
			t.Errorf("GetJiraToken() error = %v, want nil", err)
		}
		if got != testToken {
			t.Errorf("GetJiraToken() = %q, want %q", got, testToken)
		}
	})

	t.Run("file permissions", func(t *testing.T) {
		info, err := os.Stat(storePath)
		if err != nil {
			t.Fatalf("failed to stat credentials file: %v", err)
		}
		mode := info.Mode().Perm()
		if mode&0o077 != 0 {
			t.Errorf("file permissions = %o, want owner-only permissions", mode)
		}
	})

	t.Run("persistence across instances", func(t *testing.T) {
		testToken := "persistent-token-789"

		err := store.SetJiraToken(testToken)
		if err != nil {
			t.Fatalf("SetJiraToken() error = %v", err)
		}

		store2 := auth.NewFileTokenStore(storePath)

		got, err := store2.GetJiraToken()
		if err != nil {
			t.Errorf("GetJiraToken() error = %v, want nil", err)
		}
		if got != testToken {
			t.Errorf("GetJiraToken() = %q, want %q", got, testToken)
		}
	})

	t.Run("delete jira token", func(t *testing.T) {
		err := store.DeleteJiraToken()
		if err != nil {
			t.Errorf("DeleteJiraToken() error = %v, want nil", err)
		}

		if store.HasJiraToken() {
			t.Error("HasJiraToken() = true, want false after delete")
		}
	})
}

func TestTokenStore_EmptyToken(t *testing.T) {
	store := auth.NewMemoryTokenStore()

	t.Run("set empty jira token", func(t *testing.T) {
		err := store.SetJiraToken("")
		if err == nil {
			t.Error("SetJiraToken(\"\") = nil, want error")
		}
	})
}
