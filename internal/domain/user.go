package domain

import (
	"errors"
)

// User represents the authenticated user.
type User struct {
	Email string `json:"email"`
}

// Validate checks that the User has valid data.
func (u *User) Validate() error {
	if u.Email == "" {
		return errors.New("user email is required")
	}
	if !isValidEmail(u.Email) {
		return errors.New("user email is invalid")
	}
	return nil
}
