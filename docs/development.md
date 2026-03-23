# Development

## Prerequisites

- **Go 1.24+** — the module requires `go 1.24.2` (see `go.mod`)
- **HubstaffCLI** — the Hubstaff desktop app must be installed for manual end-to-end testing; unit tests mock the CLI and do not require it
- **golangci-lint v2** — for linting (see CI section below)
- **lefthook** — for local git hooks (optional but recommended)

## Common Commands

```bash
# Build all packages
go build ./...

# Run all tests
go test ./...

# Verbose test output
go test -v ./...

# Format code
go fmt ./...

# Static analysis
go vet ./...

# Lint (requires golangci-lint v2)
golangci-lint run

# Tidy dependencies
go mod tidy
```

## Cross-Compilation

The binary is CGO-free (`CGO_ENABLED=0`), which enables clean cross-compilation:

```bash
# macOS arm64
GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -o hubstaff-tui-darwin-arm64 ./cmd/hubstaff-tui

# macOS amd64
GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -o hubstaff-tui-darwin-amd64 ./cmd/hubstaff-tui

# Linux amd64
GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -o hubstaff-tui-linux-amd64 ./cmd/hubstaff-tui

# Linux arm64
GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -o hubstaff-tui-linux-arm64 ./cmd/hubstaff-tui
```

The SQLite dependency (`modernc.org/sqlite`) is a pure-Go transpilation of SQLite and does not require any system libraries.

## Package Structure

```
cmd/hubstaff-tui/
    main.go          — entry point; parses subcommands (status, setup), launches TUI
    setup.go         — writes tmux keybinding and status bar to ~/.tmux.conf

internal/
    api/
        client.go    — HubstaffCLI adapter; wraps exec.Command with 10s timeout
        models.go    — Go types for Status, Project, Task returned by the CLI
    config/
        config.go    — TOML config struct, defaults, Load(), ExpandPath()
        watcher.go   — polls config file mtime for hot-reload
    state/
        state.go     — JSON persistence of last project/task/scroll position
    store/
        store.go     — SQLite store: cache, recents, sessions, summaries, heartbeat
    ui/
        app.go       — root AppModel: Init/Update/View, navigation, layout
        projects.go  — ProjectsModel: project list with fuzzy filter
        tasks.go     — TasksModel: task list with fuzzy filter and recent tasks
        summary.go   — SummaryModel: today's tracked time table
        history.go   — HistoryModel: 7-day session history
        preview.go   — PreviewModel: task detail pane (three-pane layout)
        search.go    — SearchModel: global cross-project task search
        help.go      — HelpModel: help overlay
        header.go    — header bar with live timer and tracking indicator
        footer.go    — footer bar with key hints
        keys.go      — KeyMap: resolved keybindings from config
        styles.go    — lipgloss styles and theme definitions
        messages.go  — tea.Msg types for async command results
```

## Coding Conventions

- Follow standard Go conventions and `gofmt` formatting.
- All blocking I/O (API calls, DB queries, file reads) must run inside `tea.Cmd` goroutines — never inside `Update()`.
- Table-driven tests are preferred for logic-heavy functions.
- Keep `Update()` handlers focused; delegate to sub-models where appropriate.
- Error messages returned to the user should be actionable (explain what to do, not just what went wrong).

## Conventional Commits

This project uses [Conventional Commits](https://www.conventionalcommits.org/). Accepted types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `ci`, `bd` (Beads issue tracker).

Install [lefthook](https://github.com/evilmartians/lefthook) to get commit-msg and pre-push hooks locally:

```bash
lefthook install
```

Hooks validate commit messages with [convco](https://github.com/convco/convco).

## CI

GitHub Actions runs on every push and pull request:

| Job | What it does |
|---|---|
| Lint | `golangci-lint run` (v2, config in `.golangci.yml`) |
| Build (ubuntu) | `go build ./...` on Linux |
| Build (macos) | `go build ./...` on macOS |
| Cross-compile | Builds `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64` with `CGO_ENABLED=0` |
| Test | `go test ./...` |

All jobs must pass before a PR can be merged to `master`.

## Linting

Configuration lives in `.golangci.yml`. The project uses golangci-lint v2. To run locally:

```bash
golangci-lint run
```

If golangci-lint is not installed, see the [golangci-lint installation guide](https://golangci-lint.run/welcome/install/).
