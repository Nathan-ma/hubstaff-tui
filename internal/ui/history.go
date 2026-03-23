package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Nathan-ma/hubstaff-tui/internal/store"
)

// HistoryModel renders a scrollable multi-day session history grouped by date.
type HistoryModel struct {
	viewport viewport.Model
	rows     []store.HistorySummaryRow
	theme    Theme
	width    int
	height   int
	ready    bool
}

// NewHistoryModel creates an empty HistoryModel.
func NewHistoryModel(theme Theme) HistoryModel {
	return HistoryModel{
		theme: theme,
	}
}

// SetSize updates the viewport dimensions.
func (m *HistoryModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	if m.ready {
		m.viewport.Width = w
		m.viewport.Height = h
	}
}

// SetRows populates the history data and rebuilds the viewport content.
func (m *HistoryModel) SetRows(rows []store.HistorySummaryRow) {
	m.rows = rows
	if !m.ready {
		m.viewport = viewport.New(m.width, m.height)
		m.viewport.Style = lipgloss.NewStyle()
		m.ready = true
	}
	m.viewport.SetContent(m.renderHistory())
	m.viewport.GotoTop()
}

// Update handles viewport scrolling keys.
func (m HistoryModel) Update(msg tea.Msg) (HistoryModel, tea.Cmd) {
	if !m.ready {
		return m, nil
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the history viewport.
func (m HistoryModel) View() string {
	if !m.ready {
		return "Loading history..."
	}
	return m.viewport.View()
}

// renderHistory builds the grouped-by-date history string.
func (m HistoryModel) renderHistory() string {
	if len(m.rows) == 0 {
		return m.theme.DimmedItem.Render("No sessions in the last 7 days")
	}

	// Group rows by date preserving order (rows are already date DESC).
	type dateGroup struct {
		date  string
		rows  []store.HistorySummaryRow
		total int
	}

	var groups []dateGroup
	dateIndex := map[string]int{}

	for _, r := range m.rows {
		idx, ok := dateIndex[r.Date]
		if !ok {
			idx = len(groups)
			groups = append(groups, dateGroup{date: r.Date})
			dateIndex[r.Date] = idx
		}
		groups[idx].rows = append(groups[idx].rows, r)
		groups[idx].total += r.DurationSeconds
	}

	var sb strings.Builder

	// Title
	sb.WriteString(m.theme.HeaderTitle.Render("Session History (last 7 days)"))
	sb.WriteString("\n\n")

	for _, g := range groups {
		// Date header line: ─── 2026-03-23 (Mon) ─── Total: Xh YYm
		dayLabel := formatDateHeader(g.date)
		totalStr := formatCompactDuration(g.total)
		headerLine := fmt.Sprintf("  %s  Total: %s",
			m.theme.FooterKey.Render(dayLabel),
			m.theme.TimerText.Render(totalStr),
		)
		sb.WriteString(headerLine)
		sb.WriteString("\n")

		// Group tasks under their project name within the day.
		type projGroup struct {
			name  string
			tasks []store.HistorySummaryRow
		}
		var projGroups []projGroup
		projIndex := map[string]int{}

		for _, r := range g.rows {
			pname := r.ProjectName
			pidx, ok := projIndex[pname]
			if !ok {
				pidx = len(projGroups)
				projGroups = append(projGroups, projGroup{name: pname})
				projIndex[pname] = pidx
			}
			projGroups[pidx].tasks = append(projGroups[pidx].tasks, r)
		}

		for _, pg := range projGroups {
			sb.WriteString("  ")
			sb.WriteString(m.theme.NormalItem.Render(pg.name))
			sb.WriteString("\n")

			for _, t := range pg.tasks {
				task := t.TaskSummary
				dur := formatCompactDuration(t.DurationSeconds)
				bullet := "○"
				if t.DurationSeconds > 0 {
					bullet = "●"
				}
				// Pad task name and right-align duration in available width.
				// Available width: total - 4 (indent) - 2 (bullet+space) - len(dur) - 2 (spaces before dur).
				taskMaxW := m.width - 8 - len(dur)
				if taskMaxW < 10 {
					taskMaxW = 10
				}
				taskStr := truncate(task, taskMaxW)
				padding := taskMaxW - len([]rune(taskStr))
				if padding < 1 {
					padding = 1
				}
				line := fmt.Sprintf("    %s %s%s%s",
					m.theme.DimmedItem.Render(bullet),
					m.theme.NormalItem.Render(taskStr),
					strings.Repeat(" ", padding),
					m.theme.TimerText.Render(dur),
				)
				sb.WriteString(line)
				sb.WriteString("\n")
			}
		}
		sb.WriteString("\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

// formatDateHeader formats a "2006-01-02" date string as "2026-03-23 (Mon)".
func formatDateHeader(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return fmt.Sprintf("%s (%s)", dateStr, t.Format("Mon"))
}
