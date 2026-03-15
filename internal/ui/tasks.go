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

// taskItem wraps an api.Task for use in a bubbles/list.
type taskItem struct {
	task     api.Task
	tracking bool // currently being tracked
	active   bool // was the last tracked task
}

func (i taskItem) Title() string       { return i.task.Summary }
func (i taskItem) Description() string { return "" }
func (i taskItem) FilterValue() string { return i.task.Summary }

// TasksModel holds the state for the tasks list screen.
type TasksModel struct {
	list        list.Model
	tasks       []api.Task
	projectID   string
	projectName string
	status      api.Status
	theme       Theme
	loading     bool
	spinner     spinner.Model
}

// taskDelegate renders task items with tracking indicators.
type taskDelegate struct {
	theme Theme
}

func (d taskDelegate) Height() int  { return 1 }
func (d taskDelegate) Spacing() int { return 0 }
func (d taskDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d taskDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ti, ok := item.(taskItem)
	if !ok {
		return
	}

	indicator := d.theme.TrackingIndicator(ti.tracking, ti.active)

	name := ti.task.Summary
	isSelected := index == m.Index()

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

// NewTasksModel creates a new tasks model.
func NewTasksModel(theme Theme) TasksModel {
	delegate := taskDelegate{theme: theme}
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Tasks"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowTitle(true)
	l.SetStatusBarItemName("task", "tasks")
	l.Styles.Title = theme.HeaderTitle
	l.FilterInput.PromptStyle = theme.FilterPrompt
	l.FilterInput.TextStyle = theme.FilterText

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = theme.SpinnerStyle

	return TasksModel{
		list:    l,
		theme:   theme,
		spinner: s,
	}
}

// SetSize updates the list dimensions.
func (m *TasksModel) SetSize(width, height int) {
	m.list.SetSize(width, height)
}

// SetProject configures the tasks model for a given project.
func (m *TasksModel) SetProject(projectID, projectName string) {
	m.projectID = projectID
	m.projectName = projectName
	m.loading = true
	m.list.Title = fmt.Sprintf("Tasks - %s", projectName)
	m.list.SetItems([]list.Item{})
}

// SetTasks updates the task list items, marking tracking state.
func (m *TasksModel) SetTasks(tasks []api.Task, status api.Status) {
	m.tasks = tasks
	m.status = status
	m.loading = false

	items := make([]list.Item, len(tasks))
	for i, t := range tasks {
		tracking := status.Tracking && status.ActiveTask.ID == t.ID
		active := !status.Tracking && status.ActiveTask.ID == t.ID && status.ActiveTask.ID != ""
		items[i] = taskItem{
			task:     t,
			tracking: tracking,
			active:   active,
		}
	}
	m.list.SetItems(items)
}

// SelectedTask returns the currently selected task, if any.
func (m TasksModel) SelectedTask() (api.Task, bool) {
	item := m.list.SelectedItem()
	if item == nil {
		return api.Task{}, false
	}
	ti, ok := item.(taskItem)
	if !ok {
		return api.Task{}, false
	}
	return ti.task, true
}

// Update handles messages for the tasks list.
func (m TasksModel) Update(msg tea.Msg) (TasksModel, tea.Cmd) {
	var cmds []tea.Cmd

	if m.loading {
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

// View renders the tasks list (with spinner when loading).
func (m TasksModel) View() string {
	if m.loading {
		return fmt.Sprintf("\n  %s Loading tasks for %s...\n", m.spinner.View(), m.projectName)
	}
	return m.list.View()
}
