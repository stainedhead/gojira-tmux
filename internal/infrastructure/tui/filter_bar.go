package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// FilterFocus represents which filter is currently focused.
type FilterFocus int

const (
	// FilterFocusMember is focus on member dropdown.
	FilterFocusMember FilterFocus = iota
	// FilterFocusProject is focus on project dropdown.
	FilterFocusProject
	// FilterFocusStatus is focus on status dropdown.
	FilterFocusStatus
)

// FilterBar is a component for filtering issues.
type FilterBar struct {
	focus   FilterFocus
	focused bool
	width   int

	// Options
	members  []string
	projects []string
	statuses []string

	// Current committed selections (indices)
	memberIdx  int
	projectIdx int
	statusIdx  int

	// Dropdown overlay state
	dropdownOpen   bool
	dropdownCursor int
}

// defaultFallbackStatuses are used when no statuses are provided or fetched.
var defaultFallbackStatuses = []string{"To Do", "In Progress", "In Review", "Done"}

// NewFilterBar creates a new filter bar.
// statuses contains the Jira status names (e.g. from ListStatuses). "All" is prepended automatically.
// If statuses is empty, a built-in fallback list is used.
func NewFilterBar(team []domain.TeamMember, projects []domain.Project, statuses []string) *FilterBar {
	// Build member options using DisplayName for alias support
	members := []string{"-All-"}
	for _, m := range team {
		members = append(members, m.DisplayName())
	}

	// Build project options
	projectOpts := []string{"-All-"}
	for _, p := range projects {
		projectOpts = append(projectOpts, p.Key)
	}

	// Build status options: always start with "All", then the provided names
	statusOpts := []string{"All"}
	src := statuses
	if len(src) == 0 {
		src = defaultFallbackStatuses
	}
	for _, s := range src {
		if s != "All" {
			statusOpts = append(statusOpts, s)
		}
	}

	return &FilterBar{
		focus:    FilterFocusMember,
		focused:  false,
		members:  members,
		projects: projectOpts,
		statuses: statusOpts,
	}
}

// Focus sets the filter bar as focused.
func (f *FilterBar) Focus() {
	f.focused = true
}

// Blur removes focus from the filter bar.
func (f *FilterBar) Blur() {
	f.focused = false
	f.dropdownOpen = false
}

// Focused returns whether the filter bar is focused.
func (f *FilterBar) Focused() bool {
	return f.focused
}

// DropdownOpen reports whether a dropdown overlay is currently showing.
func (f *FilterBar) DropdownOpen() bool {
	return f.dropdownOpen
}

// SetWidth sets the width of the filter bar.
func (f *FilterBar) SetWidth(width int) {
	f.width = width
}

// GetFilter returns the current filter state.
func (f *FilterBar) GetFilter() domain.IssueFilter {
	return domain.IssueFilter{
		Assignee: f.members[f.memberIdx],
		Project:  f.projects[f.projectIdx],
		Status:   f.statuses[f.statusIdx],
	}
}

// SetFilter sets the filter bar state from a filter.
func (f *FilterBar) SetFilter(filter domain.IssueFilter) {
	// Find member index
	f.memberIdx = 0
	for i, m := range f.members {
		if m == filter.Assignee {
			f.memberIdx = i
			break
		}
	}

	// Find project index
	f.projectIdx = 0
	for i, p := range f.projects {
		if p == filter.Project {
			f.projectIdx = i
			break
		}
	}

	// Find status index
	f.statusIdx = 0
	for i, s := range f.statuses {
		if s == filter.Status {
			f.statusIdx = i
			break
		}
	}
}

// Update handles messages for the filter bar.
func (f *FilterBar) Update(msg tea.Msg) (*FilterBar, tea.Cmd) {
	if !f.focused {
		return f, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if f.dropdownOpen {
			return f.updateDropdown(msg)
		}
		return f.updateNormal(msg)
	}

	return f, nil
}

// updateDropdown handles keys when the dropdown overlay is open.
func (f *FilterBar) updateDropdown(msg tea.KeyMsg) (*FilterBar, tea.Cmd) {
	opts := f.getCurrentOptions()
	switch msg.String() {
	case "up", "k":
		f.dropdownCursor = (f.dropdownCursor + len(opts) - 1) % len(opts)
	case "down", "j":
		f.dropdownCursor = (f.dropdownCursor + 1) % len(opts)
	case "enter", " ":
		f.setCurrentIndex(f.dropdownCursor)
		f.dropdownOpen = false
		return f, f.emitFilterChanged()
	case "esc":
		f.dropdownOpen = false
	}
	return f, nil
}

// updateNormal handles keys when no dropdown is open.
func (f *FilterBar) updateNormal(msg tea.KeyMsg) (*FilterBar, tea.Cmd) {
	switch msg.String() {
	case "tab", "right", "l":
		f.nextFocus()
	case "shift+tab", "left", "h":
		f.prevFocus()
	case "up", "k":
		f.prevOption()
		return f, f.emitFilterChanged()
	case "down", "j":
		f.nextOption()
		return f, f.emitFilterChanged()
	case "enter", " ":
		f.dropdownCursor = f.getCurrentIndex()
		f.dropdownOpen = true
	}
	return f, nil
}

// getCurrentOptions returns options for the currently focused filter.
func (f *FilterBar) getCurrentOptions() []string {
	switch f.focus {
	case FilterFocusMember:
		return f.members
	case FilterFocusProject:
		return f.projects
	case FilterFocusStatus:
		return f.statuses
	}
	return nil
}

// getCurrentIndex returns the committed selection index for the focused filter.
func (f *FilterBar) getCurrentIndex() int {
	switch f.focus {
	case FilterFocusMember:
		return f.memberIdx
	case FilterFocusProject:
		return f.projectIdx
	case FilterFocusStatus:
		return f.statusIdx
	}
	return 0
}

// setCurrentIndex commits a selection index for the focused filter.
func (f *FilterBar) setCurrentIndex(idx int) {
	switch f.focus {
	case FilterFocusMember:
		f.memberIdx = idx
	case FilterFocusProject:
		f.projectIdx = idx
	case FilterFocusStatus:
		f.statusIdx = idx
	}
}

// getCurrentLabel returns the label for the focused filter.
func (f *FilterBar) getCurrentLabel() string {
	switch f.focus {
	case FilterFocusMember:
		return "Member"
	case FilterFocusProject:
		return "Project"
	case FilterFocusStatus:
		return "Status"
	}
	return ""
}

// nextFocus moves focus to next dropdown.
func (f *FilterBar) nextFocus() {
	f.focus = (f.focus + 1) % 3
}

// prevFocus moves focus to previous dropdown.
func (f *FilterBar) prevFocus() {
	f.focus = (f.focus + 2) % 3 // +2 is same as -1 mod 3
}

// nextOption cycles to next option in current dropdown.
func (f *FilterBar) nextOption() {
	switch f.focus {
	case FilterFocusMember:
		f.memberIdx = (f.memberIdx + 1) % len(f.members)
	case FilterFocusProject:
		f.projectIdx = (f.projectIdx + 1) % len(f.projects)
	case FilterFocusStatus:
		f.statusIdx = (f.statusIdx + 1) % len(f.statuses)
	}
}

// prevOption cycles to previous option in current dropdown.
func (f *FilterBar) prevOption() {
	switch f.focus {
	case FilterFocusMember:
		f.memberIdx = (f.memberIdx + len(f.members) - 1) % len(f.members)
	case FilterFocusProject:
		f.projectIdx = (f.projectIdx + len(f.projects) - 1) % len(f.projects)
	case FilterFocusStatus:
		f.statusIdx = (f.statusIdx + len(f.statuses) - 1) % len(f.statuses)
	}
}

// emitFilterChanged returns a command that emits FilterChangedMsg.
func (f *FilterBar) emitFilterChanged() tea.Cmd {
	filter := f.GetFilter()
	return func() tea.Msg {
		return FilterChangedMsg{Filter: filter}
	}
}

// View renders the filter bar, including any open dropdown overlay.
func (f *FilterBar) View() string {
	var b strings.Builder

	// Styles for dropdown
	labelStyle := Styles.FilterLabel
	valueStyle := Styles.FilterValue
	focusedStyle := lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true).
		Underline(true)
	bracketStyle := Styles.Muted

	// Helper to render a dropdown label+value
	renderDropdown := func(label, value string, isFocused bool) string {
		var result strings.Builder
		result.WriteString(labelStyle.Render(label + ": "))
		if isFocused && f.focused {
			result.WriteString(bracketStyle.Render("["))
			result.WriteString(focusedStyle.Render(value))
			result.WriteString(bracketStyle.Render("]"))
		} else {
			result.WriteString(valueStyle.Render(value))
		}
		return result.String()
	}

	memberDropdown := renderDropdown("Member", f.members[f.memberIdx], f.focus == FilterFocusMember)
	projectDropdown := renderDropdown("Project", f.projects[f.projectIdx], f.focus == FilterFocusProject)
	statusDropdown := renderDropdown("Status", f.statuses[f.statusIdx], f.focus == FilterFocusStatus)

	b.WriteString(memberDropdown)
	b.WriteString("   ")
	b.WriteString(projectDropdown)
	b.WriteString("   ")
	b.WriteString(statusDropdown)

	// Help text
	if f.focused {
		b.WriteString("   ")
		if f.dropdownOpen {
			b.WriteString(Styles.Muted.Render("(↑/↓ navigate, enter: select, esc: cancel)"))
		} else {
			b.WriteString(Styles.Muted.Render("(←/→ switch, ↑/↓ cycle, enter: open list)"))
		}
	}

	mainLine := Styles.FilterBar.Render(b.String())

	if f.dropdownOpen {
		return mainLine + "\n" + f.renderDropdownOverlay()
	}
	return mainLine
}

// renderDropdownOverlay renders the open dropdown as a bordered list below the filter bar.
func (f *FilterBar) renderDropdownOverlay() string {
	opts := f.getCurrentOptions()
	label := f.getCurrentLabel()

	highlightStyle := lipgloss.NewStyle().
		Background(Colors.Primary).
		Foreground(Colors.Background).
		Bold(true)
	normalStyle := lipgloss.NewStyle().
		Foreground(Colors.Foreground)

	rows := make([]string, len(opts))
	for i, opt := range opts {
		if i == f.dropdownCursor {
			rows[i] = highlightStyle.Render("▶ " + opt)
		} else {
			rows[i] = normalStyle.Render("  " + opt)
		}
	}

	content := strings.Join(rows, "\n")
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Primary).
		Padding(0, 1)

	title := Styles.FilterLabel.Render(label)
	return title + "\n" + boxStyle.Render(content)
}

// HasActiveFilters returns true if any filter is not set to "All".
func (f *FilterBar) HasActiveFilters() bool {
	return f.memberIdx != 0 || f.projectIdx != 0 || f.statusIdx != 0
}

// ClearFilters resets all filters to "All".
func (f *FilterBar) ClearFilters() tea.Cmd {
	f.memberIdx = 0
	f.projectIdx = 0
	f.statusIdx = 0
	return f.emitFilterChanged()
}
