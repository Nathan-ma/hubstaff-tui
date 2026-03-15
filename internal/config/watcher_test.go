package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcher_DetectsChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("[ui]\ntheme = \"plain\""), 0644); err != nil {
		t.Fatal(err)
	}

	w := NewWatcher(path)
	if w.Changed() {
		t.Fatal("should not detect change on first check")
	}

	// Ensure filesystem timestamp granularity is exceeded.
	time.Sleep(50 * time.Millisecond)
	if err := os.WriteFile(path, []byte("[ui]\ntheme = \"catppuccin-mocha\""), 0644); err != nil {
		t.Fatal(err)
	}

	if !w.Changed() {
		t.Fatal("should detect change after file modification")
	}

	if w.Changed() {
		t.Fatal("should not detect change on second check without modification")
	}
}

func TestWatcher_MissingFile(t *testing.T) {
	w := NewWatcher("/nonexistent/path/config.toml")

	// Should not panic and should not report a change.
	if w.Changed() {
		t.Fatal("should not detect change for nonexistent file")
	}
}

func TestWatcher_FileCreatedAfterInit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	// Create watcher before the file exists.
	w := NewWatcher(path)
	if w.Changed() {
		t.Fatal("should not detect change when file does not exist")
	}

	// Create the file — watcher should detect it as a change.
	if err := os.WriteFile(path, []byte("[ui]\ntheme = \"plain\""), 0644); err != nil {
		t.Fatal(err)
	}

	if !w.Changed() {
		t.Fatal("should detect change when file is created")
	}
}
