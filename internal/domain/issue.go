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

// doneStatuses are terminal statuses that indicate an issue is resolved.
// Issues in these statuses do not need attention indicators.
var doneStatuses = map[string]bool{
	"Done":      true,
	"Closed":    true,
	"Resolved":  true,
	"Cancelled": true,
	"Won't Fix": true,
	"Duplicate": true,
}

// NeedsAttention returns the attention indicator type for the issue.
// Red dot (stale) takes precedence over yellow dot (no due date).
// Issues in terminal/done statuses never need attention.
func (i *Issue) NeedsAttention() AttentionType {
	if doneStatuses[i.Status] {
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

// IsStale returns true if the issue has an assignee who hasn't engaged in 14+ days.
// When comments are loaded, the assignee's last comment date is used.
// When no assignee comment is found, Updated timestamp is used as a proxy
// (falling back to Created if Updated is not set).
func (i *Issue) IsStale() bool {
	if i.Assignee == nil {
		return false
	}

	lastAssigneeComment := i.LastAssigneeComment()
	if lastAssigneeComment != nil {
		return time.Since(lastAssigneeComment.Created) > 14*24*time.Hour
	}

	// No assignee comment found — use Updated as proxy for recent activity.
	lastActivity := i.Updated
	if lastActivity.IsZero() {
		lastActivity = i.Created
	}
	return time.Since(lastActivity) > 14*24*time.Hour
}

// LastCommentAt returns the creation time of the most recent comment,
// or nil if there are no comments.
func (i *Issue) LastCommentAt() *time.Time {
	if len(i.Comments) == 0 {
		return nil
	}
	t := i.Comments[len(i.Comments)-1].Created
	if t.IsZero() {
		return nil
	}
	return &t
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
