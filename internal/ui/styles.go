package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Catppuccin Mocha palette
// Reference: https://github.com/catppuccin/catppuccin
const (
	CatRosewater = lipgloss.Color("#f5e0dc")
	CatFlamingo  = lipgloss.Color("#f2cdcd")
	CatPink      = lipgloss.Color("#f5c2e7")
	CatMauve     = lipgloss.Color("#cba6f7")
	CatRed       = lipgloss.Color("#f38ba8")
	CatMaroon    = lipgloss.Color("#eba0ac")
	CatPeach     = lipgloss.Color("#fab387")
	CatYellow    = lipgloss.Color("#f9e2af")
	CatGreen     = lipgloss.Color("#a6e3a1")
	CatTeal      = lipgloss.Color("#94e2d5")
	CatSky       = lipgloss.Color("#89dceb")
	CatSapphire  = lipgloss.Color("#74c7ec")
	CatBlue      = lipgloss.Color("#89b4fa")
	CatLavender  = lipgloss.Color("#b4befe")
	CatText      = lipgloss.Color("#cdd6f4")
	CatSubtext1  = lipgloss.Color("#bac2de")
	CatSubtext0  = lipgloss.Color("#a6adc8")
	CatOverlay2  = lipgloss.Color("#9399b2")
	CatOverlay1  = lipgloss.Color("#7f849c")
	CatOverlay0  = lipgloss.Color("#6c7086")
	CatSurface2  = lipgloss.Color("#585b70")
	CatSurface1  = lipgloss.Color("#45475a")
	CatSurface0  = lipgloss.Color("#313244")
	CatBase      = lipgloss.Color("#1e1e2e")
	CatMantle    = lipgloss.Color("#181825")
	CatCrust     = lipgloss.Color("#11111b")
)

// Theme defines all the visual styles used by the TUI.
type Theme struct {
	// Header
	HeaderBar    lipgloss.Style
	HeaderTitle  lipgloss.Style
	TimerText    lipgloss.Style
	TimerStopped lipgloss.Style

	// Footer
	FooterBar  lipgloss.Style
	FooterKey  lipgloss.Style
	FooterDesc lipgloss.Style

	// List items
	SelectedItem lipgloss.Style
	NormalItem   lipgloss.Style
	DimmedItem   lipgloss.Style

	// Tracking indicators
	TrackingActive lipgloss.Style // ● green
	TrackingLast   lipgloss.Style // ○ yellow
	TrackingNone   lipgloss.Style // dim

	// Filter input
	FilterPrompt lipgloss.Style
	FilterText   lipgloss.Style
	FilterMatch  lipgloss.Style // highlighted matched chars

	// Status messages
	ErrorText    lipgloss.Style
	SuccessText  lipgloss.Style
	EmptyText    lipgloss.Style
	SpinnerStyle lipgloss.Style

	// Layout
	PaneBorder   lipgloss.Style
	ActiveBorder lipgloss.Style
	Separator    lipgloss.Style
}

// CatppuccinMocha returns a Theme styled with the Catppuccin Mocha color palette.
func CatppuccinMocha() Theme {
	return Theme{
		// Header
		HeaderBar:    lipgloss.NewStyle().Background(CatBase).Foreground(CatText).Padding(0, 1),
		HeaderTitle:  lipgloss.NewStyle().Bold(true).Foreground(CatBlue),
		TimerText:    lipgloss.NewStyle().Bold(true).Foreground(CatGreen),
		TimerStopped: lipgloss.NewStyle().Foreground(CatOverlay1),

		// Footer
		FooterBar:  lipgloss.NewStyle().Background(CatSurface0).Foreground(CatSubtext0).Padding(0, 1),
		FooterKey:  lipgloss.NewStyle().Bold(true).Foreground(CatMauve),
		FooterDesc: lipgloss.NewStyle().Foreground(CatSubtext0),

		// List items
		SelectedItem: lipgloss.NewStyle().Background(CatSurface1).Foreground(CatText).Bold(true).Padding(0, 1),
		NormalItem:   lipgloss.NewStyle().Foreground(CatText).Padding(0, 1),
		DimmedItem:   lipgloss.NewStyle().Foreground(CatOverlay0).Padding(0, 1),

		// Tracking indicators
		TrackingActive: lipgloss.NewStyle().Foreground(CatGreen).Bold(true),
		TrackingLast:   lipgloss.NewStyle().Foreground(CatYellow),
		TrackingNone:   lipgloss.NewStyle().Foreground(CatOverlay0),

		// Filter input
		FilterPrompt: lipgloss.NewStyle().Foreground(CatMauve).Bold(true),
		FilterText:   lipgloss.NewStyle().Foreground(CatText),
		FilterMatch:  lipgloss.NewStyle().Foreground(CatPeach).Bold(true),

		// Status messages
		ErrorText:    lipgloss.NewStyle().Foreground(CatRed).Bold(true),
		SuccessText:  lipgloss.NewStyle().Foreground(CatGreen).Bold(true),
		EmptyText:    lipgloss.NewStyle().Foreground(CatOverlay1),
		SpinnerStyle: lipgloss.NewStyle().Foreground(CatBlue),

		// Layout
		PaneBorder:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(CatSurface2),
		ActiveBorder: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(CatBlue),
		Separator:    lipgloss.NewStyle().Foreground(CatSurface2),
	}
}

// PlainTheme returns a Theme with no colors, suitable for any terminal.
// It uses bold and underline for emphasis instead of colors.
func PlainTheme() Theme {
	return Theme{
		// Header
		HeaderBar:    lipgloss.NewStyle().Padding(0, 1),
		HeaderTitle:  lipgloss.NewStyle().Bold(true),
		TimerText:    lipgloss.NewStyle().Bold(true),
		TimerStopped: lipgloss.NewStyle(),

		// Footer
		FooterBar:  lipgloss.NewStyle().Padding(0, 1),
		FooterKey:  lipgloss.NewStyle().Bold(true),
		FooterDesc: lipgloss.NewStyle(),

		// List items
		SelectedItem: lipgloss.NewStyle().Bold(true).Underline(true).Padding(0, 1),
		NormalItem:   lipgloss.NewStyle().Padding(0, 1),
		DimmedItem:   lipgloss.NewStyle().Padding(0, 1),

		// Tracking indicators
		TrackingActive: lipgloss.NewStyle().Bold(true),
		TrackingLast:   lipgloss.NewStyle(),
		TrackingNone:   lipgloss.NewStyle(),

		// Filter input
		FilterPrompt: lipgloss.NewStyle().Bold(true),
		FilterText:   lipgloss.NewStyle(),
		FilterMatch:  lipgloss.NewStyle().Bold(true).Underline(true),

		// Status messages
		ErrorText:    lipgloss.NewStyle().Bold(true),
		SuccessText:  lipgloss.NewStyle().Bold(true),
		EmptyText:    lipgloss.NewStyle(),
		SpinnerStyle: lipgloss.NewStyle(),

		// Layout
		PaneBorder:   lipgloss.NewStyle().Border(lipgloss.NormalBorder()),
		ActiveBorder: lipgloss.NewStyle().Border(lipgloss.DoubleBorder()),
		Separator:    lipgloss.NewStyle(),
	}
}

// GetTheme returns the theme by name. Falls back to PlainTheme for unknown names.
func GetTheme(name string) Theme {
	switch name {
	case "catppuccin-mocha":
		return CatppuccinMocha()
	case "plain":
		return PlainTheme()
	default:
		return PlainTheme()
	}
}

// TrackingIndicator returns the styled indicator string.
// If tracking is true and isActive is true, it shows a filled circle (●) in green.
// If tracking is false but isActive is true, it shows an empty circle (○) in yellow.
// Otherwise it shows a space for inactive items.
func (t Theme) TrackingIndicator(tracking bool, isActive bool) string {
	if tracking && isActive {
		return t.TrackingActive.Render("●")
	}
	if isActive {
		return t.TrackingLast.Render("○")
	}
	return t.TrackingNone.Render(" ")
}

// FormatTimer formats a duration as HH:MM:SS.
func FormatTimer(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
