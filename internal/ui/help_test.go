package ui

import (
	"strings"
	"testing"

	"github.com/Nathan-ma/hubstaff-tui/internal/config"
)

func TestNewHelpModel(t *testing.T) {
	theme := CatppuccinMocha()
	h := NewHelpModel(theme, NewKeyMap(config.DefaultConfig().Keybindings))

	if h.viewport.Width != 0 {
		t.Errorf("expected initial viewport width 0, got %d", h.viewport.Width)
	}
	if h.viewport.Height != 0 {
		t.Errorf("expected initial viewport height 0, got %d", h.viewport.Height)
	}
}

func TestHelpModel_SetSize(t *testing.T) {
	theme := CatppuccinMocha()
	h := NewHelpModel(theme, NewKeyMap(config.DefaultConfig().Keybindings))
	h.SetSize(80, 40)

	if h.width != 80 {
		t.Errorf("expected width 80, got %d", h.width)
	}
	if h.height != 40 {
		t.Errorf("expected height 40, got %d", h.height)
	}
	if h.viewport.Width == 0 {
		t.Error("viewport width should be non-zero after SetSize")
	}
	if h.viewport.Height == 0 {
		t.Error("viewport height should be non-zero after SetSize")
	}
}

func TestHelpModel_SetSize_NarrowTerminal(t *testing.T) {
	theme := CatppuccinMocha()
	h := NewHelpModel(theme, NewKeyMap(config.DefaultConfig().Keybindings))
	h.SetSize(30, 20)

	// Should still have reasonable dimensions
	if h.viewport.Width < 18 {
		t.Errorf("viewport width too small: %d", h.viewport.Width)
	}
}

func TestHelpModel_View_ContainsKeybindings(t *testing.T) {
	theme := CatppuccinMocha()
	h := NewHelpModel(theme, NewKeyMap(config.DefaultConfig().Keybindings))
	h.SetSize(80, 40)

	view := h.View()

	expectedSections := []string{
		"Navigation",
		"Tracking",
		"Views",
		"General",
	}
	for _, section := range expectedSections {
		if !strings.Contains(view, section) {
			t.Errorf("help view missing section %q", section)
		}
	}

	expectedKeys := []string{
		"Enter",
		"esc",
		"ctrl+e",
		"ctrl+r",
		"Ctrl+C",
	}
	for _, key := range expectedKeys {
		if !strings.Contains(view, key) {
			t.Errorf("help view missing key %q", key)
		}
	}
}

func TestHelpModel_View_ContainsTitle(t *testing.T) {
	theme := CatppuccinMocha()
	h := NewHelpModel(theme, NewKeyMap(config.DefaultConfig().Keybindings))
	h.SetSize(80, 40)

	view := h.View()
	if !strings.Contains(view, "Keybindings") {
		t.Error("help view should contain 'Keybindings' title")
	}
}

func TestHelpContent_AllSections(t *testing.T) {
	theme := CatppuccinMocha()
	h := NewHelpModel(theme, NewKeyMap(config.DefaultConfig().Keybindings))
	content := h.helpContent(56)

	// Should have all four sections
	for _, section := range []string{"Navigation", "Tracking", "Views", "General"} {
		if !strings.Contains(content, section) {
			t.Errorf("helpContent missing section %q", section)
		}
	}

	// Should have descriptions
	descriptions := []string{
		"Select project",
		"Back / Quit",
		"Stop tracking",
		"Toggle this help",
		"Force quit",
		"Fuzzy filter",
		"Refresh",
	}
	for _, desc := range descriptions {
		if !strings.Contains(content, desc) {
			t.Errorf("helpContent missing description %q", desc)
		}
	}
}
