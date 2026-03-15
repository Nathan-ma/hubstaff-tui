package main

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Nathan-ma/hubstaff-tui/internal/api"
	"github.com/Nathan-ma/hubstaff-tui/internal/config"
	"github.com/Nathan-ma/hubstaff-tui/internal/store"
	"github.com/Nathan-ma/hubstaff-tui/internal/ui"
)

var version = "dev"

func main() {
	// Parse --config flag early so all subcommands can use it
	configPath := ""
	for i, arg := range os.Args[1:] {
		if arg == "--config" && i+2 < len(os.Args) {
			configPath = os.Args[i+2]
		}
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--help", "-h":
			printUsage()
			os.Exit(0)
		case "--version", "-v":
			fmt.Println("hubstaff-tui", version)
			os.Exit(0)
		case "status":
			runStatus(configPath)
			os.Exit(0)
		case "setup":
			fmt.Println("hubstaff-tui setup: not yet implemented")
			os.Exit(0)
		}
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client := api.NewClient(cfg.Hubstaff.CLIPath)

	// Open local store for session tracking and summaries
	dbPath, err := cfg.ResolvedDBPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving DB path: %v\n", err)
		os.Exit(1)
	}
	ttl := time.Duration(cfg.Store.TTLSeconds) * time.Second
	st, err := store.Open(dbPath, ttl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening store: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = st.Close() }()

	model := ui.NewApp(cfg, client, st)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runStatus(configPath string) {
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Println("○ Hubstaff unavailable")
		return
	}

	useColor := false
	for _, arg := range os.Args[2:] {
		if arg == "--color" {
			useColor = true
		}
	}

	client := api.NewClient(cfg.Hubstaff.CLIPath)
	s, err := client.GetStatus(context.Background())
	if err != nil {
		fmt.Println("○ Hubstaff unavailable")
		return
	}

	// Side effect: update heartbeat in store
	dbPath, err := cfg.ResolvedDBPath()
	if err == nil {
		if st, err := store.Open(dbPath, time.Duration(cfg.Store.TTLSeconds)*time.Second); err == nil {
			_ = st.UpdateHeartbeat()
			_ = st.Close()
		}
	}

	if !s.Tracking {
		if useColor {
			fmt.Println("\033[2m○ Not tracking\033[0m")
		} else {
			fmt.Println("○ Not tracking")
		}
		return
	}

	// Format: ◉ HH:MM:SS  Project › Task
	tracked := s.ActiveProject.TrackedToday
	project := s.ActiveProject.Name
	task := s.ActiveTask.Name

	if useColor {
		fmt.Printf("\033[32m◉\033[0m %s  %s › %s\n", tracked, project, task)
	} else {
		fmt.Printf("◉ %s  %s › %s\n", tracked, project, task)
	}
}

func printUsage() {
	fmt.Print(`hubstaff-tui — Fast Hubstaff time tracking TUI for tmux popups

Usage:
  hubstaff-tui            Launch the interactive TUI
  hubstaff-tui status     Print current tracking status (for tmux status-right)
  hubstaff-tui setup      Configure tmux keybinding

Options:
  --help, -h              Show this help
  --version, -v           Show version
  --config <path>         Use a custom config file
`)
}
