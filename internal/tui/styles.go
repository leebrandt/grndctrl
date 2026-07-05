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
	colorRed      = lipgloss.Color("#f7768e") // red (alert, never worked)
	colorYellow   = lipgloss.Color("#e0af68") // yellow (dirty, warning)
	colorGold     = lipgloss.Color("#ff9e64") // gold (long-term star)
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
)
