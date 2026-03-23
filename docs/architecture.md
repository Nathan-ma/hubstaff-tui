# Architecture

## Overview

hubstaff-tui is structured as three distinct layers that communicate through well-defined boundaries:

```
┌─────────────────────────────────────────────┐
│  UI Layer (Bubbletea v2 Elm architecture)   │
│  internal/ui/                               │
├─────────────────────────────────────────────┤
│  Domain / Store Layer (SQLite, TTL cache)   │
│  internal/store/  internal/state/           │
├─────────────────────────────────────────────┤
│  API Adapter (wraps HubstaffCLI binary)     │
│  internal/api/                              │
└─────────────────────────────────────────────┘
```

## Three-Layer Architecture

### UI Layer — `internal/ui/`

Built on [Bubbletea v2](https://github.com/charmbracelet/bubbletea) using the Elm architecture: every model implements `Init() tea.Cmd`, `Update(tea.Msg) (tea.Model, tea.Cmd)`, and `View() string`. All rendering is done with [lipgloss v1](https://github.com/charmbracelet/lipgloss).

The root model is `AppModel`, which owns all sub-models and routes messages:

| Sub-model | Purpose |
|---|---|
| `ProjectsModel` | Scrollable project list with fuzzy filter |
| `TasksModel` | Scrollable task list with fuzzy filter |
| `SummaryModel` | Today's tracked time table (viewport scrollable) |
| `HistoryModel` | 7-day session history grouped by date |
| `PreviewModel` | Task detail panel shown in three-pane layout |
| `SearchModel` | Global cross-project task search |
| `HelpModel` | Help overlay rendered on top of current screen |

**Non-negotiable UI rules:**

- All blocking I/O (API calls, DB queries) must run inside `tea.Cmd` goroutines — never inside `Update()`.
- State transitions happen only in `Update()` via `tea.Msg` values.

### Domain / Store Layer — `internal/store/`, `internal/state/`

**`internal/store/`** — SQLite database using [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGO). The store holds:

- Cached projects and tasks (stale-while-revalidate, default TTL 300 seconds)
- Recent task usage (last-used timestamps, surfaced at top of task lists)
- Session summaries for today's summary and history views
- Heartbeat timestamp for the status subcommand

**`internal/state/`** — Lightweight JSON file at `~/.local/share/hubstaff-tui/state.json`. Persists the last viewed project ID, project name, task ID, and scroll position so the TUI reopens where you left off.

### API Adapter Layer — `internal/api/`

Wraps the HubstaffCLI binary using `exec.Command`. All commands run with a 10-second timeout. The adapter translates CLI output (JSON) into Go types (`Status`, `Project`, `Task`). It exposes:

- `GetStatus` — current tracking state, active project/task, time tracked today
- `ListProjects` — all accessible projects
- `ListTasks` — tasks for a given project ID
- `StartTracking` / `StopTracking` — start or stop time tracking
- `CheckCLI` — validates the binary exists and is executable at startup

## Multi-Pane Layout

The UI dynamically selects a layout based on the terminal width:

| Terminal width | Layout | Panes shown |
|---|---|---|
| `< 100` columns | Single-pane | Projects or Tasks (one at a time) |
| `≥ 100` columns | Two-pane | Projects on left, Tasks on right |
| `≥ 140` columns | Three-pane | Projects \| Tasks \| Preview |

Width thresholds are defined as constants in `internal/ui/app.go`:

```go
const minTwoPaneWidth  = 100
const minThreePaneWidth = 140
```

In two-pane and three-pane modes, `Tab` moves focus between panes. The preview pane shows the selected task's name, project, tracking status, and today's accumulated time.

## Stack-Based Navigation

The `screen` enum in `AppModel` acts as a navigation stack:

```
screenProjects
    └── screenTasks
            ├── screenSummary       (T key)
            ├── screenHistory       (H key)
            └── screenGlobalSearch  (Ctrl+F)
```

`Esc` moves back one level. The `previousScreen` field is stored so global search knows where to return on exit.

## Stale-While-Revalidate Caching

When a list is requested:

1. If cached data exists and is within the TTL, return it immediately.
2. If cached data is stale (or absent), return whatever is available and trigger a background refresh via a `tea.Cmd`.
3. The background fetch updates the SQLite cache, then sends a message to the UI with fresh data.

The TTL defaults to 300 seconds and is configurable via `[store] ttl_seconds` in `config.toml`. `Ctrl+R` forcibly clears the cache and triggers an immediate refresh.

## State Persistence

On every navigation change and on quit, `AppModel` calls `state.Save()` to write the current project ID, task ID, and scroll position to `~/.local/share/hubstaff-tui/state.json`. On the next launch, `state.Load()` restores this state so the cursor is positioned where the user left off.

## Config Hot-Reload

`config.Watcher` polls the config file's modification time on each Bubbletea tick. If the file's `mtime` has advanced since the last check, the config is re-read from disk and applied to the running model — changing the theme, keybindings, poll interval, or other settings without requiring a restart.

The watcher uses `os.Stat` (not `inotify`/`FSEvents`) to stay cross-platform and CGO-free.

## Background Status Polling

When `[ui] poll_interval` is greater than 0 (default 30 seconds), `AppModel.Init()` schedules a recurring `tea.Cmd` that calls `GetStatus` every interval. This keeps the tracking indicator and live timer current even if the user starts or stops tracking from outside the TUI (e.g., from the Hubstaff desktop app).

Set `poll_interval = 0` in the config to disable background polling.

## Dependencies

| Dependency | Version | Purpose |
|---|---|---|
| Go | 1.24+ | Language runtime |
| `charmbracelet/bubbletea` | v1.3.10 | TUI framework (Elm architecture) |
| `charmbracelet/lipgloss` | v1.1.0 | Styling and layout |
| `charmbracelet/bubbles` | v1.0.0 | List, spinner, viewport components |
| `modernc.org/sqlite` | v1.46.1 | Pure-Go SQLite (no CGO) |
| `BurntSushi/toml` | v1.6.0 | TOML config parsing |

## Non-Negotiables

- **No CGO** (`CGO_ENABLED=0`). The binary must be a fully static pure-Go binary. The SQLite driver (`modernc.org/sqlite`) is a C-to-Go transpilation that requires no system libraries.
- **No blocking in `Update()`**. Every call that touches the filesystem, network, or spawns a subprocess must be wrapped in a `tea.Cmd` and run in a goroutine.
- **Bubbletea v2 Elm architecture**. All models implement `Init / Update / View`. No shared mutable state between models outside of `AppModel`.

## Catppuccin Mocha Theme

Colors are defined in `internal/ui/styles.go` using the [Catppuccin Mocha](https://github.com/catppuccin/catppuccin) palette. Setting `theme = "plain"` in the config switches to a monochrome ASCII fallback that works on any terminal without true-color support.
