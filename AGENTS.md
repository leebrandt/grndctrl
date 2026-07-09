# GRNDCTRL

TUI dashboard for [Grind](https://github.com/leebrandt/grind) project workspaces — like `btop` for projects. Shows time, billing, open sessions, git state.

## Quick Facts

- **Language:** Go 1.23
- **Deps:** [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Module:** `github.com/leebrandt/grndctrl`
- **Test:** `go test ./...` (table-driven tests, table subtests, `t.Helper()` for fixtures)
- **Build:** `go build -o grndctrl .`
- **Run:** `./grndctrl` or `./grndctrl -w /path/to/workspace`

## Architecture

```
main.go                    → flag parse, workspace discovery, Bubble Tea program launch
internal/
  grind/                   → Grind data model & git helpers
    types.go               → ProjectConfig, Session, BillingConfig (+ Amount/Hours/Session helpers)
    git.go                 → CommitCount, FirstCommitDate, LastCommitDate, HasUncommittedChanges
  workspace/               → discovers workspace root via .grind.repo.git, collects projects
    workspace.go           → FindWorkspace, FindWorkspaceOrFlag, CollectProjects (git worktree list + parse .project.json)
  tui/                     → Bubble Tea UI
    model.go               → root Model (Init/Update/View), handles q/Esc/Ctrl+C quit
    styles.go              → Tokyo Night palette, Lip Gloss style vars (exported for use in main.go)
```

## Data Flow

1. `FindWorkspaceOrFlag` → walks up from CWD (or uses `--workspace`) to find `.grind.repo.git`
2. `CollectProjects` → `git worktree list` on bare repo, reads each worktree's `.project.json` → `[]grind.ProjectConfig`
3. `tui.NewModel(ws)` → Bubble Tea model renders centered title + workspace path
4. Quit with `q` / `Esc` / `Ctrl+C`

## .project.json Schema

Defined in `grind.ProjectConfig`: `name`, `type`, `idea`, `time` (`[]Session` with start/end/duration/rounded/invoiced), `billing` (roundTo/rate), `client`, `repo`, `code`, `longTerm`, `publications`.

## Code Conventions

- No comments in production code
- Tests: package-level table tests, generics helper `ptr[T]`, `t.TempDir()` for isolation, `setup*` fixtures with cleanup
- Lip Gloss styles: Tokyo Night palette (bg `#1a1b26`, accent `#7aa2f7`, green/red/yellow/gold for semantic states)
- Bubble Tea: standard Model interface, `tea.WithAltScreen()`, `tea.WindowSizeMsg` for sizing

## Spec Roadmap (under `specs/`)

1. project-scaffold ✓
2. data-layer ✓
3. main-dashboard
4. active-session-widget
5. keyboard-navigation
6. project-detail-panel
7. ideas-triage-panel
8. live-refresh

Most views are not yet implemented — TUI currently shows only a centered placeholder screen.
