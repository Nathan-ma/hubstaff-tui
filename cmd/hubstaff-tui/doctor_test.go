package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Nathan-ma/hubstaff-tui/internal/config"
	"github.com/Nathan-ma/hubstaff-tui/internal/store"
)

// --- checkCLIBinary ---

func TestCheckCLIBinary_MissingBinary(t *testing.T) {
	output := captureStdout(t, func() {
		pass := checkCLIBinary("/nonexistent/path/to/HubstaffCLI")
		if pass {
			t.Error("expected checkCLIBinary to return false for missing binary")
		}
	})

	if !strings.Contains(output, "[FAIL]") {
		t.Errorf("expected [FAIL] in output, got: %s", output)
	}
}

func TestCheckCLIBinary_DirectoryPath(t *testing.T) {
	dir := t.TempDir()
	output := captureStdout(t, func() {
		pass := checkCLIBinary(dir)
		if pass {
			t.Error("expected checkCLIBinary to return false for directory path")
		}
	})

	if !strings.Contains(output, "[FAIL]") {
		t.Errorf("expected [FAIL] in output, got: %s", output)
	}
}

func TestCheckCLIBinary_NonExecutableFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "notexec")
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho test"), 0o644); err != nil {
		t.Fatal(err)
	}

	output := captureStdout(t, func() {
		pass := checkCLIBinary(path)
		if pass {
			t.Error("expected checkCLIBinary to return false for non-executable file")
		}
	})

	if !strings.Contains(output, "[FAIL]") {
		t.Errorf("expected [FAIL] in output, got: %s", output)
	}
}

func TestCheckCLIBinary_ValidBinary(t *testing.T) {
	// Write a small wrapper script that echoes a version string.
	dir := t.TempDir()
	path := filepath.Join(dir, "FakeCLI")
	script := "#!/bin/sh\necho 'FakeCLI v9.9.9'\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	output := captureStdout(t, func() {
		pass := checkCLIBinary(path)
		if !pass {
			t.Error("expected checkCLIBinary to return true for valid executable")
		}
	})

	if !strings.Contains(output, "[PASS]") {
		t.Errorf("expected [PASS] in output, got: %s", output)
	}
	if !strings.Contains(output, "v9.9.9") {
		t.Errorf("expected version string in output, got: %s", output)
	}
}

// --- checkConfigFile ---

func TestCheckConfigFile_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does_not_exist.toml")

	output := captureStdout(t, func() {
		checkConfigFile(path)
	})

	if !strings.Contains(output, "[WARN]") {
		t.Errorf("expected [WARN] for missing config, got: %s", output)
	}
}

func TestCheckConfigFile_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("[hubstaff]\ncli_path = \"/some/path\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	output := captureStdout(t, func() {
		checkConfigFile(path)
	})

	if !strings.Contains(output, "[PASS]") {
		t.Errorf("expected [PASS] for existing config, got: %s", output)
	}
	if !strings.Contains(output, path) {
		t.Errorf("expected config path in output, got: %s", output)
	}
}

// --- checkDatabase ---

func TestCheckDatabase_MissingDB(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Store.DBPath = filepath.Join(t.TempDir(), "test.db")

	output := captureStdout(t, func() {
		pass := checkDatabase(&cfg)
		if !pass {
			t.Error("expected checkDatabase to return true when DB is absent (will be created on first run)")
		}
	})

	if !strings.Contains(output, "[WARN]") {
		t.Errorf("expected [WARN] for absent DB, got: %s", output)
	}
}

func TestCheckDatabase_ExistingDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Pre-create the DB via the store package.
	st, err := store.Open(dbPath, 300*time.Second)
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	_ = st.Close()

	cfg := config.DefaultConfig()
	cfg.Store.DBPath = dbPath
	cfg.Store.TTLSeconds = 300

	output := captureStdout(t, func() {
		pass := checkDatabase(&cfg)
		if !pass {
			t.Error("expected pass=true for existing accessible DB")
		}
	})

	if !strings.Contains(output, "[PASS]") {
		t.Errorf("expected [PASS] for existing DB, got: %s", output)
	}
	if !strings.Contains(output, dbPath) {
		t.Errorf("expected DB path in output, got: %s", output)
	}
}

// --- runDoctor exit codes ---

func TestRunDoctor_ExitCodeOneWhenCLIMissing(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Hubstaff.CLIPath = "/nonexistent/HubstaffCLI"
	cfg.Store.DBPath = filepath.Join(t.TempDir(), "test.db")

	captureStdout(t, func() {
		code := runDoctor(&cfg, "")
		if code != 1 {
			t.Errorf("expected exit code 1, got %d", code)
		}
	})
}

func TestRunDoctor_OutputContainsAllCheckLabels(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Hubstaff.CLIPath = "/nonexistent/HubstaffCLI"
	cfg.Store.DBPath = filepath.Join(t.TempDir(), "test.db")

	output := captureStdout(t, func() {
		_ = runDoctor(&cfg, "")
	})

	for _, want := range []string{"HubstaffCLI", "Config", "Database", "Active session"} {
		if !strings.Contains(output, want) {
			t.Errorf("expected %q in runDoctor output, got:\n%s", want, output)
		}
	}
}

func TestRunDoctor_ExitCodeZeroWhenAllPass(t *testing.T) {
	// Write a fake CLI script that responds to --version.
	dir := t.TempDir()
	cliBin := filepath.Join(dir, "FakeCLI")
	if err := os.WriteFile(cliBin, []byte("#!/bin/sh\necho 'FakeCLI v1.0.0'\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a minimal config file.
	cfgFile := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(cfgFile, []byte("[hubstaff]\ncli_path = \""+cliBin+"\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.Hubstaff.CLIPath = cliBin
	cfg.Store.DBPath = filepath.Join(dir, "test.db")
	cfg.Store.TTLSeconds = 300

	captureStdout(t, func() {
		code := runDoctor(&cfg, cfgFile)
		if code != 0 {
			t.Errorf("expected exit code 0 when CLI and DB checks pass, got %d", code)
		}
	})
}
