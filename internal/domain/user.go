package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// User represents the authenticated user session.
type User struct {
	Email         string    `json:"email"`
	SessionExpiry time.Time `json:"session_expiry"`
}

// Validate checks that the User has valid data.
func (u *User) Validate() error {
	if u.Email == "" {
		return errors.New("user email is required")
	}
	if !isValidEmail(u.Email) {
		return errors.New("user email is invalid")
	}
	if u.SessionExpiry.IsZero() {
		return errors.New("session expiry is required")
	}
	return nil
}

// IsSessionValid returns true if the session has not expired.
func (u *User) IsSessionValid() bool {
	if u.SessionExpiry.IsZero() {
		return false
	}
	return time.Now().Before(u.SessionExpiry)
}

// ValidateTeamMembership checks if the user email exists in the team list.
func (u *User) ValidateTeamMembership(team []TeamMember) error {
	for _, m := range team {
		if strings.EqualFold(m.Email, u.Email) {
			return nil
		}
	}
	return fmt.Errorf("user %s is not a member of the configured team", u.Email)
}
