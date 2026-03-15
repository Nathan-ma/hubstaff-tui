package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// footerView renders the footer bar with keybinding hints and status messages.
func (m AppModel) footerView() string {
	if m.width == 0 {
		return ""
	}

	var hints string
	switch m.current {
	case screenProjects:
		hints = m.keyHint("enter", "select") + "  " +
			m.keyHint("ctrl+e", "stop") + "  " +
			m.keyHint("ctrl+r", "refresh") + "  " +
			m.keyHint("/", "filter") + "  " +
			m.keyHint("T", "summary") + "  " +
			m.keyHint("esc", "quit")
	case screenTasks:
		hints = m.keyHint("enter", "start") + "  " +
			m.keyHint("ctrl+e", "stop") + "  " +
			m.keyHint("ctrl+r", "refresh") + "  " +
			m.keyHint("/", "filter") + "  " +
			m.keyHint("T", "summary") + "  " +
			m.keyHint("esc", "back")
	case screenSummary:
		hints = m.keyHint("j/k", "scroll") + "  " +
			m.keyHint("T", "back") + "  " +
			m.keyHint("esc", "back")
	}

	// If there's a status message, show it on the right
	var statusStr string
	if m.statusMsg != "" {
		if m.statusErr {
			statusStr = m.theme.ErrorText.Render(m.statusMsg)
		} else {
			statusStr = m.theme.SuccessText.Render(m.statusMsg)
		}
	}

	hintsWidth := lipgloss.Width(hints)
	statusWidth := lipgloss.Width(statusStr)
	spacing := m.width - hintsWidth - statusWidth - 2 // 2 for padding
	if spacing < 1 {
		spacing = 1
	}

	content := hints + strings.Repeat(" ", spacing) + statusStr

	return m.theme.FooterBar.Width(m.width).Render(content)
}

// keyHint renders a single keybinding hint.
func (m AppModel) keyHint(key, desc string) string {
	return m.theme.FooterKey.Render(key) + ":" + m.theme.FooterDesc.Render(desc)
}
