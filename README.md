# hubstaff-tui

Fast Hubstaff time tracking TUI for tmux floating popups — Go + Bubbletea.

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-blue)

## What is this?

A terminal UI that replaces the bash+fzf Hubstaff tracking script. Opens as a tmux floating popup — browse projects, pick a task, start tracking, close the popup. All in ≤3 keypresses.

## Documentation

- [Architecture](docs/architecture.md) — three-layer design, multi-pane layout, caching, state persistence
- [Features](docs/features.md) — full feature list with descriptions
- [Configuration](docs/configuration.md) — complete TOML reference with all options and defaults
- [Keybindings](docs/keybindings.md) — all keys by screen, plus customization guide
- [Development](docs/development.md) — build, test, lint, package structure, CI

## Features

- Projects and tasks browser with fuzzy filtering (`/`)
- Start/stop time tracking with a single keypress (`Enter` / `Ctrl+E`)
- Live ticking timer in the header (seeded from today's tracked time)
- Async API calls — UI never blocks
- **Multi-pane layout**: single (< 100 cols), two-pane (≥ 100 cols), three-pane with task preview (≥ 140 cols)
- Task preview pane — shows task name, project, tracking status, and today's time
- Recent tasks shortlist — last 5 used tasks surfaced at the top with a separator
- Today's summary view (`T`) — table of project/task/duration, scrollable
- Session history view (`H`) — 7 days of tracked time grouped by date
- Global task search (`Ctrl+F`) — search across all projects simultaneously
- `hubstaff-tui status` subcommand for tmux status-right integration
- `hubstaff-tui setup` — writes popup keybinding and status bar to `~/.tmux.conf`
- Startup dependency check — validates HubstaffCLI before launching
- State persistence — restores last viewed project/task on reopen
- Config hot-reload — changes to `config.toml` apply without restart
- Background status polling (configurable interval, default 30s)
- Terminal bell on start/stop (configurable)
- Configurable keybindings — all keys overridable in TOML config
- Catppuccin Mocha theme with plain ASCII fallback (`theme = "plain"`)
- Single static binary, no CGO

## Installation

### From source

```bash
go install github.com/Nathan-ma/hubstaff-tui/cmd/hubstaff-tui@latest
```

### From releases

Download the latest binary from [GitHub Releases](https://github.com/Nathan-ma/hubstaff-tui/releases).

### Homebrew (coming soon)

```bash
brew install Nathan-ma/tap/hubstaff-tui
```

## Prerequisites

- [Hubstaff desktop app](https://hubstaff.com/download) installed (provides `HubstaffCLI` binary)
- tmux 3.2+ (for floating popup support)

## tmux Integration

### Popup keybinding

Add to your `~/.tmux.conf`:

```bash
# Open Hubstaff popup with prefix + H
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

Then reload: `tmux source-file ~/.tmux.conf`

**Options explained:**
| Flag | Meaning |
|---|---|
| `-E` | Auto-close popup when the TUI exits |
| `-w 80%` | Popup width (80% of terminal) |
| `-h 70%` | Popup height (70% of terminal) |
| `-b rounded` | Rounded border style |
| `-T " Hubstaff "` | Popup title |
| `-e "HUBSTAFF_POPUP=1"` | Environment variable for popup detection |

### Status bar widget

Add to your `~/.tmux.conf`:

```bash
set -g status-right '#(hubstaff-tui status) | %H:%M'
set -g status-interval 10
```

Output examples:
- Tracking: `◉ 01:23:45  Mobile App › Fix login bug`
- Not tracking: `○ Not tracking`

### Automatic setup

```bash
hubstaff-tui setup
```

This appends the popup keybinding and status bar configuration to `~/.tmux.conf` and reloads it.

## Keybindings

### Projects screen

| Key | Action |
|---|---|
| `Enter` | Open project's tasks |
| `Esc` | Quit / close popup |
| `Ctrl+E` | Stop tracking |
| `Ctrl+R` | Refresh (clear cache) |
| `/` | Fuzzy filter |
| `Ctrl+F` | Global task search |
| `T` | Today's summary |
| `H` | Session history |
| `Tab` | Switch pane focus (two/three-pane mode) |
| `?` | Toggle help overlay |
| `j`/`k` or arrows | Navigate list |
| `Ctrl+C` | Force quit |

### Tasks screen

| Key | Action |
|---|---|
| `Enter` | Start tracking selected task |
| `Esc` | Back to projects |
| `Ctrl+E` | Stop tracking |
| `Ctrl+R` | Refresh |
| `/` | Fuzzy filter |
| `Ctrl+F` | Global task search |
| `T` | Today's summary |
| `H` | Session history |
| `Tab` | Switch pane focus |
| `?` | Toggle help overlay |

### Summary / History screens

| Key | Action |
|---|---|
| `j`/`k` | Scroll |
| `Esc` or `T`/`H` | Back |

See [docs/keybindings.md](docs/keybindings.md) for all screens and customization instructions.

## Configuration

Config file: `~/.config/hubstaff-tui/config.toml`

```toml
[hubstaff]
cli_path = "/Applications/Hubstaff.app/Contents/MacOS/HubstaffCLI"

[store]
ttl_seconds = 300
db_path = "~/.local/share/hubstaff-tui/hubstaff.db"

[ui]
theme = "catppuccin-mocha"   # or "plain"
poll_interval = 30           # seconds; 0 disables background polling
bell = true

[recent_tasks]
max_items = 5

[keybindings]
quit         = "esc"
stop         = "ctrl+e"
refresh      = "ctrl+r"
filter       = "/"
help         = "?"
summary      = "T"
switch_pane  = "tab"
global_search = "ctrl+f"
history      = "H"
```

All fields are optional — defaults apply when omitted. See [docs/configuration.md](docs/configuration.md) for the full reference, or [config.example.toml](config.example.toml) for an annotated example file.

## Development

```bash
git clone https://github.com/Nathan-ma/hubstaff-tui
cd hubstaff-tui
go mod download
go build ./...
go test ./...
```

### Quality gates

```bash
go build ./...         # must compile
go vet ./...           # no vet errors
go test ./...          # all tests pass
golangci-lint run      # no lint errors (v2)
```

### Conventional commits

This project uses [conventional commits](https://www.conventionalcommits.org/). Install [lefthook](https://github.com/evilmartians/lefthook) for local git hooks:

```bash
lefthook install
```

See [docs/development.md](docs/development.md) for package structure, CI details, and cross-compilation instructions.

## License

MIT
