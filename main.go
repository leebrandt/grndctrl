package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/leebrandt/grndctrl/internal/tui"
	"github.com/leebrandt/grndctrl/internal/workspace"
)

func main() {
	var workspaceFlag string
	flag.StringVar(&workspaceFlag, "workspace", "", "Path to grind workspace root")
	flag.StringVar(&workspaceFlag, "w", "", "Path to grind workspace root (shorthand)")
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

	model := tui.NewModel(ws)
	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
