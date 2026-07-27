package blood_sugar

import (
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type LogBloodSugarRequest struct {
	GlucoseValue        int                    `json:"glucose_value"           validate:"required,gt=0"`
	MeasurementTimeType domain.MeasurementTime `json:"measurement_time_type"   validate:"required,oneof=fasting before_meal after_meal before_bed random sebelum_makan sesudah_makan sewaktu puasa sebelum_tidur"`
	MeasuredAt          string                 `json:"measured_at"             validate:"required"` // format: "2006-01-02T15:04:05Z07:00"
}

type BloodSugarResponse struct {
	ID                  string                 `json:"id"`
	PatientID           string                 `json:"patient_id"`
	GlucoseValue        int                    `json:"glucose_value"`
	MeasurementTimeType domain.MeasurementTime `json:"measurement_time_type"`
	MeasurementTimeLabel string                `json:"measurement_time_label"`
	MeasuredAt          string                 `json:"measured_at"`
	Status              domain.GlucoseStatus   `json:"status"`
	ClassificationLabel string                 `json:"classification_label"`
	Severity            domain.GlucoseSeverity `json:"severity"`
	ReferenceMin        int                    `json:"reference_min"`
	ReferenceMax        int                    `json:"reference_max"`
	ReferenceRangeText  string                 `json:"reference_range_text"`
	Recommendation      string                 `json:"recommendation"`
	ColorIndicator      string                 `json:"color_indicator"`
	CreatedAt           string                 `json:"created_at"`
	UpdatedAt           string                 `json:"updated_at"`
}

type MeasurementTypeStats struct {
	MeasurementType string  `json:"measurement_type"`
	TypeLabel       string  `json:"type_label"`
	AverageValue    float64 `json:"average_value"`
	Count           int64   `json:"count"`
}

type GlucoseDistributionResponse struct {
	NormalCount        int64                  `json:"normal_count"`
	TinggiCount        int64                  `json:"tinggi_count"`
	SangatTinggiCent   int64                  `json:"sangat_tinggi_count"`
	RendahCount        int64                  `json:"rendah_count"`
	ByMeasurementType  []MeasurementTypeStats `json:"by_measurement_type,omitempty"`
}

func ToBloodSugarResponse(l *domain.BloodSugarLog) BloodSugarResponse {
	normType := domain.NormalizeMeasurementType(string(l.MeasurementTimeType))
	typeLabel := domain.GetMeasurementTypeLabel(l.MeasurementTimeType)

	// Ensure medical calculated fields are present
	medRes := domain.CalculateBloodSugarMedicalResult(l.GlucoseValue, normType, nil)
	classLabel := medRes.ClassificationLabel
	if l.Status == domain.GlucoseNormal {
		classLabel = "Normal"
	} else if l.Status == domain.GlucoseHypo {
		classLabel = "Hipoglikemia"
	} else if l.Status == domain.GlucoseSevereHypo {
		classLabel = "Hipoglikemia Berat"
	} else if l.Status == domain.GlucoseHyper {
		classLabel = "Hiperglikemia"
	} else if l.Status == domain.GlucoseSevereHyper {
		classLabel = "Hiperglikemia Berat"
	}

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

	rangeText := l.ReferenceRangeText
	if rangeText == "" {
		rangeText = medRes.ReferenceRangeText
	}

	recommendation := l.Recommendation
	if recommendation == "" {
		recommendation = medRes.Recommendation
	}

	color := l.ColorIndicator
	if color == "" {
		color = medRes.ColorIndicator
	}

	return BloodSugarResponse{
		ID:                   l.ID,
		PatientID:            l.PatientID,
		GlucoseValue:         l.GlucoseValue,
		MeasurementTimeType:  normType,
		MeasurementTimeLabel: typeLabel,
		MeasuredAt:           l.MeasuredAt.Format(time.RFC3339),
		Status:               l.Status,
		ClassificationLabel:  classLabel,
		Severity:             severity,
		ReferenceMin:         refMin,
		ReferenceMax:         refMax,
		ReferenceRangeText:   rangeText,
		Recommendation:       recommendation,
		ColorIndicator:       color,
		CreatedAt:            l.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            l.UpdatedAt.Format(time.RFC3339),
	}
}
