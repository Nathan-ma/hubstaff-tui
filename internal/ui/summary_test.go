package ui

import (
	"strings"
	"testing"

	"github.com/Nathan-ma/hubstaff-tui/internal/store"
)

func TestNewSummaryModel(t *testing.T) {
	theme := PlainTheme()
	m := NewSummaryModel(theme)
	if m.ready {
		t.Error("expected ready=false on new model")
	}
}

func TestSummaryModel_View_NotReady(t *testing.T) {
	theme := PlainTheme()
	m := NewSummaryModel(theme)
	view := m.View()
	if view != "Loading summary..." {
		t.Errorf("expected 'Loading summary...', got: %q", view)
	}
}

func TestSummaryModel_SetRows_Empty(t *testing.T) {
	theme := PlainTheme()
	m := NewSummaryModel(theme)
	m.SetSize(80, 24)
	m.SetRows(nil)
	if !m.ready {
		t.Error("expected ready=true after SetRows")
	}
	view := m.View()
	if !strings.Contains(view, "No time tracked today") {
		t.Errorf("expected 'No time tracked today', got: %q", view)
	}
}

func TestSummaryModel_SetRows_EmptySlice(t *testing.T) {
	theme := PlainTheme()
	m := NewSummaryModel(theme)
	m.SetSize(80, 24)
	m.SetRows([]store.SummaryRow{})
	view := m.View()
	if !strings.Contains(view, "No time tracked today") {
		t.Errorf("expected 'No time tracked today', got: %q", view)
	}
}

func TestSummaryModel_SetRows_SetsReady(t *testing.T) {
	theme := PlainTheme()
	m := NewSummaryModel(theme)
	m.SetSize(80, 24)
	rows := []store.SummaryRow{
		{ProjectID: "p1", ProjectName: "Alpha", TaskID: "t1", TaskSummary: "Do work", DurationSeconds: 3600},
	}
	m.SetRows(rows)
	if !m.ready {
		t.Error("expected ready=true after SetRows")
	}
}

func TestSummaryModel_SetRows_WithData(t *testing.T) {
	theme := PlainTheme()
	m := NewSummaryModel(theme)
	m.SetSize(120, 30)
	rows := []store.SummaryRow{
		{ProjectID: "p1", ProjectName: "Mobile App", TaskID: "t1", TaskSummary: "Fix login", DurationSeconds: 3600},
		{ProjectID: "p2", ProjectName: "Backend", TaskID: "t2", TaskSummary: "Migration", DurationSeconds: 1800},
	}
	m.SetRows(rows)
	view := m.View()

	if !strings.Contains(view, "Mobile App") {
		t.Errorf("expected project name 'Mobile App' in view, got: %q", view)
	}
	if !strings.Contains(view, "Fix login") {
		t.Errorf("expected task summary 'Fix login' in view, got: %q", view)
	}
	if !strings.Contains(view, "Backend") {
		t.Errorf("expected project name 'Backend' in view, got: %q", view)
	}
}

func TestSummaryModel_SetSize(t *testing.T) {
	theme := PlainTheme()
	m := NewSummaryModel(theme)
	m.SetSize(100, 40)
	if m.width != 100 || m.height != 40 {
		t.Errorf("expected 100x40, got %dx%d", m.width, m.height)
	}
}

func TestSummaryModel_SetSize_UpdatesViewport(t *testing.T) {
	theme := PlainTheme()
	m := NewSummaryModel(theme)
	m.SetSize(80, 24)
	m.SetRows([]store.SummaryRow{}) // marks ready=true
	m.SetSize(100, 40)
	if m.viewport.Width != 100 || m.viewport.Height != 40 {
		t.Errorf("expected viewport 100x40, got %dx%d", m.viewport.Width, m.viewport.Height)
	}
}

func TestSummaryModel_Update_NotReady(t *testing.T) {
	theme := PlainTheme()
	m := NewSummaryModel(theme)
	m2, cmd := m.Update(nil)
	if cmd != nil {
		t.Error("expected nil cmd when not ready")
	}
	if m2.ready {
		t.Error("expected ready=false after Update when not initialized")
	}
}

func TestSummaryModel_FallsBackToID_WhenNameEmpty(t *testing.T) {
	theme := PlainTheme()
	m := NewSummaryModel(theme)
	m.SetSize(120, 30)
	rows := []store.SummaryRow{
		{ProjectID: "proj-42", ProjectName: "", TaskID: "task-99", TaskSummary: "", DurationSeconds: 60},
	}
	m.SetRows(rows)
	view := m.View()
	if !strings.Contains(view, "proj-42") {
		t.Errorf("expected project ID as fallback, got: %q", view)
	}
	if !strings.Contains(view, "task-99") {
		t.Errorf("expected task ID as fallback, got: %q", view)
	}
}

// --- formatCompactDuration ---

func TestFormatCompactDuration(t *testing.T) {
	tests := []struct {
		seconds  int
		expected string
	}{
		{0, "0m"},
		{59, "0m"}, // 59 seconds = 0 minutes (truncated)
		{60, "1m"},
		{90, "1m"},
		{3600, "1h 00m"},
		{3661, "1h 01m"},
		{7200, "2h 00m"},
		{7260, "2h 01m"},
		{-1, "0m"}, // negative input treated as 0
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			got := formatCompactDuration(tc.seconds)
			if got != tc.expected {
				t.Errorf("formatCompactDuration(%d) = %q, want %q", tc.seconds, got, tc.expected)
			}
		})
	}
}

// --- padRight ---

func TestPadRight(t *testing.T) {
	tests := []struct {
		s        string
		width    int
		expected string
	}{
		{"hi", 5, "hi   "},
		{"hello", 5, "hello"},
		{"toolong", 4, "tool"},
		{"", 3, "   "},
	}
	for _, tc := range tests {
		got := padRight(tc.s, tc.width)
		if got != tc.expected {
			t.Errorf("padRight(%q, %d) = %q, want %q", tc.s, tc.width, got, tc.expected)
		}
	}
}

// --- padLeft ---

func TestPadLeft(t *testing.T) {
	tests := []struct {
		s        string
		width    int
		expected string
	}{
		{"hi", 5, "   hi"},
		{"hello", 5, "hello"},
		{"toolong", 4, "toolong"}, // no truncation in padLeft
		{"", 3, "   "},
	}
	for _, tc := range tests {
		got := padLeft(tc.s, tc.width)
		if got != tc.expected {
			t.Errorf("padLeft(%q, %d) = %q, want %q", tc.s, tc.width, got, tc.expected)
		}
	}
}

// --- truncate ---

func TestTruncate(t *testing.T) {
	tests := []struct {
		s        string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 8, "hello w…"},
		{"hi", 1, "h"},
		{"ab", 2, "ab"},
		{"abc", 2, "a…"},
		{"", 5, ""},
	}
	for _, tc := range tests {
		got := truncate(tc.s, tc.maxLen)
		if got != tc.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tc.s, tc.maxLen, got, tc.expected)
		}
	}
}
