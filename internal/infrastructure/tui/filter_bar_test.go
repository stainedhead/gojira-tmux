package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/stainedhead/gojira-tmux/internal/domain"
	"github.com/stainedhead/gojira-tmux/internal/infrastructure/tui"
)

var testStatuses = []string{"Open", "In Progress", "Done"}

func TestFilterBar_DisplayName_WithAlias(t *testing.T) {
	team := []domain.TeamMember{
		{Name: "John Anderson", Email: "john@example.com", Alias: "JohnA"},
		{Name: "Jane Smith", Email: "jane@example.com"},
	}
	projects := []domain.Project{
		{Key: "PROJ", Name: "Project"},
	}

	fb := tui.NewFilterBar(team, projects, testStatuses)
	view := fb.View()

	// Verify the filter bar renders without crashing
	if view == "" {
		t.Error("FilterBar.View() returned empty string")
	}
}

func TestFilterBar_GetFilter_DefaultsToAll(t *testing.T) {
	team := []domain.TeamMember{
		{Name: "John Doe", Email: "john@example.com"},
	}
	projects := []domain.Project{
		{Key: "PROJ", Name: "Project"},
	}

	fb := tui.NewFilterBar(team, projects, testStatuses)
	filter := fb.GetFilter()

	if filter.Assignee != "-All-" {
		t.Errorf("Default assignee = %q, want %q", filter.Assignee, "-All-")
	}
	if filter.Project != "-All-" {
		t.Errorf("Default project = %q, want %q", filter.Project, "-All-")
	}
	if filter.Status != "All" {
		t.Errorf("Default status = %q, want %q", filter.Status, "All")
	}
}

func TestFilterBar_BackwardCompatibility_NoAlias(t *testing.T) {
	team := []domain.TeamMember{
		{Name: "Jane Smith", Email: "jane@example.com"},
	}
	projects := []domain.Project{
		{Key: "PROJ", Name: "Project"},
	}

	fb := tui.NewFilterBar(team, projects, testStatuses)

	// Should not crash with members that have no alias
	view := fb.View()
	if view == "" {
		t.Error("FilterBar.View() returned empty string")
	}
}

func TestFilterBar_HasActiveFilters(t *testing.T) {
	team := []domain.TeamMember{
		{Name: "John Doe", Email: "john@example.com"},
	}
	projects := []domain.Project{
		{Key: "PROJ", Name: "Project"},
	}

	fb := tui.NewFilterBar(team, projects, testStatuses)

	if fb.HasActiveFilters() {
		t.Error("HasActiveFilters() = true for default filter bar")
	}
}

func TestFilterBar_ClearFilters(t *testing.T) {
	team := []domain.TeamMember{
		{Name: "John Doe", Email: "john@example.com"},
	}
	projects := []domain.Project{
		{Key: "PROJ", Name: "Project"},
	}

	fb := tui.NewFilterBar(team, projects, testStatuses)
	fb.ClearFilters()

	filter := fb.GetFilter()
	if filter.Assignee != "-All-" {
		t.Errorf("After clear, assignee = %q, want %q", filter.Assignee, "-All-")
	}
}

func TestFilterBar_SetFilter(t *testing.T) {
	team := []domain.TeamMember{
		{Name: "John Anderson", Email: "john@example.com", Alias: "JohnA"},
	}
	projects := []domain.Project{
		{Key: "PROJ", Name: "Project"},
	}

	fb := tui.NewFilterBar(team, projects, testStatuses)

	// Set a filter with the display name (which includes alias)
	fb.SetFilter(domain.IssueFilter{
		Assignee: "John Anderson (JohnA)",
		Project:  "PROJ",
		Status:   "Open",
	})

	filter := fb.GetFilter()
	if filter.Assignee != "John Anderson (JohnA)" {
		t.Errorf("After SetFilter, assignee = %q, want %q", filter.Assignee, "John Anderson (JohnA)")
	}
	if filter.Project != "PROJ" {
		t.Errorf("After SetFilter, project = %q, want %q", filter.Project, "PROJ")
	}
	if filter.Status != "Open" {
		t.Errorf("After SetFilter, status = %q, want %q", filter.Status, "Open")
	}
}

func TestFilterBar_DynamicStatuses(t *testing.T) {
	team := []domain.TeamMember{{Name: "Alice", Email: "alice@example.com"}}
	projects := []domain.Project{{Key: "PROJ", Name: "Project"}}
	customStatuses := []string{"Backlog", "In Sprint", "Review", "Released"}

	fb := tui.NewFilterBar(team, projects, customStatuses)

	// First status is always "All"
	filter := fb.GetFilter()
	if filter.Status != "All" {
		t.Errorf("Default status = %q, want All", filter.Status)
	}

	// SetFilter to a custom status
	fb.SetFilter(domain.IssueFilter{Status: "In Sprint"})
	filter = fb.GetFilter()
	if filter.Status != "In Sprint" {
		t.Errorf("After SetFilter, status = %q, want In Sprint", filter.Status)
	}
}

func TestFilterBar_EmptyStatuses_FallsBackToDefaults(t *testing.T) {
	team := []domain.TeamMember{{Name: "Alice", Email: "alice@example.com"}}
	projects := []domain.Project{{Key: "PROJ", Name: "Project"}}

	fb := tui.NewFilterBar(team, projects, nil)

	filter := fb.GetFilter()
	if filter.Status != "All" {
		t.Errorf("Default status = %q, want All", filter.Status)
	}
}

// --- Dropdown Tests ---

func TestFilterBar_Dropdown_ClosedByDefault(t *testing.T) {
	fb := tui.NewFilterBar(
		[]domain.TeamMember{{Name: "Alice", Email: "alice@example.com"}},
		[]domain.Project{{Key: "PROJ", Name: "Project"}},
		testStatuses,
	)

	if fb.DropdownOpen() {
		t.Error("Dropdown should be closed by default")
	}
}

func TestFilterBar_Dropdown_OpensOnEnter(t *testing.T) {
	fb := tui.NewFilterBar(
		[]domain.TeamMember{{Name: "Alice", Email: "alice@example.com"}},
		[]domain.Project{{Key: "PROJ", Name: "Project"}},
		testStatuses,
	)
	fb.Focus()

	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !fb.DropdownOpen() {
		t.Error("Dropdown should open on Enter")
	}
}

func TestFilterBar_Dropdown_ClosesOnEsc(t *testing.T) {
	fb := tui.NewFilterBar(
		[]domain.TeamMember{{Name: "Alice", Email: "alice@example.com"}},
		[]domain.Project{{Key: "PROJ", Name: "Project"}},
		testStatuses,
	)
	fb.Focus()

	// Open
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !fb.DropdownOpen() {
		t.Fatal("Dropdown should be open")
	}

	// Close with Esc
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if fb.DropdownOpen() {
		t.Error("Dropdown should close on Esc")
	}
}

func TestFilterBar_Dropdown_EscDoesNotChangeSelection(t *testing.T) {
	fb := tui.NewFilterBar(
		[]domain.TeamMember{{Name: "Alice", Email: "alice@example.com"}},
		[]domain.Project{{Key: "PROJ", Name: "Project"}},
		testStatuses,
	)
	fb.Focus()

	original := fb.GetFilter().Status // "All"

	// Open and move cursor down
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyDown})

	// Esc without confirming
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if fb.GetFilter().Status != original {
		t.Errorf("Status changed after Esc: got %q, want %q", fb.GetFilter().Status, original)
	}
}

func TestFilterBar_Dropdown_EnterConfirmsSelection(t *testing.T) {
	fb := tui.NewFilterBar(
		[]domain.TeamMember{{Name: "Alice", Email: "alice@example.com"}},
		[]domain.Project{{Key: "PROJ", Name: "Project"}},
		testStatuses,
	)
	fb.Focus()

	// Open dropdown (focused on Status by default — need to tab to Status)
	// Start on Member filter. Tab to Status.
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyTab})  // Member → Project
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyTab})  // Project → Status
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter}) // Open Status dropdown
	if !fb.DropdownOpen() {
		t.Fatal("Dropdown should be open on Status")
	}

	// Move cursor down once: All→Open
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyDown})
	// Confirm
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if fb.DropdownOpen() {
		t.Error("Dropdown should close after Enter confirms")
	}
	if fb.GetFilter().Status != "Open" {
		t.Errorf("Status = %q after confirm, want Open", fb.GetFilter().Status)
	}
}

func TestFilterBar_Dropdown_ViewContainsOptions(t *testing.T) {
	fb := tui.NewFilterBar(
		[]domain.TeamMember{{Name: "Alice", Email: "alice@example.com"}},
		[]domain.Project{{Key: "PROJ", Name: "Project"}},
		testStatuses,
	)
	fb.Focus()

	// Tab to Status, open dropdown
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyTab})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyTab})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter})

	view := fb.View()
	for _, s := range testStatuses {
		if !contains(view, s) {
			t.Errorf("Dropdown view missing option %q", s)
		}
	}
}

func TestFilterBar_Dropdown_HelpTextChanges(t *testing.T) {
	fb := tui.NewFilterBar(
		[]domain.TeamMember{{Name: "Alice", Email: "alice@example.com"}},
		[]domain.Project{{Key: "PROJ", Name: "Project"}},
		testStatuses,
	)
	fb.Focus()

	closedView := fb.View()
	if !contains(closedView, "enter") {
		t.Error("Closed filter bar should show 'enter' in help text")
	}

	// Open dropdown
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter})
	openView := fb.View()
	if !contains(openView, "esc") {
		t.Error("Open dropdown should show 'esc' in help text")
	}
}
