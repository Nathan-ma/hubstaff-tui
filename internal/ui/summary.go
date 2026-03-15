package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Nathan-ma/hubstaff-tui/internal/store"
)

// SummaryModel renders a scrollable table of today's tracked time by project/task.
type SummaryModel struct {
	viewport viewport.Model
	rows     []store.SummaryRow
	theme    Theme
	width    int
	height   int
	ready    bool
}

// NewSummaryModel creates an empty SummaryModel.
func NewSummaryModel(theme Theme) SummaryModel {
	return SummaryModel{
		theme: theme,
	}
}

// SetSize updates the viewport dimensions.
func (m *SummaryModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	if m.ready {
		m.viewport.Width = w
		m.viewport.Height = h
	}
}

// SetRows populates the summary data and rebuilds the viewport content.
func (m *SummaryModel) SetRows(rows []store.SummaryRow) {
	m.rows = rows
	if !m.ready {
		m.viewport = viewport.New(m.width, m.height)
		m.viewport.Style = lipgloss.NewStyle()
		m.ready = true
	}
	m.viewport.SetContent(m.renderTable())
	m.viewport.GotoTop()
}

// Update handles viewport scrolling keys.
func (m SummaryModel) Update(msg tea.Msg) (SummaryModel, tea.Cmd) {
	if !m.ready {
		return m, nil
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the summary viewport.
func (m SummaryModel) View() string {
	if !m.ready {
		return "Loading summary..."
	}
	return m.viewport.View()
}

// renderTable builds the table string from the summary rows.
func (m SummaryModel) renderTable() string {
	if len(m.rows) == 0 {
		return m.theme.DimmedItem.Render("No time tracked today")
	}

	// Determine column widths
	projHeader := "Project"
	taskHeader := "Task"
	durHeader := "Duration"

	maxProj := len(projHeader)
	maxTask := len(taskHeader)
	maxDur := len(durHeader)

	type rendered struct {
		proj string
		task string
		dur  string
	}
	var lines []rendered
	var totalSeconds int

	for _, r := range m.rows {
		proj := r.ProjectName
		if proj == "" {
			proj = r.ProjectID
		}
		task := r.TaskSummary
		if task == "" {
			task = r.TaskID
		}
		dur := formatCompactDuration(r.DurationSeconds)
		totalSeconds += r.DurationSeconds

		if len(proj) > maxProj {
			maxProj = len(proj)
		}
		if len(task) > maxTask {
			maxTask = len(task)
		}
		if len(dur) > maxDur {
			maxDur = len(dur)
		}
		lines = append(lines, rendered{proj: proj, task: task, dur: dur})
	}

	totalDur := formatCompactDuration(totalSeconds)
	if len(totalDur) > maxDur {
		maxDur = len(totalDur)
	}

	// Cap column widths to fit terminal, leaving room for separators
	availWidth := m.width - 6 // 3 columns, 2 separators of " | " (each 3 chars)
	if availWidth < 20 {
		availWidth = 20
	}
	// Allocate: duration gets its width, rest split between project and task
	remaining := availWidth - maxDur
	if remaining < 10 {
		remaining = 10
	}
	if maxProj+maxTask > remaining {
		half := remaining / 2
		if maxProj > half && maxTask > half {
			maxProj = half
			maxTask = remaining - half
		} else if maxProj > half {
			maxProj = remaining - maxTask
		} else {
			maxTask = remaining - maxProj
		}
	}

	var sb strings.Builder

	// Title
	title := m.theme.HeaderTitle.Render("Today's Summary")
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// Header row
	headerStyle := m.theme.FooterKey
	sep := m.theme.Separator.Render(" | ")

	hdr := headerStyle.Render(padRight(projHeader, maxProj)) + sep +
		headerStyle.Render(padRight(taskHeader, maxTask)) + sep +
		headerStyle.Render(padLeft(durHeader, maxDur))
	sb.WriteString(hdr)
	sb.WriteString("\n")

	// Separator line
	sb.WriteString(m.theme.Separator.Render(strings.Repeat("─", maxProj)))
	sb.WriteString(m.theme.Separator.Render("─┼─"))
	sb.WriteString(m.theme.Separator.Render(strings.Repeat("─", maxTask)))
	sb.WriteString(m.theme.Separator.Render("─┼─"))
	sb.WriteString(m.theme.Separator.Render(strings.Repeat("─", maxDur)))
	sb.WriteString("\n")

	// Data rows
	for _, l := range lines {
		row := m.theme.NormalItem.Render(padRight(truncate(l.proj, maxProj), maxProj)) + sep +
			m.theme.NormalItem.Render(padRight(truncate(l.task, maxTask), maxTask)) + sep +
			m.theme.TimerText.Render(padLeft(l.dur, maxDur))
		sb.WriteString(row)
		sb.WriteString("\n")
	}

	// Totals separator
	sb.WriteString(m.theme.Separator.Render(strings.Repeat("─", maxProj)))
	sb.WriteString(m.theme.Separator.Render("─┼─"))
	sb.WriteString(m.theme.Separator.Render(strings.Repeat("─", maxTask)))
	sb.WriteString(m.theme.Separator.Render("─┼─"))
	sb.WriteString(m.theme.Separator.Render(strings.Repeat("─", maxDur)))
	sb.WriteString("\n")

	// Total row
	totalLabel := "Total"
	totalRow := headerStyle.Render(padRight(totalLabel, maxProj)) + sep +
		m.theme.NormalItem.Render(padRight("", maxTask)) + sep +
		m.theme.TimerText.Render(padLeft(totalDur, maxDur))
	sb.WriteString(totalRow)
	sb.WriteString("\n")

	return sb.String()
}

// formatCompactDuration formats seconds as "Xh YYm" for compact display.
func formatCompactDuration(seconds int) string {
	if seconds < 0 {
		seconds = 0
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	if h > 0 {
		return fmt.Sprintf("%dh %02dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

// padRight pads a string with spaces on the right to reach width.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

// padLeft pads a string with spaces on the left to reach width.
func padLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}

// truncate shortens a string to maxLen, appending "…" if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return s[:maxLen]
	}
	return s[:maxLen-1] + "…"
}
