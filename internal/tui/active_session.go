package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

var (
	activeBannerStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1a3a2a")).
				Foreground(colorGreen).
				Padding(0, 1)

	activeBannerDim = lipgloss.NewStyle().
			Background(lipgloss.Color("#1a3a2a")).
			Foreground(colorDim)

	activeBannerText = lipgloss.NewStyle().
				Background(lipgloss.Color("#1a3a2a")).
				Foreground(colorFg)

	activeBannerBold = lipgloss.NewStyle().
				Background(lipgloss.Color("#1a3a2a")).
				Foreground(colorGreen).
				Bold(true)
)

func (m Model) activeSessionBanner() string {
	if len(m.activeSessions) == 0 {
		return ""
	}

	var lines []string
	for _, s := range m.activeSessions {
		lines = append(lines, m.renderActiveSessionLine(s))
	}

	// Apply banner background to the full width
	width := m.width
	if width == 0 {
		width = 80
	}

	padded := make([]string, len(lines))
	for i, line := range lines {
		w := runewidth.StringWidth(line)
		if w < width {
			line += strings.Repeat(" ", width-w)
		}
		padded[i] = activeBannerStyle.Render(line)
	}

	return strings.Join(padded, "\n")
}

func (m Model) renderActiveSessionLine(s activeSessionInfo) string {
	indicator := " "
	if m.tickPulse {
		indicator = "▶"
	}

	elapsed := time.Since(s.start)
	durationStr := formatDuration(elapsed)

	// Unbilled amount uses rounded time (matching billing)
	roundedSeconds := roundDuration(elapsed, s.rate, s.roundTo)
	unbilledAmount := (float64(roundedSeconds) / 3600.0) * s.rate
	unbilledStr := formatBillingDollars(unbilledAmount)

	startedStr := s.start.Local().Format("3:04 PM")

	availWidth := m.width
	if availWidth == 0 {
		availWidth = 80
	}

	// Narrow terminal: skip unbilled if width < 60
	showUnbilled := availWidth >= 60

	// Truncate project name to fit
	maxNameWidth := 20
	if availWidth < 80 {
		maxNameWidth = 10
	}
	projName := runewidth.Truncate(s.projectName, maxNameWidth, "…")

	parts := []string{
		indicator,
		"Active:",
		projName,
		"|",
		"Started:",
		startedStr,
		"|",
		"Duration:",
		durationStr,
	}

	if showUnbilled {
		parts = append(parts, "|", "Unbilled:", unbilledStr)
	}

	// Build the line with styling
	var b strings.Builder
	for i, p := range parts {
		if i > 0 {
			b.WriteString(" ")
		}
		switch {
		case p == "|":
			b.WriteString(activeBannerDim.Render(p))
		case p == "▶" || p == " ":
			b.WriteString(activeBannerBold.Render(p))
		case p == "Active:":
			b.WriteString(activeBannerDim.Render(p))
		case p == projName:
			b.WriteString(activeBannerBold.Render(p))
		case p == "Started:" || p == "Duration:" || p == "Unbilled:":
			b.WriteString(activeBannerDim.Render(p))
		default:
			b.WriteString(activeBannerText.Render(p))
		}
	}

	return b.String()
}

func formatDuration(d time.Duration) string {
	totalSecs := int(d.Seconds())
	hours := totalSecs / 3600
	minutes := (totalSecs % 3600) / 60
	secs := totalSecs % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}

func roundDuration(d time.Duration, rate float64, roundTo string) int64 {
	totalSecs := int64(d.Seconds())

	var roundSecs int64
	switch roundTo {
	case "half-hour":
		roundSecs = 1800
	case "hour":
		roundSecs = 3600
	default: // "quarter-hour" or any unknown value
		roundSecs = 900
	}

	// Round up to the nearest interval
	remainder := totalSecs % roundSecs
	if remainder > 0 {
		totalSecs = totalSecs - remainder + roundSecs
	}

	return totalSecs
}

func formatBillingDollars(amount float64) string {
	rounded := math.Round(amount*100) / 100
	if rounded == 0 {
		return "$0.00"
	}
	return fmt.Sprintf("$%.2f", rounded)
}
