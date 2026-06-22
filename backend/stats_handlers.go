package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ========== 统计API处理函数 ==========

func (s *Server) handleStatsVolume(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取天数参数
	daysStr := r.URL.Query().Get("days")
	days := 30 // 默认30天
	if daysStr != "" {
		d, err := strconv.Atoi(daysStr)
		if err == nil {
			days = d
		}
	}

	percentile := parseVolumeGrowthPercentile(r)

	analyzer := NewStatsAnalyzer(s.csv)
	stats, err := analyzer.GetVolumeStats(days, percentile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleStatsIntensity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取参数
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		d, err := strconv.Atoi(daysStr)
		if err == nil {
			days = d
		}
	}

	// 获取当前体重
	weightHandler := NewWeightRecordsHandler(s.csv)
	currentWeight, _ := weightHandler.GetLatestWeight()

	analyzer := NewStatsAnalyzer(s.csv)
	stats, err := analyzer.GetIntensityStats(days, currentWeight)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleStatsProgressRate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从URL获取exercise ID
	path := strings.TrimPrefix(r.URL.Path, "/api/stats/progress-rate/")
	exerciseID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid exercise ID", http.StatusBadRequest)
		return
	}

	// 获取目标重量
	targetWeight := 0.0
	targetStr := r.URL.Query().Get("target")
	if targetStr != "" {
		targetWeight, _ = strconv.ParseFloat(targetStr, 64)
	}

	analyzer := NewStatsAnalyzer(s.csv)
	stats, err := analyzer.GetProgressRate(exerciseID, targetWeight)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleStatsPersonalRecords(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	analyzer := NewStatsAnalyzer(s.csv)
	records, err := analyzer.GetPersonalRecords()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(records)
}

func (s *Server) handleStatsFrequency(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	analyzer := NewStatsAnalyzer(s.csv)
	stats, err := analyzer.GetTrainingFrequency()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleStatsComprehensive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取天数参数
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		d, err := strconv.Atoi(daysStr)
		if err == nil {
			days = d
		}
	}

	percentile := parseVolumeGrowthPercentile(r)

	analyzer := NewStatsAnalyzer(s.csv)
	stats, err := analyzer.GetComprehensiveStats(days, percentile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleStatsReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取参数
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		d, err := strconv.Atoi(daysStr)
		if err == nil {
			days = d
		}
	}

	percentile := parseVolumeGrowthPercentile(r)

	analyzer := NewStatsAnalyzer(s.csv)

	// 构建完整报告
	report := &StatsReport{
		StartDate:   calculateStartDate(days),
		EndDate:     getCurrentDate(),
		GeneratedAt: getCurrentTimestamp(),
	}

	// 获取各项统计
	volumeStats, _ := analyzer.GetVolumeStats(days, percentile)
	report.VolumeStats = volumeStats

	intensityStats, _ := analyzer.GetIntensityStats(days, 0)
	report.IntensityStats = intensityStats

	personalRecords, _ := analyzer.GetPersonalRecords()
	report.PersonalRecords = personalRecords

	frequency, _ := analyzer.GetTrainingFrequency()
	report.TrainingFrequency = frequency

	comprehensive, _ := analyzer.GetComprehensiveStats(days, percentile)
	report.Comprehensive = comprehensive

	// 获取体重记录
	weightHandler := NewWeightRecordsHandler(s.csv)
	weightRecords, _ := weightHandler.LoadWeightRecords()
	report.WeightRecords = weightRecords

	json.NewEncoder(w).Encode(report)
}

func parseVolumeGrowthPercentile(r *http.Request) float64 {
	value := r.URL.Query().Get("growthPercentile")
	if value == "" {
		value = r.URL.Query().Get("percentile")
	}
	if value == "" {
		return defaultVolumeGrowthPercentile
	}

	percentile, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultVolumeGrowthPercentile
	}
	return normalizePercentile(percentile)
}

// ========== 体重记录API ==========

func (s *Server) handleWeightRecords(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case http.MethodGet:
		weightHandler := NewWeightRecordsHandler(s.csv)
		records, err := weightHandler.LoadWeightRecords()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(records)

	case http.MethodPost:
		var record WeightRecord
		if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		weightHandler := NewWeightRecordsHandler(s.csv)
		if err := weightHandler.AddWeightRecord(&record); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(record)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleWeightLatest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	weightHandler := NewWeightRecordsHandler(s.csv)
	weight, err := weightHandler.GetLatestWeight()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]float64{"weight": weight})
}

func (s *Server) handleStatsExerciseDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从URL获取exercise ID
	path := strings.TrimPrefix(r.URL.Path, "/api/stats/exercise/")
	exerciseID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid exercise ID", http.StatusBadRequest)
		return
	}

	// 获取天数参数
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		d, err := strconv.Atoi(daysStr)
		if err == nil {
			days = d
		}
	}

	analyzer := NewStatsAnalyzer(s.csv)
	stats, err := analyzer.GetExerciseHistory(exerciseID, days)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleStatsOverviewHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取天数参数
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		d, err := strconv.Atoi(daysStr)
		if err == nil {
			days = d
		}
	}

	analyzer := NewStatsAnalyzer(s.csv)
	stats, err := analyzer.GetOverviewHistory(days)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleStatsCalendar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	days := 30
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil {
			days = parsedDays
		}
	}

	filterType := r.URL.Query().Get("type")
	target := r.URL.Query().Get("target")
	exerciseID := 0
	muscleGroup := ""

	switch filterType {
	case "exercise":
		parsedID, err := strconv.Atoi(target)
		if err != nil || parsedID <= 0 {
			http.Error(w, "Invalid exercise target", http.StatusBadRequest)
			return
		}
		exerciseID = parsedID
	case "muscle":
		if target == "" {
			http.Error(w, "Invalid muscle target", http.StatusBadRequest)
			return
		}
		muscleGroup = target
	case "", "overview":
		// No additional filter.
	default:
		http.Error(w, "Invalid calendar type", http.StatusBadRequest)
		return
	}

	analyzer := NewStatsAnalyzer(s.csv)
	stats, err := analyzer.GetTrainingCalendar(days, exerciseID, muscleGroup)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleStatsFiltered(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	days, exerciseID, muscleGroup, ok := parseStatsFilterParams(w, r)
	if !ok {
		return
	}
	percentile := parseVolumeGrowthPercentile(r)

	analyzer := NewStatsAnalyzer(s.csv)
	stats, err := analyzer.GetFilteredStats(days, exerciseID, muscleGroup, percentile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleStatsDayRecords(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, "Missing date", http.StatusBadRequest)
		return
	}

	analyzer := NewStatsAnalyzer(s.csv)
	if r.Method == http.MethodDelete {
		deletedSessions, deletedRecords, err := analyzer.DeleteDayRecords(date)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"date":            date,
			"deletedSessions": deletedSessions,
			"deletedRecords":  deletedRecords,
		})
		return
	}

	_, exerciseID, muscleGroup, ok := parseStatsFilterParams(w, r)
	if !ok {
		return
	}

	stats, err := analyzer.GetDayRecords(date, exerciseID, muscleGroup)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func parseStatsFilterParams(w http.ResponseWriter, r *http.Request) (int, int, string, bool) {
	days := 30
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil {
			days = parsedDays
		}
	}

	filterType := r.URL.Query().Get("type")
	target := r.URL.Query().Get("target")
	exerciseID := 0
	muscleGroup := ""

	switch filterType {
	case "exercise":
		parsedID, err := strconv.Atoi(target)
		if err != nil || parsedID <= 0 {
			http.Error(w, "Invalid exercise target", http.StatusBadRequest)
			return 0, 0, "", false
		}
		exerciseID = parsedID
	case "muscle":
		if target == "" {
			http.Error(w, "Invalid muscle target", http.StatusBadRequest)
			return 0, 0, "", false
		}
		muscleGroup = target
	case "", "overview":
		// No additional filter.
	default:
		http.Error(w, "Invalid stats type", http.StatusBadRequest)
		return 0, 0, "", false
	}

	return days, exerciseID, muscleGroup, true
}

// ========== 辅助函数 ==========

func calculateStartDate(days int) string {
	if days <= 0 {
		days = 30
	}
	return time.Now().AddDate(0, 0, -days+1).Format("2006-01-02")
}

func getCurrentDate() string {
	return time.Now().Format("2006-01-02")
}

func getCurrentTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
