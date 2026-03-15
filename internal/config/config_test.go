package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Hubstaff.CLIPath != "/Applications/Hubstaff.app/Contents/MacOS/HubstaffCLI" {
		t.Errorf("unexpected CLIPath: %s", cfg.Hubstaff.CLIPath)
	}
	if cfg.Store.TTLSeconds != 300 {
		t.Errorf("unexpected TTLSeconds: %d", cfg.Store.TTLSeconds)
	}
	if cfg.Store.DBPath != "~/.local/share/hubstaff-tui/hubstaff.db" {
		t.Errorf("unexpected DBPath: %s", cfg.Store.DBPath)
	}
	if cfg.UI.Theme != "catppuccin-mocha" {
		t.Errorf("unexpected Theme: %s", cfg.UI.Theme)
	}
	if cfg.RecentTasks.MaxItems != 5 {
		t.Errorf("unexpected MaxItems: %d", cfg.RecentTasks.MaxItems)
	}
}

func TestLoad_FileNotExist(t *testing.T) {
	cfg, err := Load("/tmp/nonexistent-hubstaff-tui-config-12345.toml")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}

	defaults := DefaultConfig()
	if cfg != defaults {
		t.Errorf("expected defaults when file missing, got: %+v", cfg)
	}
}

func TestLoad_ValidTOML(t *testing.T) {
	content := `
[hubstaff]
cli_path = "/usr/local/bin/hubstaff"

[store]
ttl_seconds = 600
db_path = "/tmp/test.db"

[ui]
theme = "plain"

[recent_tasks]
max_items = 10
`
	tmpFile := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Hubstaff.CLIPath != "/usr/local/bin/hubstaff" {
		t.Errorf("unexpected CLIPath: %s", cfg.Hubstaff.CLIPath)
	}
	if cfg.Store.TTLSeconds != 600 {
		t.Errorf("unexpected TTLSeconds: %d", cfg.Store.TTLSeconds)
	}
	if cfg.Store.DBPath != "/tmp/test.db" {
		t.Errorf("unexpected DBPath: %s", cfg.Store.DBPath)
	}
	if cfg.UI.Theme != "plain" {
		t.Errorf("unexpected Theme: %s", cfg.UI.Theme)
	}
	if cfg.RecentTasks.MaxItems != 10 {
		t.Errorf("unexpected MaxItems: %d", cfg.RecentTasks.MaxItems)
	}
}

func TestLoad_PartialTOML(t *testing.T) {
	content := `
[hubstaff]
cli_path = "/custom/path"
`
	tmpFile := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Hubstaff.CLIPath != "/custom/path" {
		t.Errorf("expected overridden CLIPath, got: %s", cfg.Hubstaff.CLIPath)
	}
	// Other sections should retain defaults
	if cfg.Store.TTLSeconds != 300 {
		t.Errorf("expected default TTLSeconds, got: %d", cfg.Store.TTLSeconds)
	}
	if cfg.Store.DBPath != "~/.local/share/hubstaff-tui/hubstaff.db" {
		t.Errorf("expected default DBPath, got: %s", cfg.Store.DBPath)
	}
	if cfg.UI.Theme != "catppuccin-mocha" {
		t.Errorf("expected default Theme, got: %s", cfg.UI.Theme)
	}
	if cfg.RecentTasks.MaxItems != 5 {
		t.Errorf("expected default MaxItems, got: %d", cfg.RecentTasks.MaxItems)
	}
}

func TestLoad_InvalidTOML(t *testing.T) {
	content := `this is not valid toml {{{{`
	tmpFile := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(tmpFile)
	if err == nil {
		t.Fatal("expected error for invalid TOML, got nil")
	}
}

func TestExpandPath_Tilde(t *testing.T) {
	expanded, err := ExpandPath("~/some/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "some/path")
	if expanded != expected {
		t.Errorf("expected %s, got %s", expected, expanded)
	}
}

func TestExpandPath_Absolute(t *testing.T) {
	path := "/absolute/path/to/file"
	expanded, err := ExpandPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expanded != path {
		t.Errorf("expected %s, got %s", path, expanded)
	}
}

func TestResolvedDBPath(t *testing.T) {
	cfg := DefaultConfig()
	resolved, err := cfg.ResolvedDBPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".local/share/hubstaff-tui/hubstaff.db")
	if resolved != expected {
		t.Errorf("expected %s, got %s", expected, resolved)
	}
}
