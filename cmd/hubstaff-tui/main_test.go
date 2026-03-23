package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout redirects os.Stdout, calls fn, then restores it and returns
// everything that was written.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old

	buf, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("io.ReadAll: %v", err)
	}
	return string(buf)
}

// --- tmuxSnippet ---

func TestTmuxSnippet_ContainsBindKey(t *testing.T) {
	if !strings.Contains(tmuxSnippet, "bind-key H") {
		t.Errorf("tmuxSnippet does not contain 'bind-key H':\n%s", tmuxSnippet)
	}
}

func TestTmuxSnippet_ContainsDisplayPopup(t *testing.T) {
	if !strings.Contains(tmuxSnippet, "display-popup") {
		t.Errorf("tmuxSnippet does not contain 'display-popup':\n%s", tmuxSnippet)
	}
}

func TestTmuxSnippet_ContainsHubstaffTUI(t *testing.T) {
	if !strings.Contains(tmuxSnippet, "hubstaff-tui") {
		t.Errorf("tmuxSnippet does not contain 'hubstaff-tui':\n%s", tmuxSnippet)
	}
}

// --- tmuxStatusSnippet ---

func TestTmuxStatusSnippet_ContainsStatusRight(t *testing.T) {
	if !strings.Contains(tmuxStatusSnippet, "status-right") {
		t.Errorf("tmuxStatusSnippet does not contain 'status-right':\n%s", tmuxStatusSnippet)
	}
}

func TestTmuxStatusSnippet_ContainsInterval(t *testing.T) {
	if !strings.Contains(tmuxStatusSnippet, "status-interval") {
		t.Errorf("tmuxStatusSnippet does not contain 'status-interval':\n%s", tmuxStatusSnippet)
	}
}

// --- printUsage ---

func TestPrintUsage_ContainsSubcommands(t *testing.T) {
	output := captureStdout(t, printUsage)

	for _, want := range []string{"status", "setup", "--help", "--version"} {
		if !strings.Contains(output, want) {
			t.Errorf("printUsage output missing %q:\n%s", want, output)
		}
	}
}
