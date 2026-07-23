package patient

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/infrastructure/email"
	"github.com/dsmes/dsmes-backend/internal/modules/auth"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	jwtpkg "github.com/dsmes/dsmes-backend/internal/pkg/jwt"
)

type patientService struct {
	repo     PatientRepository
	authRepo auth.AuthRepository
	jwt      *jwtpkg.Manager
	email    email.EmailService
	log      *zap.Logger
}

func NewPatientService(repo PatientRepository, authRepo auth.AuthRepository, jwt *jwtpkg.Manager, email email.EmailService, log *zap.Logger) PatientService {
	return &patientService{repo: repo, authRepo: authRepo, jwt: jwt, email: email, log: log}
}

func (s *patientService) RegisterPatient(ctx context.Context, req RegisterPatientRequest) (*auth.LoginResponse, error) {
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
		return nil, errs.NewBadRequest("invalid date of birth format (must be YYYY-MM-DD or ISO string)", err)
	}

	// 4. Normalize Gender & BloodType & Activity
	gender := domain.GenderLakiLaki
	if strings.EqualFold(req.Gender, "perempuan") || strings.EqualFold(req.Gender, "female") {
		gender = domain.GenderPerempuan
	}

	bloodType := domain.BloodType(req.BloodType)
	cleanBlood := strings.ToLower(strings.TrimSpace(req.BloodType))
	if cleanBlood == "tidak tahu" || cleanBlood == "tidak_tahu" {
		bloodType = domain.BloodTypeTidakTahu
	}

	activityLevel := req.GetActivity()
	if activityLevel == "" {
		activityLevel = "Ringan"
	}

	patient := &domain.Patient{
		Email:                 req.Email,
		PasswordHash:          string(hash),
		FullName:              req.FullName,
		Nickname:              req.Nickname,
		WhatsappNumber:        req.GetPhone(),
		Gender:                gender,
		DateOfBirth:           dob,
		HeightCm:              req.HeightCm,
		WeightKg:              req.WeightKg,
		BloodType:             bloodType,
		PhysicalActivityLevel: activityLevel,
		DailyCalorieTarget:    2000, // Default target
		Status:                domain.StatusAktif,
	}


	// 4. Seed default routines & times
	defaultRoutines := []domain.Routine{
		{
			RoutineType:     domain.RoutineJalanPagi,
			DescriptiveName: "Jalan Pagi Sehat",
			BaseFrequency:   "Harian",
			IsActive:        true,
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
			RoutineType:     domain.RoutineMinumAir,
			DescriptiveName: "Minum Air Putih",
			BaseFrequency:   "Setiap 4 jam",
			IsActive:        true,
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
			RoutineType:     domain.RoutineCekGula,
			DescriptiveName: "Cek Gula Darah Harian",
			BaseFrequency:   "Harian",
			IsActive:        true,
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
			s.log.Warn("patient: welcome email delivery skipped or restricted by Resend test mode", zap.String("email", patient.Email), zap.Error(err))
		}
	}()


	// Generate JWT tokens for instant mobile login
	tokens, err := s.jwt.GenerateTokenPair(patient.ID, patient.Email, "user")
	if err != nil {
		return nil, errs.NewInternal("failed to generate tokens on register", err)
	}

	session := &auth.AuthSession{
		OwnerType:    auth.OwnerTypePatient,
		OwnerID:      patient.ID,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}
	if err = s.authRepo.CreateSession(ctx, session); err != nil {
		s.log.Warn("patient: failed to persist session on register", zap.Error(err), zap.String("patient_id", patient.ID))
	}

	return &auth.LoginResponse{
		User: auth.AuthUserResponse{
			ID:       patient.ID,
			FullName: patient.FullName,
			Email:    patient.Email,
			Role:     "user",
		},
		Tokens: *tokens,
	}, nil
}


func populateSummary(res *PatientResponse, summary *PatientSummaryData) {
	if summary == nil {
		return
	}
	res.LatestBloodSugar = summary.LatestBloodSugar
	if summary.LatestBloodSugarTime != nil {
		tStr := summary.LatestBloodSugarTime.Format("2006-01-02T15:04:05Z07:00")
		res.LatestBloodSugarTime = &tStr
	}
	res.LatestBloodSugarStatus = summary.LatestBloodSugarStatus
	res.AverageBloodSugar = summary.AverageBloodSugar
	res.LatestWeight = summary.LatestWeight
	res.BMI = summary.BMI
	res.LatestMealCalories = summary.LatestMealCalories
	res.LatestMealType = summary.LatestMealType
	res.LatestActivityName = summary.LatestActivityName
	if summary.LatestActivityTime != nil {
		tStr := summary.LatestActivityTime.Format("2006-01-02T15:04:05Z07:00")
		res.LatestActivityTime = &tStr
	}
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
	if len(items) == 0 {
		return resp, total, nil
	}

	// Batch fetch summaries — O(1) queries instead of O(N)
	patientIDs := make([]string, len(items))
	for i := range items {
		patientIDs[i] = items[i].ID
	}
	summaries, err := s.repo.GetPatientSummaries(ctx, patientIDs)
	if err != nil {
		s.log.Warn("patient: failed to batch fetch summaries", zap.Error(err))
	}

	for i := range items {
		resp[i] = ToPatientResponse(&items[i])
		if summary, ok := summaries[items[i].ID]; ok && summary != nil {
			populateSummary(&resp[i], summary)
		}
	}

	return resp, total, nil
}

func (s *patientService) GetPatient(ctx context.Context, id string) (*PatientDetailResponse, error) {
	patient, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := ToPatientDetailResponse(patient)

	summary, err := s.repo.GetPatientSummary(ctx, id)
	if err == nil && summary != nil {
		populateSummary(&res.PatientResponse, summary)
	}

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
	patient.PatientCode = req.PatientCode
	patient.Address = req.Address
	if req.DiagnosisDate != "" {
		if t, err := ParseDOB(req.DiagnosisDate); err == nil {
			patient.DiagnosisDate = &t
		}
	}
	patient.CurrentMedication = req.CurrentMedication
	patient.Allergies = req.Allergies
	patient.SmokingStatus = req.SmokingStatus
	patient.PhysicalActivityLevel = req.PhysicalActivityLevel

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
