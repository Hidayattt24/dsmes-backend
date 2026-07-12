package patient

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/infrastructure/email"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type patientService struct {
	repo  PatientRepository
	email email.EmailService
	log   *zap.Logger
}

func NewPatientService(repo PatientRepository, email email.EmailService, log *zap.Logger) PatientService {
	return &patientService{repo: repo, email: email, log: log}
}

func (s *patientService) RegisterPatient(ctx context.Context, req RegisterPatientRequest) (*PatientDetailResponse, error) {
	// 1. Unique Check
	_, err := s.repo.FindByEmail(ctx, req.Email)
	if err == nil {
		return nil, errs.NewConflict("email already registered")
	}

	// 2. Hash Password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errs.NewInternal("failed to hash password", err)
	}

	// 3. Parse DOB
	dob, err := ParseDOB(req.DateOfBirth)
	if err != nil {
		return nil, errs.NewBadRequest("invalid date of birth format (must be YYYY-MM-DD)", err)
	}

	patient := &domain.Patient{
		Email:              req.Email,
		PasswordHash:       string(hash),
		FullName:           req.FullName,
		Nickname:           req.Nickname,
		WhatsappNumber:     req.WhatsappNumber,
		Gender:             req.Gender,
		DateOfBirth:        dob,
		HeightCm:           req.HeightCm,
		WeightKg:           req.WeightKg,
		BloodType:          req.BloodType,
		DailyCalorieTarget: 2000, // Default target
		Status:             domain.StatusAktif,
	}

	// 4. Seed default routines & times
	defaultRoutines := []domain.Routine{
		{
			RoutineType:      domain.RoutineJalanPagi,
			DescriptiveName:  "Jalan Pagi Sehat",
			BaseFrequency:    "Harian",
			IsActive:         true,
			RoutineTimes: []domain.RoutineTime{
				{
					TimeType:       domain.WaktuDefault,
					ScheduledTime:  strPtr("06:00:00"),
					Status:         domain.WaktuSet,
					ReminderActive: true,
				},
			},
		},
		{
			RoutineType:      domain.RoutineMinumAir,
			DescriptiveName:  "Minum Air Putih",
			BaseFrequency:    "Setiap 4 jam",
			IsActive:         true,
			RoutineTimes: []domain.RoutineTime{
				{
					TimeType:       domain.WaktuDefault,
					ScheduledTime:  strPtr("08:00:00"),
					Status:         domain.WaktuSet,
					ReminderActive: true,
				},
				{
					TimeType:       domain.WaktuDefault,
					ScheduledTime:  strPtr("12:00:00"),
					Status:         domain.WaktuSet,
					ReminderActive: true,
				},
				{
					TimeType:       domain.WaktuDefault,
					ScheduledTime:  strPtr("16:00:00"),
					Status:         domain.WaktuSet,
					ReminderActive: true,
				},
				{
					TimeType:       domain.WaktuDefault,
					ScheduledTime:  strPtr("20:00:00"),
					Status:         domain.WaktuSet,
					ReminderActive: true,
				},
			},
		},
		{
			RoutineType:      domain.RoutineCekGula,
			DescriptiveName:  "Cek Gula Darah Harian",
			BaseFrequency:    "Harian",
			IsActive:         true,
			RoutineTimes: []domain.RoutineTime{
				{
					TimeType:       domain.WaktuDefault,
					ScheduledTime:  strPtr("07:00:00"),
					Status:         domain.WaktuSet,
					ReminderActive: true,
				},
			},
		},
	}

	// 5. Seed default reminders using system templates inside repo transaction
	// Let's pass empty slice; repository will auto-query system templates
	var defaultReminders []domain.Reminder

	if err = s.repo.CreateWithOnboarding(ctx, patient, defaultRoutines, defaultReminders); err != nil {
		return nil, err
	}

	// Send welcome email in the background
	go func() {
		bgCtx := context.Background()
		if err := s.email.SendWelcomeEmail(bgCtx, patient.Email, patient.FullName); err != nil {
			s.log.Error("patient: failed to send welcome email", zap.String("email", patient.Email), zap.Error(err))
		}
	}()

	res := ToPatientDetailResponse(patient)
	return &res, nil
}

func (s *patientService) ListPatients(ctx context.Context, filter PatientFilterQuery) ([]PatientResponse, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}

	items, total, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]PatientResponse, len(items))
	for i := range items {
		resp[i] = ToPatientResponse(&items[i])
	}

	return resp, total, nil
}

func (s *patientService) GetPatient(ctx context.Context, id string) (*PatientDetailResponse, error) {
	patient, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := ToPatientDetailResponse(patient)
	return &res, nil
}

func (s *patientService) UpdateProfile(ctx context.Context, patientID string, req UpdatePatientProfileRequest) (*PatientResponse, error) {
	patient, err := s.repo.FindByID(ctx, patientID)
	if err != nil {
		return nil, err
	}

	patient.FullName = req.FullName
	patient.Nickname = req.Nickname
	patient.WhatsappNumber = req.WhatsappNumber
	patient.HeightCm = req.HeightCm
	patient.WeightKg = req.WeightKg
	if req.ProfilePhotoURL != "" {
		patient.ProfilePhotoURL = req.ProfilePhotoURL
	}
	patient.BPJS = req.BPJS
	patient.NIK = req.NIK
	patient.EmergencyName = req.EmergencyName
	patient.EmergencyRelation = req.EmergencyRelation
	patient.EmergencyPhone = req.EmergencyPhone
	patient.DiabetesType = req.DiabetesType
	patient.InterventionType = req.InterventionType

	if err = s.repo.Update(ctx, patient); err != nil {
		return nil, err
	}

	res := ToPatientResponse(patient)
	return &res, nil
}

func (s *patientService) AssignStaff(ctx context.Context, id string, req AssignStaffRequest) (*PatientDetailResponse, error) {
	patient, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	patient.AssignedStaffID = &req.StaffID

	if err = s.repo.Update(ctx, patient); err != nil {
		return nil, err
	}

	// Refetch to preload assigned staff
	refetched, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	res := ToPatientDetailResponse(refetched)
	return &res, nil
}

func (s *patientService) ToggleStatus(ctx context.Context, id string) (*PatientResponse, error) {
	patient, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if patient.Status == domain.StatusAktif {
		patient.Status = domain.StatusNonaktif
	} else {
		patient.Status = domain.StatusAktif
	}

	if err = s.repo.Update(ctx, patient); err != nil {
		return nil, err
	}

	res := ToPatientResponse(patient)
	return &res, nil
}

func (s *patientService) DeletePatient(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *patientService) GetStats(ctx context.Context, staffID string) (*PatientStats, error) {
	return s.repo.GetStats(ctx, staffID)
}

// Helpers
func strPtr(s string) *string {
	return &s
}

