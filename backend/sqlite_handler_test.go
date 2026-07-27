package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDeleteExerciseWithData(t *testing.T) {
	h, err := NewSQLiteHandler(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLiteHandler: %v", err)
	}
	defer h.Close()

	if err := h.SaveExercises([]Exercise{
		{ID: 1, Name: "test", MuscleGroup: "tmp", Unit: "kg"},
		{ID: 2, Name: "keep", MuscleGroup: "tmp", Unit: "reps"},
	}); err != nil {
		t.Fatalf("SaveExercises: %v", err)
	}
	if err := h.SaveExerciseGroups([]ExerciseGroup{
		{ID: 1, Name: "group", ExerciseIDs: []int{1, 2}},
	}); err != nil {
		t.Fatalf("SaveExerciseGroups: %v", err)
	}
	if err := h.SaveTrainingData([]TrainingSession{
		{SessionID: 1, GroupID: 1, Date: "2026-06-27", Status: "completed", DurationMinutes: 40, CreatedAt: time.Now()},
		{SessionID: 2, GroupID: 1, Date: "2026-06-28", Status: "completed", DurationMinutes: 40, CreatedAt: time.Now()},
		{SessionID: 3, GroupID: 1, Date: "2026-06-29", Status: "completed", DurationMinutes: 40, CreatedAt: time.Now()},
	}, []TrainingRecord{
		{RecordID: 1, SessionID: 1, ExerciseID: 1, SetNumber: 1, Weight: 10, Reps: 8},
		{RecordID: 2, SessionID: 2, ExerciseID: 1, SetNumber: 1, Weight: 20, Reps: 5},
		{RecordID: 3, SessionID: 2, ExerciseID: 2, SetNumber: 1, Reps: 12},
	}); err != nil {
		t.Fatalf("SaveTrainingData: %v", err)
	}

	found, deletedRecords, deletedSessions, err := h.DeleteExerciseWithData(1)
	if err != nil {
		t.Fatalf("DeleteExerciseWithData: %v", err)
	}
	if !found {
		t.Fatal("DeleteExerciseWithData found=false")
	}
	if deletedRecords != 2 {
		t.Fatalf("deletedRecords=%d, want 2", deletedRecords)
	}
	if deletedSessions != 1 {
		t.Fatalf("deletedSessions=%d, want 1", deletedSessions)
	}

	exercises, err := h.LoadExercises()
	if err != nil {
		t.Fatalf("LoadExercises: %v", err)
	}
	if len(exercises) != 1 || exercises[0].ID != 2 {
		t.Fatalf("remaining exercises=%v, want only id=2", exercises)
	}

	groups, err := h.LoadExerciseGroups()
	if err != nil {
		t.Fatalf("LoadExerciseGroups: %v", err)
	}
	if len(groups) != 1 || len(groups[0].ExerciseIDs) != 1 || groups[0].ExerciseIDs[0] != 2 {
		t.Fatalf("remaining group items=%v, want only exercise id=2", groups)
	}

	records, err := h.LoadTrainingRecords()
	if err != nil {
		t.Fatalf("LoadTrainingRecords: %v", err)
	}
	if len(records) != 1 || records[0].ExerciseID != 2 {
		t.Fatalf("remaining records=%v, want only exercise id=2", records)
	}

	sessions, err := h.LoadTrainingSessions()
	if err != nil {
		t.Fatalf("LoadTrainingSessions: %v", err)
	}
	if len(sessions) != 2 || sessions[0].SessionID != 2 || sessions[1].SessionID != 3 {
		t.Fatalf("remaining sessions=%v, want session ids 2 and 3", sessions)
	}
}

func TestExerciseGroupOrderRoundTrip(t *testing.T) {
	h, err := NewSQLiteHandler(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLiteHandler: %v", err)
	}
	defer h.Close()

	if err := h.SaveExercises([]Exercise{
		{ID: 1, Name: "深蹲", MuscleGroup: "腿部", Unit: "kg"},
		{ID: 2, Name: "硬拉", MuscleGroup: "背部", Unit: "kg"},
		{ID: 3, Name: "腿举", MuscleGroup: "腿部", Unit: "kg"},
	}); err != nil {
		t.Fatalf("SaveExercises: %v", err)
	}

	assertOrder := func(want []int) {
		t.Helper()
		groups, err := h.LoadExerciseGroups()
		if err != nil {
			t.Fatalf("LoadExerciseGroups: %v", err)
		}
		if len(groups) != 1 || len(groups[0].ExerciseIDs) != len(want) {
			t.Fatalf("groups=%v, want one group with order %v", groups, want)
		}
		for i, id := range want {
			if groups[0].ExerciseIDs[i] != id {
				t.Fatalf("exercise order=%v, want %v", groups[0].ExerciseIDs, want)
			}
		}
	}

	if err := h.SaveExerciseGroups([]ExerciseGroup{
		{ID: 1, Name: "腿部训练", ExerciseIDs: []int{3, 1, 2}},
	}); err != nil {
		t.Fatalf("SaveExerciseGroups initial: %v", err)
	}
	assertOrder([]int{3, 1, 2})

	if err := h.SaveExerciseGroups([]ExerciseGroup{
		{ID: 1, Name: "腿部训练", ExerciseIDs: []int{2, 3, 1}},
	}); err != nil {
		t.Fatalf("SaveExerciseGroups reordered: %v", err)
	}
	assertOrder([]int{2, 3, 1})
}

func TestNotes(t *testing.T) {
	h, err := NewSQLiteHandler(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLiteHandler: %v", err)
	}
	defer h.Close()

	back, err := h.AddNoteTag("背部")
	if err != nil {
		t.Fatalf("AddNoteTag back: %v", err)
	}
	chest, err := h.AddNoteTag("胸肌")
	if err != nil {
		t.Fatalf("AddNoteTag chest: %v", err)
	}
	if back.ID == chest.ID {
		t.Fatalf("tag IDs should differ: %v %v", back, chest)
	}

	if err := h.TouchNoteTag(back.ID); err != nil {
		t.Fatalf("TouchNoteTag: %v", err)
	}
	if _, err := h.SaveNote(back.ID, "今天状态很好"); err != nil {
		t.Fatalf("SaveNote: %v", err)
	}
	if _, err := h.SaveNote(back.ID, "标题：背部训练\n今天状态很好"); err != nil {
		t.Fatalf("SaveNote changed: %v", err)
	}

	note, err := h.LoadNote(back.ID)
	if err != nil {
		t.Fatalf("LoadNote: %v", err)
	}
	if note.Content != "标题：背部训练\n今天状态很好" {
		t.Fatalf("note content=%q", note.Content)
	}

	history, err := h.LoadNoteHistory(back.ID, 10)
	if err != nil {
		t.Fatalf("LoadNoteHistory: %v", err)
	}
	if len(history) != 0 {
		t.Fatalf("history len=%d, want 0 before creating a new note: %v", len(history), history)
	}

	note, err = h.CreateNewNote(back.ID)
	if err != nil {
		t.Fatalf("CreateNewNote: %v", err)
	}
	if note.Content != "" {
		t.Fatalf("new current note content=%q, want empty", note.Content)
	}

	history, err = h.LoadNoteHistory(back.ID, 10)
	if err != nil {
		t.Fatalf("LoadNoteHistory after new: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("history len=%d, want 1: %v", len(history), history)
	}
	if history[0].Summary != "背部训练" {
		t.Fatalf("history summary=%q, want title summary", history[0].Summary)
	}

	updated, err := h.UpdateNoteHistory(history[0].ID, "背部训练轻松完成，硬拉状态不错。")
	if err != nil {
		t.Fatalf("UpdateNoteHistory: %v", err)
	}
	if updated.Summary != "背部训练轻松完成，硬拉状态不错" {
		t.Fatalf("updated summary=%q", updated.Summary)
	}

	popular, err := h.LoadPopularNoteTags(1)
	if err != nil {
		t.Fatalf("LoadPopularNoteTags: %v", err)
	}
	if len(popular) != 1 || popular[0].ID != back.ID {
		t.Fatalf("popular=%v, want back tag first", popular)
	}
}

func TestTodoItems(t *testing.T) {
	h, err := NewSQLiteHandler(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLiteHandler: %v", err)
	}
	defer h.Close()

	first, err := h.AddTodoItem("买牛奶")
	if err != nil {
		t.Fatalf("AddTodoItem first: %v", err)
	}
	if first.StartAt == "" || first.StartAt != first.CreatedAt {
		t.Fatalf("first startAt=%q createdAt=%q, want automatic startAt from createdAt", first.StartAt, first.CreatedAt)
	}
	second, err := h.AddTodoItem("练背")
	if err != nil {
		t.Fatalf("AddTodoItem second: %v", err)
	}
	if first.ID == second.ID {
		t.Fatalf("todo IDs should differ: %v %v", first, second)
	}

	updated, err := h.UpdateTodoItem(first.ID, "买牛奶和鸡蛋", true)
	if err != nil {
		t.Fatalf("UpdateTodoItem: %v", err)
	}
	if updated.Title != "买牛奶和鸡蛋" || updated.StartAt != first.StartAt || !updated.Completed || updated.CompletedAt == "" {
		t.Fatalf("updated todo=%v", updated)
	}

	items, err := h.LoadTodoItems()
	if err != nil {
		t.Fatalf("LoadTodoItems: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("todo count=%d, want 2: %v", len(items), items)
	}
	if items[0].ID != second.ID || items[0].Completed {
		t.Fatalf("first visible todo=%v, want incomplete second item", items[0])
	}

	reopened, err := h.UpdateTodoItem(first.ID, first.Title, false)
	if err != nil {
		t.Fatalf("reopen todo: %v", err)
	}
	if reopened.Completed || reopened.CompletedAt != "" || reopened.StartAt != first.StartAt {
		t.Fatalf("reopened todo=%v, want incomplete with preserved startAt and empty completedAt", reopened)
	}

	if err := h.DeleteTodoItem(second.ID); err != nil {
		t.Fatalf("DeleteTodoItem: %v", err)
	}
	items, err = h.LoadTodoItems()
	if err != nil {
		t.Fatalf("LoadTodoItems after delete: %v", err)
	}
	if len(items) != 1 || items[0].ID != first.ID {
		t.Fatalf("remaining todos=%v, want first item only", items)
	}
}

func TestChallengesGenerateDailyItemsAndStats(t *testing.T) {
	h, err := NewSQLiteHandler(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLiteHandler: %v", err)
	}
	defer h.Close()

	start := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	challenge, err := h.CreateChallenge("生活习惯", start, 2, []string{"刷牙", "读书一小时"})
	if err != nil {
		t.Fatalf("CreateChallenge: %v", err)
	}
	if challenge.EndDate != time.Now().Format("2006-01-02") || len(challenge.Items) != 2 {
		t.Fatalf("challenge=%+v", challenge)
	}

	firstDay, err := h.LoadChallengeDay(start)
	if err != nil {
		t.Fatalf("LoadChallengeDay first day: %v", err)
	}
	if len(firstDay) != 1 || firstDay[0].TotalItems != 2 || firstDay[0].CompletedItems != 0 {
		t.Fatalf("firstDay=%+v", firstDay)
	}
	completed, err := h.UpdateChallengeDailyItem(firstDay[0].Items[0].ID, true)
	if err != nil {
		t.Fatalf("UpdateChallengeDailyItem: %v", err)
	}
	if !completed.Completed || completed.CompletedAt == "" {
		t.Fatalf("completed item=%+v", completed)
	}

	stats, err := h.GetChallengeStats(2)
	if err != nil {
		t.Fatalf("GetChallengeStats: %v", err)
	}
	if stats.TotalItems != 4 || stats.CompletedItems != 1 || stats.CompletionPercent != 25 {
		t.Fatalf("stats=%+v", stats)
	}
	if len(stats.Daily) != 2 || stats.Daily[0].CompletionPercent != 50 || stats.Daily[1].CompletionPercent != 0 {
		t.Fatalf("daily stats=%+v", stats.Daily)
	}
	if len(stats.ItemStats) != 2 {
		t.Fatalf("item stats=%+v", stats.ItemStats)
	}

	if _, err := h.CreateChallenge("并行挑战", start, 1, []string{"不应创建"}); !errors.Is(err, ErrActiveChallengeExists) {
		t.Fatalf("parallel challenge error=%v, want ErrActiveChallengeExists", err)
	}
	if err := h.TerminateChallenge(challenge.ID); err != nil {
		t.Fatalf("TerminateChallenge: %v", err)
	}
	if _, err := h.CreateChallenge("新挑战", time.Now().Format("2006-01-02"), 1, []string{"可以创建"}); err != nil {
		t.Fatalf("CreateChallenge after termination: %v", err)
	}
}

func TestCompletedChallengeRemainsAvailableAsReadOnlyHistory(t *testing.T) {
	h, err := NewSQLiteHandler(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLiteHandler: %v", err)
	}
	defer h.Close()

	start := time.Now().AddDate(0, 0, -8).Format("2006-01-02")
	challenge, err := h.CreateChallenge("历史挑战", start, 7, []string{"早睡", "拉伸"})
	if err != nil {
		t.Fatalf("CreateChallenge: %v", err)
	}
	if _, err := h.db.Exec(`UPDATE challenge_daily_items SET completed=1,completed_at='done' WHERE id=(SELECT MIN(id) FROM challenge_daily_items)`); err != nil {
		t.Fatalf("mark historical item complete: %v", err)
	}

	activeDay, err := h.LoadChallengeDay(start)
	if err != nil {
		t.Fatalf("LoadChallengeDay: %v", err)
	}
	if len(activeDay) != 0 {
		t.Fatalf("activeDay=%+v, want completed challenge hidden from editable view", activeDay)
	}

	history, err := h.LoadChallengeHistory()
	if err != nil {
		t.Fatalf("LoadChallengeHistory: %v", err)
	}
	if len(history) != 1 || history[0].ID != challenge.ID || history[0].Status != "completed" {
		t.Fatalf("history=%+v", history)
	}
	if history[0].TotalDays != 7 || history[0].ItemCount != 2 || history[0].TotalItems != 14 || history[0].CompletedItems != 1 {
		t.Fatalf("history summary=%+v", history[0])
	}

	detail, err := h.LoadChallengeDetail(challenge.ID)
	if err != nil {
		t.Fatalf("LoadChallengeDetail: %v", err)
	}
	if detail.Challenge.Status != "completed" || len(detail.Days) != 7 || detail.Days[0].Status != "completed" {
		t.Fatalf("detail=%+v", detail)
	}

	historyDay, err := h.LoadChallengeHistoryDay(start)
	if err != nil {
		t.Fatalf("LoadChallengeHistoryDay: %v", err)
	}
	if len(historyDay) != 1 || historyDay[0].ChallengeID != challenge.ID || historyDay[0].TotalItems != 2 {
		t.Fatalf("historyDay=%+v", historyDay)
	}
	if _, err := h.UpdateChallengeDailyItem(historyDay[0].Items[0].ID, true); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("UpdateChallengeDailyItem completed challenge error=%v, want sql.ErrNoRows", err)
	}
}

func TestSessionSubmitMergesSingleExerciseRecords(t *testing.T) {
	server, err := NewServer(t.TempDir())
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer server.csv.Close()

	if err := server.csv.SaveExercises([]Exercise{
		{ID: 1, Name: "卧推", MuscleGroup: "胸", Unit: "kg"},
		{ID: 2, Name: "飞鸟", MuscleGroup: "胸", Unit: "kg"},
	}); err != nil {
		t.Fatalf("SaveExercises: %v", err)
	}
	if err := server.csv.SaveExerciseGroups([]ExerciseGroup{
		{ID: 1, Name: "胸肌训练", ExerciseIDs: []int{1, 2}},
	}); err != nil {
		t.Fatalf("SaveExerciseGroups: %v", err)
	}

	postSession := func(payload SessionData) {
		t.Helper()
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/session", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		server.handleSessionSubmit(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("handleSessionSubmit code=%d body=%s", rec.Code, rec.Body.String())
		}
	}

	postSession(SessionData{
		GroupID:         1,
		Date:            "2026-06-28",
		DurationMinutes: 45,
		ExerciseRecords: []ExerciseRecord{{
			ExerciseID: 1,
			Sets: []TrainingRecord{
				{SetNumber: 1, Weight: 60, Reps: 8},
			},
		}},
	})
	postSession(SessionData{
		GroupID:         1,
		Date:            "2026-06-28",
		DurationMinutes: 45,
		ExerciseRecords: []ExerciseRecord{{
			ExerciseID: 2,
			Sets: []TrainingRecord{
				{SetNumber: 1, Weight: 15, Reps: 12},
			},
		}},
	})

	sessions, err := server.csv.LoadTrainingSessions()
	if err != nil {
		t.Fatalf("LoadTrainingSessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("sessions len=%d, want 1: %+v", len(sessions), sessions)
	}
	records, err := server.csv.LoadTrainingRecords()
	if err != nil {
		t.Fatalf("LoadTrainingRecords: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("records len=%d, want 2: %+v", len(records), records)
	}
	if records[0].SessionID != sessions[0].SessionID || records[1].SessionID != sessions[0].SessionID {
		t.Fatalf("records should share session %d: %+v", sessions[0].SessionID, records)
	}

	postSession(SessionData{
		GroupID:         1,
		Date:            "2026-06-28",
		DurationMinutes: 50,
		ExerciseRecords: []ExerciseRecord{{
			ExerciseID: 1,
			Sets: []TrainingRecord{
				{SetNumber: 1, Weight: 65, Reps: 6},
				{SetNumber: 2, Weight: 65, Reps: 5},
			},
		}},
	})

	sessions, err = server.csv.LoadTrainingSessions()
	if err != nil {
		t.Fatalf("LoadTrainingSessions after update: %v", err)
	}
	if len(sessions) != 1 || sessions[0].DurationMinutes != 50 {
		t.Fatalf("sessions after update=%+v, want one session with updated duration 50", sessions)
	}
	records, err = server.csv.LoadTrainingRecords()
	if err != nil {
		t.Fatalf("LoadTrainingRecords after update: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("records len after update=%d, want 3: %+v", len(records), records)
	}
	countByExercise := map[int]int{}
	for _, record := range records {
		countByExercise[record.ExerciseID]++
	}
	if countByExercise[1] != 2 || countByExercise[2] != 1 {
		t.Fatalf("countByExercise=%v, want exercise1=2 exercise2=1", countByExercise)
	}
}

func TestSessionSubmitReorderedExercisesAllocatesNewRecordIDs(t *testing.T) {
	server, err := NewServer(t.TempDir())
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer server.csv.Close()

	if err := server.csv.SaveExercises([]Exercise{
		{ID: 1, Name: "深蹲", MuscleGroup: "腿", Unit: "kg"},
		{ID: 2, Name: "臀桥", MuscleGroup: "臀", Unit: "reps"},
	}); err != nil {
		t.Fatalf("SaveExercises: %v", err)
	}
	if err := server.csv.SaveExerciseGroups([]ExerciseGroup{{ID: 1, Name: "腿部", ExerciseIDs: []int{1, 2}}}); err != nil {
		t.Fatalf("SaveExerciseGroups: %v", err)
	}

	post := func(payload SessionData) *httptest.ResponseRecorder {
		t.Helper()
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		rec := httptest.NewRecorder()
		server.handleSessionSubmit(rec, httptest.NewRequest(http.MethodPost, "/api/session", bytes.NewReader(body)))
		return rec
	}

	first := post(SessionData{GroupID: 1, Date: "2026-07-14", ExerciseRecords: []ExerciseRecord{
		{ExerciseID: 1, Sets: []TrainingRecord{{SetNumber: 1, Weight: 10, Reps: 8}}},
		{ExerciseID: 2, Sets: []TrainingRecord{{SetNumber: 1, Reps: 12}}},
	}})
	if first.Code != http.StatusCreated {
		t.Fatalf("first save code=%d body=%s", first.Code, first.Body.String())
	}

	second := post(SessionData{GroupID: 1, Date: "2026-07-14", ExerciseRecords: []ExerciseRecord{
		{ExerciseID: 2, Sets: []TrainingRecord{{SetNumber: 1, Reps: 15}}},
		{ExerciseID: 1, Sets: []TrainingRecord{{SetNumber: 1, Weight: 12, Reps: 8}}},
	}})
	if second.Code != http.StatusCreated {
		t.Fatalf("reordered save code=%d body=%s", second.Code, second.Body.String())
	}

	records, err := server.csv.LoadTrainingRecords()
	if err != nil {
		t.Fatalf("LoadTrainingRecords: %v", err)
	}
	if len(records) != 2 || records[0].RecordID <= 2 || records[1].RecordID <= 2 {
		t.Fatalf("records=%+v, want two newly allocated records", records)
	}
}

func TestSessionSubmitRejectsUnknownExercise(t *testing.T) {
	server, err := NewServer(t.TempDir())
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer server.csv.Close()

	if err := server.csv.SaveExercises([]Exercise{
		{ID: 1, Name: "小臂弯举", MuscleGroup: "小臂", Unit: "kg"},
	}); err != nil {
		t.Fatalf("SaveExercises: %v", err)
	}
	if err := server.csv.SaveExerciseGroups([]ExerciseGroup{
		{ID: 1, Name: "手臂训练", ExerciseIDs: []int{1}},
	}); err != nil {
		t.Fatalf("SaveExerciseGroups: %v", err)
	}

	body, err := json.Marshal(SessionData{
		GroupID:         1,
		Date:            "2026-07-02",
		DurationMinutes: 40,
		ExerciseRecords: []ExerciseRecord{{
			ExerciseID: 99,
			Sets: []TrainingRecord{
				{SetNumber: 1, Weight: 7.5, Reps: 8},
			},
		}},
	})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/session", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.handleSessionSubmit(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("handleSessionSubmit code=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "训练动作不存在") {
		t.Fatalf("response body=%q, want missing exercise message", rec.Body.String())
	}
}

func TestFilteredStatsBestPerformanceUsesRawMaxMetrics(t *testing.T) {
	h, err := NewSQLiteHandler(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLiteHandler: %v", err)
	}
	defer h.Close()

	if err := h.SaveExercises([]Exercise{
		{ID: 1, Name: "哑铃弯举", MuscleGroup: "手臂", Unit: "kg"},
		{ID: 2, Name: "俯卧撑", MuscleGroup: "胸肌", Unit: "reps"},
		{ID: 3, Name: "平板支撑", MuscleGroup: "核心", Unit: "duration"},
	}); err != nil {
		t.Fatalf("SaveExercises: %v", err)
	}
	if err := h.SaveExerciseGroups([]ExerciseGroup{
		{ID: 1, Name: "综合训练", ExerciseIDs: []int{1, 2, 3}},
	}); err != nil {
		t.Fatalf("SaveExerciseGroups: %v", err)
	}
	if err := h.SaveTrainingData([]TrainingSession{
		{SessionID: 1, GroupID: 1, Date: "2026-07-02", Status: "completed", DurationMinutes: 40},
	}, []TrainingRecord{
		{RecordID: 1, SessionID: 1, ExerciseID: 1, SetNumber: 1, Weight: 10, Reps: 12},
		{RecordID: 2, SessionID: 1, ExerciseID: 1, SetNumber: 2, Weight: 20, Reps: 5},
		{RecordID: 3, SessionID: 1, ExerciseID: 2, SetNumber: 1, Reps: 25},
		{RecordID: 4, SessionID: 1, ExerciseID: 3, SetNumber: 1, Reps: 1, Duration: 90},
	}); err != nil {
		t.Fatalf("SaveTrainingData: %v", err)
	}

	stats, err := NewStatsAnalyzer(h).GetFilteredStats(30, 0, "", 95)
	if err != nil {
		t.Fatalf("GetFilteredStats: %v", err)
	}
	if stats.BestPerformance.MaxWeight.Value != 20 || stats.BestPerformance.MaxWeight.Unit != "kg" {
		t.Fatalf("maxWeight=%+v, want value 20 kg", stats.BestPerformance.MaxWeight)
	}
	if stats.BestPerformance.MaxReps.Value != 25 || stats.BestPerformance.MaxReps.Unit != "reps" {
		t.Fatalf("maxReps=%+v, want value 25 reps", stats.BestPerformance.MaxReps)
	}
	if stats.BestPerformance.MaxDuration.Value != 90 || stats.BestPerformance.MaxDuration.Unit != "duration" {
		t.Fatalf("maxDuration=%+v, want value 90 duration", stats.BestPerformance.MaxDuration)
	}
}
