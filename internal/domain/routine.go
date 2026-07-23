package domain

import (
	"time"
)

type RoutineType string

const (
	RoutineJalanPagi RoutineType = "Jalan_Pagi"
	RoutineMinumAir  RoutineType = "Minum_Air"
	RoutineCekGula   RoutineType = "Cek_Gula"
)

type WaktuType string

const (
	WaktuDefault WaktuType = "Default"
	WaktuKustom  WaktuType = "Kustom"
)

type WaktuStatus string

const (
	WaktuSet   WaktuStatus = "Set"
	WaktuUnset WaktuStatus = "Unset"
)

type RoutineLogStatus string

const (
	LogCompleted RoutineLogStatus = "Completed"
	LogSkipped   RoutineLogStatus = "Skipped"
	LogPending   RoutineLogStatus = "Pending"
)

// Routine represents a daily habit routine of a patient.
type Routine struct {
	BaseModel

	PatientID       string      `gorm:"type:uuid;not null" json:"patient_id"`
	RoutineType     RoutineType `gorm:"type:varchar(50);not null" json:"routine_type"`
	DescriptiveName string      `gorm:"type:varchar(150)" json:"descriptive_name"`
	IconName        string      `gorm:"type:varchar(50)" json:"icon_name"`
	ScheduleText    string      `gorm:"type:varchar(255)" json:"schedule_text"`
	BaseFrequency   string      `gorm:"type:varchar(50);not null;default:'Daily'" json:"base_frequency"`
	IsActive        bool        `gorm:"not null;default:true" json:"is_active"`


	// Relations
	Patient      *Patient      `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	RoutineTimes []RoutineTime `gorm:"foreignKey:RoutineID;constraint:OnDelete:CASCADE" json:"routine_times,omitempty"`
}

func (Routine) TableName() string { return "routines" }

// RoutineTime represents scheduled times for a Routine.
type RoutineTime struct {
	BaseModel

	RoutineID      string      `gorm:"type:uuid;not null" json:"routine_id"`
	TimeType       WaktuType   `gorm:"type:waktu_type_enum;not null" json:"time_type"`
	ScheduledTime  *string     `gorm:"type:time" json:"scheduled_time"` // time string e.g. "08:00:00"
	Status         WaktuStatus `gorm:"type:waktu_status_enum;not null;default:Unset" json:"status"`
	ReminderActive bool        `gorm:"not null;default:false" json:"reminder_active"`
}

func (RoutineTime) TableName() string { return "routine_times" }

// RoutineLogEntry tracks whether a routine was completed, skipped or pending for a scheduled time.
type RoutineLogEntry struct {
	BaseModel

	PatientID     string           `gorm:"type:uuid;not null" json:"patient_id"`
	RoutineTimeID string           `gorm:"type:uuid;not null" json:"routine_time_id"`
	LoggedAt      time.Time        `gorm:"not null;default:now()" json:"logged_at"`
	Status        RoutineLogStatus `gorm:"type:routine_log_status_enum;not null;default:Pending" json:"status"`

	// Relations
	RoutineTime *RoutineTime `gorm:"foreignKey:RoutineTimeID" json:"routine_time,omitempty"`
}

func (RoutineLogEntry) TableName() string { return "routine_log_entries" }
