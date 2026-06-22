package main

import (
	"encoding/csv"
	"os"
	"strconv"
)

// WeightRecordsHandler keeps the existing call sites small while all weight
// records are persisted by the shared SQLite connection.
type WeightRecordsHandler struct{ store *SQLiteHandler }

func NewWeightRecordsHandler(store *SQLiteHandler) *WeightRecordsHandler {
	return &WeightRecordsHandler{store: store}
}

func (w *WeightRecordsHandler) LoadWeightRecords() ([]WeightRecord, error) {
	return w.store.LoadWeightRecords()
}

func (w *WeightRecordsHandler) AddWeightRecord(record *WeightRecord) error {
	return w.store.AddWeightRecord(record)
}

func (w *WeightRecordsHandler) GetLatestWeight() (float64, error) {
	return w.store.GetLatestWeight()
}

// loadLegacyWeightRecords is used only during the one-time CSV migration.
func loadLegacyWeightRecords(path string) ([]WeightRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []WeightRecord{}, nil
		}
		return nil, err
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, err
	}
	result := []WeightRecord{}
	for i, row := range rows {
		if i == 0 || len(row) < 3 {
			continue
		}
		id, idErr := strconv.Atoi(row[0])
		weight, weightErr := strconv.ParseFloat(row[2], 64)
		if idErr != nil || weightErr != nil || weight <= 0 {
			continue
		}
		bodyFat := 0.0
		if len(row) > 3 {
			bodyFat, _ = strconv.ParseFloat(row[3], 64)
		}
		note := ""
		if len(row) > 4 {
			note = row[4]
		}
		result = append(result, WeightRecord{ID: id, Date: row[1], Weight: weight, BodyFat: bodyFat, Note: note})
	}
	return result, nil
}
