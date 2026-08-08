package blood_sugar

import (
	"context"
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type BloodSugarRepository interface {
	Create(ctx context.Context, log *domain.BloodSugarLog) error
	FindByID(ctx context.Context, id string) (*domain.BloodSugarLog, error)
	FindAllByPatientID(ctx context.Context, patientID string, page, limit int) ([]domain.BloodSugarLog, int64, error)
	Update(ctx context.Context, log *domain.BloodSugarLog) error
	Delete(ctx context.Context, id string) error
	GetDistributionForStaff(ctx context.Context, staffID string) (*GlucoseDistributionResponse, error)
	GetPatientDOB(ctx context.Context, patientID string) (*time.Time, error)
}

type BloodSugarService interface {
	LogBloodSugar(ctx context.Context, patientID string, req LogBloodSugarRequest) (*BloodSugarResponse, error)
	UpdateBloodSugar(ctx context.Context, patientID, id string, req LogBloodSugarRequest) (*BloodSugarResponse, error)
	DeleteBloodSugar(ctx context.Context, patientID, id string) error
	GetPatientHistory(ctx context.Context, patientID string, page, limit int) ([]BloodSugarResponse, int64, error)
	GetStaffDashboard(ctx context.Context, staffID string) (*GlucoseDistributionResponse, error)
}
