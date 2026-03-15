package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", name, err)
	}
	return data
}

func TestParseStatus_Tracking(t *testing.T) {
	data := readFixture(t, "status_tracking.json")
	s, err := parseStatus(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.Tracking {
		t.Error("expected Tracking to be true")
	}
	if s.ActiveProject.ID != "proj-1" {
		t.Errorf("expected project ID proj-1, got %s", s.ActiveProject.ID)
	}
	if s.ActiveProject.Name != "Acme Backend" {
		t.Errorf("expected project name Acme Backend, got %s", s.ActiveProject.Name)
	}
	if s.ActiveProject.TrackedToday != "2:15:30" {
		t.Errorf("expected tracked today 2:15:30, got %s", s.ActiveProject.TrackedToday)
	}
	if s.ActiveTask.ID != "task-42" {
		t.Errorf("expected task ID task-42, got %s", s.ActiveTask.ID)
	}
	if s.ActiveTask.Name != "Fix login redirect" {
		t.Errorf("expected task name Fix login redirect, got %s", s.ActiveTask.Name)
	}
}

func TestParseStatus_NotTracking(t *testing.T) {
	data := readFixture(t, "status_not_tracking.json")
	s, err := parseStatus(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Tracking {
		t.Error("expected Tracking to be false")
	}
	if s.ActiveProject.ID != "" {
		t.Errorf("expected empty project ID, got %s", s.ActiveProject.ID)
	}
	if s.ActiveProject.Name != "" {
		t.Errorf("expected empty project name, got %s", s.ActiveProject.Name)
	}
	if s.ActiveProject.TrackedToday != "0:00:00" {
		t.Errorf("expected tracked today 0:00:00, got %s", s.ActiveProject.TrackedToday)
	}
	if s.ActiveTask.ID != "" {
		t.Errorf("expected empty task ID, got %s", s.ActiveTask.ID)
	}
	if s.ActiveTask.Name != "" {
		t.Errorf("expected empty task name, got %s", s.ActiveTask.Name)
	}
}

func TestParseProjects(t *testing.T) {
	data := readFixture(t, "projects.json")
	projects, err := parseProjects(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 3 {
		t.Fatalf("expected 3 projects, got %d", len(projects))
	}
	if projects[0].ID != "proj-1" || projects[0].Name != "Acme Backend" {
		t.Errorf("unexpected first project: %+v", projects[0])
	}
	if projects[1].ID != "proj-2" || projects[1].Name != "Acme Frontend" {
		t.Errorf("unexpected second project: %+v", projects[1])
	}
	if projects[2].ID != "proj-3" || projects[2].Name != "Internal Tools" {
		t.Errorf("unexpected third project: %+v", projects[2])
	}
}

func TestParseProjects_Empty(t *testing.T) {
	data := readFixture(t, "empty_projects.json")
	projects, err := parseProjects(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if projects == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestParseTasks(t *testing.T) {
	data := readFixture(t, "tasks.json")
	tasks, err := parseTasks(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}
	expected := []string{"Fix login redirect", "Update dependencies", "Refactor auth middleware"}
	for i, want := range expected {
		if tasks[i].Summary != want {
			t.Errorf("task[%d] summary: expected %q, got %q", i, want, tasks[i].Summary)
		}
	}
}

func TestParseTasks_Empty(t *testing.T) {
	data := readFixture(t, "empty_tasks.json")
	tasks, err := parseTasks(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tasks == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestParseStatus_MalformedJSON(t *testing.T) {
	_, err := parseStatus([]byte(`{not valid json`))
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestCLIError_Format(t *testing.T) {
	e := &CLIError{
		Command:  "status",
		ExitCode: 1,
		Stderr:   "not running",
	}
	expected := "hubstaff cli status failed (exit 1): not running"
	if e.Error() != expected {
		t.Errorf("expected %q, got %q", expected, e.Error())
	}
}

func TestCheckCLI_NotFound(t *testing.T) {
	c := NewClient("/nonexistent/path/HubstaffCLI")
	err := c.CheckCLI()
	if err == nil {
		t.Fatal("expected error for nonexistent CLI path")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected 'not found' in error, got: %v", err)
	}
}

func TestCheckCLI_IsDirectory(t *testing.T) {
	dir := t.TempDir()
	c := NewClient(dir)
	err := c.CheckCLI()
	if err == nil {
		t.Fatal("expected error for directory path")
	}
	if !strings.Contains(err.Error(), "directory") {
		t.Fatalf("expected 'directory' in error, got: %v", err)
	}
}

func TestCheckCLI_NotExecutable(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "hubstaff-cli-*")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	// Ensure file is not executable
	if err := os.Chmod(f.Name(), 0644); err != nil {
		t.Fatal(err)
	}
	c := NewClient(f.Name())
	err = c.CheckCLI()
	if err == nil {
		t.Fatal("expected error for non-executable file")
	}
	if !strings.Contains(err.Error(), "not executable") {
		t.Fatalf("expected 'not executable' in error, got: %v", err)
	}
}

func TestCheckCLI_Valid(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "hubstaff-cli-*")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	if err := os.Chmod(f.Name(), 0755); err != nil {
		t.Fatal(err)
	}
	c := NewClient(f.Name())
	if err := c.CheckCLI(); err != nil {
		t.Fatalf("expected no error for valid executable, got: %v", err)
	}
}
