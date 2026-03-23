package api

import (
	"encoding/json"
	"fmt"
)

// FlexibleID handles JSON values that can be either a string or a number.
// HubstaffCLI returns numeric IDs but we want to work with strings internally.
type FlexibleID string

// UnmarshalJSON implements json.Unmarshaler for FlexibleID.
func (f *FlexibleID) UnmarshalJSON(data []byte) error {
	// Reject null explicitly
	if string(data) == "null" {
		return fmt.Errorf("FlexibleID: cannot unmarshal null")
	}
	// Try string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = FlexibleID(s)
		return nil
	}
	// Try number
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexibleID(n.String())
		return nil
	}
	return fmt.Errorf("FlexibleID: cannot unmarshal %s", string(data))
}

// String returns the string representation of the ID.
func (f FlexibleID) String() string {
	return string(f)
}

// Status represents the current tracking state returned by HubstaffCLI status.
type Status struct {
	Tracking      bool          `json:"tracking"`
	ActiveProject ActiveProject `json:"active_project"`
	ActiveTask    ActiveTask    `json:"active_task"`
}

// ActiveProject holds the project currently being tracked.
type ActiveProject struct {
	ID           FlexibleID `json:"id"`
	Name         string     `json:"name"`
	TrackedToday string     `json:"tracked_today"` // "H:MM:SS"
}

// ActiveTask holds the task currently being tracked.
type ActiveTask struct {
	ID   FlexibleID `json:"id"`
	Name string     `json:"name"`
}

// Project is a Hubstaff project summary.
type Project struct {
	ID   FlexibleID `json:"id"`
	Name string     `json:"name"`
}

// Task is a Hubstaff task summary.
type Task struct {
	ID        FlexibleID `json:"id"`
	Summary   string     `json:"summary"`
	ProjectID string     `json:"-"` // injected at display time, not from API
}

// API response wrappers

type projectsResponse struct {
	Projects []Project `json:"projects"`
}

type tasksResponse struct {
	Tasks []Task `json:"tasks"`
}
