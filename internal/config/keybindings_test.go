package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultKeybindings(t *testing.T) {
	cfg := DefaultConfig()

	kb := cfg.Keybindings
	if kb.Quit != "esc" {
		t.Errorf("expected Quit=esc, got %s", kb.Quit)
	}
	if kb.Stop != "ctrl+e" {
		t.Errorf("expected Stop=ctrl+e, got %s", kb.Stop)
	}
	if kb.Refresh != "ctrl+r" {
		t.Errorf("expected Refresh=ctrl+r, got %s", kb.Refresh)
	}
	if kb.Filter != "/" {
		t.Errorf("expected Filter=/, got %s", kb.Filter)
	}
	if kb.Help != "?" {
		t.Errorf("expected Help=?, got %s", kb.Help)
	}
	if kb.Summary != "T" {
		t.Errorf("expected Summary=T, got %s", kb.Summary)
	}
	if kb.SwitchPane != "tab" {
		t.Errorf("expected SwitchPane=tab, got %s", kb.SwitchPane)
	}
	if kb.GlobalSearch != "ctrl+f" {
		t.Errorf("expected GlobalSearch=ctrl+f, got %s", kb.GlobalSearch)
	}
}

func TestCustomKeybindings(t *testing.T) {
	content := `
[keybindings]
quit = "q"
stop = "ctrl+s"
refresh = "F5"
filter = "f"
help = "h"
summary = "s"
switch_pane = "ctrl+tab"
global_search = "ctrl+g"
`
	tmpFile := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	kb := cfg.Keybindings
	if kb.Quit != "q" {
		t.Errorf("expected Quit=q, got %s", kb.Quit)
	}
	if kb.Stop != "ctrl+s" {
		t.Errorf("expected Stop=ctrl+s, got %s", kb.Stop)
	}
	if kb.Refresh != "F5" {
		t.Errorf("expected Refresh=F5, got %s", kb.Refresh)
	}
	if kb.Filter != "f" {
		t.Errorf("expected Filter=f, got %s", kb.Filter)
	}
	if kb.Help != "h" {
		t.Errorf("expected Help=h, got %s", kb.Help)
	}
	if kb.Summary != "s" {
		t.Errorf("expected Summary=s, got %s", kb.Summary)
	}
	if kb.SwitchPane != "ctrl+tab" {
		t.Errorf("expected SwitchPane=ctrl+tab, got %s", kb.SwitchPane)
	}
	if kb.GlobalSearch != "ctrl+g" {
		t.Errorf("expected GlobalSearch=ctrl+g, got %s", kb.GlobalSearch)
	}
}

func TestPartialKeybindingsOverride(t *testing.T) {
	content := `
[keybindings]
quit = "q"
summary = "s"
`
	tmpFile := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	kb := cfg.Keybindings

	// Overridden values
	if kb.Quit != "q" {
		t.Errorf("expected Quit=q, got %s", kb.Quit)
	}
	if kb.Summary != "s" {
		t.Errorf("expected Summary=s, got %s", kb.Summary)
	}

	// Non-overridden values should retain defaults
	if kb.Stop != "ctrl+e" {
		t.Errorf("expected Stop=ctrl+e, got %s", kb.Stop)
	}
	if kb.Refresh != "ctrl+r" {
		t.Errorf("expected Refresh=ctrl+r, got %s", kb.Refresh)
	}
	if kb.Filter != "/" {
		t.Errorf("expected Filter=/, got %s", kb.Filter)
	}
	if kb.Help != "?" {
		t.Errorf("expected Help=?, got %s", kb.Help)
	}
	if kb.SwitchPane != "tab" {
		t.Errorf("expected SwitchPane=tab, got %s", kb.SwitchPane)
	}
	if kb.GlobalSearch != "ctrl+f" {
		t.Errorf("expected GlobalSearch=ctrl+f, got %s", kb.GlobalSearch)
	}
}

func TestNoKeybindingsSection(t *testing.T) {
	content := `
[hubstaff]
cli_path = "/usr/local/bin/hubstaff"
`
	tmpFile := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All keybindings should be defaults when section is absent
	defaults := DefaultConfig().Keybindings
	if cfg.Keybindings != defaults {
		t.Errorf("expected default keybindings when section absent, got %+v", cfg.Keybindings)
	}
}
