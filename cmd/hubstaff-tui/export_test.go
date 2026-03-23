package main

import (
	"encoding/csv"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Nathan-ma/hubstaff-tui/internal/config"
	"github.com/Nathan-ma/hubstaff-tui/internal/store"
)

// newExportCfg builds a minimal config pointing at the given dbPath.
func newExportCfg(dbPath string) config.Config {
	cfg := config.DefaultConfig()
	cfg.Store.DBPath = dbPath
	return cfg
}

// openExportTestStore opens a fresh store at dbPath and registers cleanup.
func openExportTestStore(t *testing.T, dbPath string) *store.Store {
	t.Helper()
	st, err := store.Open(dbPath, 5*time.Minute)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

// --- Tests ---

func TestRunExport_CSV_Header_Always_Present(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "empty.db")
	// Open and immediately close to create the schema.
	st, err := store.Open(dbPath, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	_ = st.Close()

	cfg := newExportCfg(dbPath)
	output := captureStdout(t, func() {
		code := runExport(&cfg, []string{"--today"})
		if code != 0 {
			t.Errorf("expected exit 0, got %d", code)
		}
	})

	r := csv.NewReader(strings.NewReader(output))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	if len(records) < 1 {
		t.Fatal("expected at least a header row")
	}
	want := []string{"date", "project", "task", "duration_seconds", "started_at", "stopped_at"}
	for i, col := range want {
		if records[0][i] != col {
			t.Errorf("header[%d]: want %q, got %q", i, col, records[0][i])
		}
	}
	// No data rows for empty store.
	if len(records) != 1 {
		t.Errorf("expected exactly 1 row (header only), got %d", len(records))
	}
}

func TestRunExport_CSV_WithSessions(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	st := openExportTestStore(t, dbPath)

	if err := st.UpsertProjects([]store.ProjectRow{{ID: "p1", Name: "Alpha Project"}}); err != nil {
		t.Fatal(err)
	}
	if err := st.UpsertTasks("p1", []store.TaskRow{{ID: "t1", Summary: "Fix bug"}}); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, time.UTC)
	todayStop := todayStart.Add(90 * time.Minute)
	if err := st.InsertSessionForTest("t1", "p1", todayStart, &todayStop, 5400); err != nil {
		t.Fatal(err)
	}
	_ = st.Close()

	cfg := newExportCfg(dbPath)
	output := captureStdout(t, func() {
		code := runExport(&cfg, []string{"--today", "--format", "csv"})
		if code != 0 {
			t.Errorf("expected exit 0, got %d", code)
		}
	})

	r := csv.NewReader(strings.NewReader(output))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	// header + 1 data row
	if len(records) != 2 {
		t.Fatalf("expected 2 rows (header + 1 session), got %d", len(records))
	}
	row := records[1]
	if row[1] != "Alpha Project" {
		t.Errorf("project: want %q, got %q", "Alpha Project", row[1])
	}
	if row[2] != "Fix bug" {
		t.Errorf("task: want %q, got %q", "Fix bug", row[2])
	}
	if row[3] != "5400" {
		t.Errorf("duration_seconds: want %q, got %q", "5400", row[3])
	}
}

func TestRunExport_JSON_EmptyStore(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "empty.db")
	st, _ := store.Open(dbPath, time.Minute)
	_ = st.Close()

	cfg := newExportCfg(dbPath)
	output := captureStdout(t, func() {
		code := runExport(&cfg, []string{"--format", "json"})
		if code != 0 {
			t.Errorf("expected exit 0, got %d", code)
		}
	})

	var results []exportRecord
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("parse json: %v\noutput: %s", err, output)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 records, got %d", len(results))
	}
}

func TestRunExport_JSON_WithSessions(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	st := openExportTestStore(t, dbPath)

	if err := st.UpsertProjects([]store.ProjectRow{{ID: "p1", Name: "Beta Co"}}); err != nil {
		t.Fatal(err)
	}
	if err := st.UpsertTasks("p1", []store.TaskRow{{ID: "t2", Summary: "Write tests"}}); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, time.UTC)
	todayStop := todayStart.Add(30 * time.Minute)
	if err := st.InsertSessionForTest("t2", "p1", todayStart, &todayStop, 1800); err != nil {
		t.Fatal(err)
	}
	_ = st.Close()

	cfg := newExportCfg(dbPath)
	output := captureStdout(t, func() {
		code := runExport(&cfg, []string{"--format", "json", "--today"})
		if code != 0 {
			t.Errorf("expected exit 0, got %d", code)
		}
	})

	var results []exportRecord
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("parse json: %v\noutput: %s", err, output)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 record, got %d", len(results))
	}
	rec := results[0]
	if rec.Project != "Beta Co" {
		t.Errorf("project: want %q, got %q", "Beta Co", rec.Project)
	}
	if rec.Task != "Write tests" {
		t.Errorf("task: want %q, got %q", "Write tests", rec.Task)
	}
	if rec.DurationSeconds != 1800 {
		t.Errorf("duration_seconds: want 1800, got %d", rec.DurationSeconds)
	}
	if rec.StoppedAt == "" {
		t.Error("stopped_at should not be empty for a completed session")
	}
}

func TestRunExport_Week(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	st := openExportTestStore(t, dbPath)

	now := time.Now().UTC()
	// Session 3 days ago — should appear with --week.
	threeDaysAgo := now.AddDate(0, 0, -3)
	threeDaysAgoStop := threeDaysAgo.Add(time.Hour)
	if err := st.InsertSessionForTest("t1", "p1", threeDaysAgo, &threeDaysAgoStop, 3600); err != nil {
		t.Fatal(err)
	}
	// Session 10 days ago — should NOT appear.
	tenDaysAgo := now.AddDate(0, 0, -10)
	tenDaysAgoStop := tenDaysAgo.Add(time.Hour)
	if err := st.InsertSessionForTest("t1", "p1", tenDaysAgo, &tenDaysAgoStop, 3600); err != nil {
		t.Fatal(err)
	}
	_ = st.Close()

	cfg := newExportCfg(dbPath)
	output := captureStdout(t, func() {
		code := runExport(&cfg, []string{"--week", "--format", "csv"})
		if code != 0 {
			t.Errorf("expected exit 0, got %d", code)
		}
	})

	r := csv.NewReader(strings.NewReader(output))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	// header + 1 row (only the 3-days-ago session)
	if len(records) != 2 {
		t.Fatalf("expected 2 rows (header + 1), got %d", len(records))
	}
}

func TestRunExport_Since(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	st := openExportTestStore(t, dbPath)

	now := time.Now().UTC()
	// Session yesterday — should appear when --since is 2 days ago.
	yesterday := now.AddDate(0, 0, -1)
	yesterdayStop := yesterday.Add(time.Hour)
	if err := st.InsertSessionForTest("t1", "p1", yesterday, &yesterdayStop, 3600); err != nil {
		t.Fatal(err)
	}
	// Session 5 days ago — should NOT appear.
	fiveDaysAgo := now.AddDate(0, 0, -5)
	fiveDaysAgoStop := fiveDaysAgo.Add(time.Hour)
	if err := st.InsertSessionForTest("t1", "p1", fiveDaysAgo, &fiveDaysAgoStop, 3600); err != nil {
		t.Fatal(err)
	}
	_ = st.Close()

	sinceDate := now.AddDate(0, 0, -2).Format("2006-01-02")
	cfg := newExportCfg(dbPath)
	output := captureStdout(t, func() {
		code := runExport(&cfg, []string{"--since", sinceDate, "--format", "csv"})
		if code != 0 {
			t.Errorf("expected exit 0, got %d", code)
		}
	})

	r := csv.NewReader(strings.NewReader(output))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 rows (header + 1), got %d", len(records))
	}
}

func TestRunExport_InvalidFormat(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "empty.db")
	st, _ := store.Open(dbPath, time.Minute)
	_ = st.Close()

	cfg := newExportCfg(dbPath)
	code := runExport(&cfg, []string{"--format", "xml"})
	if code == 0 {
		t.Error("expected non-zero exit for invalid format")
	}
}

func TestRunExport_InvalidSinceDate(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "empty.db")
	st, _ := store.Open(dbPath, time.Minute)
	_ = st.Close()

	cfg := newExportCfg(dbPath)
	code := runExport(&cfg, []string{"--since", "not-a-date"})
	if code == 0 {
		t.Error("expected non-zero exit for invalid --since date")
	}
}
