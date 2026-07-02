package summary

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type summaryService struct {
	repo SummaryRepository
	log  *zap.Logger
}

func NewSummaryService(repo SummaryRepository, log *zap.Logger) SummaryService {
	return &summaryService{repo: repo, log: log}
}

func (s *summaryService) GetLatestSummary(ctx context.Context, patientID string) (*WeeklySummaryResponse, error) {
	item, err := s.repo.FindLatest(ctx, patientID)
	if err != nil {
		return nil, err
	}
	res := ToWeeklySummaryResponse(item)
	return &res, nil
}

func (s *summaryService) GenerateWeeklySummary(ctx context.Context, patientID string, weekStart string) (*WeeklySummaryResponse, error) {
	start, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return nil, errs.NewBadRequest("invalid week_start_date format (must be YYYY-MM-DD)", err)
	}

	end := start.AddDate(0, 0, 7) // exclusive upper bound

	// 1. Calculate values
	avg, err := s.repo.CalculateWeeklyAverage(ctx, patientID, start, end)
	if err != nil {
		return nil, err
	}

	completions, err := s.repo.CalculateCompletionsCount(ctx, patientID, start, end)
	if err != nil {
		return nil, err
	}

	// 2. Determine blood sugar trend status by comparing with previous week
	prevStart := start.AddDate(0, 0, -7)
	prevAvg, err := s.repo.CalculateWeeklyAverage(ctx, patientID, prevStart, start)

	trend := domain.TrendStabil
	if err == nil && prevAvg > 0 && avg > 0 {
		diffPercent := (avg - prevAvg) / prevAvg
		if diffPercent >= 0.05 {
			trend = domain.TrendMeningkat
		} else if diffPercent <= -0.05 {
			trend = domain.TrendMenurun
		}
	}

	summary := &domain.WeeklyHealthSummary{
		PatientID:           patientID,
		WeekStartDate:       start,
		WeekEndDate:         start.AddDate(0, 0, 6), // inclusive
		BloodSugarTrend:     trend,
		AvgBloodSugar:       avg,
		ArticlesReadCount:   completions,
		ArticlesTargetCount: 7, // standard weekly target
		GeneratedAt:         time.Now(),
	}

	if err = s.repo.Upsert(ctx, summary); err != nil {
		return nil, err
	}

	res := ToWeeklySummaryResponse(summary)
	return &res, nil
}
