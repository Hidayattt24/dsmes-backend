package domain

import "time"

type ReminderType string

const (
	ReminderSistem   ReminderType = "sistem"
	ReminderPersonal ReminderType = "personal"
)

type ReminderCategory string

const (
	CategoryMedisObat      ReminderCategory = "medis_obat"
	CategoryNutrisiAir     ReminderCategory = "nutrisi_air"
	CategoryAktivitasFisik ReminderCategory = "aktivitas_fisik"
	CategoryLainnya        ReminderCategory = "lainnya"
)

type ReminderLogStatus string

const (
	ReminderSelesai  ReminderLogStatus = "selesai"
	ReminderTerlewat ReminderLogStatus = "terlewat"
	ReminderPending  ReminderLogStatus = "pending"
)

// Reminder represents a patient's customized/system reminder.
type Reminder struct {
	BaseModel

	PatientID          string           `gorm:"type:uuid;not null" json:"patient_id"`
	ActivityName       string           `gorm:"type:varchar(150);not null" json:"activity_name"`
	ReminderType       ReminderType     `gorm:"type:reminder_type_enum;not null" json:"reminder_type"`
	Category           ReminderCategory `gorm:"type:reminder_category_enum;not null" json:"category"`
	ScheduledTime      string           `gorm:"type:time;not null" json:"scheduled_time"` // time string e.g. "08:00:00"
	IsActive           bool             `gorm:"not null;default:true" json:"is_active"`
	Notes              string           `gorm:"type:text" json:"notes"`
	IconName           string           `gorm:"type:varchar(100);default:'default'" json:"icon_name"`
	RepeatIntervalDays int              `gorm:"default:1" json:"repeat_interval_days"`

	// Relations
	ActiveDays []ReminderActiveDay `gorm:"foreignKey:ReminderID;constraint:OnDelete:CASCADE" json:"active_days,omitempty"`
}

func (Reminder) TableName() string { return "reminders" }

// ReminderActiveDay represents the days of the week a reminder is active.
type ReminderActiveDay struct {
	ReminderID string `gorm:"type:uuid;primaryKey" json:"reminder_id"`
	DayOfWeek  int    `gorm:"type:smallint;primaryKey" json:"day_of_week"` // 1=Senin, 7=Minggu
}

func (ReminderActiveDay) TableName() string { return "reminder_active_days" }

// DailyReminderLog tracks reminder execution state daily.
type DailyReminderLog struct {
	BaseModel

	ReminderID string            `gorm:"type:uuid;not null;uniqueIndex:idx_reminder_log_date" json:"reminder_id"`
	LogDate    time.Time         `gorm:"type:date;not null;uniqueIndex:idx_reminder_log_date" json:"log_date"`
	Status     ReminderLogStatus `gorm:"type:reminder_log_status_enum;not null;default:pending" json:"status"`
}

func (DailyReminderLog) TableName() string { return "daily_reminder_logs" }

// SystemReminderTemplate contains predefined reminders copied to patients during onboarding.
type SystemReminderTemplate struct {
	BaseModel

	ActivityName     string           `gorm:"type:varchar(150);not null" json:"activity_name"`
	Category         ReminderCategory `gorm:"type:reminder_category_enum;not null" json:"category"`
	DefaultTime      *string          `gorm:"type:time" json:"default_time"`
	DefaultFrequency string           `gorm:"type:varchar(50)" json:"default_frequency"`
	Description      string           `gorm:"type:text" json:"description"`
}

func (SystemReminderTemplate) TableName() string { return "system_reminder_templates" }

// NotificationLog keeps track of sent notifications.
type NotificationLog struct {
	BaseModel

	ReminderID  *string   `gorm:"type:uuid" json:"reminder_id"`
	PatientID   string    `gorm:"type:uuid;not null" json:"patient_id"`
	MessageText string    `gorm:"type:text;not null" json:"message_text"`
	NotifiedAt  time.Time `gorm:"not null;default:now()" json:"notified_at"`
	IsRead      bool      `gorm:"not null;default:false" json:"is_read"`
	NotifType   string    `gorm:"type:varchar(50);not null;default:reminder" json:"notif_type"`
	ArticleID   *string   `gorm:"type:uuid" json:"article_id"`
}

func (NotificationLog) TableName() string { return "notification_logs" }
