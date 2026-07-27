package domain

import (
	"fmt"
	"time"
)

type MeasurementTime string

const (
	TimeFasting     MeasurementTime = "fasting"     // Puasa (GDP)
	TimeBeforeMeal  MeasurementTime = "before_meal"  // Sebelum Makan
	TimeAfterMeal   MeasurementTime = "after_meal"   // 2 Jam Sesudah Makan
	TimeBeforeBed   MeasurementTime = "before_bed"   // Sebelum Tidur
	TimeRandom      MeasurementTime = "random"      // Sewaktu (GDS)

	// Legacy aliases for backward compatibility
	TimeSebelumMakan MeasurementTime = "sebelum_makan"
	TimeSesudahMakan MeasurementTime = "sesudah_makan"
	TimeSewaktu      MeasurementTime = "sewaktu"
	TimePuasa        MeasurementTime = "puasa"
	TimeSebelumTidur MeasurementTime = "sebelum_tidur"
)

// NormalizeMeasurementType converts aliases to standardized keys.
func NormalizeMeasurementType(raw string) MeasurementTime {
	switch raw {
	case "fasting", "puasa", "GDP":
		return TimeFasting
	case "before_meal", "sebelum_makan":
		return TimeBeforeMeal
	case "after_meal", "sesudah_makan", "2_jam_sesudah_makan":
		return TimeAfterMeal
	case "before_bed", "sebelum_tidur":
		return TimeBeforeBed
	case "random", "sewaktu", "GDS":
		return TimeRandom
	default:
		return TimeRandom
	}
}

// GetMeasurementTypeLabel returns a human-readable Indonesian label.
func GetMeasurementTypeLabel(mType MeasurementTime) string {
	switch NormalizeMeasurementType(string(mType)) {
	case TimeFasting:
		return "Puasa"
	case TimeBeforeMeal:
		return "Sebelum Makan"
	case TimeAfterMeal:
		return "2 Jam Sesudah Makan"
	case TimeBeforeBed:
		return "Sebelum Tidur"
	case TimeRandom:
		return "Sewaktu"
	default:
		return "Sewaktu"
	}
}

type GlucoseStatus string

const (
	GlucoseSevereHypo  GlucoseStatus = "severe_hypoglycemia"
	GlucoseHypo        GlucoseStatus = "hypoglycemia"
	GlucoseNormal      GlucoseStatus = "normal"
	GlucoseHyper       GlucoseStatus = "hyperglycemia"
	GlucoseSevereHyper GlucoseStatus = "severe_hyperglycemia"
)

type GlucoseSeverity string

const (
	SeverityNormal  GlucoseSeverity = "normal"
	SeverityWarning GlucoseSeverity = "warning"
	SeverityDanger  GlucoseSeverity = "danger"
)

type MedicalResult struct {
	Classification      GlucoseStatus   `json:"classification"`
	ClassificationLabel string          `json:"classification_label"`
	Severity            GlucoseSeverity `json:"severity"`
	ReferenceMin        int             `json:"reference_min"`
	ReferenceMax        int             `json:"reference_max"`
	ReferenceRangeText  string          `json:"reference_range_text"`
	Recommendation      string          `json:"recommendation"`
	ColorIndicator      string          `json:"color_indicator"`
}

// BloodSugarLog represents a patient's recorded blood sugar measurement.
type BloodSugarLog struct {
	BaseModel

	PatientID           string          `gorm:"type:uuid;not null;index" json:"patient_id"`
	GlucoseValue        int             `gorm:"not null" json:"glucose_value"` // mg/dL
	MeasurementTimeType MeasurementTime `gorm:"type:varchar(50);not null" json:"measurement_time_type"`
	MeasuredAt          time.Time       `gorm:"not null;index" json:"measured_at"`
	Status              GlucoseStatus   `gorm:"type:varchar(50);not null" json:"status"`
	Severity            GlucoseSeverity `gorm:"type:varchar(50);not null;default:'normal'" json:"severity"`
	ReferenceMin        int             `gorm:"not null;default:70" json:"reference_min"`
	ReferenceMax        int             `gorm:"not null;default:140" json:"reference_max"`
	ReferenceRangeText  string          `gorm:"type:varchar(100)" json:"reference_range_text"`
	Recommendation      string          `gorm:"type:text" json:"recommendation"`
	ColorIndicator      string          `gorm:"type:varchar(30)" json:"color_indicator"`
}

func (BloodSugarLog) TableName() string { return "blood_sugar_logs" }

// CalculateBloodSugarMedicalResult computes medical classification according to international guidelines and patient age.
func CalculateBloodSugarMedicalResult(val int, mType MeasurementTime, dob *time.Time) MedicalResult {
	normType := NormalizeMeasurementType(string(mType))

	// Calculate age in years if DateOfBirth is available
	age := 30 // default if DOB not specified
	if dob != nil && !dob.IsZero() {
		now := time.Now()
		age = now.Year() - dob.Year()
		if now.YearDay() < dob.YearDay() {
			age--
		}
	}

	var refMin, refMax int
	var rangeText string

	switch normType {
	case TimeFasting:
		refMin = 70
		refMax = 100
		rangeText = "70 – 100 mg/dL"
	case TimeBeforeMeal:
		refMin = 80
		refMax = 120
		rangeText = "80 – 120 mg/dL"
	case TimeBeforeBed:
		refMin = 100
		refMax = 140
		rangeText = "100 – 140 mg/dL"
	case TimeAfterMeal:
		refMin = 70
		if age < 50 {
			refMax = 140
			rangeText = "< 140 mg/dL (Usia < 50 thn)"
		} else if age <= 60 {
			refMax = 150
			rangeText = "< 150 mg/dL (Usia 50–60 thn)"
		} else {
			refMax = 160
			rangeText = "< 160 mg/dL (Usia > 60 thn)"
		}
	case TimeRandom:
		refMin = 70
		refMax = 140
		rangeText = "< 140 mg/dL"
	default:
		refMin = 70
		refMax = 140
		rangeText = "< 140 mg/dL"
	}

	// Medical Classification & Severity Rules
	if val < 40 {
		return MedicalResult{
			Classification:      GlucoseSevereHypo,
			ClassificationLabel: "Hipoglikemia Berat",
			Severity:            SeverityDanger,
			ReferenceMin:        refMin,
			ReferenceMax:        refMax,
			ReferenceRangeText:  rangeText,
			Recommendation:      "DARURAT! Segera konsumsi gula/karbohidrat cepat serap dan minta bantuan orang terdekat atau medis!",
			ColorIndicator:      "#DC2626", // Red
		}
	}

	if val < 70 {
		return MedicalResult{
			Classification:      GlucoseHypo,
			ClassificationLabel: "Hipoglikemia",
			Severity:            SeverityWarning,
			ReferenceMin:        refMin,
			ReferenceMax:        refMax,
			ReferenceRangeText:  rangeText,
			Recommendation:      "Segera konsumsi karbohidrat cepat serap (seperti jus buah, teh manis, atau permen) dan periksa kembali gula darah Anda.",
			ColorIndicator:      "#F59E0B", // Amber
		}
	}

	if val >= 350 {
		return MedicalResult{
			Classification:      GlucoseSevereHyper,
			ClassificationLabel: "Hiperglikemia Berat",
			Severity:            SeverityDanger,
			ReferenceMin:        refMin,
			ReferenceMax:        refMax,
			ReferenceRangeText:  rangeText,
			Recommendation:      "SEGERA hubungi atau kunjungi fasilitas kesehatan/dokter! Kadar gula darah Anda sangat tinggi.",
			ColorIndicator:      "#DC2626", // Red
		}
	}

	if val > 200 {
		return MedicalResult{
			Classification:      GlucoseHyper,
			ClassificationLabel: "Hiperglikemia",
			Severity:            SeverityWarning,
			ReferenceMin:        refMin,
			ReferenceMax:        refMax,
			ReferenceRangeText:  rangeText,
			Recommendation:      "Konsultasikan dengan tenaga kesehatan jika kadar gula darah tinggi berlanjut.",
			ColorIndicator:      "#F97316", // Orange
		}
	}

	if val > refMax {
		return MedicalResult{
			Classification:      GlucoseHyper,
			ClassificationLabel: "Tinggi",
			Severity:            SeverityWarning,
			ReferenceMin:        refMin,
			ReferenceMax:        refMax,
			ReferenceRangeText:  rangeText,
			Recommendation:      fmt.Sprintf("Kadar gula darah di atas rentang normal (%s) untuk jenis pengukuran ini.", rangeText),
			ColorIndicator:      "#F59E0B", // Amber
		}
	}

	// Normal
	return MedicalResult{
		Classification:      GlucoseNormal,
		ClassificationLabel: "Normal",
		Severity:            SeverityNormal,
		ReferenceMin:        refMin,
		ReferenceMax:        refMax,
		ReferenceRangeText:  rangeText,
		Recommendation:      "Pertahankan pola hidup sehat dan pemantauan rutin Anda.",
		ColorIndicator:      "#10B981", // Green
	}
}

// CalculateGlucoseStatus returns classification for backward compatibility.
func CalculateGlucoseStatus(val int, mTime MeasurementTime) GlucoseStatus {
	res := CalculateBloodSugarMedicalResult(val, mTime, nil)
	return res.Classification
}
