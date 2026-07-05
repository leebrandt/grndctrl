# Spec 4 — Active Session Widget

**Priority:** 1 (foundational)
**Parallelizable:** Can be built alongside Spec 3
**Depends on:** Spec 1, Spec 2

## Objective

Display a persistent banner at the top of the TUI when any project has an active (un-ended) session. Shows project name, start time, a live-ticking duration counter, and the estimated unbilled dollar amount accumulating in real time.

## Requirements

### Widget Placement

- Renders at the **very top** of the terminal, above the main content.
- If no active session exists, the banner is hidden (zero-height).
- If multiple projects have open sessions (edge case, but possible), show **each** as a separate row in the banner.

### Display Content

For each active session, show:

```
▶ Active: project-name  |  Started: 2:34 PM  |  Duration: 01:23:45  |  Unbilled: $312.50
```

- `▶` — a pulsing/animated indicator (use a simple timer to toggle between `▶` and ` ` every second, or use `bubbles/spinner`)
- **project-name** — the project's name, bold
- **Started** — local-time formatted start timestamp
- **Duration** — live-updating HH:MM:SS counter (recalculated every second)
- **Unbilled** — `(elapsed_seconds / 3600) * rate`, formatted as currency, updating every second

### Implementation

- In the `Update` loop, add a `tea.Tick(time.Second, ...)` command that fires every second while there's an active session.
- Each tick recalculates duration and unbilled amount based on current wall time vs. session start.
- The duration display should use the **actual elapsed time** (not the rounded billing time) since the user is looking at it live.
- The unbilled amount should use the **rounded time** (matching what will appear on the invoice).

### Styling

- Green background or green foreground for the active indicator.
- Separator (`|`) characters in dim style.
- The entire banner uses a distinct background color (e.g., dark green or a highlighted box) to make it pop.

### Edge Cases

- **Session ends** (`.project.json` updated with an end time): banner should disappear on the next data refresh.
- **Multiple active sessions** (if you forgot to close one and started another): show them stacked, each on its own line.
- **Terminal too narrow**: abbreviate durations (e.g., skip unbilled if width < 60).
- **Very long project name**: truncate to fit.

## Acceptance Criteria

1. When a project has an active session, the banner appears immediately after data load.
2. Counter ticks up every second (visually smooth).
3. Unbilled dollar amount increases proportionally.
4. Banner disappears when the session is ended (on next refresh).
5. If multiple active sessions exist, all are shown.
6. If no active sessions, the banner takes zero vertical space.
7. Ticking stops when there are no active sessions (no unnecessary CPU).
