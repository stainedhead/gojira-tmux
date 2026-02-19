package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// Fixed column content widths (excluding padding).
const (
	colWidthIndicator   = 6
	colWidthKey         = 11
	colWidthStatus      = 17
	colWidthPriority    = 8
	colWidthAssignee    = 18
	colWidthDueDate     = 10
	colWidthLastComment = 12
	colWidthLabels      = 22
	colMinSummary       = 20
	colMaxSummary       = 60
)

// colPad is the number of spaces on each side of a cell.
const colPad = 1

// tableNumCols is the total number of columns.
const tableNumCols = 9

// fixedColsWidth is the total content width of all non-summary columns.
// = 6+11+17+8+18+10+12+22 = 104
const fixedColsWidth = colWidthIndicator + colWidthKey + colWidthStatus + colWidthPriority +
	colWidthAssignee + colWidthDueDate + colWidthLastComment + colWidthLabels

// TicketsTable displays a table of Jira issues using a custom renderer.
//
// It replaces bubbles/table to work around its runewidth.Truncate limitation:
// runewidth counts ANSI escape-sequence characters as visible chars, which
// silently corrupts any cell value containing ANSI colour codes (e.g. the
// coloured indicator dots).  All cell sizing here is performed by
// lipgloss.Width / MaxWidth, which is ANSI-aware.
type TicketsTable struct {
	issues        []domain.Issue
	excludeLabels map[string]bool
	cursor        int
	offset        int
	width         int
	height        int
	focused       bool
	summaryWidth  int
}

// tableCol is a single column definition.
type tableCol struct {
	title string
	width int
}

// NewTicketsTable creates a new tickets table.
func NewTicketsTable() *TicketsTable {
	return &TicketsTable{summaryWidth: colMinSummary}
}

// cols returns the column definitions for the current summaryWidth.
func (t *TicketsTable) cols() []tableCol {
	return []tableCol{
		{"Issues", colWidthIndicator},
		{"Key", colWidthKey},
		{"Summary", t.summaryWidth},
		{"Status", colWidthStatus},
		{"Priority", colWidthPriority},
		{"Assignee", colWidthAssignee},
		{"Due Date", colWidthDueDate},
		{"Last Comment", colWidthLastComment},
		{"Labels", colWidthLabels},
	}
}

// dataRowsVisible returns the number of visible data rows
// (total height minus the 2-line header+separator block).
func (t *TicketsTable) dataRowsVisible() int {
	n := t.height - 2
	if n < 0 {
		return 0
	}
	return n
}

// clampOffset adjusts the scroll offset so the cursor stays in view.
func (t *TicketsTable) clampOffset() {
	dr := t.dataRowsVisible()
	if dr < 1 {
		dr = 1
	}
	if t.cursor < t.offset {
		t.offset = t.cursor
	}
	if t.cursor >= t.offset+dr {
		t.offset = t.cursor - dr + 1
	}
	if t.offset < 0 {
		t.offset = 0
	}
}

// SetIssues updates the table with new issues.
func (t *TicketsTable) SetIssues(issues []domain.Issue) {
	t.issues = issues
	if t.cursor >= len(issues) {
		t.cursor = max(len(issues)-1, 0)
	}
	t.offset = 0
}

// SelectedIssue returns the currently selected issue.
func (t *TicketsTable) SelectedIssue() *domain.Issue {
	if t.cursor >= 0 && t.cursor < len(t.issues) {
		return &t.issues[t.cursor]
	}
	return nil
}

// SetSize sets the table dimensions.
// height is the total lines allocated including the header + separator rows,
// matching the convention used by bubbles/table.SetHeight.
func (t *TicketsTable) SetSize(width, height int) {
	t.width = width
	t.height = height
	// Overhead = fixed col widths + padding (2×colPad per col) + separators (numCols−1).
	overhead := fixedColsWidth + tableNumCols*colPad*2 + (tableNumCols - 1)
	sumW := width - overhead
	if sumW < colMinSummary {
		sumW = colMinSummary
	}
	if sumW > colMaxSummary {
		sumW = colMaxSummary
	}
	t.summaryWidth = sumW
}

// Focus sets focus on the table.
func (t *TicketsTable) Focus() { t.focused = true }

// Blur removes focus from the table.
func (t *TicketsTable) Blur() { t.focused = false }

// Focused reports whether the table is focused.
func (t *TicketsTable) Focused() bool { return t.focused }

// SetExcludeLabels sets labels to hide from the table display.
func (t *TicketsTable) SetExcludeLabels(labels []string) {
	t.excludeLabels = make(map[string]bool, len(labels))
	for _, l := range labels {
		t.excludeLabels[l] = true
	}
}

// Update handles keyboard navigation messages.
func (t *TicketsTable) Update(msg tea.Msg) (*TicketsTable, tea.Cmd) {
	if !t.focused {
		return t, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		n := len(t.issues)
		if n == 0 {
			return t, nil
		}
		dr := max(t.dataRowsVisible(), 1)
		switch msg.String() {
		case "up", "k":
			if t.cursor > 0 {
				t.cursor--
				t.clampOffset()
			}
		case "down", "j":
			if t.cursor < n-1 {
				t.cursor++
				t.clampOffset()
			}
		case "g", "home":
			t.cursor = 0
			t.offset = 0
		case "G", "end":
			t.cursor = n - 1
			t.clampOffset()
		case "b", "pgup":
			t.cursor = max(t.cursor-dr, 0)
			t.clampOffset()
		case "f", "pgdown", " ":
			t.cursor = min(t.cursor+dr, n-1)
			t.clampOffset()
		case "u", "ctrl+u":
			t.cursor = max(t.cursor-dr/2, 0)
			t.clampOffset()
		case "d", "ctrl+d":
			t.cursor = min(t.cursor+dr/2, n-1)
			t.clampOffset()
		}
	}
	return t, nil
}

// View renders the table with column separators.
func (t *TicketsTable) View() string {
	if t.height <= 0 || t.width <= 0 {
		return ""
	}
	cols := t.cols()
	dr := t.dataRowsVisible()
	mutedSep := Styles.Muted.Render("│")

	var b strings.Builder

	// ── Header ──────────────────────────────────────────────────────────────────
	hParts := make([]string, len(cols))
	for i, col := range cols {
		rendered := lipgloss.NewStyle().
			Width(col.width).MaxWidth(col.width).
			Bold(true).Foreground(Colors.Primary).
			Render(col.title)
		hParts[i] = " " + rendered + " "
	}
	b.WriteString(strings.Join(hParts, mutedSep))
	b.WriteString("\n")

	// ── Header separator ────────────────────────────────────────────────────────
	sParts := make([]string, len(cols))
	for i, col := range cols {
		sParts[i] = strings.Repeat("─", col.width+colPad*2)
	}
	b.WriteString(Styles.Muted.Render(strings.Join(sParts, "┼")))

	// ── Data rows ───────────────────────────────────────────────────────────────
	end := t.offset + dr
	if end > len(t.issues) {
		end = len(t.issues)
	}
	for i := t.offset; i < end; i++ {
		b.WriteString("\n")
		b.WriteString(t.renderIssueRow(cols, i, mutedSep))
	}

	return b.String()
}

// renderIssueRow renders a single data row.
func (t *TicketsTable) renderIssueRow(cols []tableCol, idx int, sep string) string {
	issue := t.issues[idx]
	isSelected := idx == t.cursor

	assignee := ""
	if issue.Assignee != nil {
		assignee = issue.Assignee.Name
	}
	labels := filterExcludedLabels(issue.Labels, t.excludeLabels)

	values := []string{
		t.getAttentionIndicator(issue, isSelected),
		issue.Key,
		truncate(issue.Summary, t.summaryWidth),
		truncate(issue.Status, colWidthStatus),
		issue.Priority,
		truncate(assignee, colWidthAssignee),
		formatDueDate(issue.DueDate),
		formatRelativeTime(issue.LastCommentAt()),
		formatLabels(labels, colWidthLabels),
	}

	parts := make([]string, len(cols))
	for i, col := range cols {
		rendered := lipgloss.NewStyle().Width(col.width).MaxWidth(col.width).Render(values[i])
		parts[i] = " " + rendered + " "
	}

	// For the selected row use a plain │ so the selected background fills the
	// entire row (a muted-coloured │ would reset the background mid-row).
	if isSelected {
		return Styles.TableSelected.Render(strings.Join(parts, "│"))
	}
	return strings.Join(parts, sep)
}

// getAttentionIndicator returns three indicator circles for an issue.
//
// Non-selected rows: ANSI-coloured circles — safe here because lipgloss
// Width/MaxWidth (used in renderIssueRow) is ANSI-aware, unlike the
// runewidth.Truncate call that bubbles/table used.
//
// Selected row: plain Unicode circles — the ANSI reset codes inside coloured
// dots would otherwise cancel the selected-row background colour mid-row.
//
// Position 1 (red    ●): Stale  — assignee has not commented in 14+ days.
// Position 2 (yellow ●): No Due Date.
// Position 3 (cyan   ●): Overdue — due date has passed.
func (t *TicketsTable) getAttentionIndicator(issue domain.Issue, isSelected bool) string {
	if isSelected {
		dot := func(active bool) string {
			if active {
				return "●"
			}
			return "○"
		}
		return dot(issue.HasStaleIndicator()) + " " +
			dot(issue.HasNoDueDateIndicator()) + " " +
			dot(issue.HasOverdueIndicator())
	}
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

// ─── Helper functions ──────────────────────────────────────────────────────────

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
