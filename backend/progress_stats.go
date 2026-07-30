package main

import (
	"math"
	"sort"
	"time"
)

const (
	progressVolumeWeight       = 0.4
	progressBestWeight         = 0.6
	progressHistoryDecay       = 0.7
	progressMaxComparisons     = 6
	progressFullConfidenceAt   = 3
	progressPercentPrecision   = 10.0
	progressStatusImproved     = "improved"
	progressStatusStable       = "stable"
	progressStatusDeclined     = "declined"
	progressStatusUntrained    = "untrained"
	progressStatusInsufficient = "insufficient"
)

// ExerciseProgressSummary is the single progress result shown for an exercise.
type ExerciseProgressSummary struct {
	ExerciseID       int      `json:"exerciseId"`
	ExerciseName     string   `json:"exerciseName"`
	MuscleGroup      string   `json:"muscleGroup"`
	Unit             string   `json:"unit"`
	ProgressPercent  *float64 `json:"progressPercent"`
	Status           string   `json:"status"`
	Comparisons      int      `json:"comparisons"`
	Confidence       float64  `json:"confidence"`
	LastTrainingDate string   `json:"lastTrainingDate"`
}

type ProgressStatusCounts struct {
	Improved     int `json:"improved"`
	Stable       int `json:"stable"`
	Declined     int `json:"declined"`
	Untrained    int `json:"untrained"`
	Insufficient int `json:"insufficient"`
}

// ProgressSummary is the cycle-level progress view.
type ProgressSummary struct {
	Days                   int                       `json:"days"`
	StartDate              string                    `json:"startDate"`
	EndDate                string                    `json:"endDate"`
	OverallProgressPercent *float64                  `json:"overallProgressPercent"`
	CoveragePercent        float64                   `json:"coveragePercent"`
	Counts                 ProgressStatusCounts      `json:"counts"`
	Exercises              []ExerciseProgressSummary `json:"exercises"`
}

type DayExerciseProgress struct {
	ExerciseID      int      `json:"exerciseId"`
	ExerciseName    string   `json:"exerciseName"`
	MuscleGroup     string   `json:"muscleGroup"`
	Unit            string   `json:"unit"`
	ProgressPercent *float64 `json:"progressPercent"`
	Status          string   `json:"status"`
}

type DayProgress struct {
	Date      string                `json:"date"`
	Exercises []DayExerciseProgress `json:"exercises"`
}

type exerciseProgressObservation struct {
	Date   string
	Volume float64
	Best   float64
}

type exerciseProgressTransition struct {
	Date     string
	LogRatio float64
}

// GetProgressSummary calculates one comparable progress percentage per action,
// then aggregates the current cycle using confidence-weighted geometric mean.
func (s *StatsAnalyzer) GetProgressSummary(days, exerciseID int, muscleGroup string) (*ProgressSummary, error) {
	if days <= 0 {
		days = 30
	}

	exercises, observations, sessions, err := s.loadProgressObservations(exerciseID, muscleGroup)
	if err != nil {
		return nil, err
	}
	endDate := statsEndDate(sessions)
	startDate := endDate.AddDate(0, 0, -days+1)
	start := startDate.Format("2006-01-02")
	end := endDate.Format("2006-01-02")

	summary := &ProgressSummary{
		Days:      days,
		StartDate: start,
		EndDate:   end,
		Exercises: make([]ExerciseProgressSummary, 0, len(exercises)),
	}

	overallWeightedLog := 0.0
	overallWeight := 0.0
	validExercises := 0
	for _, exercise := range exercises {
		action := summarizeExerciseProgress(exercise, observations[exercise.ID], start, end)
		summary.Exercises = append(summary.Exercises, action)
		switch action.Status {
		case progressStatusImproved:
			summary.Counts.Improved++
		case progressStatusStable:
			summary.Counts.Stable++
		case progressStatusDeclined:
			summary.Counts.Declined++
		case progressStatusUntrained:
			summary.Counts.Untrained++
		default:
			summary.Counts.Insufficient++
		}

		if action.ProgressPercent != nil && action.Confidence > 0 {
			overallWeightedLog += action.Confidence * math.Log1p(*action.ProgressPercent/100)
			overallWeight += action.Confidence
			validExercises++
		}
	}

	if len(exercises) > 0 {
		summary.CoveragePercent = roundProgressPercent(float64(validExercises) / float64(len(exercises)) * 100)
	}
	if overallWeight > 0 {
		value := roundProgressPercent(math.Expm1(overallWeightedLog/overallWeight) * 100)
		summary.OverallProgressPercent = &value
	}
	return summary, nil
}

// GetDayProgress returns only the comprehensive progress result for each
// exercise trained on the requested day.
func (s *StatsAnalyzer) GetDayProgress(date string, exerciseID int, muscleGroup string) (*DayProgress, error) {
	exercises, observations, _, err := s.loadProgressObservations(exerciseID, muscleGroup)
	if err != nil {
		return nil, err
	}

	result := &DayProgress{Date: date, Exercises: []DayExerciseProgress{}}
	for _, exercise := range exercises {
		actionObservations := observations[exercise.ID]
		for index, current := range actionObservations {
			if current.Date != date {
				continue
			}

			action := DayExerciseProgress{
				ExerciseID:   exercise.ID,
				ExerciseName: exercise.Name,
				MuscleGroup:  exercise.MuscleGroup,
				Unit:         exercise.Unit,
				Status:       progressStatusInsufficient,
			}
			if index > 0 {
				value := roundProgressPercent(progressPercent(actionObservations[index-1], current))
				action.ProgressPercent = &value
				action.Status = progressStatus(value)
			}
			result.Exercises = append(result.Exercises, action)
			break
		}
	}

	sort.Slice(result.Exercises, func(i, j int) bool {
		if result.Exercises[i].MuscleGroup == result.Exercises[j].MuscleGroup {
			return result.Exercises[i].ExerciseName < result.Exercises[j].ExerciseName
		}
		return result.Exercises[i].MuscleGroup < result.Exercises[j].MuscleGroup
	})
	return result, nil
}

func (s *StatsAnalyzer) loadProgressObservations(exerciseID int, muscleGroup string) ([]Exercise, map[int][]exerciseProgressObservation, []TrainingSession, error) {
	exercises, err := s.csv.LoadExercises()
	if err != nil {
		return nil, nil, nil, err
	}
	sessions, err := s.csv.LoadTrainingSessions()
	if err != nil {
		return nil, nil, nil, err
	}
	records, err := s.csv.LoadTrainingRecords()
	if err != nil {
		return nil, nil, nil, err
	}

	filteredExercises := make([]Exercise, 0, len(exercises))
	exerciseMap := make(map[int]Exercise)
	for _, exercise := range exercises {
		if exercise.ID <= 0 || !matchesStatsFilter(exercise, exerciseID, muscleGroup) {
			continue
		}
		filteredExercises = append(filteredExercises, exercise)
		exerciseMap[exercise.ID] = exercise
	}
	sort.Slice(filteredExercises, func(i, j int) bool {
		if filteredExercises[i].MuscleGroup == filteredExercises[j].MuscleGroup {
			return filteredExercises[i].Name < filteredExercises[j].Name
		}
		return filteredExercises[i].MuscleGroup < filteredExercises[j].MuscleGroup
	})

	completedSessions := make([]TrainingSession, 0, len(sessions))
	sessionMap := make(map[int]TrainingSession)
	for _, session := range sessions {
		if session.Status != "completed" {
			continue
		}
		completedSessions = append(completedSessions, session)
		sessionMap[session.SessionID] = session
	}

	daily := make(map[int]map[string]*exerciseProgressObservation)
	for _, record := range records {
		session, ok := sessionMap[record.SessionID]
		if !ok {
			continue
		}
		exercise, ok := exerciseMap[record.ExerciseID]
		if !ok {
			continue
		}
		value := progressRecordValue(exercise, record)
		if value <= 0 {
			continue
		}
		if daily[exercise.ID] == nil {
			daily[exercise.ID] = make(map[string]*exerciseProgressObservation)
		}
		observation := daily[exercise.ID][session.Date]
		if observation == nil {
			observation = &exerciseProgressObservation{Date: session.Date}
			daily[exercise.ID][session.Date] = observation
		}
		observation.Volume += value
		if value > observation.Best {
			observation.Best = value
		}
	}

	observations := make(map[int][]exerciseProgressObservation)
	for currentExerciseID, byDate := range daily {
		for _, observation := range byDate {
			observations[currentExerciseID] = append(observations[currentExerciseID], *observation)
		}
		sort.Slice(observations[currentExerciseID], func(i, j int) bool {
			return observations[currentExerciseID][i].Date < observations[currentExerciseID][j].Date
		})
	}
	return filteredExercises, observations, completedSessions, nil
}

func summarizeExerciseProgress(exercise Exercise, observations []exerciseProgressObservation, start, end string) ExerciseProgressSummary {
	action := ExerciseProgressSummary{
		ExerciseID:   exercise.ID,
		ExerciseName: exercise.Name,
		MuscleGroup:  exercise.MuscleGroup,
		Unit:         exercise.Unit,
		Status:       progressStatusInsufficient,
	}
	if len(observations) > 0 {
		action.LastTrainingDate = observations[len(observations)-1].Date
	}

	trainedInPeriod := false
	transitions := make([]exerciseProgressTransition, 0, len(observations))
	for index, current := range observations {
		if current.Date >= start && current.Date <= end {
			trainedInPeriod = true
		}
		if index == 0 || current.Date < start || current.Date > end {
			continue
		}
		transitions = append(transitions, exerciseProgressTransition{
			Date:     current.Date,
			LogRatio: progressLogRatio(observations[index-1], current),
		})
	}

	if !trainedInPeriod {
		if len(observations) > 0 {
			action.Status = progressStatusUntrained
		}
		return action
	}
	if len(transitions) == 0 {
		return action
	}

	action.Comparisons = len(transitions)
	action.Confidence = math.Min(float64(action.Comparisons)/progressFullConfidenceAt, 1)
	weightedLog := 0.0
	totalWeight := 0.0
	for index := len(transitions) - 1; index >= 0 && len(transitions)-index <= progressMaxComparisons; index-- {
		recency := len(transitions) - 1 - index
		weight := math.Pow(progressHistoryDecay, float64(recency))
		weightedLog += weight * transitions[index].LogRatio
		totalWeight += weight
	}
	value := roundProgressPercent(math.Expm1(weightedLog/totalWeight) * 100)
	action.ProgressPercent = &value
	action.Status = progressStatus(value)
	return action
}

func progressRecordValue(exercise Exercise, record TrainingRecord) float64 {
	switch exercise.Unit {
	case "duration":
		return float64(record.Duration)
	case "reps":
		return float64(record.Reps)
	default:
		return record.Weight
	}
}

func progressLogRatio(previous, current exerciseProgressObservation) float64 {
	return progressVolumeWeight*math.Log(current.Volume/previous.Volume) +
		progressBestWeight*math.Log(current.Best/previous.Best)
}

func progressPercent(previous, current exerciseProgressObservation) float64 {
	return math.Expm1(progressLogRatio(previous, current)) * 100
}

func progressStatus(percent float64) string {
	switch {
	case percent > 0:
		return progressStatusImproved
	case percent < 0:
		return progressStatusDeclined
	default:
		return progressStatusStable
	}
}

func roundProgressPercent(value float64) float64 {
	rounded := math.Round(value*progressPercentPrecision) / progressPercentPrecision
	if rounded == 0 {
		return 0
	}
	return rounded
}

// Keep the date parsing behavior explicit for handler validation and tests.
func validProgressDate(value string) bool {
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}
