package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var ErrActiveChallengeExists = errors.New("an active challenge already exists")

type Challenge struct {
	ID           int             `json:"id"`
	Name         string          `json:"name"`
	StartDate    string          `json:"startDate"`
	EndDate      string          `json:"endDate"`
	Status       string          `json:"status"`
	TerminatedAt string          `json:"terminatedAt"`
	Items        []ChallengeItem `json:"items"`
	CreatedAt    string          `json:"createdAt"`
}

type ChallengeItem struct {
	ID          int    `json:"id"`
	ChallengeID int    `json:"challengeId"`
	Title       string `json:"title"`
	Position    int    `json:"position"`
}

type ChallengeDailyItem struct {
	ID              int    `json:"id"`
	ChallengeID     int    `json:"challengeId"`
	ChallengeName   string `json:"challengeName"`
	ChallengeItemID int    `json:"challengeItemId"`
	Title           string `json:"title"`
	ChallengeDate   string `json:"challengeDate"`
	Completed       bool   `json:"completed"`
	CompletedAt     string `json:"completedAt"`
}

type ChallengeDay struct {
	ChallengeID       int                  `json:"challengeId"`
	ChallengeName     string               `json:"challengeName"`
	Status            string               `json:"status"`
	Date              string               `json:"date"`
	Items             []ChallengeDailyItem `json:"items"`
	CompletedItems    int                  `json:"completedItems"`
	TotalItems        int                  `json:"totalItems"`
	CompletionPercent float64              `json:"completionPercent"`
}

type ChallengeSummary struct {
	ID                int     `json:"id"`
	Name              string  `json:"name"`
	StartDate         string  `json:"startDate"`
	EndDate           string  `json:"endDate"`
	Status            string  `json:"status"`
	TerminatedAt      string  `json:"terminatedAt"`
	TotalDays         int     `json:"totalDays"`
	ItemCount         int     `json:"itemCount"`
	TotalItems        int     `json:"totalItems"`
	CompletedItems    int     `json:"completedItems"`
	CompletionPercent float64 `json:"completionPercent"`
}

type ChallengeDetail struct {
	Challenge ChallengeSummary `json:"challenge"`
	Days      []ChallengeDay   `json:"days"`
}

type ChallengeItemStats struct {
	ChallengeItemID   int     `json:"challengeItemId"`
	ChallengeID       int     `json:"challengeId"`
	ChallengeName     string  `json:"challengeName"`
	Title             string  `json:"title"`
	CompletedDays     int     `json:"completedDays"`
	TotalDays         int     `json:"totalDays"`
	CompletionPercent float64 `json:"completionPercent"`
}

type ChallengeDailyStats struct {
	Date              string  `json:"date"`
	CompletedItems    int     `json:"completedItems"`
	TotalItems        int     `json:"totalItems"`
	CompletionPercent float64 `json:"completionPercent"`
}

type ChallengeStats struct {
	TotalItems        int                   `json:"totalItems"`
	CompletedItems    int                   `json:"completedItems"`
	CompletionPercent float64               `json:"completionPercent"`
	ItemStats         []ChallengeItemStats  `json:"itemStats"`
	Daily             []ChallengeDailyStats `json:"daily"`
}

func (h *SQLiteHandler) CreateChallenge(name, startDate string, days int, titles []string) (Challenge, error) {
	if days < 1 || days > 366 {
		return Challenge{}, fmt.Errorf("challenge days must be between 1 and 366")
	}
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return Challenge{}, fmt.Errorf("invalid start date: %w", err)
	}
	if len(titles) == 0 {
		return Challenge{}, errors.New("at least one challenge item is required")
	}

	tx, err := h.db.Begin()
	if err != nil {
		return Challenge{}, err
	}
	defer tx.Rollback()
	if err := finishExpiredChallenges(tx, time.Now().Format("2006-01-02")); err != nil {
		return Challenge{}, err
	}
	var activeCount int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM challenges WHERE status='active'`).Scan(&activeCount); err != nil {
		return Challenge{}, err
	}
	if activeCount > 0 {
		return Challenge{}, ErrActiveChallengeExists
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	endDate := start.AddDate(0, 0, days-1).Format("2006-01-02")
	result, err := tx.Exec(`INSERT INTO challenges(name,start_date,end_date,status,terminated_at,created_at) VALUES(?,?,?,'active','',?)`, name, startDate, endDate, now)
	if err != nil {
		return Challenge{}, err
	}
	challengeID, err := result.LastInsertId()
	if err != nil {
		return Challenge{}, err
	}

	items := make([]ChallengeItem, 0, len(titles))
	for position, title := range titles {
		result, err := tx.Exec(`INSERT INTO challenge_items(challenge_id,title,position) VALUES(?,?,?)`, challengeID, title, position)
		if err != nil {
			return Challenge{}, err
		}
		itemID, err := result.LastInsertId()
		if err != nil {
			return Challenge{}, err
		}
		items = append(items, ChallengeItem{ID: int(itemID), ChallengeID: int(challengeID), Title: title, Position: position})
	}

	for offset := 0; offset < days; offset++ {
		date := start.AddDate(0, 0, offset).Format("2006-01-02")
		for _, item := range items {
			if _, err := tx.Exec(`INSERT INTO challenge_daily_items(challenge_item_id,challenge_date,completed,completed_at,created_at,updated_at) VALUES(?,?,0,'',?,?)`, item.ID, date, now, now); err != nil {
				return Challenge{}, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return Challenge{}, err
	}
	return Challenge{ID: int(challengeID), Name: name, StartDate: startDate, EndDate: endDate, Status: "active", Items: items, CreatedAt: now}, nil
}

func (h *SQLiteHandler) LoadChallengeDay(date string) ([]ChallengeDay, error) {
	return h.loadChallengeDay(date, true)
}

func (h *SQLiteHandler) LoadChallengeHistoryDay(date string) ([]ChallengeDay, error) {
	return h.loadChallengeDay(date, false)
}

func (h *SQLiteHandler) loadChallengeDay(date string, activeOnly bool) ([]ChallengeDay, error) {
	if err := h.finishExpiredChallenges(); err != nil {
		return nil, err
	}
	statusFilter := ""
	if activeOnly {
		statusFilter = " AND c.status='active'"
	}
	rows, err := h.db.Query(`
SELECT d.id,c.id,c.name,c.status,i.id,i.title,d.challenge_date,d.completed,d.completed_at
FROM challenge_daily_items d
JOIN challenge_items i ON i.id=d.challenge_item_id
JOIN challenges c ON c.id=i.challenge_id
WHERE d.challenge_date=?`+statusFilter+`
ORDER BY c.id DESC,i.position ASC,d.id ASC`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byChallenge := map[int]*ChallengeDay{}
	ordered := []int{}
	for rows.Next() {
		var item ChallengeDailyItem
		var completed int
		var status string
		if err := rows.Scan(&item.ID, &item.ChallengeID, &item.ChallengeName, &status, &item.ChallengeItemID, &item.Title, &item.ChallengeDate, &completed, &item.CompletedAt); err != nil {
			return nil, err
		}
		item.Completed = completed == 1
		day := byChallenge[item.ChallengeID]
		if day == nil {
			day = &ChallengeDay{ChallengeID: item.ChallengeID, ChallengeName: item.ChallengeName, Status: status, Date: date, Items: []ChallengeDailyItem{}}
			byChallenge[item.ChallengeID] = day
			ordered = append(ordered, item.ChallengeID)
		}
		day.Items = append(day.Items, item)
		day.TotalItems++
		if item.Completed {
			day.CompletedItems++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]ChallengeDay, 0, len(ordered))
	for _, challengeID := range ordered {
		day := byChallenge[challengeID]
		day.CompletionPercent = challengeCompletionPercent(day.CompletedItems, day.TotalItems)
		result = append(result, *day)
	}
	return result, nil
}

func (h *SQLiteHandler) LoadChallengeHistory() ([]ChallengeSummary, error) {
	if err := h.finishExpiredChallenges(); err != nil {
		return nil, err
	}
	rows, err := h.db.Query(`
SELECT c.id,c.name,c.start_date,c.end_date,c.status,c.terminated_at,
       CAST(julianday(c.end_date)-julianday(c.start_date)+1 AS INTEGER),
       COUNT(DISTINCT i.id),COUNT(d.id),COALESCE(SUM(d.completed),0)
FROM challenges c
LEFT JOIN challenge_items i ON i.challenge_id=c.id
LEFT JOIN challenge_daily_items d ON d.challenge_item_id=i.id
WHERE c.status<>'active'
GROUP BY c.id,c.name,c.start_date,c.end_date,c.status,c.terminated_at
ORDER BY c.start_date DESC,c.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := []ChallengeSummary{}
	for rows.Next() {
		var summary ChallengeSummary
		if err := rows.Scan(
			&summary.ID,
			&summary.Name,
			&summary.StartDate,
			&summary.EndDate,
			&summary.Status,
			&summary.TerminatedAt,
			&summary.TotalDays,
			&summary.ItemCount,
			&summary.TotalItems,
			&summary.CompletedItems,
		); err != nil {
			return nil, err
		}
		summary.CompletionPercent = challengeCompletionPercent(summary.CompletedItems, summary.TotalItems)
		history = append(history, summary)
	}
	return history, rows.Err()
}

func (h *SQLiteHandler) LoadChallengeDetail(id int) (ChallengeDetail, error) {
	if err := h.finishExpiredChallenges(); err != nil {
		return ChallengeDetail{}, err
	}
	var summary ChallengeSummary
	err := h.db.QueryRow(`
SELECT c.id,c.name,c.start_date,c.end_date,c.status,c.terminated_at,
       CAST(julianday(c.end_date)-julianday(c.start_date)+1 AS INTEGER),
       COUNT(DISTINCT i.id),COUNT(d.id),COALESCE(SUM(d.completed),0)
FROM challenges c
LEFT JOIN challenge_items i ON i.challenge_id=c.id
LEFT JOIN challenge_daily_items d ON d.challenge_item_id=i.id
WHERE c.id=?
GROUP BY c.id,c.name,c.start_date,c.end_date,c.status,c.terminated_at`, id).Scan(
		&summary.ID,
		&summary.Name,
		&summary.StartDate,
		&summary.EndDate,
		&summary.Status,
		&summary.TerminatedAt,
		&summary.TotalDays,
		&summary.ItemCount,
		&summary.TotalItems,
		&summary.CompletedItems,
	)
	if err != nil {
		return ChallengeDetail{}, err
	}
	summary.CompletionPercent = challengeCompletionPercent(summary.CompletedItems, summary.TotalItems)

	rows, err := h.db.Query(`
SELECT d.id,c.id,c.name,c.status,i.id,i.title,d.challenge_date,d.completed,d.completed_at
FROM challenge_daily_items d
JOIN challenge_items i ON i.id=d.challenge_item_id
JOIN challenges c ON c.id=i.challenge_id
WHERE c.id=?
ORDER BY d.challenge_date,i.position,d.id`, id)
	if err != nil {
		return ChallengeDetail{}, err
	}
	defer rows.Close()

	byDate := map[string]*ChallengeDay{}
	orderedDates := []string{}
	for rows.Next() {
		var item ChallengeDailyItem
		var status string
		var completed int
		if err := rows.Scan(&item.ID, &item.ChallengeID, &item.ChallengeName, &status, &item.ChallengeItemID, &item.Title, &item.ChallengeDate, &completed, &item.CompletedAt); err != nil {
			return ChallengeDetail{}, err
		}
		item.Completed = completed == 1
		day := byDate[item.ChallengeDate]
		if day == nil {
			day = &ChallengeDay{
				ChallengeID:   item.ChallengeID,
				ChallengeName: item.ChallengeName,
				Status:        status,
				Date:          item.ChallengeDate,
				Items:         []ChallengeDailyItem{},
			}
			byDate[item.ChallengeDate] = day
			orderedDates = append(orderedDates, item.ChallengeDate)
		}
		day.Items = append(day.Items, item)
		day.TotalItems++
		if item.Completed {
			day.CompletedItems++
		}
	}
	if err := rows.Err(); err != nil {
		return ChallengeDetail{}, err
	}

	days := make([]ChallengeDay, 0, len(orderedDates))
	for _, date := range orderedDates {
		day := byDate[date]
		day.CompletionPercent = challengeCompletionPercent(day.CompletedItems, day.TotalItems)
		days = append(days, *day)
	}
	return ChallengeDetail{Challenge: summary, Days: days}, nil
}

func (h *SQLiteHandler) UpdateChallengeDailyItem(id int, completed bool) (ChallengeDailyItem, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	completedValue := 0
	completedAt := ""
	if completed {
		completedValue = 1
		completedAt = now
	}
	result, err := h.db.Exec(`
UPDATE challenge_daily_items
SET completed=?,completed_at=?,updated_at=?
WHERE id=?
  AND EXISTS (
      SELECT 1
      FROM challenge_items i
      JOIN challenges c ON c.id=i.challenge_id
      WHERE i.id=challenge_daily_items.challenge_item_id AND c.status='active'
  )`, completedValue, completedAt, now, id)
	if err != nil {
		return ChallengeDailyItem{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return ChallengeDailyItem{}, err
	}
	if affected == 0 {
		return ChallengeDailyItem{}, sql.ErrNoRows
	}
	return h.loadChallengeDailyItem(id)
}

func (h *SQLiteHandler) TerminateChallenge(id int) error {
	if err := h.finishExpiredChallenges(); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := h.db.Exec(`UPDATE challenges SET status='terminated',terminated_at=? WHERE id=? AND status='active'`, now, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (h *SQLiteHandler) finishExpiredChallenges() error {
	tx, err := h.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := finishExpiredChallenges(tx, time.Now().Format("2006-01-02")); err != nil {
		return err
	}
	return tx.Commit()
}

func finishExpiredChallenges(tx *sql.Tx, today string) error {
	_, err := tx.Exec(`UPDATE challenges SET status='completed' WHERE status='active' AND end_date < ?`, today)
	return err
}

func (h *SQLiteHandler) loadChallengeDailyItem(id int) (ChallengeDailyItem, error) {
	var item ChallengeDailyItem
	var completed int
	err := h.db.QueryRow(`
SELECT d.id,c.id,c.name,i.id,i.title,d.challenge_date,d.completed,d.completed_at
FROM challenge_daily_items d
JOIN challenge_items i ON i.id=d.challenge_item_id
JOIN challenges c ON c.id=i.challenge_id
WHERE d.id=?`, id).Scan(&item.ID, &item.ChallengeID, &item.ChallengeName, &item.ChallengeItemID, &item.Title, &item.ChallengeDate, &completed, &item.CompletedAt)
	item.Completed = completed == 1
	return item, err
}

func (h *SQLiteHandler) DeleteChallenge(id int) error {
	result, err := h.db.Exec(`DELETE FROM challenges WHERE id=?`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (h *SQLiteHandler) GetChallengeStats(days int) (ChallengeStats, error) {
	if days < 1 {
		days = 1
	}
	if err := h.finishExpiredChallenges(); err != nil {
		return ChallengeStats{}, err
	}
	today := time.Now().Format("2006-01-02")
	start := time.Now().AddDate(0, 0, -days+1).Format("2006-01-02")
	stats := ChallengeStats{ItemStats: []ChallengeItemStats{}, Daily: []ChallengeDailyStats{}}

	rows, err := h.db.Query(`
SELECT d.challenge_date,COUNT(*),COALESCE(SUM(d.completed),0)
FROM challenge_daily_items d
WHERE d.challenge_date BETWEEN ? AND ?
GROUP BY d.challenge_date
ORDER BY d.challenge_date`, start, today)
	if err != nil {
		return stats, err
	}
	for rows.Next() {
		var daily ChallengeDailyStats
		if err := rows.Scan(&daily.Date, &daily.TotalItems, &daily.CompletedItems); err != nil {
			rows.Close()
			return stats, err
		}
		daily.CompletionPercent = challengeCompletionPercent(daily.CompletedItems, daily.TotalItems)
		stats.TotalItems += daily.TotalItems
		stats.CompletedItems += daily.CompletedItems
		stats.Daily = append(stats.Daily, daily)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return stats, err
	}
	if err := rows.Close(); err != nil {
		return stats, err
	}
	stats.CompletionPercent = challengeCompletionPercent(stats.CompletedItems, stats.TotalItems)

	rows, err = h.db.Query(`
SELECT i.id,c.id,c.name,i.title,COUNT(*),COALESCE(SUM(d.completed),0)
FROM challenge_items i
JOIN challenges c ON c.id=i.challenge_id
JOIN challenge_daily_items d ON d.challenge_item_id=i.id
WHERE d.challenge_date BETWEEN ? AND ?
GROUP BY i.id,c.id,c.name,i.title,i.position
ORDER BY c.id DESC,i.position ASC`, start, today)
	if err != nil {
		return stats, err
	}
	defer rows.Close()
	for rows.Next() {
		var item ChallengeItemStats
		if err := rows.Scan(&item.ChallengeItemID, &item.ChallengeID, &item.ChallengeName, &item.Title, &item.TotalDays, &item.CompletedDays); err != nil {
			return stats, err
		}
		item.CompletionPercent = challengeCompletionPercent(item.CompletedDays, item.TotalDays)
		stats.ItemStats = append(stats.ItemStats, item)
	}
	return stats, rows.Err()
}

func challengeCompletionPercent(completed, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(completed) * 100 / float64(total)
}

func (s *Server) handleChallenges(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case http.MethodGet:
		date := r.URL.Query().Get("date")
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}
		if _, err := time.Parse("2006-01-02", date); err != nil {
			http.Error(w, "Invalid challenge date", http.StatusBadRequest)
			return
		}
		days, err := s.csv.LoadChallengeDay(date)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(days)

	case http.MethodPost:
		var payload struct {
			Name      string   `json:"name"`
			StartDate string   `json:"startDate"`
			Days      int      `json:"days"`
			Items     []string `json:"items"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if payload.StartDate == "" {
			payload.StartDate = time.Now().Format("2006-01-02")
		}
		name := strings.TrimSpace(payload.Name)
		if name == "" {
			name = "每日挑战"
		}
		items := make([]string, 0, len(payload.Items))
		for _, raw := range payload.Items {
			if item := strings.TrimSpace(raw); item != "" {
				items = append(items, item)
			}
		}
		challenge, err := s.csv.CreateChallenge(name, payload.StartDate, payload.Days, items)
		if err != nil {
			if errors.Is(err, ErrActiveChallengeExists) {
				http.Error(w, "已有正在执行的挑战，请先提前终止", http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(challenge)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleChallenge(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/challenges/"), "/")
	parts := strings.Split(path, "/")
	if path == "history" && r.Method == http.MethodGet {
		history, err := s.csv.LoadChallengeHistory()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(history)
		return
	}
	if path == "history/day" && r.Method == http.MethodGet {
		date := r.URL.Query().Get("date")
		if _, err := time.Parse("2006-01-02", date); err != nil {
			http.Error(w, "Invalid challenge date", http.StatusBadRequest)
			return
		}
		days, err := s.csv.LoadChallengeHistoryDay(date)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(days)
		return
	}
	if len(parts) == 1 && r.Method == http.MethodGet {
		id, err := strconv.Atoi(parts[0])
		if err != nil || id < 1 {
			http.Error(w, "Invalid challenge ID", http.StatusBadRequest)
			return
		}
		detail, err := s.csv.LoadChallengeDetail(id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Challenge not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(detail)
		return
	}
	if len(parts) != 2 || parts[1] != "terminate" || r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil || id < 1 {
		http.Error(w, "Invalid challenge ID", http.StatusBadRequest)
		return
	}
	if err := s.csv.TerminateChallenge(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Challenge not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Challenge terminated"})
}

func (s *Server) handleChallengeDailyItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, err := strconv.Atoi(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/challenge-daily-items/"), "/"))
	if err != nil || id < 1 {
		http.Error(w, "Invalid challenge daily item ID", http.StatusBadRequest)
		return
	}
	var payload struct {
		Completed bool `json:"completed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	item, err := s.csv.UpdateChallengeDailyItem(id, payload.Completed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Challenge daily item not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(item)
}

func (s *Server) handleChallengeStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	days := 30
	if value := r.URL.Query().Get("days"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 366 {
			http.Error(w, "days must be between 1 and 366", http.StatusBadRequest)
			return
		}
		days = parsed
	}
	stats, err := s.csv.GetChallengeStats(days)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
