package ui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Nathan-ma/hubstaff-tui/internal/api"
	"github.com/Nathan-ma/hubstaff-tui/internal/store"
)

func TestNewTasksModel(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	if m.loaded {
		t.Error("expected loaded=false on new model")
	}
	if m.loading {
		t.Error("expected loading=false on new model")
	}
	if m.loadErr != nil {
		t.Error("expected loadErr=nil on new model")
	}
}

func TestTasksModel_View_Initial(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	// Neither loading nor loaded — list view.
	view := m.View()
	// Should not show loading or error states.
	if strings.Contains(view, "Failed to load tasks") {
		t.Errorf("unexpected error in initial view: %q", view)
	}
}

func TestTasksModel_SetProject_SetsLoadingState(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	m.SetProject("proj-1", "Backend")
	if !m.loading {
		t.Error("expected loading=true after SetProject")
	}
	if m.loaded {
		t.Error("expected loaded=false after SetProject")
	}
	if m.projectID != "proj-1" {
		t.Errorf("expected projectID='proj-1', got %q", m.projectID)
	}
	if m.projectName != "Backend" {
		t.Errorf("expected projectName='Backend', got %q", m.projectName)
	}
}

func TestTasksModel_SetProject_ClearsError(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	m.SetError(errors.New("previous error"))
	m.SetProject("proj-1", "Backend")
	if m.loadErr != nil {
		t.Error("expected loadErr cleared after SetProject")
	}
}

func TestTasksModel_View_Loading(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	m.SetProject("proj-1", "My Project")
	view := m.View()
	if !strings.Contains(view, "Loading tasks for") {
		t.Errorf("expected loading message, got: %q", view)
	}
	if !strings.Contains(view, "My Project") {
		t.Errorf("expected project name in loading message, got: %q", view)
	}
}

func TestTasksModel_SetTasks_SetsLoadedState(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	m.SetProject("proj-1", "Backend")
	tasks := []api.Task{
		{ID: api.FlexibleID("t1"), Summary: "Task One"},
		{ID: api.FlexibleID("t2"), Summary: "Task Two"},
	}
	m.SetTasks(tasks, api.Status{})
	if !m.loaded {
		t.Error("expected loaded=true after SetTasks")
	}
	if m.loading {
		t.Error("expected loading=false after SetTasks")
	}
	if len(m.tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(m.tasks))
	}
}

func TestTasksModel_View_Error(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	m.SetError(errors.New("api failure"))
	view := m.View()
	if !strings.Contains(view, "Failed to load tasks") {
		t.Errorf("expected error message, got: %q", view)
	}
}

func TestTasksModel_View_Empty(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	m.SetProject("proj-1", "Backend")
	m.SetTasks([]api.Task{}, api.Status{})
	view := m.View()
	if !strings.Contains(view, "No tasks found") {
		t.Errorf("expected empty message, got: %q", view)
	}
}

func TestTasksModel_View_WithTasks(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	m.SetSize(80, 24)
	m.SetProject("proj-1", "Backend")
	tasks := []api.Task{
		{ID: api.FlexibleID("t1"), Summary: "Implement login"},
	}
	m.SetTasks(tasks, api.Status{})
	view := m.View()
	if strings.Contains(view, "Loading tasks") {
		t.Error("should not show loading state after SetTasks")
	}
	if strings.Contains(view, "No tasks found") {
		t.Error("should not show empty state after SetTasks with data")
	}
}

func TestTasksModel_SetError_SetsLoaded(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	m.SetProject("proj-1", "Backend")
	m.SetError(errors.New("boom"))
	if !m.loaded {
		t.Error("expected loaded=true after SetError")
	}
	if m.loading {
		t.Error("expected loading=false after SetError")
	}
}

func TestTasksModel_SelectedTask_NoTasks(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	_, ok := m.SelectedTask()
	if ok {
		t.Error("expected SelectedTask to return false when no tasks loaded")
	}
}

func TestTasksModel_SelectedTask_WithTasks(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	m.SetSize(80, 24)
	m.SetProject("proj-1", "Backend")
	tasks := []api.Task{
		{ID: api.FlexibleID("t1"), Summary: "Fix bug"},
		{ID: api.FlexibleID("t2"), Summary: "Write tests"},
	}
	m.SetTasks(tasks, api.Status{})

	task, ok := m.SelectedTask()
	if !ok {
		t.Fatal("expected SelectedTask to return true")
	}
	if string(task.ID) != "t1" {
		t.Errorf("expected first task selected, got ID=%q", task.ID)
	}
}

func TestTasksModel_SetRecents_WithMatchingProject(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	m.SetSize(80, 24)
	m.SetProject("proj-1", "Backend")

	tasks := []api.Task{
		{ID: api.FlexibleID("t1"), Summary: "Task One"},
		{ID: api.FlexibleID("t2"), Summary: "Task Two"},
		{ID: api.FlexibleID("t3"), Summary: "Task Three"},
	}
	m.SetTasks(tasks, api.Status{})

	recents := []store.RecentRow{
		{TaskID: "t2", ProjectID: "proj-1", UsedAt: time.Now()},
	}
	m.SetRecents(recents)

	// After setting recents, the list should have a separator item first.
	items := m.list.Items()
	if len(items) == 0 {
		t.Fatal("expected items after SetRecents")
	}
	_, isSep := items[0].(separatorItem)
	if !isSep {
		t.Error("expected first item to be a separatorItem after SetRecents with recents")
	}

	// The separator label should contain "Recent".
	sep := items[0].(separatorItem)
	if !strings.Contains(sep.label, "Recent") {
		t.Errorf("expected separator label to contain 'Recent', got %q", sep.label)
	}
}

func TestTasksModel_SetRecents_NoMatchingProject(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	m.SetSize(80, 24)
	m.SetProject("proj-1", "Backend")

	tasks := []api.Task{
		{ID: api.FlexibleID("t1"), Summary: "Task One"},
	}
	m.SetTasks(tasks, api.Status{})

	// Recents for a different project — should not affect this project's list.
	recents := []store.RecentRow{
		{TaskID: "t1", ProjectID: "proj-999", UsedAt: time.Now()},
	}
	m.SetRecents(recents)

	items := m.list.Items()
	if len(items) == 0 {
		t.Fatal("expected items")
	}
	_, isSep := items[0].(separatorItem)
	if isSep {
		t.Error("expected no separator when recents don't match project")
	}
}

func TestTasksModel_SelectedTask_ReturnsFalseForSeparator(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	m.SetSize(80, 24)
	m.SetProject("proj-1", "Backend")

	tasks := []api.Task{
		{ID: api.FlexibleID("t1"), Summary: "Task One"},
		{ID: api.FlexibleID("t2"), Summary: "Task Two"},
	}
	m.SetTasks(tasks, api.Status{})

	recents := []store.RecentRow{
		{TaskID: "t1", ProjectID: "proj-1", UsedAt: time.Now()},
	}
	m.SetRecents(recents)

	// The first item is a separator. Verify SelectedTask returns false for it.
	items := m.list.Items()
	if len(items) == 0 {
		t.Fatal("expected items")
	}
	_, isSep := items[0].(separatorItem)
	if !isSep {
		t.Skip("first item is not a separator; skipping separator selection test")
	}

	// The list starts at index 0 (the separator). SelectedTask should return false.
	_, ok := m.SelectedTask()
	if ok {
		t.Error("expected SelectedTask to return false when selection is on separator")
	}
}

func TestTasksModel_TrackingFlag(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	m.SetProject("proj-1", "Backend")

	tasks := []api.Task{
		{ID: api.FlexibleID("t1"), Summary: "Active Task"},
		{ID: api.FlexibleID("t2"), Summary: "Other Task"},
	}
	status := api.Status{
		Tracking:   true,
		ActiveTask: api.ActiveTask{ID: api.FlexibleID("t1")},
	}
	m.SetTasks(tasks, status)

	items := m.list.Items()
	ti0, ok := items[0].(taskItem)
	if !ok {
		t.Fatal("item[0] is not taskItem")
	}
	if !ti0.tracking {
		t.Error("expected item[0] tracking=true")
	}

	ti1, ok := items[1].(taskItem)
	if !ok {
		t.Fatal("item[1] is not taskItem")
	}
	if ti1.tracking {
		t.Error("expected item[1] tracking=false")
	}
}

func TestTasksModel_ActiveFlag(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	m.SetProject("proj-1", "Backend")

	tasks := []api.Task{
		{ID: api.FlexibleID("t1"), Summary: "Last Active Task"},
	}
	status := api.Status{
		Tracking:   false,
		ActiveTask: api.ActiveTask{ID: api.FlexibleID("t1")},
	}
	m.SetTasks(tasks, status)

	items := m.list.Items()
	ti0, ok := items[0].(taskItem)
	if !ok {
		t.Fatal("item[0] is not taskItem")
	}
	if ti0.tracking {
		t.Error("expected tracking=false when status.Tracking=false")
	}
	if !ti0.active {
		t.Error("expected active=true for last-tracked task")
	}
}

func TestTasksModel_SetSize(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	// Verify no panic.
	m.SetSize(100, 40)
}

func TestTasksModel_Update_NoError(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	m2, _ := m.Update(nil)
	if m2.loaded {
		t.Error("expected loaded=false after Update with no SetTasks")
	}
}

func TestTasksModel_ErrorView_ContainsDetail(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	m.SetError(errors.New("deadline exceeded"))
	view := m.View()
	if !strings.Contains(view, "deadline exceeded") {
		t.Errorf("expected error detail in view, got: %q", view)
	}
}

func TestTasksModel_SetProject_ClearsItems(t *testing.T) {
	theme := PlainTheme()
	m := NewTasksModel(theme)
	m.SetProject("proj-1", "Alpha")
	tasks := []api.Task{
		{ID: api.FlexibleID("t1"), Summary: "Task One"},
	}
	m.SetTasks(tasks, api.Status{})

	// Switch to a different project — items should be cleared.
	m.SetProject("proj-2", "Beta")
	items := m.list.Items()
	if len(items) != 0 {
		t.Errorf("expected empty list after SetProject, got %d items", len(items))
	}
}
