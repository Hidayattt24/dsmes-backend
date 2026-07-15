package patient

import (
	"context"
	"errors"
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

func (r *patientRepository) GetStats(ctx context.Context, staffID string) (*PatientStats, error) {
	var total int64
	var active int64
	var avgAge float64

	baseQuery := r.db.WithContext(ctx).Model(&domain.Patient{}).Where("deleted_at IS NULL")
	if staffID != "" {
		baseQuery = baseQuery.Where("assigned_staff_id = ?", staffID)
	}

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, errs.NewInternal("failed to count total patients for stats", err)
	}

	if err := baseQuery.Where("status = ?", domain.StatusAktif).Count(&active).Error; err != nil {
		return nil, errs.NewInternal("failed to count active patients for stats", err)
	}

	var ageResult struct {
		AvgAge float64
	}
	err := baseQuery.Select("COALESCE(AVG(EXTRACT(YEAR FROM AGE(NOW(), date_of_birth))), 0) as avg_age").Scan(&ageResult).Error
	if err != nil {
		return nil, errs.NewInternal("failed to calculate average age", err)
	}
	avgAge = ageResult.AvgAge

	return &PatientStats{
		TotalPatients:  total,
		ActivePatients: active,
		AverageAge:     int(avgAge),
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
	if err == nil {
		val := latestBS.GlucoseValue
		summary.LatestBloodSugar = &val
		summary.LatestBloodSugarTime = &latestBS.MeasuredAt
		statusStr := string(latestBS.Status)
		summary.LatestBloodSugarStatus = &statusStr
	}

	var avgBS float64
	err = r.db.WithContext(ctx).Model(&domain.BloodSugarLog{}).
		Where("patient_id = ? AND deleted_at IS NULL", patientID).
		Select("COALESCE(AVG(glucose_value), 0)").
		Scan(&avgBS).Error
	if err == nil && avgBS > 0 {
		summary.AverageBloodSugar = &avgBS
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
		}
		summary.LatestMealCalories = &cals
		mType := string(latestMeal.MealType)
		summary.LatestMealType = &mType
	}

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
		SELECT r.descriptive_name, le.logged_at
		FROM routine_log_entries le
		JOIN routine_times rt ON rt.id = le.routine_time_id
		JOIN routines r ON r.id = rt.routine_id
		WHERE le.patient_id = ? AND le.status = 'Completed' AND le.deleted_at IS NULL AND rt.deleted_at IS NULL AND r.deleted_at IS NULL
		ORDER BY le.logged_at DESC
		LIMIT 1
	`, patientID).Scan(&actResult).Error
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
	type BSResult struct {
		PatientID    string
		GlucoseValue int
		MeasuredAt   time.Time
		Status       string
	}
	var bsResults []BSResult
	r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (bs.patient_id) bs.patient_id, bs.glucose_value, bs.measured_at, bs.status
		FROM blood_sugar_logs bs
		WHERE bs.patient_id IN ? AND bs.deleted_at IS NULL
		ORDER BY bs.patient_id, bs.measured_at DESC
	`, patientIDs).Scan(&bsResults)

	bsMap := make(map[string]BSResult, len(bsResults))
	for _, bs := range bsResults {
		bsMap[bs.PatientID] = bs
	}

	// Batch fetch average blood sugar
	type AvgBSResult struct {
		PatientID string
		AvgValue  float64
	}
	var avgBSResults []AvgBSResult
	r.db.WithContext(ctx).Raw(`
		SELECT bs.patient_id, COALESCE(AVG(bs.glucose_value), 0) as avg_value
		FROM blood_sugar_logs bs
		WHERE bs.patient_id IN ? AND bs.deleted_at IS NULL
		GROUP BY bs.patient_id
	`, patientIDs).Scan(&avgBSResults)

	avgBSMap := make(map[string]float64, len(avgBSResults))
	for _, a := range avgBSResults {
		avgBSMap[a.PatientID] = a.AvgValue
	}

	// Batch fetch latest meal
	type MealResult struct {
		PatientID string
		Calories  float64
		MealType  string
	}
	var mealResults []MealResult
	r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (ml.patient_id) ml.patient_id,
			COALESCE(f.calories * ml.portion_multiplier, 0) as calories,
			ml.meal_type
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
		SELECT DISTINCT ON (le.patient_id) le.patient_id,
			r.descriptive_name, le.logged_at
		FROM routine_log_entries le
		JOIN routine_times rt ON rt.id = le.routine_time_id AND rt.deleted_at IS NULL
		JOIN routines r ON r.id = rt.routine_id AND r.deleted_at IS NULL
		WHERE le.patient_id IN ? AND le.status = 'Completed' AND le.deleted_at IS NULL
		ORDER BY le.patient_id, le.logged_at DESC
	`, patientIDs).Scan(&actResults)

	actMap := make(map[string]ActResult, len(actResults))
	for _, a := range actResults {
		actMap[a.PatientID] = a
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
		}

		if a, ok := actMap[pid]; ok && a.DescriptiveName != "" {
			name := a.DescriptiveName
			summary.LatestActivityName = &name
			summary.LatestActivityTime = &a.LoggedAt
		}

		// Get weight and height for BMI (reuse patient data already fetched)
		// BMI will be empty in batch — callers already have patient weight/height
		result[pid] = summary
	}

	return result, nil
}
