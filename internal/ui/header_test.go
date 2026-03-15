package ui

import (
	"testing"
	"time"
)

func TestFormatTimer_Various(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{0, "00:00:00"},
		{time.Second, "00:00:01"},
		{time.Minute, "00:01:00"},
		{time.Hour, "01:00:00"},
		{3*time.Hour + 45*time.Minute + 12*time.Second, "03:45:12"},
		{99*time.Hour + 59*time.Minute + 59*time.Second, "99:59:59"},
		// Sub-second durations should truncate to 0 seconds.
		{500 * time.Millisecond, "00:00:00"},
		// Durations with fractional seconds should floor.
		{61*time.Second + 999*time.Millisecond, "00:01:01"},
	}
	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			got := FormatTimer(tc.input)
			if got != tc.expected {
				t.Errorf("FormatTimer(%v) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}
