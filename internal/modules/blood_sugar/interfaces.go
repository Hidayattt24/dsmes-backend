package blood_sugar

import (
	"context"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type BloodSugarRepository interface {
	Create(ctx context.Context, log *domain.BloodSugarLog) error
	FindAllByPatientID(ctx context.Context, patientID string, page, limit int) ([]domain.BloodSugarLog, int64, error)
	GetDistributionForStaff(ctx context.Context, staffID string) (*GlucoseDistributionResponse, error)
}

type BloodSugarService interface {
	LogBloodSugar(ctx context.Context, patientID string, req LogBloodSugarRequest) (*BloodSugarResponse, error)
	GetPatientHistory(ctx context.Context, patientID string, page, limit int) ([]BloodSugarResponse, int64, error)
	GetStaffDashboard(ctx context.Context, staffID string) (*GlucoseDistributionResponse, error)
}
