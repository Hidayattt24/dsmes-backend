package settings

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type settingsRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewSettingsRepository(db *gorm.DB, log *zap.Logger) SettingsRepository {
	return &settingsRepository{db: db, log: log}
}

func (r *settingsRepository) FindAllFAQs(ctx context.Context) ([]domain.FAQ, error) {
	var items []domain.FAQ
	err := r.db.WithContext(ctx).
		Where("is_active = true AND deleted_at IS NULL").
		Order("display_order ASC").
		Find(&items).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch FAQs", err)
	}
	return items, nil
}

func (r *settingsRepository) CreateTicket(ctx context.Context, t *domain.SupportTicket) error {
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return errs.NewInternal("failed to create support ticket", err)
	}
	return nil
}

func (r *settingsRepository) FindTicketsByPatientID(ctx context.Context, patientID string) ([]domain.SupportTicket, error) {
	var items []domain.SupportTicket
	err := r.db.WithContext(ctx).
		Where("patient_id = ? AND deleted_at IS NULL", patientID).
		Order("created_at DESC").
		Find(&items).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch patient support tickets", err)
	}
	return items, nil
}

func (r *settingsRepository) FindAllTickets(ctx context.Context) ([]domain.SupportTicket, error) {
	var items []domain.SupportTicket
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Find(&items).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch all support tickets", err)
	}
	return items, nil
}

func (r *settingsRepository) FindTicketByID(ctx context.Context, id string) (*domain.SupportTicket, error) {
	var t domain.SupportTicket
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("support ticket not found")
		}
		return nil, errs.NewInternal("failed to fetch support ticket", err)
	}
	return &t, nil
}

func (r *settingsRepository) UpdateTicket(ctx context.Context, t *domain.SupportTicket) error {
	result := r.db.WithContext(ctx).Save(t)
	if result.Error != nil {
		return errs.NewInternal("failed to update support ticket", result.Error)
	}
	return nil
}
