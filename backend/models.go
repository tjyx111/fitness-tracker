package main

import "time"

// Exercise 健身动作
type Exercise struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	MuscleGroup string `json:"muscleGroup"`
	Unit        string `json:"unit"` // kg（重量）、reps（次数）或 duration（持续时间）
}

// ExerciseGroup 健身动作组（训练计划）
type ExerciseGroup struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ExerciseIDs []int  `json:"exerciseIds"`
}

// TrainingSession 训练会话（每天的训练）
type TrainingSession struct {
	SessionID       int       `json:"sessionId"`
	GroupID         int       `json:"groupId"`
	Date            string    `json:"date"`
	Status          string    `json:"status"`          // completed, planned
	DurationMinutes int       `json:"durationMinutes"` // 训练时长(分钟)
	CreatedAt       time.Time `json:"createdAt"`
}

// TrainingRecord 训练记录明细（每一组的数据）
type TrainingRecord struct {
	RecordID   int     `json:"recordId"`
	SessionID  int     `json:"sessionId"`
	ExerciseID int     `json:"exerciseId"`
	SetNumber  int     `json:"setNumber"`
	Weight     float64 `json:"weight"`   // 重量
	Reps       int     `json:"reps"`     // 次数
	Duration   int     `json:"duration"` // 持续时间(秒)
	Note       string  `json:"note"`
}

// ExerciseRecord 某个动作在一次训练中的所有组
type ExerciseRecord struct {
	ExerciseID int              `json:"exerciseId"`
	Sets       []TrainingRecord `json:"sets"`
}

// SessionData 提交训练会话的数据结构
type SessionData struct {
	GroupID         int              `json:"groupId"`
	Date            string           `json:"date"`
	DurationMinutes int              `json:"durationMinutes"`
	ExerciseRecords []ExerciseRecord `json:"exerciseRecords"`
}

// LastRecordResponse 某动作组的上一次训练记录
type LastRecordResponse struct {
	SessionID       int                      `json:"sessionId"`
	Date            string                   `json:"date"`
	ExerciseRecords map[int][]TrainingRecord `json:"exerciseRecords"` // exerciseId -> records
}

// SessionRecordsResponse 某一天某动作组的训练记录
type SessionRecordsResponse struct {
	SessionID       int                      `json:"sessionId"`
	GroupID         int                      `json:"groupId"`
	Date            string                   `json:"date"`
	DurationMinutes int                      `json:"durationMinutes"`
	ExerciseRecords map[int][]TrainingRecord `json:"exerciseRecords"` // exerciseId -> records
}

// DaySessionRecordsResponse 某一天的所有训练记录
type DaySessionRecordsResponse struct {
	Date     string                  `json:"date"`
	Sessions []SessionRecordResponse `json:"sessions"`
}

// SessionRecordResponse 某个训练会话的完整记录
type SessionRecordResponse struct {
	SessionID       int                      `json:"sessionId"`
	GroupID         int                      `json:"groupId"`
	GroupName       string                   `json:"groupName"`
	Date            string                   `json:"date"`
	DurationMinutes int                      `json:"durationMinutes"`
	Exercises       []Exercise               `json:"exercises"`
	ExerciseRecords map[int][]TrainingRecord `json:"exerciseRecords"` // exerciseId -> records
}

// ProgressData 进度图表数据
type ProgressData struct {
	ExerciseID   int           `json:"exerciseId"`
	ExerciseName string        `json:"exerciseName"`
	Dates        []string      `json:"dates"`
	Sets         []SetProgress `json:"sets"`
}

// SetProgress 某一组的变化数据
type SetProgress struct {
	SetNumber int       `json:"setNumber"`
	Weights   []float64 `json:"weights"`
	Reps      []int     `json:"reps"`
}

// MuscleProgress 某肌肉群所有动作的进度
type MuscleProgress struct {
	MuscleGroup string         `json:"muscleGroup"`
	Exercises   []ProgressData `json:"exercises"`
}
