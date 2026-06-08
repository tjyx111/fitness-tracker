package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CSVHandler 处理CSV文件读写
type CSVHandler struct {
	dataDir string
}

func NewCSVHandler(dataDir string) *CSVHandler {
	return &CSVHandler{dataDir: dataDir}
}

// 确保数据目录存在
func (h *CSVHandler) ensureDir() error {
	return os.MkdirAll(h.dataDir, 0755)
}

// 获取文件路径
func (h *CSVHandler) getPath(filename string) string {
	return fmt.Sprintf("%s/%s", h.dataDir, filename)
}

// ========== Exercises ==========

func (h *CSVHandler) LoadExercises() ([]Exercise, error) {
	h.ensureDir()
	file, err := os.Open(h.getPath("exercises.csv"))
	if err != nil {
		if os.IsNotExist(err) {
			return []Exercise{}, nil
		}
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var exercises []Exercise
	for i, record := range records {
		if i == 0 && (len(record) == 0 || strings.HasPrefix(record[0], "id")) {
			continue // 跳过header
		}
		if len(record) < 3 {
			continue
		}
		id, _ := strconv.Atoi(record[0])

		// 处理 unit 字段，默认为 "kg"
		unit := "kg"
		if len(record) >= 4 && record[3] != "" {
			unit = record[3]
		}

		exercises = append(exercises, Exercise{
			ID:          id,
			Name:        record[1],
			MuscleGroup: record[2],
			Unit:        unit,
		})
	}
	return exercises, nil
}

func (h *CSVHandler) SaveExercises(exercises []Exercise) error {
	h.ensureDir()
	file, err := os.Create(h.getPath("exercises.csv"))
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写header
	writer.Write([]string{"id(编号)", "name(动作名称)", "muscleGroup(肌肉部位)", "unit(单位类型)"})

	for _, e := range exercises {
		writer.Write([]string{
			strconv.Itoa(e.ID),
			e.Name,
			e.MuscleGroup,
			e.Unit,
		})
	}
	return nil
}

// ========== ExerciseGroups ==========

func (h *CSVHandler) LoadExerciseGroups() ([]ExerciseGroup, error) {
	h.ensureDir()
	file, err := os.Open(h.getPath("exercise_groups.csv"))
	if err != nil {
		if os.IsNotExist(err) {
			return []ExerciseGroup{}, nil
		}
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var groups []ExerciseGroup
	for i, record := range records {
		if i == 0 && (len(record) == 0 || strings.HasPrefix(record[0], "id")) {
			continue
		}
		if len(record) < 3 {
			continue
		}
		id, _ := strconv.Atoi(record[0])

		// 解析exerciseIds "1,2,3"
		var exerciseIDs []int
		if record[2] != "" {
			for _, idStr := range strings.Split(record[2], ",") {
				if idInt, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
					exerciseIDs = append(exerciseIDs, idInt)
				}
			}
		}

		groups = append(groups, ExerciseGroup{
			ID:          id,
			Name:        record[1],
			ExerciseIDs: exerciseIDs,
		})
	}
	return groups, nil
}

func (h *CSVHandler) SaveExerciseGroups(groups []ExerciseGroup) error {
	h.ensureDir()
	file, err := os.Create(h.getPath("exercise_groups.csv"))
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"id(编号)", "name(动作组名称)", "exerciseIds(包含动作ID列表)"})

	for _, g := range groups {
		var idStrs []string
		for _, id := range g.ExerciseIDs {
			idStrs = append(idStrs, strconv.Itoa(id))
		}
		writer.Write([]string{
			strconv.Itoa(g.ID),
			g.Name,
			strings.Join(idStrs, ","),
		})
	}
	return nil
}

// ========== TrainingSessions ==========

func (h *CSVHandler) LoadTrainingSessions() ([]TrainingSession, error) {
	h.ensureDir()
	file, err := os.Open(h.getPath("training_sessions.csv"))
	if err != nil {
		if os.IsNotExist(err) {
			return []TrainingSession{}, nil
		}
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var sessions []TrainingSession
	for i, record := range records {
		if i == 0 && (len(record) == 0 || strings.HasPrefix(record[0], "sessionId")) {
			continue
		}
		if len(record) < 4 {
			continue
		}
		sessionID, _ := strconv.Atoi(record[0])
		groupID, _ := strconv.Atoi(record[1])
		durationMinutes := 40
		if len(record) >= 5 && record[4] != "" {
			if parsedDuration, err := strconv.Atoi(record[4]); err == nil && parsedDuration > 0 {
				durationMinutes = parsedDuration
			}
		}

		sessions = append(sessions, TrainingSession{
			SessionID:       sessionID,
			GroupID:         groupID,
			Date:            record[2],
			Status:          record[3],
			DurationMinutes: durationMinutes,
		})
	}
	return sessions, nil
}

func (h *CSVHandler) SaveTrainingSessions(sessions []TrainingSession) error {
	h.ensureDir()
	file, err := os.Create(h.getPath("training_sessions.csv"))
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"sessionId(会话编号)", "groupId(动作组编号)", "date(训练日期)", "status(状态)", "durationMinutes(训练时长分钟)"})

	for _, s := range sessions {
		durationMinutes := s.DurationMinutes
		if durationMinutes <= 0 {
			durationMinutes = 40
		}
		writer.Write([]string{
			strconv.Itoa(s.SessionID),
			strconv.Itoa(s.GroupID),
			s.Date,
			s.Status,
			strconv.Itoa(durationMinutes),
		})
	}
	return nil
}

// ========== TrainingRecords ==========

func (h *CSVHandler) LoadTrainingRecords() ([]TrainingRecord, error) {
	h.ensureDir()
	file, err := os.Open(h.getPath("training_records.csv"))
	if err != nil {
		if os.IsNotExist(err) {
			return []TrainingRecord{}, nil
		}
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var trainingRecords []TrainingRecord
	for i, record := range records {
		if i == 0 && (len(record) == 0 || strings.HasPrefix(record[0], "recordId")) {
			continue
		}
		if len(record) < 7 {
			continue
		}
		recordID, _ := strconv.Atoi(record[0])
		sessionID, _ := strconv.Atoi(record[1])
		exerciseID, _ := strconv.Atoi(record[2])
		setNumber, _ := strconv.Atoi(record[3])
		weight, _ := strconv.ParseFloat(record[4], 64)
		reps, _ := strconv.Atoi(record[5])
		duration, _ := strconv.Atoi(record[6])

		note := ""
		if len(record) > 7 {
			note = record[7]
		}

		trainingRecords = append(trainingRecords, TrainingRecord{
			RecordID:   recordID,
			SessionID:  sessionID,
			ExerciseID: exerciseID,
			SetNumber:  setNumber,
			Weight:     weight,
			Reps:       reps,
			Duration:   duration,
			Note:       note,
		})
	}
	return trainingRecords, nil
}

func (h *CSVHandler) SaveTrainingRecords(records []TrainingRecord) error {
	h.ensureDir()
	file, err := os.Create(h.getPath("training_records.csv"))
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"recordId(记录编号)", "sessionId(会话编号)", "exerciseId(动作编号)", "setNumber(组数)", "weight(重量kg)", "reps(次数)", "duration(持续时间秒)", "note(备注)"})

	for _, r := range records {
		writer.Write([]string{
			strconv.Itoa(r.RecordID),
			strconv.Itoa(r.SessionID),
			strconv.Itoa(r.ExerciseID),
			strconv.Itoa(r.SetNumber),
			fmt.Sprintf("%.2f", r.Weight),
			strconv.Itoa(r.Reps),
			strconv.Itoa(r.Duration),
			r.Note,
		})
	}
	return nil
}

// 获取下一个ID
func (h *CSVHandler) getNextID(filename string) (int, error) {
	file, err := os.Open(h.getPath(filename))
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return 0, err
	}

	maxID := 0
	for i, record := range records {
		if i == 0 && (len(record) == 0 || strings.HasPrefix(record[0], "id") || strings.HasPrefix(record[0], "sessionId") || strings.HasPrefix(record[0], "recordId")) {
			continue
		}
		if len(record) > 0 {
			if id, err := strconv.Atoi(record[0]); err == nil && id > maxID {
				maxID = id
			}
		}
	}
	return maxID + 1, nil
}
