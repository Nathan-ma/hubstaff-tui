package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func TestCatppuccinMocha_NotZeroValues(t *testing.T) {
	theme := CatppuccinMocha()

	// Verify key styles have foreground colors set (not zero-value NoColor).
	checks := []struct {
		name  string
		style lipgloss.Style
	}{
		{"HeaderTitle", theme.HeaderTitle},
		{"TimerText", theme.TimerText},
		{"SelectedItem", theme.SelectedItem},
		{"TrackingActive", theme.TrackingActive},
		{"ErrorText", theme.ErrorText},
		{"SuccessText", theme.SuccessText},
		{"FooterKey", theme.FooterKey},
		{"FilterPrompt", theme.FilterPrompt},
		{"SpinnerStyle", theme.SpinnerStyle},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			fg := c.style.GetForeground()
			if _, ok := fg.(lipgloss.NoColor); ok {
				t.Errorf("%s: expected foreground color to be set, got NoColor", c.name)
			}
		})
	}
}

func TestPlainTheme_Renders(t *testing.T) {
	theme := PlainTheme()

	// Plain theme styles should have no foreground color set.
	plainChecks := []struct {
		name  string
		style lipgloss.Style
	}{
		{"TimerStopped", theme.TimerStopped},
		{"FooterDesc", theme.FooterDesc},
		{"NormalItem", theme.NormalItem},
		{"TrackingNone", theme.TrackingNone},
	}

	for _, c := range plainChecks {
		t.Run(c.name, func(t *testing.T) {
			fg := c.style.GetForeground()
			if _, ok := fg.(lipgloss.NoColor); !ok {
				t.Errorf("PlainTheme %s: expected NoColor foreground, got %v", c.name, fg)
			}
		})
	}
}

func TestGetTheme_Known(t *testing.T) {
	theme := GetTheme("catppuccin-mocha")

	// The catppuccin theme's HeaderTitle should have a foreground color set.
	fg := theme.HeaderTitle.GetForeground()
	if _, ok := fg.(lipgloss.NoColor); ok {
		t.Error("GetTheme('catppuccin-mocha'): HeaderTitle expected foreground color, got NoColor")
	}
}

func TestGetTheme_Unknown(t *testing.T) {
	theme := GetTheme("nonexistent-theme")

	// Unknown theme should fall back to plain. TimerStopped should have no color.
	fg := theme.TimerStopped.GetForeground()
	if _, ok := fg.(lipgloss.NoColor); !ok {
		t.Errorf("GetTheme('nonexistent-theme'): expected plain fallback with NoColor, got %v", fg)
	}
}

func TestTrackingIndicator_Active(t *testing.T) {
	theme := CatppuccinMocha()
	result := theme.TrackingIndicator(true, true)
	if !strings.Contains(result, "●") {
		t.Errorf("TrackingIndicator(true, true): expected '●' in output, got %q", result)
	}
}

func TestTrackingIndicator_Last(t *testing.T) {
	theme := CatppuccinMocha()
	result := theme.TrackingIndicator(false, true)
	if !strings.Contains(result, "○") {
		t.Errorf("TrackingIndicator(false, true): expected '○' in output, got %q", result)
	}
}

func TestTrackingIndicator_None(t *testing.T) {
	theme := CatppuccinMocha()
	result := theme.TrackingIndicator(false, false)
	if !strings.Contains(result, " ") {
		t.Errorf("TrackingIndicator(false, false): expected space in output, got %q", result)
	}
}

func TestFormatTimer(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "1h23m45s",
			duration: 1*time.Hour + 23*time.Minute + 45*time.Second,
			expected: "01:23:45",
		},
		{
			name:     "zero",
			duration: 0,
			expected: "00:00:00",
		},
		{
			name:     "just seconds",
			duration: 5 * time.Second,
			expected: "00:00:05",
		},
		{
			name:     "large hours",
			duration: 100*time.Hour + 0*time.Minute + 1*time.Second,
			expected: "100:00:01",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatTimer(tc.duration)
			if got != tc.expected {
				t.Errorf("FormatTimer(%v) = %q, want %q", tc.duration, got, tc.expected)
			}
		})
	}
}
