package ui

import (
	"github.com/Nathan-ma/hubstaff-tui/internal/api"
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

type tickMsg struct{}

type clearStatusMsg struct{}

type summaryMsg struct{ rows []store.SummaryRow }
type summaryErrMsg struct{ err error }
type recentsMsg []store.RecentRow
type recentsErrMsg struct{ err error }
