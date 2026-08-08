package checkin

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type checkinRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewCheckinRepository(db *gorm.DB, log *zap.Logger) CheckinRepository {
	return &checkinRepository{db: db, log: log}
}

func (r *checkinRepository) Upsert(ctx context.Context, c *domain.DailyMedicalCheckin) error {
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "patient_id"}, {Name: "checkin_date"}},
		DoUpdates: clause.AssignmentColumns([]string{"is_completed", "updated_at"}),
	}).Create(c).Error
	if err != nil {
		return errs.NewInternal("failed to upsert checkin", err)
	}
	return nil
}

func (r *checkinRepository) FindByMonth(ctx context.Context, patientID string, year int, month int) ([]domain.DailyMedicalCheckin, error) {
	var items []domain.DailyMedicalCheckin
	// Range predicate over checkin_date instead of EXTRACT(YEAR/MONTH ...) so the
	// UNIQUE(patient_id, checkin_date) index can be used.
	err := r.db.WithContext(ctx).
		Where("patient_id = ? AND checkin_date >= make_date(?, ?, 1) AND checkin_date < (make_date(?, ?, 1) + INTERVAL '1 month') AND deleted_at IS NULL",
			patientID, year, month, year, month).
		Find(&items).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch monthly checkins", err)
	}
	return items, nil
}
