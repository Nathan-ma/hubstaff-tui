package state

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DefaultStatePath is the default location for the state file.
const DefaultStatePath = "~/.local/share/hubstaff-tui/state.json"

// AppState holds navigation state that persists between TUI sessions.
type AppState struct {
	LastProjectID   string `json:"last_project_id"`
	LastProjectName string `json:"last_project_name"`
	LastTaskID      string `json:"last_task_id,omitempty"`
	ScrollPosition  int    `json:"scroll_position"`
}

// Load reads the state file. Returns zero state if file doesn't exist or is invalid.
func Load(path string) AppState {
	path = expandPath(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return AppState{}
	}
	var s AppState
	if err := json.Unmarshal(data, &s); err != nil {
		return AppState{}
	}
	return s
}

// Save writes the state file. Creates parent dirs if needed.
func Save(path string, s AppState) error {
	path = expandPath(path)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[1:])
		}
	}
	return path
}
