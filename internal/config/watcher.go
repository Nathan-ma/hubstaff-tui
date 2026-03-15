package config

import (
	"os"
	"time"
)

// Watcher polls a config file for changes based on modification time.
type Watcher struct {
	path    string
	lastMod time.Time
}

// NewWatcher creates a watcher for the given config file path.
func NewWatcher(path string) *Watcher {
	w := &Watcher{path: path}
	w.lastMod = w.modTime()
	return w
}

// Changed returns true if the file has been modified since the last check.
func (w *Watcher) Changed() bool {
	mod := w.modTime()
	if mod.After(w.lastMod) {
		w.lastMod = mod
		return true
	}
	return false
}

func (w *Watcher) modTime() time.Time {
	info, err := os.Stat(w.path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}
