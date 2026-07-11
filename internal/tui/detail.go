package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"github.com/leebrandt/grndctrl/internal/grind"
)

// detailView renders the full project detail panel content.
func (m Model) detailView() string {
	if m.detailProject < 0 || m.detailProject >= len(m.projects) {
		return ""
	}

	row := m.projects[m.detailProject]
	p := row.info.Config

	var b strings.Builder

	b.WriteString(m.detailHeader(p))
	b.WriteString("\n")

	b.WriteString(m.detailSummary(p))
	b.WriteString("\n")

	b.WriteString(m.detailSessionTable(p))
	b.WriteString("\n")

	b.WriteString(m.detailGitSummary(row))
	b.WriteString("\n")

	b.WriteString(m.detailProjectConfig(p))

	return b.String()
}

// detailHeader renders the project name, type badge, idea, and separator.
func (m Model) detailHeader(p grind.ProjectConfig) string {
	var b strings.Builder

	// Project name (bold, large)
	name := DetailHeaderStyle.Render(p.Name)
	b.WriteString(name)

	// Long-term star
	if p.LongTerm {
		b.WriteString(" ")
		b.WriteString(DetailGoldStyle.Render("★"))
	}

	// Type badge
	if p.Type != "" {
		b.WriteString("  ")
		b.WriteString(DetailDimStyle.Render("[" + p.Type + "]"))
	}

	b.WriteString("\n")

	// Idea (truncated to fit)
	if p.Idea != "" {
		availWidth := m.width - 8
		if availWidth < 10 {
			availWidth = 10
		}
		idea := runewidth.Truncate(p.Idea, availWidth, "…")
		b.WriteString(DetailDimStyle.Render(idea))
		b.WriteString("\n")
	}

	// Separator
	sepWidth := m.width - 4
	if sepWidth < 1 {
		sepWidth = 1
	}
	b.WriteString(DetailSeparator.Render(strings.Repeat("─", sepWidth)))

	return b.String()
}

// detailSummary renders the summary row with key-value pairs.
func (m Model) detailSummary(p grind.ProjectConfig) string {
	var b strings.Builder

	b.WriteString(DetailSectionStyle.Render("Summary"))
	b.WriteString("\n")

	totalH := p.TotalHours()
	billedH := p.BilledHours()
	unbilledH := p.UnbilledHours()

	totalAmt := p.TotalAmount()
	billedAmt := p.BilledAmount()
	unbilledAmt := p.UnbilledAmount()

	sessionCount := len(p.Time)

	// Build key-value pairs
	pairs := []struct {
		label string
		value string
		style lipgloss.Style
	}{
		{"Total Hours", formatDetailHours(totalH), DetailValueStyle},
		{"Billed Hours", formatDetailHours(billedH), DetailGreenStyle},
		{"Unbilled Hours", formatDetailHours(unbilledH), detailAmountStyle(unbilledH > 0, p.LongTerm)},
		{"Total Amount", formatDetailDollars(totalAmt), DetailValueStyle},
		{"Billed Amount", formatDetailDollars(billedAmt), DetailGreenStyle},
		{"Unbilled Amount", formatDetailDollars(unbilledAmt), detailAmountStyle(unbilledAmt > 0, p.LongTerm)},
		{"Sessions", fmt.Sprintf("%d", sessionCount), DetailValueStyle},
	}

	// Determine column layout
	availWidth := m.width - 4
	if availWidth < 40 {
		availWidth = 40
	}

	// Two columns: label + value pairs
	colWidth := availWidth / 2
	labelW := 18
	valW := colWidth - labelW
	if valW < 10 {
		valW = 10
	}

	for i := 0; i < len(pairs); i += 2 {
		left := renderPair(pairs[i].label, pairs[i].value, labelW, valW, pairs[i].style)
		if i+1 < len(pairs) {
			right := renderPair(pairs[i+1].label, pairs[i+1].value, labelW, valW, pairs[i+1].style)
			b.WriteString(left)
			b.WriteString("  ")
			b.WriteString(right)
		} else {
			b.WriteString(left)
		}
		b.WriteString("\n")
	}

	// Separator
	sepWidth := m.width - 4
	if sepWidth < 1 {
		sepWidth = 1
	}
	b.WriteString(DetailSeparator.Render(strings.Repeat("─", sepWidth)))

	return b.String()
}

// detailSessionTable renders the session history table.
func (m Model) detailSessionTable(p grind.ProjectConfig) string {
	var b strings.Builder

	b.WriteString(DetailSectionStyle.Render("Session History"))
	b.WriteString("\n")

	if len(p.Time) == 0 {
		b.WriteString(DetailDimStyle.Render("  No sessions recorded."))
		b.WriteString("\n")
		sepWidth := m.width - 4
		if sepWidth < 1 {
			sepWidth = 1
		}
		b.WriteString(DetailSeparator.Render(strings.Repeat("─", sepWidth)))
		return b.String()
	}

	// Sort sessions chronologically by start time
	sessions := sortedSessions(p.Time)

	// Determine column widths
	availWidth := m.width - 4
	if availWidth < 50 {
		availWidth = 50
	}

	// Fixed columns: # (4), Date (12), Start (9), End (9), Duration (10), Rounded (8), Invoiced (10)
	fixedW := 4 + 12 + 9 + 9 + 10 + 8 + 10 + 6 // +6 for separators
	if fixedW > availWidth {
		// Shrink date and time columns proportionally
		overflow := fixedW - availWidth
		dateW := 12 - overflow/3
		if dateW < 8 {
			dateW = 8
		}
		startW := 9 - overflow/3
		if startW < 6 {
			startW = 6
		}
		endW := 9 - overflow/3
		if endW < 6 {
			endW = 6
		}

		// Render header
		headerCells := []cellDef{
			{"#", 4, lipgloss.Left, DetailTableHeaderStyle},
			{"Date", dateW, lipgloss.Left, DetailTableHeaderStyle},
			{"Start", startW, lipgloss.Left, DetailTableHeaderStyle},
			{"End", endW, lipgloss.Left, DetailTableHeaderStyle},
			{"Duration", 10, lipgloss.Right, DetailTableHeaderStyle},
			{"Rounded", 8, lipgloss.Right, DetailTableHeaderStyle},
			{"Invoiced?", 10, lipgloss.Left, DetailTableHeaderStyle},
		}

		var hdrBuf strings.Builder
		for i, c := range headerCells {
			if i > 0 {
				hdrBuf.WriteString(" ")
			}
			hdrBuf.WriteString(formatCell(c.value, c.width, c.align, c.style))
		}
		b.WriteString(hdrBuf.String())
		b.WriteString("\n")

		for i, s := range sessions {
			bg := lipgloss.NewStyle()
			if i%2 == 1 {
				bg = DetailAltRowStyle
			}

			num := fmt.Sprintf("%d", i+1)
			dateStr := s.Start.Local().Format("2006-01-02")
			startStr := s.Start.Local().Format("3:04 PM")
			endStr := "—"
			if s.End != nil {
				endStr = s.End.Local().Format("3:04 PM")
			}
			durStr := s.DurationHuman()
			roundedStr := formatRoundedHours(s.Rounded)
			invoicedStr := "—"
			invoicedStyle := DetailDimStyle
			if s.Invoiced != nil && *s.Invoiced {
				invoicedStr = "✓"
				invoicedStyle = DetailGreenStyle
			}

			// Active session badge
			if s.End == nil {
				endStr = DetailYellowStyle.Render("Active")
			}

			rowCells := []cellDef{
				{num, 4, lipgloss.Left, bg},
				{dateStr, dateW, lipgloss.Left, bg},
				{startStr, startW, lipgloss.Left, bg},
				{endStr, endW, lipgloss.Left, bg},
				{durStr, 10, lipgloss.Right, bg},
				{roundedStr, 8, lipgloss.Right, bg},
				{invoicedStr, 10, lipgloss.Left, invoicedStyle},
			}

			for j, c := range rowCells {
				if j > 0 {
					b.WriteString(" ")
				}
				b.WriteString(formatCell(c.value, c.width, c.align, c.style))
			}
			b.WriteString("\n")
		}
	} else {
		// Render header with full widths
		headerCells := []cellDef{
			{"#", 4, lipgloss.Left, DetailTableHeaderStyle},
			{"Date", 12, lipgloss.Left, DetailTableHeaderStyle},
			{"Start", 9, lipgloss.Left, DetailTableHeaderStyle},
			{"End", 9, lipgloss.Left, DetailTableHeaderStyle},
			{"Duration", 10, lipgloss.Right, DetailTableHeaderStyle},
			{"Rounded", 8, lipgloss.Right, DetailTableHeaderStyle},
			{"Invoiced?", 10, lipgloss.Left, DetailTableHeaderStyle},
		}

		var hdrBuf strings.Builder
		for i, c := range headerCells {
			if i > 0 {
				hdrBuf.WriteString(" ")
			}
			hdrBuf.WriteString(formatCell(c.value, c.width, c.align, c.style))
		}
		b.WriteString(hdrBuf.String())
		b.WriteString("\n")

		for i, s := range sessions {
			bg := lipgloss.NewStyle()
			if i%2 == 1 {
				bg = DetailAltRowStyle
			}

			num := fmt.Sprintf("%d", i+1)
			dateStr := s.Start.Local().Format("2006-01-02")
			startStr := s.Start.Local().Format("3:04 PM")
			endStr := "—"
			if s.End != nil {
				endStr = s.End.Local().Format("3:04 PM")
			}
			durStr := s.DurationHuman()
			roundedStr := formatRoundedHours(s.Rounded)
			invoicedStr := "—"
			invoicedStyle := DetailDimStyle
			if s.Invoiced != nil && *s.Invoiced {
				invoicedStr = "✓"
				invoicedStyle = DetailGreenStyle
			}

			// Active session badge
			if s.End == nil {
				endStr = DetailYellowStyle.Render("Active")
			}

			rowCells := []cellDef{
				{num, 4, lipgloss.Left, bg},
				{dateStr, 12, lipgloss.Left, bg},
				{startStr, 9, lipgloss.Left, bg},
				{endStr, 9, lipgloss.Left, bg},
				{durStr, 10, lipgloss.Right, bg},
				{roundedStr, 8, lipgloss.Right, bg},
				{invoicedStr, 10, lipgloss.Left, invoicedStyle},
			}

			for j, c := range rowCells {
				if j > 0 {
					b.WriteString(" ")
				}
				b.WriteString(formatCell(c.value, c.width, c.align, c.style))
			}
			b.WriteString("\n")
		}
	}

	// Separator
	sepWidth := m.width - 4
	if sepWidth < 1 {
		sepWidth = 1
	}
	b.WriteString(DetailSeparator.Render(strings.Repeat("─", sepWidth)))

	return b.String()
}

// detailGitSummary renders the git timeline section.
func (m Model) detailGitSummary(row projectRow) string {
	var b strings.Builder

	b.WriteString(DetailSectionStyle.Render("Git Summary"))
	b.WriteString("\n")

	p := row.info.Config

	// First commit / Last commit dates
	firstCommit := "?"
	lastCommit := "?"
	commitCount := 0

	// We need the bare repo path. Derive it from the workspace.
	// The workspace is stored in m.workspace.
	bareRepoPath := m.workspace + "/.grind.repo.git"

	fc, err := grind.FirstCommitDate(bareRepoPath, row.info.Branch)
	if err == nil && fc != "" {
		t, err := time.Parse(time.RFC3339, fc)
		if err == nil {
			firstCommit = relativeTime(t)
		}
	}

	lc, err := grind.LastCommitDate(bareRepoPath, row.info.Branch)
	if err == nil && lc != "" {
		t, err := time.Parse(time.RFC3339, lc)
		if err == nil {
			lastCommit = relativeTime(t)
		}
	}

	cc, err := grind.CommitCount(bareRepoPath, row.info.Branch)
	if err == nil {
		commitCount = cc
	}

	dirty := row.dirty
	dirtyStr := "clean"
	dirtyStyle := DetailGreenStyle
	if dirty {
		dirtyStr = "!"
		dirtyStyle = DetailYellowStyle
	}

	availWidth := m.width - 4
	if availWidth < 40 {
		availWidth = 40
	}

	labelW := 20
	valW := availWidth/2 - labelW
	if valW < 10 {
		valW = 10
	}

	pairs := []struct {
		label string
		value string
		style lipgloss.Style
	}{
		{"First Commit", firstCommit, DetailValueStyle},
		{"Last Commit", lastCommit, DetailValueStyle},
		{"Commit Count", fmt.Sprintf("%d", commitCount), DetailValueStyle},
		{"Uncommitted Changes", dirtyStr, dirtyStyle},
	}

	for i := 0; i < len(pairs); i += 2 {
		left := renderPair(pairs[i].label, pairs[i].value, labelW, valW, pairs[i].style)
		if i+1 < len(pairs) {
			right := renderPair(pairs[i+1].label, pairs[i+1].value, labelW, valW, pairs[i+1].style)
			b.WriteString(left)
			b.WriteString("  ")
			b.WriteString(right)
		} else {
			b.WriteString(left)
		}
		b.WriteString("\n")
	}

	// Repo URL
	if p.Repo != "" {
		b.WriteString(DetailLabelStyle.Render("  Repo:"))
		b.WriteString(" ")
		b.WriteString(DetailValueStyle.Render(p.Repo))
		b.WriteString("\n")
	}

	// Separator
	sepWidth := m.width - 4
	if sepWidth < 1 {
		sepWidth = 1
	}
	b.WriteString(DetailSeparator.Render(strings.Repeat("─", sepWidth)))

	return b.String()
}

// detailProjectConfig renders the project configuration section.
func (m Model) detailProjectConfig(p grind.ProjectConfig) string {
	var b strings.Builder

	b.WriteString(DetailSectionStyle.Render("Project Config"))
	b.WriteString("\n")

	availWidth := m.width - 4
	if availWidth < 40 {
		availWidth = 40
	}

	labelW := 18
	valW := availWidth/2 - labelW
	if valW < 10 {
		valW = 10
	}

	// Billing info
	roundTo := p.Billing.RoundTo
	if roundTo == "" {
		roundTo = "quarter-hour"
	}
	rate := p.Billing.Rate

	pairs := []struct {
		label string
		value string
		style lipgloss.Style
	}{
		{"Type", orDash(p.Type), DetailValueStyle},
		{"Billing Rate", formatDetailDollars(rate) + "/hr", DetailValueStyle},
		{"Round To", roundTo, DetailValueStyle},
	}

	for i := 0; i < len(pairs); i += 2 {
		left := renderPair(pairs[i].label, pairs[i].value, labelW, valW, pairs[i].style)
		if i+1 < len(pairs) {
			right := renderPair(pairs[i+1].label, pairs[i+1].value, labelW, valW, pairs[i+1].style)
			b.WriteString(left)
			b.WriteString("  ")
			b.WriteString(right)
		} else {
			b.WriteString(left)
		}
		b.WriteString("\n")
	}

	// Client info
	if p.Client != nil {
		b.WriteString("\n")
		b.WriteString(DetailLabelStyle.Render("  Client:"))
		b.WriteString("\n")

		clientPairs := []struct {
			label string
			value string
		}{
			{"Contact", p.Client.Contact},
			{"Company", p.Client.Company},
			{"Email", p.Client.Email},
		}

		for _, cp := range clientPairs {
			if cp.value != "" {
				b.WriteString(DetailDimStyle.Render("    " + cp.label + ":"))
				b.WriteString(" ")
				b.WriteString(DetailValueStyle.Render(cp.value))
				b.WriteString("\n")
			}
		}
	}

	// Publications
	if len(p.Publications) > 0 {
		b.WriteString("\n")
		b.WriteString(DetailLabelStyle.Render("  Publications:"))
		b.WriteString("\n")

		for _, pub := range p.Publications {
			pubStr := pub.URL
			if pub.PublishedAt != "" {
				pubStr = pub.PublishedAt + " — " + pub.URL
			}
			b.WriteString(DetailDimStyle.Render("    • "))
			b.WriteString(DetailValueStyle.Render(runewidth.Truncate(pubStr, availWidth-6, "…")))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// sortedSessions returns sessions sorted chronologically by start time.
func sortedSessions(sessions []grind.Session) []grind.Session {
	sorted := make([]grind.Session, len(sessions))
	copy(sorted, sessions)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Start.Before(sorted[i].Start) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

// renderPair renders a label: value pair with fixed widths.
func renderPair(label, value string, labelW, valW int, valueStyle lipgloss.Style) string {
	label = DetailLabelStyle.Render(label + ":")
	label = runewidth.Truncate(label, labelW, "…")
	padded := label + strings.Repeat(" ", labelW-runewidth.StringWidth(label))

	val := runewidth.Truncate(value, valW, "…")
	padded += valueStyle.Render(val)

	return padded
}

// formatDetailHours formats hours as a human-readable string like "12.5h".
func formatDetailHours(h float64) string {
	return fmt.Sprintf("%.1fh", h)
}

// formatDetailDollars formats a dollar amount with 2 decimal places.
func formatDetailDollars(amount float64) string {
	rounded := math.Round(amount*100) / 100
	if rounded == 0 {
		return "$0.00"
	}
	return fmt.Sprintf("$%.2f", rounded)
}

// formatRoundedHours formats rounded seconds as a human-readable string like "0.25h".
func formatRoundedHours(seconds int64) string {
	hours := float64(seconds) / 3600.0
	return fmt.Sprintf("%.2fh", hours)
}

// detailAmountStyle returns the appropriate style for an amount value.
func detailAmountStyle(positive bool, longTerm bool) lipgloss.Style {
	if !positive {
		return DetailValueStyle
	}
	if longTerm {
		return DetailMutedBadgeStyle
	}
	return DetailGreenStyle
}

// orDash returns the value or "—" if empty.
func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
