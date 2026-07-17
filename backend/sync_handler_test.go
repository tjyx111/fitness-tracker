package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDatabaseSyncDoesNotRequireAuthorization(t *testing.T) {
	server, err := NewServer(t.TempDir())
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer server.csv.Close()

	if err := server.csv.SaveExercises([]Exercise{
		{ID: 1, Name: "squat", MuscleGroup: "legs", Unit: "kg"},
	}); err != nil {
		t.Fatalf("SaveExercises: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/sync/database", nil)
	rec := httptest.NewRecorder()
	server.handleDatabaseSync(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, body=%q", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/vnd.sqlite3" {
		t.Fatalf("Content-Type=%q, want application/vnd.sqlite3", got)
	}

	dbPath := filepath.Join(t.TempDir(), "fitness.db")
	if err := os.WriteFile(dbPath, rec.Body.Bytes(), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	var check string
	if err := db.QueryRow(`PRAGMA integrity_check;`).Scan(&check); err != nil {
		t.Fatalf("integrity_check: %v", err)
	}
	if check != "ok" {
		t.Fatalf("integrity_check=%q, want ok", check)
	}

	var name string
	if err := db.QueryRow(`SELECT name FROM exercises WHERE id=1`).Scan(&name); err != nil {
		t.Fatalf("query synced exercise: %v", err)
	}
	if name != "squat" {
		t.Fatalf("synced exercise name=%q, want squat", name)
	}
}

func TestDatabaseSyncRejectsNonGet(t *testing.T) {
	server, err := NewServer(t.TempDir())
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer server.csv.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/sync/database", nil)
	rec := httptest.NewRecorder()
	server.handleDatabaseSync(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status=%d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if got := rec.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("Allow=%q, want GET", got)
	}
}
