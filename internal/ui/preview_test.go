package ui

import (
	"strings"
	"testing"

	"github.com/Nathan-ma/hubstaff-tui/internal/api"
)

func TestNewPreviewModel(t *testing.T) {
	theme := PlainTheme()
	m := NewPreviewModel(theme)
	if m.width != 0 {
		t.Errorf("expected width=0 on new model, got %d", m.width)
	}
	if m.todayLoaded {
		t.Error("expected todayLoaded=false on new model")
	}
}

func TestPreviewModel_View_ZeroWidth(t *testing.T) {
	theme := PlainTheme()
	m := NewPreviewModel(theme)
	view := m.View()
	if view != "" {
		t.Errorf("expected empty string when width=0, got: %q", view)
	}
}

func TestPreviewModel_View_NoTask(t *testing.T) {
	theme := PlainTheme()
	m := NewPreviewModel(theme)
	m.SetSize(80, 24)
	view := m.View()
	if !strings.Contains(view, "Select a task to preview") {
		t.Errorf("expected placeholder message, got: %q", view)
	}
}

func TestPreviewModel_SetTask_SetsFields(t *testing.T) {
	theme := PlainTheme()
	m := NewPreviewModel(theme)
	task := api.Task{ID: api.FlexibleID("t1"), Summary: "Fix the bug"}
	m.SetTask(task, "My Project", false)
	if string(m.task.ID) != "t1" {
		t.Errorf("expected task ID 't1', got %q", m.task.ID)
	}
	if m.projectName != "My Project" {
		t.Errorf("expected projectName='My Project', got %q", m.projectName)
	}
	if m.tracking {
		t.Error("expected tracking=false")
	}
	if m.todayLoaded {
		t.Error("expected todayLoaded=false after SetTask (reset)")
	}
	if m.todaySecs != 0 {
		t.Errorf("expected todaySecs=0 after SetTask reset, got %d", m.todaySecs)
	}
}

func TestPreviewModel_View_WithTask(t *testing.T) {
	theme := PlainTheme()
	m := NewPreviewModel(theme)
	m.SetSize(80, 24)
	task := api.Task{ID: api.FlexibleID("t1"), Summary: "Implement feature"}
	m.SetTask(task, "Alpha Project", false)
	view := m.View()
	if !strings.Contains(view, "Implement feature") {
		t.Errorf("expected task summary in view, got: %q", view)
	}
	if !strings.Contains(view, "Alpha Project") {
		t.Errorf("expected project name in view, got: %q", view)
	}
}

func TestPreviewModel_View_TodayNotLoaded(t *testing.T) {
	theme := PlainTheme()
	m := NewPreviewModel(theme)
	m.SetSize(80, 24)
	task := api.Task{ID: api.FlexibleID("t1"), Summary: "Some Task"}
	m.SetTask(task, "Project", false)
	// todayLoaded is false by default after SetTask.
	view := m.View()
	if !strings.Contains(view, "…") {
		t.Errorf("expected '…' when todayLoaded=false, got: %q", view)
	}
}

func TestPreviewModel_SetTodaySeconds_Zero(t *testing.T) {
	theme := PlainTheme()
	m := NewPreviewModel(theme)
	m.SetSize(80, 24)
	task := api.Task{ID: api.FlexibleID("t1"), Summary: "Some Task"}
	m.SetTask(task, "Project", false)
	m.SetTodaySeconds(0)
	view := m.View()
	if !strings.Contains(view, "No time logged") {
		t.Errorf("expected 'No time logged' when todaySecs=0, got: %q", view)
	}
}

func TestPreviewModel_SetTodaySeconds_NonZero(t *testing.T) {
	theme := PlainTheme()
	m := NewPreviewModel(theme)
	m.SetSize(80, 24)
	task := api.Task{ID: api.FlexibleID("t1"), Summary: "Some Task"}
	m.SetTask(task, "Project", false)
	m.SetTodaySeconds(3600)
	view := m.View()
	if !strings.Contains(view, "1h 00m") {
		t.Errorf("expected '1h 00m' for 3600 seconds, got: %q", view)
	}
}

func TestPreviewModel_SetTodaySeconds_SetsLoaded(t *testing.T) {
	theme := PlainTheme()
	m := NewPreviewModel(theme)
	m.SetTodaySeconds(120)
	if !m.todayLoaded {
		t.Error("expected todayLoaded=true after SetTodaySeconds")
	}
	if m.todaySecs != 120 {
		t.Errorf("expected todaySecs=120, got %d", m.todaySecs)
	}
}

func TestPreviewModel_View_TrackingNow(t *testing.T) {
	theme := PlainTheme()
	m := NewPreviewModel(theme)
	m.SetSize(80, 24)
	task := api.Task{ID: api.FlexibleID("t1"), Summary: "Active Task"}
	m.SetTask(task, "Project", true)
	view := m.View()
	if !strings.Contains(view, "Tracking now") {
		t.Errorf("expected 'Tracking now' when tracking=true, got: %q", view)
	}
}

func TestPreviewModel_View_NotTracking(t *testing.T) {
	theme := PlainTheme()
	m := NewPreviewModel(theme)
	m.SetSize(80, 24)
	task := api.Task{ID: api.FlexibleID("t1"), Summary: "Inactive Task"}
	m.SetTask(task, "Project", false)
	view := m.View()
	if !strings.Contains(view, "Not tracking") {
		t.Errorf("expected 'Not tracking' when tracking=false, got: %q", view)
	}
}

func TestPreviewModel_SetSize(t *testing.T) {
	theme := PlainTheme()
	m := NewPreviewModel(theme)
	m.SetSize(100, 50)
	if m.width != 100 || m.height != 50 {
		t.Errorf("expected 100x50, got %dx%d", m.width, m.height)
	}
}

func TestPreviewModel_View_ReturnsEmptyWhenWidthZeroAfterSetTask(t *testing.T) {
	theme := PlainTheme()
	m := NewPreviewModel(theme)
	task := api.Task{ID: api.FlexibleID("t1"), Summary: "Some Task"}
	m.SetTask(task, "Project", false)
	// Width is still 0.
	view := m.View()
	if view != "" {
		t.Errorf("expected empty string when width=0 even with task set, got: %q", view)
	}
}

func TestPreviewModel_SetTask_ResetsTracking(t *testing.T) {
	theme := PlainTheme()
	m := NewPreviewModel(theme)
	m.SetSize(80, 24)

	task1 := api.Task{ID: api.FlexibleID("t1"), Summary: "First"}
	m.SetTask(task1, "Project A", true)
	m.SetTodaySeconds(1800)

	// Now set a different task without tracking.
	task2 := api.Task{ID: api.FlexibleID("t2"), Summary: "Second"}
	m.SetTask(task2, "Project B", false)

	if m.tracking {
		t.Error("expected tracking=false after new SetTask")
	}
	if m.todayLoaded {
		t.Error("expected todayLoaded=false after new SetTask")
	}
	if m.todaySecs != 0 {
		t.Errorf("expected todaySecs=0 after new SetTask, got %d", m.todaySecs)
	}
}

func TestPreviewModel_View_DurationVariants(t *testing.T) {
	theme := PlainTheme()

	tests := []struct {
		secs     int
		expected string
	}{
		{60, "1m"},
		{3660, "1h 01m"},
		{7200, "2h 00m"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			m := NewPreviewModel(theme)
			m.SetSize(80, 24)
			task := api.Task{ID: api.FlexibleID("t1"), Summary: "Task"}
			m.SetTask(task, "Project", false)
			m.SetTodaySeconds(tc.secs)
			view := m.View()
			if !strings.Contains(view, tc.expected) {
				t.Errorf("SetTodaySeconds(%d): expected %q in view, got: %q", tc.secs, tc.expected, view)
			}
		})
	}
}
