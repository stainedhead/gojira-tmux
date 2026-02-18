package tui

import (
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

// CommentsPanel displays issue comments.
type CommentsPanel struct {
	comments []domain.Comment
	focused  bool
	width    int
	height   int
	scrollY  int
}

// NewCommentsPanel creates a new comments panel.
func NewCommentsPanel() *CommentsPanel {
	return &CommentsPanel{}
}

// SetComments sets the comments to display (sorted newest first).
func (c *CommentsPanel) SetComments(comments []domain.Comment) {
	sorted := make([]domain.Comment, len(comments))
	copy(sorted, comments)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Created.After(sorted[j].Created)
	})
	c.comments = sorted
	c.scrollY = 0
}

// Focus sets the panel as focused.
func (c *CommentsPanel) Focus() {
	c.focused = true
}

// Blur removes focus from the panel.
func (c *CommentsPanel) Blur() {
	c.focused = false
}

// Focused returns whether the panel is focused.
func (c *CommentsPanel) Focused() bool {
	return c.focused
}

// SetSize sets the panel dimensions.
func (c *CommentsPanel) SetSize(width, height int) {
	c.width = width
	c.height = height
}

// Update handles messages for the panel.
func (c *CommentsPanel) Update(msg tea.Msg) (*CommentsPanel, tea.Cmd) {
	if !c.focused {
		return c, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if c.scrollY > 0 {
				c.scrollY--
			}
		case "down", "j":
			c.scrollY++
		}
	}

	return c, nil
}

// View renders the comments panel clamped to its set height.
func (c *CommentsPanel) View() string {
	panelStyle := Styles.Panel
	if c.focused {
		panelStyle = Styles.FocusedPanel
	}

	// Build every content line up-front so we can apply viewport clamping.
	var lines []string

	count := len(c.comments)
	title := "Comments"
	if count > 0 {
		title += " (" + intToString(count) + ")"
	}
	lines = append(lines, Styles.PanelTitle.Render(title))

	if len(c.comments) == 0 {
		lines = append(lines, Styles.Muted.Render("No comments"))
	} else {
		wrapWidth := c.width - 4
		if wrapWidth < 1 {
			wrapWidth = 1
		}
		sepWidth := c.width - 4
		if sepWidth < 1 {
			sepWidth = 1
		}

		for i, comment := range c.comments {
			if i > 0 {
				lines = append(lines, "")
				lines = append(lines, Styles.Muted.Render(strings.Repeat("─", sepWidth)))
			}
			header := Styles.FilterValue.Render(comment.Author) +
				Styles.Muted.Render(" · ") +
				Styles.Muted.Render(comment.Created.Format("Jan 2, 2006 15:04"))
			lines = append(lines, header)
			for _, bodyLine := range strings.Split(wrapText(comment.Body, wrapWidth), "\n") {
				lines = append(lines, Styles.Paragraph.Render(bodyLine))
			}
		}
	}

	return c.renderViewport(panelStyle, lines)
}

// renderViewport clamps lines to the visible height and renders the panel.
// The panel always occupies exactly c.height terminal lines (outer, including borders).
func (c *CommentsPanel) renderViewport(panelStyle lipgloss.Style, lines []string) string {
	viewHeight := max(c.height-2, 1)

	maxScroll := max(len(lines)-viewHeight, 0)
	if c.scrollY > maxScroll {
		c.scrollY = maxScroll
	}

	end := min(c.scrollY+viewHeight, len(lines))
	visible := strings.Join(lines[c.scrollY:end], "\n")

	return panelStyle.Width(max(c.width, 0)).Height(max(c.height, 0)).Render(visible)
}
