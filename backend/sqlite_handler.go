package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
CREATE INDEX IF NOT EXISTS idx_sessions_date_group ON training_sessions(training_date, group_id);
CREATE INDEX IF NOT EXISTS idx_records_exercise_session ON training_records(exercise_id, session_id);
CREATE INDEX IF NOT EXISTS idx_group_items_exercise ON exercise_group_items(exercise_id);
CREATE INDEX IF NOT EXISTS idx_weight_date ON weight_records(record_date DESC);
`
	if _, err := h.db.Exec(schema); err != nil {
		return fmt.Errorf("initialize sqlite schema: %w", err)
	}
	return h.importCSVOnce()
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
	if err = h.replaceTx(tx, "training_records", idsOfRecords(records), func(tx *sql.Tx) error {
		for _, v := range records {
			if _, err := tx.Exec(`INSERT INTO training_records(id,session_id,exercise_id,set_number,weight,reps,duration_seconds,note) VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET session_id=excluded.session_id,exercise_id=excluded.exercise_id,set_number=excluded.set_number,weight=excluded.weight,reps=excluded.reps,duration_seconds=excluded.duration_seconds,note=excluded.note`, v.RecordID, v.SessionID, v.ExerciseID, v.SetNumber, v.Weight, v.Reps, v.Duration, v.Note); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
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
