package blood_sugar

import (
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type LogBloodSugarRequest struct {
	GlucoseValue        int                    `json:"glucose_value"           validate:"required,gt=0"`
	MeasurementTimeType domain.MeasurementTime `json:"measurement_time_type"   validate:"required,oneof=fasting before_meal after_meal before_bed random sebelum_makan sesudah_makan sewaktu puasa sebelum_tidur"`
	MeasuredAt          string                 `json:"measured_at"             validate:"required"`
}

type BloodSugarResponse struct {
	ID                   string                 `json:"id"`
	PatientID            string                 `json:"patient_id"`
	GlucoseValue         int                    `json:"glucose_value"`
	MeasurementTimeType  domain.MeasurementTime `json:"measurement_time_type"`
	MeasurementTimeLabel string                 `json:"measurement_time_label"`
	MeasuredAt           string                 `json:"measured_at"`
	Category             domain.GlucoseCategory `json:"category"`
	CategoryLabel        string                 `json:"category_label"`
	Severity             domain.GlucoseSeverity `json:"severity"`
	ReferenceMin         int                    `json:"reference_min"`
	ReferenceMax         int                    `json:"reference_max"`
	ReferenceRange       string                 `json:"reference_range"`
	Recommendation       string                 `json:"recommendation"`
	Color                string                 `json:"color"`
	CreatedAt            string                 `json:"created_at"`
	UpdatedAt            string                 `json:"updated_at"`
}

type MeasurementTypeStats struct {
	MeasurementType string  `json:"measurement_type"`
	TypeLabel       string  `json:"type_label"`
	AverageValue    float64 `json:"average_value"`
	Count           int64   `json:"count"`
}

type GlucoseDistributionResponse struct {
	HypoglycemiaCount  int64                  `json:"hypoglycemia_count"`
	NormalCount        int64                  `json:"normal_count"`
	TargetCount        int64                  `json:"target_count"`
	PrediabetesCount   int64                  `json:"prediabetes_count"`
	ElevatedCount      int64                  `json:"elevated_count"`
	HyperglycemiaCount int64                  `json:"hyperglycemia_count"`
	ByMeasurementType  []MeasurementTypeStats `json:"by_measurement_type,omitempty"`
}

func ToBloodSugarResponse(l *domain.BloodSugarLog) BloodSugarResponse {
	normType := domain.NormalizeMeasurementType(string(l.MeasurementTimeType))
	typeLabel := domain.GetMeasurementTypeLabel(l.MeasurementTimeType)

	// Recompute via the single classifier for the label (the stored Category
	// enum value is correct, but the human-readable label is derived here for
	// consistency with the domain metadata).
	medRes := domain.ClassifyBloodGlucose(l.GlucoseValue, normType, nil)

	category := l.Category
	classLabel := medRes.CategoryLabel

	severity := l.Severity
	if severity == "" {
		severity = medRes.Severity
	}

	refMin := l.ReferenceMin
	if refMin <= 0 {
		refMin = medRes.ReferenceMin
	}

	refMax := l.ReferenceMax
	if refMax <= 0 {
		refMax = medRes.ReferenceMax
	}

	refRange := l.ReferenceRange
	if refRange == "" {
		refRange = medRes.ReferenceRange
	}

	recommendation := l.Recommendation
	if recommendation == "" {
		recommendation = medRes.Recommendation
	}

	color := l.Color
	if color == "" {
		color = medRes.Color
	}

	return BloodSugarResponse{
		ID:                   l.ID,
		PatientID:            l.PatientID,
		GlucoseValue:         l.GlucoseValue,
		MeasurementTimeType:  normType,
		MeasurementTimeLabel: typeLabel,
		MeasuredAt:           l.MeasuredAt.Format(time.RFC3339),
		Category:             category,
		CategoryLabel:        classLabel,
		Severity:             severity,
		ReferenceMin:         refMin,
		ReferenceMax:         refMax,
		ReferenceRange:       refRange,
		Recommendation:       recommendation,
		Color:                color,
		CreatedAt:            l.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            l.UpdatedAt.Format(time.RFC3339),
	}
}
