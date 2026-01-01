package domain

import (
	"errors"
	"fmt"
	"regexp"
)

// projectKeyRegex validates that project keys are uppercase letters only.
var projectKeyRegex = regexp.MustCompile(`^[A-Z]+$`)

// Project represents a Jira project.
type Project struct {
	Key  string `yaml:"key" json:"key"`
	Name string `yaml:"name" json:"name"`
}

// Validate checks that the Project has valid data.
func (p *Project) Validate() error {
	if p.Key == "" {
		return errors.New("project key is required")
	}
	if !projectKeyRegex.MatchString(p.Key) {
		return errors.New("project key must be uppercase letters only")
	}
	if p.Name == "" {
		return errors.New("project name is required")
	}
	return nil
}

// String returns a formatted string representation.
func (p *Project) String() string {
	return fmt.Sprintf("%s - %s", p.Key, p.Name)
}
