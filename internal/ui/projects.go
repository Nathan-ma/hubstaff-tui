package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
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
	loaded   bool
	loadErr  error
	spinner  spinner.Model
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

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = theme.SpinnerStyle

	return ProjectsModel{
		list:    l,
		theme:   theme,
		spinner: s,
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
	m.loaded = true
	m.loadErr = nil

	items := make([]list.Item, len(projects))
	for i, p := range projects {
		tracking := status.Tracking && status.ActiveProject.ID == p.ID
		active := !status.Tracking && status.ActiveProject.ID == p.ID && string(status.ActiveProject.ID) != ""
		items[i] = projectItem{
			project:  p,
			tracking: tracking,
			active:   active,
		}
	}
	m.list.SetItems(items)
}

// SetError records a load error for the projects model.
func (m *ProjectsModel) SetError(err error) {
	m.loadErr = err
	m.loaded = true
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
	var cmds []tea.Cmd

	// Update spinner when not yet loaded
	if !m.loaded {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the projects list.
func (m ProjectsModel) View() string {
	// Loading state: show spinner
	if !m.loaded {
		return fmt.Sprintf("\n  %s Loading projects...\n", m.spinner.View())
	}

	// Error state: show error message with retry hint
	if m.loadErr != nil {
		errMsg := m.theme.ErrorText.Render(fmt.Sprintf("Failed to load projects: %v", m.loadErr))
		hint := m.theme.EmptyText.Render("Press ctrl+r to retry")
		return fmt.Sprintf("\n\n  %s\n\n  %s\n", errMsg, hint)
	}

	// Empty state: no projects found
	if len(m.projects) == 0 {
		emptyMsg := m.theme.EmptyText.Render("No projects found")
		hint := m.theme.EmptyText.Render("Press ctrl+r to refresh")
		return fmt.Sprintf("\n\n  %s\n\n  %s\n", emptyMsg, hint)
	}

	return m.list.View()
}
