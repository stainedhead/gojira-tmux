package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// TicketsTable displays a table of Jira issues.
type TicketsTable struct {
	table    table.Model
	issues   []domain.Issue
	selected int
	width    int
	height   int
}

// NewTicketsTable creates a new tickets table.
func NewTicketsTable() *TicketsTable {
	columns := []table.Column{
		{Title: "!", Width: 1},
		{Title: "Key", Width: 12},
		{Title: "Summary", Width: 50},
		{Title: "Status", Width: 15},
		{Title: "Priority", Width: 10},
		{Title: "Assignee", Width: 20},
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

		rows[i] = table.Row{
			indicator,
			issue.Key,
			truncate(issue.Summary, 48),
			issue.Status,
			issue.Priority,
			truncate(assignee, 18),
		}
	}

	t.table.SetRows(rows)
}

// getAttentionIndicator returns the attention indicator for an issue.
func (t *TicketsTable) getAttentionIndicator(issue domain.Issue) string {
	attention := issue.NeedsAttention()
	switch attention {
	case domain.AttentionStale:
		return Styles.DotRed.String()
	case domain.AttentionNoDueDate:
		return Styles.DotYellow.String()
	default:
		return " "
	}
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

	// Adjust summary column width based on available space
	cols := t.table.Columns()
	if len(cols) > 2 {
		fixedWidth := 1 + 12 + 15 + 10 + 20 + 20 // indicator + key + status + priority + assignee + padding
		summaryWidth := width - fixedWidth
		if summaryWidth < 20 {
			summaryWidth = 20
		}
		if summaryWidth > 80 {
			summaryWidth = 80
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

	// Attention indicator
	attention := issue.NeedsAttention()
	switch attention {
	case domain.AttentionStale:
		b.WriteString(Styles.DotRed.String())
	case domain.AttentionNoDueDate:
		b.WriteString(Styles.DotYellow.String())
	default:
		b.WriteString(" ")
	}
	b.WriteString(" ")

	// Key
	key := fmt.Sprintf("%-12s", issue.Key)
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
