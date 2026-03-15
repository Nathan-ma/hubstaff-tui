package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const defaultCLIPath = "/Applications/Hubstaff.app/Contents/MacOS/HubstaffCLI"
const defaultTimeout = 10 * time.Second

// CLIError is returned when the HubstaffCLI command exits with a non-zero status.
type CLIError struct {
	Command  string
	ExitCode int
	Stderr   string
}

func (e *CLIError) Error() string {
	return fmt.Sprintf("hubstaff cli %s failed (exit %d): %s", e.Command, e.ExitCode, e.Stderr)
}

// Client wraps the HubstaffCLI binary.
type Client struct {
	cliPath string
	timeout time.Duration
}

// NewClient creates a new Client. If cliPath is empty the default macOS path is used.
func NewClient(cliPath string) *Client {
	if cliPath == "" {
		cliPath = defaultCLIPath
	}
	return &Client{cliPath: cliPath, timeout: defaultTimeout}
}

// CheckCLI verifies the HubstaffCLI binary exists and is executable.
// Returns nil if OK, or an error describing the problem.
func (c *Client) CheckCLI() error {
	path := c.cliPath
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("HubstaffCLI not found at %s\n\nMake sure the Hubstaff desktop app is installed.\nOr set a custom path in ~/.config/hubstaff-tui/config.toml:\n\n  [hubstaff]\n  cli_path = \"/path/to/HubstaffCLI\"", path)
	}
	if err != nil {
		return fmt.Errorf("cannot access HubstaffCLI at %s: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("HubstaffCLI path is a directory, not a file: %s", path)
	}
	// Check if executable (unix only)
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("HubstaffCLI at %s is not executable", path)
	}
	return nil
}

// run executes the HubstaffCLI binary with the given arguments and returns stdout bytes.
func (c *Client) run(ctx context.Context, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.cliPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		exitCode := 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		cmdName := strings.Join(args, " ")
		if cmdName == "" {
			cmdName = "unknown"
		}
		return nil, &CLIError{
			Command:  cmdName,
			ExitCode: exitCode,
			Stderr:   strings.TrimSpace(stderr.String()),
		}
	}

	return stdout.Bytes(), nil
}

// GetStatus returns the current tracking status.
func (c *Client) GetStatus(ctx context.Context) (Status, error) {
	data, err := c.run(ctx, "status")
	if err != nil {
		return Status{}, err
	}
	return parseStatus(data)
}

// ListProjects returns all available projects.
func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	data, err := c.run(ctx, "projects")
	if err != nil {
		return nil, err
	}
	return parseProjects(data)
}

// ListTasks returns the tasks for a given project.
func (c *Client) ListTasks(ctx context.Context, projectID string) ([]Task, error) {
	data, err := c.run(ctx, "tasks", projectID)
	if err != nil {
		return nil, err
	}
	return parseTasks(data)
}

// StartTask starts tracking the given task.
func (c *Client) StartTask(ctx context.Context, taskID string) error {
	_, err := c.run(ctx, "start_task", taskID)
	return err
}

// Stop stops the current tracking session.
func (c *Client) Stop(ctx context.Context) error {
	_, err := c.run(ctx, "stop")
	return err
}

// parse helpers

func parseStatus(data []byte) (Status, error) {
	var s Status
	if err := json.Unmarshal(data, &s); err != nil {
		return Status{}, fmt.Errorf("parse status: %w", err)
	}
	return s, nil
}

func parseProjects(data []byte) ([]Project, error) {
	var resp projectsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse projects: %w", err)
	}
	if resp.Projects == nil {
		resp.Projects = []Project{}
	}
	return resp.Projects, nil
}

func parseTasks(data []byte) ([]Task, error) {
	var resp tasksResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse tasks: %w", err)
	}
	if resp.Tasks == nil {
		resp.Tasks = []Task{}
	}
	return resp.Tasks, nil
}
