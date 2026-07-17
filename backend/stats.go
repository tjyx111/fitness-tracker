package main

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// WeightRecord 体重记录
type WeightRecord struct {
	ID      int     `json:"id"`
	Date    string  `json:"date"`
	Weight  float64 `json:"weight"`  // kg
	BodyFat float64 `json:"bodyFat"` // 体脂率(%)
	Note    string  `json:"note"`
}

// VolumeStats 容量统计
type VolumeStats struct {
	TotalVolume               float64            `json:"totalVolume"`               // 总训练容量(kg)
	AverageVolume             float64            `json:"averageVolume"`             // 平均容量
	MaxVolume                 float64            `json:"maxVolume"`                 // 最大容量
	MinVolume                 float64            `json:"minVolume"`                 // 最小容量
	VolumeGrowthRate          float64            `json:"volumeGrowthRate"`          // 容量增长率(%)
	VolumeGrowthPercentile    float64            `json:"volumeGrowthPercentile"`    // 增长率使用的分位数
	RecentPeriodVolumeIndex   float64            `json:"recentPeriodVolumeIndex"`   // 最近半周期容量分位指数
	PreviousPeriodVolumeIndex float64            `json:"previousPeriodVolumeIndex"` // 上一个半周期容量分位指数
	DailyVolumes              map[string]float64 `json:"dailyVolumes"`              // 每日容量
	WeeklyVolumes             map[string]float64 `json:"weeklyVolumes"`             // 每周容量
	MonthlyVolumes            map[string]float64 `json:"monthlyVolumes"`            // 每月容量
}

// IntensityStats 强度统计
type IntensityStats struct {
	AverageIntensity  float64         `json:"averageIntensity"`  // 平均强度(kg)
	MaxIntensity      float64         `json:"maxIntensity"`      // 最大强度
	IntensityTrend    []float64       `json:"intensityTrend"`    // 强度趋势
	RelativeIntensity float64         `json:"relativeIntensity"` // 相对强度(容量/体重)
	Estimated1RM      map[int]float64 `json:"estimated1RM"`      // 估算1RM(动作ID->1RM)
}

// ProgressRateStats 进度速度统计
type ProgressRateStats struct {
	WeightGrowthRate float64   `json:"weightGrowthRate"` // 重量增长率(%/周)
	RepsGrowthRate   float64   `json:"repsGrowthRate"`   // 次数增长率(%/周)
	VolumeGrowthRate float64   `json:"volumeGrowthRate"` // 容量增长率(%/周)
	TimeToTarget     int       `json:"timeToTarget"`     // 预计达成目标时间(周)
	WeeklyProgress   []float64 `json:"weeklyProgress"`   // 每周进度
}

// PersonalRecord 个人记录
type PersonalRecord struct {
	ExerciseID    int     `json:"exerciseId"`
	ExerciseName  string  `json:"exerciseName"`
	MaxWeight     float64 `json:"maxWeight"`     // 最大重量(kg)
	MaxVolume     float64 `json:"maxVolume"`     // 最大容量(kg)
	MaxReps       int     `json:"maxReps"`       // 某重量下的最大次数
	MaxDuration   int     `json:"maxDuration"`   // 最长持续时间(秒)
	MaxWeightDate string  `json:"maxWeightDate"` // 最大重量日期
	MaxVolumeDate string  `json:"maxVolumeDate"` // 最大容量日期
	AchievedAt    string  `json:"achievedAt"`    // 达成时间
}

// TrainingFrequency 训练频率统计
type TrainingFrequency struct {
	WeeklyFrequency  int             `json:"weeklyFrequency"`  // 本周训练次数
	MuscleGroupFreq  map[string]int  `json:"muscleGroupFreq"`  // 各肌群训练频率
	TrainingStreak   int             `json:"trainingStreak"`   // 连续训练周数
	AverageRestDays  float64         `json:"averageRestDays"`  // 平均休息天数
	TrainingCalendar map[string]bool `json:"trainingCalendar"` // 训练日历(日期->是否训练)
}

// ComprehensiveStats 综合统计
type ComprehensiveStats struct {
	TrainingQualityScore float64  `json:"trainingQualityScore"` // 训练质量分(0-100)
	ProgressiveOverload  float64  `json:"progressiveOverload"`  // 渐进超负荷指数
	FatigueIndex         float64  `json:"fatigueIndex"`         // 疲劳指数
	BalanceCoefficient   float64  `json:"balanceCoefficient"`   // 平衡系数(0-1)
	OverallProgress      float64  `json:"overallProgress"`      // 总体进度(%)
	WeakMuscleGroups     []string `json:"weakMuscleGroups"`     // 薄弱肌群
	Recommendations      []string `json:"recommendations"`      // 训练建议
}

// DailyStats 每日统计
type DailyStats struct {
	Date              string          `json:"date"`
	TotalVolume       float64         `json:"totalVolume"`
	TotalSets         int             `json:"totalSets"`
	ExercisesCount    int             `json:"exercisesCount"`
	Duration          int             `json:"duration"`          // 训练时长(分钟)
	ExerciseBreakdown map[int]float64 `json:"exerciseBreakdown"` // 各动作容量
}

// StatsReport 统计报告
type StatsReport struct {
	StartDate         string              `json:"startDate"`
	EndDate           string              `json:"endDate"`
	VolumeStats       *VolumeStats        `json:"volumeStats"`
	IntensityStats    *IntensityStats     `json:"intensityStats"`
	ProgressRate      *ProgressRateStats  `json:"progressRate"`
	PersonalRecords   []PersonalRecord    `json:"personalRecords"`
	TrainingFrequency *TrainingFrequency  `json:"trainingFrequency"`
	Comprehensive     *ComprehensiveStats `json:"comprehensive"`
	DailyStats        []DailyStats        `json:"dailyStats"`
	WeightRecords     []WeightRecord      `json:"weightRecords"`
	GeneratedAt       string              `json:"generatedAt"`
}

// ExerciseDailyStats 单个动作每日统计
type ExerciseDailyStats struct {
	Date          string  `json:"date"`
	MaxWeight     float64 `json:"maxWeight"`     // 最大重量
	TotalVolume   float64 `json:"totalVolume"`   // 总容量 (重量×次数)
	TotalReps     int     `json:"totalReps"`     // 总次数
	AvgWeight     float64 `json:"avgWeight"`     // 平均重量
	MaxDuration   int     `json:"maxDuration"`   // 最长持续时间(秒)
	TotalDuration int     `json:"totalDuration"` // 总持续时间
	AvgDuration   float64 `json:"avgDuration"`   // 平均持续时间
	SetsCount     int     `json:"setsCount"`     // 组数
}

// ExerciseHistoryStats 单个动作历史统计
type ExerciseHistoryStats struct {
	ExerciseID      int                  `json:"exerciseId"`
	ExerciseName    string               `json:"exerciseName"`
	Unit            string               `json:"unit"`
	MuscleGroup     string               `json:"muscleGroup"`
	DailyStats      []ExerciseDailyStats `json:"dailyStats"`
	BestPerformance ExerciseDailyStats   `json:"bestPerformance"`
	OverallStats    struct {
		TotalSets    int     `json:"totalSets"`
		TotalVolume  float64 `json:"totalVolume"`
		AvgVolume    float64 `json:"avgVolume"`
		MaxWeight    float64 `json:"maxWeight"`
		MaxDuration  int     `json:"maxDuration"`
		TrainingDays int     `json:"trainingDays"`
	} `json:"overallStats"`
}

// StatsAnalyzer 统计分析器
type StatsAnalyzer struct {
	csv *SQLiteHandler
}

func NewStatsAnalyzer(csv *SQLiteHandler) *StatsAnalyzer {
	return &StatsAnalyzer{csv: csv}
}

// CalculateVolume 计算单次训练容量
func (s *StatsAnalyzer) CalculateVolume(records []TrainingRecord) float64 {
	total := 0.0
	for _, r := range records {
		total += r.Weight * float64(r.Reps)
	}
	return total
}

// Estimate1RM 估算单次最大重量 (使用Epley公式)
func (s *StatsAnalyzer) Estimate1RM(weight float64, reps int) float64 {
	if reps == 1 {
		return weight
	}
	// Epley公式: 1RM = weight × (1 + reps/30)
	return weight * (1 + float64(reps)/30.0)
}

// CalculateIntensity 计算训练强度
func (s *StatsAnalyzer) CalculateIntensity(records []TrainingRecord) float64 {
	if len(records) == 0 {
		return 0
	}
	total := 0.0
	count := 0
	for _, r := range records {
		if r.Weight > 0 {
			total += r.Weight
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

const defaultVolumeGrowthPercentile = 95.0

// GetVolumeStats 获取容量统计
func (s *StatsAnalyzer) GetVolumeStats(days int, growthPercentile float64) (*VolumeStats, error) {
	growthPercentile = normalizePercentile(growthPercentile)

	records, _ := s.csv.LoadTrainingRecords()
	sessions, _ := s.csv.LoadTrainingSessions()

	// 按日期分组
	dailyVolumes := make(map[string]float64)
	weeklyVolumes := make(map[string]float64)
	monthlyVolumes := make(map[string]float64)

	for _, session := range sessions {
		if !isSessionWithinDays(session.Date, days) {
			continue
		}

		sessionRecords := filterRecordsBySession(records, session.SessionID)
		volume := s.CalculateVolume(sessionRecords)

		dailyVolumes[session.Date] += volume

		// 计算周和月
		date, _ := time.Parse("2006-01-02", session.Date)
		week := date.Format("2006-W01")
		month := date.Format("2006-01")
		weeklyVolumes[week] += volume
		monthlyVolumes[month] += volume
	}

	// 计算统计指标
	dates := make([]string, 0, len(dailyVolumes))
	for date := range dailyVolumes {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	volumes := make([]float64, 0, len(dates))
	for _, date := range dates {
		volumes = append(volumes, dailyVolumes[date])
	}

	if len(volumes) == 0 {
		return &VolumeStats{}, nil
	}

	totalVolume := sum(volumes)
	averageVolume := totalVolume / float64(len(volumes))
	maxVolume := max(volumes)
	minVolume := min(volumes)

	// 计算增长率：按日期拆成前后半周期，使用指定分位数代替均值。
	volumeGrowthRate := 0.0
	recentPeriodVolumeIndex := 0.0
	previousPeriodVolumeIndex := 0.0
	if len(volumes) >= 2 {
		recentPeriodVolumeIndex = percentile(volumes[len(volumes)/2:], growthPercentile)
		previousPeriodVolumeIndex = percentile(volumes[:len(volumes)/2], growthPercentile)
		if previousPeriodVolumeIndex > 0 {
			volumeGrowthRate = ((recentPeriodVolumeIndex - previousPeriodVolumeIndex) / previousPeriodVolumeIndex) * 100
		}
	}

	return &VolumeStats{
		TotalVolume:               totalVolume,
		AverageVolume:             averageVolume,
		MaxVolume:                 maxVolume,
		MinVolume:                 minVolume,
		VolumeGrowthRate:          volumeGrowthRate,
		VolumeGrowthPercentile:    growthPercentile,
		RecentPeriodVolumeIndex:   recentPeriodVolumeIndex,
		PreviousPeriodVolumeIndex: previousPeriodVolumeIndex,
		DailyVolumes:              dailyVolumes,
		WeeklyVolumes:             weeklyVolumes,
		MonthlyVolumes:            monthlyVolumes,
	}, nil
}

// GetIntensityStats 获取强度统计
func (s *StatsAnalyzer) GetIntensityStats(days int, currentWeight float64) (*IntensityStats, error) {
	records, _ := s.csv.LoadTrainingRecords()
	sessions, _ := s.csv.LoadTrainingSessions()

	estimated1RM := make(map[int]float64)
	dailyIntensity := make(map[string]float64)
	totalVolume := 0.0
	totalWeight := 0.0
	weightCount := 0

	for _, session := range sessions {
		if !isSessionWithinDays(session.Date, days) {
			continue
		}

		sessionRecords := filterRecordsBySession(records, session.SessionID)
		sessionIntensity := s.CalculateIntensity(sessionRecords)
		dailyIntensity[session.Date] = sessionIntensity

		// 计算总容量和强度
		volume := s.CalculateVolume(sessionRecords)
		totalVolume += volume

		for _, r := range sessionRecords {
			if r.Weight > 0 {
				totalWeight += r.Weight
				weightCount++

				// 估算1RM
				current1RM := s.Estimate1RM(r.Weight, r.Reps)
				if current1RM > estimated1RM[r.ExerciseID] {
					estimated1RM[r.ExerciseID] = current1RM
				}
			}
		}
	}

	if weightCount == 0 {
		return &IntensityStats{}, nil
	}

	avgIntensity := totalWeight / float64(weightCount)
	maxIntensity := 0.0
	for _, intensity := range dailyIntensity {
		if intensity > maxIntensity {
			maxIntensity = intensity
		}
	}

	// 相对强度
	relativeIntensity := 0.0
	if currentWeight > 0 {
		relativeIntensity = totalVolume / currentWeight
	}

	// 强度趋势
	intensities := make([]float64, 0, len(dailyIntensity))
	for _, intensity := range dailyIntensity {
		intensities = append(intensities, intensity)
	}
	sort.Float64s(intensities)

	return &IntensityStats{
		AverageIntensity:  avgIntensity,
		MaxIntensity:      maxIntensity,
		IntensityTrend:    intensities,
		RelativeIntensity: relativeIntensity,
		Estimated1RM:      estimated1RM,
	}, nil
}

// GetProgressRate 获取进度速度
func (s *StatsAnalyzer) GetProgressRate(exerciseID int, targetWeight float64) (*ProgressRateStats, error) {
	records, _ := s.csv.LoadTrainingRecords()

	// 筛选该动作的记录
	var exerciseRecords []TrainingRecord
	for _, r := range records {
		if r.ExerciseID == exerciseID {
			exerciseRecords = append(exerciseRecords, r)
		}
	}

	if len(exerciseRecords) == 0 {
		return &ProgressRateStats{}, nil
	}

	// 按时间排序
	sort.Slice(exerciseRecords, func(i, j int) bool {
		return exerciseRecords[i].RecordID < exerciseRecords[j].RecordID
	})

	// 计算增长率
	firstWeight := exerciseRecords[0].Weight
	firstReps := exerciseRecords[0].Reps
	lastWeight := exerciseRecords[len(exerciseRecords)-1].Weight
	lastReps := exerciseRecords[len(exerciseRecords)-1].Reps

	weightGrowthRate := 0.0
	if firstWeight > 0 {
		weightGrowthRate = ((lastWeight - firstWeight) / firstWeight) * 100
	}

	repsGrowthRate := 0.0
	if firstReps > 0 {
		repsGrowthRate = ((float64(lastReps) - float64(firstReps)) / float64(firstReps)) * 100
	}

	// 计算容量增长率
	firstVolume := firstWeight * float64(firstReps)
	lastVolume := lastWeight * float64(lastReps)
	volumeGrowthRate := 0.0
	if firstVolume > 0 {
		volumeGrowthRate = ((lastVolume - firstVolume) / firstVolume) * 100
	}

	// 预计达成目标时间
	timeToTarget := 0
	if targetWeight > 0 && lastWeight > 0 && weightGrowthRate > 0 {
		weeksNeeded := (targetWeight - lastWeight) / (lastWeight * weightGrowthRate / 100)
		timeToTarget = int(math.Ceil(weeksNeeded))
	}

	return &ProgressRateStats{
		WeightGrowthRate: weightGrowthRate,
		RepsGrowthRate:   repsGrowthRate,
		VolumeGrowthRate: volumeGrowthRate,
		TimeToTarget:     timeToTarget,
	}, nil
}

// GetPersonalRecords 获取个人记录
func (s *StatsAnalyzer) GetPersonalRecords() ([]PersonalRecord, error) {
	records, _ := s.csv.LoadTrainingRecords()
	sessions, _ := s.csv.LoadTrainingSessions()
	exercises, _ := s.csv.LoadExercises()

	// 按动作分组
	exercisePRs := make(map[int]PersonalRecord)

	for _, r := range records {
		// 过滤掉id=0的无效动作
		if r.ExerciseID <= 0 {
			continue
		}

		if _, exists := exercisePRs[r.ExerciseID]; !exists {
			// 初始化
			var exerciseName string
			for _, e := range exercises {
				if e.ID == r.ExerciseID {
					exerciseName = e.Name
					break
				}
			}

			exercisePRs[r.ExerciseID] = PersonalRecord{
				ExerciseID:   r.ExerciseID,
				ExerciseName: exerciseName,
				MaxWeight:    r.Weight,
				MaxReps:      r.Reps,
				MaxDuration:  r.Duration,
			}
		}

		pr := exercisePRs[r.ExerciseID]

		// 更新最大重量
		if r.Weight > pr.MaxWeight {
			pr.MaxWeight = r.Weight
			// 查找日期
			for _, s := range sessions {
				if s.SessionID == r.SessionID {
					pr.MaxWeightDate = s.Date
					break
				}
			}
		}

		// 更新最大次数
		if r.Reps > pr.MaxReps {
			pr.MaxReps = r.Reps
		}

		// 更新最长持续时间
		if r.Duration > pr.MaxDuration {
			pr.MaxDuration = r.Duration
		}

		// 计算最大容量
		volume := r.Weight * float64(r.Reps)
		if volume > pr.MaxVolume {
			pr.MaxVolume = volume
			for _, s := range sessions {
				if s.SessionID == r.SessionID {
					pr.MaxVolumeDate = s.Date
					break
				}
			}
		}

		exercisePRs[r.ExerciseID] = pr
	}

	// 转换为切片
	var prs []PersonalRecord
	for _, pr := range exercisePRs {
		prs = append(prs, pr)
	}

	sort.Slice(prs, func(i, j int) bool {
		return prs[i].ExerciseID < prs[j].ExerciseID
	})

	return prs, nil
}

// GetTrainingFrequency 获取训练频率
func (s *StatsAnalyzer) GetTrainingFrequency() (*TrainingFrequency, error) {
	sessions, _ := s.csv.LoadTrainingSessions()
	records, _ := s.csv.LoadTrainingRecords()
	exercises, _ := s.csv.LoadExercises()

	// 训练日历
	trainingCalendar := make(map[string]bool)
	muscleGroupFreq := make(map[string]int)

	// 计算本周训练频率
	now := time.Now()
	weekStart := beginningOfWeek(now)
	weeklyFrequency := 0

	for _, session := range sessions {
		trainingCalendar[session.Date] = true

		sessionDate, _ := time.Parse("2006-01-02", session.Date)
		if sessionDate.After(weekStart) || sessionDate.Equal(weekStart) {
			weeklyFrequency++
		}

		// 统计肌群频率
		sessionRecords := filterRecordsBySession(records, session.SessionID)
		for _, r := range sessionRecords {
			for _, e := range exercises {
				if e.ID == r.ExerciseID && e.ID > 0 { // 过滤掉id=0的标题行
					muscleGroupFreq[e.MuscleGroup]++
					break
				}
			}
		}
	}

	// 计算连续训练周数
	trainingStreak := s.calculateTrainingStreak(sessions)

	// 计算平均休息天数
	averageRestDays := s.calculateAverageRestDays(sessions)

	return &TrainingFrequency{
		WeeklyFrequency:  weeklyFrequency,
		MuscleGroupFreq:  muscleGroupFreq,
		TrainingStreak:   trainingStreak,
		AverageRestDays:  averageRestDays,
		TrainingCalendar: trainingCalendar,
	}, nil
}

// GetComprehensiveStats 获取综合统计
func (s *StatsAnalyzer) GetComprehensiveStats(days int, growthPercentile float64) (*ComprehensiveStats, error) {
	volumeStats, _ := s.GetVolumeStats(days, growthPercentile)
	intensityStats, _ := s.GetIntensityStats(days, 0)
	frequency, _ := s.GetTrainingFrequency()

	// 训练质量分 (0-100)
	qualityScore := 0.0
	if volumeStats.VolumeGrowthRate > 0 {
		qualityScore += math.Min(volumeStats.VolumeGrowthRate, 50) * 0.4
	}
	if intensityStats.AverageIntensity > 0 {
		qualityScore += math.Min(intensityStats.AverageIntensity, 100) * 0.3
	}
	qualityScore += float64(frequency.WeeklyFrequency) * 5
	qualityScore = math.Min(qualityScore, 100)

	// 渐进超负荷指数
	progressiveOverload := volumeStats.VolumeGrowthRate

	// 疲劳指数
	fatigueIndex := 0.0
	if progressiveOverload < 0 {
		fatigueIndex = math.Abs(progressiveOverload)
	}

	// 平衡系数 (各肌群训练容量的标准差)
	balanceCoefficient := s.calculateBalanceCoefficient(frequency.MuscleGroupFreq)

	// 总体进度
	overallProgress := math.Max(0, math.Min(100, volumeStats.VolumeGrowthRate*2))

	// 薄弱肌群
	weakMuscleGroups := s.identifyWeakMuscleGroups(frequency.MuscleGroupFreq)

	// 训练建议
	recommendations := s.generateRecommendations(qualityScore, progressiveOverload, fatigueIndex, balanceCoefficient)

	return &ComprehensiveStats{
		TrainingQualityScore: qualityScore,
		ProgressiveOverload:  progressiveOverload,
		FatigueIndex:         fatigueIndex,
		BalanceCoefficient:   balanceCoefficient,
		OverallProgress:      overallProgress,
		WeakMuscleGroups:     weakMuscleGroups,
		Recommendations:      recommendations,
	}, nil
}

// 辅助函数
func filterRecordsBySession(records []TrainingRecord, sessionID int) []TrainingRecord {
	var filtered []TrainingRecord
	for _, r := range records {
		if r.SessionID == sessionID {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func sum(values []float64) float64 {
	total := 0.0
	for _, v := range values {
		total += v
	}
	return total
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return sum(values) / float64(len(values))
}

func normalizePercentile(value float64) float64 {
	if value <= 0 {
		return defaultVolumeGrowthPercentile
	}
	if value > 100 {
		return 100
	}
	return value
}

func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}

	p = normalizePercentile(p)
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)

	if len(sorted) == 1 {
		return sorted[0]
	}

	position := (p / 100) * float64(len(sorted)-1)
	lower := int(math.Floor(position))
	upper := int(math.Ceil(position))
	if lower == upper {
		return sorted[lower]
	}

	weight := position - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	middle := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[middle]
	}
	return (sorted[middle-1] + sorted[middle]) / 2
}

func max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	m := values[0]
	for _, v := range values {
		if v > m {
			m = v
		}
	}
	return m
}

func min(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	m := values[0]
	for _, v := range values {
		if v < m {
			m = v
		}
	}
	return m
}

func (s *StatsAnalyzer) calculateTrainingStreak(sessions []TrainingSession) int {
	if len(sessions) == 0 {
		return 0
	}

	trainedWeeks := make(map[string]bool)
	for _, session := range sessions {
		date, err := time.Parse("2006-01-02", session.Date)
		if err != nil {
			continue
		}
		year, week := date.ISOWeek()
		trainedWeeks[formatISOWeek(year, week)] = true
	}

	streak := 0
	weekStart := beginningOfWeek(time.Now())
	for {
		year, week := weekStart.ISOWeek()
		if !trainedWeeks[formatISOWeek(year, week)] {
			break
		}
		streak++
		weekStart = weekStart.AddDate(0, 0, -7)
	}

	return streak
}

func (s *StatsAnalyzer) calculateAverageRestDays(sessions []TrainingSession) float64 {
	if len(sessions) < 2 {
		return 0
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Date < sessions[j].Date
	})

	totalRestDays := 0.0
	for i := 1; i < len(sessions); i++ {
		date1, _ := time.Parse("2006-01-02", sessions[i-1].Date)
		date2, _ := time.Parse("2006-01-02", sessions[i].Date)
		diff := date2.Sub(date1).Hours() / 24
		totalRestDays += diff
	}

	return totalRestDays / float64(len(sessions)-1)
}

func (s *StatsAnalyzer) calculateBalanceCoefficient(muscleFreq map[string]int) float64 {
	if len(muscleFreq) == 0 {
		return 0
	}

	values := make([]float64, 0, len(muscleFreq))
	for _, v := range muscleFreq {
		values = append(values, float64(v))
	}

	avg := average(values)
	if avg == 0 {
		return 0
	}

	// 计算标准差
	variance := 0.0
	for _, v := range values {
		diff := v - avg
		variance += diff * diff
	}
	variance /= float64(len(values))
	stdDev := math.Sqrt(variance)

	// 转换为0-1的系数 (越接近1越均衡)
	balance := 1 - math.Min(stdDev/avg, 1)
	return balance
}

func (s *StatsAnalyzer) identifyWeakMuscleGroups(muscleFreq map[string]int) []string {
	if len(muscleFreq) == 0 {
		return []string{}
	}

	avg := 0.0
	for _, v := range muscleFreq {
		avg += float64(v)
	}
	avg /= float64(len(muscleFreq))

	var weak []string
	for muscle, freq := range muscleFreq {
		if float64(freq) < avg*0.7 {
			weak = append(weak, muscle)
		}
	}

	return weak
}

func (s *StatsAnalyzer) generateRecommendations(qualityScore, progressiveOverload, fatigueIndex, balanceCoefficient float64) []string {
	var recommendations []string

	if qualityScore < 50 {
		recommendations = append(recommendations, "训练质量偏低，建议增加训练强度或频率")
	}

	if progressiveOverload < 0 {
		recommendations = append(recommendations, "训练容量下降，可能需要调整训练计划或增加休息")
	}

	if fatigueIndex > 20 {
		recommendations = append(recommendations, "疲劳指数较高，建议增加休息时间或减少训练量")
	}

	if balanceCoefficient < 0.6 {
		recommendations = append(recommendations, "肌群训练不平衡，建议加强薄弱肌群的训练")
	}

	if qualityScore > 80 && progressiveOverload > 10 {
		recommendations = append(recommendations, "训练进展良好！继续保持并可以尝试新的挑战")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "训练状态良好，继续保持！")
	}

	return recommendations
}

// beginningOfWeek 辅助函数
func beginningOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	return t.AddDate(0, 0, -int(weekday)+1)
}

// GetExerciseHistory 获取单个动作的历史统计数据
func (s *StatsAnalyzer) GetExerciseHistory(exerciseID int, days int) (*ExerciseHistoryStats, error) {
	records, _ := s.csv.LoadTrainingRecords()
	sessions, _ := s.csv.LoadTrainingSessions()
	exercises, _ := s.csv.LoadExercises()
	endDate := statsEndDate(sessions)

	// 查找动作信息
	var exercise *Exercise
	for _, e := range exercises {
		if e.ID == exerciseID {
			exercise = &e
			break
		}
	}
	if exercise == nil {
		return &ExerciseHistoryStats{}, nil
	}

	// 筛选该动作的记录
	var exerciseRecords []TrainingRecord
	for _, r := range records {
		if r.ExerciseID == exerciseID {
			exerciseRecords = append(exerciseRecords, r)
		}
	}

	// 按日期分组计算每日统计
	dailyStatsMap := make(map[string]*ExerciseDailyStats)

	for _, session := range sessions {
		// 获取该会话中该动作的记录
		sessionRecords := filterRecordsBySession(exerciseRecords, session.SessionID)
		if len(sessionRecords) == 0 {
			continue
		}

		dayStats := &ExerciseDailyStats{
			Date: session.Date,
		}

		// 根据单位类型计算不同的指标
		if exercise.Unit == "kg" {
			// 重量类型动作
			for _, r := range sessionRecords {
				dayStats.TotalVolume += r.Weight * float64(r.Reps)
				dayStats.TotalReps += r.Reps
				dayStats.SetsCount++
				if r.Weight > dayStats.MaxWeight {
					dayStats.MaxWeight = r.Weight
				}
			}
			if dayStats.SetsCount > 0 {
				dayStats.AvgWeight = dayStats.TotalVolume / float64(dayStats.TotalReps)
			}
		} else {
			// 持续时间类型动作
			for _, r := range sessionRecords {
				dayStats.TotalDuration += r.Duration
				dayStats.TotalReps += r.Reps
				dayStats.SetsCount++
				if r.Duration > dayStats.MaxDuration {
					dayStats.MaxDuration = r.Duration
				}
			}
			if dayStats.SetsCount > 0 {
				dayStats.AvgDuration = float64(dayStats.TotalDuration) / float64(dayStats.SetsCount)
			}
		}

		dailyStatsMap[session.Date] = dayStats
	}

	// 填充缺失日期为0，确保连续的时间序列
	var dailyStats []ExerciseDailyStats
	for i := days - 1; i >= 0; i-- {
		date := endDate.AddDate(0, 0, -i).Format("2006-01-02")
		if stats, exists := dailyStatsMap[date]; exists {
			dailyStats = append(dailyStats, *stats)
		} else {
			// 填充空数据
			dailyStats = append(dailyStats, ExerciseDailyStats{
				Date:          date,
				MaxWeight:     0,
				TotalVolume:   0,
				TotalReps:     0,
				AvgWeight:     0,
				MaxDuration:   0,
				TotalDuration: 0,
				AvgDuration:   0,
				SetsCount:     0,
			})
		}
	}

	// 找出最佳表现
	var bestPerformance ExerciseDailyStats
	if len(dailyStats) > 0 {
		bestPerformance = dailyStats[0]
		for _, stats := range dailyStats {
			if exercise.Unit == "kg" {
				if stats.TotalVolume > bestPerformance.TotalVolume {
					bestPerformance = stats
				}
			} else {
				if stats.TotalDuration > bestPerformance.TotalDuration {
					bestPerformance = stats
				}
			}
		}
	}

	// 计算总体统计
	totalSets := 0
	totalVolume := 0.0
	maxWeight := 0.0
	maxDuration := 0
	trainingDays := 0
	for _, stats := range dailyStats {
		totalSets += stats.SetsCount
		totalVolume += stats.TotalVolume
		if stats.SetsCount > 0 {
			trainingDays++
		}
		if stats.MaxWeight > maxWeight {
			maxWeight = stats.MaxWeight
		}
		if stats.MaxDuration > maxDuration {
			maxDuration = stats.MaxDuration
		}
	}

	avgVolume := 0.0
	if trainingDays > 0 {
		avgVolume = totalVolume / float64(trainingDays)
	}

	return &ExerciseHistoryStats{
		ExerciseID:      exercise.ID,
		ExerciseName:    exercise.Name,
		Unit:            exercise.Unit,
		MuscleGroup:     exercise.MuscleGroup,
		DailyStats:      dailyStats,
		BestPerformance: bestPerformance,
		OverallStats: struct {
			TotalSets    int     `json:"totalSets"`
			TotalVolume  float64 `json:"totalVolume"`
			AvgVolume    float64 `json:"avgVolume"`
			MaxWeight    float64 `json:"maxWeight"`
			MaxDuration  int     `json:"maxDuration"`
			TrainingDays int     `json:"trainingDays"`
		}{
			TotalSets:    totalSets,
			TotalVolume:  totalVolume,
			AvgVolume:    avgVolume,
			MaxWeight:    maxWeight,
			MaxDuration:  maxDuration,
			TrainingDays: trainingDays,
		},
	}, nil
}

// OverviewDailyStats 总览每日统计
type OverviewDailyStats struct {
	Date        string  `json:"date"`
	TotalVolume float64 `json:"totalVolume"`
	TotalSets   int     `json:"totalSets"`
	MaxWeight   float64 `json:"maxWeight"`
	MaxDuration int     `json:"maxDuration"`
	TotalReps   int     `json:"totalReps"`
}

// OverviewHistoryStats 总览历史统计
type OverviewHistoryStats struct {
	DailyStats   []OverviewDailyStats `json:"dailyStats"`
	OverallStats struct {
		TotalVolume   float64 `json:"totalVolume"`
		AverageVolume float64 `json:"averageVolume"`
		MaxVolume     float64 `json:"maxVolume"`
		TotalSets     int     `json:"totalSets"`
		TrainingDays  int     `json:"trainingDays"`
	} `json:"overallStats"`
}

// TrainingCalendarDay 训练热力图单日数据
type TrainingCalendarDay struct {
	Date     string  `json:"date"`
	Trained  bool    `json:"trained"`
	Sets     int     `json:"sets"`
	Volume   float64 `json:"volume"`
	Reps     int     `json:"reps"`
	Duration int     `json:"duration"`
	Level    int     `json:"level"`
}

// TrainingCalendarStats 训练热力图数据
type TrainingCalendarStats struct {
	Days    []TrainingCalendarDay `json:"days"`
	Summary struct {
		TrainingDays int     `json:"trainingDays"`
		TotalSets    int     `json:"totalSets"`
		TotalVolume  float64 `json:"totalVolume"`
		MaxSets      int     `json:"maxSets"`
	} `json:"summary"`
}

// FilteredStats 按当前筛选条件汇总的统计页数据
type FilteredStats struct {
	OverallStats struct {
		TrainingDays int `json:"trainingDays"`
		TotalSets    int `json:"totalSets"`
	} `json:"overallStats"`
	BestPerformance struct {
		MaxSets struct {
			Date string `json:"date"`
			Sets int    `json:"sets"`
		} `json:"maxSets"`
		MaxWeight   SetPerformance `json:"maxWeight"`
		MaxReps     SetPerformance `json:"maxReps"`
		MaxDuration SetPerformance `json:"maxDuration"`
	} `json:"bestPerformance"`
	DailyScores []DailyMuscleScore `json:"dailyScores"`
}

// DailyMuscleScore 单日动作指数汇总分
type DailyMuscleScore struct {
	Date          string  `json:"date"`
	MuscleScore   float64 `json:"muscleScore"`
	RawValue      float64 `json:"rawValue"`
	TotalRawValue float64 `json:"totalRawValue"`
	RawUnit       string  `json:"rawUnit"`
}

// SetPerformance 单组最佳表现
type SetPerformance struct {
	Date         string  `json:"date"`
	SessionID    int     `json:"sessionId"`
	GroupID      int     `json:"groupId"`
	GroupName    string  `json:"groupName"`
	ExerciseID   int     `json:"exerciseId"`
	ExerciseName string  `json:"exerciseName"`
	MuscleGroup  string  `json:"muscleGroup"`
	Unit         string  `json:"unit"`
	SetNumber    int     `json:"setNumber"`
	Weight       float64 `json:"weight"`
	Reps         int     `json:"reps"`
	Duration     int     `json:"duration"`
	Value        float64 `json:"value"`
}

// DayRecordsStats 单日组明细
type DayRecordsStats struct {
	Date    string           `json:"date"`
	Records []SetPerformance `json:"records"`
}

// GetOverviewHistory 获取总览历史统计数据
func (s *StatsAnalyzer) GetOverviewHistory(days int) (*OverviewHistoryStats, error) {
	records, _ := s.csv.LoadTrainingRecords()
	sessions, _ := s.csv.LoadTrainingSessions()
	endDate := statsEndDate(sessions)

	// 按日期分组计算每日统计
	dailyStatsMap := make(map[string]*OverviewDailyStats)

	for _, session := range sessions {
		if !isSessionWithinDaysFromEnd(session.Date, days, endDate) {
			continue
		}

		sessionRecords := filterRecordsBySession(records, session.SessionID)
		if len(sessionRecords) == 0 {
			continue
		}

		dayStats := &OverviewDailyStats{
			Date: session.Date,
		}

		// 计算该日期的总容量和总组数
		for _, r := range sessionRecords {
			dayStats.TotalVolume += r.Weight * float64(r.Reps)
			dayStats.TotalSets++
			dayStats.TotalReps += r.Reps
			if r.Weight > dayStats.MaxWeight {
				dayStats.MaxWeight = r.Weight
			}
			if r.Duration > dayStats.MaxDuration {
				dayStats.MaxDuration = r.Duration
			}
		}

		dailyStatsMap[session.Date] = dayStats
	}

	// 填充缺失日期为0，确保连续的时间序列
	var dailyStats []OverviewDailyStats
	for i := days - 1; i >= 0; i-- {
		date := endDate.AddDate(0, 0, -i).Format("2006-01-02")
		if stats, exists := dailyStatsMap[date]; exists {
			dailyStats = append(dailyStats, *stats)
		} else {
			dailyStats = append(dailyStats, OverviewDailyStats{
				Date:        date,
				TotalVolume: 0,
				TotalSets:   0,
				MaxWeight:   0,
				MaxDuration: 0,
				TotalReps:   0,
			})
		}
	}

	// 计算总体统计
	totalVolume := 0.0
	totalSets := 0
	maxVolume := 0.0
	trainingDays := 0
	for _, stats := range dailyStats {
		totalVolume += stats.TotalVolume
		totalSets += stats.TotalSets
		if stats.TotalVolume > maxVolume {
			maxVolume = stats.TotalVolume
		}
		if stats.TotalSets > 0 {
			trainingDays++
		}
	}

	averageVolume := 0.0
	if trainingDays > 0 {
		averageVolume = totalVolume / float64(trainingDays)
	}

	return &OverviewHistoryStats{
		DailyStats: dailyStats,
		OverallStats: struct {
			TotalVolume   float64 `json:"totalVolume"`
			AverageVolume float64 `json:"averageVolume"`
			MaxVolume     float64 `json:"maxVolume"`
			TotalSets     int     `json:"totalSets"`
			TrainingDays  int     `json:"trainingDays"`
		}{
			TotalVolume:   totalVolume,
			AverageVolume: averageVolume,
			MaxVolume:     maxVolume,
			TotalSets:     totalSets,
			TrainingDays:  trainingDays,
		},
	}, nil
}

// GetTrainingCalendar 获取总览、单动作或单肌群的训练热力图数据
func (s *StatsAnalyzer) GetTrainingCalendar(days int, exerciseID int, muscleGroup string) (*TrainingCalendarStats, error) {
	if days <= 0 {
		days = 30
	}

	records, _ := s.csv.LoadTrainingRecords()
	sessions, _ := s.csv.LoadTrainingSessions()
	endDate := statsEndDate(sessions)

	exerciseMuscleGroups := make(map[int]string)
	if muscleGroup != "" {
		exercises, _ := s.csv.LoadExercises()
		for _, exercise := range exercises {
			exerciseMuscleGroups[exercise.ID] = exercise.MuscleGroup
		}
	}

	dailyStats := make(map[string]*TrainingCalendarDay)
	for _, session := range sessions {
		if !isSessionWithinDaysFromEnd(session.Date, days, endDate) {
			continue
		}

		sessionRecords := filterRecordsBySession(records, session.SessionID)
		for _, record := range sessionRecords {
			if exerciseID > 0 && record.ExerciseID != exerciseID {
				continue
			}
			if muscleGroup != "" && exerciseMuscleGroups[record.ExerciseID] != muscleGroup {
				continue
			}

			day := dailyStats[session.Date]
			if day == nil {
				day = &TrainingCalendarDay{Date: session.Date}
				dailyStats[session.Date] = day
			}

			day.Trained = true
			day.Sets++
			day.Volume += record.Weight * float64(record.Reps)
			day.Reps += record.Reps
			day.Duration += record.Duration
		}
	}

	stats := &TrainingCalendarStats{}
	for i := days - 1; i >= 0; i-- {
		date := endDate.AddDate(0, 0, -i).Format("2006-01-02")
		day := TrainingCalendarDay{Date: date}
		if existing, ok := dailyStats[date]; ok {
			day = *existing
			day.Level = calculateHeatmapLevel(day.Sets)
		}
		stats.Days = append(stats.Days, day)

		if day.Trained {
			stats.Summary.TrainingDays++
		}
		stats.Summary.TotalSets += day.Sets
		stats.Summary.TotalVolume += day.Volume
		if day.Sets > stats.Summary.MaxSets {
			stats.Summary.MaxSets = day.Sets
		}
	}

	return stats, nil
}

// GetFilteredStats 获取总览、单动作或单肌群统计页数据
func (s *StatsAnalyzer) GetFilteredStats(days int, exerciseID int, muscleGroup string, growthPercentile float64) (*FilteredStats, error) {
	if days <= 0 {
		days = 30
	}

	records, _ := s.csv.LoadTrainingRecords()
	sessions, _ := s.csv.LoadTrainingSessions()
	exercises, _ := s.csv.LoadExercises()
	exerciseMap := buildExerciseMap(exercises)
	sessionMap := buildSessionMap(sessions)
	endDate := statsEndDate(sessions)

	stats := &FilteredStats{}
	setsByDate := make(map[string]int)
	rawByDateExercise := make(map[string]map[int]float64)
	bestSetByDateExercise := make(map[string]map[int]float64)

	for _, record := range records {
		session, ok := sessionMap[record.SessionID]
		if !ok || !isSessionWithinDaysFromEnd(session.Date, days, endDate) {
			continue
		}
		exercise, ok := exerciseMap[record.ExerciseID]
		if !ok || !matchesStatsFilter(exercise, exerciseID, muscleGroup) {
			continue
		}

		stats.OverallStats.TotalSets++
		setsByDate[session.Date]++

		if rawByDateExercise[session.Date] == nil {
			rawByDateExercise[session.Date] = make(map[int]float64)
		}
		rawByDateExercise[session.Date][record.ExerciseID] += recordRaw(exercise, record)
		if bestSetByDateExercise[session.Date] == nil {
			bestSetByDateExercise[session.Date] = make(map[int]float64)
		}
		if bestSet := recordBestSetValue(exercise, record); bestSet > bestSetByDateExercise[session.Date][record.ExerciseID] {
			bestSetByDateExercise[session.Date][record.ExerciseID] = bestSet
		}

		if record.Weight > 0 && record.Weight > stats.BestPerformance.MaxWeight.Value {
			stats.BestPerformance.MaxWeight = buildSetPerformanceWithUnit(session.Date, exercise, record, record.Weight, "kg")
		}
		if record.Reps > 0 && float64(record.Reps) > stats.BestPerformance.MaxReps.Value {
			stats.BestPerformance.MaxReps = buildSetPerformanceWithUnit(session.Date, exercise, record, float64(record.Reps), "reps")
		}
		if record.Duration > 0 && float64(record.Duration) > stats.BestPerformance.MaxDuration.Value {
			stats.BestPerformance.MaxDuration = buildSetPerformanceWithUnit(session.Date, exercise, record, float64(record.Duration), "duration")
		}
	}

	setDates := make([]string, 0, len(setsByDate))
	for date := range setsByDate {
		setDates = append(setDates, date)
	}
	sort.Strings(setDates)
	for _, date := range setDates {
		sets := setsByDate[date]
		stats.OverallStats.TrainingDays++
		if sets > stats.BestPerformance.MaxSets.Sets {
			stats.BestPerformance.MaxSets.Date = date
			stats.BestPerformance.MaxSets.Sets = sets
		}
	}

	baselines := s.calculateExerciseBaselines(exerciseMap, sessionMap, records, exerciseID, muscleGroup, growthPercentile)
	rawUnit := inferTrendRawUnit(exerciseMap, sessionMap, records, exerciseID, days)
	if exerciseID > 0 {
		if exercise, ok := exerciseMap[exerciseID]; ok {
			rawUnit = exercise.Unit
		}
	}
	for i := days - 1; i >= 0; i-- {
		date := endDate.AddDate(0, 0, -i).Format("2006-01-02")
		score := 0.0
		rawValue := 0.0
		totalRawValue := 0.0
		for currentExerciseID, raw := range rawByDateExercise[date] {
			if baseline := baselines[currentExerciseID]; baseline > 0 {
				score += raw / baseline
			}
			if exerciseID > 0 && currentExerciseID == exerciseID {
				rawValue = bestSetByDateExercise[date][currentExerciseID]
				totalRawValue = raw
			}
		}
		stats.DailyScores = append(stats.DailyScores, DailyMuscleScore{
			Date:          date,
			MuscleScore:   score,
			RawValue:      rawValue,
			TotalRawValue: totalRawValue,
			RawUnit:       rawUnit,
		})
	}

	return stats, nil
}

// GetDayRecords 获取某一天在当前筛选条件下的所有组
func (s *StatsAnalyzer) GetDayRecords(date string, exerciseID int, muscleGroup string) (*DayRecordsStats, error) {
	records, _ := s.csv.LoadTrainingRecords()
	sessions, _ := s.csv.LoadTrainingSessions()
	exercises, _ := s.csv.LoadExercises()
	groups, _ := s.csv.LoadExerciseGroups()
	exerciseMap := buildExerciseMap(exercises)
	sessionMap := buildSessionMap(sessions)
	groupMap := make(map[int]string)
	for _, group := range groups {
		groupMap[group.ID] = group.Name
	}

	stats := &DayRecordsStats{Date: date}
	for _, record := range records {
		session, ok := sessionMap[record.SessionID]
		if !ok || session.Date != date {
			continue
		}
		exercise, ok := exerciseMap[record.ExerciseID]
		if !ok || !matchesStatsFilter(exercise, exerciseID, muscleGroup) {
			continue
		}

		performance := buildSetPerformance(date, exercise, record, recordRaw(exercise, record))
		performance.GroupID = session.GroupID
		performance.GroupName = groupMap[session.GroupID]
		stats.Records = append(stats.Records, performance)
	}

	sort.Slice(stats.Records, func(i, j int) bool {
		if stats.Records[i].ExerciseName == stats.Records[j].ExerciseName {
			return stats.Records[i].SetNumber < stats.Records[j].SetNumber
		}
		return stats.Records[i].ExerciseName < stats.Records[j].ExerciseName
	})

	return stats, nil
}

// DeleteDayRecords 删除某一天的全部训练会话和组记录
func (s *StatsAnalyzer) DeleteDayRecords(date string) (int, int, error) {
	records, _ := s.csv.LoadTrainingRecords()
	sessions, _ := s.csv.LoadTrainingSessions()

	deletedSessionIDs := make(map[int]bool)
	remainingSessions := make([]TrainingSession, 0, len(sessions))
	for _, session := range sessions {
		if session.Date == date {
			deletedSessionIDs[session.SessionID] = true
			continue
		}
		remainingSessions = append(remainingSessions, session)
	}

	deletedRecords := 0
	remainingRecords := make([]TrainingRecord, 0, len(records))
	for _, record := range records {
		if deletedSessionIDs[record.SessionID] {
			deletedRecords++
			continue
		}
		remainingRecords = append(remainingRecords, record)
	}

	if err := s.csv.SaveTrainingData(remainingSessions, remainingRecords); err != nil {
		return 0, 0, err
	}

	return len(deletedSessionIDs), deletedRecords, nil
}

func (s *StatsAnalyzer) calculateExerciseBaselines(exerciseMap map[int]Exercise, sessionMap map[int]TrainingSession, records []TrainingRecord, exerciseID int, muscleGroup string, growthPercentile float64) map[int]float64 {
	rawByExerciseDate := make(map[int]map[string]float64)

	for _, record := range records {
		session, ok := sessionMap[record.SessionID]
		if !ok || !isSessionWithinDays(session.Date, 30) {
			continue
		}
		exercise, ok := exerciseMap[record.ExerciseID]
		if !ok || !matchesStatsFilter(exercise, exerciseID, muscleGroup) {
			continue
		}

		if rawByExerciseDate[record.ExerciseID] == nil {
			rawByExerciseDate[record.ExerciseID] = make(map[string]float64)
		}
		rawByExerciseDate[record.ExerciseID][session.Date] += recordRaw(exercise, record)
	}

	baselines := make(map[int]float64)
	for currentExerciseID, rawByDate := range rawByExerciseDate {
		values := make([]float64, 0, len(rawByDate))
		for _, value := range rawByDate {
			if value > 0 {
				values = append(values, value)
			}
		}
		if len(values) == 0 {
			continue
		}
		if len(values) < 3 {
			baselines[currentExerciseID] = median(values)
		} else {
			baselines[currentExerciseID] = percentile(values, growthPercentile)
		}
	}

	return baselines
}

func buildExerciseMap(exercises []Exercise) map[int]Exercise {
	exerciseMap := make(map[int]Exercise)
	for _, exercise := range exercises {
		exerciseMap[exercise.ID] = exercise
	}
	return exerciseMap
}

func buildSessionMap(sessions []TrainingSession) map[int]TrainingSession {
	sessionMap := make(map[int]TrainingSession)
	for _, session := range sessions {
		sessionMap[session.SessionID] = session
	}
	return sessionMap
}

func matchesStatsFilter(exercise Exercise, exerciseID int, muscleGroup string) bool {
	if exerciseID > 0 && exercise.ID != exerciseID {
		return false
	}
	return muscleGroup == "" || exercise.MuscleGroup == muscleGroup
}

func recordRaw(exercise Exercise, record TrainingRecord) float64 {
	switch exercise.Unit {
	case "duration":
		return float64(record.Duration)
	case "reps":
		return float64(record.Reps)
	default:
		if record.Weight <= 0 && record.Reps > 0 {
			return float64(record.Reps)
		}
		return record.Weight * float64(record.Reps)
	}
}

func recordBestSetValue(exercise Exercise, record TrainingRecord) float64 {
	switch effectiveRecordUnit(exercise, record) {
	case "duration":
		return float64(record.Duration)
	case "reps":
		return float64(record.Reps)
	default:
		return record.Weight
	}
}

func buildSetPerformance(date string, exercise Exercise, record TrainingRecord, value float64) SetPerformance {
	return SetPerformance{
		Date:         date,
		SessionID:    record.SessionID,
		ExerciseID:   record.ExerciseID,
		ExerciseName: exercise.Name,
		MuscleGroup:  exercise.MuscleGroup,
		Unit:         effectiveRecordUnit(exercise, record),
		SetNumber:    record.SetNumber,
		Weight:       record.Weight,
		Reps:         record.Reps,
		Duration:     record.Duration,
		Value:        value,
	}
}

func buildSetPerformanceWithUnit(date string, exercise Exercise, record TrainingRecord, value float64, unit string) SetPerformance {
	performance := buildSetPerformance(date, exercise, record, value)
	performance.Unit = unit
	return performance
}

func effectiveRecordUnit(exercise Exercise, record TrainingRecord) string {
	if exercise.Unit == "kg" && record.Weight <= 0 && record.Reps > 0 {
		return "reps"
	}
	return exercise.Unit
}

func inferTrendRawUnit(exerciseMap map[int]Exercise, sessionMap map[int]TrainingSession, records []TrainingRecord, exerciseID int, days int) string {
	if exerciseID <= 0 {
		return "score"
	}

	exercise, ok := exerciseMap[exerciseID]
	if !ok {
		return "score"
	}
	if exercise.Unit != "kg" {
		return exercise.Unit
	}

	for _, record := range records {
		session, ok := sessionMap[record.SessionID]
		if !ok || !isSessionWithinDays(session.Date, days) || record.ExerciseID != exerciseID {
			continue
		}
		if record.Weight > 0 {
			return "kg"
		}
	}

	return "reps"
}

func calculateHeatmapLevel(sets int) int {
	switch {
	case sets <= 0:
		return 0
	case sets <= 2:
		return 1
	case sets <= 5:
		return 2
	case sets <= 8:
		return 3
	default:
		return 4
	}
}

func isSessionWithinDays(date string, days int) bool {
	if days <= 0 {
		return true
	}

	sessionDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false
	}

	now := time.Now()
	startDate := now.AddDate(0, 0, -days+1)
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, now.Location())
	return !sessionDate.Before(startDate)
}

func statsEndDate(sessions []TrainingSession) time.Time {
	now := time.Now()
	endDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	for _, session := range sessions {
		sessionDate, err := time.ParseInLocation("2006-01-02", session.Date, now.Location())
		if err != nil {
			continue
		}
		if sessionDate.After(endDate) {
			endDate = sessionDate
		}
	}

	return endDate
}

func isSessionWithinDaysFromEnd(date string, days int, endDate time.Time) bool {
	if days <= 0 {
		return true
	}

	sessionDate, err := time.ParseInLocation("2006-01-02", date, endDate.Location())
	if err != nil {
		return false
	}

	startDate := endDate.AddDate(0, 0, -days+1)
	return !sessionDate.Before(startDate) && !sessionDate.After(endDate)
}

func formatISOWeek(year int, week int) string {
	return fmt.Sprintf("%04d-W%02d", year, week)
}
