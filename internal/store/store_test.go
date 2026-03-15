package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	s, err := Open(dbPath, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func newTestStoreWithTTL(t *testing.T, ttl time.Duration) *Store {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	s, err := Open(dbPath, ttl)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestOpen_CreatesDatabase(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	s, err := Open(dbPath, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("database file was not created")
	}
}

func TestOpen_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "a", "b", "c")
	dbPath := filepath.Join(nested, "test.db")

	s, err := Open(dbPath, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("database file was not created in nested directory")
	}
}

func TestUpsertProjects(t *testing.T) {
	s := newTestStore(t)

	projects := []ProjectRow{
		{ID: "p1", Name: "Project Alpha"},
		{ID: "p2", Name: "Project Beta"},
	}
	if err := s.UpsertProjects(projects); err != nil {
		t.Fatal(err)
	}

	got, stale, err := s.ListProjects()
	if err != nil {
		t.Fatal(err)
	}
	if stale {
		t.Error("expected fresh, got stale")
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(got))
	}
	// Ordered by name
	if got[0].Name != "Project Alpha" {
		t.Errorf("expected Project Alpha first, got %s", got[0].Name)
	}
	if got[1].Name != "Project Beta" {
		t.Errorf("expected Project Beta second, got %s", got[1].Name)
	}
}

func TestUpsertProjects_Update(t *testing.T) {
	s := newTestStore(t)

	if err := s.UpsertProjects([]ProjectRow{{ID: "p1", Name: "Old Name"}}); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertProjects([]ProjectRow{{ID: "p1", Name: "New Name"}}); err != nil {
		t.Fatal(err)
	}

	got, _, err := s.ListProjects()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 project, got %d", len(got))
	}
	if got[0].Name != "New Name" {
		t.Errorf("expected New Name, got %s", got[0].Name)
	}
}

func TestListProjects_Stale(t *testing.T) {
	s := newTestStoreWithTTL(t, 1*time.Millisecond)

	if err := s.UpsertProjects([]ProjectRow{{ID: "p1", Name: "P1"}}); err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Millisecond)

	_, stale, err := s.ListProjects()
	if err != nil {
		t.Fatal(err)
	}
	if !stale {
		t.Error("expected stale, got fresh")
	}
}

func TestListProjects_Fresh(t *testing.T) {
	s := newTestStore(t) // 5 min TTL

	if err := s.UpsertProjects([]ProjectRow{{ID: "p1", Name: "P1"}}); err != nil {
		t.Fatal(err)
	}

	_, stale, err := s.ListProjects()
	if err != nil {
		t.Fatal(err)
	}
	if stale {
		t.Error("expected fresh, got stale")
	}
}

func TestUpsertTasks(t *testing.T) {
	s := newTestStore(t)

	tasks := []TaskRow{
		{ID: "t1", Summary: "Task One"},
		{ID: "t2", Summary: "Task Two"},
	}
	if err := s.UpsertTasks("p1", tasks); err != nil {
		t.Fatal(err)
	}

	got, stale, err := s.ListTasks("p1")
	if err != nil {
		t.Fatal(err)
	}
	if stale {
		t.Error("expected fresh, got stale")
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(got))
	}
}

func TestListTasks_FiltersByProject(t *testing.T) {
	s := newTestStore(t)

	if err := s.UpsertTasks("p1", []TaskRow{{ID: "t1", Summary: "Task for P1"}}); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertTasks("p2", []TaskRow{{ID: "t2", Summary: "Task for P2"}}); err != nil {
		t.Fatal(err)
	}

	got, _, err := s.ListTasks("p1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 task for p1, got %d", len(got))
	}
	if got[0].ID != "t1" {
		t.Errorf("expected task t1, got %s", got[0].ID)
	}

	got2, _, err := s.ListTasks("p2")
	if err != nil {
		t.Fatal(err)
	}
	if len(got2) != 1 {
		t.Fatalf("expected 1 task for p2, got %d", len(got2))
	}
	if got2[0].ID != "t2" {
		t.Errorf("expected task t2, got %s", got2[0].ID)
	}
}

func TestInvalidateProjects(t *testing.T) {
	s := newTestStore(t)

	if err := s.UpsertProjects([]ProjectRow{{ID: "p1", Name: "P1"}}); err != nil {
		t.Fatal(err)
	}

	if err := s.InvalidateProjects(); err != nil {
		t.Fatal(err)
	}

	_, stale, err := s.ListProjects()
	if err != nil {
		t.Fatal(err)
	}
	if !stale {
		t.Error("expected stale after invalidation")
	}
}

func TestStartSession(t *testing.T) {
	s := newTestStore(t)

	if err := s.StartSession("t1", "p1"); err != nil {
		t.Fatal(err)
	}

	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("expected 1 session, got %d", count)
	}

	var taskID, projectID string
	var stoppedAt *string
	err := s.db.QueryRow("SELECT task_id, project_id, stopped_at FROM sessions WHERE id = 1").
		Scan(&taskID, &projectID, &stoppedAt)
	if err != nil {
		t.Fatal(err)
	}
	if taskID != "t1" || projectID != "p1" {
		t.Errorf("unexpected session data: task=%s project=%s", taskID, projectID)
	}
	if stoppedAt != nil {
		t.Error("expected stopped_at to be NULL")
	}
}

func TestStopSession(t *testing.T) {
	s := newTestStore(t)

	if err := s.StartSession("t1", "p1"); err != nil {
		t.Fatal(err)
	}

	// Small delay so duration > 0
	time.Sleep(10 * time.Millisecond)

	if err := s.StopSession(); err != nil {
		t.Fatal(err)
	}

	var stoppedAt *string
	var duration int
	err := s.db.QueryRow("SELECT stopped_at, duration_seconds FROM sessions WHERE id = 1").
		Scan(&stoppedAt, &duration)
	if err != nil {
		t.Fatal(err)
	}
	if stoppedAt == nil {
		t.Error("expected stopped_at to be set")
	}
	// Duration may be 0 for very short intervals, just check it's non-negative
	if duration < 0 {
		t.Errorf("expected non-negative duration, got %d", duration)
	}
}

func TestUpdateHeartbeat(t *testing.T) {
	s := newTestStore(t)

	if err := s.StartSession("t1", "p1"); err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Millisecond)

	if err := s.UpdateHeartbeat(); err != nil {
		t.Fatal(err)
	}

	var duration int
	err := s.db.QueryRow("SELECT duration_seconds FROM sessions WHERE id = 1").Scan(&duration)
	if err != nil {
		t.Fatal(err)
	}
	// Duration is computed; for very short intervals it may be 0
	if duration < 0 {
		t.Errorf("expected non-negative duration, got %d", duration)
	}

	// Verify session is still active (stopped_at IS NULL)
	var stoppedAt *string
	err = s.db.QueryRow("SELECT stopped_at FROM sessions WHERE id = 1").Scan(&stoppedAt)
	if err != nil {
		t.Fatal(err)
	}
	if stoppedAt != nil {
		t.Error("expected stopped_at to remain NULL after heartbeat")
	}
}

func TestTodaySummary(t *testing.T) {
	s := newTestStore(t)

	// Insert project and task for join
	if err := s.UpsertProjects([]ProjectRow{{ID: "p1", Name: "Project One"}}); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertTasks("p1", []TaskRow{{ID: "t1", Summary: "Task One"}}); err != nil {
		t.Fatal(err)
	}

	// Insert a session with known duration
	now := time.Now().UTC()
	_, err := s.db.Exec(
		"INSERT INTO sessions (task_id, project_id, started_at, stopped_at, duration_seconds) VALUES (?, ?, ?, ?, ?)",
		"t1", "p1", fmtTime(now.Add(-time.Hour)), fmtTime(now), 3600,
	)
	if err != nil {
		t.Fatal(err)
	}

	summary, err := s.TodaySummary()
	if err != nil {
		t.Fatal(err)
	}
	if len(summary) != 1 {
		t.Fatalf("expected 1 summary row, got %d", len(summary))
	}
	if summary[0].ProjectName != "Project One" {
		t.Errorf("expected Project One, got %s", summary[0].ProjectName)
	}
	if summary[0].TaskSummary != "Task One" {
		t.Errorf("expected Task One, got %s", summary[0].TaskSummary)
	}
	if summary[0].DurationSeconds != 3600 {
		t.Errorf("expected 3600s, got %d", summary[0].DurationSeconds)
	}
}

func TestTouchRecent(t *testing.T) {
	s := newTestStore(t)

	if err := s.TouchRecent("t1", "p1"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Millisecond)
	if err := s.TouchRecent("t2", "p1"); err != nil {
		t.Fatal(err)
	}

	recents, err := s.ListRecents(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(recents) != 2 {
		t.Fatalf("expected 2 recents, got %d", len(recents))
	}
	// Most recent first
	if recents[0].TaskID != "t2" {
		t.Errorf("expected t2 first, got %s", recents[0].TaskID)
	}
	if recents[1].TaskID != "t1" {
		t.Errorf("expected t1 second, got %s", recents[1].TaskID)
	}
}

func TestListRecents_Limit(t *testing.T) {
	s := newTestStore(t)

	for i := 0; i < 10; i++ {
		taskID := "t" + string(rune('0'+i))
		if err := s.TouchRecent(taskID, "p1"); err != nil {
			t.Fatal(err)
		}
		time.Sleep(1 * time.Millisecond)
	}

	recents, err := s.ListRecents(5)
	if err != nil {
		t.Fatal(err)
	}
	if len(recents) != 5 {
		t.Fatalf("expected 5 recents, got %d", len(recents))
	}
}

func TestInvalidateAll(t *testing.T) {
	s := newTestStore(t)

	if err := s.UpsertProjects([]ProjectRow{{ID: "p1", Name: "P1"}}); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertTasks("p1", []TaskRow{{ID: "t1", Summary: "T1"}}); err != nil {
		t.Fatal(err)
	}

	// Verify fresh before invalidation
	_, stale, err := s.ListProjects()
	if err != nil {
		t.Fatal(err)
	}
	if stale {
		t.Error("expected fresh before invalidation")
	}

	if err := s.InvalidateAll(); err != nil {
		t.Fatal(err)
	}

	_, stale, err = s.ListProjects()
	if err != nil {
		t.Fatal(err)
	}
	if !stale {
		t.Error("expected projects stale after InvalidateAll")
	}

	_, stale, err = s.ListTasks("p1")
	if err != nil {
		t.Fatal(err)
	}
	if !stale {
		t.Error("expected tasks stale after InvalidateAll")
	}
}

func TestInvalidateTasks(t *testing.T) {
	s := newTestStore(t)

	if err := s.UpsertTasks("p1", []TaskRow{{ID: "t1", Summary: "T1"}}); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertTasks("p2", []TaskRow{{ID: "t2", Summary: "T2"}}); err != nil {
		t.Fatal(err)
	}

	if err := s.InvalidateTasks("p1"); err != nil {
		t.Fatal(err)
	}

	_, stale, err := s.ListTasks("p1")
	if err != nil {
		t.Fatal(err)
	}
	if !stale {
		t.Error("expected p1 tasks stale after invalidation")
	}

	_, stale, err = s.ListTasks("p2")
	if err != nil {
		t.Fatal(err)
	}
	if stale {
		t.Error("expected p2 tasks still fresh")
	}
}
