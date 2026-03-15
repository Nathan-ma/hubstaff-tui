package ui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Nathan-ma/hubstaff-tui/internal/api"
	"github.com/Nathan-ma/hubstaff-tui/internal/config"
	"github.com/Nathan-ma/hubstaff-tui/internal/store"
)

type screen int

const (
	screenProjects screen = iota
	screenTasks
	screenSummary
)

// AppModel is the root Bubbletea model for the TUI.
type AppModel struct {
	cfg    config.Config
	client *api.Client
	store  *store.Store
	theme  Theme

	// Navigation
	current screen

	// Sub-models
	projects ProjectsModel
	tasks    TasksModel
	summary  SummaryModel

	// Global state
	status api.Status
	width  int
	height int

	// Timer
	timerBase  time.Duration // from status.tracked_today at load time
	timerStart time.Time     // when we started ticking
	tracking   bool

	// Help overlay
	showHelp bool
	help     HelpModel

	// Error/status messages
	statusMsg string
	statusErr bool
}

// NewApp creates a new AppModel ready for tea.NewProgram.
func NewApp(cfg config.Config, client *api.Client, st *store.Store) AppModel {
	theme := GetTheme(cfg.UI.Theme)
	return AppModel{
		cfg:      cfg,
		client:   client,
		store:    st,
		theme:    theme,
		current:  screenProjects,
		projects: NewProjectsModel(theme),
		tasks:    NewTasksModel(theme),
		help:     NewHelpModel(theme),
		summary:  NewSummaryModel(theme),
	}
}

// Init fetches the initial status and projects concurrently.
func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchStatus(),
		m.fetchProjects(),
		m.projects.spinner.Tick,
	)
}

// Update handles all messages.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := 1
		footerHeight := 1
		contentHeight := m.height - headerHeight - footerHeight
		if contentHeight < 1 {
			contentHeight = 1
		}
		m.projects.SetSize(m.width, contentHeight)
		m.tasks.SetSize(m.width, contentHeight)
		if m.showHelp {
			m.help.SetSize(m.width, m.height)
		}
		m.summary.SetSize(m.width, contentHeight)
		return m, nil

	case tea.KeyMsg:
		// When help overlay is visible, only handle dismiss keys and scrolling
		if m.showHelp {
			switch msg.String() {
			case "?", "esc":
				m.showHelp = false
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			default:
				// Route to viewport for scrolling (j/k/up/down/pgup/pgdown)
				var cmd tea.Cmd
				m.help, cmd = m.help.Update(msg)
				return m, cmd
			}
		}

		// Handle global keys first, but only when not filtering
		if !m.isFiltering() {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "?":
				m.showHelp = true
				m.help.SetSize(m.width, m.height)
				return m, nil
			case "ctrl+e":
				return m, m.stopTracking()
			case "ctrl+r":
				m.statusMsg = "Refreshing..."
				m.statusErr = false
				// Reset loaded state so spinners show again
				m.projects.loaded = false
				m.projects.loadErr = nil
				cmds = append(cmds, m.fetchStatus(), m.fetchProjects(), m.projects.spinner.Tick)
				if m.current == screenTasks {
					m.tasks.loading = true
					m.tasks.loaded = false
					m.tasks.loadErr = nil
					cmds = append(cmds, m.fetchTasks(m.tasks.projectID), m.tasks.spinner.Tick)
				}
				return m, tea.Batch(cmds...)
			case "T":
				if m.current != screenSummary {
					m.current = screenSummary
					m.summary.SetSize(m.width, m.height-2) // header + footer
					return m, m.fetchSummary()
				}
			}
		}

		// Screen-specific key handling (only when not filtering)
		if !m.isFiltering() {
			switch m.current {
			case screenProjects:
				switch msg.String() {
				case "enter":
					if p, ok := m.projects.SelectedProject(); ok {
						m.current = screenTasks
						m.tasks.SetProject(p.ID, p.Name)
						return m, tea.Batch(m.fetchTasks(p.ID), m.fetchRecents(), m.tasks.spinner.Tick)
					}
				case "esc":
					return m, tea.Quit
				}
			case screenTasks:
				switch msg.String() {
				case "enter":
					if t, ok := m.tasks.SelectedTask(); ok {
						return m, m.startTask(t.ID, m.tasks.projectID)
					}
				case "esc":
					m.current = screenProjects
					return m, nil
				}
			case screenSummary:
				switch msg.String() {
				case "esc", "T":
					m.current = screenProjects
					return m, nil
				}
			}
		}

	// --- Async message handlers ---

	case statusMsg:
		m.status = msg.status
		m.tracking = msg.status.Tracking
		if m.tracking {
			m.timerBase = parseTrackedToday(msg.status.ActiveProject.TrackedToday)
			m.timerStart = time.Now()
			cmds = append(cmds, tickCmd())
		}
		// Refresh project list tracking indicators
		m.projects.SetProjects(m.projects.projects, m.status)
		if m.current == screenTasks {
			m.tasks.SetTasks(m.tasks.tasks, m.status)
		}
		return m, tea.Batch(cmds...)

	case statusErrMsg:
		m.statusMsg = fmt.Sprintf("Status error: %v", msg.err)
		m.statusErr = true
		cmds = append(cmds, m.clearStatusAfter())
		return m, tea.Batch(cmds...)

	case projectsMsg:
		m.projects.SetProjects(msg.projects, m.status)
		return m, nil

	case projectsErrMsg:
		m.projects.SetError(msg.err)
		m.statusMsg = fmt.Sprintf("Projects error: %v", msg.err)
		m.statusErr = true
		cmds = append(cmds, m.clearStatusAfter())
		return m, tea.Batch(cmds...)

	case tasksMsg:
		m.tasks.SetTasks(msg.tasks, m.status)
		return m, nil

	case recentsMsg:
		m.tasks.SetRecents([]store.RecentRow(msg))
		return m, nil

	case recentsErrMsg:
		// Non-critical: just log to status briefly.
		m.statusMsg = fmt.Sprintf("Recents error: %v", msg.err)
		m.statusErr = true
		cmds = append(cmds, m.clearStatusAfter())
		return m, tea.Batch(cmds...)

	case tasksErrMsg:
		m.tasks.SetError(msg.err)
		m.statusMsg = fmt.Sprintf("Tasks error: %v", msg.err)
		m.statusErr = true
		cmds = append(cmds, m.clearStatusAfter())
		return m, tea.Batch(cmds...)

	case startedMsg:
		m.statusMsg = "Tracking started"
		m.statusErr = false
		if m.store != nil {
			_ = m.store.TouchRecent(msg.taskID, msg.projectID)
		}
		cmds = append(cmds, m.fetchStatus(), m.clearStatusAfter())
		return m, tea.Batch(cmds...)

	case startErrMsg:
		m.statusMsg = fmt.Sprintf("Start error: %v", msg.err)
		m.statusErr = true
		cmds = append(cmds, m.clearStatusAfter())
		return m, tea.Batch(cmds...)

	case stoppedMsg:
		m.tracking = false
		m.statusMsg = "Tracking stopped"
		m.statusErr = false
		cmds = append(cmds, m.fetchStatus(), m.clearStatusAfter())
		return m, tea.Batch(cmds...)

	case stopErrMsg:
		m.statusMsg = fmt.Sprintf("Stop error: %v", msg.err)
		m.statusErr = true
		cmds = append(cmds, m.clearStatusAfter())
		return m, tea.Batch(cmds...)

	case tickMsg:
		if m.tracking {
			cmds = append(cmds, tickCmd())
		}
		return m, tea.Batch(cmds...)

	case summaryMsg:
		m.summary.SetRows(msg.rows)
		return m, nil

	case summaryErrMsg:
		m.statusMsg = fmt.Sprintf("Summary error: %v", msg.err)
		m.statusErr = true
		m.current = screenProjects
		cmds = append(cmds, m.clearStatusAfter())
		return m, tea.Batch(cmds...)

	case clearStatusMsg:
		m.statusMsg = ""
		m.statusErr = false
		return m, nil
	}

	// Route to active sub-model
	switch m.current {
	case screenProjects:
		var cmd tea.Cmd
		m.projects, cmd = m.projects.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case screenTasks:
		var cmd tea.Cmd
		m.tasks, cmd = m.tasks.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case screenSummary:
		var cmd tea.Cmd
		m.summary, cmd = m.summary.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the full TUI.
func (m AppModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.headerView()
	footer := m.footerView()

	var content string
	switch m.current {
	case screenProjects:
		content = m.projects.View()
	case screenTasks:
		content = m.tasks.View()
	case screenSummary:
		content = m.summary.View()
	}

	view := header + "\n" + content + "\n" + footer

	if m.showHelp {
		helpBox := m.help.View()
		overlay := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpBox)
		return overlay
	}

	return view
}

// --- Commands ---

func (m AppModel) fetchStatus() tea.Cmd {
	client := m.client
	return func() tea.Msg {
		s, err := client.GetStatus(context.Background())
		if err != nil {
			return statusErrMsg{err: err}
		}
		return statusMsg{status: s}
	}
}

func (m AppModel) fetchProjects() tea.Cmd {
	client := m.client
	return func() tea.Msg {
		projects, err := client.ListProjects(context.Background())
		if err != nil {
			return projectsErrMsg{err: err}
		}
		return projectsMsg{projects: projects}
	}
}

func (m AppModel) fetchTasks(projectID string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		tasks, err := client.ListTasks(context.Background(), projectID)
		if err != nil {
			return tasksErrMsg{err: err}
		}
		return tasksMsg{tasks: tasks}
	}
}

func (m AppModel) fetchSummary() tea.Cmd {
	st := m.store
	return func() tea.Msg {
		if st == nil {
			return summaryErrMsg{err: fmt.Errorf("store not configured")}
		}
		rows, err := st.TodaySummary()
		if err != nil {
			return summaryErrMsg{err: err}
		}
		return summaryMsg{rows: rows}
	}
}

func (m AppModel) startTask(taskID, projectID string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		err := client.StartTask(context.Background(), taskID)
		if err != nil {
			return startErrMsg{err: err}
		}
		return startedMsg{taskID: taskID, projectID: projectID}
	}
}

func (m AppModel) fetchRecents() tea.Cmd {
	s := m.store
	limit := m.cfg.RecentTasks.MaxItems
	return func() tea.Msg {
		if s == nil {
			return recentsMsg(nil)
		}
		recents, err := s.ListRecents(limit)
		if err != nil {
			return recentsErrMsg{err: err}
		}
		return recentsMsg(recents)
	}
}

func (m AppModel) stopTracking() tea.Cmd {
	client := m.client
	return func() tea.Msg {
		err := client.Stop(context.Background())
		if err != nil {
			return stopErrMsg{err: err}
		}
		return stoppedMsg{}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(_ time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m AppModel) clearStatusAfter() tea.Cmd {
	return tea.Tick(3*time.Second, func(_ time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

// --- Helpers ---

// isFiltering returns true if the active list is in filtering mode.
func (m AppModel) isFiltering() bool {
	switch m.current {
	case screenProjects:
		return m.projects.list.FilterState() == list.Filtering
	case screenTasks:
		return m.tasks.list.FilterState() == list.Filtering
	case screenSummary:
		return false
	}
	return false
}

// parseTrackedToday parses a "H:MM:SS" string into a time.Duration.
func parseTrackedToday(s string) time.Duration {
	if s == "" {
		return 0
	}
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0
	}
	h, _ := strconv.Atoi(parts[0])
	min, _ := strconv.Atoi(parts[1])
	sec, _ := strconv.Atoi(parts[2])
	return time.Duration(h)*time.Hour + time.Duration(min)*time.Minute + time.Duration(sec)*time.Second
}
