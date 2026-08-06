package domain

import (
	"fmt"
	"time"
)

// ── Measurement Type ──────────────────────────────────────────────────────────

type MeasurementTime string

const (
	TimeFasting    MeasurementTime = "fasting"
	TimeBeforeMeal MeasurementTime = "before_meal"
	TimeAfterMeal  MeasurementTime = "after_meal"
	TimeBeforeBed  MeasurementTime = "before_bed"
	TimeRandom     MeasurementTime = "random"

	// Legacy aliases — kept for backward compatibility.
	TimeSebelumMakan MeasurementTime = "sebelum_makan"
	TimeSesudahMakan MeasurementTime = "sesudah_makan"
	TimeSewaktu      MeasurementTime = "sewaktu"
	TimePuasa        MeasurementTime = "puasa"
	TimeSebelumTidur MeasurementTime = "sebelum_tidur"
)

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

// ── Glucose Category (diagnostic) ─────────────────────────────────────────────
//
// The six clinical categories are the SINGLE source of truth for blood sugar
// classification across the entire system (Mobile, Staff, Admin, Reports, AI).
// No other code anywhere may classify a glucose value — all UI layers consume
// the category + label + colour emitted here.

type GlucoseCategory string

const (
	CategoryHypoglycemia  GlucoseCategory = "hypoglycemia"
	CategoryNormal        GlucoseCategory = "normal"
	CategoryTarget        GlucoseCategory = "target"
	CategoryPrediabetes   GlucoseCategory = "prediabetes"
	CategoryElevated      GlucoseCategory = "elevated"
	CategoryHyperglycemia GlucoseCategory = "hyperglycemia"
)

// CategoryInfo holds the display metadata for one category — returned by the
// classifier and serialised to JSON so all frontends can render identically.
type CategoryInfo struct {
	Category    GlucoseCategory `json:"category"`
	Label       string          `json:"label"`
	Color       string          `json:"color"`
	Severity    GlucoseSeverity `json:"severity"`
	Description string          `json:"description"`
}

// ── Severity ──────────────────────────────────────────────────────────────────

type GlucoseSeverity string

const (
	SeverityNormal  GlucoseSeverity = "normal"
	SeverityWarning GlucoseSeverity = "warning"
	SeverityDanger  GlucoseSeverity = "danger"
)

// ── Classifier Result ─────────────────────────────────────────────────────────

// BloodSugarClassification is the unified return type of the classifier. Every
// consumer (API handler, dashboard query, report) receives the SAME shape.
type BloodSugarClassification struct {
	Category       GlucoseCategory `json:"category"`
	CategoryLabel  string          `json:"category_label"`
	Severity       GlucoseSeverity `json:"severity"`
	Color          string          `json:"color"`
	Description    string          `json:"description"`
	ReferenceMin   int             `json:"reference_min"`
	ReferenceMax   int             `json:"reference_max"`
	ReferenceRange string          `json:"reference_range"`
	Recommendation string          `json:"recommendation"`
}

// ── Category metadata ─────────────────────────────────────────────────────────

var categoryInfo = map[GlucoseCategory]CategoryInfo{
	CategoryHypoglycemia: {
		Category:    CategoryHypoglycemia,
		Label:       "Hipoglikemia",
		Color:       "#DC2626",
		Severity:    SeverityDanger,
		Description: "Kadar gula darah di bawah rentang normal. Memerlukan tindakan segera.",
	},
	CategoryNormal: {
		Category:    CategoryNormal,
		Label:       "Normal",
		Color:       "#10B981",
		Severity:    SeverityNormal,
		Description: "Kadar gula darah berada dalam rentang normal.",
	},
	CategoryTarget: {
		Category:    CategoryTarget,
		Label:       "Target",
		Color:       "#10B981",
		Severity:    SeverityNormal,
		Description: "Kadar gula darah dalam rentang target pengelolaan diabetes.",
	},
	CategoryPrediabetes: {
		Category:    CategoryPrediabetes,
		Label:       "Prediabetes",
		Color:       "#F59E0B",
		Severity:    SeverityWarning,
		Description: "Kadar gula darah berada di rentang prediabetes. Risiko berkembang menjadi diabetes.",
	},
	CategoryElevated: {
		Category:    CategoryElevated,
		Label:       "Elevated",
		Color:       "#F97316",
		Severity:    SeverityWarning,
		Description: "Kadar gula darah di atas target pengelolaan.",
	},
	CategoryHyperglycemia: {
		Category:    CategoryHyperglycemia,
		Label:       "Hiperglikemia",
		Color:       "#DC2626",
		Severity:    SeverityDanger,
		Description: "Kadar gula darah di atas ambang diabetes. Perlu perhatian medis.",
	},
}

// ── BloodSugarLog ─────────────────────────────────────────────────────────────

type BloodSugarLog struct {
	BaseModel

	PatientID           string          `gorm:"type:uuid;not null;index" json:"patient_id"`
	GlucoseValue        int             `gorm:"not null" json:"glucose_value"`
	MeasurementTimeType MeasurementTime `gorm:"type:varchar(50);not null" json:"measurement_time_type"`
	MeasuredAt          time.Time       `gorm:"not null;index" json:"measured_at"`
	Category            GlucoseCategory `gorm:"type:varchar(50);not null;column:status" json:"category"`
	Severity            GlucoseSeverity `gorm:"type:varchar(50);not null;default:'normal'" json:"severity"`
	ReferenceMin        int             `gorm:"not null;default:70" json:"reference_min"`
	ReferenceMax        int             `gorm:"not null;default:140" json:"reference_max"`
	ReferenceRange      string          `gorm:"type:varchar(100)" json:"reference_range"`
	Recommendation      string          `gorm:"type:text" json:"recommendation"`
	Color               string          `gorm:"type:varchar(30)" json:"color"`
}

func (BloodSugarLog) TableName() string { return "blood_sugar_logs" }

// ── The Single Classifier ─────────────────────────────────────────────────────
//
// ClassifyBloodGlucose is the ONLY function in the entire system that determines
// a blood sugar reading's clinical category. All frontends (Mobile, Staff,
// Admin, Dashboard, Reports) MUST consume this output rather than computing
// their own classification.

func ClassifyBloodGlucose(val int, mType MeasurementTime, dob *time.Time) BloodSugarClassification {
	normType := NormalizeMeasurementType(string(mType))

	age := 30
	if dob != nil && !dob.IsZero() {
		now := time.Now()
		age = now.Year() - dob.Year()
		if now.YearDay() < dob.YearDay() {
			age--
		}
	}

	var refMin, refMax int
	var refRange string

	switch normType {
	case TimeFasting:
		refMin = 70
		refMax = 100
		refRange = fmt.Sprintf("%d – %d mg/dL (Puasa)", refMin, refMax)
		info := classifyGDP(val, refMin, refMax, refRange)
		adjustRangeForAge(&info, age, normType)
		return info

	case TimeBeforeMeal:
		refMin = 70
		refMax = 100
		refRange = fmt.Sprintf("%d – %d mg/dL (Sebelum Makan)", refMin, refMax)
		info := classifyBeforeMeal(val, refMin, refMax, refRange)
		adjustRangeForAge(&info, age, normType)
		return info

	case TimeAfterMeal:
		refMin = 70
		if age < 50 {
			refMax = 140
			refRange = "< 140 mg/dL (2 Jam Sesudah Makan)"
		} else if age <= 60 {
			refMax = 150
			refRange = "< 150 mg/dL (2 Jam Sesudah Makan)"
		} else {
			refMax = 160
			refRange = "< 160 mg/dL (2 Jam Sesudah Makan)"
		}
		return classifyGD2PP(val, refMin, refMax, refRange)

	case TimeBeforeBed:
		refMin = 70
		refMax = 140
		refRange = fmt.Sprintf("%d – %d mg/dL (Sebelum Tidur)", refMin, refMax)
		return classifyBeforeBed(val, refMin, refMax, refRange)

	case TimeRandom:
		refMin = 70
		refMax = 200
		refRange = "< 200 mg/dL (Sewaktu)"
		return classifyGDS(val, refMin, refMax, refRange)

	default:
		refMin = 70
		refMax = 140
		refRange = "< 140 mg/dL"
		return classifyGDS(val, refMin, refMax, refRange)
	}
}

// ── Per-Type Classification ───────────────────────────────────────────────────

// GDP (Gula Darah Puasa):
//
//	< 70             → Hypoglycemia
//	70 – 99          → Normal
//	100 – 125        → Prediabetes
//	≥ 126            → Hyperglycemia
func classifyGDP(val, refMin, refMax int, refRange string) BloodSugarClassification {
	if val < 70 {
		return hypoResult(refMin, refMax, refRange)
	}
	if val < 100 {
		return catResult(CategoryNormal, refMin, refMax, refRange)
	}
	if val < 126 {
		return catResult(CategoryPrediabetes, refMin, refMax, refRange)
	}
	return catResult(CategoryHyperglycemia, refMin, refMax, refRange)
}

// GD2PP (2 Jam Setelah Makan):
//
//	< 70             → Hypoglycemia
//	70 – 139         → Normal
//	140 – 199        → Prediabetes
//	≥ 200            → Hyperglycemia
func classifyGD2PP(val, refMin, refMax int, refRange string) BloodSugarClassification {
	if val < 70 {
		return hypoResult(refMin, refMax, refRange)
	}
	if val < 140 {
		return catResult(CategoryNormal, refMin, refMax, refRange)
	}
	if val < 200 {
		return catResult(CategoryPrediabetes, refMin, refMax, refRange)
	}
	return catResult(CategoryHyperglycemia, refMin, refMax, refRange)
}

// GDS (Sewaktu):
//
//	< 70             → Hypoglycemia
//	70 – 199         → Normal
//	≥ 200            → Hyperglycemia
func classifyGDS(val, refMin, refMax int, refRange string) BloodSugarClassification {
	if val < 70 {
		return hypoResult(refMin, refMax, refRange)
	}
	if val < 200 {
		return catResult(CategoryNormal, refMin, refMax, refRange)
	}
	return catResult(CategoryHyperglycemia, refMin, refMax, refRange)
}

// before_meal (Sebelum Makan — management class):
//
//	< 70             → Hypoglycemia
//	70 – 99          → Target
//	100 – 199        → Elevated
//	≥ 200            → Hyperglycemia
func classifyBeforeMeal(val, refMin, refMax int, refRange string) BloodSugarClassification {
	if val < 70 {
		return hypoResult(refMin, refMax, refRange)
	}
	if val < 100 {
		return catResult(CategoryTarget, refMin, refMax, refRange)
	}
	if val < 200 {
		return catResult(CategoryElevated, refMin, refMax, refRange)
	}
	return catResult(CategoryHyperglycemia, refMin, refMax, refRange)
}

// before_bed (Sebelum Tidur — management class):
//
//	< 70             → Hypoglycemia
//	70 – 139         → Target
//	140 – 199        → Elevated
//	≥ 200            → Hyperglycemia
func classifyBeforeBed(val, refMin, refMax int, refRange string) BloodSugarClassification {
	if val < 70 {
		return hypoResult(refMin, refMax, refRange)
	}
	if val < 140 {
		return catResult(CategoryTarget, refMin, refMax, refRange)
	}
	if val < 200 {
		return catResult(CategoryElevated, refMin, refMax, refRange)
	}
	return catResult(CategoryHyperglycemia, refMin, refMax, refRange)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func hypoResult(refMin, refMax int, refRange string) BloodSugarClassification {
	cat := categoryInfo[CategoryHypoglycemia]
	return BloodSugarClassification{
		Category:       cat.Category,
		CategoryLabel:  cat.Label,
		Severity:       SeverityDanger,
		Color:          cat.Color,
		Description:    cat.Description,
		ReferenceMin:   refMin,
		ReferenceMax:   refMax,
		ReferenceRange: refRange,
		Recommendation: "Konsumsi karbohidrat cepat serap (jus buah, teh manis, atau permen) dan periksa kembali gula darah Anda dalam 15 menit.",
	}
}

func catResult(cat GlucoseCategory, refMin, refMax int, refRange string) BloodSugarClassification {
	info := categoryInfo[cat]
	rec := info.Description + " Pertahankan pola hidup sehat dan pemantauan rutin."
	switch cat {
	case CategoryPrediabetes:
		rec = "Kadar gula darah menunjukkan risiko prediabetes. Konsultasikan dengan tenaga kesehatan untuk evaluasi lebih lanjut."
	case CategoryElevated:
		rec = "Kadar gula darah di atas target pengelolaan. Tetap utamakan konsumsi makanan bergizi seimbang dan aktivitas fisik teratur."
	case CategoryHyperglycemia:
		rec = "Kadar gula darah tinggi. Pastikan rutin minum obat sesuai anjuran. Konsultasikan dengan tenaga kesehatan jika kadar gula darah tinggi berlanjut."
	}
	return BloodSugarClassification{
		Category:       cat,
		CategoryLabel:  info.Label,
		Severity:       info.Severity,
		Color:          info.Color,
		Description:    info.Description,
		ReferenceMin:   refMin,
		ReferenceMax:   refMax,
		ReferenceRange: refRange,
		Recommendation: rec,
	}
}

func adjustRangeForAge(info *BloodSugarClassification, age int, _ MeasurementTime) {
	// Age adjustments are handled per measurement type in the caller.
	_ = age
}

// ── Backward Compatibility ────────────────────────────────────────────────────
//
// CalculateBloodSugarMedicalResult and CalculateGlucoseStatus remain as thin
// wrappers around the new classifier so existing call-sites still compile.

func CalculateBloodSugarMedicalResult(val int, mType MeasurementTime, dob *time.Time) BloodSugarClassification {
	return ClassifyBloodGlucose(val, mType, dob)
}

func CalculateGlucoseStatus(val int, mType MeasurementTime) GlucoseCategory {
	return ClassifyBloodGlucose(val, mType, nil).Category
}
