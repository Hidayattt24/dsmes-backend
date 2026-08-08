package domain

import "time"

// PatientActivityLog represents a direct physical activity log from the mobile app.
type PatientActivityLog struct {
	BaseModel

	PatientID       string    `gorm:"type:uuid;not null" json:"patient_id"`
	ActivityName    string    `gorm:"type:varchar(255);not null;default:''" json:"activity_name"`
	DurationMinutes int       `gorm:"not null;default:0" json:"duration_minutes"`
	Intensity       string    `gorm:"type:varchar(50);not null;default:'Ringan'" json:"intensity"`
	Notes           string    `gorm:"type:text;not null;default:''" json:"notes"`
	LoggedAt        time.Time `gorm:"not null;default:now()" json:"logged_at"`
}

func (PatientActivityLog) TableName() string { return "patient_activity_logs" }
