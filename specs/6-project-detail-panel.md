# Spec 6 — Project Detail Panel

**Priority:** 2
**Depends on:** Spec 5 (needs selection/navigation)
**Parallelizable:** Can begin once Spec 5 is partially done (selection exists)

## Objective

When a project is selected with `Enter`, show a detailed view of that project: session history, billing summary, git timeline, project metadata. `Esc` returns to the main dashboard.

## Requirements

### View Transition

- `Enter` on a project row transitions the main content area (not a separate window) to the detail view.
- `Esc` returns to the project list (preserving scroll position and cursor).
- The detail view replaces the table entirely (not a split).

### Layout

Organized into sections, top to bottom:

#### 1. Header
- Project name (bold, large), type badge, long-term star if applicable
- Idea title (truncated to fit)
- Separator line

#### 2. Summary Row (key-value pairs in a single line or small grid)
- Total hours / Billed hours / Unbilled hours
- Total amount / Billed amount / Unbilled amount (dollar-formatted, with rate)
- Session count
- Commit count
- Separator line

#### 3. Session History Table

| # | Date | Start | End | Duration | Rounded | Invoiced? |
|---|---|---|---|---|---|---|
| 1 | 2026-02-11 | 1:49 PM | 2:03 PM | 14m | 0.25h | ✓ |
| 2 | 2026-03-07 | 5:07 AM | 8:23 AM | 3h 16m | 3.25h | — |

- Sorted chronologically by start time.
- Invoiced sessions have a green checkmark; unbilled are dim/dash.
- If there's an active session (no end), show it as the last row with a live "Active" badge.

#### 4. Git Summary
- First commit date / Last commit date
- Commit count (vs main)
- Uncommitted changes indicator (`!` / clean)
- Repo URL (if set), with a note like `gh issue list` count if available (stretch)

#### 5. Project Config (collapsible / only if `C` pressed, or always shown below)
- Type, billing rate, round-to strategy
- Client info (if present): name, company, email
- Publications (if any): URL + date

### Scrolling

- If the detail content exceeds the terminal height, the view scrolls (use `bubbles/viewport`).
- `j`/`k` scroll the viewport (not selection, since there's nothing to select in this view yet; future: select individual sessions to mark as invoiced).

### Keybindings (Detail View)

| Key | Action |
|---|---|
| `Esc` | Return to dashboard |
| `j` / `k` | Scroll down/up |
| `g` / `G` | Scroll to top/bottom |
| `Ctrl+d` / `Ctrl+u` | Half-page scroll |

### Styling

- Section headers: bold, underlined or colored.
- Table rows: alternating subtle background.
- Active/invoiced indicators: colored (green = invoiced, yellow = active, red = not invoiced).
- Monetary values: formatted with `$` and 2 decimal places.
- Time durations: human-readable (e.g., "3h 16m" not "3.27h").
- Use the Lip Gloss styles from Spec 1 consistently.

## Acceptance Criteria

1. Pressing `Enter` on a project opens its detail view.
2. All sections render with correct data.
3. Session history table shows the right sessions with correct formatting.
4. Active session (if any) shows with live indicator.
5. Billing summary math matches manual calculation from `.project.json`.
6. `Esc` returns to dashboard with preserved scroll position.
7. Scrolling works when content overflows terminal height.
8. Every field that is null/empty in the data is handled gracefully (shows `—` or hides section).
