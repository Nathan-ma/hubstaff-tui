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

// globalTaskItem wraps a task with its project info for display in global search.
type globalTaskItem struct {
	task        api.Task
	projectID   string
	projectName string
	tracking    bool
}

func (i globalTaskItem) Title() string       { return fmt.Sprintf("%s > %s", i.projectName, i.task.Summary) }
func (i globalTaskItem) Description() string { return "" }
func (i globalTaskItem) FilterValue() string { return i.projectName + " " + i.task.Summary }

// globalTaskDelegate renders global task items with project name prefix.
type globalTaskDelegate struct {
	theme Theme
}

func (d globalTaskDelegate) Height() int                             { return 1 }
func (d globalTaskDelegate) Spacing() int                            { return 0 }
func (d globalTaskDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d globalTaskDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	gi, ok := item.(globalTaskItem)
	if !ok {
		return
	}

	indicator := d.theme.TrackingIndicator(gi.tracking, false)

	title := fmt.Sprintf("%s > %s", gi.projectName, gi.task.Summary)
	isSelected := index == m.Index()

	maxWidth := m.Width() - 4
	if maxWidth > 0 {
		title = ansi.Truncate(title, maxWidth, "...")
	}

	var line string
	if isSelected {
		line = d.theme.SelectedItem.Render(fmt.Sprintf("%s %s", indicator, title))
	} else {
		line = d.theme.NormalItem.Render(fmt.Sprintf("%s %s", indicator, title))
	}

	_, _ = fmt.Fprint(w, line)
}

// SearchModel holds the state for the global task search screen.
type SearchModel struct {
	list    list.Model
	tasks   []globalTaskItem
	theme   Theme
	loading bool
	spinner spinner.Model
	loaded  int // number of projects loaded so far
	total   int // total projects to load
}

// NewSearchModel creates a new search model.
func NewSearchModel(theme Theme) SearchModel {
	delegate := globalTaskDelegate{theme: theme}
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Global Task Search"
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

	return SearchModel{
		list:    l,
		theme:   theme,
		spinner: s,
	}
}

// SetSize updates the list dimensions.
func (m *SearchModel) SetSize(width, height int) {
	m.list.SetSize(width, height)
}

// Activate resets the search model for a new search session.
func (m *SearchModel) Activate(totalProjects int) {
	m.tasks = nil
	m.loading = true
	m.loaded = 0
	m.total = totalProjects
	m.list.SetItems([]list.Item{})
	m.list.Title = "Global Task Search"
}

// AddTasks appends tasks from a project to the search results incrementally.
func (m *SearchModel) AddTasks(projectID, projectName string, tasks []api.Task, status api.Status) {
	for _, t := range tasks {
		tracking := status.Tracking && status.ActiveTask.ID == t.ID
		item := globalTaskItem{
			task:        t,
			projectID:   projectID,
			projectName: projectName,
			tracking:    tracking,
		}
		m.tasks = append(m.tasks, item)
	}
	m.loaded++
	if m.loaded >= m.total {
		m.loading = false
	}
	m.rebuildList()
}

// MarkDone marks loading as complete (in case there are zero projects).
func (m *SearchModel) MarkDone() {
	m.loading = false
}

// rebuildList updates the list items from the current tasks slice.
func (m *SearchModel) rebuildList() {
	items := make([]list.Item, len(m.tasks))
	for i, t := range m.tasks {
		items[i] = t
	}
	m.list.SetItems(items)
}

// SelectedTask returns the currently selected global task item, if any.
func (m SearchModel) SelectedTask() (globalTaskItem, bool) {
	item := m.list.SelectedItem()
	if item == nil {
		return globalTaskItem{}, false
	}
	gi, ok := item.(globalTaskItem)
	if !ok {
		return globalTaskItem{}, false
	}
	return gi, true
}

// Update handles messages for the search list.
func (m SearchModel) Update(msg tea.Msg) (SearchModel, tea.Cmd) {
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

// View renders the search list with loading indicator.
func (m SearchModel) View() string {
	if m.loading && len(m.tasks) == 0 {
		return fmt.Sprintf("\n  %s Loading tasks... (%d/%d projects)\n", m.spinner.View(), m.loaded, m.total)
	}

	if m.loading {
		// Show the list with a loading indicator in the title
		m.list.Title = fmt.Sprintf("Global Task Search (%d/%d projects)", m.loaded, m.total)
	} else {
		m.list.Title = "Global Task Search"
	}

	return m.list.View()
}
