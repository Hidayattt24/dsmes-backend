package reminder

import (
	"github.com/dsmes/dsmes-backend/internal/domain"
)

type CreateReminderRequest struct {
	ActivityName       string                  `json:"activity_name"        validate:"required,min=2,max=150"`
	Category           domain.ReminderCategory `json:"category"             validate:"required,oneof=medis_obat nutrisi_air aktivitas_fisik lainnya"`
	ScheduledTime      string                  `json:"scheduled_time"       validate:"required"` // format: "HH:MM:SS"
	Notes              string                  `json:"notes"`
	RepeatIntervalDays int                     `json:"repeat_interval_days" validate:"required,min=1"`
	ActiveDays         []int                   `json:"active_days"          validate:"required,min=1"` // 1-7
}

type ReminderResponse struct {
	ID                 string                  `json:"id"`
	ActivityName       string                  `json:"activity_name"`
	ReminderType       domain.ReminderType     `json:"reminder_type"`
	Category           domain.ReminderCategory `json:"category"`
	ScheduledTime      string                  `json:"scheduled_time"`
	IsActive           bool                    `json:"is_active"`
	Notes              string                  `json:"notes"`
	RepeatIntervalDays int                     `json:"repeat_interval_days"`
	ActiveDays         []int                   `json:"active_days"`
}

type NotificationResponse struct {
	ID          string `json:"id"`
	ReminderID  *string `json:"reminder_id"`
	MessageText string `json:"message_text"`
	NotifiedAt  string `json:"notified_at"`
	IsRead      bool   `json:"is_read"`
}

func ToReminderResponse(r *domain.Reminder) ReminderResponse {
	days := make([]int, len(r.ActiveDays))
	for i, d := range r.ActiveDays {
		days[i] = d.DayOfWeek
	}
	return ReminderResponse{
		ID:                 r.ID,
		ActivityName:       r.ActivityName,
		ReminderType:       r.ReminderType,
		Category:           r.Category,
		ScheduledTime:      r.ScheduledTime,
		IsActive:           r.IsActive,
		Notes:              r.Notes,
		RepeatIntervalDays: r.RepeatIntervalDays,
		ActiveDays:         days,
	}
}

type MedicationLogResponse struct {
	ID            string                  `json:"id"`
	ReminderID    string                  `json:"reminder_id"`
	ActivityName  string                  `json:"activity_name"`
	Category      domain.ReminderCategory `json:"category"`
	ScheduledTime string                  `json:"scheduled_time"`
	Status        domain.ReminderLogStatus `json:"status"`
	LoggedDate    string                  `json:"logged_date"`
}
