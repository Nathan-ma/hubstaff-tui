package ui

import (
	"strings"
	"testing"

	"github.com/Nathan-ma/hubstaff-tui/internal/store"
)

func TestNewHistoryModel(t *testing.T) {
	theme := GetTheme("default")
	m := NewHistoryModel(theme)
	if m.ready {
		t.Error("expected ready=false on new model")
	}
}

func TestHistoryModel_View_NotReady(t *testing.T) {
	theme := GetTheme("default")
	m := NewHistoryModel(theme)
	view := m.View()
	if !strings.Contains(view, "Loading") {
		t.Errorf("expected loading message, got: %q", view)
	}
}

func TestHistoryModel_SetRows_Empty(t *testing.T) {
	theme := GetTheme("default")
	m := NewHistoryModel(theme)
	m.SetSize(80, 24)
	m.SetRows(nil)
	view := m.View()
	if !strings.Contains(view, "No sessions") {
		t.Errorf("expected no-sessions message, got: %q", view)
	}
}

func TestHistoryModel_SetRows(t *testing.T) {
	theme := GetTheme("default")
	m := NewHistoryModel(theme)
	m.SetSize(100, 30)

	rows := []store.HistorySummaryRow{
		{Date: "2026-03-23", ProjectName: "Mobile App", TaskSummary: "Fix login bug", DurationSeconds: 8100},
		{Date: "2026-03-23", ProjectName: "Backend", TaskSummary: "Database migration", DurationSeconds: 4500},
		{Date: "2026-03-22", ProjectName: "Mobile App", TaskSummary: "Add OAuth", DurationSeconds: 3600},
	}
	m.SetRows(rows)

	if !m.ready {
		t.Error("expected ready=true after SetRows")
	}
	view := m.View()
	if !strings.Contains(view, "2026-03-23") {
		t.Errorf("expected date 2026-03-23 in view, got: %q", view)
	}
	if !strings.Contains(view, "Mobile App") {
		t.Errorf("expected project name in view, got: %q", view)
	}
	if !strings.Contains(view, "Fix login bug") {
		t.Errorf("expected task name in view, got: %q", view)
	}
}

func TestHistoryModel_SetSize(t *testing.T) {
	theme := GetTheme("default")
	m := NewHistoryModel(theme)
	m.SetSize(120, 40)
	if m.width != 120 || m.height != 40 {
		t.Errorf("expected 120x40, got %dx%d", m.width, m.height)
	}
}

func TestFormatDateHeader(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2026-03-23", "2026-03-23 (Mon)"},
		{"2026-03-22", "2026-03-22 (Sun)"},
		{"invalid", "invalid"},
	}
	for _, tc := range tests {
		got := formatDateHeader(tc.input)
		if got != tc.expected {
			t.Errorf("formatDateHeader(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}
