package ui

import "testing"

func TestBellCmd(t *testing.T) {
	cmd := bellCmd()
	if cmd == nil {
		t.Fatal("bellCmd returned nil")
	}
	// Execute it (writes \a to stderr, harmless in tests)
	msg := cmd()
	// Should return nil msg
	if msg != nil {
		t.Fatalf("expected nil msg, got %v", msg)
	}
}
