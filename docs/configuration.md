# Configuration

## Config File Location

```
~/.config/hubstaff-tui/config.toml
```

The file is optional. If it does not exist, all defaults apply. You only need to specify the options you want to override.

A custom path can be passed at runtime:

```bash
hubstaff-tui --config /path/to/my-config.toml
```

## Hot-Reload

The TUI watches the config file for changes by polling its modification time on each tick. When the file changes, the new config is applied immediately — there is no need to restart the TUI. This applies to theme, keybindings, poll interval, TTL, and all other options.

## Full Reference

```toml
[hubstaff]
# Absolute path to the HubstaffCLI binary.
# Default: /Applications/Hubstaff.app/Contents/MacOS/HubstaffCLI
cli_path = "/Applications/Hubstaff.app/Contents/MacOS/HubstaffCLI"


[store]
# Cache TTL in seconds. Cached project and task lists are considered fresh
# for this duration. After expiry the UI serves the stale data immediately
# and refreshes in the background (stale-while-revalidate).
# Default: 300
ttl_seconds = 300

# Path to the SQLite database file. ~ is expanded to the home directory.
# Default: ~/.local/share/hubstaff-tui/hubstaff.db
db_path = "~/.local/share/hubstaff-tui/hubstaff.db"


[ui]
# Color theme. Supported values: "catppuccin-mocha", "plain"
# "plain" uses monochrome ASCII and works on any terminal.
# Default: "catppuccin-mocha"
theme = "catppuccin-mocha"

# Background status polling interval in seconds.
# The TUI calls the Hubstaff API every poll_interval seconds to keep the
# tracking state and live timer in sync.
# Set to 0 to disable background polling.
# Default: 30
poll_interval = 30

# Emit a terminal bell character (\a) when a tracking session starts or stops.
# Default: true
bell = true


[recent_tasks]
# Number of recently used tasks to display at the top of the task list,
# separated from the full list by a "── Recent ──" divider.
# Default: 5
max_items = 5


[keybindings]
# All values are Bubbletea key strings.
# Only specify the keys you want to change — defaults apply to omitted keys.

# Quit the TUI (or close a popup/overlay).
# Default: "esc"
quit = "esc"

# Stop the current tracking session.
# Default: "ctrl+e"
stop = "ctrl+e"

# Refresh: clear the cache and reload projects/tasks from the API.
# Default: "ctrl+r"
refresh = "ctrl+r"

# Open the fuzzy filter input for the current list.
# Default: "/"
filter = "/"

# Toggle the help overlay.
# Default: "?"
help = "?"

# Open today's summary view.
# Default: "T"
summary = "T"

# Switch focus between panes in two-pane and three-pane layout.
# Default: "tab"
switch_pane = "tab"

# Open the global task search screen (search across all projects).
# Default: "ctrl+f"
global_search = "ctrl+f"

# Open the session history view (7 days).
# Default: "H"
history = "H"
```

## Option Descriptions

### `[hubstaff]`

| Key | Type | Default | Description |
|---|---|---|---|
| `cli_path` | string | `/Applications/Hubstaff.app/Contents/MacOS/HubstaffCLI` | Absolute path to the HubstaffCLI binary shipped with the Hubstaff desktop app |

### `[store]`

| Key | Type | Default | Description |
|---|---|---|---|
| `ttl_seconds` | int | `300` | Cache TTL in seconds for project and task lists |
| `db_path` | string | `~/.local/share/hubstaff-tui/hubstaff.db` | SQLite database path; `~` is expanded |

### `[ui]`

| Key | Type | Default | Description |
|---|---|---|---|
| `theme` | string | `"catppuccin-mocha"` | Color theme; `"plain"` for monochrome ASCII fallback |
| `poll_interval` | int | `30` | Background polling interval in seconds; `0` disables |
| `bell` | bool | `true` | Terminal bell on tracking start/stop |

### `[recent_tasks]`

| Key | Type | Default | Description |
|---|---|---|---|
| `max_items` | int | `5` | Number of recent tasks shown at the top of the task list |

### `[keybindings]`

| Key | Default | Action |
|---|---|---|
| `quit` | `"esc"` | Quit / close popup or overlay |
| `stop` | `"ctrl+e"` | Stop tracking |
| `refresh` | `"ctrl+r"` | Clear cache and refresh |
| `filter` | `"/"` | Open fuzzy filter |
| `help` | `"?"` | Toggle help overlay |
| `summary` | `"T"` | Open today's summary |
| `switch_pane` | `"tab"` | Switch pane focus |
| `global_search` | `"ctrl+f"` | Open global task search |
| `history` | `"H"` | Open session history |

## Minimal Example

This is the minimum configuration needed on a non-macOS system where HubstaffCLI lives elsewhere:

```toml
[hubstaff]
cli_path = "/usr/local/bin/HubstaffCLI"
```

All other values fall back to their defaults.
