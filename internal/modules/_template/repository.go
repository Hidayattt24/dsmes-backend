package _template

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

// templateRepository is the concrete GORM implementation of TemplateRepository.
// It is unexported — callers only see the TemplateRepository interface.
type templateRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewTemplateRepository creates a new repository instance.
// Inject *gorm.DB from the container; never use a global DB variable.
func NewTemplateRepository(db *gorm.DB, log *zap.Logger) TemplateRepository {
	return &templateRepository{db: db, log: log}
}

// FindAll returns a paginated slice of Template records and the total count.
func (r *templateRepository) FindAll(ctx context.Context, page, limit int) ([]Template, int64, error) {
	var items []Template
	var total int64

	offset := (page - 1) * limit

	if err := r.db.WithContext(ctx).Model(&Template{}).Count(&total).Error; err != nil {
		return nil, 0, errs.NewInternal("failed to count templates", err)
	}

	if err := r.db.WithContext(ctx).
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, 0, errs.NewInternal("failed to fetch templates", err)
	}

	return items, total, nil
}

// FindByID fetches a single Template by UUID primary key.
func (r *templateRepository) FindByID(ctx context.Context, id string) (*Template, error) {
	var item Template
	if err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("template not found")
		}
		return nil, errs.NewInternal("failed to fetch template", err)
	}
	return &item, nil
}

// Create inserts a new Template record.
func (r *templateRepository) Create(ctx context.Context, t *Template) error {
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return errs.NewInternal("failed to create template", err)
	}
	return nil
}

// Update saves changes to an existing Template record.
func (r *templateRepository) Update(ctx context.Context, t *Template) error {
	result := r.db.WithContext(ctx).Save(t)
	if result.Error != nil {
		return errs.NewInternal("failed to update template", result.Error)
	}
	if result.RowsAffected == 0 {
		return errs.NewNotFound("template not found")
	}
	return nil
}

// Delete soft-deletes a Template by UUID.
func (r *templateRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&Template{})
	if result.Error != nil {
		return errs.NewInternal("failed to delete template", result.Error)
	}
	if result.RowsAffected == 0 {
		return errs.NewNotFound("template not found")
	}
	return nil
}
