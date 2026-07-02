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

	return items, total, nil
}

func (r *bloodSugarRepository) GetDistributionForPuskesmas(ctx context.Context, puskesmasID string) (*GlucoseDistributionResponse, error) {
	var res GlucoseDistributionResponse
	err := r.db.WithContext(ctx).Raw(`
		SELECT 
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'normal'), 0) AS normal_count,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'tinggi'), 0) AS tinggi_count,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'sangat_tinggi'), 0) AS sangat_tinggi_count,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'rendah'), 0) AS rendah_count
		FROM blood_sugar_logs b
		JOIN patients p ON p.id = b.patient_id
		WHERE p.assigned_puskesmas_id = ? AND b.deleted_at IS NULL AND p.deleted_at IS NULL
	`, puskesmasID).Scan(&res).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch distribution", err)
	}
	return &res, nil
}
