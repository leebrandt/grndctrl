package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette.
const (
	colorBg       = lipgloss.Color("#1a1b26") // dark background
	colorFg       = lipgloss.Color("#a9b1d6") // light foreground
	colorDim      = lipgloss.Color("#565f89") // dim/gray
	colorAccent   = lipgloss.Color("#7aa2f7") // blue accent
	colorGreen    = lipgloss.Color("#9ece6a") // green (active, invoiced)
	colorRed      = lipgloss.Color("#ff4466") // red (alert, never worked)
	colorYellow   = lipgloss.Color("#e0af68") // yellow (dirty, warning)
	colorGold     = lipgloss.Color("#ff9e64") // gold (long-term star)

	colorGreenMuted = lipgloss.Color("#5a7a3a") // muted green (long-term active)
	colorYellowMuted = lipgloss.Color("#8a7040") // muted yellow (long-term dirty)
	colorDimMuted   = lipgloss.Color("#444c66") // muted dim (long-term default)
)

// Base styles.
var (
	AppStyle = lipgloss.NewStyle().
		Background(colorBg).
		Foreground(colorFg).
		Padding(1, 2)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			MarginBottom(1)

	DimStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	AccentStyle = lipgloss.NewStyle().
			Foreground(colorAccent)

	GreenStyle = lipgloss.NewStyle().
			Foreground(colorGreen)

	RedStyle = lipgloss.NewStyle().
			Foreground(colorRed)

	YellowStyle = lipgloss.NewStyle().
			Foreground(colorYellow)

	GoldStyle = lipgloss.NewStyle().
			Foreground(colorGold)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Italic(true)

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Italic(true).
			MarginBottom(0)

	SelectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2f3346"))

	ActiveRowStyle = lipgloss.NewStyle().
			Foreground(colorGreen)

	DirtyRowStyle = lipgloss.NewStyle().
			Foreground(colorYellow)

	NeverWorkedStyle = lipgloss.NewStyle().
			Foreground(colorRed)

	StarStyle = lipgloss.NewStyle().
			Foreground(colorGold)

	ActiveRowMutedStyle = lipgloss.NewStyle().
				Foreground(colorGreenMuted)

	DirtyRowMutedStyle = lipgloss.NewStyle().
				Foreground(colorYellowMuted)

	LongTermDefaultStyle = lipgloss.NewStyle().
				Foreground(colorGreenMuted)

	RowStyle = lipgloss.NewStyle()

	AltRowStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1c1e2b"))

	FilterBadgeStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	FilterPromptStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorAccent).
				Padding(0, 1).
				Foreground(colorFg)

	// Detail view styles
	DetailHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent)

	DetailSectionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent).
				Underline(true)

	DetailLabelStyle = lipgloss.NewStyle().
				Foreground(colorDim)

	DetailValueStyle = lipgloss.NewStyle().
				Foreground(colorFg)

	DetailGreenStyle = lipgloss.NewStyle().
				Foreground(colorGreen)

	DetailRedStyle = lipgloss.NewStyle().
				Foreground(colorRed)

	DetailYellowStyle = lipgloss.NewStyle().
				Foreground(colorYellow)

	DetailGoldStyle = lipgloss.NewStyle().
				Foreground(colorGold)

	DetailDimStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	DetailSeparator = lipgloss.NewStyle().
				Foreground(colorDim)

	DetailTableHeaderStyle = lipgloss.NewStyle().
				Foreground(colorDim).
				Italic(true)

	DetailAltRowStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1c1e2b"))

	DetailBadgeStyle = lipgloss.NewStyle().
				Foreground(colorGreen).
				Bold(true)

	DetailMutedBadgeStyle = lipgloss.NewStyle().
				Foreground(colorGreenMuted).
				Bold(true)
)
