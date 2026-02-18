package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// PropertiesPanel displays detailed issue properties.
type PropertiesPanel struct {
	issue   *domain.Issue
	focused bool
	width   int
	height  int
	scrollY int
}

// NewPropertiesPanel creates a new properties panel.
func NewPropertiesPanel() *PropertiesPanel {
	return &PropertiesPanel{}
}

// SetIssue sets the issue to display.
func (p *PropertiesPanel) SetIssue(issue *domain.Issue) {
	p.issue = issue
	p.scrollY = 0
}

// Focus sets the panel as focused.
func (p *PropertiesPanel) Focus() {
	p.focused = true
}

// Blur removes focus from the panel.
func (p *PropertiesPanel) Blur() {
	p.focused = false
}

// Focused returns whether the panel is focused.
func (p *PropertiesPanel) Focused() bool {
	return p.focused
}

// SetSize sets the panel dimensions.
func (p *PropertiesPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// Update handles messages for the panel.
func (p *PropertiesPanel) Update(msg tea.Msg) (*PropertiesPanel, tea.Cmd) {
	if !p.focused {
		return p, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if p.scrollY > 0 {
				p.scrollY--
			}
		case "down", "j":
			p.scrollY++
		}
	}

	return p, nil
}

// View renders the properties panel clamped to its set height.
func (p *PropertiesPanel) View() string {
	panelStyle := Styles.Panel
	if p.focused {
		panelStyle = Styles.FocusedPanel
	}

	// Build every content line up-front so we can apply viewport clamping.
	var lines []string
	lines = append(lines, Styles.PanelTitle.Render("Properties"))

	if p.issue == nil {
		lines = append(lines, Styles.Muted.Render("Select an issue to view details"))
	} else {
		labelStyle := Styles.FilterLabel.Width(12)
		for _, prop := range []struct{ label, value string }{
			{"Key", p.issue.Key},
			{"Status", p.issue.Status},
			{"Priority", p.issue.Priority},
			{"Reporter", p.getReporter()},
			{"Assignee", p.getAssignee()},
			{"Created", p.formatTime(p.issue.Created)},
			{"Updated", p.formatTime(p.issue.Updated)},
			{"Due Date", p.getDueDate()},
			{"Sprint", p.issue.Sprint},
			{"Epic", p.issue.Epic},
			{"Story Points", p.getStoryPoints()},
			{"Labels", p.getLabels()},
		} {
			if prop.value == "" {
				continue
			}
			lines = append(lines,
				labelStyle.Render(prop.label+":")+" "+Styles.Paragraph.Render(prop.value),
			)
		}

		if p.issue.Description != "" {
			lines = append(lines, "")
			lines = append(lines, Styles.FilterLabel.Render("Description:"))
			wrapWidth := p.width - 4
			if wrapWidth < 1 {
				wrapWidth = 1
			}
			for _, dl := range strings.Split(wrapText(p.issue.Description, wrapWidth), "\n") {
				lines = append(lines, Styles.Muted.Render(dl))
			}
		}
	}

	return p.renderViewport(panelStyle, lines)
}

// renderViewport clamps lines to the visible height and renders the panel.
// The panel always occupies exactly p.height terminal lines (outer, including borders).
func (p *PropertiesPanel) renderViewport(panelStyle lipgloss.Style, lines []string) string {
	// viewHeight is the number of content lines that fit inside the border.
	// Height() in lipgloss v1 is the outer total; borders consume 2 lines.
	viewHeight := max(p.height-2, 1)

	// Clamp scroll position.
	maxScroll := max(len(lines)-viewHeight, 0)
	if p.scrollY > maxScroll {
		p.scrollY = maxScroll
	}

	end := min(p.scrollY+viewHeight, len(lines))
	visible := strings.Join(lines[p.scrollY:end], "\n")

	return panelStyle.Width(max(p.width, 0)).Height(max(p.height, 0)).Render(visible)
}

func (p *PropertiesPanel) getReporter() string {
	if p.issue.Reporter != nil {
		return p.issue.Reporter.Name
	}
	return ""
}

func (p *PropertiesPanel) getAssignee() string {
	if p.issue.Assignee != nil {
		return p.issue.Assignee.Name
	}
	return "Unassigned"
}

func (p *PropertiesPanel) getDueDate() string {
	if p.issue.DueDate != nil {
		return p.issue.DueDate.Format("2006-01-02")
	}
	return ""
}

func (p *PropertiesPanel) getStoryPoints() string {
	if p.issue.StoryPoints > 0 {
		return intToString(p.issue.StoryPoints)
	}
	return ""
}

func (p *PropertiesPanel) getLabels() string {
	if len(p.issue.Labels) > 0 {
		return strings.Join(p.issue.Labels, ", ")
	}
	return ""
}

func (p *PropertiesPanel) formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}

// wrapText wraps text to the specified width.
func wrapText(s string, width int) string {
	if width <= 0 {
		return s
	}

	var result strings.Builder
	lines := strings.Split(s, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}

		words := strings.Fields(line)
		lineLen := 0
		for j, word := range words {
			if j > 0 {
				if lineLen+1+len(word) > width {
					result.WriteString("\n")
					lineLen = 0
				} else {
					result.WriteString(" ")
					lineLen++
				}
			}
			result.WriteString(word)
			lineLen += len(word)
		}
	}

	return result.String()
}

// DetailsPanelStyle returns the style for a details panel.
func DetailsPanelStyle(focused bool, width, height int) lipgloss.Style {
	style := Styles.Panel
	if focused {
		style = Styles.FocusedPanel
	}
	return style.Width(width).Height(height)
}
