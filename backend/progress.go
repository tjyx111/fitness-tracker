package main

import (
	"sort"
)

// ProgressAnalyzer 进度分析器
type ProgressAnalyzer struct {
	csv *CSVHandler
}

func NewProgressAnalyzer(csv *CSVHandler) *ProgressAnalyzer {
	return &ProgressAnalyzer{csv: csv}
}

// GetExerciseProgress 获取某个动作的进度数据
func (p *ProgressAnalyzer) GetExerciseProgress(exerciseID int) (*ProgressData, error) {
	// 加载所有数据
	exercises, err := p.csv.LoadExercises()
	if err != nil {
		return nil, err
	}

	records, err := p.csv.LoadTrainingRecords()
	if err != nil {
		return nil, err
	}

	sessions, err := p.csv.LoadTrainingSessions()
	if err != nil {
		return nil, err
	}

	// 查找动作名称
	var exerciseName string
	for _, e := range exercises {
		if e.ID == exerciseID {
			exerciseName = e.Name
			break
		}
	}

	// 按session分组该动作的记录
	sessionRecords := make(map[int][]TrainingRecord) // sessionID -> records
	for _, r := range records {
		if r.ExerciseID == exerciseID {
			sessionRecords[r.SessionID] = append(sessionRecords[r.SessionID], r)
		}
	}

	// 按日期排序
	type sessionWithDate struct {
		sessionID int
		date      string
		records   []TrainingRecord
	}
	var sortedSessions []sessionWithDate
	for sid, recs := range sessionRecords {
		var date string
		for _, s := range sessions {
			if s.SessionID == sid {
				date = s.Date
				break
			}
		}
		sortedSessions = append(sortedSessions, sessionWithDate{
			sessionID: sid,
			date:      date,
			records:   recs,
		})
	}

	sort.Slice(sortedSessions, func(i, j int) bool {
		return sortedSessions[i].date < sortedSessions[j].date
	})

	// 构建进度数据
	progress := &ProgressData{
		ExerciseID:   exerciseID,
		ExerciseName: exerciseName,
		Dates:        []string{},
		Sets:         []SetProgress{},
	}

	// 找出最大组数
	maxSets := 0
	for _, s := range sortedSessions {
		if len(s.records) > maxSets {
			maxSets = len(s.records)
		}
	}

	// 初始化每一组的数据
	for i := 0; i < maxSets; i++ {
		progress.Sets = append(progress.Sets, SetProgress{
			SetNumber: i + 1,
			Weights:   []float64{},
			Reps:      []int{},
		})
	}

	// 填充数据
	for _, s := range sortedSessions {
		progress.Dates = append(progress.Dates, s.date)

		// 按setNumber排序记录
		sort.Slice(s.records, func(i, j int) bool {
			return s.records[i].SetNumber < s.records[j].SetNumber
		})

		// 为每一组填充数据（如果没有该组，填充0）
		for setIdx := 0; setIdx < maxSets; setIdx++ {
			if setIdx < len(s.records) {
				progress.Sets[setIdx].Weights = append(progress.Sets[setIdx].Weights, s.records[setIdx].Weight)
				progress.Sets[setIdx].Reps = append(progress.Sets[setIdx].Reps, s.records[setIdx].Reps)
			} else {
				progress.Sets[setIdx].Weights = append(progress.Sets[setIdx].Weights, 0)
				progress.Sets[setIdx].Reps = append(progress.Sets[setIdx].Reps, 0)
			}
		}
	}

	return progress, nil
}

// GetMuscleProgress 获取某肌肉群所有动作的进度
func (p *ProgressAnalyzer) GetMuscleProgress(muscleGroup string) (*MuscleProgress, error) {
	exercises, err := p.csv.LoadExercises()
	if err != nil {
		return nil, err
	}

	var exerciseIDs []int
	for _, e := range exercises {
		if e.MuscleGroup == muscleGroup {
			exerciseIDs = append(exerciseIDs, e.ID)
		}
	}

	progress := &MuscleProgress{
		MuscleGroup: muscleGroup,
		Exercises:   []ProgressData{},
	}

	for _, eid := range exerciseIDs {
		exerciseProgress, err := p.GetExerciseProgress(eid)
		if err != nil {
			continue
		}
		progress.Exercises = append(progress.Exercises, *exerciseProgress)
	}

	return progress, nil
}

// GetLastRecordForGroup 获取某动作组的上一次训练记录
func (p *ProgressAnalyzer) GetLastRecordForGroup(groupID int) (*LastRecordResponse, error) {
	sessions, err := p.csv.LoadTrainingSessions()
	if err != nil {
		return nil, err
	}

	// 找出该group最后一次完成的session
	var lastSession *TrainingSession
	for _, s := range sessions {
		if s.GroupID == groupID && s.Status == "completed" {
			if lastSession == nil || s.Date > lastSession.Date {
				lastSession = &s
			}
		}
	}

	if lastSession == nil {
		return &LastRecordResponse{
			SessionID:        0,
			Date:             "",
			ExerciseRecords:  map[int][]TrainingRecord{},
		}, nil
	}

	// 加载该session的所有记录
	records, err := p.csv.LoadTrainingRecords()
	if err != nil {
		return nil, err
	}

	exerciseRecords := make(map[int][]TrainingRecord)
	for _, r := range records {
		if r.SessionID == lastSession.SessionID {
			exerciseRecords[r.ExerciseID] = append(exerciseRecords[r.ExerciseID], r)
		}
	}

	// 每个动作的记录按setNumber排序
	for eid, recs := range exerciseRecords {
		sort.Slice(recs, func(i, j int) bool {
			return recs[i].SetNumber < recs[j].SetNumber
		})
		exerciseRecords[eid] = recs
	}

	return &LastRecordResponse{
		SessionID:       lastSession.SessionID,
		Date:            lastSession.Date,
		ExerciseRecords: exerciseRecords,
	}, nil
}
