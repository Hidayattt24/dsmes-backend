package domain

import (
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
	CategoryLowWarning    GlucoseCategory = "low_warning"
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
		Description: "Kadar gula darah terlalu rendah. Memerlukan tindakan segera.",
	},
	CategoryLowWarning: {
		Category:    CategoryLowWarning,
		Label:       "Waspada Rendah",
		Color:       "#F59E0B",
		Severity:    SeverityWarning,
		Description: "Kadar gula darah mendekati batas bawah normal. Waspadai gejala hipoglikemia.",
	},
	CategoryNormal: {
		Category:    CategoryNormal,
		Label:       "Normal",
		Color:       "#10B981",
		Severity:    SeverityNormal,
		Description: "Kadar gula darah berada dalam target ideal / terkontrol.",
	},
	CategoryTarget: {
		Category:    CategoryTarget,
		Label:       "Normal",
		Color:       "#10B981",
		Severity:    SeverityNormal,
		Description: "Kadar gula darah dalam rentang target aman pengelolaan diabetes.",
	},
	CategoryPrediabetes: {
		Category:    CategoryPrediabetes,
		Label:       "Waspada",
		Color:       "#F59E0B",
		Severity:    SeverityWarning,
		Description: "Kadar gula darah berada di atas target ideal (toleransi glukosa terganggu / prediabetes).",
	},
	CategoryElevated: {
		Category:    CategoryElevated,
		Label:       "Waspada",
		Color:       "#F59E0B",
		Severity:    SeverityWarning,
		Description: "Kadar gula darah di atas target pengelolaan / elevated.",
	},
	CategoryHyperglycemia: {
		Category:    CategoryHyperglycemia,
		Label:       "Hiperglikemia",
		Color:       "#DC2626",
		Severity:    SeverityDanger,
		Description: "Kadar gula darah tinggi di atas target aman. Perlu perhatian medis.",
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

	var refMin, refMax int
	var refRange string

	switch normType {
	case TimeFasting:
		refMin = 80
		refMax = 130
		refRange = "80 – 130 mg/dL (Puasa)"
		return classifyGDP(val, refMin, refMax, refRange)

	case TimeBeforeMeal:
		refMin = 80
		refMax = 130
		refRange = "80 – 130 mg/dL (Sebelum Makan)"
		return classifyBeforeMeal(val, refMin, refMax, refRange)

	case TimeAfterMeal:
		refMin = 70
		refMax = 179
		refRange = "< 180 mg/dL (2 Jam Sesudah Makan)"
		return classifyGD2PP(val, refMin, refMax, refRange)

	case TimeBeforeBed:
		refMin = 100
		refMax = 140
		refRange = "100 – 140 mg/dL (Sebelum Tidur)"
		return classifyBeforeBed(val, refMin, refMax, refRange)

	case TimeRandom:
		refMin = 70
		refMax = 139
		refRange = "< 140 mg/dL (Sewaktu)"
		return classifyGDS(val, refMin, refMax, refRange)

	default:
		refMin = 70
		refMax = 139
		refRange = "< 140 mg/dL (Sewaktu)"
		return classifyGDS(val, refMin, refMax, refRange)
	}
}

// ── Per-Type Classification ───────────────────────────────────────────────────

// GDP (Gula Darah Puasa):
//
//	< 70             → Hipoglikemia (Kritis)
//	70 – 79          → Waspada Rendah
//	80 – 130         → Normal / Terkontrol
//	131 – 180        → Waspada / Elevated
//	> 180            → Hiperglikemia
func classifyGDP(val, refMin, refMax int, refRange string) BloodSugarClassification {
	if val < 70 {
		return hypoResult(refMin, refMax, refRange)
	}
	if val <= 79 {
		return catResult(CategoryLowWarning, refMin, refMax, refRange)
	}
	if val <= 130 {
		return catResult(CategoryNormal, refMin, refMax, refRange)
	}
	if val <= 180 {
		return catResult(CategoryElevated, refMin, refMax, refRange)
	}
	return catResult(CategoryHyperglycemia, refMin, refMax, refRange)
}

// GD2PP (2 Jam Setelah Makan):
//
//	< 70             → Hipoglikemia
//	70 – 179         → Normal / Terkontrol
//	≥ 180            → Hiperglikemia
func classifyGD2PP(val, refMin, refMax int, refRange string) BloodSugarClassification {
	if val < 70 {
		return hypoResult(refMin, refMax, refRange)
	}
	if val < 180 {
		return catResult(CategoryNormal, refMin, refMax, refRange)
	}
	return catResult(CategoryHyperglycemia, refMin, refMax, refRange)
}

// GDS (Sewaktu):
//
//	< 70             → Hipoglikemia
//	70 – 139         → Normal / Terkontrol
//	140 – 199        → Waspada / TGT
//	≥ 200            → Hiperglikemia
func classifyGDS(val, refMin, refMax int, refRange string) BloodSugarClassification {
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

// before_meal (Sebelum Makan):
//
//	< 70             → Hipoglikemia
//	70 – 79          → Waspada Rendah
//	80 – 130         → Normal / Terkontrol
//	> 130            → Waspada / Elevated
func classifyBeforeMeal(val, refMin, refMax int, refRange string) BloodSugarClassification {
	if val < 70 {
		return hypoResult(refMin, refMax, refRange)
	}
	if val <= 79 {
		return catResult(CategoryLowWarning, refMin, refMax, refRange)
	}
	if val <= 130 {
		return catResult(CategoryNormal, refMin, refMax, refRange)
	}
	if val <= 180 {
		return catResult(CategoryElevated, refMin, refMax, refRange)
	}
	return catResult(CategoryHyperglycemia, refMin, refMax, refRange)
}

// before_bed (Sebelum Tidur):
//
//	< 100            → Hipoglikemia (Risiko Hipoglikemia Malam)
//	100 – 140        → Normal / Terkontrol
//	> 140            → Waspada / Elevated
func classifyBeforeBed(val, refMin, refMax int, refRange string) BloodSugarClassification {
	if val < 100 {
		return hypoResult(refMin, refMax, refRange)
	}
	if val <= 140 {
		return catResult(CategoryNormal, refMin, refMax, refRange)
	}
	if val <= 180 {
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
