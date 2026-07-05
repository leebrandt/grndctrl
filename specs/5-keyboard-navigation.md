# Spec 5 — Keyboard Navigation

**Priority:** 2
**Depends on:** Spec 3 (needs a table to navigate)
**Parallelizable:** Can begin once Spec 3 is stable

## Objective

Implement full keyboard-driven navigation of the dashboard: movement, filtering, focus switching, help overlay, and consistent keybindings across views.

## Requirements

### Keybinding Table (Dashboard View)

| Key | Action |
|---|---|
| `j` / `↓` | Move cursor down one row |
| `k` / `↑` | Move cursor up one row |
| `g` | Jump to first row |
| `G` | Jump to last row |
| `Ctrl+d` / `Ctrl+u` | Page down / page up (half screen) |
| `/` | Enter filter mode (see below) |
| `Enter` | Open project detail panel (Spec 6) |
| `i` | Switch to ideas triage panel (Spec 7) |
| `r` | Force refresh data |
| `?` | Toggle help overlay |
| `q` / `Esc` | Quit |

### Filter Mode

- Pressing `/` shows a prompt at the bottom of the screen: `Filter: _`
- As you type, the table filters to show only projects whose name or type contains the typed string (case-insensitive).
- `Esc` exits filter mode and clears the filter.
- `Enter` applies the filter and exits edit mode (cursor stays in filtered list).
- `Backspace` removes characters.
- The filter prompt has its own style (e.g., box with border).
- When filter is active, a small badge shows `Filter: <text>` in the top-right corner.

### Help Overlay

- Pressing `?` toggles a centered overlay box showing all keybindings.
- The overlay has a border, title "Keyboard Shortcuts", and is semi-transparent (or just has a distinct background).
- Pressing `?` again or `Esc` closes it.
- Overlay does not resize awkwardly — min width 40 cols.

### Scrolling

- When the project list exceeds terminal height (minus header/active-session banner), the table scrolls.
- The selected row should always be visible (scroll follows cursor).
- Use Bubble Tea's built-in viewport or implement manual scroll offset tracking.

### Focus Management

- At this point there are two "views": Dashboard and Ideas. Navigation keys apply to the currently focused panel.
- `Tab` switches focus between panels (when multiple are visible).
- Focus is visually indicated (e.g., brighter border or highlighted title bar).

### Consistent Help Text

- Bottom of screen shows a mini status bar with the most important hints:
  `j/k  move  |  /  filter  |  Enter  detail  |  ?  help  |  q  quit`
- Status bar is dim, sticky at the bottom.

## Acceptance Criteria

1. All keys in the table work as specified.
2. Filtering narrows the project list in real time as you type.
3. Filter clears when `Esc` is pressed.
4. Help overlay shows all bindings and dismisses correctly.
5. Scroll follows cursor so selection is never off-screen.
6. Status bar hints are always visible at the bottom.
7. Switching to ideas panel works and `Tab` switches focus.
