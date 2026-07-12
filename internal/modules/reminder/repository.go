package reminder

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type reminderRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewReminderRepository(db *gorm.DB, log *zap.Logger) ReminderRepository {
	return &reminderRepository{db: db, log: log}
}

func (r *reminderRepository) FindAllByPatientID(ctx context.Context, patientID string) ([]domain.Reminder, error) {
	var items []domain.Reminder
	err := r.db.WithContext(ctx).
		Preload("ActiveDays").
		Where("patient_id = ? AND deleted_at IS NULL", patientID).
		Find(&items).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch reminders", err)
	}
	return items, nil
}

func (r *reminderRepository) FindByID(ctx context.Context, id string) (*domain.Reminder, error) {
	var rem domain.Reminder
	err := r.db.WithContext(ctx).
		Preload("ActiveDays").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&rem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("reminder not found")
		}
		return nil, errs.NewInternal("failed to fetch reminder", err)
	}
	return &rem, nil
}

func (r *reminderRepository) Create(ctx context.Context, rem *domain.Reminder, activeDays []int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(rem).Error; err != nil {
			return err
		}

		for _, d := range activeDays {
			ad := domain.ReminderActiveDay{
				ReminderID: rem.ID,
				DayOfWeek:  d,
			}
			if err := tx.Create(&ad).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *reminderRepository) Update(ctx context.Context, rem *domain.Reminder, activeDays []int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(rem).Error; err != nil {
			return err
		}

		// Delete existing active days first
		if err := tx.Where("reminder_id = ?", rem.ID).Delete(&domain.ReminderActiveDay{}).Error; err != nil {
			return err
		}

		// Re-insert new active days
		for _, d := range activeDays {
			ad := domain.ReminderActiveDay{
				ReminderID: rem.ID,
				DayOfWeek:  d,
			}
			if err := tx.Create(&ad).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *reminderRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Model(&domain.Reminder{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()"))
	if result.Error != nil {
		return errs.NewInternal("failed to soft delete reminder", result.Error)
	}
	if result.RowsAffected == 0 {
		return errs.NewNotFound("reminder not found")
	}
	return nil
}

func (r *reminderRepository) FindNotificationsByPatientID(ctx context.Context, patientID string) ([]domain.NotificationLog, error) {
	var items []domain.NotificationLog
	err := r.db.WithContext(ctx).
		Where("patient_id = ? AND deleted_at IS NULL", patientID).
		Order("notified_at DESC").
		Limit(50).
		Find(&items).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch notifications", err)
	}
	return items, nil
}

func (r *reminderRepository) MarkNotificationsAsRead(ctx context.Context, patientID string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.NotificationLog{}).
		Where("patient_id = ? AND is_read = false", patientID).
		Update("is_read", true)
	if result.Error != nil {
		return errs.NewInternal("failed to mark notifications as read", result.Error)
	}
	return nil
}

func (r *reminderRepository) FindLogsByPatientAndDate(ctx context.Context, patientID string, dateStr string) ([]domain.DailyReminderLog, error) {
	var logs []domain.DailyReminderLog
	err := r.db.WithContext(ctx).
		Joins("JOIN reminders ON reminders.id = daily_reminder_logs.reminder_id").
		Where("reminders.patient_id = ? AND daily_reminder_logs.log_date = ? AND daily_reminder_logs.deleted_at IS NULL", patientID, dateStr).
		Find(&logs).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch daily reminder logs", err)
	}
	return logs, nil
}

