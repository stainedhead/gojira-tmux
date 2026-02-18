package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/stainedhead/gojira-tmux/internal/domain"
)

func makeComment(author, body string, daysAgo int) domain.Comment {
	return domain.Comment{
		ID:      author,
		Author:  author,
		Body:    body,
		Created: time.Now().Add(-time.Duration(daysAgo) * 24 * time.Hour),
	}
}

func TestCommentsPreview_View_NoComments(t *testing.T) {
	p := NewCommentsPreview()
	p.SetWidth(80)
	p.SetComments(nil)
	view := p.View()

	if !strings.Contains(view, "No comments") {
		t.Errorf("expected 'No comments' in view, got: %q", view)
	}
}

func TestCommentsPreview_View_ShowsComments(t *testing.T) {
	p := NewCommentsPreview()
	p.SetWidth(80)
	p.SetComments([]domain.Comment{
		makeComment("Alice", "Fixed the auth flow", 2),
		makeComment("Bob", "Looks good to me", 1),
	})
	view := p.View()

	if !strings.Contains(view, "Alice") {
		t.Errorf("expected 'Alice' in view, got: %q", view)
	}
	if !strings.Contains(view, "Bob") {
		t.Errorf("expected 'Bob' in view, got: %q", view)
	}
}

func TestCommentsPreview_View_ShowsAtMostFive(t *testing.T) {
	p := NewCommentsPreview()
	p.SetWidth(80)
	comments := []domain.Comment{
		makeComment("User1", "Comment 1", 10),
		makeComment("User2", "Comment 2", 9),
		makeComment("User3", "Comment 3", 8),
		makeComment("User4", "Comment 4", 7),
		makeComment("User5", "Comment 5", 6),
		makeComment("User6", "Comment 6", 5),
		makeComment("User7", "Comment 7", 4),
	}
	p.SetComments(comments)
	view := p.View()

	// User1 is the oldest - should NOT appear (only 5 most recent shown)
	if strings.Contains(view, "User1") {
		t.Error("oldest comment (User1) should not appear — only 5 most recent shown")
	}
	// User7 is the most recent - should appear
	if !strings.Contains(view, "User7") {
		t.Errorf("most recent comment (User7) should appear in view, got: %q", view)
	}
}

func TestCommentsPreview_View_MostRecentFirst(t *testing.T) {
	p := NewCommentsPreview()
	p.SetWidth(80)
	p.SetComments([]domain.Comment{
		makeComment("OldComment", "old", 20),
		makeComment("NewComment", "new", 1),
	})
	view := p.View()

	oldIdx := strings.Index(view, "OldComment")
	newIdx := strings.Index(view, "NewComment")
	if oldIdx == -1 || newIdx == -1 {
		t.Fatalf("expected both comments in view, got: %q", view)
	}
	if newIdx > oldIdx {
		t.Error("most recent comment should appear before older comment")
	}
}

func TestCommentsPreview_View_NoPanicWithZeroWidth(t *testing.T) {
	p := NewCommentsPreview()
	p.SetWidth(0)
	p.SetComments([]domain.Comment{
		makeComment("Alice", "Comment", 1),
		makeComment("Bob", "Another", 2),
	})
	_ = p.View() // must not panic
}

func TestCommentsPreview_View_ContainsHeader(t *testing.T) {
	p := NewCommentsPreview()
	p.SetWidth(80)
	p.SetComments(nil)
	view := p.View()

	if !strings.Contains(view, "Recent Comments") {
		t.Errorf("expected header 'Recent Comments' in view, got: %q", view)
	}
}
