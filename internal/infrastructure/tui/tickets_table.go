package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// TicketsTable displays a table of Jira issues.
type TicketsTable struct {
	table         table.Model
	issues        []domain.Issue
	excludeLabels map[string]bool
	selected      int
	width         int
	height        int
}

// NewTicketsTable creates a new tickets table.
func NewTicketsTable() *TicketsTable {
	columns := []table.Column{
		{Title: "        ", Width: 8},
		{Title: "Key", Width: 15},
		{Title: "Summary", Width: 50},
		{Title: "Status", Width: 15},
		{Title: "Priority", Width: 8},
		{Title: "Assignee", Width: 18},
		{Title: "Due Date", Width: 10},
		{Title: "Last Comment", Width: 12},
		{Title: "Labels", Width: 18},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(Colors.Border).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(Colors.Foreground).
		Background(Colors.Primary).
		Bold(true)
	t.SetStyles(s)

	return &TicketsTable{
		table: t,
	}
}

// SetIssues updates the table with new issues.
func (t *TicketsTable) SetIssues(issues []domain.Issue) {
	t.issues = issues

	rows := make([]table.Row, len(issues))
	for i, issue := range issues {
		indicator := t.getAttentionIndicator(issue)
		assignee := ""
		if issue.Assignee != nil {
			assignee = issue.Assignee.Name
		}

		visibleLabels := filterExcludedLabels(issue.Labels, t.excludeLabels)
		rows[i] = table.Row{
			indicator,
			issue.Key,
			truncate(issue.Summary, 48),
			issue.Status,
			issue.Priority,
			truncate(assignee, 16),
			formatDueDate(issue.DueDate),
			formatRelativeTime(issue.LastCommentAt()),
			formatLabels(visibleLabels, 16),
		}
	}

	t.table.SetRows(rows)
}

// getAttentionIndicator returns three independent indicator circles for an issue.
// Each circle is filled (●) when its condition is active, empty (○) otherwise.
// Position 1 (red):    Stale — assignee inactive 14+ days
// Position 2 (yellow): No Due Date — due date not set
// Position 3 (cyan):   Overdue — due date is in the past
func (t *TicketsTable) getAttentionIndicator(issue domain.Issue) string {
	dot := func(active bool, style lipgloss.Style) string {
		if active {
			return style.String()
		}
		return Styles.DotEmpty.String()
	}
	return dot(issue.HasStaleIndicator(), Styles.DotRed) + " " +
		dot(issue.HasNoDueDateIndicator(), Styles.DotYellow) + " " +
		dot(issue.HasOverdueIndicator(), Styles.DotCyan)
}

// SelectedIssue returns the currently selected issue.
func (t *TicketsTable) SelectedIssue() *domain.Issue {
	idx := t.table.Cursor()
	if idx >= 0 && idx < len(t.issues) {
		return &t.issues[idx]
	}
	return nil
}

// SetSize sets the table size.
func (t *TicketsTable) SetSize(width, height int) {
	t.width = width
	t.height = height
	t.table.SetHeight(height)

	// Adjust summary column width based on available space.
	// Fixed cols: indicators + Key + Status + Priority + Assignee + DueDate + LastComment + Labels + padding
	cols := t.table.Columns()
	if len(cols) > 2 {
		fixedWidth := 8 + 15 + 15 + 8 + 18 + 10 + 12 + 18 + 20 // columns + padding
		summaryWidth := width - fixedWidth
		if summaryWidth < 20 {
			summaryWidth = 20
		}
		if summaryWidth > 60 {
			summaryWidth = 60
		}
		cols[2].Width = summaryWidth
		t.table.SetColumns(cols)
	}
}

// Update handles messages for the table.
func (t *TicketsTable) Update(msg tea.Msg) (*TicketsTable, tea.Cmd) {
	var cmd tea.Cmd
	t.table, cmd = t.table.Update(msg)
	return t, cmd
}

// View renders the table.
func (t *TicketsTable) View() string {
	return t.table.View()
}

// Focus sets focus to the table.
func (t *TicketsTable) Focus() {
	t.table.Focus()
}

// Blur removes focus from the table.
func (t *TicketsTable) Blur() {
	t.table.Blur()
}

// Focused returns whether the table is focused.
func (t *TicketsTable) Focused() bool {
	return t.table.Focused()
}

// SetExcludeLabels sets the label values that should be hidden from the table.
// Existing rows are rebuilt to apply the new exclusions.
func (t *TicketsTable) SetExcludeLabels(labels []string) {
	t.excludeLabels = make(map[string]bool, len(labels))
	for _, l := range labels {
		t.excludeLabels[l] = true
	}
	// Rebuild rows so the exclusion takes effect immediately.
	t.SetIssues(t.issues)
}

// filterExcludedLabels returns a new slice with any labels in the exclude set removed.
func filterExcludedLabels(labels []string, exclude map[string]bool) []string {
	if labels == nil {
		return nil
	}
	if len(exclude) == 0 {
		return labels
	}
	out := make([]string, 0, len(labels))
	for _, l := range labels {
		if !exclude[l] {
			out = append(out, l)
		}
	}
	return out
}

// formatDueDate formats a due date as "YYYY-MM-DD" or "none" if nil.
func formatDueDate(d *time.Time) string {
	if d == nil {
		return "none"
	}
	return d.Format("2006-01-02")
}

// formatRelativeTime formats a time as a human-readable relative string like "3d ago".
// Returns "never" for nil, "today" for times within the past 24 hours.
func formatRelativeTime(t *time.Time) string {
	if t == nil {
		return "never"
	}
	hours := int(time.Since(*t).Hours())
	if hours < 24 {
		return "today"
	}
	return fmt.Sprintf("%dd ago", hours/24)
}

// formatLabels formats a slice of labels as a truncated comma-separated string.
func formatLabels(labels []string, maxLen int) string {
	if len(labels) == 0 {
		return "none"
	}
	joined := strings.Join(labels, ", ")
	if len(joined) > maxLen {
		return joined[:maxLen-3] + "..."
	}
	return joined
}

// truncate truncates a string to the given length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// FormatIssueRow formats an issue as a table row string (for non-table display).
func FormatIssueRow(issue domain.Issue, width int) string {
	var b strings.Builder

	// Attention indicators — three independent circles
	dot := func(active bool, style lipgloss.Style) string {
		if active {
			return style.String()
		}
		return Styles.DotEmpty.String()
	}
	b.WriteString(dot(issue.HasStaleIndicator(), Styles.DotRed))
	b.WriteString(" ")
	b.WriteString(dot(issue.HasNoDueDateIndicator(), Styles.DotYellow))
	b.WriteString(" ")
	b.WriteString(dot(issue.HasOverdueIndicator(), Styles.DotCyan))
	b.WriteString(" ")

	// Key
	key := fmt.Sprintf("%-15s", issue.Key)
	b.WriteString(key)
	b.WriteString(" ")

	// Summary (truncated)
	summaryWidth := width - 50
	if summaryWidth < 20 {
		summaryWidth = 20
	}
	summary := truncate(issue.Summary, summaryWidth)
	b.WriteString(summary)

	return b.String()
}
