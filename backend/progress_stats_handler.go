package main

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleStatsProgressSummary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	days, exerciseID, muscleGroup, ok := parseStatsFilterParams(w, r)
	if !ok {
		return
	}
	stats, err := NewStatsAnalyzer(s.csv).GetProgressSummary(days, exerciseID, muscleGroup)
	if err != nil {
		http.Error(w, "calculate progress summary failed", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "encode progress summary failed", http.StatusInternalServerError)
	}
}

func (s *Server) handleStatsProgressDay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	date := r.URL.Query().Get("date")
	if !validProgressDate(date) {
		http.Error(w, "invalid date", http.StatusBadRequest)
		return
	}
	_, exerciseID, muscleGroup, ok := parseStatsFilterParams(w, r)
	if !ok {
		return
	}
	stats, err := NewStatsAnalyzer(s.csv).GetDayProgress(date, exerciseID, muscleGroup)
	if err != nil {
		http.Error(w, "calculate day progress failed", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "encode day progress failed", http.StatusInternalServerError)
	}
}
