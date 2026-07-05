package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model is the root Bubble Tea model for the GRNDCTRL application.
type Model struct {
	ready       bool
	width       int
	height      int
	workspace   string
}

// NewModel creates the root model with the given workspace path.
func NewModel(workspace string) Model {
	return Model{
		workspace: workspace,
	}
}

// Init returns any initial commands. For now, just wait for the window to
// be sized.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and returns the updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.ready = true
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	return m, nil
}

// View renders the current state of the UI.
func (m Model) View() string {
	if !m.ready {
		return "\n  Loading..."
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		TitleStyle.Render("GRNDCTRL"),
		"",
		DimStyle.Render("grind workspace dashboard"),
		"",
		DimStyle.Render(fmt.Sprintf("workspace: %s", m.workspace)),
		"",
		"",
		DimStyle.Render("press q or Ctrl+C to quit"),
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
