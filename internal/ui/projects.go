package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/Nathan-ma/hubstaff-tui/internal/api"
)

// projectItem wraps an api.Project for use in a bubbles/list.
type projectItem struct {
	project  api.Project
	tracking bool // currently being tracked
	active   bool // was the last tracked project
}

func (i projectItem) Title() string       { return i.project.Name }
func (i projectItem) Description() string { return "" }
func (i projectItem) FilterValue() string { return i.project.Name }

// ProjectsModel holds the state for the projects list screen.
type ProjectsModel struct {
	list     list.Model
	projects []api.Project
	status   api.Status
	theme    Theme
}

// projectDelegate renders project items with tracking indicators.
type projectDelegate struct {
	theme Theme
}

func (d projectDelegate) Height() int                             { return 1 }
func (d projectDelegate) Spacing() int                            { return 0 }
func (d projectDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d projectDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	pi, ok := item.(projectItem)
	if !ok {
		return
	}

	indicator := d.theme.TrackingIndicator(pi.tracking, pi.active)

	name := pi.project.Name
	isSelected := index == m.Index()

	// Truncate name to fit available width (leave room for indicator + spacing)
	maxWidth := m.Width() - 4
	if maxWidth > 0 {
		name = ansi.Truncate(name, maxWidth, "...")
	}

	var line string
	if isSelected {
		line = d.theme.SelectedItem.Render(fmt.Sprintf("%s %s", indicator, name))
	} else {
		line = d.theme.NormalItem.Render(fmt.Sprintf("%s %s", indicator, name))
	}

	_, _ = fmt.Fprint(w, line)
}

// NewProjectsModel creates a new projects model.
func NewProjectsModel(theme Theme) ProjectsModel {
	delegate := projectDelegate{theme: theme}
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Projects"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowTitle(true)
	l.SetStatusBarItemName("project", "projects")
	l.Styles.Title = theme.HeaderTitle
	l.FilterInput.PromptStyle = theme.FilterPrompt
	l.FilterInput.TextStyle = theme.FilterText

	return ProjectsModel{
		list:  l,
		theme: theme,
	}
}

// SetSize updates the list dimensions.
func (m *ProjectsModel) SetSize(width, height int) {
	m.list.SetSize(width, height)
}

// SetProjects updates the project list items, marking tracking state.
func (m *ProjectsModel) SetProjects(projects []api.Project, status api.Status) {
	m.projects = projects
	m.status = status

	items := make([]list.Item, len(projects))
	for i, p := range projects {
		tracking := status.Tracking && status.ActiveProject.ID == p.ID
		active := !status.Tracking && status.ActiveProject.ID == p.ID && status.ActiveProject.ID != ""
		items[i] = projectItem{
			project:  p,
			tracking: tracking,
			active:   active,
		}
	}
	m.list.SetItems(items)
}

// SelectedProject returns the currently selected project, if any.
func (m ProjectsModel) SelectedProject() (api.Project, bool) {
	item := m.list.SelectedItem()
	if item == nil {
		return api.Project{}, false
	}
	pi, ok := item.(projectItem)
	if !ok {
		return api.Project{}, false
	}
	return pi.project, true
}

// Update handles messages for the projects list.
func (m ProjectsModel) Update(msg tea.Msg) (ProjectsModel, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the projects list.
func (m ProjectsModel) View() string {
	return m.list.View()
}
