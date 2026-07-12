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

	if filter.Gender != "" && filter.Gender != "Semua" {
		genderVal := domain.Gender(filter.Gender)
		if filter.Gender == "Laki-laki" || filter.Gender == "laki_laki" {
			genderVal = domain.GenderLakiLaki
		} else if filter.Gender == "Perempuan" || filter.Gender == "perempuan" {
			genderVal = domain.GenderPerempuan
		}
		q = q.Where("gender = ?", genderVal)
	}

	if filter.Status != "" && filter.Status != "Semua" {
		statusVal := domain.AccountStatus(filter.Status)
		if filter.Status == "Aktif" || filter.Status == "aktif" {
			statusVal = domain.StatusAktif
		} else if filter.Status == "Nonaktif" || filter.Status == "nonaktif" {
			statusVal = domain.StatusNonaktif
		}
		q = q.Where("status = ?", statusVal)
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

func (r *patientRepository) GetStats(ctx context.Context, puskesmasID string) (*PatientStats, error) {
	var total int64
	var active int64
	var avgAge float64

	baseQuery := r.db.WithContext(ctx).Model(&domain.Patient{}).Where("deleted_at IS NULL")
	if puskesmasID != "" {
		baseQuery = baseQuery.Where("assigned_puskesmas_id = ?", puskesmasID)
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
