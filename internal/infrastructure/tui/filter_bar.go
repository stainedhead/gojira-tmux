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
	focus    FilterFocus
	focused  bool
	width    int

	// Options
	members  []string
	projects []string
	statuses []string

	// Current selections (indices)
	memberIdx  int
	projectIdx int
	statusIdx  int
}

// NewFilterBar creates a new filter bar.
func NewFilterBar(team []domain.TeamMember, projects []domain.Project) *FilterBar {
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

	// Status options (fixed)
	statuses := []string{"All", "Open", "Ready", "In Test", "Done"}

	return &FilterBar{
		focus:    FilterFocusMember,
		focused:  false,
		members:  members,
		projects: projectOpts,
		statuses: statuses,
	}
}

// Focus sets the filter bar as focused.
func (f *FilterBar) Focus() {
	f.focused = true
}

// Blur removes focus from the filter bar.
func (f *FilterBar) Blur() {
	f.focused = false
}

// Focused returns whether the filter bar is focused.
func (f *FilterBar) Focused() bool {
	return f.focused
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
		switch msg.String() {
		case "tab", "right", "l":
			f.nextFocus()
			return f, nil
		case "shift+tab", "left", "h":
			f.prevFocus()
			return f, nil
		case "up", "k":
			f.prevOption()
			return f, f.emitFilterChanged()
		case "down", "j":
			f.nextOption()
			return f, f.emitFilterChanged()
		case "enter":
			// Enter could expand dropdown or just cycle
			f.nextOption()
			return f, f.emitFilterChanged()
		}
	}

	return f, nil
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

// View renders the filter bar.
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

	// Helper to render a dropdown
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

	// Render all dropdowns
	memberDropdown := renderDropdown("Member", f.members[f.memberIdx], f.focus == FilterFocusMember)
	projectDropdown := renderDropdown("Project", f.projects[f.projectIdx], f.focus == FilterFocusProject)
	statusDropdown := renderDropdown("Status", f.statuses[f.statusIdx], f.focus == FilterFocusStatus)

	// Join with spacing
	b.WriteString(memberDropdown)
	b.WriteString("   ")
	b.WriteString(projectDropdown)
	b.WriteString("   ")
	b.WriteString(statusDropdown)

	// Add help text if focused
	if f.focused {
		b.WriteString("   ")
		b.WriteString(Styles.Muted.Render("(←/→ switch, ↑/↓ change)"))
	}

	return Styles.FilterBar.Render(b.String())
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
