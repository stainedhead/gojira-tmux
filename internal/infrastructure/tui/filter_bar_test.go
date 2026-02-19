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

	fb := tui.NewFilterBar(team, projects, testStatuses, nil)
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

	fb := tui.NewFilterBar(team, projects, testStatuses, nil)
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

	fb := tui.NewFilterBar(team, projects, testStatuses, nil)

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

	fb := tui.NewFilterBar(team, projects, testStatuses, nil)

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

	fb := tui.NewFilterBar(team, projects, testStatuses, nil)
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

	fb := tui.NewFilterBar(team, projects, testStatuses, nil)

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

	fb := tui.NewFilterBar(team, projects, customStatuses, nil)

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

	fb := tui.NewFilterBar(team, projects, nil, nil)

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
		nil,
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
		nil,
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
		nil,
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
		nil,
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
		nil,
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

	// Move cursor down three times: All → -Open- → -Active- → Open
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyDown})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyDown})
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
		nil,
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
		nil,
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

// --- -Open- sentinel tests ---

func TestFilterBar_OpenSentinel_InStatusList(t *testing.T) {
	fb := tui.NewFilterBar(
		[]domain.TeamMember{{Name: "Alice", Email: "alice@example.com"}},
		[]domain.Project{{Key: "PROJ", Name: "Project"}},
		testStatuses,
		nil,
	)
	fb.Focus()

	// Tab to Status, open dropdown, check view
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyTab})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyTab})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter})

	view := fb.View()
	if !contains(view, "-Open-") {
		t.Error("Status dropdown should contain -Open- sentinel")
	}
}

func TestFilterBar_OpenSentinel_CanBeSelected(t *testing.T) {
	fb := tui.NewFilterBar(
		[]domain.TeamMember{{Name: "Alice", Email: "alice@example.com"}},
		[]domain.Project{{Key: "PROJ", Name: "Project"}},
		testStatuses,
		nil,
	)
	fb.Focus()

	// Tab to Status, open, move down once (All → -Open-), confirm
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyTab})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyTab})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyDown})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if fb.GetFilter().Status != "-Open-" {
		t.Errorf("Status = %q, want -Open-", fb.GetFilter().Status)
	}
}

// --- SetStatuses tests ---

func TestFilterBar_SetStatuses_KeepsValidSelection(t *testing.T) {
	fb := tui.NewFilterBar(
		[]domain.TeamMember{{Name: "Alice", Email: "alice@example.com"}},
		[]domain.Project{{Key: "PROJ", Name: "Project"}},
		[]string{"In Progress", "Done"},
		nil,
	)

	// Select "In Progress"
	fb.SetFilter(domain.IssueFilter{Status: "In Progress"})
	if fb.GetFilter().Status != "In Progress" {
		t.Fatalf("pre-condition: status = %q, want In Progress", fb.GetFilter().Status)
	}

	// Update statuses — "In Progress" is still in the new list
	fb.SetStatuses([]string{"To Do", "In Progress", "Closed"})
	if fb.GetFilter().Status != "In Progress" {
		t.Errorf("Status = %q after SetStatuses, want In Progress (preserved)", fb.GetFilter().Status)
	}
}

func TestFilterBar_SetStatuses_ResetsInvalidSelection(t *testing.T) {
	fb := tui.NewFilterBar(
		[]domain.TeamMember{{Name: "Alice", Email: "alice@example.com"}},
		[]domain.Project{{Key: "PROJ", Name: "Project"}},
		[]string{"In Progress", "Done"},
		nil,
	)

	// Select "Done"
	fb.SetFilter(domain.IssueFilter{Status: "Done"})

	// New statuses don't include "Done"
	fb.SetStatuses([]string{"To Do", "In Progress"})
	if fb.GetFilter().Status != "All" {
		t.Errorf("Status = %q after SetStatuses, want All (reset)", fb.GetFilter().Status)
	}
}

func TestFilterBar_SetStatuses_KeepsOpenSentinel(t *testing.T) {
	fb := tui.NewFilterBar(
		[]domain.TeamMember{{Name: "Alice", Email: "alice@example.com"}},
		[]domain.Project{{Key: "PROJ", Name: "Project"}},
		testStatuses,
		nil,
	)

	// Select -Open-
	fb.SetFilter(domain.IssueFilter{Status: "-Open-"})

	// Replace statuses — -Open- is always injected by buildStatusOptions
	fb.SetStatuses([]string{"To Do", "Done"})
	if fb.GetFilter().Status != "-Open-" {
		t.Errorf("Status = %q after SetStatuses, want -Open- (always present)", fb.GetFilter().Status)
	}
}

func TestFilterBar_SetStatuses_OpenSentinelAlwaysPresent(t *testing.T) {
	fb := tui.NewFilterBar(
		[]domain.TeamMember{{Name: "Alice", Email: "alice@example.com"}},
		[]domain.Project{{Key: "PROJ", Name: "Project"}},
		[]string{"To Do"},
		nil,
	)
	fb.SetStatuses([]string{"Backlog", "Released"})

	// The status list must always contain "All" and "-Open-"
	filter := fb.GetFilter()
	if filter.Status != "All" {
		t.Errorf("default status = %q, want All", filter.Status)
	}
	// Tab to Status, open dropdown, check -Open- is present
	fb.Focus()
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyTab})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyTab})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !contains(fb.View(), "-Open-") {
		t.Error("Status dropdown should always contain -Open-")
	}
}

// --- Custom status filter group tests ---

func TestFilterBar_CustomStatusFilters_AppearInDropdown(t *testing.T) {
	customFilters := []domain.StatusFilter{
		{Name: "-Mine-", Statuses: []string{"In Progress", "In Review"}},
		{Name: "-Blocked-", Statuses: []string{"On Hold", "Escalated"}},
	}
	fb := tui.NewFilterBar(
		[]domain.TeamMember{{Name: "Alice", Email: "alice@example.com"}},
		[]domain.Project{{Key: "PROJ", Name: "Project"}},
		testStatuses,
		customFilters,
	)
	fb.Focus()

	// Tab to Status, open dropdown
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyTab})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyTab})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view := fb.View()

	if !contains(view, "-Mine-") {
		t.Error("Status dropdown should contain custom group -Mine-")
	}
	if !contains(view, "-Blocked-") {
		t.Error("Status dropdown should contain custom group -Blocked-")
	}
	// Default sentinels should NOT appear when custom filters are provided
	if contains(view, "-Open-") {
		t.Error("Status dropdown should not contain -Open- when custom filters are defined")
	}
	if contains(view, "-Active-") {
		t.Error("Status dropdown should not contain -Active- when custom filters are defined")
	}
}

func TestFilterBar_CustomStatusFilters_CanBeSelected(t *testing.T) {
	customFilters := []domain.StatusFilter{
		{Name: "-Mine-", Statuses: []string{"In Progress", "In Review"}},
	}
	fb := tui.NewFilterBar(
		[]domain.TeamMember{{Name: "Alice", Email: "alice@example.com"}},
		[]domain.Project{{Key: "PROJ", Name: "Project"}},
		testStatuses,
		customFilters,
	)
	fb.Focus()

	// Tab to Status, open, move down once (All → -Mine-), confirm
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyTab})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyTab})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyDown})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if fb.GetFilter().Status != "-Mine-" {
		t.Errorf("Status = %q, want -Mine-", fb.GetFilter().Status)
	}
}

// --- -Team- / -Not Team- member tests ---

func TestFilterBar_TeamSentinels_PresentWithNonEmptyTeam(t *testing.T) {
	team := []domain.TeamMember{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}
	fb := tui.NewFilterBar(team, []domain.Project{{Key: "PROJ", Name: "Project"}}, testStatuses, nil)
	fb.Focus()

	// Open Member dropdown
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view := fb.View()

	if !contains(view, "-Team-") {
		t.Error("Member dropdown should contain -Team- when team is non-empty")
	}
	if !contains(view, "-Not Team-") {
		t.Error("Member dropdown should contain -Not Team- when team is non-empty")
	}
}

func TestFilterBar_TeamSentinels_AbsentWithEmptyTeam(t *testing.T) {
	fb := tui.NewFilterBar(nil, []domain.Project{{Key: "PROJ", Name: "Project"}}, testStatuses, nil)
	fb.Focus()

	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view := fb.View()

	if contains(view, "-Team-") {
		t.Error("Member dropdown should NOT contain -Team- when team is empty")
	}
	if contains(view, "-Not Team-") {
		t.Error("Member dropdown should NOT contain -Not Team- when team is empty")
	}
}

func TestFilterBar_TeamSentinel_CanBeSelected(t *testing.T) {
	team := []domain.TeamMember{
		{Name: "Alice", Email: "alice@example.com"},
	}
	fb := tui.NewFilterBar(team, []domain.Project{{Key: "PROJ", Name: "Project"}}, testStatuses, nil)
	fb.Focus()

	// Open dropdown, move down once (-All- → -Team-), confirm
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyDown})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if fb.GetFilter().Assignee != "-Team-" {
		t.Errorf("Assignee = %q, want -Team-", fb.GetFilter().Assignee)
	}
}

func TestFilterBar_NotTeamSentinel_CanBeSelected(t *testing.T) {
	team := []domain.TeamMember{
		{Name: "Alice", Email: "alice@example.com"},
	}
	fb := tui.NewFilterBar(team, []domain.Project{{Key: "PROJ", Name: "Project"}}, testStatuses, nil)
	fb.Focus()

	// Open dropdown, move down twice (-All- → -Team- → -Not Team-), confirm
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyDown})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyDown})
	fb, _ = fb.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if fb.GetFilter().Assignee != "-Not Team-" {
		t.Errorf("Assignee = %q, want -Not Team-", fb.GetFilter().Assignee)
	}
}
