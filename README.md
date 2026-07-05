# GRNDCTRL

A TUI dashboard for [Grind](https://github.com/leebrandt/grind) project workspaces.

Think `btop` / `htop` for your projects — shows all projects, time spent, billing status, open sessions, and more.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

## Build

```sh
go build -o grndctrl .
```

## Run

```sh
# Auto-detect workspace from current directory:
./grndctrl

# Or point to a specific workspace:
./grndctrl --workspace /path/to/workspace
./grndctrl -w /path/to/workspace
```

## Quit

Press `q` or `Ctrl+C`.
