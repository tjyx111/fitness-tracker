package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteHandler is the application's local persistent store. The load/save
// methods intentionally preserve the old storage API so HTTP behavior remains
// compatible while the data is stored in normalized, transactional tables.
type SQLiteHandler struct {
	db      *sql.DB
	dataDir string
}

func NewSQLiteHandler(dataDir string) (*SQLiteHandler, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(dataDir, "fitness.db")
	dsn := fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL", dbPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	h := &SQLiteHandler{db: db, dataDir: dataDir}
	if err := h.initialize(); err != nil {
		db.Close()
		return nil, err
	}
	return h, nil
}

func (h *SQLiteHandler) Close() error { return h.db.Close() }

func (h *SQLiteHandler) initialize() error {
	const schema = `
CREATE TABLE IF NOT EXISTS schema_meta (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS exercises (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL CHECK (trim(name) <> ''),
    muscle_group TEXT NOT NULL DEFAULT '',
    unit TEXT NOT NULL DEFAULT 'kg' CHECK (unit IN ('kg', 'reps', 'duration'))
);
CREATE TABLE IF NOT EXISTS exercise_groups (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL CHECK (trim(name) <> '')
);
CREATE TABLE IF NOT EXISTS exercise_group_items (
    group_id INTEGER NOT NULL REFERENCES exercise_groups(id) ON DELETE CASCADE,
    exercise_id INTEGER NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    position INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (group_id, exercise_id)
);
CREATE TABLE IF NOT EXISTS training_sessions (
    id INTEGER PRIMARY KEY,
    group_id INTEGER NOT NULL REFERENCES exercise_groups(id) ON DELETE RESTRICT,
    training_date TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'completed' CHECK (status IN ('completed', 'planned')),
    duration_minutes INTEGER NOT NULL DEFAULT 40 CHECK (duration_minutes > 0),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE TABLE IF NOT EXISTS training_records (
    id INTEGER PRIMARY KEY,
    session_id INTEGER NOT NULL REFERENCES training_sessions(id) ON DELETE CASCADE,
    exercise_id INTEGER NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT,
    set_number INTEGER NOT NULL CHECK (set_number > 0),
    weight REAL NOT NULL DEFAULT 0 CHECK (weight >= 0),
    reps INTEGER NOT NULL DEFAULT 0 CHECK (reps >= 0),
    duration_seconds INTEGER NOT NULL DEFAULT 0 CHECK (duration_seconds >= 0),
    note TEXT NOT NULL DEFAULT '',
    UNIQUE (session_id, exercise_id, set_number)
);
CREATE TABLE IF NOT EXISTS weight_records (
    id INTEGER PRIMARY KEY,
    record_date TEXT NOT NULL,
    weight REAL NOT NULL CHECK (weight > 0),
    body_fat REAL NOT NULL DEFAULT 0 CHECK (body_fat >= 0 AND body_fat <= 100),
    note TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS note_tags (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE CHECK (trim(name) <> ''),
    use_count INTEGER NOT NULL DEFAULT 0 CHECK (use_count >= 0),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE TABLE IF NOT EXISTS notes (
    id INTEGER PRIMARY KEY,
    tag_id INTEGER NOT NULL UNIQUE REFERENCES note_tags(id) ON DELETE CASCADE,
    content TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE TABLE IF NOT EXISTS note_history (
    id INTEGER PRIMARY KEY,
    tag_id INTEGER NOT NULL REFERENCES note_tags(id) ON DELETE CASCADE,
    summary TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE TABLE IF NOT EXISTS todo_items (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL CHECK (trim(title) <> ''),
    start_at TEXT NOT NULL DEFAULT '',
    completed INTEGER NOT NULL DEFAULT 0 CHECK (completed IN (0, 1)),
    completed_at TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE TABLE IF NOT EXISTS challenges (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL CHECK (trim(name) <> ''),
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'terminated')),
    terminated_at TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    CHECK (end_date >= start_date)
);
CREATE TABLE IF NOT EXISTS challenge_items (
    id INTEGER PRIMARY KEY,
    challenge_id INTEGER NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    title TEXT NOT NULL CHECK (trim(title) <> ''),
    position INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS challenge_daily_items (
    id INTEGER PRIMARY KEY,
    challenge_item_id INTEGER NOT NULL REFERENCES challenge_items(id) ON DELETE CASCADE,
    challenge_date TEXT NOT NULL,
    completed INTEGER NOT NULL DEFAULT 0 CHECK (completed IN (0, 1)),
    completed_at TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (challenge_item_id, challenge_date)
);
CREATE INDEX IF NOT EXISTS idx_sessions_date_group ON training_sessions(training_date, group_id);
CREATE INDEX IF NOT EXISTS idx_records_exercise_session ON training_records(exercise_id, session_id);
CREATE INDEX IF NOT EXISTS idx_group_items_exercise ON exercise_group_items(exercise_id);
CREATE INDEX IF NOT EXISTS idx_weight_date ON weight_records(record_date DESC);
CREATE INDEX IF NOT EXISTS idx_note_tags_popular ON note_tags(use_count DESC, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_note_history_tag_created ON note_history(tag_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_todo_items_completed_updated ON todo_items(completed, updated_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_challenge_daily_items_date ON challenge_daily_items(challenge_date, completed);
CREATE INDEX IF NOT EXISTS idx_challenge_items_challenge_position ON challenge_items(challenge_id, position);
`
	if _, err := h.db.Exec(schema); err != nil {
		return fmt.Errorf("initialize sqlite schema: %w", err)
	}
	if err := h.ensureNoteHistorySummaryColumn(); err != nil {
		return err
	}
	if err := h.ensureTodoItemColumns(); err != nil {
		return err
	}
	if err := h.ensureChallengeColumns(); err != nil {
		return err
	}
	return h.importCSVOnce()
}

func (h *SQLiteHandler) ensureChallengeColumns() error {
	if err := h.ensureColumnExists("challenges", "status", "TEXT NOT NULL DEFAULT 'active'"); err != nil {
		return err
	}
	if err := h.ensureColumnExists("challenges", "terminated_at", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	_, err := h.db.Exec(`UPDATE challenges SET status='active' WHERE trim(status)=''`)
	return err
}

func (h *SQLiteHandler) ensureTodoItemColumns() error {
	if err := h.ensureColumnExists("todo_items", "start_at", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := h.ensureColumnExists("todo_items", "completed_at", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	_, err := h.db.Exec(`UPDATE todo_items SET start_at=created_at WHERE trim(start_at)=''`)
	return err
}

func (h *SQLiteHandler) ensureColumnExists(table string, column string, definition string) error {
	rows, err := h.db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return err
		}
		if name == column {
			return rows.Err()
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	_, err = h.db.Exec(`ALTER TABLE ` + table + ` ADD COLUMN ` + column + ` ` + definition)
	return err
}

func (h *SQLiteHandler) ensureNoteHistorySummaryColumn() error {
	rows, err := h.db.Query(`PRAGMA table_info(note_history)`)
	if err != nil {
		return err
	}

	hasSummary := false
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			rows.Close()
			return err
		}
		if name == "summary" {
			hasSummary = true
			break
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if !hasSummary {
		if _, err := h.db.Exec(`ALTER TABLE note_history ADD COLUMN summary TEXT NOT NULL DEFAULT ''`); err != nil {
			return err
		}
	}
	_, err = h.db.Exec(`UPDATE note_history SET summary='' WHERE summary IS NULL`)
	return err
}

func (h *SQLiteHandler) importCSVOnce() error {
	var done string
	err := h.db.QueryRow(`SELECT value FROM schema_meta WHERE key='csv_imported'`).Scan(&done)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	legacy := NewCSVHandler(h.dataDir)
	exercises, err := legacy.LoadExercises()
	if err != nil {
		return fmt.Errorf("load legacy exercises: %w", err)
	}
	groups, err := legacy.LoadExerciseGroups()
	if err != nil {
		return fmt.Errorf("load legacy groups: %w", err)
	}
	sessions, err := legacy.LoadTrainingSessions()
	if err != nil {
		return fmt.Errorf("load legacy sessions: %w", err)
	}
	records, err := legacy.LoadTrainingRecords()
	if err != nil {
		return fmt.Errorf("load legacy records: %w", err)
	}
	weights, err := loadLegacyWeightRecords(filepath.Join(h.dataDir, "weight_records.csv"))
	if err != nil {
		return fmt.Errorf("load legacy weights: %w", err)
	}

	tx, err := h.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, e := range exercises {
		if e.Unit == "" {
			e.Unit = "kg"
		}
		if _, err = tx.Exec(`INSERT OR IGNORE INTO exercises(id,name,muscle_group,unit) VALUES(?,?,?,?)`, e.ID, e.Name, e.MuscleGroup, e.Unit); err != nil {
			return fmt.Errorf("import exercise %d: %w", e.ID, err)
		}
	}
	for _, g := range groups {
		if _, err = tx.Exec(`INSERT OR IGNORE INTO exercise_groups(id,name) VALUES(?,?)`, g.ID, g.Name); err != nil {
			return fmt.Errorf("import group %d: %w", g.ID, err)
		}
		for pos, exerciseID := range g.ExerciseIDs {
			if _, err = tx.Exec(`INSERT OR IGNORE INTO exercise_group_items(group_id,exercise_id,position) VALUES(?,?,?)`, g.ID, exerciseID, pos); err != nil {
				return fmt.Errorf("import group %d exercise %d: %w", g.ID, exerciseID, err)
			}
		}
	}
	for _, s := range sessions {
		if s.DurationMinutes <= 0 {
			s.DurationMinutes = 40
		}
		createdAt := s.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		if _, err = tx.Exec(`INSERT OR IGNORE INTO training_sessions(id,group_id,training_date,status,duration_minutes,created_at) VALUES(?,?,?,?,?,?)`, s.SessionID, s.GroupID, s.Date, s.Status, s.DurationMinutes, createdAt.Format(time.RFC3339Nano)); err != nil {
			return fmt.Errorf("import session %d: %w", s.SessionID, err)
		}
	}
	for _, r := range records {
		if _, err = tx.Exec(`INSERT OR IGNORE INTO training_records(id,session_id,exercise_id,set_number,weight,reps,duration_seconds,note) VALUES(?,?,?,?,?,?,?,?)`, r.RecordID, r.SessionID, r.ExerciseID, r.SetNumber, r.Weight, r.Reps, r.Duration, r.Note); err != nil {
			return fmt.Errorf("import record %d: %w", r.RecordID, err)
		}
	}
	for _, w := range weights {
		if _, err = tx.Exec(`INSERT OR IGNORE INTO weight_records(id,record_date,weight,body_fat,note) VALUES(?,?,?,?,?)`, w.ID, w.Date, w.Weight, w.BodyFat, w.Note); err != nil {
			return fmt.Errorf("import weight %d: %w", w.ID, err)
		}
	}
	if _, err = tx.Exec(`INSERT INTO schema_meta(key,value) VALUES('csv_imported',?)`, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return err
	}
	return tx.Commit()
}

func (h *SQLiteHandler) LoadExercises() ([]Exercise, error) {
	rows, err := h.db.Query(`SELECT id,name,muscle_group,unit FROM exercises ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Exercise{}
	for rows.Next() {
		var v Exercise
		if err := rows.Scan(&v.ID, &v.Name, &v.MuscleGroup, &v.Unit); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (h *SQLiteHandler) SaveExercises(values []Exercise) error {
	return h.replace("exercises", idsOfExercises(values), func(tx *sql.Tx) error {
		for _, v := range values {
			if v.Unit == "" {
				v.Unit = "kg"
			}
			if _, err := tx.Exec(`INSERT INTO exercises(id,name,muscle_group,unit) VALUES(?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,muscle_group=excluded.muscle_group,unit=excluded.unit`, v.ID, v.Name, v.MuscleGroup, v.Unit); err != nil {
				return err
			}
		}
		return nil
	})
}

func (h *SQLiteHandler) DeleteExerciseWithData(id int) (bool, int64, int64, error) {
	tx, err := h.db.Begin()
	if err != nil {
		return false, 0, 0, err
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRow(`SELECT COUNT(1) FROM exercises WHERE id=?`, id).Scan(&exists); err != nil {
		return false, 0, 0, err
	}
	if exists == 0 {
		return false, 0, 0, nil
	}

	var deletedRecords int64
	var deletedSessions int64
	if err := tx.QueryRow(`SELECT COUNT(1) FROM training_records WHERE exercise_id=?`, id).Scan(&deletedRecords); err != nil {
		return false, 0, 0, err
	}
	if err := tx.QueryRow(`
SELECT COUNT(1)
FROM training_sessions s
WHERE s.id IN (SELECT session_id FROM training_records WHERE exercise_id=?)
  AND NOT EXISTS (
      SELECT 1
      FROM training_records r
      WHERE r.session_id=s.id
        AND r.exercise_id<>?
  )
`, id, id).Scan(&deletedSessions); err != nil {
		return false, 0, 0, err
	}

	if _, err := tx.Exec(`
DELETE FROM training_sessions
WHERE id IN (SELECT session_id FROM training_records WHERE exercise_id=?)
  AND NOT EXISTS (
      SELECT 1
      FROM training_records r
      WHERE r.session_id=training_sessions.id
        AND r.exercise_id<>?
  )
`, id, id); err != nil {
		return false, 0, 0, err
	}

	if _, err := tx.Exec(`DELETE FROM training_records WHERE exercise_id=?`, id); err != nil {
		return false, 0, 0, err
	}

	if _, err := tx.Exec(`DELETE FROM exercise_group_items WHERE exercise_id=?`, id); err != nil {
		return false, 0, 0, err
	}
	if _, err := tx.Exec(`DELETE FROM exercises WHERE id=?`, id); err != nil {
		return false, 0, 0, err
	}

	if err := tx.Commit(); err != nil {
		return false, 0, 0, err
	}
	return true, deletedRecords, deletedSessions, nil
}

func (h *SQLiteHandler) LoadExerciseGroups() ([]ExerciseGroup, error) {
	rows, err := h.db.Query(`SELECT g.id,g.name,i.exercise_id FROM exercise_groups g LEFT JOIN exercise_group_items i ON i.group_id=g.id ORDER BY g.id,i.position`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []ExerciseGroup{}
	index := map[int]int{}
	for rows.Next() {
		var id int
		var name string
		var eid sql.NullInt64
		if err := rows.Scan(&id, &name, &eid); err != nil {
			return nil, err
		}
		p, ok := index[id]
		if !ok {
			p = len(result)
			index[id] = p
			result = append(result, ExerciseGroup{ID: id, Name: name, ExerciseIDs: []int{}})
		}
		if eid.Valid {
			result[p].ExerciseIDs = append(result[p].ExerciseIDs, int(eid.Int64))
		}
	}
	return result, rows.Err()
}

func (h *SQLiteHandler) SaveExerciseGroups(values []ExerciseGroup) error {
	return h.replace("exercise_groups", idsOfGroups(values), func(tx *sql.Tx) error {
		for _, v := range values {
			if _, err := tx.Exec(`INSERT INTO exercise_groups(id,name) VALUES(?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name`, v.ID, v.Name); err != nil {
				return err
			}
			if _, err := tx.Exec(`DELETE FROM exercise_group_items WHERE group_id=?`, v.ID); err != nil {
				return err
			}
			for pos, eid := range v.ExerciseIDs {
				if _, err := tx.Exec(`INSERT INTO exercise_group_items(group_id,exercise_id,position) VALUES(?,?,?)`, v.ID, eid, pos); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (h *SQLiteHandler) LoadTrainingSessions() ([]TrainingSession, error) {
	rows, err := h.db.Query(`SELECT id,group_id,training_date,status,duration_minutes,created_at FROM training_sessions ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []TrainingSession{}
	for rows.Next() {
		var v TrainingSession
		var created string
		if err := rows.Scan(&v.SessionID, &v.GroupID, &v.Date, &v.Status, &v.DurationMinutes, &created); err != nil {
			return nil, err
		}
		v.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		result = append(result, v)
	}
	return result, rows.Err()
}
func (h *SQLiteHandler) SaveTrainingSessions(values []TrainingSession) error {
	return h.replace("training_sessions", idsOfSessions(values), func(tx *sql.Tx) error {
		for _, v := range values {
			if v.DurationMinutes <= 0 {
				v.DurationMinutes = 40
			}
			created := v.CreatedAt
			if created.IsZero() {
				created = time.Now().UTC()
			}
			if _, err := tx.Exec(`INSERT INTO training_sessions(id,group_id,training_date,status,duration_minutes,created_at) VALUES(?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET group_id=excluded.group_id,training_date=excluded.training_date,status=excluded.status,duration_minutes=excluded.duration_minutes`, v.SessionID, v.GroupID, v.Date, v.Status, v.DurationMinutes, created.Format(time.RFC3339Nano)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (h *SQLiteHandler) LoadTrainingRecords() ([]TrainingRecord, error) {
	rows, err := h.db.Query(`SELECT id,session_id,exercise_id,set_number,weight,reps,duration_seconds,note FROM training_records ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []TrainingRecord{}
	for rows.Next() {
		var v TrainingRecord
		if err := rows.Scan(&v.RecordID, &v.SessionID, &v.ExerciseID, &v.SetNumber, &v.Weight, &v.Reps, &v.Duration, &v.Note); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}
func (h *SQLiteHandler) SaveTrainingRecords(values []TrainingRecord) error {
	return h.replace("training_records", idsOfRecords(values), func(tx *sql.Tx) error {
		for _, v := range values {
			if _, err := tx.Exec(`INSERT INTO training_records(id,session_id,exercise_id,set_number,weight,reps,duration_seconds,note) VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET session_id=excluded.session_id,exercise_id=excluded.exercise_id,set_number=excluded.set_number,weight=excluded.weight,reps=excluded.reps,duration_seconds=excluded.duration_seconds,note=excluded.note`, v.RecordID, v.SessionID, v.ExerciseID, v.SetNumber, v.Weight, v.Reps, v.Duration, v.Note); err != nil {
				return err
			}
		}
		return nil
	})
}

// SaveTrainingData commits a session and all of its sets atomically. This
// prevents a valid session from being left without records (or vice versa).
func (h *SQLiteHandler) SaveTrainingData(sessions []TrainingSession, records []TrainingRecord) error {
	tx, err := h.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = h.replaceTx(tx, "training_sessions", idsOfSessions(sessions), func(tx *sql.Tx) error {
		for _, v := range sessions {
			if v.DurationMinutes <= 0 {
				v.DurationMinutes = 40
			}
			created := v.CreatedAt
			if created.IsZero() {
				created = time.Now().UTC()
			}
			if _, err := tx.Exec(`INSERT INTO training_sessions(id,group_id,training_date,status,duration_minutes,created_at) VALUES(?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET group_id=excluded.group_id,training_date=excluded.training_date,status=excluded.status,duration_minutes=excluded.duration_minutes`, v.SessionID, v.GroupID, v.Date, v.Status, v.DurationMinutes, created.Format(time.RFC3339Nano)); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if err = h.deleteRowsNotInTx(tx, "training_records", idsOfRecords(records)); err != nil {
		return err
	}
	for _, v := range records {
		if _, err := tx.Exec(`INSERT INTO training_records(id,session_id,exercise_id,set_number,weight,reps,duration_seconds,note) VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET session_id=excluded.session_id,exercise_id=excluded.exercise_id,set_number=excluded.set_number,weight=excluded.weight,reps=excluded.reps,duration_seconds=excluded.duration_seconds,note=excluded.note`, v.RecordID, v.SessionID, v.ExerciseID, v.SetNumber, v.Weight, v.Reps, v.Duration, v.Note); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (h *SQLiteHandler) LoadWeightRecords() ([]WeightRecord, error) {
	rows, err := h.db.Query(`SELECT id,record_date,weight,body_fat,note FROM weight_records ORDER BY record_date,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []WeightRecord{}
	for rows.Next() {
		var v WeightRecord
		if err := rows.Scan(&v.ID, &v.Date, &v.Weight, &v.BodyFat, &v.Note); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}
func (h *SQLiteHandler) AddWeightRecord(v *WeightRecord) error {
	res, err := h.db.Exec(`INSERT INTO weight_records(record_date,weight,body_fat,note) VALUES(?,?,?,?)`, v.Date, v.Weight, v.BodyFat, v.Note)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	v.ID = int(id)
	return err
}
func (h *SQLiteHandler) GetLatestWeight() (float64, error) {
	var weight float64
	err := h.db.QueryRow(`SELECT weight FROM weight_records ORDER BY record_date DESC,id DESC LIMIT 1`).Scan(&weight)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return weight, err
}

func (h *SQLiteHandler) LoadNoteTags() ([]NoteTag, error) {
	rows, err := h.db.Query(`SELECT id,name,use_count,created_at,updated_at FROM note_tags ORDER BY name COLLATE NOCASE,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []NoteTag{}
	for rows.Next() {
		var v NoteTag
		if err := rows.Scan(&v.ID, &v.Name, &v.UseCount, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (h *SQLiteHandler) LoadPopularNoteTags(limit int) ([]NoteTag, error) {
	if limit <= 0 {
		limit = 4
	}
	rows, err := h.db.Query(`SELECT id,name,use_count,created_at,updated_at FROM note_tags ORDER BY use_count DESC,updated_at DESC,id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []NoteTag{}
	for rows.Next() {
		var v NoteTag
		if err := rows.Scan(&v.ID, &v.Name, &v.UseCount, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (h *SQLiteHandler) AddNoteTag(name string) (NoteTag, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := h.db.Exec(`
INSERT INTO note_tags(name,created_at,updated_at)
VALUES(?,?,?)
ON CONFLICT(name) DO UPDATE SET updated_at=excluded.updated_at
`, name, now, now); err != nil {
		return NoteTag{}, err
	}
	var tag NoteTag
	err := h.db.QueryRow(`SELECT id,name,use_count,created_at,updated_at FROM note_tags WHERE name=?`, name).
		Scan(&tag.ID, &tag.Name, &tag.UseCount, &tag.CreatedAt, &tag.UpdatedAt)
	return tag, err
}

func (h *SQLiteHandler) TouchNoteTag(id int) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := h.db.Exec(`UPDATE note_tags SET use_count=use_count+1,updated_at=? WHERE id=?`, now, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (h *SQLiteHandler) LoadNote(tagID int) (Note, error) {
	var note Note
	err := h.db.QueryRow(`SELECT id,tag_id,content,created_at,updated_at FROM notes WHERE tag_id=?`, tagID).
		Scan(&note.ID, &note.TagID, &note.Content, &note.CreatedAt, &note.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		note.TagID = tagID
		return note, nil
	}
	return note, err
}

func (h *SQLiteHandler) SaveNote(tagID int, content string) (Note, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	tx, err := h.db.Begin()
	if err != nil {
		return Note{}, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
INSERT INTO notes(tag_id,content,created_at,updated_at)
VALUES(?,?,?,?)
ON CONFLICT(tag_id) DO UPDATE SET content=excluded.content,updated_at=excluded.updated_at
`, tagID, content, now, now); err != nil {
		return Note{}, err
	}

	if _, err := tx.Exec(`UPDATE note_tags SET use_count=use_count+1,updated_at=? WHERE id=?`, now, tagID); err != nil {
		return Note{}, err
	}
	if err := tx.Commit(); err != nil {
		return Note{}, err
	}
	return h.LoadNote(tagID)
}

func (h *SQLiteHandler) CreateNewNote(tagID int) (Note, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	tx, err := h.db.Begin()
	if err != nil {
		return Note{}, err
	}
	defer tx.Rollback()

	var current string
	err = tx.QueryRow(`SELECT content FROM notes WHERE tag_id=?`, tagID).Scan(&current)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Note{}, err
	}
	if strings.TrimSpace(current) != "" {
		if _, err := tx.Exec(
			`INSERT INTO note_history(tag_id,summary,content,created_at) VALUES(?,?,?,?)`,
			tagID, summarizeNote(current), current, now,
		); err != nil {
			return Note{}, err
		}
	}

	if _, err := tx.Exec(`
INSERT INTO notes(tag_id,content,created_at,updated_at)
VALUES(?,?,?,?)
ON CONFLICT(tag_id) DO UPDATE SET content=excluded.content,updated_at=excluded.updated_at
`, tagID, "", now, now); err != nil {
		return Note{}, err
	}
	if _, err := tx.Exec(`UPDATE note_tags SET use_count=use_count+1,updated_at=? WHERE id=?`, now, tagID); err != nil {
		return Note{}, err
	}
	if err := tx.Commit(); err != nil {
		return Note{}, err
	}
	return h.LoadNote(tagID)
}

func (h *SQLiteHandler) LoadNoteHistory(tagID int, limit int) ([]NoteHistory, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := h.db.Query(`SELECT id,tag_id,summary,content,created_at FROM note_history WHERE tag_id=? ORDER BY created_at DESC,id DESC LIMIT ?`, tagID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []NoteHistory{}
	for rows.Next() {
		var v NoteHistory
		if err := rows.Scan(&v.ID, &v.TagID, &v.Summary, &v.Content, &v.CreatedAt); err != nil {
			return nil, err
		}
		if v.Summary == "" {
			v.Summary = summarizeNote(v.Content)
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (h *SQLiteHandler) UpdateNoteHistory(id int, content string) (NoteHistory, error) {
	nowSummary := summarizeNote(content)
	res, err := h.db.Exec(`UPDATE note_history SET content=?,summary=? WHERE id=?`, content, nowSummary, id)
	if err != nil {
		return NoteHistory{}, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return NoteHistory{}, err
	}
	if affected == 0 {
		return NoteHistory{}, sql.ErrNoRows
	}
	var v NoteHistory
	err = h.db.QueryRow(`SELECT id,tag_id,summary,content,created_at FROM note_history WHERE id=?`, id).
		Scan(&v.ID, &v.TagID, &v.Summary, &v.Content, &v.CreatedAt)
	return v, err
}

func (h *SQLiteHandler) LoadTodoItems() ([]TodoItem, error) {
	rows, err := h.db.Query(`
SELECT id,title,start_at,completed,completed_at,created_at,updated_at
FROM todo_items
ORDER BY completed ASC,
         CASE WHEN completed=0 AND start_at<>'' THEN 0 WHEN completed=0 THEN 1 ELSE 0 END ASC,
         CASE WHEN completed=0 THEN start_at END ASC,
         CASE WHEN completed=1 THEN completed_at END DESC,
         updated_at DESC,
         id DESC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []TodoItem{}
	for rows.Next() {
		var v TodoItem
		var completed int
		if err := rows.Scan(&v.ID, &v.Title, &v.StartAt, &completed, &v.CompletedAt, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		v.Completed = completed == 1
		result = append(result, v)
	}
	return result, rows.Err()
}

func (h *SQLiteHandler) AddTodoItem(title string) (TodoItem, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := h.db.Exec(`INSERT INTO todo_items(title,start_at,completed,completed_at,created_at,updated_at) VALUES(?,?,?,?,?,?)`, title, now, 0, "", now, now)
	if err != nil {
		return TodoItem{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return TodoItem{}, err
	}
	return h.loadTodoItem(int(id))
}

func (h *SQLiteHandler) UpdateTodoItem(id int, title string, completed bool) (TodoItem, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	var currentCompleted int
	var currentCompletedAt string
	var currentStartAt string
	var createdAt string
	if err := h.db.QueryRow(`SELECT completed,completed_at,start_at,created_at FROM todo_items WHERE id=?`, id).Scan(&currentCompleted, &currentCompletedAt, &currentStartAt, &createdAt); err != nil {
		return TodoItem{}, err
	}

	startAt := strings.TrimSpace(currentStartAt)
	if startAt == "" {
		startAt = strings.TrimSpace(createdAt)
	}
	if startAt == "" {
		startAt = now
	}

	completedValue := 0
	completedAt := currentCompletedAt
	if completed {
		completedValue = 1
		if currentCompleted == 0 || strings.TrimSpace(completedAt) == "" {
			completedAt = now
		}
	} else {
		completedAt = ""
	}
	res, err := h.db.Exec(`UPDATE todo_items SET title=?,start_at=?,completed=?,completed_at=?,updated_at=? WHERE id=?`, title, startAt, completedValue, completedAt, now, id)
	if err != nil {
		return TodoItem{}, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return TodoItem{}, err
	}
	if affected == 0 {
		return TodoItem{}, sql.ErrNoRows
	}
	return h.loadTodoItem(id)
}

func (h *SQLiteHandler) DeleteTodoItem(id int) error {
	res, err := h.db.Exec(`DELETE FROM todo_items WHERE id=?`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (h *SQLiteHandler) loadTodoItem(id int) (TodoItem, error) {
	var v TodoItem
	var completed int
	err := h.db.QueryRow(`SELECT id,title,start_at,completed,completed_at,created_at,updated_at FROM todo_items WHERE id=?`, id).
		Scan(&v.ID, &v.Title, &v.StartAt, &completed, &v.CompletedAt, &v.CreatedAt, &v.UpdatedAt)
	v.Completed = completed == 1
	return v, err
}

func summarizeNote(content string) string {
	for _, rawLine := range strings.Split(content, "\n") {
		line := cleanNoteSummaryLine(rawLine)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		for _, prefix := range []string{"标题:", "标题：", "title:", "主题:", "主题："} {
			if strings.HasPrefix(lower, strings.ToLower(prefix)) {
				title := strings.TrimSpace(line[len(prefix):])
				if title != "" {
					return truncateRunes(title, 28)
				}
			}
		}
		if len([]rune(line)) <= 28 && !endsWithSentencePunctuation(line) {
			return line
		}
		break
	}

	text := strings.Join(strings.Fields(strings.ReplaceAll(content, "\n", " ")), " ")
	text = strings.TrimSpace(text)
	if text == "" {
		return "空笔记"
	}
	for _, sep := range []string{"。", "！", "？", "；", ".", "!", "?", ";"} {
		if idx := strings.Index(text, sep); idx > 0 {
			text = strings.TrimSpace(text[:idx])
			break
		}
	}
	text = cleanNoteSummaryLine(text)
	if text == "" {
		return "未命名笔记"
	}
	return truncateRunes(text, 36)
}

func cleanNoteSummaryLine(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimLeft(value, "#>-*+ \t")
	value = strings.TrimSpace(value)
	value = strings.Trim(value, " \t\r\n\"'“”‘’`")
	return strings.Join(strings.Fields(value), " ")
}

func endsWithSentencePunctuation(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	for _, suffix := range []string{"。", "！", "？", "；", ".", "!", "?", ";"} {
		if strings.HasSuffix(value, suffix) {
			return true
		}
	}
	return false
}

func truncateRunes(value string, limit int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= limit {
		return string(runes)
	}
	if limit <= 1 {
		return string(runes[:limit])
	}
	return string(runes[:limit-1]) + "…"
}

func (h *SQLiteHandler) replace(table string, ids []int, upsert func(*sql.Tx) error) error {
	tx, err := h.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = h.replaceTx(tx, table, ids, upsert); err != nil {
		return err
	}
	return tx.Commit()
}

func (h *SQLiteHandler) replaceTx(tx *sql.Tx, table string, ids []int, upsert func(*sql.Tx) error) error {
	if err := upsert(tx); err != nil {
		return err
	}
	return h.deleteRowsNotInTx(tx, table, ids)
}

func (h *SQLiteHandler) deleteRowsNotInTx(tx *sql.Tx, table string, ids []int) error {
	query := `DELETE FROM ` + table
	if len(ids) > 0 {
		query += ` WHERE id NOT IN (`
		args := make([]any, len(ids))
		for i, id := range ids {
			if i > 0 {
				query += `,`
			}
			query += `?`
			args[i] = id
		}
		query += `)`
		_, err := tx.Exec(query, args...)
		return err
	} else {
		_, err := tx.Exec(query)
		return err
	}
}

func idsOfExercises(v []Exercise) []int {
	r := make([]int, len(v))
	for i, x := range v {
		r[i] = x.ID
	}
	return r
}
func idsOfGroups(v []ExerciseGroup) []int {
	r := make([]int, len(v))
	for i, x := range v {
		r[i] = x.ID
	}
	return r
}
func idsOfSessions(v []TrainingSession) []int {
	r := make([]int, len(v))
	for i, x := range v {
		r[i] = x.SessionID
	}
	return r
}
func idsOfRecords(v []TrainingRecord) []int {
	r := make([]int, len(v))
	for i, x := range v {
		r[i] = x.RecordID
	}
	return r
}
