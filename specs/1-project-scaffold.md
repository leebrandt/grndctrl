# Spec 1 — Project Scaffold & Foundation

**Priority:** 1 (foundational)
**Parallelizable:** No (others depend on this)

## Objective

Initialize the Go module, set up the Bubble Tea skeleton, establish the project directory structure, and wire a running (empty) TUI that can be quit with `q` or `Ctrl+C`. No business logic yet.

## Requirements

### Go Module

- Module path: `github.com/leebrandt/grndctrl` (adjustable)
- Go 1.22+ (use `go mod init`)
- Dependencies:
  - `github.com/charmbracelet/bubbletea`
  - `github.com/charmbracelet/lipgloss`
  - `github.com/charmbracelet/bubbles` (for prebuilt components)

### Directory Structure

```
.
├── main.go                  # Entrypoint: parse flags, start program
├── go.mod / go.sum
├── internal/
│   ├── grind/               # Data layer (Spec 2)
│   │   └── types.go
│   ├── tui/                 # TUI components
│   │   ├── model.go         # Root Bubble Tea model
│   │   └── styles.go        # Lip Gloss style definitions
│   └── workspace/           # Workspace discovery (Spec 2)
│       └── workspace.go
├── specs/                   # This directory
└── README.md                # Short: what it is, how to build/run
```

### Entrypoint (`main.go`)

- Parse a `--workspace` / `-w` flag (path to workspace root). If omitted, auto-detect by walking up from `cwd` looking for `.grind.repo.git`.
- Fallback error if neither flag nor auto-detect finds a workspace: print a clear error and exit.
- Instantiate and start the Bubble Tea program.
- Use `tea.WithAltScreen()` for full-screen rendering.

### Bubble Tea Skeleton (`internal/tui/model.go`)

- `Model` struct with at minimum a `ready bool` field.
- `Init()` returns `tea.Batch()` with any initial commands (none yet).
- `Update(msg tea.Msg) (tea.Model, tea.Cmd)`:
  - Handle `tea.KeyMsg` for `q` and `ctrl+c`: return `tea.Quit`.
  - Handle `tea.WindowSizeMsg`: store terminal dimensions.
  - Everything else: pass through.
- `View() string`: render a centered "GRNDCTRL — loading..." or similar placeholder.
- Style constants in `styles.go` using Lip Gloss (colors, padding, borders).

### Build & Run

- `go build -o grndctrl .` must succeed.
- `./grndctrl` (from within a grind workspace) shows the TUI, `q` quits.
- `./grndctrl -w /path/to/workspace` works from anywhere.

## Acceptance Criteria

1. `go build` produces a binary with no errors.
2. Running the binary inside `/home/lee/Work` auto-detects the workspace.
3. Running with `-w /home/lee/Work` from any directory auto-detects correctly.
4. Pressing `q` or `Ctrl+C` quits the TUI cleanly.
5. Terminal is restored to its original state on quit.
6. Running outside a workspace without `-w` prints a clear error message.
