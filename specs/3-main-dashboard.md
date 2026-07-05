# Spec 3 — Main Dashboard View

**Priority:** 1 (foundational)
**Parallelizable:** Can be built alongside Spec 4
**Depends on:** Spec 1, Spec 2

## Objective

Render the primary project overview table — a richer, interactive version of `grind status`. This is the default view shown at launch.

## Requirements

### Data Model (in `internal/tui/model.go` additions)

```go
type DashboardModel struct {
    projects  []grind.ProjectConfig
    workspace string
    width     int
    height    int
    cursor    int             // currently selected row index
    ready     bool
}
```

On `Init()`, run a command to load projects (via `CollectProjects`), then send them as a `tea.Msg`.

### Table Columns

| Column | Content | Alignment |
|---|---|---|
| Status | `▶` for open session, `!` for dirty, `★` for long-term, blank | left |
| Name | `project.Name` | left |
| Type | `project.Type` or `—` if unset | left |
| Worked | Total hours (e.g., `12.5h`) | right |
| Billed | Billed hours (e.g., `8.0h`) | right |
| Unbilled $ | Dollar amount of unbilled time (e.g., `$675`) | right |
| Last Session | Relative time (e.g., `3d ago`, `just now`) | left |
| Last Commit | Relative time | left |

### Color & Styling (Lip Gloss)

- **Header row**: dim/italic style, separator line below.
- **Open session** (`▶`): row highlighted in green, name in bold.
- **Dirty worktree** (`!`): row highlighted in yellow/amber.
- **Never worked** (no sessions): name shown in red/dim.
- **Unbilled amount > 0**: unbilled column in green.
- **Selected row**: inverted or highlighted background.
- **Long-term** (`★`): gold/yellow star prefix.
- Alternating row background (very subtle — 1-2 shade difference).

### Sorting

Default sort: ascending by last-session timestamp (most neglected first). Projects with zero sessions go to the top, sorted alphabetically among themselves.

### Commands

- `loadProjects()` — calls `workspace.CollectProjects`, returns `tea.Cmd` that sends `ProjectsLoadedMsg`
- `refreshProjects()` — same but re-queries (for future auto-refresh)
- On startup: show a spinner while loading (use `bubbles/spinner`)

### Edge Cases

- **Empty workspace** (no projects): show a centered message "No active projects. Create one with `grind new project`."
- **Terminal too narrow**: show a truncated table with a "→ scroll right" indicator (or just hide columns gracefully).
- **Long project names**: truncate with ellipsis to fit column width.
- **Git commands fail**: show `?` in git-derived columns rather than crashing.

## Acceptance Criteria

1. Launching GRNDCTRL shows the project table within 1 second.
2. Spinner is shown while data loads.
3. All columns display correct values matching `grind status` output.
4. Color coding is applied correctly per the rules above.
5. `j`/`k` move the selection cursor up/down (basic navigation — full keyboard spec comes later).
6. `q`/`Esc` quits from this view.
7. Empty workspace shows the friendly message (not a blank screen).
8. Table respects terminal width (no horizontal overflow into wrapping).
