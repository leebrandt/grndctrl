# Spec 2 — Data Layer

**Priority:** 1 (foundational)
**Parallelizable:** No (others depend on this)
**Depends on:** Spec 1

## Objective

Implement Go types mirroring the GrindCLI `.project.json` schema, a reader for project configs, workspace discovery (walking up from cwd or using an explicit path), and git query functions (commit count, first/last commit date, dirty worktree check). All in a well-tested `internal/` package.

## Requirements

### Types (`internal/grind/types.go`)

Match the GrindCLI types exactly:

```go
package grind

import "time"

type Session struct {
    Start    time.Time `json:"start"`
    End      *time.Time `json:"end,omitempty"`
    Duration int64     `json:"duration"`     // seconds
    Rounded  int64     `json:"rounded"`      // rounded seconds
    Invoiced *bool     `json:"invoiced,omitempty"`
}

type BillingConfig struct {
    RoundTo string  `json:"roundTo"` // "quarter-hour" | "half-hour" | "hour"
    Rate    float64 `json:"rate"`
}

type ClientInfo struct {
    Contact string `json:"contact,omitempty"`
    Company string `json:"company,omitempty"`
    Address string `json:"address,omitempty"`
    Phone   string `json:"phone,omitempty"`
    Email   string `json:"email,omitempty"`
}

type Publication struct {
    URL         string `json:"url"`
    PublishedAt string `json:"publishedAt"`
}

type ProjectConfig struct {
    Name         string       `json:"name"`
    Type         string       `json:"type,omitempty"`
    Idea         string       `json:"idea"`
    Time         []Session    `json:"time"`
    Billing      BillingConfig `json:"billing"`
    Client       *ClientInfo  `json:"client,omitempty"`
    Repo         string       `json:"repo,omitempty"`
    Code         string       `json:"code,omitempty"`
    LongTerm     bool         `json:"longTerm,omitempty"`
    Publications []Publication `json:"publications,omitempty"`
}
```

Also add computed/derived fields via methods:

- `ProjectConfig.TotalSeconds() int64` — sum of all `session.Rounded`
- `ProjectConfig.BilledSeconds() int64` — sum of all `session.Rounded` where `Invoiced == true`
- `ProjectConfig.UnbilledSeconds() int64` — `TotalSeconds() - BilledSeconds()`
- `ProjectConfig.TotalHours() float64` / `BilledHours() / UnbilledHours()`
- `ProjectConfig.TotalAmount() float64` etc. (hours * rate)
- `ProjectConfig.ActiveSession() *Session` — returns the session with `End == nil`, if any
- `ProjectConfig.LastSession() *Session` — returns the most recent `Session` by `Start` time, or nil
- `Session.DurationHuman() string` — e.g., "2h 15m"

### Workspace Discovery (`internal/workspace/workspace.go`)

```go
// FindWorkspace walks up from startDir looking for .grind.repo.git.
// Returns the workspace root directory, or an error if not found.
func FindWorkspace(startDir string) (string, error)

// CollectProjects reads all active project worktrees (via git worktree list)
// and loads their .project.json files.
func CollectProjects(workspaceRoot string) ([]grind.ProjectConfig, error)

// FindWorkspaceOrFlag checks the --workspace flag first, then auto-detects.
func FindWorkspaceOrFlag(flagValue string) (string, error)
```

Git queries (use `os/exec` to shell out to `git`):

```go
// CommitCount returns commits on branch that aren't in main.
func CommitCount(bareRepo, branch string) (int, error)

// FirstCommitDate returns the ISO datetime of the first commit on branch.
func FirstCommitDate(bareRepo, branch string) (string, error)

// LastCommitDate returns the ISO datetime of the most recent commit on branch.
func LastCommitDate(bareRepo, branch string) (string, error)

// HasUncommittedChanges checks if the worktree has dirty files.
func HasUncommittedChanges(worktreePath string) (bool, error)
```

Use `git -C <path> <command>` just like GrindCLI does. Match the same `--not main` logic for commit counting.

### Testing

- Write table-driven unit tests for all math methods (TotalSeconds, BilledSeconds, etc.).
- Write tests for workspace discovery with a mock/temp directory structure.
- Use `testing.T` with temp dirs — no external dependencies.

## Acceptance Criteria

1. Reading a real `.project.json` file produces a correctly-typed `ProjectConfig`.
2. `TotalSeconds` / `BilledSeconds` / `UnbilledSeconds` match manual calculation from test fixtures.
3. `ActiveSession()` returns the correct session (or nil) given test data.
4. `LastSession()` returns the right session sorted by start time.
5. `FindWorkspace("/home/lee/Work")` returns `/home/lee/Work`.
6. `CollectProjects` returns the same projects as `grind status` for the same workspace.
7. All git query functions return sensible results (or zero/nil for empty repos).
8. `go test ./...` passes.
