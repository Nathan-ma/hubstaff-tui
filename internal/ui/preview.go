package ui

import (
	"strings"

	"github.com/Nathan-ma/hubstaff-tui/internal/api"
)

// PreviewModel renders task details in the preview pane (US-026).
type PreviewModel struct {
	theme       Theme
	task        api.Task
	projectName string
	tracking    bool
	todaySecs   int
	todayLoaded bool
	width       int
	height      int
}

// NewPreviewModel creates a new PreviewModel.
func NewPreviewModel(theme Theme) PreviewModel {
	return PreviewModel{theme: theme}
}

// SetTask updates the preview with the selected task. Resets today's duration
// until a fresh taskPreviewMsg arrives.
func (m *PreviewModel) SetTask(task api.Task, projectName string, tracking bool) {
	m.task = task
	m.projectName = projectName
	m.tracking = tracking
	m.todaySecs = 0
	m.todayLoaded = false
}

// SetTodaySeconds records today's tracked duration for the current task.
func (m *PreviewModel) SetTodaySeconds(secs int) {
	m.todaySecs = secs
	m.todayLoaded = true
}

// SetSize updates the pane content dimensions.
func (m *PreviewModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// View renders the preview pane content.
func (m PreviewModel) View() string {
	if m.width == 0 {
		return ""
	}
	if string(m.task.ID) == "" {
		return "\n  " + m.theme.EmptyText.Render("Select a task to preview")
	}

	maxW := m.width - 2 // leave 2 chars of left indent
	if maxW < 1 {
		maxW = 1
	}

	var b strings.Builder

	// Task name
	b.WriteString("\n")
	name := truncate(m.task.Summary, maxW)
	b.WriteString("  " + m.theme.SelectedItem.Render(name) + "\n\n")

	// Project
	b.WriteString("  " + m.theme.DimmedItem.Render("Project") + "\n")
	b.WriteString("  " + truncate(m.projectName, maxW) + "\n\n")

	// Tracking status
	indicator := m.theme.TrackingIndicator(m.tracking, false)
	statusText := "Not tracking"
	if m.tracking {
		statusText = "Tracking now"
	}
	b.WriteString("  " + indicator + " " + statusText + "\n\n")

	// Today's time
	b.WriteString("  " + m.theme.DimmedItem.Render("Today") + "\n")
	if !m.todayLoaded {
		b.WriteString("  " + m.theme.EmptyText.Render("…") + "\n")
	} else if m.todaySecs == 0 {
		b.WriteString("  " + m.theme.EmptyText.Render("No time logged") + "\n")
	} else {
		b.WriteString("  " + m.theme.TimerText.Render(formatCompactDuration(m.todaySecs)) + "\n")
	}

	return b.String()
}
