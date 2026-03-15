# hubstaff-tui — Research & Project Plan

> Findings from a three-agent research session on building a proper Hubstaff time tracker TUI to replace the bash+fzf script.

---

## Table of Contents

1. [Bash Script Analysis](#1-bash-script-analysis)
2. [Language & Ecosystem Research](#2-language--ecosystem-research)
3. [tmux Floating Panel Research](#3-tmux-floating-panel-research)
4. [Project Plan](#4-project-plan)

---

## 1. Bash Script Analysis

### 1.1 Core Data Model

**Status entity**
- `tracking` (boolean) — whether time is currently being tracked
- `active_project.id`, `active_project.name`, `active_project.tracked_today` (H:MM:SS)
- `active_task.id`, `active_task.name`

**Project entity**
- `id`, `name`
- Derived display indicator: `●` (active+tracking), `○` (active+stopped), ` ` (inactive)

**Task entity**
- `id`, `summary` (note: field is `summary`, not `name`)
- `project_id` / `project_name` — injected at display time, not from API

**JSON shapes:**
```json
// HubstaffCLI status
{ "tracking": bool, "active_project": { "id": "...", "name": "...", "tracked_today": "H:MM:SS" }, "active_task": { "id": "...", "name": "..." } }

// HubstaffCLI projects
{ "projects": [{ "id": "...", "name": "..." }] }

// HubstaffCLI tasks <project_id>
{ "tasks": [{ "id": "...", "summary": "..." }] }
```

### 1.2 API Surface

| Command | Args | Notes |
|---|---|---|
| `HubstaffCLI status` | none | Returns status JSON |
| `HubstaffCLI projects` | none | Returns projects JSON |
| `HubstaffCLI tasks <project_id>` | project ID | Returns tasks JSON |
| `HubstaffCLI start_task <task_id>` | task ID | Fire and forget |
| `HubstaffCLI stop` | none | Fire and forget |

- `stop` and `start_task` return values are never inspected
- All calls use `2>/dev/null` — stderr silently swallowed
- Auth is handled by the Hubstaff desktop app; no auth commands needed
- No `start <project_id>` command — you can only start tracking via `start_task`

### 1.3 Navigation Model

```
[Projects Screen]
      |
      | enter (select project)
      v
[Tasks Screen]
      |
      | enter (select task) → start_task API call
      |
      | ctrl-p → become($0) → [Projects Screen] (full process restart)
```

Back navigation is a full process restart via `become($0)` — no real stack, no state preserved.

### 1.4 Current Features

1. Projects list view with fzf fuzzy search
2. Tasks list view for a selected project with fzf fuzzy search
3. Status header: tracking indicator, active task name, today's tracked time, active project name, keybinding reference
4. Active indicator on list items (`●` / `○`)
5. Start tracking — select task with `enter`
6. Stop tracking — `ctrl-e`
7. Refresh list — `ctrl-r` (clears cache + reloads)
8. Jump to tasks — `ctrl-t` from projects screen
9. Back to projects — `ctrl-p` from tasks screen
10. Preview pane — right-side 40% panel
11. File-based caching with 5-minute TTL
12. Dependency checks at startup

### 1.5 Pain Points & Limitations

**Critical bugs:**
- `ctrl-t` is broken — `become($0 --tasks)` passes no project ID
- Tasks view `reload(...)` stubs — stop, refresh, and start-task confirmation are literally the string `"..."`, not implemented
- `ctrl-e` reload in tasks view — also a stub

**Structural limits (bash+fzf):**
- No real navigation stack — `become()` kills the process on every screen transition
- Status fetched 2–3 times per screen (never cached in-process)
- Timer is a snapshot, does not tick in real time
- Silent failures everywhere — user gets no feedback if `start_task` fails
- No global task search across projects
- `stat -f %m` is macOS-only (BSD syntax)

**UX gaps:**
- No confirmation or error feedback on start/stop
- No optimistic UI after state changes
- Empty task list has no message distinguishing "no tasks" from "API error"

### 1.6 Primary User Flow (Context Swap)

The critical workflow: *"I was tracking Task A in Project X, I need to switch to Task B in Project Y."*

Current required steps:
1. Launch script → Projects screen (3 API calls: status×2 + projects)
2. Press `ctrl-e` to stop → triggers reload (3 more API calls)
3. Navigate to Project Y → press `enter` → Tasks screen (3 API calls: status×2 + tasks)
4. Navigate to Task B → press `enter` → `start_task` fires (stub, no confirmation)
5. Trust it worked — no feedback

Ideal (not currently possible): select a task and have stop+start happen atomically with visible confirmation.

### 1.7 Keybindings Inventory

| Key | Screen | Action | Status |
|---|---|---|---|
| `enter` | Projects | Navigate into project's tasks | Working |
| `enter` | Tasks | Start tracking selected task | Partially (reload is a stub) |
| `ctrl-e` | Projects | Stop current tracking | Working |
| `ctrl-e` | Tasks | Stop current tracking | Broken (stub reload) |
| `ctrl-r` | Projects | Clear cache + refresh | Working |
| `ctrl-r` | Tasks | Refresh | Broken (stub reload) |
| `ctrl-t` | Projects | Jump to tasks | Broken (no project context) |
| `ctrl-p` | Tasks | Return to projects | Works via full restart |
| `esc` | Both | Exit / go back | fzf default |

### 1.8 State Management

There is no persistent in-process state. All state is derived from API calls at screen-load time.

| State | Source | Lifetime |
|---|---|---|
| tracking, active project/task | `HubstaffCLI status` | Per API call (never cached) |
| Projects list | `HubstaffCLI projects` + file cache | Up to 5 min |
| Tasks list (per project) | `HubstaffCLI tasks <id>` + file cache | Up to 5 min |
| Selected project ID/name | bash variable | Lifetime of single fzf session only |

On `ctrl-p` → `become($0)`: all bash variables are lost. Previously selected project is not remembered.

### 1.9 Caching Strategy

| Aspect | Detail |
|---|---|
| Cache location | `${TMPDIR:-/tmp}/hubstaff-cache/` |
| Projects file | `projects.json` |
| Tasks files | `tasks_<project_id>.json` |
| TTL | 300 seconds, checked via `stat -f %m` (macOS-only) |
| Invalidation | `ctrl-r` → `clear_cache` → `rm -f *.json` |
| Status | Never cached |
| Poison risk | If CLI call fails, empty/null is still written to cache |

### 1.10 Error States

| Error | Current Handling | Gap |
|---|---|---|
| `HubstaffCLI` not found | Startup check + exit | Good |
| `jq` not installed | Startup check + exit | Good |
| `fzf` not installed | No check | Crashes at runtime |
| `status` fails | `//` fallbacks in jq | Silently shows "not tracking" |
| `projects` fails | `2>/dev/null`, empty written to cache | Silent empty list |
| `tasks` fails | `2>/dev/null`, null guard | Silent empty list |
| `start_task` fails | `execute-silent` + `2>/dev/null` | Zero user feedback |
| `stop` fails | `execute-silent` + `2>/dev/null` | Zero user feedback |

---

## 2. Language & Ecosystem Research

### 2.1 Go Ecosystem

**Primary TUI stack: Bubbletea v2 + Bubbles + Lipgloss (charmbracelet)**

- Bubbletea v2.0.2 released March 2026 — rewritten renderer, breaking API change from v1
- 40.6k GitHub stars, 18,700+ dependents, backed by Charm (corporate)
- Elm-inspired architecture: `Init()` / `Update(msg)` / `View()` functional model
- `bubbles/list` — built-in fuzzy-filtered list with cursor navigation, status messages
- `bubbles/spinner`, `bubbles/stopwatch`, `bubbles/textinput` — ready-made components
- Lipgloss — CSS-like styling engine for colors, borders, layout

**Async model:** `tea.Cmd` — run a function in a goroutine, return a `Msg` to the update loop. Exactly right for "call external binary, parse JSON, update UI."

**Fuzzy search:** `bubbles/list` has built-in filter via `sahilm/fuzzy`. Sufficient for lists of <1000 items. No external dependency needed.

**Startup time:** 10–30ms cold start. Clears 200ms target easily.

**Binary size:** 8–15MB unoptimized, 6–10MB stripped.

**Distribution:** Single binary. goreleaser handles cross-compilation + GitHub releases + Homebrew formula generation. lazygit (74.3k stars) ships this way.

### 2.2 Rust Ecosystem

**Primary TUI stack: Ratatui + crossterm**

- Ratatui v0.30.0 (Dec 2025), 19.1k stars, 13.3k dependents
- Immediate-mode rendering — you call `frame.render_widget(widget, area)` every frame
- No built-in component library comparable to Bubbles — more widgets written from scratch
- Production-proven: Helix editor, skim, gitui, bottom/btm all use it
- `ratatui::init()` / `ratatui::restore()` pattern (v0.28+) makes setup trivial

**Fuzzy search:** `nucleo-matcher` (from Helix) — best scoring quality and performance. `skim` crate for a full embedded fuzzy picker. For 50–200 items, any option is fine.

**Async model:** tokio + `mpsc::channel`, or `std::thread` + channels. More boilerplate than `tea.Cmd`. The `asyncgit` pattern from gitui is the reference.

**Startup time:** 2–8ms cold start. Marginally faster than Go but both are fine.

**Binary size:** 1.5–4MB stripped with LTO. Noticeably smaller than Go.

**Distribution:** cargo-dist or cargo-release + Homebrew. gitui, skim, broot all ship this way.

### 2.3 Comparison Matrix

| Criterion | Go | Rust |
|---|---|---|
| Cold start | ~20ms | ~5ms |
| Dev iteration speed | Fast | Slower (borrow checker) |
| Built-in fuzzy list | `bubbles/list` ✓ | DIY + nucleo |
| Async API calls | `tea.Cmd` — 5 lines | channels — 30+ lines |
| Binary size | ~8MB stripped | ~3MB stripped |
| Terminal raw mode | Excellent | Excellent |
| Distribution | goreleaser + Homebrew | cargo-dist + Homebrew |
| Reference tool | lazygit (74.3k ★) | gitui (21.6k ★), broot |

### 2.4 Recommendation: Go + Bubbletea v2

**Rationale:**

1. Both languages clear the 200ms cold-start bar — startup time is not a differentiator here.
2. `bubbles/list` solves ~60% of the UI problem out of the box (filterable list, cursor, navigation).
3. `tea.Cmd` is the perfect model for calling an external binary asynchronously — 5 lines vs 30+.
4. `tea.Tick` gives a real-time status poller in one line.
5. lazygit (74.3k stars, built solo in Go) demonstrates the ceiling for this stack.
6. Bubbletea v2 just shipped — starting now means building on the modern API.

**Choose Rust instead if:** you're already comfortable with it and the borrow checker doesn't slow you down, or if you want sub-5ms startup and <2MB binary as hard constraints.

**Stack:**
```
bubbletea  v2.0.2  — TUI framework (Elm architecture)
bubbles    v2.x    — list, textinput, spinner, stopwatch components
lipgloss   v1.x    — styling (status bar colors, borders, layout)
goreleaser v2.x    — release automation + Homebrew formula
```

---

## 3. tmux Floating Panel Research

### 3.1 `display-popup` Mechanics

```bash
tmux display-popup [-BCEkN] [-b border-lines] [-h height] [-w width]
                   [-x position] [-y position] [-T title] [-s style]
                   [-S border-style] [-e environment] [-d start-directory]
                   [shell-command [argument ...]]
```

**Key flags:**

| Flag | Meaning |
|---|---|
| `-E` | Auto-close popup when command exits (any exit code) |
| `-EE` | Auto-close only on exit code 0 |
| `-k` | Any key dismisses popup |
| `-B` | No border |
| `-b border-lines` | Border style: `single`, `double`, `heavy`, `rounded`, `padded`, `none` |
| `-w` / `-h` | Width/height — percentage (`80%`) or absolute |
| `-x` / `-y` | Position — supports `#{popup_centre_x}`, `#{popup_centre_y}`, etc. |
| `-T` | Title string (supports tmux format strings) |
| `-s` / `-S` | Popup and border style (fg/bg colors) |
| `-e VAR=val` | Inject environment variable (repeatable) |
| `-d path` | Working directory |
| `-C` | Close any existing popup |

**Closing behavior:**
- Without `-E`: popup stays until ESC or `C-c`
- With `-E`: popup closes when command exits (any exit code) — **recommended**
- With `-EE`: popup closes only on exit 0 — keep open to show error state

**Key constraint:** While a popup is open, panes underneath are frozen (not updated).

**Popup vs. floating pane:** A popup is a transient overlay with no persistent tmux pane/window. This is the correct primitive for a quick-access tool.

### 3.2 Recommended `.tmux.conf` Binding

```bash
bind-key H display-popup \
  -E \
  -w 80% \
  -h 70% \
  -b rounded \
  -T " Hubstaff " \
  -e "HUBSTAFF_POPUP=1" \
  -d "#{pane_current_path}" \
  "hubstaff-tui"
```

This is the only line a user needs to add. The tool auto-closes when the user presses `esc` (which exits the binary).

**Passing current directory context** (useful for auto-filtering by project):
```bash
-d "#{pane_current_path}" -e "HUBSTAFF_CWD=#{pane_current_path}"
```

### 3.3 TUI App Behavior Inside Popups

**Terminal sizing:** The popup gets its own pseudo-terminal sized to `-w`/`-h`. The TUI receives correct `COLUMNS`/`LINES`. Must handle `SIGWINCH` for resize.

**Detecting popup context:** Pass an env var from the keybinding — `HUBSTAFF_POPUP=1`. Check with `os.Getenv("HUBSTAFF_POPUP")`.

**Closing the popup:** Simply `os.Exit(0)`. With `-E`, tmux closes the popup automatically. No special tmux signaling needed.

**How lazygit does it:**
lazygit has zero tmux-specific logic. Users open it as:
```bash
bind g display-popup -E -d "#{pane_current_path}" -w 90% -h 90% lazygit
```
The popup closes when lazygit exits. The TUI is completely unaware it's inside a popup.

**fzf `--tmux` lesson:** fzf calls `display-popup -E` internally when `--tmux` is set, and silently ignores it when not in tmux. This graceful-degradation pattern is worth adopting.

### 3.4 State Persistence Between Invocations

**Options:**

| Approach | Pros | Cons |
|---|---|---|
| No persistence | Zero complexity | User re-selects project every time |
| File-based state | Simple, survives restarts | Stale state if project deleted |
| In-memory daemon | Instant re-open, status bar widget | Complex, process management |

**Recommendation: file-based state.** Write on exit, read on startup.

```
~/.local/share/hubstaff-tui/state.json
{
  "last_project_id": "abc123",
  "last_project_name": "Acme Backend",
  "last_task_id": "xyz789",
  "scroll_position": 3
}
```

Reference: watson, timetrap, fzf history all use this pattern. No daemon needed.

### 3.5 Auto-configuration UX

The tool can configure tmux on first run:

```bash
hubstaff-tui setup
# Detects tmux, appends bind-key line to ~/.tmux.conf, runs: tmux source-file ~/.tmux.conf
```

Live binding (no file edit, doesn't persist across server restart):
```bash
tmux bind-key H display-popup -E -w 80% -h 70% -e HUBSTAFF_POPUP=1 "hubstaff-tui"
```

---

## 4. Project Plan

### 4.1 Project Identity

**Name:** `hubstaff-tui`  
**Language:** Go + Bubbletea v2  
**Target:** macOS arm64/amd64 primary, Linux secondary  
**Distribution:** Single binary, goreleaser → GitHub releases → Homebrew tap

### 4.2 Feature Set

| Feature | Bash script | hubstaff-tui |
|---|---|---|
| Projects list with fuzzy search | ✓ (fzf) | ✓ (bubbles/list) |
| Tasks list with fuzzy search | ✓ (fzf) | ✓ |
| Real back navigation | ✗ | ✓ (navigation stack) |
| Start tracking with feedback | ✗ (stub) | ✓ (spinner → ● confirm) |
| Stop tracking with feedback | ✓ (no confirm) | ✓ |
| Live ticking timer | ✗ (snapshot) | ✓ (bubbles/stopwatch) |
| Non-blocking API calls | ✗ (blocks UI) | ✓ (tea.Cmd goroutines) |
| Loading spinners | ✗ | ✓ (bubbles/spinner) |
| Error display | ✗ (silent) | ✓ (status bar) |
| Status poll interval | ✗ | ✓ (configurable, default 5s) |
| Last-used state restored on open | ✗ | ✓ (state.json) |
| tmux setup command | ✗ | ✓ (`hubstaff-tui setup`) |
| Global task search (cross-project) | ✗ | Phase 3 |
| Linux compatible | ✗ (BSD stat) | ✓ |
| Configurable keybindings | ✗ | Phase 2 |

### 4.3 Screen Model

```
┌─────────────────────────────────────────────┐
│ ● Acme Backend / TICKET-42                  │  ← header (status)
│ ⏱  00:47:23  |  Today: 3h 12m              │  ← live ticking
│─────────────────────────────────────────────│
│ > [fuzzy filter input]                      │  ← always focused
│─────────────────────────────────────────────│
│   ○ Acme Backend                            │  ← projects list
│   ● Acme Frontend                           │    (● = tracking)
│     Internal Tools                          │
│     ...                                     │
│─────────────────────────────────────────────│
│ enter:select  ctrl-e:stop  ctrl-r:refresh   │  ← footer
└─────────────────────────────────────────────┘

       enter on project
            ▼

┌─────────────────────────────────────────────┐
│ ● Acme Backend / TICKET-42                  │
│ ⏱  00:47:23  |  Today: 3h 12m              │
│─────────────────────────────────────────────│
│ > [fuzzy filter input]           📂 Acme    │
│─────────────────────────────────────────────│
│   TICKET-38  Fix login redirect             │
│ ● TICKET-42  Refactor auth middleware       │  ← ● = currently tracking
│   TICKET-51  Update dependencies            │
│─────────────────────────────────────────────│
│ enter:start  ctrl-e:stop  esc:back          │
└─────────────────────────────────────────────┘
```

### 4.4 Keybinding Design

| Key | Screen | Action |
|---|---|---|
| `enter` | Projects | Open tasks for selected project |
| `enter` | Tasks | Start tracking selected task (with spinner feedback) |
| `esc` / `ctrl-p` | Tasks | Back to projects |
| `esc` | Projects | Exit (closes popup) |
| `ctrl-e` | Both | Stop tracking (with confirmation) |
| `ctrl-r` | Both | Refresh (clear cache + reload) |
| `?` | Both | Toggle keybinding help overlay |
| `ctrl-c` | Both | Force exit |

### 4.5 Phase Breakdown

#### Phase 1 — MVP

- Go module init, Bubbletea v2 app skeleton
- `internal/hubstaff/` — adapter for `HubstaffCLI` binary (exec + JSON parse)
- `internal/cache/` — file-based cache with TTL (cross-platform)
- Projects list view with fuzzy filter
- Tasks list view with back navigation (real stack, not process restart)
- Start/stop tracking with spinner feedback
- Live ticking header (bubbles/stopwatch)
- Async API calls — all non-blocking
- `hubstaff-tui setup` — auto-configure tmux keybinding

#### Phase 2 — Polish

- State persistence (`~/.local/share/hubstaff-tui/state.json`) — restore last project/task on open
- Error state display — distinguish API error from empty list
- Loading spinners during initial data fetch
- Configurable poll interval and cache TTL via config file
- Keybinding help overlay (`?`)
- Linux support verified + CI matrix

#### Phase 3 — Power Features

- Global task search across all projects (background-load all project tasks)
- Time log view — today/week breakdown per project
- Quick-switch shortcut — if tracking, `enter` on a different task shows "stop X and start Y?" with single keypress confirm
- tmux status-right widget — `hubstaff-tui status` outputs a short string for status bar

### 4.6 Repository Layout

```
hubstaff-tui/
├── main.go
├── go.mod
├── go.sum
├── .goreleaser.yaml
├── internal/
│   ├── hubstaff/       # CLI adapter (exec + JSON)
│   │   ├── client.go
│   │   └── types.go
│   ├── cache/          # File-based cache with TTL
│   │   └── cache.go
│   ├── state/          # Persistent state (last project/task)
│   │   └── state.go
│   └── ui/             # Bubbletea models
│       ├── app.go      # Root model + navigation stack
│       ├── projects.go # Projects list model
│       ├── tasks.go    # Tasks list model
│       └── styles.go   # Lipgloss styles (Catppuccin theme)
└── cmd/
    └── setup.go        # `hubstaff-tui setup` tmux config command
```

### 4.7 tmux Integration Reference

**User's `.tmux.conf`** (added by `hubstaff-tui setup` or manually):
```bash
bind-key H display-popup \
  -E \
  -w 80% \
  -h 70% \
  -b rounded \
  -T " Hubstaff " \
  -e "HUBSTAFF_POPUP=1" \
  -d "#{pane_current_path}" \
  "hubstaff-tui"
```

**App exit behavior:**
- `esc` on projects screen → `os.Exit(0)` → popup closes (via `-E`)
- After `start_task` succeeds → `os.Exit(0)` → popup closes automatically
- Use `-EE` variant if you want popup to stay on errors

---

*Research conducted March 2026. Bubbletea v2.0.2, Ratatui v0.30.0, tmux 3.5.*
