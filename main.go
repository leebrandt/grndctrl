package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/leebrandt/grndctrl/internal/tui"
	"github.com/leebrandt/grndctrl/internal/workspace"
)

func main() {
	var workspaceFlag string
	var refreshInterval int
	var noWatch bool
	flag.StringVar(&workspaceFlag, "workspace", "", "Path to grind workspace root")
	flag.StringVar(&workspaceFlag, "w", "", "Path to grind workspace root (shorthand)")
	flag.IntVar(&refreshInterval, "refresh", 10, "Auto-refresh interval in seconds (0 to disable)")
	flag.IntVar(&refreshInterval, "r", 10, "Auto-refresh interval in seconds (shorthand)")
	flag.BoolVar(&noWatch, "no-watch", false, "Disable auto-refresh (manual 'r' still works)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "A TUI dashboard for Grind project workspaces.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	ws, err := workspace.FindWorkspaceOrFlag(workspaceFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, tui.ErrorStyle.Render("Error:"), err.Error())
		os.Exit(1)
	}

	autoRefresh := !noWatch && refreshInterval > 0
	interval := time.Duration(refreshInterval) * time.Second

	model := tui.NewModel(ws, interval, autoRefresh)
	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
