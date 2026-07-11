package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"github.com/leebrandt/grndctrl/internal/workspace"
)

func (m Model) ideasView() string {
	availWidth := m.width - 4

	var b strings.Builder

	b.WriteString(TitleStyle.Render("GRNDCTRL"))
	b.WriteString("\n\n")

	if !m.ideasLoaded {
		b.WriteString(DimStyle.Render("Loading ideas..."))
		b.WriteString("\n")
		return m.padToHeight(b.String())
	}

	if m.ideasErr != nil {
		b.WriteString(ErrorStyle.Render("Error loading ideas:"))
		b.WriteString("\n")
		b.WriteString(DimStyle.Render(m.ideasErr.Error()))
		b.WriteString("\n")
		return m.padToHeight(b.String())
	}

	visible := m.visibleIdeas()
	if len(visible) == 0 {
		emptyMsg := "No ideas yet."
		if m.ideasRejected {
			emptyMsg = "No rejected ideas."
		}
		content := lipgloss.JoinVertical(
			lipgloss.Center,
			DimStyle.Render(emptyMsg),
		)
		b.WriteString(content)
		b.WriteString("\n")
		return m.padToHeight(b.String())
	}

	b.WriteString(renderIdeasHeader(availWidth))
	b.WriteString("\n")

	b.WriteString(DimStyle.Render(strings.Repeat("─", availWidth)))
	b.WriteString("\n")

	ch := m.ideasContentHeight()

	if m.ideasScroll > len(visible)-ch && len(visible) > ch {
		m.ideasScroll = len(visible) - ch
	}
	if m.ideasScroll < 0 {
		m.ideasScroll = 0
	}

	start := m.ideasScroll
	end := start + ch
	if end > len(visible) {
		end = len(visible)
	}

	for i := start; i < end; i++ {
		b.WriteString(m.renderIdeaRow(visible[i], i == m.ideasCursor, availWidth))
		b.WriteString("\n")
	}

	if len(visible) < ch {
		for i := 0; i < ch-len(visible); i++ {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.ideasStatusBar(availWidth))

	return m.padToHeight(b.String())
}

func (m Model) padToHeight(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) < m.height {
		content += strings.Repeat("\n", m.height-len(lines))
	}
	return content
}

func renderIdeasHeader(availWidth int) string {
	numW := 4
	statusW := 10
	createdW := 12
	titleW := availWidth - numW - statusW - createdW - 3
	if titleW < 10 {
		titleW = 10
	}

	cells := []cellDef{
		{"#", numW, lipgloss.Left, TableHeaderStyle},
		{"Title", titleW, lipgloss.Left, TableHeaderStyle},
		{"Status", statusW, lipgloss.Left, TableHeaderStyle},
		{"Created", createdW, lipgloss.Left, TableHeaderStyle},
	}

	var b strings.Builder
	for i, c := range cells {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(formatCell(c.value, c.width, c.align, c.style))
	}

	h := b.String()
	if w := runewidth.StringWidth(h); w < availWidth {
		h += strings.Repeat(" ", availWidth-w)
	}

	return h
}

func (m Model) renderIdeaRow(idea workspace.Idea, selected bool, availWidth int) string {
	numW := 4
	statusW := 10
	createdW := 12
	titleW := availWidth - numW - statusW - createdW - 3
	if titleW < 10 {
		titleW = 10
	}

	numStr := fmt.Sprintf("%d", idea.Number)
	createdStr := idea.Created.Format("2006-01-02")

	var statusStr string
	var statusStyle lipgloss.Style
	switch {
	case idea.Rejected:
		statusStr = "Rejected"
		statusStyle = RedStyle
	default:
		statusStr = "Idea"
		statusStyle = GreenStyle
	}

	bgStyle := lipgloss.NewStyle()
	if selected {
		bgStyle = SelectedStyle
	}

	titleStyle := GreenStyle
	if idea.Rejected {
		titleStyle = RedStyle
	}

	var b strings.Builder
	b.WriteString(bgStyle.Render(formatCell(numStr, numW, lipgloss.Left, statusStyle)))
	b.WriteString(" ")
	b.WriteString(bgStyle.Render(formatCell(idea.Title, titleW, lipgloss.Left, titleStyle)))
	b.WriteString(" ")
	b.WriteString(bgStyle.Render(formatCell(statusStr, statusW, lipgloss.Left, statusStyle)))
	b.WriteString(" ")
	b.WriteString(bgStyle.Render(formatCell(createdStr, createdW, lipgloss.Left, DimStyle)))

	return b.String()
}

func (m Model) ideasStatusBar(availWidth int) string {
	hints := "j/k  move  |  a  all  |  r  rejected  |  Esc  back  |  q  quit"

	modeLabel := ""
	switch {
	case m.ideasAll:
		modeLabel = "All"
	case m.ideasRejected:
		modeLabel = "Rejected"
	default:
		modeLabel = "Ideas"
	}

	badge := fmt.Sprintf("Mode: %s", modeLabel)
	badge = FilterBadgeStyle.Render(badge)
	hintsW := runewidth.StringWidth(hints)
	badgeW := runewidth.StringWidth(badge)
	if hintsW+badgeW+3 < availWidth {
		padding := availWidth - hintsW - badgeW
		hints = hints + strings.Repeat(" ", padding) + badge
	}

	return HelpStyle.Render(hints)
}
