package checkin

import (
	"context"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type CheckinRepository interface {
	Upsert(ctx context.Context, c *domain.DailyMedicalCheckin) error
	FindByMonth(ctx context.Context, patientID string, year int, month int) ([]domain.DailyMedicalCheckin, error)
}

type CheckinService interface {
	PerformCheckin(ctx context.Context, patientID string, checkinDate string) (*CheckinResponse, error)
	GetCheckinCalendar(ctx context.Context, patientID string, year int, month int) (*CheckinCalendarResponse, error)
}
