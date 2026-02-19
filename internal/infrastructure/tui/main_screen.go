package tui

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// MainFocus represents which component is currently focused.
type MainFocus int

const (
	// MainFocusFilter is focus on the filter bar.
	MainFocusFilter MainFocus = iota
	// MainFocusTable is focus on the tickets table.
	MainFocusTable
	// MainFocusProperties is focus on the properties panel.
	MainFocusProperties
	// MainFocusComments is focus on the comments panel.
	MainFocusComments
)

// MainScreen is the main ticket list view.
type MainScreen struct {
	jiraPort   domain.JiraPort
	configPort domain.ConfigPort
	cfg        *domain.Config
	user       *domain.User

	filterBar       *FilterBar
	table           *TicketsTable
	propertiesPanel *PropertiesPanel
	commentsPanel   *CommentsPanel
	focus           MainFocus
	issues          []domain.Issue
	filter          domain.IssueFilter
	selectedIssue   *domain.Issue
	showDetails     bool
	showLegend      bool
	loading         bool
	loadingDetails  bool
	err             error
	width           int
	height          int
	keys            KeyMap

	// allStatuses is the full instance-level status list, used when no project
	// is selected or as fallback when project-scoped status loading fails.
	allStatuses []string
}

// NewMainScreenModel creates a new main screen.
// cfg is the current application config (used to restore last filter and persist changes).
// statuses is the list of Jira status names to populate the status filter.
func NewMainScreenModel(jiraPort domain.JiraPort, configPort domain.ConfigPort, user *domain.User, cfg *domain.Config, statuses []string) *MainScreen {
	// Get team and projects from config
	var team []domain.TeamMember
	var projects []domain.Project
	if configPort != nil {
		team = configPort.GetTeamMembers()
		projects = configPort.GetProjects()
	}

	var statusFilters []domain.StatusFilter
	if cfg != nil {
		statusFilters = cfg.StatusFilters
	}
	filterBar := NewFilterBar(team, projects, statuses, statusFilters)
	table := NewTicketsTable()
	propertiesPanel := NewPropertiesPanel()
	commentsPanel := NewCommentsPanel()

	// Apply label exclusions from config
	if cfg != nil && len(cfg.ExcludeLabels) > 0 {
		table.SetExcludeLabels(cfg.ExcludeLabels)
	}

	// Restore last saved filter if available
	if cfg != nil {
		saved := cfg.LastFilter
		if saved.Assignee != "" || saved.Project != "" || saved.Status != "" {
			filterBar.SetFilter(domain.IssueFilter{
				Assignee: saved.Assignee,
				Project:  saved.Project,
				Status:   saved.Status,
			})
		}
	}

	// Start with table focused
	filterBar.Blur()
	table.Focus()
	propertiesPanel.Blur()
	commentsPanel.Blur()

	// Capture the restored filter so initial loadIssues uses it
	initialFilter := filterBar.GetFilter()

	return &MainScreen{
		jiraPort:        jiraPort,
		configPort:      configPort,
		cfg:             cfg,
		user:            user,
		filterBar:       filterBar,
		table:           table,
		propertiesPanel: propertiesPanel,
		commentsPanel:   commentsPanel,
		focus:           MainFocusTable,
		filter:          initialFilter,
		keys:            DefaultKeyMap(),
		allStatuses:     statuses,
	}
}

// Init initializes the main screen and loads issues.
func (s *MainScreen) Init() tea.Cmd {
	return s.loadIssues()
}

// loadIssues fetches issues from Jira.
func (s *MainScreen) loadIssues() tea.Cmd {
	s.loading = true
	s.err = nil

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		issues, err := s.jiraPort.SearchIssues(ctx, s.filter)
		if err != nil {
			return issuesLoadErrorMsg{err: err}
		}

		return issuesLoadedInternalMsg{issues: issues}
	}
}

// loadIssueDetails fetches full details for an issue.
func (s *MainScreen) loadIssueDetails(key string) tea.Cmd {
	s.loadingDetails = true

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		issue, err := s.jiraPort.GetIssue(ctx, key)
		if err != nil {
			return issueDetailsErrorMsg{err: err}
		}

		// Fetch comments if not included
		if len(issue.Comments) == 0 {
			comments, err := s.jiraPort.GetIssueComments(ctx, key)
			if err == nil {
				issue.Comments = comments
			}
		}

		return issueDetailsLoadedInternalMsg{issue: issue}
	}
}

// loadProjectStatuses fetches statuses valid for a specific project.
// On failure it falls back to allStatuses so the filter bar stays populated.
func (s *MainScreen) loadProjectStatuses(projectKey string) tea.Cmd {
	allStatuses := s.allStatuses
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		statuses, err := s.jiraPort.ListProjectStatuses(ctx, projectKey)
		if err != nil || len(statuses) == 0 {
			return projectStatusesLoadedMsg{statuses: allStatuses}
		}
		return projectStatusesLoadedMsg{statuses: statuses}
	}
}

// Internal messages
type issuesLoadedInternalMsg struct {
	issues []domain.Issue
}

type issuesLoadErrorMsg struct {
	err error
}

type issueDetailsLoadedInternalMsg struct {
	issue *domain.Issue
}

type issueDetailsErrorMsg struct {
	err error
}

type projectStatusesLoadedMsg struct {
	statuses []string
}

// Update handles messages for the main screen.
func (s *MainScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.updateComponentSizes()
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "?":
			s.showLegend = !s.showLegend
			return s, nil
		case "r":
			if s.showLegend {
				return s, nil
			}
			return s, s.loadIssues()
		case "q", "ctrl+c":
			return s, tea.Quit
		case "f":
			if s.showLegend {
				return s, nil
			}
			// Toggle focus to filter bar
			s.setFocus(MainFocusFilter)
			return s, nil
		case "tab":
			// Absorb tab while filter dropdown is open
			if s.focus == MainFocusFilter && s.filterBar.DropdownOpen() {
				return s, nil
			}
			s.cycleFocus()
			return s, nil
		case "shift+tab":
			if s.focus == MainFocusFilter && s.filterBar.DropdownOpen() {
				return s, nil
			}
			s.cycleFocusReverse()
			return s, nil
		case "esc":
			// Escape closes the legend popup first
			if s.showLegend {
				s.showLegend = false
				return s, nil
			}
			// Escape closes the detail view (from any sub-panel or the table itself)
			if s.showDetails {
				s.showDetails = false
				s.selectedIssue = nil
				s.setFocus(MainFocusTable)
				s.updateComponentSizes()
				return s, nil
			}
			if s.focus == MainFocusFilter {
				if s.filterBar.DropdownOpen() {
					// Route esc to filter bar to close the dropdown
					s.filterBar, _ = s.filterBar.Update(msg)
					return s, nil
				}
				s.setFocus(MainFocusTable)
				return s, nil
			}
		case "enter":
			// Enter on table selects issue and shows details
			if s.focus == MainFocusTable {
				if issue := s.table.SelectedIssue(); issue != nil {
					s.showDetails = true
					s.selectedIssue = issue
					s.propertiesPanel.SetIssue(issue)
					s.commentsPanel.SetComments(issue.Comments)
					s.updateComponentSizes()
					// Load full details
					return s, s.loadIssueDetails(issue.Key)
				}
			}
		}

	case issuesLoadedInternalMsg:
		s.loading = false
		s.issues = msg.issues
		s.table.SetIssues(msg.issues)
		return s, nil

	case issuesLoadErrorMsg:
		s.loading = false
		s.err = msg.err
		return s, nil

	case issueDetailsLoadedInternalMsg:
		s.loadingDetails = false
		if msg.issue != nil {
			s.selectedIssue = msg.issue
			s.propertiesPanel.SetIssue(msg.issue)
			s.commentsPanel.SetComments(msg.issue.Comments)
		}
		return s, nil

	case issueDetailsErrorMsg:
		s.loadingDetails = false
		// Keep showing the basic issue info even if details failed
		return s, nil

	case FilterChangedMsg:
		projectChanged := s.filter.Project != msg.Filter.Project
		s.filter = msg.Filter
		cmds := []tea.Cmd{s.loadIssues(), s.saveFilter(msg.Filter)}
		if projectChanged {
			if msg.Filter.Project == "" || msg.Filter.Project == "-All-" {
				s.filterBar.SetStatuses(s.allStatuses)
			} else {
				cmds = append(cmds, s.loadProjectStatuses(msg.Filter.Project))
			}
		}
		return s, tea.Batch(cmds...)

	case projectStatusesLoadedMsg:
		s.filterBar.SetStatuses(msg.statuses)
		return s, nil
	}

	// Route to focused component
	var cmd tea.Cmd
	switch s.focus {
	case MainFocusFilter:
		s.filterBar, cmd = s.filterBar.Update(msg)
		// Recalculate sizes whenever the dropdown opens/closes to keep total
		// rendered height within the terminal bounds.
		s.updateComponentSizes()
	case MainFocusTable:
		// Issue 1: live detail panel update — capture cursor before routing.
		var prevKey string
		if s.showDetails {
			if prev := s.table.SelectedIssue(); prev != nil {
				prevKey = prev.Key
			}
		}
		s.table, cmd = s.table.Update(msg)
		// If the cursor moved while details are open, update the panels.
		if s.showDetails {
			if curr := s.table.SelectedIssue(); curr != nil && curr.Key != prevKey {
				s.selectedIssue = curr
				s.propertiesPanel.SetIssue(curr)
				s.commentsPanel.SetComments(curr.Comments)
				cmd = tea.Batch(cmd, s.loadIssueDetails(curr.Key))
			}
		}
	case MainFocusProperties:
		s.propertiesPanel, cmd = s.propertiesPanel.Update(msg)
	case MainFocusComments:
		s.commentsPanel, cmd = s.commentsPanel.Update(msg)
	}
	return s, cmd
}

// maxTableRows is the maximum number of data rows shown in the tickets table
// when the detail panels are visible. Capping the table height guarantees
// predictable vertical space for the panels below.
const maxTableRows = 12

// updateComponentSizes updates sizes of all components based on current layout.
func (s *MainScreen) updateComponentSizes() {
	contentWidth := max(s.width-4, 0)
	s.filterBar.SetWidth(contentWidth)

	// Reserve vertical space for any open dropdown overlay.
	filterExtra := s.filterBar.ExtraHeight()

	if s.showDetails {
		// Stacked layout: table fills full width at top (capped to maxTableRows),
		// detail panels side-by-side below the table.
		//
		// Vertical accounting (App padding 1top+1bot = 2, captured in s.height):
		//   title+blank(2) + filterbar+blank(2+filterExtra) + count+blank(2)
		//   + table(maxTableRows+2) + gap(1) + panels(detailHeight)
		//   + newline-after-panels(1) + footer-blank+help(2)
		// Total = 12 + filterExtra + maxTableRows + detailHeight  → solve for detailHeight
		detailHeight := max(s.height-12-filterExtra-maxTableRows, 4)

		s.table.SetSize(contentWidth, maxTableRows)

		// Split panels 50/50 horizontally with a 1-char gap between them.
		leftWidth := (contentWidth - 1) / 2
		rightWidth := contentWidth - leftWidth - 1
		if leftWidth < 10 {
			leftWidth = 10
		}
		if rightWidth < 10 {
			rightWidth = 10
		}
		s.propertiesPanel.SetSize(leftWidth, detailHeight)
		s.commentsPanel.SetSize(rightWidth, detailHeight)
	} else {
		// Full-width table fills all available vertical space.
		s.table.SetSize(contentWidth, max(s.height-10-filterExtra, 5))
	}
}

// setFocus sets focus to the specified component.
func (s *MainScreen) setFocus(focus MainFocus) {
	s.focus = focus

	// Blur all components
	s.filterBar.Blur()
	s.table.Blur()
	s.propertiesPanel.Blur()
	s.commentsPanel.Blur()

	// Focus the target component
	switch focus {
	case MainFocusFilter:
		s.filterBar.Focus()
	case MainFocusTable:
		s.table.Focus()
	case MainFocusProperties:
		s.propertiesPanel.Focus()
	case MainFocusComments:
		s.commentsPanel.Focus()
	}
}

// cycleFocus moves focus to the next component.
func (s *MainScreen) cycleFocus() {
	if s.showDetails {
		// With details: Filter -> Table -> Properties -> Comments -> Filter
		switch s.focus {
		case MainFocusFilter:
			s.setFocus(MainFocusTable)
		case MainFocusTable:
			s.setFocus(MainFocusProperties)
		case MainFocusProperties:
			s.setFocus(MainFocusComments)
		case MainFocusComments:
			s.setFocus(MainFocusFilter)
		}
	} else {
		// Without details: Filter -> Table -> Filter
		switch s.focus {
		case MainFocusFilter:
			s.setFocus(MainFocusTable)
		case MainFocusTable:
			s.setFocus(MainFocusFilter)
		}
	}
}

// cycleFocusReverse moves focus to the previous component.
func (s *MainScreen) cycleFocusReverse() {
	if s.showDetails {
		// Reverse: Comments -> Properties -> Table -> Filter -> Comments
		switch s.focus {
		case MainFocusFilter:
			s.setFocus(MainFocusComments)
		case MainFocusTable:
			s.setFocus(MainFocusFilter)
		case MainFocusProperties:
			s.setFocus(MainFocusTable)
		case MainFocusComments:
			s.setFocus(MainFocusProperties)
		}
	} else {
		s.cycleFocus()
	}
}

// View renders the main screen.
func (s *MainScreen) View() string {
	var b strings.Builder

	// Header
	s.renderHeader(&b)

	// Filter bar
	s.renderFilterBar(&b)

	// Main content
	if s.showLegend {
		s.renderLegend(&b)
	} else if s.loading {
		s.renderLoading(&b)
	} else if s.err != nil {
		s.renderError(&b)
	} else if len(s.issues) == 0 {
		s.renderEmpty(&b)
	} else if s.showDetails {
		s.renderSplitView(&b)
	} else {
		s.renderTable(&b)
	}

	// Footer
	s.renderFooter(&b)

	return Styles.App.Render(b.String())
}

func (s *MainScreen) renderHeader(b *strings.Builder) {
	title := Styles.Title.Render("📋 Jira Tickets")
	b.WriteString(title)

	if s.user != nil {
		userInfo := Styles.Muted.Render(" (" + s.user.Email + ")")
		b.WriteString(userInfo)
	}
	b.WriteString("\n\n")
}

func (s *MainScreen) renderFilterBar(b *strings.Builder) {
	// Render the interactive filter bar
	b.WriteString(s.filterBar.View())
	b.WriteString("\n\n")
}

func (s *MainScreen) renderLoading(b *strings.Builder) {
	loading := Styles.Muted.Render("Loading issues...")
	b.WriteString(loading)
	b.WriteString("\n")
}

func (s *MainScreen) renderError(b *strings.Builder) {
	errMsg := Styles.Error.Render("Error loading issues: " + s.err.Error())
	b.WriteString(errMsg)
	b.WriteString("\n\n")
	b.WriteString(Styles.Muted.Render("Press 'r' to retry"))
	b.WriteString("\n")
}

func (s *MainScreen) renderEmpty(b *strings.Builder) {
	empty := Styles.Muted.Render("No issues found matching the current filters.")
	b.WriteString(empty)
	b.WriteString("\n\n")
	b.WriteString(Styles.Muted.Render("Press 'r' to refresh"))
	b.WriteString("\n")
}

func (s *MainScreen) renderTable(b *strings.Builder) {
	// Issue count
	count := Styles.Muted.Render(formatCount(len(s.issues)))
	b.WriteString(count)
	b.WriteString("\n\n")

	// Table
	b.WriteString(s.table.View())
	b.WriteString("\n")
}

func (s *MainScreen) renderSplitView(b *strings.Builder) {
	// Issue count
	count := Styles.Muted.Render(formatCount(len(s.issues)))
	b.WriteString(count)
	b.WriteString("\n\n")

	// Full-width table at top
	b.WriteString(s.table.View())
	b.WriteString("\n")

	// Detail panels side-by-side below the table.
	// Properties (field details) on the left, Comments on the right.
	detailRow := lipgloss.JoinHorizontal(lipgloss.Top,
		s.propertiesPanel.View(),
		" ",
		s.commentsPanel.View(),
	)
	b.WriteString(detailRow)
	b.WriteString("\n")
}

func (s *MainScreen) renderLegend(b *strings.Builder) {
	title := Styles.PanelTitle.Render("Indicator Legend")

	row := func(dotStyle lipgloss.Style, name, desc string) string {
		return lipgloss.JoinHorizontal(lipgloss.Top,
			dotStyle.String(),
			"  ",
			Styles.HelpKey.Render(name),
		) + "\n    " + Styles.HelpDesc.Render(desc)
	}

	content := strings.Join([]string{
		title,
		"",
		row(Styles.DotRed, "Stale", "Assignee has not commented in 14+ days"),
		"",
		row(Styles.DotYellow, "No Due Date", "Issue has no due date set"),
		"",
		row(Styles.DotCyan, "Overdue", "Due date has passed"),
		"",
		Styles.DotEmpty.String() + "  " + Styles.HelpDesc.Render("Indicator not triggered"),
		"",
		Styles.Muted.Render("Press ? or esc to close"),
	}, "\n")

	b.WriteString(Styles.Panel.Render(content))
	b.WriteString("\n")
}

func (s *MainScreen) renderFooter(b *strings.Builder) {
	var help string
	if s.showDetails {
		help = lipgloss.JoinHorizontal(lipgloss.Top,
			Styles.HelpKey.Render("tab"),
			Styles.HelpDesc.Render(" switch panel  "),
			Styles.HelpKey.Render("esc"),
			Styles.HelpDesc.Render(" close details  "),
			Styles.HelpKey.Render("r"),
			Styles.HelpDesc.Render(" refresh  "),
			Styles.HelpKey.Render("q"),
			Styles.HelpDesc.Render(" quit"),
		)
	} else {
		help = lipgloss.JoinHorizontal(lipgloss.Top,
			Styles.HelpKey.Render("↑/↓"),
			Styles.HelpDesc.Render(" navigate  "),
			Styles.HelpKey.Render("enter"),
			Styles.HelpDesc.Render(" view details  "),
			Styles.HelpKey.Render("tab"),
			Styles.HelpDesc.Render(" switch focus  "),
			Styles.HelpKey.Render("f"),
			Styles.HelpDesc.Render(" filter  "),
			Styles.HelpKey.Render("?"),
			Styles.HelpDesc.Render(" legend  "),
			Styles.HelpKey.Render("r"),
			Styles.HelpDesc.Render(" refresh  "),
			Styles.HelpKey.Render("q"),
			Styles.HelpDesc.Render(" quit"),
		)
	}
	b.WriteString("\n")
	b.WriteString(help)
}

// saveFilter persists the current filter state to the config file non-blocking.
func (s *MainScreen) saveFilter(filter domain.IssueFilter) tea.Cmd {
	if s.configPort == nil || s.cfg == nil {
		return nil
	}
	s.cfg.LastFilter = domain.FilterState{
		Assignee: filter.Assignee,
		Project:  filter.Project,
		Status:   filter.Status,
	}
	cfg := s.cfg
	return func() tea.Msg {
		_ = s.configPort.Save(cfg) // silently ignore errors - non-critical
		return nil
	}
}

func formatCount(n int) string {
	if n == 1 {
		return "1 issue"
	}
	return intToString(n) + " issues"
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
