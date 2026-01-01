package domain

import (
	"errors"
	"regexp"
	"time"
)

// issueKeyRegex validates issue keys match pattern like PROJ-123.
var issueKeyRegex = regexp.MustCompile(`^[A-Z]+-\d+$`)

// AttentionType indicates the type of attention indicator for an issue.
type AttentionType int

const (
	// AttentionNone indicates no attention needed.
	AttentionNone AttentionType = iota
	// AttentionNoDueDate indicates a yellow dot (open issue without due date).
	AttentionNoDueDate
	// AttentionStale indicates a red dot (assignee hasn't commented in 14+ days).
	AttentionStale
)

// Issue represents a Jira ticket.
type Issue struct {
	Key         string       `json:"key"`
	Summary     string       `json:"summary"`
	Description string       `json:"description"`
	Status      string       `json:"status"`
	Priority    string       `json:"priority"`
	Assignee    *TeamMember  `json:"assignee,omitempty"`
	Reporter    *TeamMember  `json:"reporter,omitempty"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
	Created     time.Time    `json:"created"`
	Updated     time.Time    `json:"updated"`
	Sprint      string       `json:"sprint,omitempty"`
	Epic        string       `json:"epic,omitempty"`
	Labels      []string     `json:"labels,omitempty"`
	StoryPoints int          `json:"story_points,omitempty"`
	Comments    []Comment    `json:"comments,omitempty"`
}

// Validate checks that the Issue has valid data.
func (i *Issue) Validate() error {
	if i.Key == "" {
		return errors.New("issue key is required")
	}
	if !issueKeyRegex.MatchString(i.Key) {
		return errors.New("issue key must match pattern PROJECT-123")
	}
	if i.Summary == "" {
		return errors.New("issue summary is required")
	}
	if i.Status == "" {
		return errors.New("issue status is required")
	}
	if i.Created.IsZero() {
		return errors.New("issue created time is required")
	}
	if i.Updated.IsZero() {
		return errors.New("issue updated time is required")
	}
	return nil
}

// NeedsAttention returns the attention indicator type for the issue.
// Red dot (stale) takes precedence over yellow dot (no due date).
func (i *Issue) NeedsAttention() AttentionType {
	if i.Status != "Open" {
		return AttentionNone
	}

	// Check for stale (assignee hasn't commented in 14+ days) - red dot takes precedence
	if i.IsStale() {
		return AttentionStale
	}

	// Check for missing due date - yellow dot
	if i.DueDate == nil {
		return AttentionNoDueDate
	}

	return AttentionNone
}

// IsStale returns true if the issue has an assignee who hasn't commented in 14+ days.
func (i *Issue) IsStale() bool {
	if i.Assignee == nil {
		return false
	}

	lastAssigneeComment := i.LastAssigneeComment()
	if lastAssigneeComment == nil {
		// No assignee comment ever - check if issue is older than 14 days
		return time.Since(i.Created) > 14*24*time.Hour
	}

	return time.Since(lastAssigneeComment.Created) > 14*24*time.Hour
}

// LastAssigneeComment returns the most recent comment by the assignee.
// Returns nil if no assignee or no comments from assignee.
func (i *Issue) LastAssigneeComment() *Comment {
	if i.Assignee == nil {
		return nil
	}

	// Iterate backwards to find the most recent comment from assignee
	for j := len(i.Comments) - 1; j >= 0; j-- {
		if i.Comments[j].Author == i.Assignee.Name {
			return &i.Comments[j]
		}
	}
	return nil
}
