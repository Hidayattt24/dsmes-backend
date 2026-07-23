package routine

import "github.com/dsmes/dsmes-backend/internal/domain"

type UpdateRoutineTimeRequest struct {
	TimeType       domain.WaktuType   `json:"time_type"       validate:"required,oneof=Default Kustom"`
	ScheduledTime  *string            `json:"scheduled_time"` // "HH:MM:SS" or null
	Status         domain.WaktuStatus `json:"status"          validate:"required,oneof=Set Unset"`
	ReminderActive bool               `json:"reminder_active"`
}

type LogRoutineRequest struct {
	RoutineTimeID string                  `json:"routine_time_id" validate:"required,uuid4"`
	Status        domain.RoutineLogStatus `json:"status"          validate:"required,oneof=Completed Skipped Pending"`
}

type RoutineResponse struct {
	ID              string                `json:"id"`
	RoutineType     domain.RoutineType    `json:"routine_type"`
	DescriptiveName string                `json:"descriptive_name"`
	IconName        string                `json:"icon_name"`
	ScheduleText    string                `json:"schedule_text"`
	BaseFrequency   string                `json:"base_frequency"`
	IsActive        bool                  `json:"is_active"`
	RoutineTimes    []RoutineTimeResponse `json:"routine_times"`
}

type RoutineTimeResponse struct {
	ID             string             `json:"id"`
	TimeType       domain.WaktuType   `json:"time_type"`
	ScheduledTime  *string            `json:"scheduled_time"`
	Status         domain.WaktuStatus `json:"status"`
	ReminderActive bool               `json:"reminder_active"`
}

type RoutineSetupItem struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	IconName     string   `json:"icon_name"`
	ScheduleText string   `json:"schedule_text"`
	IsPredefined bool     `json:"is_predefined"`
	CustomTimes  []string `json:"custom_times"`
}

type BulkSetupRoutinesRequest struct {
	UseReminder bool               `json:"use_reminder"`
	Routines    []RoutineSetupItem `json:"routines"`
}

type RoutineLogResponse struct {
	ID            string                  `json:"id"`
	RoutineTimeID string                  `json:"routine_time_id"`
	LoggedAt      string                  `json:"logged_at"`
	Status        domain.RoutineLogStatus `json:"status"`
}

type OnboardingStatusResponse struct {
	IsReady        bool `json:"is_ready"`
	ReminderActive bool `json:"reminder_active"`
}

func ToRoutineResponse(r *domain.Routine) RoutineResponse {
	times := make([]RoutineTimeResponse, len(r.RoutineTimes))
	for i, t := range r.RoutineTimes {
		times[i] = RoutineTimeResponse{
			ID:             t.ID,
			TimeType:       t.TimeType,
			ScheduledTime:  t.ScheduledTime,
			Status:         t.Status,
			ReminderActive: t.ReminderActive,
		}
	}
	return RoutineResponse{
		ID:              r.ID,
		RoutineType:     r.RoutineType,
		DescriptiveName: r.DescriptiveName,
		IconName:        r.IconName,
		ScheduleText:    r.ScheduleText,
		BaseFrequency:   r.BaseFrequency,
		IsActive:        r.IsActive,
		RoutineTimes:    times,
	}
}

type ActivityLogResponse struct {
	ID              string                  `json:"id"`
	RoutineType     domain.RoutineType      `json:"routine_type"`
	DescriptiveName string                  `json:"descriptive_name"`
	ScheduledTime   *string                 `json:"scheduled_time"`
	Status          domain.RoutineLogStatus `json:"status"`
	LoggedAt        string                  `json:"logged_at"`
}
