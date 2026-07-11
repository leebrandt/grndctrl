package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

var (
	overlayBgStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1a1b26")).
			Padding(1, 2)

	overlayBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorAccent).
				Padding(1, 2)

	overlayTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent)

	overlayKeyStyle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	overlayDescStyle = lipgloss.NewStyle().
				Foreground(colorFg)

	overlayDimStyle = lipgloss.NewStyle().
			Foreground(colorDim)
)

type helpEntry struct {
	key         string
	description string
}

var helpEntries = []helpEntry{
	{"j / ↓", "Move cursor down"},
	{"k / ↑", "Move cursor up"},
	{"g", "Jump to first row"},
	{"G", "Jump to last row"},
	{"Ctrl+d", "Page down (half screen)"},
	{"Ctrl+u", "Page up (half screen)"},
	{"/", "Enter filter mode"},
	{"Enter", "Open project detail"},
	{"i", "Ideas triage panel"},
	{"r", "Force refresh data"},
	{"?", "Toggle this help overlay"},
	{"Tab", "Switch focus between panels"},
	{"q / Esc", "Quit"},
}

func (m Model) helpOverlay(viewContent string) string {
	// Build the overlay box
	title := overlayTitleStyle.Render("Keyboard Shortcuts")
	separator := overlayDimStyle.Render(strings.Repeat("─", 30))

	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(separator)
	b.WriteString("\n\n")

	for _, entry := range helpEntries {
		key := overlayKeyStyle.Render(entry.key)
		keyW := runewidth.StringWidth(entry.key)
		padding := 18 - keyW
		if padding < 1 {
			padding = 1
		}
		b.WriteString(key)
		b.WriteString(strings.Repeat(" ", padding))
		b.WriteString(overlayDescStyle.Render(entry.description))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(overlayDimStyle.Render("Press ? or Esc to close"))

	boxContent := b.String()

	// Determine overlay dimensions
	overlayWidth := 48

	// Wrap in border
	box := overlayBorderStyle.Width(overlayWidth - 4).Render(boxContent)

	// Place in center of viewport
	viewWidth := m.width
	viewHeight := m.height
	if viewWidth == 0 {
		viewWidth = 80
	}
	if viewHeight == 0 {
		viewHeight = 24
	}

	// Create a semi-transparent-like overlay by placing the box in the center
	overlay := lipgloss.Place(
		viewWidth,
		viewHeight,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)

	return overlay
}
