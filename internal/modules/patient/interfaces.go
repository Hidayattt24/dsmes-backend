package patient

import (
	"context"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type PatientRepository interface {
	FindAll(ctx context.Context, filter PatientFilterQuery) ([]domain.Patient, int64, error)
	FindByID(ctx context.Context, id string) (*domain.Patient, error)
	FindByEmail(ctx context.Context, email string) (*domain.Patient, error)
	Create(ctx context.Context, p *domain.Patient) error
	Update(ctx context.Context, p *domain.Patient) error
	Delete(ctx context.Context, id string) error

	// Transactional onboarding helpers
	CreateWithOnboarding(ctx context.Context, p *domain.Patient, defaultRoutines []domain.Routine, defaultReminders []domain.Reminder) error
}

type PatientService interface {
	RegisterPatient(ctx context.Context, req RegisterPatientRequest) (*PatientDetailResponse, error)
	ListPatients(ctx context.Context, filter PatientFilterQuery) ([]PatientResponse, int64, error)
	GetPatient(ctx context.Context, id string) (*PatientDetailResponse, error)
	UpdateProfile(ctx context.Context, patientID string, req UpdatePatientProfileRequest) (*PatientResponse, error)
	AssignPuskesmas(ctx context.Context, id string, req AssignPuskesmasRequest) (*PatientDetailResponse, error)
	ToggleStatus(ctx context.Context, id string) (*PatientResponse, error)
	DeletePatient(ctx context.Context, id string) error
}
