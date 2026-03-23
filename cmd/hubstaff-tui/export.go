package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Nathan-ma/hubstaff-tui/internal/config"
	"github.com/Nathan-ma/hubstaff-tui/internal/store"
)

// exportRecord is the flat representation written to CSV or JSON output.
type exportRecord struct {
	Date            string `json:"date"`
	Project         string `json:"project"`
	Task            string `json:"task"`
	DurationSeconds int    `json:"duration_seconds"`
	StartedAt       string `json:"started_at"`
	StoppedAt       string `json:"stopped_at"`
}

func runExport(cfg *config.Config, args []string) int {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	format := fs.String("format", "csv", "Output format: csv or json")
	today := fs.Bool("today", false, "Export today's sessions (default if no range flag given)")
	week := fs.Bool("week", false, "Export the last 7 days of sessions")
	since := fs.String("since", "", "Export sessions from this ISO date (YYYY-MM-DD) to now")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "export: %v\n", err)
		return 1
	}

	if *format != "csv" && *format != "json" {
		fmt.Fprintf(os.Stderr, "export: --format must be csv or json\n")
		return 1
	}

	// Determine date range.
	now := time.Now().UTC()
	var start, end time.Time
	end = now.Add(24 * time.Hour) // inclusive upper bound (tomorrow midnight)

	switch {
	case *since != "":
		parsed, err := time.Parse("2006-01-02", *since)
		if err != nil {
			fmt.Fprintf(os.Stderr, "export: invalid --since date %q (expected YYYY-MM-DD)\n", *since)
			return 1
		}
		start = parsed.UTC()
	case *week:
		start = now.AddDate(0, 0, -7).Truncate(24 * time.Hour)
	default:
		// --today is the default
		_ = today
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	}

	// Open the store.
	dbPath, err := cfg.ResolvedDBPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "export: resolve db path: %v\n", err)
		return 1
	}
	ttl := time.Duration(cfg.Store.TTLSeconds) * time.Second
	st, err := store.Open(dbPath, ttl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "export: open store: %v\n", err)
		return 1
	}
	defer func() { _ = st.Close() }()

	sessions, err := st.GetSessionsInRange(start, end)
	if err != nil {
		fmt.Fprintf(os.Stderr, "export: query sessions: %v\n", err)
		return 1
	}

	records := make([]exportRecord, 0, len(sessions))
	for _, s := range sessions {
		stoppedAt := ""
		if s.StoppedAt != nil {
			stoppedAt = s.StoppedAt.UTC().Format(time.RFC3339)
		}
		records = append(records, exportRecord{
			Date:            s.Date,
			Project:         s.ProjectName,
			Task:            s.TaskSummary,
			DurationSeconds: s.DurationSeconds,
			StartedAt:       s.StartedAt.UTC().Format(time.RFC3339),
			StoppedAt:       stoppedAt,
		})
	}

	switch *format {
	case "json":
		return writeJSON(records)
	default:
		return writeCSV(records)
	}
}

func writeCSV(records []exportRecord) int {
	w := csv.NewWriter(os.Stdout)
	header := []string{"date", "project", "task", "duration_seconds", "started_at", "stopped_at"}
	if err := w.Write(header); err != nil {
		fmt.Fprintf(os.Stderr, "export: write csv header: %v\n", err)
		return 1
	}
	for _, r := range records {
		row := []string{
			r.Date,
			r.Project,
			r.Task,
			strconv.Itoa(r.DurationSeconds),
			r.StartedAt,
			r.StoppedAt,
		}
		if err := w.Write(row); err != nil {
			fmt.Fprintf(os.Stderr, "export: write csv row: %v\n", err)
			return 1
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "export: flush csv: %v\n", err)
		return 1
	}
	return 0
}

func writeJSON(records []exportRecord) int {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if records == nil {
		records = []exportRecord{}
	}
	if err := enc.Encode(records); err != nil {
		fmt.Fprintf(os.Stderr, "export: write json: %v\n", err)
		return 1
	}
	return 0
}
