package auth_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stainedhead/gojira-tmux/internal/adapter/auth"
)

func TestMemoryTokenStore(t *testing.T) {
	store := auth.NewMemoryTokenStore()

	// Test initial state
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

	// Test set and get
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

	// Test delete
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

	// Test refresh token
	t.Run("set and get refresh token", func(t *testing.T) {
		testRefreshToken := "refresh-token-456"

		err := store.SetRefreshToken(testRefreshToken)
		if err != nil {
			t.Errorf("SetRefreshToken() error = %v, want nil", err)
		}

		got, err := store.GetRefreshToken()
		if err != nil {
			t.Errorf("GetRefreshToken() error = %v, want nil", err)
		}
		if got != testRefreshToken {
			t.Errorf("GetRefreshToken() = %q, want %q", got, testRefreshToken)
		}
	})

	// Test delete refresh token
	t.Run("delete refresh token", func(t *testing.T) {
		err := store.DeleteRefreshToken()
		if err != nil {
			t.Errorf("DeleteRefreshToken() error = %v, want nil", err)
		}

		token, err := store.GetRefreshToken()
		if err != nil {
			t.Errorf("GetRefreshToken() error = %v, want nil", err)
		}
		if token != "" {
			t.Errorf("GetRefreshToken() = %q, want empty string", token)
		}
	})
}

func TestFileTokenStore(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "credentials")

	store := auth.NewFileTokenStore(storePath)

	// Test initial state
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

	// Test set and get
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

	// Verify file was created with restricted permissions
	t.Run("file permissions", func(t *testing.T) {
		info, err := os.Stat(storePath)
		if err != nil {
			t.Fatalf("failed to stat credentials file: %v", err)
		}
		// Check that file is not world or group readable (0600 or similar)
		mode := info.Mode().Perm()
		if mode&0o077 != 0 {
			t.Errorf("file permissions = %o, want owner-only permissions", mode)
		}
	})

	// Test persistence
	t.Run("persistence across instances", func(t *testing.T) {
		testToken := "persistent-token-789"

		err := store.SetJiraToken(testToken)
		if err != nil {
			t.Fatalf("SetJiraToken() error = %v", err)
		}

		// Create new store instance
		store2 := auth.NewFileTokenStore(storePath)

		got, err := store2.GetJiraToken()
		if err != nil {
			t.Errorf("GetJiraToken() error = %v, want nil", err)
		}
		if got != testToken {
			t.Errorf("GetJiraToken() = %q, want %q", got, testToken)
		}
	})

	// Test delete
	t.Run("delete jira token", func(t *testing.T) {
		err := store.DeleteJiraToken()
		if err != nil {
			t.Errorf("DeleteJiraToken() error = %v, want nil", err)
		}

		if store.HasJiraToken() {
			t.Error("HasJiraToken() = true, want false after delete")
		}
	})

	// Test refresh token
	t.Run("set and get refresh token", func(t *testing.T) {
		testRefreshToken := "refresh-token-456"

		err := store.SetRefreshToken(testRefreshToken)
		if err != nil {
			t.Errorf("SetRefreshToken() error = %v, want nil", err)
		}

		got, err := store.GetRefreshToken()
		if err != nil {
			t.Errorf("GetRefreshToken() error = %v, want nil", err)
		}
		if got != testRefreshToken {
			t.Errorf("GetRefreshToken() = %q, want %q", got, testRefreshToken)
		}
	})
}

func TestTokenStore_EmptyToken(t *testing.T) {
	store := auth.NewMemoryTokenStore()

	// Setting empty token should fail
	t.Run("set empty jira token", func(t *testing.T) {
		err := store.SetJiraToken("")
		if err == nil {
			t.Error("SetJiraToken(\"\") = nil, want error")
		}
	})

	t.Run("set empty refresh token", func(t *testing.T) {
		err := store.SetRefreshToken("")
		if err == nil {
			t.Error("SetRefreshToken(\"\") = nil, want error")
		}
	})
}
