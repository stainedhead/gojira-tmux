package jira

import (
	"fmt"
	"strings"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// JQLBuilder builds JQL queries from filter criteria.
type JQLBuilder struct {
	projects      []domain.Project
	team          []domain.TeamMember
	statusFilters []domain.StatusFilter
}

// NewJQLBuilder creates a new JQL builder.
// statusFilters defines the named filter groups shown in the status dropdown.
// If nil or empty, domain.DefaultStatusFilters() is used.
func NewJQLBuilder(projects []domain.Project, team []domain.TeamMember, statusFilters []domain.StatusFilter) *JQLBuilder {
	sf := statusFilters
	if len(sf) == 0 {
		sf = domain.DefaultStatusFilters()
	}
	return &JQLBuilder{
		projects:      projects,
		team:          team,
		statusFilters: sf,
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
func (b *JQLBuilder) buildAssigneeCondition(identifier string) string {
	if identifier == "" || identifier == "-All-" {
		return ""
	}

	// Synthetic: match any team member
	if identifier == "-Team-" {
		if len(b.team) == 0 {
			return ""
		}
		emails := teamEmails(b.team)
		return fmt.Sprintf("assignee in (%s)", strings.Join(emails, ", "))
	}

	// Synthetic: exclude all team members
	if identifier == "-Not Team-" {
		if len(b.team) == 0 {
			return ""
		}
		emails := teamEmails(b.team)
		return fmt.Sprintf("assignee not in (%s)", strings.Join(emails, ", "))
	}

	// Find team member by name or alias
	var member *domain.TeamMember
	for i := range b.team {
		if b.team[i].MatchesIdentifier(identifier) {
			member = &b.team[i]
			break
		}
	}

	if member == nil {
		return ""
	}

	return fmt.Sprintf(`assignee = %s`, escapeJQL(member.Email))
}

// buildStatusCondition builds the status filter condition.
// If the status matches a named filter group, it expands to a multi-status clause.
// Otherwise the value is treated as a single Jira status name.
func (b *JQLBuilder) buildStatusCondition(status string) string {
	if status == "" || status == "All" {
		return ""
	}
	for _, sf := range b.statusFilters {
		if sf.Name == status {
			if len(sf.Statuses) == 0 {
				return ""
			}
			escaped := make([]string, len(sf.Statuses))
			for i, s := range sf.Statuses {
				escaped[i] = escapeJQL(s)
			}
			return fmt.Sprintf(`status in (%s)`, strings.Join(escaped, ", "))
		}
	}
	return fmt.Sprintf(`status in (%s)`, escapeJQL(status))
}

// teamEmails returns the JQL-escaped email addresses for all team members.
func teamEmails(team []domain.TeamMember) []string {
	emails := make([]string, len(team))
	for i, m := range team {
		emails[i] = escapeJQL(m.Email)
	}
	return emails
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
