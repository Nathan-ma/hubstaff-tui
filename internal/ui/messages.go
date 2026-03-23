package ui

import (
	"github.com/Nathan-ma/hubstaff-tui/internal/api"
	"github.com/Nathan-ma/hubstaff-tui/internal/config"
	"github.com/Nathan-ma/hubstaff-tui/internal/store"
)

// Messages used to communicate between async commands and the TUI.

type statusMsg struct{ status api.Status }
type statusErrMsg struct{ err error }

type projectsMsg struct{ projects []api.Project }
type projectsErrMsg struct{ err error }

type tasksMsg struct{ tasks []api.Task }
type tasksErrMsg struct{ err error }

type startedMsg struct {
	taskID    string
	projectID string
}
type startErrMsg struct{ err error }

type stoppedMsg struct{}
type stopErrMsg struct{ err error }

// switchedMsg is sent after a successful quick-switch (stop + start) operation.
type switchedMsg struct {
	taskID    string
	projectID string
}

type tickMsg struct{}

type clearStatusMsg struct{}

type summaryMsg struct{ rows []store.SummaryRow }
type summaryErrMsg struct{ err error }
type historyMsg struct{ rows []store.HistorySummaryRow }
type historyErrMsg struct{ err error }
type recentsMsg []store.RecentRow
type recentsErrMsg struct{ err error }

type pollTickMsg struct{}

// configCheckMsg triggers a check for config file changes.
type configCheckMsg struct{}

// configReloadedMsg carries a freshly loaded config after a file change.
type configReloadedMsg struct{ cfg config.Config }

// debounceMsg is sent after a debounce delay when the project selection changes
// in two-pane mode. If the projectID still matches the current selection,
// tasks are fetched for that project.
type debounceMsg struct {
	projectID   string
	projectName string
}

// taskPreviewMsg carries today's tracked seconds for the preview pane.
type taskPreviewMsg struct {
	taskID  string
	seconds int
}

// globalTasksMsg delivers tasks for a single project during global search.
type globalTasksMsg struct {
	projectID   string
	projectName string
	tasks       []api.Task
}

// globalTasksErrMsg is sent when a project's task fetch fails during global search.
type globalTasksErrMsg struct {
	projectID   string
	projectName string
	err         error
}

// globalSearchDoneMsg signals that all project task fetches are complete.
type globalSearchDoneMsg struct{}
