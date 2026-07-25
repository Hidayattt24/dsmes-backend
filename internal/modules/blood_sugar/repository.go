package blood_sugar

import (
	"context"

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
	err := q.Offset(offset).Limit(limit).Order("measured_at DESC").Find(&items).Error
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

func (r *bloodSugarRepository) GetDistributionForStaff(ctx context.Context, staffID string) (*GlucoseDistributionResponse, error) {
	var res GlucoseDistributionResponse
	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'normal'), 0) AS normal_count,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'tinggi'), 0) AS tinggi_count,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'sangat_tinggi'), 0) AS sangat_tinggi_count,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'rendah'), 0) AS rendah_count
		FROM blood_sugar_logs b
		JOIN patients p ON p.id = b.patient_id
		WHERE p.assigned_staff_id = ? AND b.deleted_at IS NULL AND p.deleted_at IS NULL
	`, staffID).Scan(&res).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch distribution", err)
	}
	return &res, nil
}
