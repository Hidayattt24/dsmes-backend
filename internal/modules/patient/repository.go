package patient

import (
	"context"
	"errors"
	"math"
	"sort"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type patientRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewPatientRepository(db *gorm.DB, log *zap.Logger) PatientRepository {
	return &patientRepository{db: db, log: log}
}

func (r *patientRepository) FindAll(ctx context.Context, filter PatientFilterQuery) ([]domain.Patient, int64, error) {
	var items []domain.Patient
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.Patient{}).Where("patients.deleted_at IS NULL")

	if filter.StaffID != "" {
		q = q.Where("patients.assigned_staff_id = ?", filter.StaffID)
	}

	if filter.HealthFacility != "" {
		q = q.Where("patients.health_facility = ?", filter.HealthFacility)
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		q = q.Where("patients.full_name ILIKE ? OR patients.email ILIKE ?", searchPattern, searchPattern)
	}

	if filter.Gender != "" && filter.Gender != "Semua" {
		genderVal := domain.Gender(filter.Gender)
		if filter.Gender == "Laki-laki" || filter.Gender == "laki_laki" {
			genderVal = domain.GenderLakiLaki
		} else if filter.Gender == "Perempuan" || filter.Gender == "perempuan" {
			genderVal = domain.GenderPerempuan
		}
		q = q.Where("patients.gender = ?", genderVal)
	}

	if filter.Status != "" && filter.Status != "Semua" {
		statusVal := domain.AccountStatus(filter.Status)
		if filter.Status == "Aktif" || filter.Status == "aktif" {
			statusVal = domain.StatusAktif
		} else if filter.Status == "Nonaktif" || filter.Status == "nonaktif" {
			statusVal = domain.StatusNonaktif
		}
		q = q.Where("patients.status = ?", statusVal)
	}

	if filter.AgeMin != nil {
		q = q.Where("EXTRACT(YEAR FROM AGE(NOW(), patients.date_of_birth)) >= ?", *filter.AgeMin)
	}
	if filter.AgeMax != nil {
		q = q.Where("EXTRACT(YEAR FROM AGE(NOW(), patients.date_of_birth)) <= ?", *filter.AgeMax)
	}

	if filter.ComplianceMin != nil {
		q = q.Where("patients.compliance >= ?", *filter.ComplianceMin)
	}
	if filter.ComplianceMax != nil {
		q = q.Where("patients.compliance <= ?", *filter.ComplianceMax)
	}

	// Track whether we already joined blood_sugar_logs to avoid duplicate JOINs
	hasBSJoin := false

	if filter.BloodSugarStatus != "" && filter.BloodSugarStatus != "Semua" {
		bsStatus := filter.BloodSugarStatus
		if bsStatus == "Tinggi" {
			bsStatus = "tinggi"
		} else if bsStatus == "Sangat Tinggi" || bsStatus == "sangat_tinggi" {
			bsStatus = "sangat_tinggi"
		} else if bsStatus == "Rendah" || bsStatus == "rendah" {
			bsStatus = "rendah"
		} else if bsStatus == "Normal" || bsStatus == "normal" {
			bsStatus = "normal"
		}
		q = q.Joins("LEFT JOIN (SELECT DISTINCT ON (patient_id) patient_id, status FROM blood_sugar_logs WHERE deleted_at IS NULL ORDER BY patient_id, measured_at DESC) latest_bs ON latest_bs.patient_id = patients.id")
		hasBSJoin = true
		q = q.Where("latest_bs.status = ?", bsStatus)
	}

	if filter.RiskLevel != "" && filter.RiskLevel != "Semua" {
		if !hasBSJoin {
			q = q.Joins("LEFT JOIN (SELECT DISTINCT ON (patient_id) patient_id, status FROM blood_sugar_logs WHERE deleted_at IS NULL ORDER BY patient_id, measured_at DESC) latest_bs ON latest_bs.patient_id = patients.id")
			hasBSJoin = true
		}
		riskSQL := `
			CASE
				WHEN latest_bs.status = 'sangat_tinggi' OR patients.compliance < 30 THEN 'sangat_tinggi'
				WHEN latest_bs.status = 'tinggi' OR (patients.compliance >= 30 AND patients.compliance < 50) THEN 'tinggi'
				WHEN latest_bs.status = 'rendah' OR (patients.compliance >= 50 AND patients.compliance < 70) THEN 'sedang'
				ELSE 'rendah'
			END
		`
		rl := filter.RiskLevel
		if rl == "Sangat Tinggi" {
			rl = "sangat_tinggi"
		} else if rl == "Tinggi" {
			rl = "tinggi"
		} else if rl == "Sedang" {
			rl = "sedang"
		} else if rl == "Rendah" {
			rl = "rendah"
		}
		q = q.Where(riskSQL+" = ?", rl)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errs.NewInternal("failed to count patients", err)
	}

	order := "DESC"
	if filter.SortOrder == "asc" {
		order = "ASC"
	}

	switch filter.SortBy {
	case "name":
		q = q.Order("patients.full_name " + order)
	case "newest":
		q = q.Order("patients.created_at DESC")
	case "oldest":
		q = q.Order("patients.created_at ASC")
	case "latest_record":
		q = q.Order("patients.last_active_at DESC NULLS LAST")
	case "highest_blood_sugar":
		if !hasBSJoin {
			q = q.Joins("LEFT JOIN (SELECT DISTINCT ON (patient_id) patient_id, glucose_value FROM blood_sugar_logs WHERE deleted_at IS NULL ORDER BY patient_id, measured_at DESC) latest_bs ON latest_bs.patient_id = patients.id")
		}
		q = q.Order("latest_bs.glucose_value " + order + " NULLS LAST")
	default:
		q = q.Order("patients.created_at DESC")
	}

	offset := (filter.Page - 1) * filter.Limit
	err := q.Preload("AssignedStaff").
		Offset(offset).Limit(filter.Limit).
		Find(&items).Error
	if err != nil {
		return nil, 0, errs.NewInternal("failed to fetch patients", err)
	}

	return items, total, nil
}

func (r *patientRepository) FindByID(ctx context.Context, id string) (*domain.Patient, error) {
	var p domain.Patient
	err := r.db.WithContext(ctx).
		Preload("AssignedStaff").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("patient not found")
		}
		return nil, errs.NewInternal("failed to fetch patient", err)
	}
	return &p, nil
}

func (r *patientRepository) FindByEmail(ctx context.Context, email string) (*domain.Patient, error) {
	var p domain.Patient
	err := r.db.WithContext(ctx).
		Preload("AssignedStaff").
		Where("email = ? AND deleted_at IS NULL", email).
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("patient not found")
		}
		return nil, errs.NewInternal("failed to fetch patient by email", err)
	}
	return &p, nil
}

func (r *patientRepository) FindByPhoneNumber(ctx context.Context, phone string) (*domain.Patient, error) {
	var p domain.Patient
	err := r.db.WithContext(ctx).
		Preload("AssignedStaff").
		Where("phone_number = ? AND deleted_at IS NULL", phone).
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("patient not found")
		}
		return nil, errs.NewInternal("failed to fetch patient by phone number", err)
	}
	return &p, nil
}

func (r *patientRepository) Create(ctx context.Context, p *domain.Patient) error {
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		return errs.NewInternal("failed to create patient", err)
	}
	return nil
}

func (r *patientRepository) Update(ctx context.Context, p *domain.Patient) error {
	result := r.db.WithContext(ctx).Save(p)
	if result.Error != nil {
		return errs.NewInternal("failed to update patient", result.Error)
	}
	return nil
}

func (r *patientRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Model(&domain.Patient{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()"))
	if result.Error != nil {
		return errs.NewInternal("failed to soft delete patient", result.Error)
	}
	if result.RowsAffected == 0 {
		return errs.NewNotFound("patient not found")
	}
	return nil
}

func (r *patientRepository) CreateWithOnboarding(ctx context.Context, p *domain.Patient, defaultRoutines []domain.Routine, defaultReminders []domain.Reminder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(p).Error; err != nil {
			return err
		}

		for i := range defaultRoutines {
			defaultRoutines[i].PatientID = p.ID
			if err := tx.Create(&defaultRoutines[i]).Error; err != nil {
				return err
			}
		}

		for i := range defaultReminders {
			defaultReminders[i].PatientID = p.ID
			// Set the default timestamps & UUID for child records
			if err := tx.Create(&defaultReminders[i]).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *patientRepository) GetStats(ctx context.Context, facilityName string) (*PatientStats, error) {
	var total int64
	var active int64

	q := r.db.WithContext(ctx).Model(&domain.Patient{}).Where("deleted_at IS NULL")
	if facilityName != "" {
		q = q.Where("health_facility = ?", facilityName)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, errs.NewInternal("failed to count total patients for stats", err)
	}

	qActive := r.db.WithContext(ctx).Model(&domain.Patient{}).Where("deleted_at IS NULL AND status = ?", domain.StatusAktif)
	if facilityName != "" {
		qActive = qActive.Where("health_facility = ?", facilityName)
	}
	if err := qActive.Count(&active).Error; err != nil {
		return nil, errs.NewInternal("failed to count active patients for stats", err)
	}

	var agesResult struct {
		Youngest float64
		Oldest   float64
	}
	qAge := r.db.WithContext(ctx).Table("patients").Where("deleted_at IS NULL AND date_of_birth IS NOT NULL")
	if facilityName != "" {
		qAge = qAge.Where("health_facility = ?", facilityName)
	}
	err := qAge.Select("COALESCE(MIN(EXTRACT(YEAR FROM AGE(NOW(), date_of_birth))), 0) as youngest, COALESCE(MAX(EXTRACT(YEAR FROM AGE(NOW(), date_of_birth))), 0) as oldest").Scan(&agesResult).Error
	if err != nil {
		return nil, errs.NewInternal("failed to calculate youngest and oldest age", err)
	}

	return &PatientStats{
		TotalPatients:  total,
		ActivePatients: active,
		YoungestAge:    int(agesResult.Youngest),
		OldestAge:      int(agesResult.Oldest),
	}, nil
}

func (r *patientRepository) GetPatientSummary(ctx context.Context, patientID string) (*PatientSummaryData, error) {
	var summary PatientSummaryData

	// 1. Get average blood sugar and latest blood sugar
	var latestBS domain.BloodSugarLog
	err := r.db.WithContext(ctx).
		Where("patient_id = ? AND deleted_at IS NULL", patientID).
		Order("measured_at DESC").
		First(&latestBS).Error
	if err == nil && latestBS.GlucoseValue > 0 {
		val := latestBS.GlucoseValue
		summary.LatestBloodSugar = &val
		summary.LatestBloodSugarTime = &latestBS.MeasuredAt
		statusStr := string(latestBS.Category)
		summary.LatestBloodSugarStatus = &statusStr
	} else {
		// Fallback to patient_measurements table
		var latestM domain.PatientMeasurement
		mErr := r.db.WithContext(ctx).
			Where("patient_id = ? AND blood_sugar IS NOT NULL AND blood_sugar > 0 AND deleted_at IS NULL", patientID).
			Order("measured_at DESC, created_at DESC").
			First(&latestM).Error
		if mErr == nil && latestM.BloodSugar != nil && *latestM.BloodSugar > 0 {
			val := *latestM.BloodSugar
			summary.LatestBloodSugar = &val
			summary.LatestBloodSugarTime = &latestM.MeasuredAt
			st := string(domain.CalculateGlucoseStatus(val, domain.TimeSewaktu))
			summary.LatestBloodSugarStatus = &st
		}
	}

	var avgBS float64
	err = r.db.WithContext(ctx).Model(&domain.BloodSugarLog{}).
		Where("patient_id = ? AND deleted_at IS NULL", patientID).
		Select("COALESCE(AVG(glucose_value), 0)").
		Scan(&avgBS).Error
	if err == nil && avgBS > 0 {
		summary.AverageBloodSugar = &avgBS
	} else {
		var avgM float64
		_ = r.db.WithContext(ctx).Model(&domain.PatientMeasurement{}).
			Where("patient_id = ? AND blood_sugar IS NOT NULL AND blood_sugar > 0 AND deleted_at IS NULL", patientID).
			Select("COALESCE(AVG(blood_sugar), 0)").
			Scan(&avgM).Error
		if avgM > 0 {
			summary.AverageBloodSugar = &avgM
		}
	}

	// 1b. Get latest meal calories and type
	var latestMeal domain.MealLog
	err = r.db.WithContext(ctx).
		Preload("Food").
		Where("patient_id = ? AND deleted_at IS NULL", patientID).
		Order("logged_at DESC").
		First(&latestMeal).Error
	if err == nil {
		var cals float64
		if latestMeal.Food != nil {
			cals = latestMeal.Food.Calories * latestMeal.PortionMultiplier
			foodName := latestMeal.Food.Name
			summary.LatestMealName = &foodName
		}
		summary.LatestMealCalories = &cals
		mType := string(latestMeal.MealType)
		summary.LatestMealType = &mType
	}

	// 1c. Get today's total consumed calories
	var todayCals float64
	r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(COALESCE(f.calories, 0) * COALESCE(ml.portion_multiplier, 1)), 0)
		FROM meal_logs ml
		LEFT JOIN foods f ON f.id = ml.food_id
		WHERE ml.patient_id = ? AND ml.logged_at >= CURRENT_DATE AND ml.deleted_at IS NULL
	`, patientID).Scan(&todayCals)
	summary.TodayConsumedCalories = &todayCals

	// 2. Get latest weight and height from patient to calculate BMI
	var patient domain.Patient
	err = r.db.WithContext(ctx).
		Select("weight_kg, height_cm").
		Where("id = ? AND deleted_at IS NULL", patientID).
		First(&patient).Error
	if err == nil {
		w := patient.WeightKg
		summary.LatestWeight = &w
		if patient.HeightCm > 0 {
			heightM := patient.HeightCm / 100.0
			bmi := w / (heightM * heightM)
			summary.BMI = &bmi
		}
	}

	// 3. Get latest completed activity
	type ActivityResult struct {
		DescriptiveName string
		LoggedAt        time.Time
	}
	var actResult ActivityResult
	err = r.db.WithContext(ctx).Raw(`
		SELECT descriptive_name, logged_at
		FROM (
			SELECT pal.activity_name AS descriptive_name, pal.logged_at
			FROM patient_activity_logs pal
			WHERE pal.patient_id = ? AND pal.deleted_at IS NULL
			UNION ALL
			SELECT COALESCE(r.descriptive_name, CAST(r.routine_type AS VARCHAR)) AS descriptive_name, le.logged_at
			FROM routine_log_entries le
			JOIN routine_times rt ON rt.id = le.routine_time_id
			JOIN routines r ON r.id = rt.routine_id
			WHERE le.patient_id = ? AND le.status = 'Completed' AND le.deleted_at IS NULL AND rt.deleted_at IS NULL AND r.deleted_at IS NULL
		) combined_act
		ORDER BY logged_at DESC
		LIMIT 1
	`, patientID, patientID).Scan(&actResult).Error
	if err == nil && actResult.DescriptiveName != "" {
		name := actResult.DescriptiveName
		summary.LatestActivityName = &name
		t := actResult.LoggedAt
		summary.LatestActivityTime = &t
	}

	return &summary, nil
}

func (r *patientRepository) GetPatientSummaries(ctx context.Context, patientIDs []string) (map[string]*PatientSummaryData, error) {
	if len(patientIDs) == 0 {
		return map[string]*PatientSummaryData{}, nil
	}

	result := make(map[string]*PatientSummaryData, len(patientIDs))
	for _, id := range patientIDs {
		result[id] = nil
	}

	// Batch fetch latest blood sugar for all patients
	// Prioritizes blood_sugar_logs; falls back to patient_measurements.blood_sugar if none
	type BSResult struct {
		PatientID    string
		GlucoseValue int
		MeasuredAt   time.Time
		Status       string
	}
	var bsResults []BSResult
	r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (patient_id) patient_id, glucose_value, measured_at, status
		FROM (
			SELECT bs.patient_id, bs.glucose_value, bs.measured_at, bs.status::text
			FROM blood_sugar_logs bs
			WHERE bs.patient_id IN ? AND bs.deleted_at IS NULL
			UNION ALL
			SELECT pm.patient_id, pm.blood_sugar AS glucose_value, pm.measured_at,
				CASE
					WHEN pm.blood_sugar < 70 THEN 'rendah'
					WHEN pm.blood_sugar >= 200 THEN 'sangat_tinggi'
					WHEN pm.blood_sugar >= 140 THEN 'tinggi'
					ELSE 'normal'
				END AS status
			FROM patient_measurements pm
			WHERE pm.patient_id IN ? AND pm.blood_sugar IS NOT NULL AND pm.blood_sugar > 0 AND pm.deleted_at IS NULL
		) combined
		ORDER BY patient_id, measured_at DESC
	`, patientIDs, patientIDs).Scan(&bsResults)

	bsMap := make(map[string]BSResult, len(bsResults))
	for _, bs := range bsResults {
		bsMap[bs.PatientID] = bs
	}

	// Batch fetch average blood sugar (includes patient_measurements as fallback)
	type AvgBSResult struct {
		PatientID string
		AvgValue  float64
	}
	var avgBSResults []AvgBSResult
	r.db.WithContext(ctx).Raw(`
		SELECT patient_id, COALESCE(AVG(glucose_value), 0) as avg_value
		FROM (
			SELECT bs.patient_id, bs.glucose_value
			FROM blood_sugar_logs bs
			WHERE bs.patient_id IN ? AND bs.deleted_at IS NULL
			UNION ALL
			SELECT pm.patient_id, pm.blood_sugar AS glucose_value
			FROM patient_measurements pm
			WHERE pm.patient_id IN ? AND pm.blood_sugar IS NOT NULL AND pm.blood_sugar > 0 AND pm.deleted_at IS NULL
		) combined
		GROUP BY patient_id
	`, patientIDs, patientIDs).Scan(&avgBSResults)

	avgBSMap := make(map[string]float64, len(avgBSResults))
	for _, a := range avgBSResults {
		avgBSMap[a.PatientID] = a.AvgValue
	}

	// Batch fetch latest meal
	type MealResult struct {
		PatientID string
		Calories  float64
		MealType  string
		FoodName  string
	}
	var mealResults []MealResult
	r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (ml.patient_id) ml.patient_id,
			COALESCE(f.calories * ml.portion_multiplier, 0) as calories,
			ml.meal_type,
			COALESCE(f.name, '') as food_name
		FROM meal_logs ml
		LEFT JOIN foods f ON f.id = ml.food_id AND f.deleted_at IS NULL
		WHERE ml.patient_id IN ? AND ml.deleted_at IS NULL
		ORDER BY ml.patient_id, ml.logged_at DESC
	`, patientIDs).Scan(&mealResults)

	mealMap := make(map[string]MealResult, len(mealResults))
	for _, m := range mealResults {
		mealMap[m.PatientID] = m
	}

	// Batch fetch latest activity
	type ActResult struct {
		PatientID       string
		DescriptiveName string
		LoggedAt        time.Time
	}
	var actResults []ActResult
	r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (patient_id) patient_id, descriptive_name, logged_at
		FROM (
			SELECT pal.patient_id, pal.activity_name AS descriptive_name, pal.logged_at
			FROM patient_activity_logs pal
			WHERE pal.patient_id IN ? AND pal.deleted_at IS NULL
			UNION ALL
			SELECT le.patient_id, COALESCE(r.descriptive_name, CAST(r.routine_type AS VARCHAR)) AS descriptive_name, le.logged_at
			FROM routine_log_entries le
			JOIN routine_times rt ON rt.id = le.routine_time_id AND rt.deleted_at IS NULL
			JOIN routines r ON r.id = rt.routine_id AND r.deleted_at IS NULL
			WHERE le.patient_id IN ? AND le.status = 'Completed' AND le.deleted_at IS NULL
		) combined_act
		ORDER BY patient_id, logged_at DESC
	`, patientIDs, patientIDs).Scan(&actResults)

	actMap := make(map[string]ActResult, len(actResults))
	for _, a := range actResults {
		actMap[a.PatientID] = a
	}

	// Batch fetch today's total consumed calories
	type TodayCalResult struct {
		PatientID string
		TotalCal  float64
	}
	var todayCalResults []TodayCalResult
	r.db.WithContext(ctx).Raw(`
		SELECT ml.patient_id, COALESCE(SUM(COALESCE(f.calories, 0) * COALESCE(ml.portion_multiplier, 1)), 0) as total_cal
		FROM meal_logs ml
		LEFT JOIN foods f ON f.id = ml.food_id
		WHERE ml.patient_id IN ? AND ml.logged_at >= CURRENT_DATE AND ml.deleted_at IS NULL
		GROUP BY ml.patient_id
	`, patientIDs).Scan(&todayCalResults)

	todayCalMap := make(map[string]float64, len(todayCalResults))
	for _, tc := range todayCalResults {
		todayCalMap[tc.PatientID] = tc.TotalCal
	}

	for _, pid := range patientIDs {
		summary := &PatientSummaryData{}

		if bs, ok := bsMap[pid]; ok {
			val := bs.GlucoseValue
			summary.LatestBloodSugar = &val
			summary.LatestBloodSugarTime = &bs.MeasuredAt
			statusStr := bs.Status
			summary.LatestBloodSugarStatus = &statusStr
		}

		if avg, ok := avgBSMap[pid]; ok && avg > 0 {
			summary.AverageBloodSugar = &avg
		}

		if m, ok := mealMap[pid]; ok {
			summary.LatestMealCalories = &m.Calories
			mType := m.MealType
			summary.LatestMealType = &mType
			if m.FoodName != "" {
				fName := m.FoodName
				summary.LatestMealName = &fName
			}
		}

		if a, ok := actMap[pid]; ok && a.DescriptiveName != "" {
			name := a.DescriptiveName
			summary.LatestActivityName = &name
			summary.LatestActivityTime = &a.LoggedAt
		}

		if tc, ok := todayCalMap[pid]; ok {
			summary.TodayConsumedCalories = &tc
		} else {
			zero := 0.0
			summary.TodayConsumedCalories = &zero
		}

		// Get weight and height for BMI (reuse patient data already fetched)
		// BMI will be empty in batch — callers already have patient weight/height
		result[pid] = summary
	}

	return result, nil
}

func (r *patientRepository) GetPatientActivityAnalytics(ctx context.Context, patientID string, days int) (*PatientActivityAnalyticsResponse, error) {
	var bsCount, mealCount, actCount, medCount int64

	bsQuery := r.db.WithContext(ctx).Table("blood_sugar_logs").Where("patient_id = ? AND deleted_at IS NULL", patientID)
	mealQuery := r.db.WithContext(ctx).Table("meal_logs").Where("patient_id = ? AND deleted_at IS NULL", patientID)
	actQuery := r.db.WithContext(ctx).Table("routine_log_entries").Where("patient_id = ? AND status = 'Completed' AND deleted_at IS NULL", patientID)
	medQuery := r.db.WithContext(ctx).Table("daily_reminder_logs d").
		Joins("JOIN reminders r ON r.id = d.reminder_id AND r.deleted_at IS NULL").
		Where("r.patient_id = ? AND d.status = 'selesai' AND d.deleted_at IS NULL", patientID)

	if days > 0 {
		cutoff := time.Now().AddDate(0, 0, -days)
		bsQuery = bsQuery.Where("measured_at >= ?", cutoff)
		mealQuery = mealQuery.Where("logged_at >= ?", cutoff)
		actQuery = actQuery.Where("logged_at >= ?", cutoff)
		medQuery = medQuery.Where("d.log_date >= ?", cutoff)
	}

	if err := bsQuery.Count(&bsCount).Error; err != nil {
		return nil, errs.NewInternal("failed to count blood sugar logs", err)
	}

	if err := mealQuery.Count(&mealCount).Error; err != nil {
		return nil, errs.NewInternal("failed to count meal logs", err)
	}

	if err := actQuery.Count(&actCount).Error; err != nil {
		return nil, errs.NewInternal("failed to count activity logs", err)
	}

	if err := medQuery.Count(&medCount).Error; err != nil {
		return nil, errs.NewInternal("failed to count medication logs", err)
	}

	total := bsCount + mealCount + actCount + medCount

	calcPct := func(cnt int64) float64 {
		if total == 0 {
			return 0.0
		}
		return math.Round((float64(cnt)/float64(total))*1000) / 10
	}

	categories := []struct {
		Name  string
		Count int64
	}{
		{"Gula Darah", bsCount},
		{"Asupan Makanan", mealCount},
		{"Aktivitas Fisik", actCount},
		{"Obat", medCount},
	}

	mostUsed := "-"
	leastUsed := "-"

	if total > 0 {
		maxCnt := int64(-1)
		minCnt := int64(1<<63 - 1)

		for _, c := range categories {
			if c.Count > maxCnt {
				maxCnt = c.Count
				mostUsed = c.Name
			}
			if c.Count < minCnt {
				minCnt = c.Count
				leastUsed = c.Name
			}
		}
	}

	return &PatientActivityAnalyticsResponse{
		TotalRecords: total,
		BloodSugar: ActivityAnalyticsItem{
			Count:      bsCount,
			Percentage: calcPct(bsCount),
		},
		Food: ActivityAnalyticsItem{
			Count:      mealCount,
			Percentage: calcPct(mealCount),
		},
		PhysicalActivity: ActivityAnalyticsItem{
			Count:      actCount,
			Percentage: calcPct(actCount),
		},
		Medication: ActivityAnalyticsItem{
			Count:      medCount,
			Percentage: calcPct(medCount),
		},
		MostUsed:  mostUsed,
		LeastUsed: leastUsed,
	}, nil
}

func (r *patientRepository) GetPatientDailyLogsAggregate(ctx context.Context, patientID string, startDate, endDate time.Time) (map[string]*DailyLogsAggregate, error) {
	all, err := r.GetPatientDailyLogsAggregates(ctx, []string{patientID}, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return all[patientID], nil
}

// GetPatientDailyLogsAggregates computes daily logs aggregates for many patients
// in a handful of GROUP BY queries (IN (...) + GROUP BY) instead of one query
// per patient, eliminating the N+1 pattern in ListPatients.
func (r *patientRepository) GetPatientDailyLogsAggregates(ctx context.Context, patientIDs []string, startDate, endDate time.Time) (map[string]map[string]*DailyLogsAggregate, error) {
	result := make(map[string]map[string]*DailyLogsAggregate)
	if len(patientIDs) == 0 {
		return result, nil
	}

	getDay := func(patientID, dateStr string) *DailyLogsAggregate {
		if result[patientID] == nil {
			result[patientID] = make(map[string]*DailyLogsAggregate)
		}
		if result[patientID][dateStr] == nil {
			result[patientID][dateStr] = &DailyLogsAggregate{}
		}
		return result[patientID][dateStr]
	}

	// 1. Blood sugar counts by day
	type BSGroup struct {
		PatientID string
		DateStr   string
		Count     int
	}
	var bsGroups []BSGroup
	if err := r.db.WithContext(ctx).Raw(`
		SELECT patient_id, TO_CHAR(measured_at, 'YYYY-MM-DD') as date_str, COUNT(*) as count
		FROM blood_sugar_logs
		WHERE patient_id IN (?) AND measured_at >= ? AND measured_at <= ? AND deleted_at IS NULL
		GROUP BY patient_id, date_str
	`, patientIDs, startDate, endDate).Scan(&bsGroups).Error; err != nil {
		return nil, errs.NewInternal("failed to aggregate blood sugar logs", err)
	}
	for _, g := range bsGroups {
		getDay(g.PatientID, g.DateStr).BloodSugarCount = g.Count
	}

	// 2. Meal calories and counts by day
	type MealGroup struct {
		PatientID string
		DateStr   string
		TotalCal  float64
		Count     int
	}
	var mealGroups []MealGroup
	if err := r.db.WithContext(ctx).Raw(`
		SELECT ml.patient_id, TO_CHAR(ml.logged_at, 'YYYY-MM-DD') as date_str,
			SUM(COALESCE(f.calories, 0) * COALESCE(ml.portion_multiplier, 1)) as total_cal,
			COUNT(*) as count
		FROM meal_logs ml
		LEFT JOIN foods f ON f.id = ml.food_id
		WHERE ml.patient_id IN (?) AND ml.logged_at >= ? AND ml.logged_at <= ? AND ml.deleted_at IS NULL
		GROUP BY ml.patient_id, date_str
	`, patientIDs, startDate, endDate).Scan(&mealGroups).Error; err != nil {
		return nil, errs.NewInternal("failed to aggregate meal logs", err)
	}
	for _, g := range mealGroups {
		day := getDay(g.PatientID, g.DateStr)
		day.TotalMealCalories = g.TotalCal
		day.MealCount = g.Count
	}

	// 3. Activity minutes by day
	type ActGroup struct {
		PatientID    string
		DateStr      string
		TotalMinutes int
	}
	var actGroups []ActGroup
	if err := r.db.WithContext(ctx).Raw(`
		SELECT le.patient_id, TO_CHAR(le.logged_at, 'YYYY-MM-DD') as date_str,
			COUNT(*) * 30 as total_minutes
		FROM routine_log_entries le
		JOIN routine_times rt ON rt.id = le.routine_time_id AND rt.deleted_at IS NULL
		JOIN routines r ON r.id = rt.routine_id AND r.deleted_at IS NULL
		WHERE le.patient_id IN (?) AND le.logged_at >= ? AND le.logged_at <= ? AND le.status = 'Completed' AND le.deleted_at IS NULL
		GROUP BY le.patient_id, date_str
	`, patientIDs, startDate, endDate).Scan(&actGroups).Error; err != nil {
		return nil, errs.NewInternal("failed to aggregate activity logs", err)
	}
	for _, g := range actGroups {
		getDay(g.PatientID, g.DateStr).TotalActivityMinutes = g.TotalMinutes
	}

	// 4. Medication completed by day
	type MedGroup struct {
		PatientID      string
		DateStr        string
		CompletedCount int
	}
	var medGroups []MedGroup
	if err := r.db.WithContext(ctx).Raw(`
		SELECT r.patient_id, TO_CHAR(d.log_date, 'YYYY-MM-DD') as date_str, COUNT(*) as completed_count
		FROM daily_reminder_logs d
		JOIN reminders r ON r.id = d.reminder_id AND r.deleted_at IS NULL
		WHERE r.patient_id IN (?) AND d.log_date >= ? AND d.log_date <= ? AND d.status = 'selesai' AND d.deleted_at IS NULL
		GROUP BY r.patient_id, date_str
	`, patientIDs, startDate, endDate).Scan(&medGroups).Error; err != nil {
		return nil, errs.NewInternal("failed to aggregate medication logs", err)
	}
	for _, g := range medGroups {
		getDay(g.PatientID, g.DateStr).MedicationCompletedCount = g.CompletedCount
	}

	// 5. Total active medication reminders count per patient
	type RemGroup struct {
		PatientID string
		Count     int64
	}
	var remGroups []RemGroup
	if err := r.db.WithContext(ctx).Raw(`
		SELECT patient_id, COUNT(*) as count
		FROM reminders
		WHERE patient_id IN (?) AND is_active = true AND category = 'medis_obat' AND deleted_at IS NULL
		GROUP BY patient_id
	`, patientIDs).Scan(&remGroups).Error; err != nil {
		return nil, errs.NewInternal("failed to count active reminders", err)
	}
	reminderCount := make(map[string]int, len(remGroups))
	for _, g := range remGroups {
		reminderCount[g.PatientID] = int(g.Count)
	}
	for pid, days := range result {
		for _, agg := range days {
			agg.MedicationScheduledCount = reminderCount[pid]
		}
	}

	return result, nil
}

func (r *patientRepository) CreateMeasurement(ctx context.Context, m *domain.PatientMeasurement) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(m).Error; err != nil {
			return errs.NewInternal("failed to create measurement record", err)
		}

		// Update patient summary fields (weight_kg, height_cm, daily_calorie_target) with latest measurement
		updates := map[string]interface{}{}
		if m.WeightKg != nil && *m.WeightKg > 0 {
			updates["weight_kg"] = *m.WeightKg
		}
		if m.HeightCm != nil && *m.HeightCm > 0 {
			updates["height_cm"] = *m.HeightCm
		}
		if m.DailyCalorieTarget != nil && *m.DailyCalorieTarget > 0 {
			updates["daily_calorie_target"] = *m.DailyCalorieTarget
		}

		if len(updates) > 0 {
			if err := tx.Model(&domain.Patient{}).Where("id = ?", m.PatientID).Updates(updates).Error; err != nil {
				return errs.NewInternal("failed to update patient summary", err)
			}
		}
		return nil
	})
}

func (r *patientRepository) CreateMeasurementWithSync(ctx context.Context, patient *domain.Patient, m *domain.PatientMeasurement, bsLog *domain.BloodSugarLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(patient).Error; err != nil {
			return errs.NewInternal("failed to update patient profile", err)
		}
		if err := tx.Create(m).Error; err != nil {
			return errs.NewInternal("failed to create measurement record", err)
		}
		if bsLog != nil {
			if err := tx.Create(bsLog).Error; err != nil {
				return errs.NewInternal("failed to create blood sugar log", err)
			}
		}
		return nil
	})
}

func (r *patientRepository) GetPatientMeasurements(ctx context.Context, patientID string) ([]domain.PatientMeasurement, error) {
	var items []domain.PatientMeasurement
	err := r.db.WithContext(ctx).
		Where("patient_id = ? AND deleted_at IS NULL", patientID).
		Order("measured_at DESC, created_at DESC").
		Limit(200). // server-side cap
		Find(&items).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch patient measurements history", err)
	}

	// Fetch blood_sugar_logs to ensure patient-entered blood sugar logs are merged into health measurement history
	var bsLogs []domain.BloodSugarLog
	_ = r.db.WithContext(ctx).
		Where("patient_id = ? AND deleted_at IS NULL", patientID).
		Order("measured_at DESC").
		Limit(200). // server-side cap
		Find(&bsLogs).Error

	if len(bsLogs) > 0 {
		existingIDs := make(map[string]bool)
		for _, m := range items {
			existingIDs[m.ID] = true
		}

		var patientName string = "Pasien"
		var pat domain.Patient
		if pErr := r.db.WithContext(ctx).Select("full_name").Where("id = ?", patientID).First(&pat).Error; pErr == nil && pat.FullName != "" {
			patientName = pat.FullName
		}

		for _, bs := range bsLogs {
			if !existingIDs[bs.ID] {
				val := bs.GlucoseValue
				measType := string(bs.MeasurementTimeType)
				m := domain.PatientMeasurement{
					BaseModel: domain.BaseModel{
						ID:        bs.ID,
						CreatedAt: bs.CreatedAt,
						UpdatedAt: bs.UpdatedAt,
					},
					PatientID:          bs.PatientID,
					BloodSugar:         &val,
					BloodSugarTimeType: &measType,
					RecordedByID:       &bs.PatientID,
					RecordedByName:     patientName,
					RecordedByRole:     "patient",
					MeasuredAt:         bs.MeasuredAt,
				}
				items = append(items, m)
			}
		}

		sort.Slice(items, func(i, j int) bool {
			return items[i].MeasuredAt.After(items[j].MeasuredAt)
		})
	}

	return items, nil
}

func (r *patientRepository) GetLatestMeasurement(ctx context.Context, patientID string) (*domain.PatientMeasurement, error) {
	var item domain.PatientMeasurement
	err := r.db.WithContext(ctx).
		Where("patient_id = ? AND deleted_at IS NULL", patientID).
		Order("measured_at DESC, created_at DESC").
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *patientRepository) FindMeasurementByID(ctx context.Context, measurementID string) (*domain.PatientMeasurement, error) {
	var item domain.PatientMeasurement
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", measurementID).
		First(&item).Error
	if err != nil {
		return nil, errs.NewNotFound("measurement record not found", err)
	}
	return &item, nil
}

func (r *patientRepository) UpdateMeasurement(ctx context.Context, m *domain.PatientMeasurement) error {
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return errs.NewInternal("failed to update measurement record", err)
	}
	return nil
}

func (r *patientRepository) CreateBloodSugarLog(ctx context.Context, bsLog *domain.BloodSugarLog) error {
	if err := r.db.WithContext(ctx).Create(bsLog).Error; err != nil {
		return errs.NewInternal("failed to create blood sugar log entry", err)
	}
	return nil
}
