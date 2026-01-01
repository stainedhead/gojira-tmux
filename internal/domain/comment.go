package domain

import (
	"errors"
	"time"
)

// Comment represents a Jira issue comment.
type Comment struct {
	ID      string    `json:"id"`
	Author  string    `json:"author"`
	Body    string    `json:"body"`
	Created time.Time `json:"created"`
}

// Validate checks that the Comment has valid data.
func (c *Comment) Validate() error {
	if c.ID == "" {
		return errors.New("comment ID is required")
	}
	if c.Author == "" {
		return errors.New("comment author is required")
	}
	if c.Created.IsZero() {
		return errors.New("comment created time is required")
	}
	return nil
}

// AgeDays returns the number of days since the comment was created.
func (c *Comment) AgeDays() int {
	duration := time.Since(c.Created)
	return int(duration.Hours() / 24)
}
