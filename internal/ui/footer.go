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
			m.keyHint(m.keys.Stop, "stop") + "  " +
			m.keyHint(m.keys.Refresh, "refresh") + "  " +
			m.keyHint(m.keys.Filter, "filter") + "  " +
			m.keyHint(m.keys.GlobalSearch, "search") + "  " +
			m.keyHint(m.keys.Summary, "summary") + "  " +
			m.keyHint(m.keys.Help, "help") + "  " +
			m.keyHint(m.keys.Quit, "quit")
	case screenTasks:
		hints = m.keyHint("enter", "start/switch") + "  " +
			m.keyHint(m.keys.Stop, "stop") + "  " +
			m.keyHint(m.keys.Refresh, "refresh") + "  " +
			m.keyHint(m.keys.Filter, "filter") + "  " +
			m.keyHint(m.keys.GlobalSearch, "search") + "  " +
			m.keyHint(m.keys.Summary, "summary") + "  " +
			m.keyHint(m.keys.Help, "help") + "  " +
			m.keyHint(m.keys.Quit, "back")
	case screenGlobalSearch:
		hints = m.keyHint("enter", "start") + "  " +
			m.keyHint(m.keys.Stop, "stop") + "  " +
			m.keyHint(m.keys.Filter, "filter") + "  " +
			m.keyHint(m.keys.Help, "help") + "  " +
			m.keyHint(m.keys.Quit, "back")
	case screenSummary:
		hints = m.keyHint("j/k", "scroll") + "  " +
			m.keyHint(m.keys.Summary, "back") + "  " +
			m.keyHint(m.keys.Quit, "back")
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
