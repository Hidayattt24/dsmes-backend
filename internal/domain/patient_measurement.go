package domain

import (
	"time"
)

// PatientMeasurement represents a historical health measurement record for a patient.
type PatientMeasurement struct {
	BaseModel

	PatientID              string     `gorm:"type:uuid;index;not null" json:"patient_id"`
	WeightKg               *float64   `gorm:"type:numeric(5,2)" json:"weight_kg,omitempty"`
	HeightCm               *float64   `gorm:"type:numeric(5,2)" json:"height_cm,omitempty"`
	BMI                    *float64   `gorm:"type:numeric(4,1)" json:"bmi,omitempty"`
	BloodPressureSystolic  *int       `gorm:"type:int" json:"blood_pressure_systolic,omitempty"`
	BloodPressureDiastolic *int       `gorm:"type:int" json:"blood_pressure_diastolic,omitempty"`
	BloodSugar             *int       `gorm:"type:int" json:"blood_sugar,omitempty"`
	WaistCircumferenceCm   *float64   `gorm:"type:numeric(5,2)" json:"waist_circumference_cm,omitempty"`
	DailyCalorieTarget     *int       `gorm:"type:int" json:"daily_calorie_target,omitempty"`
	Notes                  string     `gorm:"type:text" json:"notes,omitempty"`
	RecordedByID           *string    `gorm:"type:uuid" json:"recorded_by_id,omitempty"`
	RecordedByName         string     `gorm:"type:varchar(150);not null" json:"recorded_by_name"`
	RecordedByRole         string     `gorm:"type:varchar(50);not null;default:'admin'" json:"recorded_by_role"` // admin | patient
	MeasuredAt             time.Time  `gorm:"type:timestamptz;not null;index" json:"measured_at"`

	// Relation
	Patient *Patient `gorm:"foreignKey:PatientID" json:"-"`
}
