package domain

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// TeamMember represents a team member for filtering and display.
type TeamMember struct {
	Name  string `yaml:"name" json:"name"`
	Email string `yaml:"email" json:"email"`
	Alias string `yaml:"alias,omitempty" json:"alias,omitempty"`
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
	if t.Alias != "" && !isValidAlias(t.Alias) {
		return errors.New("team member alias must be alphanumeric (no spaces)")
	}
	return nil
}

// isValidAlias checks that an alias contains only letters and numbers.
func isValidAlias(alias string) bool {
	if alias == "" {
		return true
	}
	for _, r := range alias {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

// MatchesIdentifier checks if the team member matches the given identifier
// by alias or name with the following priority:
// 1. Exact alias match (case-sensitive)
// 2. Exact name match (case-sensitive)
// 3. Case-insensitive alias match
// 4. Case-insensitive name match
func (t *TeamMember) MatchesIdentifier(identifier string) bool {
	if t.Alias != "" && t.Alias == identifier {
		return true
	}
	if t.Name == identifier {
		return true
	}
	if t.Alias != "" && strings.EqualFold(t.Alias, identifier) {
		return true
	}
	return strings.EqualFold(t.Name, identifier)
}

// DisplayName returns the display name including alias if present.
func (t *TeamMember) DisplayName() string {
	if t.Alias != "" {
		return fmt.Sprintf("%s (%s)", t.Name, t.Alias)
	}
	return t.Name
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
