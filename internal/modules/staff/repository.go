package staff

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type staffRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewStaffRepository(db *gorm.DB, log *zap.Logger) StaffRepository {
	return &staffRepository{db: db, log: log}
}

func (r *staffRepository) FindAll(ctx context.Context, search, status string, role *domain.StaffRole, page, limit int) ([]domain.StaffAccount, int64, error) {
	var items []domain.StaffAccount
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.StaffAccount{}).Where("deleted_at IS NULL")

	if search != "" {
		q = q.Where("(LOWER(full_name) LIKE LOWER(?) OR LOWER(username) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?))", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if status != "" {
		q = q.Where("LOWER(status) = LOWER(?)", status)
	}

	if role != nil && *role != "" {
		q = q.Where("role = ?", *role)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errs.NewInternal("failed to count staff", err)
	}

	offset := (page - 1) * limit
	if err := q.Preload("HealthFacility").Offset(offset).Limit(limit).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, 0, errs.NewInternal("failed to fetch staff", err)
	}

	return items, total, nil
}

func (r *staffRepository) FindByID(ctx context.Context, id string) (*domain.StaffAccount, error) {
	var s domain.StaffAccount
	err := r.db.WithContext(ctx).Preload("HealthFacility").Where("id = ? AND deleted_at IS NULL", id).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("staff not found")
		}
		return nil, errs.NewInternal("failed to fetch staff", err)
	}
	return &s, nil
}

func (r *staffRepository) FindByEmail(ctx context.Context, email string) (*domain.StaffAccount, error) {
	var s domain.StaffAccount
	err := r.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("staff not found")
		}
		return nil, errs.NewInternal("failed to fetch staff by email", err)
	}
	return &s, nil
}

func (r *staffRepository) FindByUsername(ctx context.Context, username string) (*domain.StaffAccount, error) {
	var s domain.StaffAccount
	err := r.db.WithContext(ctx).Where("username = ? AND deleted_at IS NULL", username).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("staff not found")
		}
		return nil, errs.NewInternal("failed to fetch staff by username", err)
	}
	return &s, nil
}

func (r *staffRepository) Create(ctx context.Context, s *domain.StaffAccount) error {
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		return errs.NewInternal("failed to create staff", err)
	}
	return nil
}

func (r *staffRepository) Update(ctx context.Context, s *domain.StaffAccount) error {
	result := r.db.WithContext(ctx).Save(s)
	if result.Error != nil {
		return errs.NewInternal("failed to update staff", result.Error)
	}
	return nil
}

func (r *staffRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Model(&domain.StaffAccount{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()"))
	if result.Error != nil {
		return errs.NewInternal("failed to soft delete staff", result.Error)
	}
	if result.RowsAffected == 0 {
		return errs.NewNotFound("staff not found")
	}
	return nil
}
