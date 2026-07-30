package main

import (
	"math"
	"testing"
	"time"
)

func TestProgressSummaryAndDayProgress(t *testing.T) {
	h, err := NewSQLiteHandler(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()

	exercises := []Exercise{
		{ID: 1, Name: "负重动作", MuscleGroup: "腿部", Unit: "kg"},
		{ID: 2, Name: "次数动作", MuscleGroup: "胸部", Unit: "reps"},
		{ID: 3, Name: "持续动作", MuscleGroup: "核心", Unit: "duration"},
		{ID: 4, Name: "未训练动作", MuscleGroup: "背部", Unit: "reps"},
		{ID: 5, Name: "周期未训练动作", MuscleGroup: "肩部", Unit: "kg"},
	}
	if err := h.SaveExercises(exercises); err != nil {
		t.Fatal(err)
	}
	if err := h.SaveExerciseGroups([]ExerciseGroup{
		{ID: 1, Name: "测试训练", ExerciseIDs: []int{1, 2, 3, 4, 5}},
	}); err != nil {
		t.Fatal(err)
	}

	today := time.Now().Format("2006-01-02")
	previous := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	old := time.Now().AddDate(0, 0, -60).Format("2006-01-02")
	if err := h.SaveTrainingData([]TrainingSession{
		{SessionID: 1, GroupID: 1, Date: previous, Status: "completed", DurationMinutes: 40},
		{SessionID: 2, GroupID: 1, Date: today, Status: "completed", DurationMinutes: 40},
		{SessionID: 3, GroupID: 1, Date: today, Status: "planned", DurationMinutes: 40},
		{SessionID: 4, GroupID: 1, Date: old, Status: "completed", DurationMinutes: 40},
	}, []TrainingRecord{
		{RecordID: 1, SessionID: 1, ExerciseID: 1, SetNumber: 1, Weight: 10, Reps: 20},
		{RecordID: 2, SessionID: 1, ExerciseID: 1, SetNumber: 2, Weight: 10, Reps: 20},
		{RecordID: 3, SessionID: 1, ExerciseID: 2, SetNumber: 1, Reps: 10},
		{RecordID: 4, SessionID: 1, ExerciseID: 2, SetNumber: 2, Reps: 10},
		{RecordID: 5, SessionID: 1, ExerciseID: 3, SetNumber: 1, Duration: 30},
		{RecordID: 6, SessionID: 1, ExerciseID: 3, SetNumber: 2, Duration: 30},
		{RecordID: 7, SessionID: 2, ExerciseID: 1, SetNumber: 1, Weight: 11, Reps: 1},
		{RecordID: 8, SessionID: 2, ExerciseID: 1, SetNumber: 2, Weight: 11, Reps: 1},
		{RecordID: 9, SessionID: 2, ExerciseID: 2, SetNumber: 1, Reps: 12},
		{RecordID: 10, SessionID: 2, ExerciseID: 2, SetNumber: 2, Reps: 12},
		{RecordID: 11, SessionID: 2, ExerciseID: 3, SetNumber: 1, Duration: 33},
		{RecordID: 12, SessionID: 2, ExerciseID: 3, SetNumber: 2, Duration: 33},
		{RecordID: 13, SessionID: 3, ExerciseID: 1, SetNumber: 1, Weight: 100, Reps: 100},
		{RecordID: 14, SessionID: 4, ExerciseID: 5, SetNumber: 1, Weight: 8, Reps: 20},
	}); err != nil {
		t.Fatal(err)
	}

	summary, err := NewStatsAnalyzer(h).GetProgressSummary(30, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.Exercises) != 5 || summary.Counts.Improved != 3 ||
		summary.Counts.Insufficient != 1 || summary.Counts.Untrained != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if summary.CoveragePercent != 60 {
		t.Fatalf("coverage=%v, want 60", summary.CoveragePercent)
	}

	progressByID := make(map[int]ExerciseProgressSummary)
	for _, action := range summary.Exercises {
		progressByID[action.ExerciseID] = action
	}
	assertProgressPercent(t, progressByID[1].ProgressPercent, 10)
	assertProgressPercent(t, progressByID[2].ProgressPercent, 20)
	assertProgressPercent(t, progressByID[3].ProgressPercent, 10)
	if progressByID[4].ProgressPercent != nil || progressByID[4].Status != progressStatusInsufficient {
		t.Fatalf("untrained action=%+v", progressByID[4])
	}
	if progressByID[5].ProgressPercent != nil || progressByID[5].Status != progressStatusUntrained {
		t.Fatalf("out-of-period action=%+v", progressByID[5])
	}
	// The kg result must ignore reps: 20kg -> 22kg is +10%, even though
	// repetitions deliberately fell from 40 to 2.
	wantOverall := (math.Pow(1.1*1.2*1.1, 1.0/3.0) - 1) * 100
	assertProgressPercent(t, summary.OverallProgressPercent, math.Round(wantOverall*10)/10)
	if got := recordRaw(exercises[0], TrainingRecord{Weight: 10, Reps: 20}); got != 10 {
		t.Fatalf("kg raw value=%v, want weight only", got)
	}

	day, err := NewStatsAnalyzer(h).GetDayProgress(today, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(day.Exercises) != 3 {
		t.Fatalf("day exercises=%+v", day.Exercises)
	}
	for _, action := range day.Exercises {
		if action.ProgressPercent == nil || action.Status != progressStatusImproved {
			t.Fatalf("day action=%+v", action)
		}
	}

	firstDay, err := NewStatsAnalyzer(h).GetDayProgress(previous, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range firstDay.Exercises {
		if action.ProgressPercent != nil || action.Status != progressStatusInsufficient {
			t.Fatalf("first day action=%+v", action)
		}
	}
}

func assertProgressPercent(t *testing.T, got *float64, want float64) {
	t.Helper()
	if got == nil || math.Abs(*got-want) > 0.001 {
		t.Fatalf("progress=%v, want %.1f", got, want)
	}
}
