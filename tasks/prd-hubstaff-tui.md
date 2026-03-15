[PRD]
# PRD: hubstaff-tui

## Overview

`hubstaff-tui` is a Go TUI application for fast Hubstaff context switching inside a tmux floating popup. It replaces an existing bash+fzf script with a polished, self-contained binary that wraps the local `HubstaffCLI` binary. The primary use case is: open a popup, see your projects/tasks at a glance, switch tracking context in one keypress, close the popup — all without leaving the terminal.

A secondary `status` subcommand prints a compact string for tmux's `status-right`, keeping the current timer visible at all times outside the popup.

---

## Goals

- Launch from tmux as a floating popup via `display-popup` with no friction
- Browse projects and tasks in a two-pane layout with fuzzy filtering
- Start/stop a Hubstaff timer in a single keypress
- Display a live ticking timer inside the popup header
- Pipe current status to tmux's status bar (`hubstaff-tui status`)
- Surface recently used tasks at the top for fast re-switching
- Show a today's summary view (time tracked per project/task)
- Send macOS/Linux desktop notifications for idle or long-session alerts
- Be fully configurable via a TOML file
- Ship as a single static binary, installable via `go install` or Homebrew

---

## Quality Gates

These checks must pass for every user story before it is considered done:

- `go build ./...` — project must compile cleanly
- `go vet ./...` — no vet errors
- `go test ./...` — all tests pass
- Manual smoke test: launch `hubstaff-tui` in a terminal and verify the described behavior

---

## User Stories

### US-001: Project scaffolding and binary entrypoint
**Description:** As a developer, I want a clean Go project structure so that the codebase is maintainable and follows standard Go conventions.

**Acceptance Criteria:**
- [ ] `go.mod` created with module path `github.com/<owner>/hubstaff-tui`
- [ ] `cmd/hubstaff-tui/main.go` as the binary entrypoint
- [ ] Internal packages under `internal/` (api, config, ui, notify)
- [ ] `go build ./...` produces a binary at `./hubstaff-tui`
- [ ] Running `./hubstaff-tui --help` prints usage without panicking
- [ ] `go vet ./...` passes clean
- [ ] `.goreleaser.yaml` skeleton present for future release automation

---

### US-002: Hubstaff CLI client (API layer)
**Description:** As a developer, I want a typed Go API client that wraps the local `HubstaffCLI` binary so that all data access goes through a single, testable interface.

**Acceptance Criteria:**
- [ ] `internal/api/client.go` defines a `Client` struct with configurable CLI path
- [ ] `internal/api/models.go` defines `Project`, `Task`, `Timer`, and `Status` structs
- [ ] Client exposes methods: `GetStatus()`, `ListProjects()`, `ListTasks(projectID)`, `StartTask(taskID)`, `Stop()`
- [ ] Each method runs the appropriate `HubstaffCLI` subcommand and parses the JSON response
- [ ] CLI stderr is captured and returned as a typed error, not silently discarded
- [ ] Unit tests exist for JSON parsing using fixture files (no real CLI calls in tests)
- [ ] Client methods return errors on non-zero exit codes or malformed JSON

---

### US-003: Disk cache with TTL
**Description:** As a user, I want projects and tasks to load instantly from cache so that the popup feels snappy even with slow API responses.

**Acceptance Criteria:**
- [ ] `internal/api/cache.go` implements a file-based cache in `~/.cache/hubstaff-tui/`
- [ ] Cache key per resource: `projects.json`, `tasks_<projectID>.json`
- [ ] Configurable TTL (default 300s); cache is considered stale after TTL
- [ ] On cache hit: return cached data immediately, trigger background refresh if stale (stale-while-revalidate)
- [ ] On cache miss: fetch from CLI, store result, return data
- [ ] `Status` is never cached — always a live call
- [ ] Cache can be busted programmatically (called on `r` keypress)
- [ ] Cache directory is created automatically if it does not exist
- [ ] Unit tests for hit/miss/stale/invalidation logic using temp dirs

---

### US-004: TOML configuration loader
**Description:** As a user, I want a config file so that I can customize the CLI path, cache TTL, keybindings, theme, and notification settings without recompiling.

**Acceptance Criteria:**
- [ ] Config is loaded from `~/.config/hubstaff-tui/config.toml`
- [ ] If the file does not exist, built-in defaults are used silently (no error)
- [ ] `internal/config/config.go` exposes a `Config` struct with all supported fields
- [ ] Supported fields:
  - `[hubstaff] cli_path` (default: `/Applications/Hubstaff.app/Contents/MacOS/HubstaffCLI`)
  - `[cache] ttl_seconds` (default: `300`), `dir` (default: `~/.cache/hubstaff-tui`)
  - `[ui] theme` (default: `catppuccin-mocha`)
  - `[notify] enabled`, `idle_reminder_minutes`, `max_session_hours`
  - `[recent_tasks] max_items` (default: `5`)
- [ ] `hubstaff-tui --config /path/to/config.toml` overrides the default path
- [ ] Invalid TOML returns a human-readable error and exits non-zero
- [ ] `config.example.toml` committed to repo with all fields documented

---

### US-005: Lipgloss theme and layout primitives
**Description:** As a developer, I want a centralized styles module so that all UI components share a consistent visual language.

**Acceptance Criteria:**
- [ ] `internal/ui/styles.go` defines all lipgloss styles for the Catppuccin Mocha palette
- [ ] Styles defined for: header bar, footer bar, active pane border, inactive pane border, selected item, tracking indicator (`●`/`○`/` `), timer text, filter input
- [ ] A `Theme` interface allows swapping palettes (Catppuccin Mocha implemented; plain ASCII fallback for no-color terminals)
- [ ] Terminal color depth is detected at startup; 256-color and true-color paths both work
- [ ] Nerd Font glyphs are used by default; a `plain` theme falls back to ASCII-only glyphs

---

### US-006: Two-pane project/task browser (core TUI)
**Description:** As a user, I want to see my projects on the left and tasks on the right so that I can browse and select without multiple navigation steps.

**Acceptance Criteria:**
- [ ] `internal/ui/model.go` implements the root bubbletea `Model` with `Update`/`View`/`Init`
- [ ] Left pane shows the project list with status indicators (`●` tracking, `○` last active, ` ` inactive)
- [ ] Right pane shows tasks for the currently focused project; auto-loads when project focus changes
- [ ] Tasks pane shows a loading spinner while tasks are being fetched
- [ ] The currently tracking task is visually highlighted in the task list
- [ ] `Tab`/`Shift+Tab` switches focus between left and right pane
- [ ] `j`/`k` and arrow keys navigate within the focused pane
- [ ] Layout adapts to terminal width: below 100 cols, collapse to single-pane drill-down mode
- [ ] Both panes have visible border labels: `Projects` and `Tasks`

---

### US-007: Embedded fuzzy filter
**Description:** As a user, I want to type to filter the focused list so that I can find a project or task instantly without scrolling.

**Acceptance Criteria:**
- [ ] Pressing `/` activates the filter input at the bottom of the focused pane
- [ ] Typing filters the list in real-time using fuzzy matching (`sahilm/fuzzy`)
- [ ] Matched characters are highlighted in the list items
- [ ] `Esc` clears the filter and returns focus to the list
- [ ] Filter state is independent per pane (project filter does not affect task filter)
- [ ] Filtering works on project name and task summary text
- [ ] Empty filter shows all items

---

### US-008: Start/stop tracking actions
**Description:** As a user, I want to start tracking a task with Enter and stop with `s` so that I can switch context without leaving the keyboard.

**Acceptance Criteria:**
- [ ] Pressing `Enter` on a task row calls `client.StartTask(taskID)` asynchronously
- [ ] While the start command is running, a spinner is shown in the task row; the UI remains responsive
- [ ] On success, the tracking indicator updates immediately and the header timer resets
- [ ] On error, a dismissible error message appears in the footer (not a crash)
- [ ] Pressing `s` calls `client.Stop()` asynchronously with the same spinner/error behavior
- [ ] `q` and `Ctrl+C` exit the TUI cleanly (restores terminal, hides cursor)
- [ ] After a successful `Start` or `Stop`, the status is re-fetched and the UI refreshes

---

### US-009: Live timer in header
**Description:** As a user, I want to see the running timer tick in real-time in the popup header so that I always know how long I've been tracking.

**Acceptance Criteria:**
- [ ] Header bar shows: tracking indicator, active task name, active project name, and elapsed time
- [ ] Elapsed time updates every second using a `time.Tick`-based bubbletea `Cmd`
- [ ] When no timer is running, header shows `○ Not tracking` with no time display
- [ ] Header reflows correctly when the terminal is resized
- [ ] Timer shows format `HH:MM:SS`
- [ ] Timer is derived from `status.active_project.tracked_today` at load time and ticks from there

---

### US-010: Recent tasks shortlist
**Description:** As a user, I want my most recently tracked tasks to appear at the top of the task list so that I can re-switch context without browsing.

**Acceptance Criteria:**
- [ ] A local LRU list of up to `config.recent_tasks.max_items` (default 5) task IDs is persisted to `~/.cache/hubstaff-tui/recents.json`
- [ ] When a task is started, its ID is prepended to the recents list
- [ ] In the task pane, recent tasks appear at the top under a `── Recent ──` separator, then all other tasks below
- [ ] If a recent task ID no longer appears in the current project's task list, it is silently omitted
- [ ] Recents list is scoped globally (across all projects), not per-project

---

### US-011: Today's summary view
**Description:** As a user, I want to see a breakdown of time tracked today so that I can review my work without leaving the terminal.

**Acceptance Criteria:**
- [ ] Pressing `T` toggles a full-screen summary view that replaces the two-pane layout
- [ ] Summary view shows a table: `Project | Task | Duration` sorted by project, then task
- [ ] A totals row at the bottom shows total time tracked today
- [ ] Data is fetched from `client.GetStatus()` and any available cache
- [ ] Pressing `T` again or `Esc` returns to the two-pane browser
- [ ] If no data is available for today, shows `No time tracked today`

---

### US-012: `hubstaff-tui status` subcommand (tmux status-bar)
**Description:** As a user, I want a subcommand that prints the current tracking state to stdout so that tmux's status-right can show my timer without opening the popup.

**Acceptance Criteria:**
- [ ] `hubstaff-tui status` prints a single line to stdout and exits
- [ ] Format when tracking: `◉ 01:23:45  Mobile App › Fix login bug`
- [ ] Format when not tracking: `○ Not tracking`
- [ ] Output uses plain text by default; `--color` flag enables ANSI color codes
- [ ] Exit code 0 in both cases (tmux requires a zero exit for status-right)
- [ ] Command completes in under 500ms (uses cache; does not wait for live API call)
- [ ] If CLI is unavailable or status call fails, prints `○ Hubstaff unavailable` and exits 0

---

### US-013: Desktop notifications
**Description:** As a user, I want to receive a desktop notification when I forget to start tracking or have been tracking too long so that I stay on top of my time.

**Acceptance Criteria:**
- [ ] `internal/notify/notify.go` wraps `osascript` (macOS) and `notify-send` (Linux) for sending notifications
- [ ] On macOS, notification uses `osascript -e 'display notification ...'`
- [ ] A background goroutine checks tracking state every `config.notify.idle_reminder_minutes` minutes
- [ ] If no timer is running for longer than the configured interval, sends: `Hubstaff — No timer running for X minutes`
- [ ] If a timer has been running for longer than `config.notify.max_session_hours`, sends: `Hubstaff — You've been tracking for X hours`
- [ ] Notifications are disabled when `config.notify.enabled = false`
- [ ] Notification goroutine runs only when the `hubstaff-tui` TUI is open (not the `status` subcommand)
- [ ] If `osascript`/`notify-send` is not found, notifications are silently skipped (no crash)

---

### US-014: Help overlay
**Description:** As a user, I want to press `?` to see all available keybindings so that I don't need to memorize them.

**Acceptance Criteria:**
- [ ] Pressing `?` renders a modal overlay on top of the two-pane layout
- [ ] Overlay lists all keybindings grouped by context: Navigation, Tracking, Views, General
- [ ] Pressing `?` again or `Esc` dismisses the overlay
- [ ] Overlay is scrollable if it exceeds terminal height
- [ ] Keybindings shown in the overlay reflect any user overrides from `config.toml` (future-proof)

---

### US-015: tmux integration docs and config snippet
**Description:** As a user, I want ready-made tmux configuration so that I can integrate the popup and status bar with a copy-paste setup.

**Acceptance Criteria:**
- [ ] `README.md` includes a **tmux integration** section
- [ ] Section provides a `~/.tmux.conf` snippet with:
  - `display-popup` bind for opening the popup (`prefix + H`)
  - `status-right` entry using `hubstaff-tui status`
- [ ] Snippet includes comments explaining each option (`-w`, `-h`, `-T`, `-E`)
- [ ] Required tmux version (3.2+) is noted
- [ ] A `Makefile` target `make install` copies the binary to `/usr/local/bin`

---

## Functional Requirements

- **FR-1:** The binary must compile to a single static binary with `CGO_ENABLED=0 go build ./...`
- **FR-2:** All Hubstaff data access must go through `internal/api/Client`; no direct CLI calls in UI code
- **FR-3:** The TUI must switch to the alternate screen buffer on start and restore it on exit
- **FR-4:** Terminal resize (`SIGWINCH`) must reflow the layout without crashing
- **FR-5:** All blocking operations (CLI calls, file I/O) must be performed inside bubbletea `Cmd` functions, never in `Update`
- **FR-6:** The `status` subcommand must not launch the TUI or the notification goroutine
- **FR-7:** Config file absence must never cause a non-zero exit or error log
- **FR-8:** The binary must be cross-compilable for `darwin/amd64`, `darwin/arm64`, and `linux/amd64`

---

## Non-Goals (Out of Scope)

- Direct Hubstaff HTTP API integration (no OAuth — CLI binary is the sole data source)
- Windows support
- Mouse support inside the TUI
- Multiple simultaneous timers
- Creating new projects or tasks from within the TUI
- Editing task descriptions
- Team member time tracking views
- Web dashboard or any non-terminal interface
- Plugin/extension system

---

## Technical Considerations

- **Bubbletea v2** (March 2026): use v2 API; do not mix v1 patterns
- **lipgloss v2** for styling (released alongside bubbletea v2)
- **bubbles** components to use: `list`, `textinput`, `spinner`, `viewport` (for summary scroll)
- **sahilm/fuzzy** for embedded fuzzy matching
- **BurntSushi/toml** for config parsing
- The `HubstaffCLI` binary is macOS-only but the config allows overriding the path for Linux compatibility
- Cache files use the same JSON structure as the CLI output to avoid transformation overhead
- The `status` subcommand must be safe to call from tmux's `status-interval` refresh loop (every 5–15s); it must not leave zombie processes

---

## Success Metrics

- Popup opens and is fully interactive in under 300ms on a warm cache
- Switching tracking context requires at most 3 keypresses from popup open to close
- `hubstaff-tui status` exits in under 500ms
- Zero terminal corruption on exit (cursor restored, alt-screen exited)
- Config file with all defaults documented passes a `go vet`-equivalent schema check

---

## Open Questions

- **Module path / GitHub username:** What org/username should the Go module path use? (e.g., `github.com/yourname/hubstaff-tui`)
- **Single-pane fallback width:** 100 columns chosen as the breakpoint for collapsing to single-pane — confirm or adjust
- **Notification granularity:** Should idle reminders repeat (every N minutes until tracking starts) or fire once?
- **Today's summary data source:** `HubstaffCLI status` returns today's total per active project — confirm whether the CLI exposes per-task breakdowns or if summary is project-level only

[/PRD]
