# 🧑‍🏫 Smacchia Agent: Go Code Health Report

**Repo:** `github.com/leebrandt/grndctrl`
**Date:** 2026-07-09
**Scanned:** 6 source files, 3 test files (excluded per scan rules)
**Style:** Tokyo Night dashboard for Grind workspaces

---

## 1. 📊 The Metric Snapshot

| Metric | Value | Verdict |
|---|---|---|
| **Highest Cyclomatic Complexity** | **15** in function `Update` (`internal/tui/model.go:58`) | Above the Go comfort zone of < 10 |
| **Explicit Error Coverage** | **~86%** of error-returning calls handled | Solid, but some errors are silently swallowed |
| **Concurrency Check** | ✅ No issues found | Bubble Tea manages goroutines; no leaks detected |
| **Interface Count** | **0 custom interfaces** | Good restraint — no premature abstraction |
| **Code Duplication Hotspots** | **2** (see below) | `CollectProjects` and `CollectProjectInfos` share ~80% body |

### Cyclomatic Complexity — Top Offenders

| Function | Complexity | File |
|---|---|---|
| `Update` | **15** | `internal/tui/model.go:58` |
| `computeColumnWidths` | **9** | `internal/tui/dashboard.go:47` |
| `FindWorkspace` | **7** | `internal/workspace/workspace.go:22` |
| `relativeTime` | **7** | `internal/tui/dashboard.go:256` |

---

## 2. 🔍 What I Found & Why It Matters

### 🔴 Issue #1 — High Complexity in `Update` (model.go:58)

**What I noticed:**
`Update` is your Bubble Tea message dispatcher, and it packs a `switch msg.(type)` → `switch msg.String()` → three branches for `tea.KeyMsg`, one for `tea.WindowSizeMsg`, one for `ProjectsLoadedMsg`, one for `spinner.TickMsg` — all in a single function with 15 independent paths.

**Why Go prefers a different way:**
Go's philosophy is that functions should do *one thing*. A message dispatch with 15 paths is hard to unit test, hard to reason about, and easy to accidentally break when adding the next feature (like keyboard navigation into sub-panels). Flat beats nested.

**The Go Pro-Tip:**
Extract each message handler into its own method: `handleKeyMsg`, `handleWindowSizeMsg`, `handleProjectsLoaded`, `handleSpinnerTick`. This keeps `Update` as a thin router and each handler testable in isolation.

---

### 🟡 Issue #2 — Duplicated Worktree Parsing in `CollectProjects` & `CollectProjectInfos`

**What I noticed:**
Both functions in `internal/workspace/workspace.go` start with identical code:

```
git worktree list → split lines → fields → stat .project.json → read → unmarshal → skip errors
```

The only difference is that `CollectProjectInfos` also captures `branch` and `worktreePath` into a richer struct.

**Why Go prefers a different way:**
Go has no inheritance, but it *loves* composition. Duplication is the #1 maintenance smell. When you add a third variant (say `CollectProjectMap`), you will have to repeat this same 30-line block again.

**The Go Pro-Tip:**
Extract a private `collectWorktreeEntries` helper that returns a `[]worktreeEntry` struct (`{path, branch, project}`). Then both `CollectProjects` and `CollectProjectInfos` become thin wrappers that map over it. One change point, one place to fix.

---

### 🟡 Issue #3 — Silent Error Swallowing on File I/O

**What I noticed:**
In `CollectProjects` (workspace.go:101-113), three different errors are silently swallowed with `continue`:
- `os.Stat` not-exist → `continue` ✅ (by design)
- `os.ReadFile` failure (e.g. permissions) → `continue` ❌
- `json.Unmarshal` failure → `continue` ❌

**Why Go prefers a different way:**
Go's `if err != nil` convention exists precisely so you *see* every failure point. Swallowing errors makes debugging a nightmare. A permission error on `.project.json` will silently make that project disappear from the dashboard with zero feedback.

**The Go Pro-Tip:**
At minimum, log skipped projects or return them in a `SkippedProjects []string` field alongside the results. Something like:

```go
type CollectResult struct {
    Projects []grind.ProjectConfig
    Skipped  []string  // paths that were skipped and why
}
```

---

### 🟢 Issue #4 — No Interface Pollution (Good!)

**What I noticed:**
Zero custom interfaces across 6 source files. The code directly composes Bubble Tea's existing interfaces (`tea.Model`) and uses concrete types everywhere else.

**Why Go prefers this way:**
Go interfaces should be discovered, not declared upfront. You are doing it right. When (and only when) you need to test `workspace.CollectProjects` with a mock, you can extract a `gitWorktreeLister` interface then. Not before.

---

## 3. 📝 Conceptual "Before & After" Lesson

### On Silent Error Skipping (Issue #3)

```go
// ❌ The Current Pattern — The caller has no idea what was skipped
func CollectProjects(root string) ([]grind.ProjectConfig, error) {
    // ...
    for _, line := range lines {
        data, err := os.ReadFile(projectFilePath)
        if err != nil {
            continue // 🔴 silently drops permission-denied files
        }
        var project grind.ProjectConfig
        if err := json.Unmarshal(data, &project); err != nil {
            continue // 🔴 silently drops malformed JSON
        }
        projects = append(projects, project)
    }
    return projects, nil // caller thinks everything is fine
}
```

```go
// ✅ The Go Way — Be explicit about what was skipped
type CollectResult struct {
    Projects []grind.ProjectConfig
    Skipped  []string // human-readable reasons
}

func CollectProjects(root string) (*CollectResult, error) {
    // ...
    var result CollectResult
    for _, line := range lines {
        data, err := os.ReadFile(projectFilePath)
        if err != nil {
            result.Skipped = append(result.Skipped,
                fmt.Sprintf("%s: %v", projectFilePath, err))
            continue
        }
        var project grind.ProjectConfig
        if err := json.Unmarshal(data, &project); err != nil {
            result.Skipped = append(result.Skipped,
                fmt.Sprintf("%s: bad JSON: %v", projectFilePath, err))
            continue
        }
        result.Projects = append(result.Projects, project)
    }
    return &result, nil
}
```

**Why this matters:** Now the TUI can show a dimmed "2 projects skipped" footer, and the developer debugging a missing project gets instant feedback instead of head-scratching.

---

### On Cyclomatic Complexity in `Update` (Issue #1)

```go
// ❌ The Current Pattern — 15 paths, all in one function
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "esc", "ctrl+c": return m, tea.Quit
        case "j", "down":          // cursor down
        case "k", "up":            // cursor up
        }
    case tea.WindowSizeMsg:        // resize
    case ProjectsLoadedMsg:        // data arrived
    case spinner.TickMsg:          // spinner animation
    }
    return m, nil
}
```

```go
// ✅ The Go Way — Router + Handlers
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:         return m.handleKey(msg)
    case tea.WindowSizeMsg:  return m.handleWindowSize(msg)
    case ProjectsLoadedMsg:  return m.handleProjectsLoaded(msg)
    case spinner.TickMsg:    return m.handleSpinnerTick(msg)
    default:                 return m, nil
    }
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "q", "esc", "ctrl+c": return m, tea.Quit
    case "j", "down":          return m.handleCursorDown()
    case "k", "up":            return m.handleCursorUp()
    default:                   return m, nil
    }
}
// Each handler is ≤ 5 paths and independently testable.
```

**Why this matters:** When you add spec #5 (keyboard navigation) and #6 (project detail panel), these handlers grow independently without the complexity ballooning exponentially.

---

## 4. ✅ What Is Already Great

| Practice | Where | Why It Is Good |
|---|---|---|
| **Error wrapping with `%w`** | All `git.go`, `workspace.go` | Preserves error chain for `errors.Is`/`errors.As` |
| **Table-driven tests** | All `_test.go` files | Idiomatic Go, easy to add cases |
| **`t.Helper()`** | `setupBareRepo`, `setupWorktree`, `setupCollectProjectsFixture` | Clean test failure traces |
| **`t.TempDir()`** | `TestFindWorkspace`, `TestFindWorkspaceOrFlag` | Auto-cleanup, no leaks |
| **Generic `ptr[T]` helper** | `types_test.go` | Very idiomatic Go 1.23 |
| **No naked returns** | All files | Explicitness > brevity |
| **Meaningful zero values** | `ProjectConfig{}` works with `Time: nil` | No constructor ceremony |
| **No premature interfaces** | Zero custom interfaces | Postponing abstraction until it is needed |

---

## 5. 🏁 Final Thoughts

This is a **clean, well-structured Go codebase** that follows most idioms correctly. The two main things to watch as you build out specs 4–8:

1. **Do not let `Update` grow unchecked** — Extract handlers now before the complexity compounds.
2. **Decouple worktree parsing** — Unify `CollectProjects`/`CollectProjectInfos` behind a shared helper before a third variant appears.

The architecture (`main.go` → `workspace` discovery → `grind` data model → `tui` rendering) is textbook Go package layering. Keep it flat, keep errors visible, keep interfaces discovered organically. 🚀
