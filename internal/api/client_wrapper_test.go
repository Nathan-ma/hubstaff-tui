package api

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func makeFakeCLI(t *testing.T, script string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "fake-cli")
	content := "#!/bin/sh\n" + script
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}
	return path
}

// --- GetStatus ---

func TestClient_GetStatus_Success(t *testing.T) {
	path := makeFakeCLI(t, `
case "$1" in
  status)
    printf '{"tracking":true,"active_project":{"id":"1","name":"Acme Backend","tracked_today":"2:15:30"},"active_task":{"id":"42","name":"Fix login redirect"}}'
    ;;
  *)
    exit 1
    ;;
esac
`)
	c := NewClient(path)
	s, err := c.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.Tracking {
		t.Error("expected Tracking to be true")
	}
	if s.ActiveProject.Name != "Acme Backend" {
		t.Errorf("expected project name 'Acme Backend', got %q", s.ActiveProject.Name)
	}
	if string(s.ActiveProject.ID) != "1" {
		t.Errorf("expected project ID '1', got %q", s.ActiveProject.ID)
	}
	if s.ActiveTask.Name != "Fix login redirect" {
		t.Errorf("expected task name 'Fix login redirect', got %q", s.ActiveTask.Name)
	}
}

func TestClient_GetStatus_NotTracking(t *testing.T) {
	path := makeFakeCLI(t, `
case "$1" in
  status)
    printf '{"tracking":false,"active_project":{"tracked_today":"0:00:00"}}'
    ;;
  *)
    exit 1
    ;;
esac
`)
	c := NewClient(path)
	s, err := c.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Tracking {
		t.Error("expected Tracking to be false")
	}
	if s.ActiveProject.TrackedToday != "0:00:00" {
		t.Errorf("expected tracked_today '0:00:00', got %q", s.ActiveProject.TrackedToday)
	}
}

func TestClient_GetStatus_CLIError(t *testing.T) {
	path := makeFakeCLI(t, `exit 1`)
	c := NewClient(path)
	_, err := c.GetStatus(context.Background())
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected *CLIError, got %T: %v", err, err)
	}
	if cliErr.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", cliErr.ExitCode)
	}
}

// --- ListProjects ---

func TestClient_ListProjects_Success(t *testing.T) {
	path := makeFakeCLI(t, `
case "$1" in
  projects)
    printf '{"projects":[{"id":"1","name":"Acme Backend"},{"id":"2","name":"Acme Frontend"},{"id":"3","name":"Internal Tools"}]}'
    ;;
  *)
    exit 1
    ;;
esac
`)
	c := NewClient(path)
	projects, err := c.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 3 {
		t.Fatalf("expected 3 projects, got %d", len(projects))
	}
	if projects[0].Name != "Acme Backend" {
		t.Errorf("expected first project 'Acme Backend', got %q", projects[0].Name)
	}
	if projects[2].Name != "Internal Tools" {
		t.Errorf("expected third project 'Internal Tools', got %q", projects[2].Name)
	}
}

func TestClient_ListProjects_Error(t *testing.T) {
	path := makeFakeCLI(t, `exit 1`)
	c := NewClient(path)
	_, err := c.ListProjects(context.Background())
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected *CLIError, got %T: %v", err, err)
	}
}

// --- ListTasks ---

func TestClient_ListTasks_Success(t *testing.T) {
	// Capture the argument passed as projectID via an env var written to a temp file.
	argsFile := filepath.Join(t.TempDir(), "args.txt")
	path := makeFakeCLI(t, `
case "$1" in
  tasks)
    echo "$2" > `+argsFile+`
    printf '{"tasks":[{"id":"1","summary":"Fix login redirect"},{"id":"2","summary":"Update dependencies"},{"id":"3","summary":"Refactor auth middleware"}]}'
    ;;
  *)
    exit 1
    ;;
esac
`)
	c := NewClient(path)
	tasks, err := c.ListTasks(context.Background(), "proj-99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}
	if tasks[0].Summary != "Fix login redirect" {
		t.Errorf("expected first task 'Fix login redirect', got %q", tasks[0].Summary)
	}

	// Verify the project ID was forwarded as the second CLI argument.
	argsData, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("could not read args file: %v", err)
	}
	// echo appends a newline; trim it.
	got := string(argsData)
	if len(got) > 0 && got[len(got)-1] == '\n' {
		got = got[:len(got)-1]
	}
	if got != "proj-99" {
		t.Errorf("expected projectID 'proj-99' passed to CLI, got %q", got)
	}
}

func TestClient_ListTasks_Error(t *testing.T) {
	path := makeFakeCLI(t, `exit 1`)
	c := NewClient(path)
	_, err := c.ListTasks(context.Background(), "1")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected *CLIError, got %T: %v", err, err)
	}
}

// --- StartTask ---

func TestClient_StartTask_Success(t *testing.T) {
	path := makeFakeCLI(t, `
case "$1" in
  start_task)
    exit 0
    ;;
  *)
    exit 1
    ;;
esac
`)
	c := NewClient(path)
	if err := c.StartTask(context.Background(), "42"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_StartTask_Error(t *testing.T) {
	path := makeFakeCLI(t, `exit 1`)
	c := NewClient(path)
	err := c.StartTask(context.Background(), "42")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected *CLIError, got %T: %v", err, err)
	}
	if cliErr.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", cliErr.ExitCode)
	}
}

// --- Stop ---

func TestClient_Stop_Success(t *testing.T) {
	path := makeFakeCLI(t, `
case "$1" in
  stop)
    exit 0
    ;;
  *)
    exit 1
    ;;
esac
`)
	c := NewClient(path)
	if err := c.Stop(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Stop_Error(t *testing.T) {
	path := makeFakeCLI(t, `exit 1`)
	c := NewClient(path)
	err := c.Stop(context.Background())
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected *CLIError, got %T: %v", err, err)
	}
}
