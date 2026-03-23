# Keybindings

## Projects Screen

| Key | Action |
|---|---|
| `Enter` | Open the selected project's task list |
| `Esc` | Quit the TUI (close the popup) |
| `Ctrl+E` | Stop the current tracking session |
| `Ctrl+R` | Refresh — clear cache and reload from the API |
| `/` | Open fuzzy filter input |
| `Ctrl+F` | Open global task search (search across all projects) |
| `T` | Open today's summary view |
| `H` | Open session history view (7 days) |
| `Tab` | Switch pane focus (two-pane / three-pane layout only) |
| `?` | Toggle help overlay |
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `Ctrl+C` | Force quit |

## Tasks Screen

| Key | Action |
|---|---|
| `Enter` | Start tracking the selected task (prompts for confirmation if already tracking a different task) |
| `Esc` | Back to projects |
| `Ctrl+E` | Stop the current tracking session |
| `Ctrl+R` | Refresh — clear cache and reload tasks from the API |
| `/` | Open fuzzy filter input |
| `Ctrl+F` | Open global task search |
| `T` | Open today's summary view |
| `H` | Open session history view |
| `Tab` | Switch pane focus (two-pane / three-pane layout only) |
| `?` | Toggle help overlay |
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `Ctrl+C` | Force quit |

## Summary Screen (Today's Summary)

| Key | Action |
|---|---|
| `j` / `↓` | Scroll down |
| `k` / `↑` | Scroll up |
| `T` | Close summary and return to previous screen |
| `Esc` | Close summary and return to previous screen |

## History Screen (Session History)

| Key | Action |
|---|---|
| `j` / `↓` | Scroll down |
| `k` / `↑` | Scroll up |
| `H` | Close history and return to previous screen |
| `Esc` | Close history and return to previous screen |

## Global Search Screen

| Key | Action |
|---|---|
| `Enter` | Start tracking the selected search result |
| `Ctrl+E` | Stop the current tracking session |
| `/` | Filter within the search results |
| `Esc` | Back to the previous screen |

## Quick-Switch Confirmation Prompt

When you press `Enter` on a task while already tracking a different task, a confirmation prompt appears:

| Key | Action |
|---|---|
| `y` | Confirm and switch tracking to the new task |
| `n` / `Esc` | Cancel — keep tracking the current task |

## Customizing Keybindings

All keys listed above (except `j`/`k`, arrows, and `Ctrl+C`) can be remapped in `~/.config/hubstaff-tui/config.toml`:

```toml
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

Only include the keys you want to change — defaults apply to omitted entries. Changes take effect immediately if config hot-reload is active (no restart needed).

Key values follow Bubbletea's key string format: `"ctrl+<letter>"`, `"alt+<letter>"`, or a single character such as `"/"`, `"?"`, `"T"`. Case matters for single characters (`"T"` and `"t"` are different keys).

See [configuration.md](configuration.md) for the full configuration reference.
