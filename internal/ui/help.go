package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpModel is a modal overlay that displays all keybindings.
type HelpModel struct {
	viewport viewport.Model
	theme    Theme
	keys     KeyMap
	width    int
	height   int
}

// NewHelpModel creates a new HelpModel with default dimensions.
func NewHelpModel(theme Theme, keys KeyMap) HelpModel {
	vp := viewport.New(0, 0)
	return HelpModel{
		viewport: vp,
		theme:    theme,
		keys:     keys,
	}
}

// SetSize updates the help overlay dimensions based on terminal size.
// The overlay occupies at most 60 columns and 80% of the terminal height.
func (m *HelpModel) SetSize(termWidth, termHeight int) {
	m.width = termWidth
	m.height = termHeight

	boxWidth := 60
	if termWidth-4 < boxWidth {
		boxWidth = termWidth - 4
	}
	if boxWidth < 20 {
		boxWidth = 20
	}

	// Content width inside the border (border takes 2 cols)
	contentWidth := boxWidth - 2

	// Build the help text content
	content := m.helpContent(contentWidth)

	// Calculate how tall the viewport needs to be.
	// Reserve 2 lines for the border top/bottom.
	contentLines := strings.Count(content, "\n") + 1
	maxVPHeight := termHeight - 6 // leave room for border + margins
	if maxVPHeight < 5 {
		maxVPHeight = 5
	}
	vpHeight := contentLines
	if vpHeight > maxVPHeight {
		vpHeight = maxVPHeight
	}

	m.viewport.Width = contentWidth
	m.viewport.Height = vpHeight
	m.viewport.SetContent(content)
}

// Update handles input for the help viewport (scrolling).
func (m HelpModel) Update(msg tea.Msg) (HelpModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the help overlay as a bordered, centered box.
func (m HelpModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.HeaderTitle.GetForeground()).
		Padding(0, 1)

	title := titleStyle.Render("Keybindings")

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.ActiveBorder.GetBorderTopForeground()).
		Padding(1, 2)

	content := m.viewport.View()

	// Show scroll hint if content is scrollable
	if m.viewport.TotalLineCount() > m.viewport.Height {
		scrollHint := m.theme.FooterDesc.Render("  scroll: j/k  ")
		content += "\n" + scrollHint
	}

	box := boxStyle.Render(content)

	// Place the title on the top border
	boxLines := strings.Split(box, "\n")
	if len(boxLines) > 0 {
		// Insert title into the top border line
		borderLine := boxLines[0]
		titleStr := title
		titleWidth := lipgloss.Width(titleStr)
		borderWidth := lipgloss.Width(borderLine)
		if titleWidth+4 < borderWidth {
			// Replace part of the border with the title
			// The border line starts with the corner character
			runes := []rune(borderLine)
			titleRunes := []rune(titleStr)
			insertAt := 2 // after corner + one border char
			if insertAt+len(titleRunes) < len(runes) {
				for i, r := range titleRunes {
					runes[insertAt+i] = r
				}
				boxLines[0] = string(runes)
			}
		}
		box = strings.Join(boxLines, "\n")
	}

	return box
}

// helpContent builds the formatted help text.
func (m HelpModel) helpContent(width int) string {
	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.HeaderTitle.GetForeground())

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.FooterKey.GetForeground()).
		Width(14).
		Align(lipgloss.Left)

	descStyle := lipgloss.NewStyle().
		Foreground(m.theme.FooterDesc.GetForeground())

	_ = width

	var b strings.Builder

	sections := []struct {
		title string
		keys  []struct{ key, desc string }
	}{
		{
			title: "Navigation",
			keys: []struct{ key, desc string }{
				{"Enter", "Select project / Start task"},
				{m.keys.SwitchPane, "Switch pane (two-pane mode)"},
				{m.keys.Quit, "Back / Quit"},
				{"Ctrl+P", "Back to projects"},
				{"j / k", "Navigate list"},
				{m.keys.Filter, "Fuzzy filter"},
				{m.keys.Quit, "Clear filter"},
			},
		},
		{
			title: "Tracking",
			keys: []struct{ key, desc string }{
				{"Enter", "Start tracking selected task"},
				{m.keys.Stop, "Stop tracking"},
			},
		},
		{
			title: "Views",
			keys: []struct{ key, desc string }{
				{m.keys.GlobalSearch, "Global task search"},
				{m.keys.Summary, "Today's summary"},
				{m.keys.Help, "Toggle this help"},
				{m.keys.Refresh, "Refresh (clear cache)"},
			},
		},
		{
			title: "General",
			keys: []struct{ key, desc string }{
				{"Ctrl+C", "Force quit"},
			},
		},
	}

	for i, section := range sections {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(sectionStyle.Render(section.title))
		b.WriteString("\n")
		for _, kv := range section.keys {
			b.WriteString("  ")
			b.WriteString(keyStyle.Render(kv.key))
			b.WriteString(descStyle.Render(kv.desc))
			b.WriteString("\n")
		}
	}

	return strings.TrimRight(b.String(), "\n")
}
