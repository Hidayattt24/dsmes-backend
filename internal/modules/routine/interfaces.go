package routine

import (
	"context"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type RoutineRepository interface {
	FindAllByPatientID(ctx context.Context, patientID string) ([]domain.Routine, error)
	FindTimeByID(ctx context.Context, id string) (*domain.RoutineTime, error)
	UpdateTime(ctx context.Context, t *domain.RoutineTime) error
	CreateLog(ctx context.Context, log *domain.RoutineLogEntry) error
	FindLogsByPatientAndDate(ctx context.Context, patientID string, dateStr string) ([]domain.RoutineLogEntry, error)
	ReplacePatientRoutines(ctx context.Context, patientID string, routines []domain.Routine) error
	CreateActivityLog(ctx context.Context, log *domain.PatientActivityLog) error
	FindFreeActivityLogsByPatientAndDate(ctx context.Context, patientID string, dateStr string) ([]domain.PatientActivityLog, error)
}

type RoutineService interface {
	ListRoutines(ctx context.Context, patientID string) ([]RoutineResponse, error)
	ConfigureRoutineTime(ctx context.Context, patientID string, routineTimeID string, req UpdateRoutineTimeRequest) (*RoutineTimeResponse, error)
	BulkSetupRoutines(ctx context.Context, patientID string, req BulkSetupRoutinesRequest) ([]RoutineResponse, error)
	LogRoutine(ctx context.Context, patientID string, req LogRoutineRequest) (*RoutineLogResponse, error)
	GetOnboardingStatus(ctx context.Context, patientID string) (*OnboardingStatusResponse, error)
	GetPatientActivityLogs(ctx context.Context, patientID string, dateStr string) ([]ActivityLogResponse, error)
	LogActivity(ctx context.Context, patientID string, req LogActivityRequest) (*LogActivityResponse, error)
}
