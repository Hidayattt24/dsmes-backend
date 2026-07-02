package patient

import (
	"context"
	"errors"

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

	q := r.db.WithContext(ctx).Model(&domain.Patient{}).Where("deleted_at IS NULL")

	if filter.PuskesmasID != "" {
		q = q.Where("assigned_puskesmas_id = ?", filter.PuskesmasID)
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		q = q.Where("full_name ILIKE ? OR email ILIKE ?", searchPattern, searchPattern)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errs.NewInternal("failed to count patients", err)
	}

	offset := (filter.Page - 1) * filter.Limit
	err := q.Preload("AssignedPuskesmas").
		Offset(offset).Limit(filter.Limit).
		Order("created_at DESC").
		Find(&items).Error
	if err != nil {
		return nil, 0, errs.NewInternal("failed to fetch patients", err)
	}

	return items, total, nil
}

func (r *patientRepository) FindByID(ctx context.Context, id string) (*domain.Patient, error) {
	var p domain.Patient
	err := r.db.WithContext(ctx).
		Preload("AssignedPuskesmas").
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
		Preload("AssignedPuskesmas").
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
