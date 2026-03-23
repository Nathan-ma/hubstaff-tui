package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Nathan-ma/hubstaff-tui/internal/api"
	"github.com/Nathan-ma/hubstaff-tui/internal/config"
	"github.com/Nathan-ma/hubstaff-tui/internal/store"
)

const (
	labelPass = "  [PASS]"
	labelFail = "  [FAIL]"
	labelWarn = "  [WARN]"
)

// runDoctor runs all setup diagnostic checks and prints a plain-text report.
// Returns exit code 0 if all required checks pass, 1 if any required check fails.
func runDoctor(cfg *config.Config, configPath string) int {
	fmt.Println("hubstaff-tui doctor")
	fmt.Println()

	exitCode := 0

	// Check 1: HubstaffCLI binary
	if !checkCLIBinary(cfg.Hubstaff.CLIPath) {
		exitCode = 1
	}

	// Check 2: Config file
	checkConfigFile(configPath)

	// Check 3: Database
	if !checkDatabase(cfg) {
		exitCode = 1
	}

	// Check 4: Active session (optional — never causes failure)
	checkActiveSession(cfg)

	fmt.Println()
	if exitCode == 0 {
		fmt.Println("All required checks passed.")
	} else {
		fmt.Println("One or more checks failed. See above for details.")
	}

	return exitCode
}

// checkCLIBinary verifies the HubstaffCLI binary exists, is executable, and responds to --version.
// Returns true if the check passes.
func checkCLIBinary(cliPath string) bool {
	label := "HubstaffCLI"

	// Resolve path: if not absolute, try PATH lookup.
	resolvedPath := cliPath
	if !strings.HasPrefix(cliPath, "/") {
		if p, err := exec.LookPath(cliPath); err == nil {
			resolvedPath = p
		}
	}

	info, err := os.Stat(resolvedPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("%s %s: not found at %s\n", labelFail, label, resolvedPath)
		} else {
			fmt.Printf("%s %s: cannot access %s: %v\n", labelFail, label, resolvedPath, err)
		}
		return false
	}

	if info.IsDir() {
		fmt.Printf("%s %s: path is a directory, not a file: %s\n", labelFail, label, resolvedPath)
		return false
	}

	if info.Mode()&0111 == 0 {
		fmt.Printf("%s %s: not executable: %s\n", labelFail, label, resolvedPath)
		return false
	}

	// Run --version with a 5-second timeout to verify the binary works.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, resolvedPath, "--version")
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("%s %s: %s (binary found but --version failed: %v)\n", labelFail, label, resolvedPath, err)
		return false
	}

	version := strings.TrimSpace(string(out))
	fmt.Printf("%s %s: %s (%s)\n", labelPass, label, resolvedPath, version)
	return true
}

// checkConfigFile reports whether the config file exists and where it lives.
// A missing config file is a warning (defaults are used), not a failure.
func checkConfigFile(configPath string) {
	label := "Config"

	path := configPath
	if path == "" {
		path = config.DefaultConfigPath
	}

	expanded, err := config.ExpandPath(path)
	if err != nil {
		fmt.Printf("%s %s: cannot expand path %s: %v\n", labelWarn, label, path, err)
		return
	}

	if _, err := os.Stat(expanded); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("%s %s: not found at %s (defaults in use)\n", labelWarn, label, expanded)
		} else {
			fmt.Printf("%s %s: cannot access %s: %v\n", labelWarn, label, expanded, err)
		}
		return
	}

	fmt.Printf("%s %s: %s\n", labelPass, label, expanded)
}

// checkDatabase verifies the SQLite database path is accessible (or can be created).
// Returns true if the check passes.
func checkDatabase(cfg *config.Config) bool {
	label := "Database"

	dbPath, err := cfg.ResolvedDBPath()
	if err != nil {
		fmt.Printf("%s %s: cannot resolve path %s: %v\n", labelFail, label, cfg.Store.DBPath, err)
		return false
	}

	if _, err := os.Stat(dbPath); err != nil {
		if os.IsNotExist(err) {
			// DB doesn't exist yet — not a hard failure; it will be created on first run.
			fmt.Printf("%s %s: %s (will be created on first run)\n", labelWarn, label, dbPath)
			return true
		}
		fmt.Printf("%s %s: cannot access %s: %v\n", labelFail, label, dbPath, err)
		return false
	}

	// File exists; try opening it to confirm accessibility.
	ttl := time.Duration(cfg.Store.TTLSeconds) * time.Second
	st, err := store.Open(dbPath, ttl)
	if err != nil {
		fmt.Printf("%s %s: %s (open failed: %v)\n", labelFail, label, dbPath, err)
		return false
	}
	_ = st.Close()

	fmt.Printf("%s %s: %s\n", labelPass, label, dbPath)
	return true
}

// checkActiveSession calls GetStatus and prints the tracking state.
// This check never causes a failure — errors are shown as warnings.
func checkActiveSession(cfg *config.Config) {
	label := "Active session"

	client := api.NewClient(cfg.Hubstaff.CLIPath)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s, err := client.GetStatus(ctx)
	if err != nil {
		fmt.Printf("%s %s: could not query status (%v)\n", labelWarn, label, err)
		return
	}

	if !s.Tracking {
		fmt.Printf("%s %s: not tracking\n", labelWarn, label)
		return
	}

	fmt.Printf("%s %s: tracking %s › %s (%s)\n",
		labelPass, label,
		s.ActiveProject.Name,
		s.ActiveTask.Name,
		s.ActiveProject.TrackedToday,
	)
}
