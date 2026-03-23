# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-03-23

### Added
- Interactive TUI with single/two/three-pane adaptive layout
- Projects and tasks browser with fuzzy filtering (`/` key)
- Start/stop time tracking with async feedback
- Live ticking timer in header (HH:MM:SS format)
- Recent tasks shortlist (top 5, configurable) with separator
- Today's summary view (`T` key) — project/task/duration table
- 7-day session history view (`H` key) grouped by date
- Global cross-project task search (`Ctrl+F`)
- Task preview pane in three-pane mode (≥140 cols)
- Help overlay (`?` key) with all keybindings
- `hubstaff-tui status` subcommand for tmux status-right
- `hubstaff-tui setup` subcommand for auto-configuring tmux
- TOML configuration at `~/.config/hubstaff-tui/config.toml`
- Config hot-reload (no restart needed)
- Configurable keybindings via TOML
- Catppuccin Mocha theme + plain ASCII fallback
- State persistence (restores last project/task/scroll on reopen)
- SQLite local cache (stale-while-revalidate, TTL configurable)
- Background status polling (configurable interval)
- Startup dependency check (validates HubstaffCLI)
- Terminal bell on start/stop (configurable)
- Cross-compiled static binaries: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64
- GoReleaser-based release automation with Homebrew tap support

[Unreleased]: https://github.com/Nathan-ma/hubstaff-tui/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Nathan-ma/hubstaff-tui/releases/tag/v0.1.0
