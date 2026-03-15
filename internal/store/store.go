package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Row types

// ProjectRow represents a cached project.
type ProjectRow struct {
	ID        string
	Name      string
	UpdatedAt time.Time
}

// TaskRow represents a cached task.
type TaskRow struct {
	ID        string
	Summary   string
	ProjectID string
	UpdatedAt time.Time
}

// SessionRow represents a tracking session.
type SessionRow struct {
	ID              int64
	TaskID          string
	ProjectID       string
	StartedAt       time.Time
	StoppedAt       *time.Time
	DurationSeconds int
}

// SummaryRow represents a today-summary entry.
type SummaryRow struct {
	ProjectID       string
	ProjectName     string
	TaskID          string
	TaskSummary     string
	DurationSeconds int
}

// RecentRow represents a recently-used task.
type RecentRow struct {
	TaskID    string
	ProjectID string
	UsedAt    time.Time
}

const schema = `
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    summary TEXT NOT NULL,
    project_id TEXT NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    started_at DATETIME NOT NULL,
    stopped_at DATETIME,
    duration_seconds INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS recents (
    task_id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    used_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sessions_started_at ON sessions(started_at);
CREATE INDEX IF NOT EXISTS idx_tasks_project_id ON tasks(project_id);
`

// sqliteTimeFormat is the time format understood by SQLite datetime functions.
// Uses fractional seconds for sub-second precision in ordering and TTL checks.
const sqliteTimeFormat = "2006-01-02 15:04:05.000"

// fmtTime formats a time.Time for SQLite storage.
func fmtTime(t time.Time) string {
	return t.UTC().Format(sqliteTimeFormat)
}

// parseTime parses a SQLite time string back to time.Time.
// It handles multiple formats since the SQLite driver may return times in
// different formats depending on the column type affinity.
func parseTime(s string) time.Time {
	for _, fmt := range []string{
		sqliteTimeFormat,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		time.RFC3339Nano,
		time.RFC3339,
	} {
		if t, err := time.Parse(fmt, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}

// Store wraps a SQLite database for local caching and session tracking.
type Store struct {
	db  *sql.DB
	ttl time.Duration
}

// Open creates or opens the SQLite database at dbPath, creates tables, and
// returns a Store. Parent directories are created if they do not exist.
func Open(dbPath string, ttl time.Duration) (*Store, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("store: create directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("store: open db: %w", err)
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("store: set WAL mode: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("store: create schema: %w", err)
	}

	return &Store{db: db, ttl: ttl}, nil
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// --- Projects ---

// UpsertProjects bulk-upserts projects and sets updated_at to now.
func (s *Store) UpsertProjects(projects []ProjectRow) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("store: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`
		INSERT INTO projects (id, name, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name=excluded.name, updated_at=excluded.updated_at
	`)
	if err != nil {
		return fmt.Errorf("store: prepare upsert projects: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	now := fmtTime(time.Now())
	for _, p := range projects {
		if _, err := stmt.Exec(p.ID, p.Name, now); err != nil {
			return fmt.Errorf("store: upsert project %s: %w", p.ID, err)
		}
	}

	return tx.Commit()
}

// ListProjects returns all projects. The stale flag is true when any project's
// updated_at is older than the configured TTL, or when there are no projects.
func (s *Store) ListProjects() ([]ProjectRow, bool, error) {
	rows, err := s.db.Query("SELECT id, name, updated_at FROM projects ORDER BY name")
	if err != nil {
		return nil, false, fmt.Errorf("store: list projects: %w", err)
	}
	defer func() { _ = rows.Close() }()

	cutoff := time.Now().UTC().Add(-s.ttl)
	var projects []ProjectRow
	stale := false

	for rows.Next() {
		var p ProjectRow
		var updatedAt string
		if err := rows.Scan(&p.ID, &p.Name, &updatedAt); err != nil {
			return nil, false, fmt.Errorf("store: scan project: %w", err)
		}
		p.UpdatedAt = parseTime(updatedAt)
		if p.UpdatedAt.Before(cutoff) {
			stale = true
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("store: rows err: %w", err)
	}

	if len(projects) == 0 {
		stale = true
	}

	return projects, stale, nil
}

// InvalidateProjects resets updated_at to epoch, forcing a re-fetch.
func (s *Store) InvalidateProjects() error {
	_, err := s.db.Exec("UPDATE projects SET updated_at = ?", fmtTime(time.Time{}))
	return err
}

// --- Tasks ---

// UpsertTasks bulk-upserts tasks for a project and sets updated_at to now.
func (s *Store) UpsertTasks(projectID string, tasks []TaskRow) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("store: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`
		INSERT INTO tasks (id, summary, project_id, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET summary=excluded.summary, project_id=excluded.project_id, updated_at=excluded.updated_at
	`)
	if err != nil {
		return fmt.Errorf("store: prepare upsert tasks: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	now := fmtTime(time.Now())
	for _, t := range tasks {
		if _, err := stmt.Exec(t.ID, t.Summary, projectID, now); err != nil {
			return fmt.Errorf("store: upsert task %s: %w", t.ID, err)
		}
	}

	return tx.Commit()
}

// ListTasks returns tasks for a given project. The stale flag is true when any
// task's updated_at is older than the configured TTL, or when there are no tasks.
func (s *Store) ListTasks(projectID string) ([]TaskRow, bool, error) {
	rows, err := s.db.Query(
		"SELECT id, summary, project_id, updated_at FROM tasks WHERE project_id = ? ORDER BY summary",
		projectID,
	)
	if err != nil {
		return nil, false, fmt.Errorf("store: list tasks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	cutoff := time.Now().UTC().Add(-s.ttl)
	var tasks []TaskRow
	stale := false

	for rows.Next() {
		var t TaskRow
		var updatedAt string
		if err := rows.Scan(&t.ID, &t.Summary, &t.ProjectID, &updatedAt); err != nil {
			return nil, false, fmt.Errorf("store: scan task: %w", err)
		}
		t.UpdatedAt = parseTime(updatedAt)
		if t.UpdatedAt.Before(cutoff) {
			stale = true
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("store: rows err: %w", err)
	}

	if len(tasks) == 0 {
		stale = true
	}

	return tasks, stale, nil
}

// InvalidateTasks resets updated_at for all tasks belonging to a project.
func (s *Store) InvalidateTasks(projectID string) error {
	_, err := s.db.Exec("UPDATE tasks SET updated_at = ? WHERE project_id = ?", fmtTime(time.Time{}), projectID)
	return err
}

// --- Sessions ---

// StartSession records a new tracking session with the current time.
func (s *Store) StartSession(taskID, projectID string) error {
	_, err := s.db.Exec(
		"INSERT INTO sessions (task_id, project_id, started_at) VALUES (?, ?, ?)",
		taskID, projectID, fmtTime(time.Now()),
	)
	return err
}

// StopSession sets stopped_at and computes duration for the active session
// (where stopped_at IS NULL). If no active session exists, it returns nil.
func (s *Store) StopSession() error {
	now := fmtTime(time.Now())
	_, err := s.db.Exec(`
		UPDATE sessions
		SET stopped_at = ?,
		    duration_seconds = CAST((julianday(?) - julianday(started_at)) * 86400 AS INTEGER)
		WHERE stopped_at IS NULL
	`, now, now)
	return err
}

// UpdateHeartbeat updates the duration of the current active session.
func (s *Store) UpdateHeartbeat() error {
	now := fmtTime(time.Now())
	_, err := s.db.Exec(`
		UPDATE sessions
		SET duration_seconds = CAST((julianday(?) - julianday(started_at)) * 86400 AS INTEGER)
		WHERE stopped_at IS NULL
	`, now)
	return err
}

// TodaySummary returns time tracked today grouped by project and task.
func (s *Store) TodaySummary() ([]SummaryRow, error) {
	today := time.Now().UTC().Format("2006-01-02")
	rows, err := s.db.Query(`
		SELECT
			s.project_id,
			COALESCE(p.name, '') AS project_name,
			s.task_id,
			COALESCE(t.summary, '') AS task_summary,
			COALESCE(SUM(s.duration_seconds), 0) AS duration_seconds
		FROM sessions s
		LEFT JOIN projects p ON p.id = s.project_id
		LEFT JOIN tasks t ON t.id = s.task_id
		WHERE date(s.started_at) = ?
		GROUP BY s.project_id, s.task_id
		ORDER BY duration_seconds DESC
	`, today)
	if err != nil {
		return nil, fmt.Errorf("store: today summary: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var summary []SummaryRow
	for rows.Next() {
		var r SummaryRow
		if err := rows.Scan(&r.ProjectID, &r.ProjectName, &r.TaskID, &r.TaskSummary, &r.DurationSeconds); err != nil {
			return nil, fmt.Errorf("store: scan summary: %w", err)
		}
		summary = append(summary, r)
	}
	return summary, rows.Err()
}

// --- Recents ---

// TouchRecent upserts a task into the recents table with the current timestamp.
func (s *Store) TouchRecent(taskID, projectID string) error {
	_, err := s.db.Exec(`
		INSERT INTO recents (task_id, project_id, used_at)
		VALUES (?, ?, ?)
		ON CONFLICT(task_id) DO UPDATE SET project_id=excluded.project_id, used_at=excluded.used_at
	`, taskID, projectID, fmtTime(time.Now()))
	return err
}

// ListRecents returns the most recent N task entries.
func (s *Store) ListRecents(limit int) ([]RecentRow, error) {
	rows, err := s.db.Query(
		"SELECT task_id, project_id, used_at FROM recents ORDER BY used_at DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list recents: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var recents []RecentRow
	for rows.Next() {
		var r RecentRow
		var usedAt string
		if err := rows.Scan(&r.TaskID, &r.ProjectID, &usedAt); err != nil {
			return nil, fmt.Errorf("store: scan recent: %w", err)
		}
		r.UsedAt = parseTime(usedAt)
		recents = append(recents, r)
	}
	return recents, rows.Err()
}

// InvalidateAll clears all TTL timestamps for projects and tasks, forcing
// re-fetch on next access.
func (s *Store) InvalidateAll() error {
	epoch := fmtTime(time.Time{})
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec("UPDATE projects SET updated_at = ?", epoch); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE tasks SET updated_at = ?", epoch); err != nil {
		return err
	}
	return tx.Commit()
}
