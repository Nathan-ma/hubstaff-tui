# Features

## Projects and Tasks Browser

The main interface is a browsable list of your Hubstaff projects and their tasks. Selecting a project opens its task list. Both lists support fuzzy filtering: press `/` to enter a filter query and the list narrows in real time. Press `Esc` to clear the filter.

## Start and Stop Tracking

Press `Enter` on a task to start tracking it. If you are already tracking a different task, a confirmation prompt appears before switching. Press `Ctrl+E` at any time to stop the current tracking session.

## Live Ticking Timer

While tracking is active, the header shows a live timer that counts up every second. The timer is seeded from the `tracked_today` value returned by the Hubstaff API so it reflects accumulated time, not just the current session.

## Async API Calls — UI Never Blocks

Every call to the HubstaffCLI binary runs inside a `tea.Cmd` goroutine. The UI remains fully responsive while projects, tasks, and status are loading. A spinner indicates pending background operations.

## Multi-Pane Layout

The layout adapts to the terminal width automatically:

- **Single-pane** (< 100 columns): projects and tasks occupy the full screen, navigated with `Enter`/`Esc`.
- **Two-pane** (≥ 100 columns): projects on the left, tasks on the right — both visible simultaneously.
- **Three-pane** (≥ 140 columns): adds a preview pane on the right showing task details.

Use `Tab` to move focus between panes in two-pane and three-pane modes.

## Task Preview Pane

In three-pane mode (≥ 140 columns) the rightmost pane shows the currently highlighted task's name, project name, tracking status, and today's accumulated time. The preview updates as you move through the task list without requiring a separate keypress.

## Recent Tasks Shortlist

The SQLite store tracks when each task was last used. The task list prepends the most recently used tasks (default: 5) at the top under a `── Recent ──` separator, so your most common tasks are always a quick keystroke away without filtering.

## Today's Summary View

Press `T` to open the summary screen: a table of every project and task tracked today with the accumulated duration for each. The table is rendered in a scrollable viewport; use `j`/`k` to scroll. Press `T` or `Esc` to return.

## Session History View

Press `H` to open the history screen: 7 days of tracked sessions grouped by date. Each entry shows the date, project, task, and duration. Use `j`/`k` to scroll. Press `H` or `Esc` to return.

## Global Task Search

Press `Ctrl+F` to open the global search screen. Type to search across tasks in all projects simultaneously. Press `Enter` on a result to start tracking it immediately. Press `Esc` to return to the previous screen.

## Hubstaff Status Subcommand

`hubstaff-tui status` prints the current tracking state to stdout, designed for embedding in a tmux status bar:

- Tracking: `◉ 01:23:45  Mobile App › Fix login bug`
- Not tracking: `○ Not tracking`

Pass `--color` for ANSI-colored output.

## tmux Setup Subcommand

`hubstaff-tui setup` writes the popup keybinding and status bar configuration to `~/.tmux.conf` and reloads it. The popup is bound to `prefix + H` and opens the TUI in a floating window.

## Startup Dependency Check

On launch, the binary verifies that the HubstaffCLI binary exists at the configured path and is executable. If not, a clear error message is printed and the TUI does not start. This prevents confusing failures later during first API calls.

## State Persistence

When you close the TUI, the last viewed project, task, and scroll position are written to `~/.local/share/hubstaff-tui/state.json`. On the next launch, the cursor is restored to where you left off.

## Config Hot-Reload

The config file at `~/.config/hubstaff-tui/config.toml` is polled for modification time changes on each tick. If the file changes, the new config is applied immediately — theme, keybindings, poll interval, and other settings update without restarting the TUI.

## Background Status Polling

The TUI polls the Hubstaff API at a configurable interval (default 30 seconds) to keep the tracking state and timer in sync even when tracking is started or stopped from outside the TUI (e.g., the Hubstaff desktop app). Set `poll_interval = 0` to disable.

## Terminal Bell on Start/Stop

When a tracking session starts or stops, the TUI emits a terminal bell character (`\a`) so you get an audible or visual notification. This can be disabled with `bell = false` in the config.

## Configurable Keybindings

Every action key can be remapped in the `[keybindings]` section of `config.toml`. See [configuration.md](configuration.md) for the full list.

## Catppuccin Mocha Theme and Plain ASCII Fallback

The default theme uses the Catppuccin Mocha color palette, which requires a 256-color or true-color terminal. Set `theme = "plain"` for a monochrome ASCII fallback that works on any terminal.
