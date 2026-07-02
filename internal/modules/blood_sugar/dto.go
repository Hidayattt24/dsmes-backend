package blood_sugar

import (
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type LogBloodSugarRequest struct {
	GlucoseValue        int                    `json:"glucose_value"           validate:"required,gt=0"`
	MeasurementTimeType domain.MeasurementTime `json:"measurement_time_type"   validate:"required,oneof=sebelum_makan sesudah_makan sewaktu"`
	MeasuredAt          string                 `json:"measured_at"             validate:"required"` // format: "2006-01-02T15:04:05Z07:00"
}

type BloodSugarResponse struct {
	ID                  string                 `json:"id"`
	GlucoseValue        int                    `json:"glucose_value"`
	MeasurementTimeType domain.MeasurementTime `json:"measurement_time_type"`
	MeasuredAt          string                 `json:"measured_at"`
	Status              domain.GlucoseStatus   `json:"status"`
}

type GlucoseDistributionResponse struct {
	NormalCount      int64 `json:"normal_count"`
	TinggiCount      int64 `json:"tinggi_count"`
	SangatTinggiCent int64 `json:"sangat_tinggi_count"`
	RendahCount      int64 `json:"rendah_count"`
}

func ToBloodSugarResponse(l *domain.BloodSugarLog) BloodSugarResponse {
	return BloodSugarResponse{
		ID:                  l.ID,
		GlucoseValue:        l.GlucoseValue,
		MeasurementTimeType: l.MeasurementTimeType,
		MeasuredAt:          l.MeasuredAt.Format(time.RFC3339),
		Status:              l.Status,
	}
}
