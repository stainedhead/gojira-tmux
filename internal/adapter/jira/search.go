package jira

import (
	"fmt"
	"strings"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// JQLBuilder builds JQL queries from filter criteria.
type JQLBuilder struct {
	projects []domain.Project
	team     []domain.TeamMember
}

// NewJQLBuilder creates a new JQL builder.
func NewJQLBuilder(projects []domain.Project, team []domain.TeamMember) *JQLBuilder {
	return &JQLBuilder{
		projects: projects,
		team:     team,
	}
}

// Build constructs a JQL query from the filter.
func (b *JQLBuilder) Build(filter domain.IssueFilter) string {
	var conditions []string

	// Project filter
	conditions = append(conditions, b.buildProjectCondition(filter.Project))

	// Assignee filter
	if assigneeCondition := b.buildAssigneeCondition(filter.Assignee); assigneeCondition != "" {
		conditions = append(conditions, assigneeCondition)
	}

	// Status filter
	if statusCondition := b.buildStatusCondition(filter.Status); statusCondition != "" {
		conditions = append(conditions, statusCondition)
	}

	// Join conditions and add ORDER BY
	query := strings.Join(conditions, " AND ")
	return query + " ORDER BY updated DESC"
}

// buildProjectCondition builds the project filter condition.
func (b *JQLBuilder) buildProjectCondition(projectKey string) string {
	// If specific project selected (not empty and not "-All-")
	if projectKey != "" && projectKey != "-All-" {
		return fmt.Sprintf(`project = %s`, escapeJQL(projectKey))
	}

	// All configured projects
	if len(b.projects) == 1 {
		return fmt.Sprintf(`project IN (%s)`, escapeJQL(b.projects[0].Key))
	}

	keys := make([]string, len(b.projects))
	for i, p := range b.projects {
		keys[i] = escapeJQL(p.Key)
	}
	return fmt.Sprintf(`project IN (%s)`, strings.Join(keys, ", "))
}

// buildAssigneeCondition builds the assignee filter condition.
func (b *JQLBuilder) buildAssigneeCondition(assigneeName string) string {
	if assigneeName == "" || assigneeName == "-All-" {
		return ""
	}

	// Find email for the given name
	var email string
	for _, m := range b.team {
		if m.Name == assigneeName {
			email = m.Email
			break
		}
	}

	if email == "" {
		return ""
	}

	return fmt.Sprintf(`assignee = %s`, escapeJQL(email))
}

// buildStatusCondition builds the status filter condition.
func (b *JQLBuilder) buildStatusCondition(status string) string {
	if status == "" || status == "All" {
		return ""
	}

	jqlStatus := MapStatus(status)
	if jqlStatus == "" {
		return ""
	}

	return fmt.Sprintf(`status = %s`, escapeJQL(jqlStatus))
}

// MapStatus maps UI status to JQL status.
func MapStatus(uiStatus string) string {
	switch uiStatus {
	case "Open":
		return "Open"
	case "Ready":
		return "Ready for Development"
	case "In Test":
		return "In Test"
	case "Done":
		return "Done"
	case "All", "":
		return ""
	default:
		return uiStatus
	}
}

// escapeJQL escapes special characters in JQL values and wraps in quotes.
func escapeJQL(s string) string {
	// JQL reserved: + - & | ! ( ) { } [ ] ^ " ~ * ? \ /
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
	)
	return `"` + replacer.Replace(s) + `"`
}
