package patient

import (
	"context"
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/modules/auth"
)

type PatientRepository interface {
	FindAll(ctx context.Context, filter PatientFilterQuery) ([]domain.Patient, int64, error)
	FindByID(ctx context.Context, id string) (*domain.Patient, error)
	FindByEmail(ctx context.Context, email string) (*domain.Patient, error)
	FindByPhoneNumber(ctx context.Context, phone string) (*domain.Patient, error)
	Create(ctx context.Context, p *domain.Patient) error
	Update(ctx context.Context, p *domain.Patient) error
	Delete(ctx context.Context, id string) error
	GetStats(ctx context.Context, staffID string) (*PatientStats, error)
	GetPatientSummary(ctx context.Context, patientID string) (*PatientSummaryData, error)
	GetPatientSummaries(ctx context.Context, patientIDs []string) (map[string]*PatientSummaryData, error)
	GetPatientActivityAnalytics(ctx context.Context, patientID string, days int) (*PatientActivityAnalyticsResponse, error)
	GetPatientDailyLogsAggregate(ctx context.Context, patientID string, startDate, endDate time.Time) (map[string]*DailyLogsAggregate, error)

	// Health Measurements
	CreateMeasurement(ctx context.Context, m *domain.PatientMeasurement) error
	GetPatientMeasurements(ctx context.Context, patientID string) ([]domain.PatientMeasurement, error)
	GetLatestMeasurement(ctx context.Context, patientID string) (*domain.PatientMeasurement, error)
	FindMeasurementByID(ctx context.Context, measurementID string) (*domain.PatientMeasurement, error)
	UpdateMeasurement(ctx context.Context, m *domain.PatientMeasurement) error
	CreateBloodSugarLog(ctx context.Context, bsLog *domain.BloodSugarLog) error

	// Transactional onboarding helpers
	CreateWithOnboarding(ctx context.Context, p *domain.Patient, defaultRoutines []domain.Routine, defaultReminders []domain.Reminder) error
}

type PatientService interface {
	RegisterPatient(ctx context.Context, req RegisterPatientRequest) (*auth.LoginResponse, error)
	SetupHealthProfile(ctx context.Context, patientID string, req SetupHealthProfileRequest) (*PatientDetailResponse, error)

	ListPatients(ctx context.Context, filter PatientFilterQuery) ([]PatientResponse, int64, error)
	GetPatient(ctx context.Context, id string) (*PatientDetailResponse, error)
	UpdateProfile(ctx context.Context, patientID string, req UpdatePatientProfileRequest) (*PatientResponse, error)
	ChangePassword(ctx context.Context, patientID string, req ChangePasswordRequest) error
	UpdatePatientByAdmin(ctx context.Context, patientID string, req UpdatePatientRequest) (*PatientDetailResponse, error)
	AssignStaff(ctx context.Context, id string, req AssignStaffRequest) (*PatientDetailResponse, error)
	ToggleStatus(ctx context.Context, id string) (*PatientResponse, error)
	DeletePatient(ctx context.Context, id string) error
	GetStats(ctx context.Context, staffID string) (*PatientStats, error)
	GetPatientActivityAnalytics(ctx context.Context, patientID string, days int) (*PatientActivityAnalyticsResponse, error)

	// Health Measurements
	CreateMeasurement(ctx context.Context, patientID string, req CreateMeasurementRequest, recordedByID, recordedByName, recordedByRole string) (*PatientMeasurementResponse, error)
	GetPatientMeasurements(ctx context.Context, patientID string) ([]PatientMeasurementResponse, error)
	UpdateMeasurement(ctx context.Context, patientID, measurementID string, req UpdateMeasurementRequest) (*PatientMeasurementResponse, error)
}
