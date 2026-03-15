package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Nathan-ma/hubstaff-tui/internal/api"
	"github.com/Nathan-ma/hubstaff-tui/internal/config"
	"github.com/Nathan-ma/hubstaff-tui/internal/ui"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--help", "-h":
			printUsage()
			os.Exit(0)
		case "--version", "-v":
			fmt.Println("hubstaff-tui", version)
			os.Exit(0)
		case "status":
			fmt.Println("○ Not tracking")
			os.Exit(0)
		case "setup":
			fmt.Println("hubstaff-tui setup: not yet implemented")
			os.Exit(0)
		}
	}

	// Load config
	configPath := ""
	for i, arg := range os.Args[1:] {
		if arg == "--config" && i+2 < len(os.Args) {
			configPath = os.Args[i+2]
		}
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client := api.NewClient(cfg.Hubstaff.CLIPath)

	model := ui.NewApp(cfg, client)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
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
