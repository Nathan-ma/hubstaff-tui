package ui

import (
	"testing"
	"time"
)

func TestParseTrackedToday(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"0:00:00", 0},
		{"1:00:00", time.Hour},
		{"2:15:30", 2*time.Hour + 15*time.Minute + 30*time.Second},
		{"0:05:10", 5*time.Minute + 10*time.Second},
		{"", 0},
		{"invalid", 0},
		{"1:2", 0}, // wrong format — only 2 parts
		{"10:30:45", 10*time.Hour + 30*time.Minute + 45*time.Second},
		{"0:00:01", time.Second},
		{"abc:de:fg", 0}, // non-numeric parts produce 0 via Atoi
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := parseTrackedToday(tc.input)
			if got != tc.expected {
				t.Errorf("parseTrackedToday(%q) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestIsFiltering_DefaultScreens(t *testing.T) {
	// Verify isFiltering returns false for a freshly-created summary screen.
	m := AppModel{current: screenSummary}
	if m.isFiltering() {
		t.Error("expected isFiltering() = false for screenSummary")
	}
}

func TestScreenConstants(t *testing.T) {
	// Ensure the screen enum values are distinct.
	screens := []screen{screenProjects, screenTasks, screenSummary}
	seen := make(map[screen]bool)
	for _, s := range screens {
		if seen[s] {
			t.Errorf("duplicate screen constant value: %d", s)
		}
		seen[s] = true
	}
}
