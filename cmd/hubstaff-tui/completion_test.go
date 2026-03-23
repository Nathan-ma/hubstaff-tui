package main

import (
	"strings"
	"testing"
)

func TestRunCompletion_Bash(t *testing.T) {
	// capture would be ideal, but since we use fmt.Print, just verify no panic
	// and that constants have expected content
	if !strings.Contains(bashCompletion, "hubstaff-tui") {
		t.Error("bash completion should reference hubstaff-tui")
	}
}

func TestRunCompletion_Zsh(t *testing.T) {
	if !strings.Contains(zshCompletion, "_hubstaff_tui") {
		t.Error("zsh completion should define _hubstaff_tui function")
	}
}

func TestRunCompletion_Fish(t *testing.T) {
	if !strings.Contains(fishCompletion, "complete -c hubstaff-tui") {
		t.Error("fish completion should have complete commands")
	}
}

func TestRunCompletion_UnknownShell(t *testing.T) {
	code := runCompletion([]string{"powershell"})
	if code != 1 {
		t.Errorf("expected exit code 1 for unknown shell, got %d", code)
	}
}

func TestRunCompletion_NoArgs(t *testing.T) {
	code := runCompletion([]string{})
	if code != 0 {
		t.Errorf("expected exit code 0 for no args (shows help), got %d", code)
	}
}
