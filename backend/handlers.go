package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// Server 服务器和依赖
type Server struct {
	csv      *SQLiteHandler
	analyzer *ProgressAnalyzer
}

func NewServer(dataDir string) (*Server, error) {
	csv, err := NewSQLiteHandler(dataDir)
	if err != nil {
		return nil, err
	}
	return &Server{
		csv:      csv,
		analyzer: NewProgressAnalyzer(csv),
	}, nil
}

// ========== 动作管理 ==========

func (s *Server) handleExercises(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case http.MethodGet:
		exercises, err := s.csv.LoadExercises()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(exercises)

	case http.MethodPost:
		var exercise Exercise
		if err := json.NewDecoder(r.Body).Decode(&exercise); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 生成新ID
		exercises, _ := s.csv.LoadExercises()
		maxID := 0
		for _, e := range exercises {
			if e.ID > maxID {
				maxID = e.ID
			}
		}
		exercise.ID = maxID + 1

		exercises = append(exercises, exercise)
		if err := s.csv.SaveExercises(exercises); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(exercise)

	case http.MethodPut:
		// 更新动作
		var exercise Exercise
		if err := json.NewDecoder(r.Body).Decode(&exercise); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if exercise.ID <= 0 {
			http.Error(w, "Invalid exercise ID", http.StatusBadRequest)
			return
		}

		exercises, _ := s.csv.LoadExercises()
		found := false
		for i, e := range exercises {
			if e.ID == exercise.ID {
				exercises[i] = exercise
				found = true
				break
			}
		}

		if !found {
			http.Error(w, "Exercise not found", http.StatusNotFound)
			return
		}

		if err := s.csv.SaveExercises(exercises); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(exercise)

	case http.MethodDelete:
		// 从URL获取要删除的exercise ID
		idStr := r.URL.Path[len("/api/exercises/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid exercise ID", http.StatusBadRequest)
			return
		}

		exercises, _ := s.csv.LoadExercises()
		var updatedExercises []Exercise
		found := false
		for _, e := range exercises {
			if e.ID != id {
				updatedExercises = append(updatedExercises, e)
			} else {
				found = true
			}
		}

		if !found {
			http.Error(w, "Exercise not found", http.StatusNotFound)
			return
		}

		if err := s.csv.SaveExercises(updatedExercises); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		groups, _ := s.csv.LoadExerciseGroups()
		groupsChanged := false
		for i, group := range groups {
			var updatedIDs []int
			for _, exerciseID := range group.ExerciseIDs {
				if exerciseID != id {
					updatedIDs = append(updatedIDs, exerciseID)
				} else {
					groupsChanged = true
				}
			}
			groups[i].ExerciseIDs = updatedIDs
		}
		if groupsChanged {
			if err := s.csv.SaveExerciseGroups(groups); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Exercise deleted successfully",
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ========== 动作组管理 ==========

func (s *Server) handleGroups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case http.MethodGet:
		groups, err := s.csv.LoadExerciseGroups()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 加载exercise详情
		exercises, _ := s.csv.LoadExercises()
		exerciseMap := make(map[int]Exercise)
		for _, e := range exercises {
			exerciseMap[e.ID] = e
		}

		// 为每个group填充exercise详情
		type GroupWithDetails struct {
			ID          int        `json:"id"`
			Name        string     `json:"name"`
			ExerciseIDs []int      `json:"exerciseIds"`
			Exercises   []Exercise `json:"exercises"`
		}

		var result []GroupWithDetails
		for _, g := range groups {
			var exList []Exercise
			for _, eid := range g.ExerciseIDs {
				if e, ok := exerciseMap[eid]; ok {
					exList = append(exList, e)
				}
			}
			result = append(result, GroupWithDetails{
				ID:          g.ID,
				Name:        g.Name,
				ExerciseIDs: g.ExerciseIDs,
				Exercises:   exList,
			})
		}

		json.NewEncoder(w).Encode(result)

	case http.MethodPost:
		var group ExerciseGroup
		if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		groups, _ := s.csv.LoadExerciseGroups()
		maxID := 0
		for _, g := range groups {
			if g.ID > maxID {
				maxID = g.ID
			}
		}
		group.ID = maxID + 1

		groups = append(groups, group)
		if err := s.csv.SaveExerciseGroups(groups); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(group)

	case http.MethodPut:
		var group ExerciseGroup
		if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if group.ID <= 0 {
			http.Error(w, "Invalid group ID", http.StatusBadRequest)
			return
		}

		groups, _ := s.csv.LoadExerciseGroups()
		found := false
		for i, g := range groups {
			if g.ID == group.ID {
				groups[i] = group
				found = true
				break
			}
		}

		if !found {
			http.Error(w, "Group not found", http.StatusNotFound)
			return
		}

		if err := s.csv.SaveExerciseGroups(groups); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(group)

	case http.MethodDelete:
		idStr := strings.TrimPrefix(r.URL.Path, "/api/groups/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid group ID", http.StatusBadRequest)
			return
		}

		groups, _ := s.csv.LoadExerciseGroups()
		var updatedGroups []ExerciseGroup
		found := false
		for _, g := range groups {
			if g.ID != id {
				updatedGroups = append(updatedGroups, g)
			} else {
				found = true
			}
		}

		if !found {
			http.Error(w, "Group not found", http.StatusNotFound)
			return
		}

		if err := s.csv.SaveExerciseGroups(updatedGroups); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Group deleted successfully",
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ========== 获取上次训练记录 ==========

func (s *Server) handleGroupLastRecord(w http.ResponseWriter, r *http.Request) {
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

	// 从URL提取group ID
	path := strings.TrimPrefix(r.URL.Path, "/api/groups/")
	path = strings.TrimSuffix(path, "/last-record")
	groupID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	lastRecord, err := s.analyzer.GetLastRecordForGroup(groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(lastRecord)
}

// ========== 提交训练会话 ==========

func (s *Server) handleSessionSubmit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var sessionData SessionData
	if err := json.NewDecoder(r.Body).Decode(&sessionData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 加载现有数据
	sessions, _ := s.csv.LoadTrainingSessions()
	records, _ := s.csv.LoadTrainingRecords()
	sessionDateByID := make(map[int]string)
	for _, session := range sessions {
		sessionDateByID[session.SessionID] = session.Date
	}

	resetExerciseIDs := make(map[int]bool)
	for _, exRecord := range sessionData.ExerciseRecords {
		resetExerciseIDs[exRecord.ExerciseID] = true
	}

	filteredRecords := make([]TrainingRecord, 0, len(records))
	for _, record := range records {
		if sessionDateByID[record.SessionID] == sessionData.Date && resetExerciseIDs[record.ExerciseID] {
			continue
		}
		filteredRecords = append(filteredRecords, record)
	}
	records = filteredRecords

	sessionHasRecords := make(map[int]bool)
	for _, record := range records {
		sessionHasRecords[record.SessionID] = true
	}
	filteredSessions := make([]TrainingSession, 0, len(sessions))
	for _, session := range sessions {
		if session.Date == sessionData.Date && !sessionHasRecords[session.SessionID] {
			continue
		}
		filteredSessions = append(filteredSessions, session)
	}
	sessions = filteredSessions

	newSetCount := 0
	for _, exRecord := range sessionData.ExerciseRecords {
		newSetCount += len(exRecord.Sets)
	}
	if newSetCount == 0 {
		if err := s.csv.SaveTrainingData(sessions, records); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Training records cleared successfully",
		})
		return
	}

	// 生成新session ID
	maxSessionID := 0
	for _, s := range sessions {
		if s.SessionID > maxSessionID {
			maxSessionID = s.SessionID
		}
	}
	newSessionID := maxSessionID + 1

	durationMinutes := sessionData.DurationMinutes
	if durationMinutes <= 0 {
		durationMinutes = 40
	}

	// 创建session
	session := TrainingSession{
		SessionID:       newSessionID,
		GroupID:         sessionData.GroupID,
		Date:            sessionData.Date,
		Status:          "completed",
		DurationMinutes: durationMinutes,
	}
	sessions = append(sessions, session)

	// 生成record ID并保存records
	maxRecordID := 0
	for _, r := range records {
		if r.RecordID > maxRecordID {
			maxRecordID = r.RecordID
		}
	}

	for _, exRecord := range sessionData.ExerciseRecords {
		for _, set := range exRecord.Sets {
			maxRecordID++
			set.RecordID = maxRecordID
			set.SessionID = newSessionID
			set.ExerciseID = exRecord.ExerciseID
			records = append(records, set)
		}
	}

	// 保存
	if err := s.csv.SaveTrainingData(sessions, records); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessionId": newSessionID,
		"message":   "Training session saved successfully",
	})
}

// ========== 查询某天某动作组训练记录 ==========

func (s *Server) handleSessionRecords(w http.ResponseWriter, r *http.Request) {
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

	date := r.URL.Query().Get("date")
	groupID, err := strconv.Atoi(r.URL.Query().Get("groupId"))
	if date == "" || err != nil || groupID <= 0 {
		http.Error(w, "date and groupId are required", http.StatusBadRequest)
		return
	}

	sessions, _ := s.csv.LoadTrainingSessions()
	records, _ := s.csv.LoadTrainingRecords()

	matchedSessionIDs := make(map[int]bool)
	response := SessionRecordsResponse{
		GroupID:         groupID,
		Date:            date,
		DurationMinutes: 40,
		ExerciseRecords: make(map[int][]TrainingRecord),
	}

	for _, session := range sessions {
		if session.Date != date || session.GroupID != groupID {
			continue
		}
		matchedSessionIDs[session.SessionID] = true
		if session.SessionID >= response.SessionID {
			response.SessionID = session.SessionID
			response.DurationMinutes = session.DurationMinutes
		}
	}

	for _, record := range records {
		if !matchedSessionIDs[record.SessionID] {
			continue
		}
		response.ExerciseRecords[record.ExerciseID] = append(response.ExerciseRecords[record.ExerciseID], record)
	}

	for exerciseID := range response.ExerciseRecords {
		sort.Slice(response.ExerciseRecords[exerciseID], func(i, j int) bool {
			return response.ExerciseRecords[exerciseID][i].SetNumber < response.ExerciseRecords[exerciseID][j].SetNumber
		})
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleSessionDayRecords(w http.ResponseWriter, r *http.Request) {
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

	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, "date is required", http.StatusBadRequest)
		return
	}

	sessions, _ := s.csv.LoadTrainingSessions()
	records, _ := s.csv.LoadTrainingRecords()
	exercises, _ := s.csv.LoadExercises()
	groups, _ := s.csv.LoadExerciseGroups()

	exerciseMap := make(map[int]Exercise)
	for _, exercise := range exercises {
		exerciseMap[exercise.ID] = exercise
	}

	groupMap := make(map[int]string)
	for _, group := range groups {
		groupMap[group.ID] = group.Name
	}

	recordsBySession := make(map[int][]TrainingRecord)
	for _, record := range records {
		recordsBySession[record.SessionID] = append(recordsBySession[record.SessionID], record)
	}

	response := DaySessionRecordsResponse{Date: date}
	for _, session := range sessions {
		if session.Date != date {
			continue
		}

		sessionRecords := recordsBySession[session.SessionID]
		if len(sessionRecords) == 0 {
			continue
		}

		exerciseRecords := make(map[int][]TrainingRecord)
		exerciseSeen := make(map[int]bool)
		exerciseList := []Exercise{}
		for _, record := range sessionRecords {
			exerciseRecords[record.ExerciseID] = append(exerciseRecords[record.ExerciseID], record)
			if exerciseSeen[record.ExerciseID] {
				continue
			}
			exerciseSeen[record.ExerciseID] = true
			exercise, ok := exerciseMap[record.ExerciseID]
			if !ok {
				exercise = Exercise{
					ID:          record.ExerciseID,
					Name:        "动作" + strconv.Itoa(record.ExerciseID),
					MuscleGroup: "未知",
					Unit:        "kg",
				}
			}
			exerciseList = append(exerciseList, exercise)
		}

		for exerciseID := range exerciseRecords {
			sort.Slice(exerciseRecords[exerciseID], func(i, j int) bool {
				return exerciseRecords[exerciseID][i].SetNumber < exerciseRecords[exerciseID][j].SetNumber
			})
		}

		groupName := groupMap[session.GroupID]
		if groupName == "" {
			groupName = "动作组" + strconv.Itoa(session.GroupID)
		}

		response.Sessions = append(response.Sessions, SessionRecordResponse{
			SessionID:       session.SessionID,
			GroupID:         session.GroupID,
			GroupName:       groupName,
			Date:            session.Date,
			DurationMinutes: session.DurationMinutes,
			Exercises:       exerciseList,
			ExerciseRecords: exerciseRecords,
		})
	}

	json.NewEncoder(w).Encode(response)
}

// ========== 进度查询 ==========

func (s *Server) handleExerciseProgress(w http.ResponseWriter, r *http.Request) {
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

	// 从URL提取exercise ID
	path := strings.TrimPrefix(r.URL.Path, "/api/progress/exercise/")
	exerciseID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid exercise ID", http.StatusBadRequest)
		return
	}

	progress, err := s.analyzer.GetExerciseProgress(exerciseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(progress)
}

func (s *Server) handleMuscleProgress(w http.ResponseWriter, r *http.Request) {
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

	// 从URL提取muscle group
	path := strings.TrimPrefix(r.URL.Path, "/api/progress/muscle/")
	muscleGroup := path

	progress, err := s.analyzer.GetMuscleProgress(muscleGroup)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(progress)
}
