package summary

import (
	"context"
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type SummaryRepository interface {
	FindLatest(ctx context.Context, patientID string) (*domain.WeeklyHealthSummary, error)
	Upsert(ctx context.Context, s *domain.WeeklyHealthSummary) error

	// Aggregation helpers inside repo for calculating values
	CalculateWeeklyAverage(ctx context.Context, patientID string, start, end time.Time) (float64, error)
	CalculateCompletionsCount(ctx context.Context, patientID string, start, end time.Time) (int, error)
}

type SummaryService interface {
	GetLatestSummary(ctx context.Context, patientID string) (*WeeklySummaryResponse, error)
	GenerateWeeklySummary(ctx context.Context, patientID string, weekStart string) (*WeeklySummaryResponse, error)
}
