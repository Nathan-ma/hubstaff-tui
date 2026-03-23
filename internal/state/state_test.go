package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FileNotExist(t *testing.T) {
	s := Load("/tmp/hubstaff-tui-test-nonexistent/state.json")
	if s.LastProjectID != "" {
		t.Errorf("expected empty LastProjectID, got %q", s.LastProjectID)
	}
	if s.LastProjectName != "" {
		t.Errorf("expected empty LastProjectName, got %q", s.LastProjectName)
	}
	if s.LastTaskID != "" {
		t.Errorf("expected empty LastTaskID, got %q", s.LastTaskID)
	}
	if s.ScrollPosition != 0 {
		t.Errorf("expected ScrollPosition 0, got %d", s.ScrollPosition)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	want := AppState{
		LastProjectID:   "proj-123",
		LastProjectName: "My Project",
		LastTaskID:      "task-456",
		ScrollPosition:  3,
	}

	if err := Save(path, want); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got := Load(path)
	if got != want {
		t.Errorf("Load returned %+v, want %+v", got, want)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	if err := os.WriteFile(path, []byte("{invalid json}"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	s := Load(path)
	if s.LastProjectID != "" {
		t.Errorf("expected empty LastProjectID for invalid JSON, got %q", s.LastProjectID)
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deep", "state.json")

	want := AppState{
		LastProjectID:   "proj-789",
		LastProjectName: "Nested Project",
		ScrollPosition:  1,
	}

	if err := Save(path, want); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify the file was created.
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("state file not created: %v", err)
	}

	got := Load(path)
	if got != want {
		t.Errorf("Load returned %+v, want %+v", got, want)
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	s := Load(path)
	// Empty file is invalid JSON, should return zero state.
	if s.LastProjectID != "" {
		t.Errorf("expected empty LastProjectID for empty file, got %q", s.LastProjectID)
	}
}

func TestSave_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	first := AppState{LastProjectID: "proj-1", LastProjectName: "First"}
	second := AppState{LastProjectID: "proj-2", LastProjectName: "Second", ScrollPosition: 5}

	if err := Save(path, first); err != nil {
		t.Fatalf("Save first failed: %v", err)
	}
	if err := Save(path, second); err != nil {
		t.Fatalf("Save second failed: %v", err)
	}

	got := Load(path)
	if got != second {
		t.Errorf("Load returned %+v, want %+v", got, second)
	}
}

func TestLoad_ZeroState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	// Save a zero-valued state and load it back.
	zero := AppState{}
	if err := Save(path, zero); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got := Load(path)
	if got != zero {
		t.Errorf("Load returned %+v, want %+v", got, zero)
	}
}

func TestSaveAndLoad_EmptyTaskID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	want := AppState{
		LastProjectID:   "proj-100",
		LastProjectName: "No Task",
		ScrollPosition:  0,
	}

	if err := Save(path, want); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got := Load(path)
	if got != want {
		t.Errorf("Load returned %+v, want %+v", got, want)
	}
}
