package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors defines the color palette for the application.
var Colors = struct {
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Error      lipgloss.Color
	Muted      lipgloss.Color
	Background lipgloss.Color
	Foreground lipgloss.Color
	Border     lipgloss.Color
}{
	Primary:    lipgloss.Color("#7C3AED"),
	Secondary:  lipgloss.Color("#06B6D4"),
	Success:    lipgloss.Color("#10B981"),
	Warning:    lipgloss.Color("#F59E0B"),
	Error:      lipgloss.Color("#EF4444"),
	Muted:      lipgloss.Color("#6B7280"),
	Background: lipgloss.Color("#1F2937"),
	Foreground: lipgloss.Color("#F9FAFB"),
	Border:     lipgloss.Color("#374151"),
}

// Styles contains all the lipgloss styles for the application.
var Styles = struct {
	// Base styles
	App           lipgloss.Style
	Title         lipgloss.Style
	Subtitle      lipgloss.Style
	Paragraph     lipgloss.Style
	Muted         lipgloss.Style
	Error         lipgloss.Style
	Success       lipgloss.Style
	Warning       lipgloss.Style

	// Component styles
	Panel         lipgloss.Style
	PanelTitle    lipgloss.Style
	FocusedPanel  lipgloss.Style

	// Table styles
	TableHeader   lipgloss.Style
	TableRow      lipgloss.Style
	TableSelected lipgloss.Style

	// Input styles
	Input         lipgloss.Style
	InputFocused  lipgloss.Style
	InputLabel    lipgloss.Style

	// Button styles
	Button        lipgloss.Style
	ButtonFocused lipgloss.Style

	// Status indicators
	DotRed        lipgloss.Style
	DotYellow     lipgloss.Style
	DotGreen      lipgloss.Style
	DotCyan       lipgloss.Style
	DotEmpty      lipgloss.Style

	// Filter bar
	FilterBar     lipgloss.Style
	FilterLabel   lipgloss.Style
	FilterValue   lipgloss.Style

	// Help
	HelpKey       lipgloss.Style
	HelpDesc      lipgloss.Style
}{
	// Base styles
	App: lipgloss.NewStyle().
		Padding(1, 2),

	Title: lipgloss.NewStyle().
		Bold(true).
		Foreground(Colors.Primary).
		MarginBottom(1),

	Subtitle: lipgloss.NewStyle().
		Foreground(Colors.Secondary),

	Paragraph: lipgloss.NewStyle().
		Foreground(Colors.Foreground),

	Muted: lipgloss.NewStyle().
		Foreground(Colors.Muted),

	Error: lipgloss.NewStyle().
		Foreground(Colors.Error).
		Bold(true),

	Success: lipgloss.NewStyle().
		Foreground(Colors.Success),

	Warning: lipgloss.NewStyle().
		Foreground(Colors.Warning),

	// Component styles
	Panel: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Border).
		Padding(0, 1),

	PanelTitle: lipgloss.NewStyle().
		Bold(true).
		Foreground(Colors.Primary).
		MarginBottom(1),

	FocusedPanel: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Primary).
		Padding(0, 1),

	// Table styles
	TableHeader: lipgloss.NewStyle().
		Bold(true).
		Foreground(Colors.Primary).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(Colors.Border),

	TableRow: lipgloss.NewStyle().
		Foreground(Colors.Foreground),

	TableSelected: lipgloss.NewStyle().
		Background(Colors.Primary).
		Foreground(Colors.Foreground).
		Bold(true),

	// Input styles
	Input: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Border).
		Padding(0, 1),

	InputFocused: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Primary).
		Padding(0, 1),

	InputLabel: lipgloss.NewStyle().
		Foreground(Colors.Muted).
		MarginBottom(1),

	// Button styles
	Button: lipgloss.NewStyle().
		Foreground(Colors.Foreground).
		Background(Colors.Border).
		Padding(0, 2).
		MarginRight(1),

	ButtonFocused: lipgloss.NewStyle().
		Foreground(Colors.Foreground).
		Background(Colors.Primary).
		Padding(0, 2).
		MarginRight(1),

	// Status indicators
	DotRed: lipgloss.NewStyle().
		Foreground(Colors.Error).
		SetString("●"),

	DotYellow: lipgloss.NewStyle().
		Foreground(Colors.Warning).
		SetString("●"),

	DotGreen: lipgloss.NewStyle().
		Foreground(Colors.Success).
		SetString("●"),

	DotCyan: lipgloss.NewStyle().
		Foreground(Colors.Secondary).
		SetString("●"),

	DotEmpty: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		SetString("○"),

	// Filter bar
	FilterBar: lipgloss.NewStyle().
		Foreground(Colors.Foreground).
		MarginBottom(1),

	FilterLabel: lipgloss.NewStyle().
		Foreground(Colors.Muted),

	FilterValue: lipgloss.NewStyle().
		Foreground(Colors.Secondary).
		Bold(true),

	// Help
	HelpKey: lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true),

	HelpDesc: lipgloss.NewStyle().
		Foreground(Colors.Muted),
}

// RenderAttentionIndicator renders the attention indicator for an issue.
func RenderAttentionIndicator(indicator int) string {
	switch indicator {
	case 1: // AttentionNoDueDate
		return Styles.DotYellow.String()
	case 2: // AttentionStale
		return Styles.DotRed.String()
	default:
		return " "
	}
}
