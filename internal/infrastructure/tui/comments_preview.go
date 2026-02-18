package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

const previewMaxComments = 5

// CommentsPreview displays the most recent comments for the currently selected
// issue. It is read-only and cannot be focused.
type CommentsPreview struct {
	comments []domain.Comment
	width    int
}

// NewCommentsPreview creates a new CommentsPreview.
func NewCommentsPreview() *CommentsPreview {
	return &CommentsPreview{}
}

// SetComments replaces the displayed comments with the most recent
// previewMaxComments from the supplied slice (sorted newest-first).
func (c *CommentsPreview) SetComments(comments []domain.Comment) {
	sorted := make([]domain.Comment, len(comments))
	copy(sorted, comments)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Created.After(sorted[j].Created)
	})
	if len(sorted) > previewMaxComments {
		sorted = sorted[:previewMaxComments]
	}
	c.comments = sorted
}

// SetWidth sets the available render width.
func (c *CommentsPreview) SetWidth(width int) {
	c.width = width
}

// View renders the comments preview as muted, non-interactive text.
func (c *CommentsPreview) View() string {
	var b strings.Builder

	// Header separator line
	headerLabel := " Recent Comments "
	sepWidth := c.width - len(headerLabel) - 2
	if sepWidth < 1 {
		sepWidth = 1
	}
	header := Styles.Muted.Render(strings.Repeat("─", 1) + headerLabel + strings.Repeat("─", sepWidth))
	b.WriteString(header)
	b.WriteString("\n")

	if len(c.comments) == 0 {
		b.WriteString(Styles.Muted.Render("  No comments"))
		return b.String()
	}

	const authorCol = 16
	const dateCol = 8

	bodyWidth := c.width - authorCol - dateCol - 6
	if bodyWidth < 10 {
		bodyWidth = 10
	}

	for _, comment := range c.comments {
		author := truncate(comment.Author, authorCol)
		date := formatRelativeTime(&comment.Created)
		body := strings.ReplaceAll(comment.Body, "\n", " ")
		body = truncate(body, bodyWidth)

		row := fmt.Sprintf("  %-*s  %-*s  %s", authorCol, author, dateCol, date, body)
		b.WriteString(Styles.Muted.Render(row))
		b.WriteString("\n")
	}

	return b.String()
}
