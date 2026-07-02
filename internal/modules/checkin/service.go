package checkin

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type checkinService struct {
	repo CheckinRepository
	log  *zap.Logger
}

func NewCheckinService(repo CheckinRepository, log *zap.Logger) CheckinService {
	return &checkinService{repo: repo, log: log}
}

func (s *checkinService) PerformCheckin(ctx context.Context, patientID string, checkinDate string) (*CheckinResponse, error) {
	d, err := time.Parse("2006-01-02", checkinDate)
	if err != nil {
		return nil, errs.NewBadRequest("invalid checkin_date format (must be YYYY-MM-DD)", err)
	}

	checkin := &domain.DailyMedicalCheckin{
		PatientID:   patientID,
		CheckinDate: d,
		IsCompleted: true,
	}

	if err = s.repo.Upsert(ctx, checkin); err != nil {
		return nil, err
	}

	return &CheckinResponse{
		ID:          checkin.ID,
		CheckinDate: checkinDate,
		IsCompleted: checkin.IsCompleted,
	}, nil
}

func (s *checkinService) GetCheckinCalendar(ctx context.Context, patientID string, year int, month int) (*CheckinCalendarResponse, error) {
	items, err := s.repo.FindByMonth(ctx, patientID, year, month)
	if err != nil {
		return nil, err
	}

	completedDates := make([]string, 0, len(items))
	for _, item := range items {
		if item.IsCompleted {
			completedDates = append(completedDates, item.CheckinDate.Format("2006-01-02"))
		}
	}

	return &CheckinCalendarResponse{
		CompletedDates: completedDates,
	}, nil
}
