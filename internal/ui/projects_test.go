package ui

import (
	"errors"
	"strings"
	"testing"

	"github.com/Nathan-ma/hubstaff-tui/internal/api"
)

func TestNewProjectsModel(t *testing.T) {
	theme := PlainTheme()
	m := NewProjectsModel(theme)
	if m.loaded {
		t.Error("expected loaded=false on new model")
	}
	if m.loadErr != nil {
		t.Error("expected loadErr=nil on new model")
	}
	if len(m.projects) != 0 {
		t.Error("expected empty projects slice on new model")
	}
}

func TestProjectsModel_View_Loading(t *testing.T) {
	theme := PlainTheme()
	m := NewProjectsModel(theme)
	view := m.View()
	if !strings.Contains(view, "Loading projects") {
		t.Errorf("expected loading message, got: %q", view)
	}
}

func TestProjectsModel_View_Error(t *testing.T) {
	theme := PlainTheme()
	m := NewProjectsModel(theme)
	m.SetError(errors.New("network timeout"))
	view := m.View()
	if !strings.Contains(view, "Failed to load projects") {
		t.Errorf("expected error message, got: %q", view)
	}
}

func TestProjectsModel_View_Empty(t *testing.T) {
	theme := PlainTheme()
	m := NewProjectsModel(theme)
	m.SetProjects([]api.Project{}, api.Status{})
	view := m.View()
	if !strings.Contains(view, "No projects found") {
		t.Errorf("expected empty message, got: %q", view)
	}
}

func TestProjectsModel_SetProjects_SetsLoaded(t *testing.T) {
	theme := PlainTheme()
	m := NewProjectsModel(theme)
	projects := []api.Project{
		{ID: api.FlexibleID("1"), Name: "Alpha"},
		{ID: api.FlexibleID("2"), Name: "Beta"},
	}
	m.SetProjects(projects, api.Status{})
	if !m.loaded {
		t.Error("expected loaded=true after SetProjects")
	}
	if m.loadErr != nil {
		t.Error("expected loadErr=nil after SetProjects")
	}
	if len(m.projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(m.projects))
	}
}

func TestProjectsModel_SetProjects_ClearsError(t *testing.T) {
	theme := PlainTheme()
	m := NewProjectsModel(theme)
	m.SetError(errors.New("previous error"))
	m.SetProjects([]api.Project{{ID: api.FlexibleID("1"), Name: "Alpha"}}, api.Status{})
	if m.loadErr != nil {
		t.Error("expected loadErr to be cleared after SetProjects")
	}
}

func TestProjectsModel_SetProjects_TrackingFlag(t *testing.T) {
	theme := PlainTheme()
	m := NewProjectsModel(theme)

	projects := []api.Project{
		{ID: api.FlexibleID("42"), Name: "Tracked Project"},
		{ID: api.FlexibleID("99"), Name: "Other Project"},
	}
	status := api.Status{
		Tracking: true,
		ActiveProject: api.ActiveProject{
			ID:   api.FlexibleID("42"),
			Name: "Tracked Project",
		},
	}
	m.SetProjects(projects, status)

	// Verify tracking item has tracking=true in the list
	items := m.list.Items()
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	pi0, ok := items[0].(projectItem)
	if !ok {
		t.Fatal("item[0] is not projectItem")
	}
	if !pi0.tracking {
		t.Error("expected item[0] tracking=true")
	}
	if pi0.active {
		t.Error("expected item[0] active=false when tracking=true")
	}

	pi1, ok := items[1].(projectItem)
	if !ok {
		t.Fatal("item[1] is not projectItem")
	}
	if pi1.tracking {
		t.Error("expected item[1] tracking=false")
	}
}

func TestProjectsModel_SetProjects_ActiveFlag(t *testing.T) {
	theme := PlainTheme()
	m := NewProjectsModel(theme)

	projects := []api.Project{
		{ID: api.FlexibleID("42"), Name: "Last Project"},
		{ID: api.FlexibleID("99"), Name: "Other Project"},
	}
	status := api.Status{
		Tracking: false,
		ActiveProject: api.ActiveProject{
			ID:   api.FlexibleID("42"),
			Name: "Last Project",
		},
	}
	m.SetProjects(projects, status)

	items := m.list.Items()
	pi0, ok := items[0].(projectItem)
	if !ok {
		t.Fatal("item[0] is not projectItem")
	}
	if pi0.tracking {
		t.Error("expected item[0] tracking=false when status.Tracking=false")
	}
	if !pi0.active {
		t.Error("expected item[0] active=true for last-tracked project")
	}
}

func TestProjectsModel_SetProjects_NoActiveWhenIDEmpty(t *testing.T) {
	theme := PlainTheme()
	m := NewProjectsModel(theme)

	projects := []api.Project{
		{ID: api.FlexibleID("42"), Name: "Project"},
	}
	// Status with empty ActiveProject ID and Tracking=false
	status := api.Status{
		Tracking:      false,
		ActiveProject: api.ActiveProject{ID: api.FlexibleID("")},
	}
	m.SetProjects(projects, status)

	items := m.list.Items()
	pi0, ok := items[0].(projectItem)
	if !ok {
		t.Fatal("item[0] is not projectItem")
	}
	if pi0.active {
		t.Error("expected active=false when ActiveProject.ID is empty")
	}
}

func TestProjectsModel_SelectedProject_NoProjects(t *testing.T) {
	theme := PlainTheme()
	m := NewProjectsModel(theme)
	_, ok := m.SelectedProject()
	if ok {
		t.Error("expected SelectedProject to return false when no projects loaded")
	}
}

func TestProjectsModel_SelectedProject_WithProjects(t *testing.T) {
	theme := PlainTheme()
	m := NewProjectsModel(theme)
	m.SetSize(80, 24)
	projects := []api.Project{
		{ID: api.FlexibleID("1"), Name: "First"},
		{ID: api.FlexibleID("2"), Name: "Second"},
	}
	m.SetProjects(projects, api.Status{})

	proj, ok := m.SelectedProject()
	if !ok {
		t.Fatal("expected SelectedProject to return true when projects are loaded")
	}
	if string(proj.ID) != "1" {
		t.Errorf("expected first project selected, got ID=%q", proj.ID)
	}
	if proj.Name != "First" {
		t.Errorf("expected project name 'First', got %q", proj.Name)
	}
}

func TestProjectsModel_SetError_SetsLoaded(t *testing.T) {
	theme := PlainTheme()
	m := NewProjectsModel(theme)
	m.SetError(errors.New("boom"))
	if !m.loaded {
		t.Error("expected loaded=true after SetError")
	}
	if m.loadErr == nil {
		t.Error("expected loadErr to be set after SetError")
	}
}

func TestProjectsModel_SetSize(t *testing.T) {
	theme := PlainTheme()
	m := NewProjectsModel(theme)
	// Just verify it doesn't panic.
	m.SetSize(100, 40)
}

func TestProjectsModel_Update_NoError(t *testing.T) {
	theme := PlainTheme()
	m := NewProjectsModel(theme)
	// Verify Update returns without panic in unloaded state.
	m2, _ := m.Update(nil)
	if m2.loaded {
		t.Error("expected loaded=false after Update with no SetProjects")
	}
}

func TestProjectsModel_View_WithProjects(t *testing.T) {
	theme := PlainTheme()
	m := NewProjectsModel(theme)
	m.SetSize(80, 24)
	projects := []api.Project{
		{ID: api.FlexibleID("1"), Name: "My Project"},
	}
	m.SetProjects(projects, api.Status{})
	view := m.View()
	// The list view should be rendered (not loading, not empty, not error).
	if strings.Contains(view, "Loading") {
		t.Error("should not show loading state after SetProjects")
	}
	if strings.Contains(view, "No projects found") {
		t.Error("should not show empty state after SetProjects with data")
	}
}

func TestProjectsModel_ErrorView_ContainsErrorDetails(t *testing.T) {
	theme := PlainTheme()
	m := NewProjectsModel(theme)
	m.SetError(errors.New("connection refused"))
	view := m.View()
	if !strings.Contains(view, "connection refused") {
		t.Errorf("expected error detail in view, got: %q", view)
	}
}
