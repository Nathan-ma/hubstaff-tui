package api

// Status represents the current tracking state returned by HubstaffCLI status.
type Status struct {
	Tracking      bool          `json:"tracking"`
	ActiveProject ActiveProject `json:"active_project"`
	ActiveTask    ActiveTask    `json:"active_task"`
}

// ActiveProject holds the project currently being tracked.
type ActiveProject struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	TrackedToday string `json:"tracked_today"` // "H:MM:SS"
}

// ActiveTask holds the task currently being tracked.
type ActiveTask struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Project is a Hubstaff project summary.
type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Task is a Hubstaff task summary.
type Task struct {
	ID        string `json:"id"`
	Summary   string `json:"summary"`
	ProjectID string `json:"-"` // injected at display time, not from API
}

// API response wrappers

type projectsResponse struct {
	Projects []Project `json:"projects"`
}

type tasksResponse struct {
	Tasks []Task `json:"tasks"`
}
