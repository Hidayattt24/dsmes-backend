package domain

import (
	"time"
)

type MeasurementTime string

const (
	TimeSebelumMakan MeasurementTime = "sebelum_makan"
	TimeSesudahMakan MeasurementTime = "sesudah_makan"
	TimeSewaktu      MeasurementTime = "sewaktu"
)

type GlucoseStatus string

const (
	GlucoseRendah      GlucoseStatus = "rendah"
	GlucoseNormal      GlucoseStatus = "normal"
	GlucoseTinggi      GlucoseStatus = "tinggi"
	GlucoseSangatTinggi GlucoseStatus = "sangat_tinggi"
)

// BloodSugarLog represents a patient's recorded blood sugar measurement.
type BloodSugarLog struct {
	BaseModel

	PatientID           string          `gorm:"type:uuid;not null" json:"patient_id"`
	GlucoseValue        int             `gorm:"not null" json:"glucose_value"` // mg/dL
	MeasurementTimeType MeasurementTime `gorm:"type:measurement_time_enum;not null" json:"measurement_time_type"`
	MeasuredAt          time.Time       `gorm:"not null" json:"measured_at"`
	Status              GlucoseStatus   `gorm:"type:glucose_status_enum;not null" json:"status"`
}

func (BloodSugarLog) TableName() string { return "blood_sugar_logs" }

// CalculateGlucoseStatus determines the status based on medical rules:
// - Sebelum makan: Normal < 130, Tinggi >= 130, Sangat Tinggi >= 180, Rendah < 70
// - Sesudah makan: Normal < 180, Tinggi >= 180, Sangat Tinggi >= 250, Rendah < 70
// - Sewaktu: Normal < 140, Tinggi >= 140, Sangat Tinggi >= 200, Rendah < 70
func CalculateGlucoseStatus(val int, mTime MeasurementTime) GlucoseStatus {
	if val < 70 {
		return GlucoseRendah
	}

	switch mTime {
	case TimeSebelumMakan:
		if val >= 180 {
			return GlucoseSangatTinggi
		} else if val >= 130 {
			return GlucoseTinggi
		}
		return GlucoseNormal
	case TimeSesudahMakan:
		if val >= 250 {
			return GlucoseSangatTinggi
		} else if val >= 180 {
			return GlucoseTinggi
		}
		return GlucoseNormal
	case TimeSewaktu:
		if val >= 200 {
			return GlucoseSangatTinggi
		} else if val >= 140 {
			return GlucoseTinggi
		}
		return GlucoseNormal
	default:
		if val >= 200 {
			return GlucoseSangatTinggi
		} else if val >= 140 {
			return GlucoseTinggi
		}
		return GlucoseNormal
	}
}
