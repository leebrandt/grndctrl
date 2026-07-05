# Learning Go by Building a Terminal Dashboard

**A Web Dev's First Look at GRNDCTRL**

*By a 30-year web dev, for fellow web devs who never went to CS school*

---

## 1. "Wait, No Classes?" — Go's Big Ideas

If you're coming from C# or JavaScript, Go will feel familiar but stripped down. Here's the gist:

**Go has structs, not classes.** No inheritance, no constructors, no `this` keyword. You define data shapes like this:

```go
type Model struct {
    ready     bool
    width     int
    height    int
    workspace string
}
```

That's it. No `: base()`, no `public`/`private`/`protected` keywords (we'll get to visibility in a minute). If you've written TypeScript interfaces or C# records, this will click immediately.

**No exceptions.** In Go, errors are just values that functions return. You check them inline:

```go
ws, err := workspace.FindWorkspaceOrFlag(workspaceFlag)
if err != nil {
    // handle it
}
```

Coming from C# `try/catch` or JS `try/catch`, this felt weird at first. But it means you can never forget that a function can fail — the compiler won't let you ignore the return value.

**Interfaces are implicit.** A type satisfies an interface automatically if it has the right methods. You don't write `implements` or `: IWhatever`. If it walks like a duck and quacks like a duck, it's a duck.

---

## 2. The Go Toolchain — What You Actually Type

If you're used to `npm run dev` or `dotnet build`, Go's toolchain is refreshingly simple.

| What you want | Command |
|---|---|
| Compile everything | `go build ./...` |
| Compile + run | `go run .` |
| Run all tests | `go test ./...` |
| Run a single test | `go test -run TestName ./...` |
| Format your code | `go fmt ./...` |
| Add a dependency | `go get github.com/some/module` |

There is no build config file. No Webpack. No tsconfig. No project file. The code *is* the config.

**The `go.mod` file** is like `package.json` — it lists your module name, Go version, and dependencies. `go.sum` is the lockfile (like `package-lock.json`). You never edit either by hand; `go get` and `go mod tidy` manage them.

**No runtime, no VM.** Go compiles to a single static binary. On Linux, `go build .` gives you an executable called `grndctrl` — one file, nothing else. You can copy it to another machine and it just runs. Coming from C# (which needs the .NET runtime) or JS (which needs Node), this is almost magical.

---

## 3. Tour of the Codebase — File by File

Let's walk through every file in GRNDCTRL and understand what it does.

### `main.go` — The Entry Point

This is where the program starts. Every Go program has a `package main` and a `func main()`.

```go
func main() {
    var workspaceFlag string
    flag.StringVar(&workspaceFlag, "workspace", "", "Path to grind workspace root")
    flag.StringVar(&workspaceFlag, "w", "", "Path to grind workspace root (shorthand)")
    flag.Parse()

    ws, err := workspace.FindWorkspaceOrFlag(workspaceFlag)
    if err != nil {
        fmt.Fprintln(os.Stderr, tui.ErrorStyle.Render("Error:"), err.Error())
        os.Exit(1)
    }

    model := tui.NewModel(ws)
    program := tea.NewProgram(model, tea.WithAltScreen())
    if _, err := program.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

Things to notice:
- **`flag` package** handles CLI flags. Think `process.argv` but with built-in help text, default values, and parsing. `flag.Parse()` fills in the variables.
- **`err` everywhere.** Two different places can fail here (workspace discovery and program run), and both are checked.
- **`os.Stderr`** — Go makes you be explicit about where output goes. Error messages go to stderr, not stdout. This is a Unix convention that Go enforces by habit.
- **`tea.NewProgram(model)`** hands control to Bubble Tea, which runs its own event loop. More on that later.

### `internal/workspace/workspace.go` — Finding the Workspace

This file walks up directories looking for a `.grind.repo.git` folder.

```go
func FindWorkspace(startDir string) (string, error) {
    abs, err := filepath.Abs(startDir)
    // ...
    for {
        bareRepoPath := filepath.Join(current, bareRepoName)
        if info, err := os.Stat(bareRepoPath); err == nil && info.IsDir() {
            return current, nil
        }
        parent := filepath.Dir(current)
        if parent == current {
            return "", errors.New("not in a grind workspace")
        }
        current = parent
    }
}
```

Key Go-isms:
- **Multiple return values.** `(string, error)` means "return a path and possibly an error." This is the standard Go pattern.
- **`os.Stat`** checks if a file or directory exists. Combined with `info.IsDir()`, it confirms the directory is there.
- **`for { }`** is Go's `while(true)`.
- **`filepath.Join`** handles path separators cross-platform. Never concatenate paths with string interpolation.

### `internal/grind/types.go` — The Data Layer (Placeholder)

This file is essentially empty — just a package declaration and a comment.

```go
package grind

// This file is a placeholder for Spec 2.
```

In C# terms, this is a blank namespace where the data models will go. In JS terms, it's an empty module waiting for types. The `specs/` folder has the design document (`2-data-layer.md`) that describes what will go here eventually.

### `internal/tui/model.go` — The Bubble Tea State Machine

This is the heart of the TUI (Text User Interface). Bubble Tea is a Go framework based on a pattern called **The Elm Architecture** (or if you've used React, think `useReducer` on steroids).

The core idea: your UI is a **model** (state), **update** (message handler), and **view** (renderer).

```go
type Model struct {
    ready     bool
    width     int
    height    int
    workspace string
}
```

**`Init()`** — Returns the initial command (like `useEffect` on mount). Currently returns `nil` (nothing to do on startup).

```go
func (m Model) Init() tea.Cmd {
    return nil
}
```

**`Update(msg)`** — Handles messages (key presses, window resize, etc.) and returns a new state.

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "esc", "ctrl+c":
            return m, tea.Quit
        }
    case tea.WindowSizeMsg:
        m.ready = true
        m.width = msg.Width
        m.height = msg.Height
    }
    return m, nil
}
```

This is like a reducer in Redux or React's `useReducer` — you get a message and return new state. The difference is you also return a command (think "side effect" like quitting the program).

**`View()`** — Renders the screen as a string.

```go
func (m Model) View() string {
    if !m.ready {
        return "\n  Loading..."
    }
    content := lipgloss.JoinVertical(
        lipgloss.Center,
        TitleStyle.Render("GRNDCTRL"),
        "",
        DimStyle.Render("grind workspace dashboard"),
        "",
        DimStyle.Render(fmt.Sprintf("workspace: %s", m.workspace)),
    )
    return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
```

Your view is just a function that returns a string. No JSX. No DOM. Just text rendered to a terminal. Lip Gloss (`lipgloss`) is a styling library that adds colors, margins, padding, and alignment to terminal text — like CSS for the command line.

### `internal/tui/styles.go` — The Color Palette

This is where all the visual design lives:

```go
const (
    colorBg     = lipgloss.Color("#1a1b26") // dark background
    colorFg     = lipgloss.Color("#a9b1d6") // light foreground
    colorAccent = lipgloss.Color("#7aa2f7") // blue accent
    // ...
)
```

Think of this as a CSS variables file — a single source of truth for colors and reusable styles.

### `specs/` — Design Docs as Code

The `specs/` folder contains numbered Markdown files that lay out the future of the project:

1. **Project Scaffold** — what's already built
2. **Data Layer** — how to read `.project.json` files
3. **Main Dashboard** — the project list view
4. **Active Session Widget** — timer display
5. **Keyboard Navigation** — vim-like movement
6. **Project Detail Panel** — deep dive on one project
7. **Ideas Triage Panel** — quick capture
8. **Live Refresh** — auto-updating data

This is a refreshing approach: the spec lives *in the repo*, right next to the code. No Jira ticket lost to time, no Notion page nobody reads.

---

## 4. The Moving Parts — How They Fit Together

Here's the flow when you run `./grndctrl`:

```
┌──────────────┐     ┌──────────────────┐     ┌───────────────┐
│  main.go     │────→│ workspace.go     │────→│ tui/model.go  │
│  parse flags │     │ find .grind.repo │     │ create model  │
│  detect ws   │     │ return path      │     │ start tea     │
└──────────────┘     └──────────────────┘     └───────┬───────┘
                                                       │
                                                       ▼
                                              ┌──────────────────┐
                                              │ Bubble Tea Loop  │
                                              │                  │
                                              │  Key press?  ──→│ Update()
                                              │  Resize?     ──→│ Update()
                                              │  Timer tick? ──→│ Update()
                                              │                  │
                                              │  After update,  │
                                              │  call View()    │
                                              │  → render text  │
                                              └──────────────────┘
```

The separation is clean:
- **`workspace`** package deals with the filesystem (discovery)
- **`grind`** package will deal with data (parsing project files)
- **`tui`** package deals with the user interface (rendering, input)
- **`main.go`** wires them together

This is standard Go architecture: packages grouped by responsibility, with `main.go` as the thin glue.

---

## 5. Go-isms That Will Surprise a C#/JS Dev

### Capitalization = Export

In Go, a function or variable that starts with a **capital letter** is exported (public). Lowercase = unexported (private to the package).

```go
func FindWorkspace(...)    // public - can be called from outside the package
func findWorkspace(...)    // private - only visible inside the workspace package
const bareRepoName = ...   // private - lowercase
```

This replaces `public`/`private`/`export` keywords entirely. You just look at the first letter.

### Error as Return Value

In C# you might write:

```csharp
try {
    var data = File.ReadAllText(path);
    // use data
} catch (Exception ex) {
    Console.Error.WriteLine($"Failed: {ex.Message}");
}
```

In Go you write:

```go
data, err := os.ReadFile(path)
if err != nil {
    fmt.Fprintln(os.Stderr, "Failed:", err)
    return
}
```

This pattern is everywhere. You'll see `if err != nil` so often it becomes muscle memory. The advantage? Error handling is visible, not hidden in stack unwinding. Every path that can fail is explicit.

### `defer` — A Finally Block That Goes at the Call Site

```go
file, err := os.Open(path)
if err != nil {
    return err
}
defer file.Close()
// use file...
```

`defer` runs when the function returns. Unlike C# `finally` (at the end of the try block) or JS `.finally()` (at the end of the chain), `defer` is declared *right next to the resource acquisition*. You can't forget to close the file because the close call is directly below the open call.

### Pointers (But Friendlier Than C)

Go has pointers, but they're simpler than C pointers. You use `&` to get a pointer and `*` to follow one:

```go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // 'm' is a pointer to Model
    m.ready = true // Go auto-follows the pointer; no -> operator needed
}
```

The `*Model` syntax means "this function receives a pointer to Model." Think of it like C# `ref` or `out` — you're passing the original, not a copy.

The `&` in `flag.StringVar(&workspaceFlag, ...)` means "pass the address of `workspaceFlag` so the function can write into it." The `flag` package needs to modify the variable directly.

### `:=` — Short Assignment (Type Inference)

```go
ws, err := workspace.FindWorkspaceOrFlag(workspaceFlag)
```

The `:=` declares and assigns in one step. Go infers the types from the return values. It's like `var ws = ...` in C# or `const ws = ...` in JS, but shorter.

You use `=` when the variable already exists (reassigning), and `:=` when declaring a new one.

### No Async/Await

In JS, you do `await fetch(url)`. In Go with Bubble Tea, you send **commands** — messages that get processed by the event loop:

```go
func fetchData() tea.Msg {
    data, _ := http.Get("https://api.example.com")
    return dataMsg{data: data}
}

// In Update:
case dataMsg:
    m.data = msg.data
    return m, nil
```

No promises, no `async`, no threads. Bubble Tea runs your commands in a goroutine (Go's lightweight thread) and delivers the result back as a message. If you've used Redux Saga or Elm effects, this will feel familiar.

---

## 6. "Where's the Math?" — What You Don't Need

I'll be honest: when I first looked at Go, I worried it would require CS concepts I never learned in school. Here's what this codebase **doesn't** use:

- **No generics** — The current code doesn't use any. (Go has them as of 1.18, but this codebase doesn't need them.)
- **No concurrency** — Yet. Go's famous goroutines and channels aren't used yet (they'll come with Spec 8: Live Refresh).
- **No complex data structures** — Just slices (dynamic arrays) and maps (hash tables). Nothing more exotic than a `map[string]string`.
- **No algorithm puzzles** — The workspace discovery walks up a directory tree. That's a `for` loop with a `filepath.Dir()` call. No tree rotations. No big-O analysis needed.

The hardest concept in the codebase is **pointers**, and even those work the same as C# `ref` parameters.

---

## 7. Try It Yourself

```sh
# From the repo root:
go build -o grndctrl .
./grndctrl
```

If you're not in a Grind workspace, it will error. You can see this by trying:

```sh
cd /tmp && /path/to/grndctrl
# Error: not in a grind workspace: could not find .grind.repo.git
```

To see the TUI working, create a minimal grind workspace:

```sh
mkdir -p /tmp/test-workspace/.grind.repo.git
/path/to/grndctrl -w /tmp/test-workspace
```

You'll see a centered terminal screen that says:

```
             GRNDCTRL

     grind workspace dashboard

   workspace: /tmp/test-workspace

       press q or Ctrl+C to quit
```

Press `q` to exit.

---

## What's Next

This is a scaffold — a skeleton with the bones in place but no organs yet. The specs in `specs/` describe what comes next: reading project data, building a dashboard with lists and panels, adding keyboard navigation, and eventually live-refreshing data.

For a web developer learning Go, this project is an ideal playground: small enough to understand top to bottom, but real enough to teach the fundamentals — packages, error handling, structs, the standard library, and the Bubble Tea model for building interactive terminal applications.

---

*Next post in the series: Adding the Data Layer — Reading JSON files with Go's standard library. Or whatever comes next.*
