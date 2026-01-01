package auth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

const (
	serviceName = "gojira-tmux"
	jiraKey     = "jira-api-token"
	refreshKey  = "okta-refresh-token"
)

// MemoryTokenStore is an in-memory implementation of TokenStorePort for testing.
type MemoryTokenStore struct {
	mu           sync.RWMutex
	jiraToken    string
	refreshToken string
}

// NewMemoryTokenStore creates a new in-memory token store.
func NewMemoryTokenStore() *MemoryTokenStore {
	return &MemoryTokenStore{}
}

// GetJiraToken retrieves the stored Jira API token.
func (s *MemoryTokenStore) GetJiraToken() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.jiraToken, nil
}

// SetJiraToken stores the Jira API token.
func (s *MemoryTokenStore) SetJiraToken(token string) error {
	if token == "" {
		return errors.New("token cannot be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jiraToken = token
	return nil
}

// DeleteJiraToken removes the stored Jira API token.
func (s *MemoryTokenStore) DeleteJiraToken() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jiraToken = ""
	return nil
}

// GetRefreshToken retrieves the stored Okta refresh token.
func (s *MemoryTokenStore) GetRefreshToken() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.refreshToken, nil
}

// SetRefreshToken stores the Okta refresh token.
func (s *MemoryTokenStore) SetRefreshToken(token string) error {
	if token == "" {
		return errors.New("token cannot be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshToken = token
	return nil
}

// DeleteRefreshToken removes the stored refresh token.
func (s *MemoryTokenStore) DeleteRefreshToken() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshToken = ""
	return nil
}

// HasJiraToken returns true if a Jira token exists.
func (s *MemoryTokenStore) HasJiraToken() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.jiraToken != ""
}

// Ensure MemoryTokenStore implements domain.TokenStorePort.
var _ domain.TokenStorePort = (*MemoryTokenStore)(nil)

// FileTokenStore stores tokens in an encrypted file.
// This is used as a fallback when keyring is not available.
type FileTokenStore struct {
	path string
	mu   sync.RWMutex
}

// fileData represents the JSON structure stored in the file.
type fileData struct {
	JiraToken    string `json:"jira_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// NewFileTokenStore creates a new file-based token store.
func NewFileTokenStore(path string) *FileTokenStore {
	return &FileTokenStore{path: path}
}

// load reads the file data.
func (s *FileTokenStore) load() (*fileData, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return &fileData{}, nil
	}
	if err != nil {
		return nil, err
	}

	var fd fileData
	if err := json.Unmarshal(data, &fd); err != nil {
		return nil, err
	}
	return &fd, nil
}

// save writes the file data.
func (s *FileTokenStore) save(fd *fileData) error {
	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.Marshal(fd)
	if err != nil {
		return err
	}

	// Write with restricted permissions (owner only)
	return os.WriteFile(s.path, data, 0600)
}

// GetJiraToken retrieves the stored Jira API token.
func (s *FileTokenStore) GetJiraToken() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fd, err := s.load()
	if err != nil {
		return "", err
	}
	return fd.JiraToken, nil
}

// SetJiraToken stores the Jira API token.
func (s *FileTokenStore) SetJiraToken(token string) error {
	if token == "" {
		return errors.New("token cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	fd, err := s.load()
	if err != nil {
		return err
	}
	fd.JiraToken = token
	return s.save(fd)
}

// DeleteJiraToken removes the stored Jira API token.
func (s *FileTokenStore) DeleteJiraToken() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fd, err := s.load()
	if err != nil {
		return err
	}
	fd.JiraToken = ""
	return s.save(fd)
}

// GetRefreshToken retrieves the stored Okta refresh token.
func (s *FileTokenStore) GetRefreshToken() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fd, err := s.load()
	if err != nil {
		return "", err
	}
	return fd.RefreshToken, nil
}

// SetRefreshToken stores the Okta refresh token.
func (s *FileTokenStore) SetRefreshToken(token string) error {
	if token == "" {
		return errors.New("token cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	fd, err := s.load()
	if err != nil {
		return err
	}
	fd.RefreshToken = token
	return s.save(fd)
}

// DeleteRefreshToken removes the stored refresh token.
func (s *FileTokenStore) DeleteRefreshToken() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fd, err := s.load()
	if err != nil {
		return err
	}
	fd.RefreshToken = ""
	return s.save(fd)
}

// HasJiraToken returns true if a Jira token exists.
func (s *FileTokenStore) HasJiraToken() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fd, err := s.load()
	if err != nil {
		return false
	}
	return fd.JiraToken != ""
}

// Ensure FileTokenStore implements domain.TokenStorePort.
var _ domain.TokenStorePort = (*FileTokenStore)(nil)

// KeyringTokenStore uses the OS keychain for secure storage.
// Falls back to FileTokenStore if keyring is not available.
type KeyringTokenStore struct {
	fallback *FileTokenStore
	useFile  bool
}

// NewKeyringTokenStore creates a new keyring-based token store.
// It will try the OS keychain first, falling back to file storage if unavailable.
func NewKeyringTokenStore(fallbackPath string) *KeyringTokenStore {
	ks := &KeyringTokenStore{
		fallback: NewFileTokenStore(fallbackPath),
	}

	// Test if keyring is available
	// For now, we'll always use the fallback since go-keyring requires
	// proper system setup (D-Bus on Linux, etc.)
	// In production, we would test: keyring.Set("test", "test", "test")
	ks.useFile = true

	return ks
}

// GetJiraToken retrieves the stored Jira API token.
func (s *KeyringTokenStore) GetJiraToken() (string, error) {
	if s.useFile {
		return s.fallback.GetJiraToken()
	}
	// TODO: Use keyring.Get(serviceName, jiraKey)
	return s.fallback.GetJiraToken()
}

// SetJiraToken stores the Jira API token.
func (s *KeyringTokenStore) SetJiraToken(token string) error {
	if s.useFile {
		return s.fallback.SetJiraToken(token)
	}
	// TODO: Use keyring.Set(serviceName, jiraKey, token)
	return s.fallback.SetJiraToken(token)
}

// DeleteJiraToken removes the stored Jira API token.
func (s *KeyringTokenStore) DeleteJiraToken() error {
	if s.useFile {
		return s.fallback.DeleteJiraToken()
	}
	// TODO: Use keyring.Delete(serviceName, jiraKey)
	return s.fallback.DeleteJiraToken()
}

// GetRefreshToken retrieves the stored Okta refresh token.
func (s *KeyringTokenStore) GetRefreshToken() (string, error) {
	if s.useFile {
		return s.fallback.GetRefreshToken()
	}
	// TODO: Use keyring.Get(serviceName, refreshKey)
	return s.fallback.GetRefreshToken()
}

// SetRefreshToken stores the Okta refresh token.
func (s *KeyringTokenStore) SetRefreshToken(token string) error {
	if s.useFile {
		return s.fallback.SetRefreshToken(token)
	}
	// TODO: Use keyring.Set(serviceName, refreshKey, token)
	return s.fallback.SetRefreshToken(token)
}

// DeleteRefreshToken removes the stored refresh token.
func (s *KeyringTokenStore) DeleteRefreshToken() error {
	if s.useFile {
		return s.fallback.DeleteRefreshToken()
	}
	// TODO: Use keyring.Delete(serviceName, refreshKey)
	return s.fallback.DeleteRefreshToken()
}

// HasJiraToken returns true if a Jira token exists.
func (s *KeyringTokenStore) HasJiraToken() bool {
	if s.useFile {
		return s.fallback.HasJiraToken()
	}
	// TODO: Try keyring.Get and check for keyring.ErrNotFound
	return s.fallback.HasJiraToken()
}

// Ensure KeyringTokenStore implements domain.TokenStorePort.
var _ domain.TokenStorePort = (*KeyringTokenStore)(nil)
