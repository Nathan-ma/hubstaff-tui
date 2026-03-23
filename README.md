# hubstaff-tui

Fast Hubstaff time tracking TUI for tmux floating popups — Go + Bubbletea.

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-blue)

## What is this?

A terminal UI that replaces the bash+fzf Hubstaff tracking script. Opens as a tmux floating popup — browse projects, pick a task, start tracking, close the popup. All in ≤3 keypresses.

## Features

- Projects and tasks browser with fuzzy filtering
- Start/stop time tracking in a single keypress
- Live ticking timer in the header
- Async API calls — UI never blocks
- Catppuccin Mocha theme (with plain ASCII fallback)
- SQLite local cache with stale-while-revalidate
- TOML configuration
- Single static binary (no CGO)

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
- Tracking: `◉ 01:23:45  Mobile App > Fix login bug`
- Not tracking: `○ Not tracking`

### Automatic setup

```bash
hubstaff-tui setup
```

This appends the popup keybinding to `~/.tmux.conf` and reloads it.

## Keybindings

| Key | Projects screen | Tasks screen |
|---|---|---|
| `Enter` | Open project's tasks | Start tracking task |
| `Esc` | Quit (close popup) | Back to projects |
| `Ctrl+E` | Stop tracking | Stop tracking |
| `Ctrl+R` | Refresh (clear cache) | Refresh |
| `/` | Fuzzy filter | Fuzzy filter |
| `j`/`k` | Navigate list | Navigate list |
| `Ctrl+C` | Force quit | Force quit |

## Configuration

Config file: `~/.config/hubstaff-tui/config.toml`

```toml
[hubstaff]
cli_path = "/Applications/Hubstaff.app/Contents/MacOS/HubstaffCLI"

[store]
ttl_seconds = 300
db_path = "~/.local/share/hubstaff-tui/hubstaff.db"

[ui]
theme = "catppuccin-mocha"  # or "plain"

[recent_tasks]
max_items = 5
```

All fields are optional — defaults apply when omitted. See [config.example.toml](config.example.toml).

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
go build ./...   # must compile
go vet ./...     # no vet errors
go test ./...    # all tests pass
```

### Conventional commits

This project uses [conventional commits](https://www.conventionalcommits.org/). Install [lefthook](https://github.com/evilmartians/lefthook) for local git hooks:

```bash
lefthook install
```

## License

MIT
