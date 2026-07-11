package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"github.com/leebrandt/grndctrl/internal/grind"
)

func (m Model) dashboardView() string {
	availWidth := m.width - 4

	nameW, typeW, lastSessW, lastCommitW := computeColumnWidths(availWidth)

	var b strings.Builder

	b.WriteString(TitleStyle.Render("GRNDCTRL"))
	b.WriteString("\n\n")

	b.WriteString(renderHeader(availWidth, nameW, typeW, lastSessW, lastCommitW))
	b.WriteString("\n")

	b.WriteString(DimStyle.Render(strings.Repeat("─", availWidth)))
	b.WriteString("\n")

	visible := m.visibleProjects()
	ch := m.contentHeight()

	// Ensure scroll offset is valid
	if m.scrollOffset > len(visible)-ch && len(visible) > ch {
		m.scrollOffset = len(visible) - ch
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}

	start := m.scrollOffset
	end := start + ch
	if end > len(visible) {
		end = len(visible)
	}

	for i := start; i < end; i++ {
		// Map visible index back to original index for cursor/background
		origIdx := m.visibleToOriginal(i)
		b.WriteString(m.renderRow(origIdx, nameW, typeW, lastSessW, lastCommitW))
		b.WriteString("\n")
	}

	// Fill remaining rows if fewer projects than content height
	if len(visible) < ch {
		for i := 0; i < ch-len(visible); i++ {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.statusBar(availWidth))

	out := b.String()
	lines := strings.Split(out, "\n")
	if len(lines) < m.height {
		out += strings.Repeat("\n", m.height-len(lines))
	}

	return out
}

// visibleToOriginal maps a visible index back to the original projects slice index.
func (m *Model) visibleToOriginal(visibleIdx int) int {
	if m.filtered != nil {
		if visibleIdx >= 0 && visibleIdx < len(m.filtered) {
			return m.filtered[visibleIdx]
		}
	}
	return visibleIdx
}

func (m Model) statusBar(availWidth int) string {
	hints := "j/k  move  |  /  filter  |  Enter  detail  |  ?  help  |  q  quit"

	// Filter badge
	if m.filterText != "" {
		badge := fmt.Sprintf("Filter: %s", m.filterText)
		badge = FilterBadgeStyle.Render(badge)
		hintsW := runewidth.StringWidth(hints)
		badgeW := runewidth.StringWidth(badge)
		if hintsW+badgeW+3 < availWidth {
			padding := availWidth - hintsW - badgeW
			hints = hints + strings.Repeat(" ", padding) + badge
		}
	}

	return HelpStyle.Render(hints)
}

func computeColumnWidths(availWidth int) (nameW, typeW, lastSessW, lastCommitW int) {
	fixedWidth := 2 + 7 + 7 + 9 + 7 // status + worked + billed + unbilled + separators
	flexWidth := availWidth - fixedWidth
	if flexWidth < 0 {
		flexWidth = 0
	}

	nameW = flexWidth * 2 / 5
	typeW = flexWidth * 1 / 5
	lastSessW = flexWidth * 1 / 5
	lastCommitW = flexWidth - nameW - typeW - lastSessW

	if nameW < 10 {
		nameW = 10
	}
	if typeW < 4 {
		typeW = 4
	}
	if lastSessW < 8 {
		lastSessW = 8
	}
	if lastCommitW < 8 {
		lastCommitW = 8
	}

	// If total exceeds availWidth, ensure minimums work
	totalFlex := nameW + typeW + lastSessW + lastCommitW
	if totalFlex > flexWidth && flexWidth > 0 {
		shrink := totalFlex - flexWidth
		nameW -= shrink
		if nameW < 10 {
			typeW += (nameW - 10)
			nameW = 10
		}
		if typeW < 4 {
			typeW = 4
		}
		lastCommitW = flexWidth - nameW - typeW - lastSessW
		if lastCommitW < 0 {
			lastCommitW = 0
		}
	}

	return nameW, typeW, lastSessW, lastCommitW
}

type cellDef struct {
	value string
	width int
	align lipgloss.Position
	style lipgloss.Style
}

func renderHeader(availWidth, nameW, typeW, lastSessW, lastCommitW int) string {
	cells := []cellDef{
		{"St", 2, lipgloss.Left, TableHeaderStyle},
		{"Name", nameW, lipgloss.Left, TableHeaderStyle},
		{"Type", typeW, lipgloss.Left, TableHeaderStyle},
		{"Worked", 7, lipgloss.Right, TableHeaderStyle},
		{"Billed", 7, lipgloss.Right, TableHeaderStyle},
		{"Unbilled $", 9, lipgloss.Right, TableHeaderStyle},
		{"Last Session", lastSessW, lipgloss.Left, TableHeaderStyle},
		{"Last Commit", lastCommitW, lipgloss.Left, TableHeaderStyle},
	}

	var b strings.Builder
	for i, c := range cells {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(formatCell(c.value, c.width, c.align, c.style))
	}

	// Ensure header spans full width
	h := b.String()
	if w := runewidth.StringWidth(h); w < availWidth {
		h += strings.Repeat(" ", availWidth-w)
	}

	return h
}

func (m Model) renderRow(i int, nameW, typeW, lastSessW, lastCommitW int) string {
	row := m.projects[i]
	p := row.info.Config

	bgStyle := m.rowBackgroundStyle(i)
	rowStyle := m.rowForegroundStyle(i)

	status := statusChar(row)
	typeVal := p.Type
	if typeVal == "" {
		typeVal = "—"
	}
	worked := formatHours(p.TotalHours())
	billed := formatHours(p.BilledHours())
	unbilled := formatDollars(p.UnbilledAmount())

	lastSess := lastSessionTime(p)
	lastCommit := lastCommitTime(row.lastCommitDate, row.gitErr)

	nameStyle := rowStyle
	if p.LastSession() == nil {
		nameStyle = NeverWorkedStyle
	}

	var b strings.Builder
	b.WriteString(bgStyle.Render(formatCell(status, 2, lipgloss.Left, rowStyle)))
	b.WriteString(" ")
	b.WriteString(bgStyle.Render(formatCell(p.Name, nameW, lipgloss.Left, nameStyle)))
	b.WriteString(" ")
	b.WriteString(bgStyle.Render(formatCell(typeVal, typeW, lipgloss.Left, rowStyle)))
	b.WriteString(" ")
	b.WriteString(bgStyle.Render(formatCell(worked, 7, lipgloss.Right, rowStyle)))
	b.WriteString(" ")
	b.WriteString(bgStyle.Render(formatCell(billed, 7, lipgloss.Right, rowStyle)))
	b.WriteString(" ")

	ubStyle := rowStyle
	if p.UnbilledAmount() > 0 {
		if p.LongTerm {
			ubStyle = ActiveRowMutedStyle
		} else {
			ubStyle = GreenStyle
		}
	}
	b.WriteString(bgStyle.Render(formatCell(unbilled, 9, lipgloss.Right, ubStyle)))
	b.WriteString(" ")
	b.WriteString(bgStyle.Render(formatCell(lastSess, lastSessW, lipgloss.Left, rowStyle)))
	b.WriteString(" ")
	b.WriteString(bgStyle.Render(formatCell(lastCommit, lastCommitW, lipgloss.Left, rowStyle)))

	return b.String()
}

func (m Model) rowBackgroundStyle(i int) lipgloss.Style {
	if i == m.cursor {
		return SelectedStyle
	}
	if i%2 == 1 {
		return AltRowStyle
	}
	return lipgloss.NewStyle()
}

func (m Model) rowForegroundStyle(i int) lipgloss.Style {
	row := m.projects[i]
	p := row.info.Config

	if p.LongTerm {
		switch {
		case p.ActiveSession() != nil:
			return ActiveRowMutedStyle
		case row.dirty:
			return DirtyRowMutedStyle
		default:
			return LongTermDefaultStyle
		}
	}

	switch {
	case p.ActiveSession() != nil:
		return ActiveRowStyle
	case row.dirty:
		return DirtyRowStyle
	default:
		return GreenStyle
	}
}

func statusChar(row projectRow) string {
	p := row.info.Config
	switch {
	case p.ActiveSession() != nil:
		return "▶"
	case row.dirty:
		return "!"
	case p.LongTerm:
		return "★"
	default:
		return " "
	}
}

func formatCell(value string, width int, align lipgloss.Position, style lipgloss.Style) string {
	w := runewidth.StringWidth(value)
	if w > width {
		value = runewidth.Truncate(value, width, "…")
		w = runewidth.StringWidth(value)
	}

	var padded string
	switch align {
	case lipgloss.Right:
		if w < width {
			padded = strings.Repeat(" ", width-w) + value
		} else {
			padded = value
		}
	default:
		if w < width {
			padded = value + strings.Repeat(" ", width-w)
		} else {
			padded = value
		}
	}

	return style.Render(padded)
}

func formatHours(h float64) string {
	return fmt.Sprintf("%.1fh", h)
}

func formatDollars(amount float64) string {
	if amount == 0 {
		return "$0"
	}
	return fmt.Sprintf("$%.0f", amount)
}

func relativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		m := int(diff.Minutes())
		if m == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", m)
	case diff < 24*time.Hour:
		h := int(diff.Hours())
		if h == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", h)
	case diff < 30*24*time.Hour:
		d := int(diff.Hours() / 24)
		if d == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", d)
	case diff < 365*24*time.Hour:
		mo := int(diff.Hours() / 24 / 30)
		if mo == 1 {
			return "1mo ago"
		}
		return fmt.Sprintf("%dmo ago", mo)
	default:
		y := int(diff.Hours() / 24 / 365)
		if y == 1 {
			return "1y ago"
		}
		return fmt.Sprintf("%dy ago", y)
	}
}

func lastSessionTime(p grind.ProjectConfig) string {
	s := p.LastSession()
	if s == nil {
		return "never"
	}
	if s.End != nil {
		return relativeTime(*s.End)
	}
	return "just now"
}

func lastCommitTime(date string, gitErr bool) string {
	if gitErr || date == "" {
		return "?"
	}
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return "?"
	}
	return relativeTime(t)
}
