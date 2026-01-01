package domain

import (
	"errors"
	"fmt"
	"strings"
)

// TeamMember represents a team member for filtering and display.
type TeamMember struct {
	Name  string `yaml:"name" json:"name"`
	Email string `yaml:"email" json:"email"`
}

// Validate checks that the TeamMember has valid data.
func (t *TeamMember) Validate() error {
	if t.Name == "" {
		return errors.New("team member name is required")
	}
	if t.Email == "" {
		return errors.New("team member email is required")
	}
	if !isValidEmail(t.Email) {
		return errors.New("team member email is invalid")
	}
	return nil
}

// String returns a formatted string representation.
func (t *TeamMember) String() string {
	return fmt.Sprintf("%s <%s>", t.Name, t.Email)
}

// isValidEmail performs basic email validation.
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	local, domain := parts[0], parts[1]
	if local == "" || domain == "" {
		return false
	}
	// Check domain has at least one dot or is a valid TLD
	if !strings.Contains(domain, ".") && len(domain) < 2 {
		return false
	}
	return true
}
