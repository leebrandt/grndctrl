# Spec 8 — Live Refresh / File Watching

**Priority:** 3
**Depends on:** Spec 3 (needs data to refresh)
**Parallelizable:** Can be built in parallel with Spec 7

## Objective

Keep the dashboard up to date without requiring manual `r` presses. Support two refresh strategies: polling (simple, cross-platform) and file-system watching (instant, may require platform-specific code). Configurable refresh interval.

## Requirements

### Polling (Default)

- A `tea.Tick` fires every N seconds (default: 10s).
- On tick, re-run `CollectProjects` and re-query git state for all projects.
- Only update the UI if data has actually changed (compare session count, last session end time, dirty flag, commit count — a simple hash or modification time check).

### File Watching (Optional Enhancement)

- Use `github.com/fsnotify/fsnotify` or similar to watch `.project.json` files for changes.
- When a change is detected, re-read that single project's config (more efficient).
- Start/stop file watchers when projects are added/removed (worktree list changes).
- Fall back to polling if file watching fails or is unavailable.

### Refresh Indicator

- Show a subtle indicator in the top-right corner or status bar:
  - `◐` spinning while refresh is in progress
  - A timestamp of last successful refresh (e.g., `Last: 2:45 PM`)
  - `!` if last refresh failed

### Configurable Interval

- `--refresh 30` flag (or `-r`) sets the polling interval in seconds.
- `--no-watch` disables auto-refresh entirely.
- Can also be toggled at runtime (future — not in this spec).

### Data Change Detection

To avoid unnecessary re-renders, detect actual changes by:
1. Comparing the number of sessions per project.
2. Comparing the `end` field of the last session per project (since that's the most frequent change).
3. Comparing the git dirty flag.

If nothing changed, skip the view update (but could still update the "last refresh" timestamp).

### Error Handling

- If a project's `.project.json` fails to parse (malformed JSON), skip that project and show a warning indicator (e.g., `⚠` after the project name).
- If git commands fail for a project, show `?` for git-derived columns but keep the rest of the data.
- Don't crash on individual project failures — aggregate errors and show a status indicator.

## Acceptance Criteria

1. Dashboard refreshes automatically every 10 seconds by default.
2. Active session counter ticks every second (via Spec 4) but full data refresh is on 10s cadence.
3. When a `.project.json` is modified externally (e.g., via `grind save`), the dashboard reflects the change within the polling interval.
4. Refresh indicator shows in the status bar.
5. `--refresh 5` sets a 5-second interval.
6. `--no-watch` disables auto-refresh (only manual `r` works).
7. Malformed project JSON doesn't crash the app — shows a warning for that project.
8. No unnecessary re-renders when data hasn't changed.
