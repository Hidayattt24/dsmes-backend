package domain

import "time"

// DailyMedicalCheckin represents check-in calendars for patients.
type DailyMedicalCheckin struct {
	BaseModel

	PatientID   string    `gorm:"type:uuid;not null;uniqueIndex:idx_patient_checkin" json:"patient_id"`
	CheckinDate time.Time `gorm:"type:date;not null;uniqueIndex:idx_patient_checkin" json:"checkin_date"`
	IsCompleted bool      `gorm:"not null;default:false" json:"is_completed"`
}

func (DailyMedicalCheckin) TableName() string { return "daily_medical_checkins" }
