package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const DefaultConfigPath = "~/.config/hubstaff-tui/config.toml"

type Config struct {
	Hubstaff    HubstaffConfig    `toml:"hubstaff"`
	Store       StoreConfig       `toml:"store"`
	UI          UIConfig          `toml:"ui"`
	RecentTasks RecentTasksConfig `toml:"recent_tasks"`
	Keybindings KeybindingsConfig `toml:"keybindings"`
}

// KeybindingsConfig holds user-configurable keybindings.
// All fields default to sensible values; users only need to specify overrides.
type KeybindingsConfig struct {
	Quit         string `toml:"quit"`
	Stop         string `toml:"stop"`
	Refresh      string `toml:"refresh"`
	Filter       string `toml:"filter"`
	Help         string `toml:"help"`
	Summary      string `toml:"summary"`
	SwitchPane   string `toml:"switch_pane"`
	GlobalSearch string `toml:"global_search"`
	History      string `toml:"history"`
}

type HubstaffConfig struct {
	CLIPath string `toml:"cli_path"`
}

type StoreConfig struct {
	TTLSeconds int    `toml:"ttl_seconds"`
	DBPath     string `toml:"db_path"`
}

type UIConfig struct {
	Theme        string `toml:"theme"`
	PollInterval int    `toml:"poll_interval"` // seconds, default 30; 0 disables polling
	Bell         *bool  `toml:"bell"`          // terminal bell on start/stop, default true
}

// BellEnabled returns whether the terminal bell is enabled.
// Defaults to true if not explicitly set.
func (u UIConfig) BellEnabled() bool {
	if u.Bell == nil {
		return true
	}
	return *u.Bell
}

type RecentTasksConfig struct {
	MaxItems int `toml:"max_items"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Hubstaff: HubstaffConfig{
			CLIPath: "/Applications/Hubstaff.app/Contents/MacOS/HubstaffCLI",
		},
		Store: StoreConfig{
			TTLSeconds: 300,
			DBPath:     "~/.local/share/hubstaff-tui/hubstaff.db",
		},
		UI: UIConfig{
			Theme:        "catppuccin-mocha",
			PollInterval: 30,
		},
		RecentTasks: RecentTasksConfig{
			MaxItems: 5,
		},
		Keybindings: KeybindingsConfig{
			Quit:         "esc",
			Stop:         "ctrl+e",
			Refresh:      "ctrl+r",
			Filter:       "/",
			Help:         "?",
			Summary:      "T",
			SwitchPane:   "tab",
			GlobalSearch: "ctrl+f",
			History:      "H",
		},
	}
}

// Load reads the config from the given path. If path is empty, uses DefaultConfigPath.
// If the file does not exist, returns DefaultConfig() with no error.
// If the file exists but is invalid TOML, returns an error.
func Load(path string) (Config, error) {
	if path == "" {
		path = DefaultConfigPath
	}

	expanded, err := ExpandPath(path)
	if err != nil {
		return Config{}, fmt.Errorf("expanding config path: %w", err)
	}

	cfg := DefaultConfig()

	_, err = toml.DecodeFile(expanded, &cfg)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}
		// Also handle *os.PathError wrapping a not-exist error
		var pathErr *os.PathError
		if errors.As(err, &pathErr) && errors.Is(pathErr.Err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}
		return Config{}, fmt.Errorf("invalid config file %s: %w", expanded, err)
	}

	return cfg, nil
}

// ExpandPath expands ~ to the user's home directory.
func ExpandPath(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("getting home directory: %w", err)
		}
		return filepath.Join(home, path[1:]), nil
	}
	return path, nil
}

// ResolvedDBPath returns the expanded absolute path for the database.
func (c Config) ResolvedDBPath() (string, error) {
	return ExpandPath(c.Store.DBPath)
}
