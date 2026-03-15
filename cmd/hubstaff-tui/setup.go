package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const tmuxSnippet = `
# Hubstaff popup (prefix + H) — added by hubstaff-tui setup
bind-key H display-popup \
  -E \
  -w 80%% \
  -h 70%% \
  -b rounded \
  -T " Hubstaff " \
  -e "HUBSTAFF_POPUP=1" \
  -d "#{pane_current_path}" \
  "hubstaff-tui"
`

const tmuxStatusSnippet = `
# Hubstaff status bar — added by hubstaff-tui setup
set -g status-right '#(hubstaff-tui status) | %H:%M'
set -g status-interval 10
`

func runSetup() {
	// 1. Check if tmux is available
	if _, err := exec.LookPath("tmux"); err != nil {
		fmt.Fprintln(os.Stderr, "Error: tmux not found in PATH")
		fmt.Fprintln(os.Stderr, "hubstaff-tui requires tmux 3.2+ for popup support")
		os.Exit(1)
	}

	// 2. Find tmux.conf path
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not determine home directory: %v\n", err)
		os.Exit(1)
	}
	confPath := filepath.Join(home, ".tmux.conf")

	// 3. Read existing config (or empty)
	existing := ""
	if data, err := os.ReadFile(confPath); err == nil {
		existing = string(data)
	}

	// 4. Check if binding already exists
	if strings.Contains(existing, "hubstaff-tui") {
		fmt.Println("hubstaff-tui is already configured in ~/.tmux.conf")
		fmt.Println("  To reconfigure, remove the hubstaff-tui lines and run setup again.")
		os.Exit(0)
	}

	// 5. Append snippets
	f, err := os.OpenFile(confPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not write to %s: %v\n", confPath, err)
		os.Exit(1)
	}
	defer f.Close()

	if _, err := f.WriteString("\n" + tmuxSnippet + "\n" + tmuxStatusSnippet); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not append to %s: %v\n", confPath, err)
		os.Exit(1)
	}

	fmt.Println("Added hubstaff-tui keybinding to ~/.tmux.conf")
	fmt.Println("  Popup: prefix + H")
	fmt.Println("  Status bar: shows current tracking state")

	// 6. Reload tmux config if tmux is running
	if err := exec.Command("tmux", "source-file", confPath).Run(); err != nil {
		fmt.Println("\n  Note: tmux config written but could not reload.")
		fmt.Println("  Run: tmux source-file ~/.tmux.conf")
	} else {
		fmt.Println("  tmux config reloaded")
	}
}
