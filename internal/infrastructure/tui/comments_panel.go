package tui

import (
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

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
	// Sort by created time, newest first
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

// View renders the comments panel.
func (c *CommentsPanel) View() string {
	var b strings.Builder

	// Panel style
	panelStyle := Styles.Panel
	if c.focused {
		panelStyle = Styles.FocusedPanel
	}

	// Title with count
	count := len(c.comments)
	title := "Comments"
	if count > 0 {
		title += " (" + intToString(count) + ")"
	}
	b.WriteString(Styles.PanelTitle.Render(title))
	b.WriteString("\n")

	if len(c.comments) == 0 {
		b.WriteString(Styles.Muted.Render("No comments"))
		return panelStyle.Width(c.width).Height(c.height).Render(b.String())
	}

	// Render comments
	for i, comment := range c.comments {
		if i > 0 {
			b.WriteString("\n")
			b.WriteString(Styles.Muted.Render(strings.Repeat("─", c.width-4)))
			b.WriteString("\n")
		}

		c.renderComment(&b, comment)
	}

	return panelStyle.Width(c.width).Height(c.height).Render(b.String())
}

func (c *CommentsPanel) renderComment(b *strings.Builder, comment domain.Comment) {
	// Header: Author - Date
	header := Styles.FilterValue.Render(comment.Author)
	header += Styles.Muted.Render(" · ")
	header += Styles.Muted.Render(comment.Created.Format("Jan 2, 2006 15:04"))
	b.WriteString(header)
	b.WriteString("\n")

	// Body (wrapped)
	body := wrapText(comment.Body, c.width-4)
	b.WriteString(Styles.Paragraph.Render(body))
	b.WriteString("\n")
}
