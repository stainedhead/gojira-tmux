package tui

import (
	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// TokenStoredMsg indicates API token was successfully stored.
type TokenStoredMsg struct{}

// IssuesLoadedMsg indicates issues were loaded.
type IssuesLoadedMsg struct {
	Issues []domain.Issue
}

// IssueSelectedMsg indicates an issue was selected.
type IssueSelectedMsg struct {
	Issue *domain.Issue
}

// IssueDetailsLoadedMsg indicates issue details were loaded.
type IssueDetailsLoadedMsg struct {
	Issue *domain.Issue
}

// ErrorMsg indicates an error occurred.
type ErrorMsg struct {
	Err error
}

// RefreshMsg triggers a data refresh.
type RefreshMsg struct{}

// FilterChangedMsg indicates filter state changed.
type FilterChangedMsg struct {
	Filter domain.IssueFilter
}

// LoadingMsg indicates loading state changed.
type LoadingMsg struct {
	Loading bool
	Message string
}

// LogoutMsg indicates user requested logout.
type LogoutMsg struct{}
