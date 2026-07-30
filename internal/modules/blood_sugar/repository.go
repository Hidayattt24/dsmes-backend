package blood_sugar

import (
	"context"
	"errors"
	"math"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type bloodSugarRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewBloodSugarRepository(db *gorm.DB, log *zap.Logger) BloodSugarRepository {
	return &bloodSugarRepository{db: db, log: log}
}

func (r *bloodSugarRepository) Create(ctx context.Context, log *domain.BloodSugarLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return errs.NewInternal("failed to log blood sugar", err)
	}

	// Create an immutable snapshot in patient_measurements (INSERT ONLY)
	var patient domain.Patient
	var weight *float64
	var height *float64
	var bmi *float64
	var waist *float64
	var calorieTarget *int
	var patientName string = "Pasien"

	if pErr := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", log.PatientID).First(&patient).Error; pErr == nil {
		if patient.FullName != "" {
			patientName = patient.FullName
		}
		if patient.WeightKg > 0 {
			w := patient.WeightKg
			weight = &w
		}
		if patient.HeightCm > 0 {
			h := patient.HeightCm
			height = &h
		}
		if patient.WeightKg > 0 && patient.HeightCm > 0 {
			hM := patient.HeightCm / 100.0
			val := math.Round((patient.WeightKg/(hM*hM))*10) / 10
			bmi = &val
		}
		if patient.DailyCalorieTarget > 0 {
			t := patient.DailyCalorieTarget
			calorieTarget = &t
		}
	}

	// Fetch latest patient_measurement to get latest waist circumference or blood pressure if patient profile didn't store it
	var latestM domain.PatientMeasurement
	if mErr := r.db.WithContext(ctx).
		Where("patient_id = ? AND deleted_at IS NULL", log.PatientID).
		Order("measured_at DESC, created_at DESC").
		First(&latestM).Error; mErr == nil {
		if weight == nil && latestM.WeightKg != nil {
			weight = latestM.WeightKg
		}
		if height == nil && latestM.HeightCm != nil {
			height = latestM.HeightCm
		}
		if bmi == nil && latestM.BMI != nil {
			bmi = latestM.BMI
		}
		if calorieTarget == nil && latestM.DailyCalorieTarget != nil {
			calorieTarget = latestM.DailyCalorieTarget
		}
		if latestM.WaistCircumferenceCm != nil {
			waist = latestM.WaistCircumferenceCm
		}
	}

	val := log.GlucoseValue
	mTypeStr := string(log.MeasurementTimeType)

	snapshot := &domain.PatientMeasurement{
		BaseModel: domain.BaseModel{
			ID: log.ID,
		},
		PatientID:            log.PatientID,
		WeightKg:             weight,
		HeightCm:             height,
		BMI:                  bmi,
		BloodSugar:           &val,
		BloodSugarTimeType:   &mTypeStr,
		WaistCircumferenceCm: waist,
		DailyCalorieTarget:   calorieTarget,
		Notes:                "Pencatatan Mandiri Gula Darah Pasien",
		RecordedByID:         &log.PatientID,
		RecordedByName:       patientName,
		RecordedByRole:       "patient",
		MeasuredAt:           log.MeasuredAt,
	}

	_ = r.db.WithContext(ctx).Create(snapshot).Error

	return nil
}

func (r *bloodSugarRepository) FindByID(ctx context.Context, id string) (*domain.BloodSugarLog, error) {
	var log domain.BloodSugarLog
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&log).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("blood sugar log not found")
		}
		return nil, errs.NewInternal("failed to fetch blood sugar log", err)
	}
	return &log, nil
}

func (r *bloodSugarRepository) Update(ctx context.Context, log *domain.BloodSugarLog) error {
	if err := r.db.WithContext(ctx).Save(log).Error; err != nil {
		return errs.NewInternal("failed to update blood sugar log", err)
	}
	return nil
}

func (r *bloodSugarRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Exec("DELETE FROM blood_sugar_logs WHERE id = ?", id)
	if result.Error != nil {
		return errs.NewInternal("failed to delete blood sugar log", result.Error)
	}
	if result.RowsAffected == 0 {
		return errs.NewNotFound("blood sugar log not found")
	}
	return nil
}

func (r *bloodSugarRepository) FindAllByPatientID(ctx context.Context, patientID string, page, limit int) ([]domain.BloodSugarLog, int64, error) {
	var items []domain.BloodSugarLog
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.BloodSugarLog{}).Where("patient_id = ? AND deleted_at IS NULL", patientID)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errs.NewInternal("failed to count blood sugar logs", err)
	}

	offset := (page - 1) * limit
	err := q.Offset(offset).Limit(limit).Order("measured_at DESC, created_at DESC").Find(&items).Error
	if err != nil {
		return nil, 0, errs.NewInternal("failed to fetch blood sugar logs", err)
	}

	if len(items) == 0 {
		var measurements []domain.PatientMeasurement
		mErr := r.db.WithContext(ctx).
			Where("patient_id = ? AND blood_sugar IS NOT NULL AND blood_sugar > 0 AND deleted_at IS NULL", patientID).
			Order("measured_at DESC, created_at DESC").
			Find(&measurements).Error
		if mErr == nil && len(measurements) > 0 {
			for _, m := range measurements {
				if m.BloodSugar != nil && *m.BloodSugar > 0 {
					status := domain.CalculateGlucoseStatus(*m.BloodSugar, domain.TimeSewaktu)
					items = append(items, domain.BloodSugarLog{
						BaseModel: domain.BaseModel{
							ID:        m.ID,
							CreatedAt: m.CreatedAt,
							UpdatedAt: m.UpdatedAt,
						},
						PatientID:           m.PatientID,
						GlucoseValue:        *m.BloodSugar,
						MeasurementTimeType: domain.TimeSewaktu,
						MeasuredAt:          m.MeasuredAt,
						Status:              status,
					})
				}
			}
			total = int64(len(items))
		}
	}

	return items, total, nil
}

func (r *bloodSugarRepository) GetPatientDOB(ctx context.Context, patientID string) (*time.Time, error) {
	var patient domain.Patient
	err := r.db.WithContext(ctx).Select("date_of_birth").Where("id = ? AND deleted_at IS NULL", patientID).First(&patient).Error
	if err != nil {
		return nil, nil
	}
	return &patient.DateOfBirth, nil
}

func (r *bloodSugarRepository) GetDistributionForStaff(ctx context.Context, staffID string) (*GlucoseDistributionResponse, error) {
	var res GlucoseDistributionResponse
	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'normal'), 0) AS normal_count,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'hyperglycemia' OR b.status = 'tinggi'), 0) AS tinggi_count,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'severe_hyperglycemia' OR b.status = 'sangat_tinggi'), 0) AS sangat_tinggi_count,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'hypoglycemia' OR b.status = 'severe_hypoglycemia' OR b.status = 'rendah'), 0) AS rendah_count
		FROM blood_sugar_logs b
		JOIN patients p ON p.id = b.patient_id
		WHERE p.assigned_staff_id = ? AND b.deleted_at IS NULL AND p.deleted_at IS NULL
	`, staffID).Scan(&res).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch distribution", err)
	}

	type TypeRow struct {
		MeasurementTimeType string  `gorm:"column:measurement_time_type"`
		AverageValue        float64 `gorm:"column:average_value"`
		Count               int64   `gorm:"column:count"`
	}
	var rows []TypeRow
	_ = r.db.WithContext(ctx).Raw(`
		SELECT 
			b.measurement_time_type,
			COALESCE(ROUND(AVG(b.glucose_value), 1), 0) AS average_value,
			COUNT(*) AS count
		FROM blood_sugar_logs b
		JOIN patients p ON p.id = b.patient_id
		WHERE p.assigned_staff_id = ? AND b.deleted_at IS NULL AND p.deleted_at IS NULL
		GROUP BY b.measurement_time_type
	`, staffID).Scan(&rows).Error

	res.ByMeasurementType = make([]MeasurementTypeStats, 0, len(rows))
	for _, row := range rows {
		mType := domain.NormalizeMeasurementType(row.MeasurementTimeType)
		res.ByMeasurementType = append(res.ByMeasurementType, MeasurementTypeStats{
			MeasurementType: string(mType),
			TypeLabel:       domain.GetMeasurementTypeLabel(mType),
			AverageValue:    row.AverageValue,
			Count:           row.Count,
		})
	}

	return &res, nil
}
