package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/Nathan-ma/hubstaff-tui/internal/api"
	"github.com/Nathan-ma/hubstaff-tui/internal/store"
)

// taskItem wraps an api.Task for use in a bubbles/list.
type taskItem struct {
	task     api.Task
	tracking bool // currently being tracked
	active   bool // was the last tracked task
	recent   bool // recently used
}

func (i taskItem) Title() string       { return i.task.Summary }
func (i taskItem) Description() string { return "" }
func (i taskItem) FilterValue() string { return i.task.Summary }

// separatorItem is a non-selectable visual separator in the task list.
type separatorItem struct {
	label string
}

func (s separatorItem) Title() string       { return s.label }
func (s separatorItem) Description() string { return "" }
func (s separatorItem) FilterValue() string { return "" }

// TasksModel holds the state for the tasks list screen.
type TasksModel struct {
	list        list.Model
	tasks       []api.Task
	recents     []store.RecentRow
	projectID   string
	projectName string
	status      api.Status
	theme       Theme
	loading     bool
	loaded      bool
	loadErr     error
	spinner     spinner.Model
}

// taskDelegate renders task items with tracking indicators.
type taskDelegate struct {
	theme Theme
}

func (d taskDelegate) Height() int                             { return 1 }
func (d taskDelegate) Spacing() int                            { return 0 }
func (d taskDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d taskDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	// Handle separator items
	if sep, ok := item.(separatorItem); ok {
		line := d.theme.Separator.Render(sep.label)
		fmt.Fprint(w, line)
		return
	}

	ti, ok := item.(taskItem)
	if !ok {
		return
	}

	indicator := d.theme.TrackingIndicator(ti.tracking, ti.active)

	name := ti.task.Summary
	isSelected := index == m.Index()

	recentLabel := ""
	if ti.recent {
		recentLabel = " " + d.theme.DimmedItem.Render("recent")
	}

	maxWidth := m.Width() - 4
	if ti.recent && maxWidth > 0 {
		// Leave room for the "recent" label (about 8 chars with styling)
		maxWidth -= 8
	}
	if maxWidth > 0 {
		name = ansi.Truncate(name, maxWidth, "...")
	}

	var line string
	if isSelected {
		line = d.theme.SelectedItem.Render(fmt.Sprintf("%s %s", indicator, name)) + recentLabel
	} else {
		line = d.theme.NormalItem.Render(fmt.Sprintf("%s %s", indicator, name)) + recentLabel
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
	m.loaded = false
	m.loadErr = nil
	m.list.Title = fmt.Sprintf("Tasks - %s", projectName)
	m.list.SetItems([]list.Item{})
}

// SetTasks updates the task list items, marking tracking state.
// It rebuilds the list incorporating any previously loaded recents.
func (m *TasksModel) SetTasks(tasks []api.Task, status api.Status) {
	m.tasks = tasks
	m.status = status
	m.loading = false
	m.loaded = true
	m.loadErr = nil
	m.rebuildList()
}

// SetRecents updates the recents list and rebuilds the task list display.
func (m *TasksModel) SetRecents(recents []store.RecentRow) {
	m.recents = recents
	// Only rebuild if tasks are already loaded (not still loading).
	if !m.loading && len(m.tasks) > 0 {
		m.rebuildList()
	}
}

// rebuildList constructs the list items with recents shown first under a separator,
// followed by the remaining tasks.
func (m *TasksModel) rebuildList() {
	// Build a set of recent task IDs for this project.
	recentIDs := make(map[string]bool)
	for _, r := range m.recents {
		if r.ProjectID == m.projectID {
			recentIDs[r.TaskID] = true
		}
	}

	// Separate tasks into recent and non-recent groups.
	var recentTasks, otherTasks []api.Task
	for _, t := range m.tasks {
		if recentIDs[t.ID] {
			recentTasks = append(recentTasks, t)
		} else {
			otherTasks = append(otherTasks, t)
		}
	}

	// Order recent tasks by their recency (order from m.recents).
	if len(recentTasks) > 1 {
		orderMap := make(map[string]int)
		for i, r := range m.recents {
			orderMap[r.TaskID] = i
		}
		// Simple insertion sort for small N.
		for i := 1; i < len(recentTasks); i++ {
			for j := i; j > 0 && orderMap[recentTasks[j].ID] < orderMap[recentTasks[j-1].ID]; j-- {
				recentTasks[j], recentTasks[j-1] = recentTasks[j-1], recentTasks[j]
			}
		}
	}

	var items []list.Item

	if len(recentTasks) > 0 {
		// Add "Recent" separator
		items = append(items, separatorItem{label: "── Recent ──"})

		for _, t := range recentTasks {
			tracking := m.status.Tracking && m.status.ActiveTask.ID == t.ID
			active := !m.status.Tracking && m.status.ActiveTask.ID == t.ID && m.status.ActiveTask.ID != ""
			items = append(items, taskItem{
				task:     t,
				tracking: tracking,
				active:   active,
				recent:   true,
			})
		}

		// Add "All Tasks" separator if there are other tasks
		if len(otherTasks) > 0 {
			items = append(items, separatorItem{label: "── All Tasks ──"})
		}
	}

	for _, t := range otherTasks {
		tracking := m.status.Tracking && m.status.ActiveTask.ID == t.ID
		active := !m.status.Tracking && m.status.ActiveTask.ID == t.ID && m.status.ActiveTask.ID != ""
		items = append(items, taskItem{
			task:     t,
			tracking: tracking,
			active:   active,
		})
	}

	m.list.SetItems(items)
}

// SetError records a load error for the tasks model.
func (m *TasksModel) SetError(err error) {
	m.loadErr = err
	m.loading = false
	m.loaded = true
}

// SelectedTask returns the currently selected task, if any.
// Returns false if the selection is on a separator item.
func (m TasksModel) SelectedTask() (api.Task, bool) {
	item := m.list.SelectedItem()
	if item == nil {
		return api.Task{}, false
	}
	ti, ok := item.(taskItem)
	if !ok {
		// Selected item is a separator, not a task.
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
	// Loading state: show spinner
	if m.loading {
		return fmt.Sprintf("\n  %s Loading tasks for %s...\n", m.spinner.View(), m.projectName)
	}

	// Error state: show error message with retry hint
	if m.loadErr != nil {
		errMsg := m.theme.ErrorText.Render(fmt.Sprintf("Failed to load tasks: %v", m.loadErr))
		hint := m.theme.EmptyText.Render("Press ctrl+r to retry")
		return fmt.Sprintf("\n\n  %s\n\n  %s\n", errMsg, hint)
	}

	// Empty state: no tasks found for this project
	if m.loaded && len(m.tasks) == 0 {
		emptyMsg := m.theme.EmptyText.Render("No tasks found for this project")
		hint := m.theme.EmptyText.Render("Press ctrl+r to refresh or esc to go back")
		return fmt.Sprintf("\n\n  %s\n\n  %s\n", emptyMsg, hint)
	}

	return m.list.View()
}
