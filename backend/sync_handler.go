package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// handleDatabaseSync returns a consistent SQLite snapshot for pulling cloud
// data to another machine.
func (s *Server) handleDatabaseSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path, err := s.csv.CreateSnapshot()
	if err != nil {
		http.Error(w, "create database snapshot failed", http.StatusInternalServerError)
		return
	}
	defer os.Remove(path)

	f, err := os.Open(path)
	if err != nil {
		http.Error(w, "open database snapshot failed", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "application/vnd.sqlite3")
	w.Header().Set("Content-Disposition", `attachment; filename="fitness.db"`)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, "fitness.db", time.Now(), f)
}

// CreateSnapshot uses SQLite's online VACUUM INTO operation so the downloaded
// database is complete even while the source database is using WAL mode.
func (h *SQLiteHandler) CreateSnapshot() (string, error) {
	f, err := os.CreateTemp("", "fitness-snapshot-*.db")
	if err != nil {
		return "", err
	}
	path := f.Name()
	if err := f.Close(); err != nil {
		os.Remove(path)
		return "", err
	}
	// VACUUM INTO requires the destination not to exist.
	if err := os.Remove(path); err != nil {
		return "", err
	}

	if _, err := h.db.Exec(`VACUUM INTO ?`, path); err != nil {
		os.Remove(path)
		return "", fmt.Errorf("vacuum into %s: %w", filepath.Base(path), err)
	}
	return path, nil
}
