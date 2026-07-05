# Spec 7 — Ideas Triage Panel

**Priority:** 3
**Depends on:** Spec 3 (shares the same data infrastructure)
**Parallelizable:** Can be built in parallel with Spec 8

## Objective

Add a panel showing ideas from the workspace's `ideas/` directory — the same ideas visible via `grind list ideas`. Filterable by status (all, pending, rejected). Provides a quick way to see what ideas haven't been started as projects yet.

## Requirements

### Data Layer Addition (`internal/workspace/`)

```go
// Idea represents a single idea file.
type Idea struct {
    Number   int
    Filename string
    Title    string
    Rejected bool
    Created  time.Time  // parsed from filename timestamp
}

// CollectIdeas reads the ideas directory in the main worktree.
// If includeRejected is false, filters out files starting with "rejected-".
func CollectIdeas(workspaceRoot string, includeRejected bool) ([]Idea, error)
```

Idea files are timestamped filenames in `<workspace>/grind/ideas/`. Parse the title from the first `#` heading in the file content. Parse rejected status from the `rejected-` prefix.

### TUI Panel

- Press `i` from the dashboard (or `Tab` to switch to ideas panel).
- Shows a table with columns:

| # | Title | Status | Created |
|---|---|---|---|
| 0 | The Misfortune of Meaning | Project | 2026-02-01 |
| 1 | AI Dotfiles | Idea | 2026-02-03 |
| — | — | — | — |
| 3 | [REJECTED] Bad Idea | Rejected | 2026-01-15 |

- **Status**: "Project" if the idea has been turned into a project (a worktree with that idea's project exists), "Idea" if pending, "Rejected" if rejected.
- **Created**: date parsed from the idea file's timestamp.
- Rejected rows are shown in red/dim.

### Filtering

- `a` — toggle show all (including rejected)
- `r` — show only rejected
- Default: show only non-rejected ideas.

### Navigation

- Same `j`/`k` / `g`/`G` bindings as the project table.
- `Enter` on a non-rejected, non-project idea could open the idea file (stretch — would need to shell out to `$EDITOR`).
- `i` again or `Tab` switches back to the dashboard.
- `Esc` returns to the dashboard.

### Edge Cases

- **No ideas directory**: show a centered message "No ideas yet."
- **Ideas with no title** (empty file or no `#` heading): show "(untitled)".
- **Ideas that are also projects**: cross-reference with the project list (compare idea title to project idea field, or filename to project name).

## Acceptance Criteria

1. `i` switches to the ideas panel.
2. Ideas are listed with correct title, status badge, and creation date.
3. Filtering by rejected/all works correctly.
4. Ideas that have been turned into projects show "Project" status.
5. Navigation keys work the same as the dashboard.
6. `Esc` returns to the main dashboard.
7. Empty state renders correctly.
