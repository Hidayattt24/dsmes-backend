package facility

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type facilityRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewFacilityRepository(db *gorm.DB, log *zap.Logger) FacilityRepository {
	return &facilityRepository{db: db, log: log}
}

func (r *facilityRepository) FindAll(ctx context.Context, search string, page, limit int) ([]domain.HealthFacility, int64, error) {
	var items []domain.HealthFacility
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.HealthFacility{}).Where("deleted_at IS NULL")

	if search != "" {
		q = q.Where("LOWER(name) LIKE LOWER(?)", "%"+search+"%")
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errs.NewInternal("failed to count health facilities", err)
	}

	offset := (page - 1) * limit
	if err := q.Offset(offset).Limit(limit).Order("name ASC").Find(&items).Error; err != nil {
		return nil, 0, errs.NewInternal("failed to fetch health facilities", err)
	}

	return items, total, nil
}

func (r *facilityRepository) FindByID(ctx context.Context, id string) (*domain.HealthFacility, error) {
	var f domain.HealthFacility
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&f).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("health facility not found")
		}
		return nil, errs.NewInternal("failed to fetch health facility", err)
	}
	return &f, nil
}

func (r *facilityRepository) FindByName(ctx context.Context, name string) (*domain.HealthFacility, error) {
	var f domain.HealthFacility
	err := r.db.WithContext(ctx).Where("LOWER(name) = LOWER(?) AND deleted_at IS NULL", name).First(&f).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("health facility not found")
		}
		return nil, errs.NewInternal("failed to fetch health facility by name", err)
	}
	return &f, nil
}

func (r *facilityRepository) Create(ctx context.Context, f *domain.HealthFacility) error {
	if err := r.db.WithContext(ctx).Create(f).Error; err != nil {
		return errs.NewInternal("failed to create health facility", err)
	}
	return nil
}

func (r *facilityRepository) Update(ctx context.Context, f *domain.HealthFacility) error {
	if err := r.db.WithContext(ctx).Save(f).Error; err != nil {
		return errs.NewInternal("failed to update health facility", err)
	}
	return nil
}

func (r *facilityRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Model(&domain.HealthFacility{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()"))
	if result.Error != nil {
		return errs.NewInternal("failed to soft delete health facility", result.Error)
	}
	if result.RowsAffected == 0 {
		return errs.NewNotFound("health facility not found")
	}
	return nil
}
