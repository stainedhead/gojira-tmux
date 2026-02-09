package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/stainedhead/gojira-tmux/internal/adapter/auth"
	"github.com/stainedhead/gojira-tmux/internal/infrastructure/tui"
)

func TestSetupScreen_View_ShowsInstructions(t *testing.T) {
	tokenStore := auth.NewMemoryTokenStore()
	authPort := &mockAuthPort{}

	screen := tui.NewSetupScreenModel(tokenStore, authPort)
	view := screen.View()

	// Should contain setup instructions
	if !contains(view, "API Token") {
		t.Errorf("Setup screen should mention API Token, got: %s", view)
	}

	// Should contain the token generation URL
	if !contains(view, "atlassian.com") {
		t.Errorf("Setup screen should contain Atlassian URL, got: %s", view)
	}
}

func TestSetupScreen_View_ShowsEmailAndTokenFields(t *testing.T) {
	tokenStore := auth.NewMemoryTokenStore()
	authPort := &mockAuthPort{}

	screen := tui.NewSetupScreenModel(tokenStore, authPort)
	view := screen.View()

	if !contains(view, "Email") {
		t.Errorf("Setup screen should have Email label, got: %s", view)
	}
	if !contains(view, "Token") || !contains(view, "token") {
		t.Errorf("Setup screen should have Token label, got: %s", view)
	}
}

func TestSetupScreen_Init_ReturnsBlink(t *testing.T) {
	tokenStore := auth.NewMemoryTokenStore()
	authPort := &mockAuthPort{}

	screen := tui.NewSetupScreenModel(tokenStore, authPort)
	cmd := screen.Init()

	if cmd == nil {
		t.Error("Init() should return blink command")
	}
}

func TestSetupScreen_Update_EscQuits(t *testing.T) {
	tokenStore := auth.NewMemoryTokenStore()
	authPort := &mockAuthPort{}

	screen := tui.NewSetupScreenModel(tokenStore, authPort)

	_, cmd := screen.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Error("Esc should return quit command")
	}
}

func TestSetupScreen_Update_CtrlCQuits(t *testing.T) {
	tokenStore := auth.NewMemoryTokenStore()
	authPort := &mockAuthPort{}

	screen := tui.NewSetupScreenModel(tokenStore, authPort)

	_, cmd := screen.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("Ctrl+C should return quit command")
	}
}

func TestSetupScreen_Update_TabSwitchesFields(t *testing.T) {
	tokenStore := auth.NewMemoryTokenStore()
	authPort := &mockAuthPort{}

	screen := tui.NewSetupScreenModel(tokenStore, authPort)

	// Tab should switch focus (not crash)
	model, cmd := screen.Update(tea.KeyMsg{Type: tea.KeyTab})
	if model == nil {
		t.Error("Tab should return non-nil model")
	}
	// Should return blink command for new focused input
	if cmd == nil {
		t.Error("Tab should return blink command")
	}
}

func TestSetupScreen_HelpText(t *testing.T) {
	tokenStore := auth.NewMemoryTokenStore()
	authPort := &mockAuthPort{}

	screen := tui.NewSetupScreenModel(tokenStore, authPort)
	view := screen.View()

	// Should show help keys
	if !contains(view, "tab") {
		t.Errorf("Setup screen should show tab key hint, got: %s", view)
	}
	if !contains(view, "enter") {
		t.Errorf("Setup screen should show enter key hint, got: %s", view)
	}
	if !contains(view, "esc") {
		t.Errorf("Setup screen should show esc key hint, got: %s", view)
	}
}
