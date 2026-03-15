package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// headerView renders the header bar with tracking status and timer.
func (m AppModel) headerView() string {
	if m.width == 0 {
		return ""
	}

	var left string
	if m.tracking {
		left = m.theme.TrackingActive.Render("●") + " " +
			m.theme.HeaderTitle.Render(m.status.ActiveProject.Name+" / "+m.status.ActiveTask.Name)
	} else if m.status.ActiveProject.ID != "" {
		left = m.theme.TrackingLast.Render("○") + " " +
			m.theme.TimerStopped.Render("Not tracking (last: "+m.status.ActiveProject.Name+")")
	} else {
		left = m.theme.TrackingNone.Render("○") + " " +
			m.theme.TimerStopped.Render("Not tracking")
	}

	var right string
	if m.tracking {
		elapsed := m.timerBase + time.Since(m.timerStart)
		right = m.theme.TimerText.Render(fmt.Sprintf("⏱ %s", FormatTimer(elapsed)))
	} else if m.status.ActiveProject.TrackedToday != "" {
		right = m.theme.TimerStopped.Render(fmt.Sprintf("Today: %s", m.status.ActiveProject.TrackedToday))
	}

	// Calculate spacing to fill the width
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	spacing := m.width - leftWidth - rightWidth - 2 // 2 for padding
	if spacing < 1 {
		spacing = 1
	}

	content := left + strings.Repeat(" ", spacing) + right

	return m.theme.HeaderBar.Width(m.width).Render(content)
}
