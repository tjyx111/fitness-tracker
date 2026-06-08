package main

import (
	"encoding/csv"
	"os"
	"strconv"
)

// WeightRecordsHandler 体重记录处理器
type WeightRecordsHandler struct {
	weightFile string
}

func NewWeightRecordsHandler(dataDir string) *WeightRecordsHandler {
	return &WeightRecordsHandler{
		weightFile: dataDir + "/weight_records.csv",
	}
}

// LoadWeightRecords 加载体重记录
func (w *WeightRecordsHandler) LoadWeightRecords() ([]WeightRecord, error) {
	file, err := os.Open(w.weightFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 创建文件并写入标题
			return w.createWeightFile()
		}
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var weightRecords []WeightRecord
	for i, record := range records {
		if i == 0 {
			continue // 跳过标题行
		}
		if len(record) < 3 {
			continue
		}

		id, _ := strconv.Atoi(record[0])
		weight, _ := strconv.ParseFloat(record[2], 64)
		bodyFat := 0.0
		if len(record) > 3 {
			bodyFat, _ = strconv.ParseFloat(record[3], 64)
		}
		note := ""
		if len(record) > 4 {
			note = record[4]
		}

		weightRecords = append(weightRecords, WeightRecord{
			ID:      id,
			Date:    record[1],
			Weight:  weight,
			BodyFat: bodyFat,
			Note:    note,
		})
	}

	return weightRecords, nil
}

// SaveWeightRecords 保存体重记录
func (w *WeightRecordsHandler) SaveWeightRecords(records []WeightRecord) error {
	file, err := os.Create(w.weightFile)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入标题
	writer.Write([]string{"id(编号)", "date(日期)", "weight(体重kg)", "bodyFat(体脂率%)", "note(备注)"})

	// 写入数据
	for _, r := range records {
		writer.Write([]string{
			strconv.Itoa(r.ID),
			r.Date,
			strconv.FormatFloat(r.Weight, 'f', 2, 64),
			strconv.FormatFloat(r.BodyFat, 'f', 2, 64),
			r.Note,
		})
	}

	return nil
}

// AddWeightRecord 添加体重记录
func (w *WeightRecordsHandler) AddWeightRecord(record WeightRecord) error {
	records, err := w.LoadWeightRecords()
	if err != nil {
		return err
	}

	// 生成新ID
	maxID := 0
	for _, r := range records {
		if r.ID > maxID {
			maxID = r.ID
		}
	}
	record.ID = maxID + 1

	records = append(records, record)
	return w.SaveWeightRecords(records)
}

// GetLatestWeight 获取最新体重
func (w *WeightRecordsHandler) GetLatestWeight() (float64, error) {
	records, err := w.LoadWeightRecords()
	if err != nil {
		return 0, err
	}

	if len(records) == 0 {
		return 0, nil
	}

	// 按日期排序，返回最新的
	latest := records[0]
	for _, r := range records {
		if r.Date > latest.Date {
			latest = r
		}
	}

	return latest.Weight, nil
}

func (w *WeightRecordsHandler) createWeightFile() ([]WeightRecord, error) {
	file, err := os.Create(w.weightFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入标题
	writer.Write([]string{"id(编号)", "date(日期)", "weight(体重kg)", "bodyFat(体脂率%)", "note(备注)"})

	return []WeightRecord{}, nil
}
