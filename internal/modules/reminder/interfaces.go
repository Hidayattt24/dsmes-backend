package reminder

import (
	"context"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type ReminderRepository interface {
	FindAllByPatientID(ctx context.Context, patientID string) ([]domain.Reminder, error)
	FindByID(ctx context.Context, id string) (*domain.Reminder, error)
	Create(ctx context.Context, r *domain.Reminder, activeDays []int) error
	Update(ctx context.Context, r *domain.Reminder, activeDays []int) error
	Delete(ctx context.Context, id string) error

	FindNotificationsByPatientID(ctx context.Context, patientID string) ([]domain.NotificationLog, error)
	MarkNotificationsAsRead(ctx context.Context, patientID string) error
	FindLogsByPatientAndDate(ctx context.Context, patientID string, dateStr string) ([]domain.DailyReminderLog, error)
}

type ReminderService interface {
	ListReminders(ctx context.Context, patientID string) ([]ReminderResponse, error)
	CreateReminder(ctx context.Context, patientID string, req CreateReminderRequest) (*ReminderResponse, error)
	UpdateReminder(ctx context.Context, patientID string, id string, req CreateReminderRequest) (*ReminderResponse, error)
	DeleteReminder(ctx context.Context, patientID string, id string) error
	ToggleReminder(ctx context.Context, patientID string, id string) (*ReminderResponse, error)

	GetNotifications(ctx context.Context, patientID string) ([]NotificationResponse, error)
	MarkAllRead(ctx context.Context, patientID string) error
	GetPatientMedicationLogs(ctx context.Context, patientID string, dateStr string) ([]MedicationLogResponse, error)
}
