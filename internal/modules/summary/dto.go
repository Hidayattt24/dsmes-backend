package summary

import (
	"github.com/dsmes/dsmes-backend/internal/domain"
)

type WeeklySummaryResponse struct {
	ID                  string             `json:"id"`
	WeekStartDate       string             `json:"week_start_date"`
	WeekEndDate         string             `json:"week_end_date"`
	BloodSugarTrend     domain.TrendStatus `json:"blood_sugar_trend"`
	AvgBloodSugar       float64            `json:"avg_blood_sugar"`
	ArticlesReadCount   int                `json:"articles_read_count"`
	ArticlesTargetCount int                `json:"articles_target_count"`
	GeneratedAt         string             `json:"generated_at"`
}

type GenerateSummaryRequest struct {
	PatientID     string `json:"patient_id" validate:"required,uuid4"`
	WeekStartDate string `json:"week_start_date" validate:"required"` // format: "YYYY-MM-DD"
}

func ToWeeklySummaryResponse(w *domain.WeeklyHealthSummary) WeeklySummaryResponse {
	return WeeklySummaryResponse{
		ID:                  w.ID,
		WeekStartDate:       w.WeekStartDate.Format("2006-01-02"),
		WeekEndDate:         w.WeekEndDate.Format("2006-01-02"),
		BloodSugarTrend:     w.BloodSugarTrend,
		AvgBloodSugar:       w.AvgBloodSugar,
		ArticlesReadCount:   w.ArticlesReadCount,
		ArticlesTargetCount: w.ArticlesTargetCount,
		GeneratedAt:         w.GeneratedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
