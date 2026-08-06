package routine

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type routineRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewRoutineRepository(db *gorm.DB, log *zap.Logger) RoutineRepository {
	return &routineRepository{db: db, log: log}
}

func (r *routineRepository) FindAllByPatientID(ctx context.Context, patientID string) ([]domain.Routine, error) {
	var items []domain.Routine
	err := r.db.WithContext(ctx).
		Preload("RoutineTimes").
		Where("patient_id = ? AND deleted_at IS NULL", patientID).
		Find(&items).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch routines", err)
	}
	return items, nil
}

func (r *routineRepository) FindTimeByID(ctx context.Context, id string) (*domain.RoutineTime, error) {
	var t domain.RoutineTime
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("routine time not found")
		}
		return nil, errs.NewInternal("failed to fetch routine time", err)
	}
	return &t, nil
}

func (r *routineRepository) UpdateTime(ctx context.Context, t *domain.RoutineTime) error {
	result := r.db.WithContext(ctx).Save(t)
	if result.Error != nil {
		return errs.NewInternal("failed to update routine time", result.Error)
	}
	return nil
}

func (r *routineRepository) CreateLog(ctx context.Context, log *domain.RoutineLogEntry) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return errs.NewInternal("failed to create routine log entry", err)
	}
	return nil
}

func (r *routineRepository) FindLogsByPatientAndDate(ctx context.Context, patientID string, dateStr string) ([]domain.RoutineLogEntry, error) {
	var logs []domain.RoutineLogEntry
	q := r.db.WithContext(ctx).Preload("RoutineTime").Preload("RoutineTime.Routine").Where("patient_id = ? AND deleted_at IS NULL", patientID)
	if dateStr != "" {
		// Range predicate instead of DATE(col) = ? so the (patient_id, logged_at)
		// index can be used.
		q = q.Where("logged_at >= ?::date AND logged_at < (?::date + INTERVAL '1 day')", dateStr, dateStr)
	} else {
		q = q.Where("logged_at >= NOW() - INTERVAL '30 days'")
	}
	err := q.Order("logged_at DESC").Find(&logs).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch routine logs", err)
	}
	return logs, nil
}

func (r *routineRepository) FindFreeActivityLogsByPatientAndDate(ctx context.Context, patientID string, dateStr string) ([]domain.PatientActivityLog, error) {
	var logs []domain.PatientActivityLog
	q := r.db.WithContext(ctx).Where("patient_id = ? AND deleted_at IS NULL", patientID)
	if dateStr != "" {
		q = q.Where("logged_at >= ?::date AND logged_at < (?::date + INTERVAL '1 day')", dateStr, dateStr)
	}
	err := q.Order("logged_at DESC").Find(&logs).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch free activity logs", err)
	}
	return logs, nil
}

func (r *routineRepository) CreateActivityLog(ctx context.Context, log *domain.PatientActivityLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return errs.NewInternal("failed to create activity log", err)
	}
	return nil
}

func (r *routineRepository) ReplacePatientRoutines(ctx context.Context, patientID string, routines []domain.Routine) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete previous routines for this patient
		var existingIDs []string
		tx.Model(&domain.Routine{}).Where("patient_id = ?", patientID).Pluck("id", &existingIDs)

		if len(existingIDs) > 0 {
			tx.Where("routine_id IN ?", existingIDs).Delete(&domain.RoutineTime{})
			tx.Where("patient_id = ?", patientID).Delete(&domain.Routine{})
		}

		for i := range routines {
			routines[i].PatientID = patientID
			if err := tx.Create(&routines[i]).Error; err != nil {
				return errs.NewInternal("failed to save routine setup", err)
			}
		}
		return nil
	})
}
