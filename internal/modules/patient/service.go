package patient

import (
	"context"
	"fmt"
	"math"
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


func CalculateCalorieStatus(target int, consumed float64) *CalorieStatusInfo {
	if target <= 0 {
		target = 2000
	}

	achievement := (consumed / float64(target)) * 100.0
	achievement = math.Round(achievement*10) / 10

	diff := consumed - float64(target)
	diff = math.Round(diff)

	diffStr := ""
	if diff > 0 {
		diffStr = fmt.Sprintf("+%d kcal", int(diff))
	} else if diff < 0 {
		diffStr = fmt.Sprintf("%d kcal", int(diff))
	} else {
		diffStr = "0 kcal"
	}

	status := "Target Tercapai"
	code := "excellent"
	desc := "Asupan kalori harian sesuai dengan target rekomendasi."

	if achievement >= 95.0 && achievement <= 105.0 {
		status = "Target Tercapai"
		code = "excellent"
		desc = "Asupan kalori harian sesuai dengan target rekomendasi."
	} else if achievement >= 80.0 && achievement < 95.0 {
		status = "Sedikit di Bawah Target"
		code = "slightly_below"
		desc = "Pasien mengonsumsi sedikit lebih sedikit kalori dari target."
	} else if achievement >= 60.0 && achievement < 80.0 {
		status = "di Bawah Target"
		code = "below"
		desc = "Asupan kalori signifikan di bawah rekomendasi."
	} else if achievement < 60.0 {
		status = "Asupan Sangat Rendah"
		code = "very_low"
		desc = "Pasien mengonsumsi kalori sangat rendah dari target."
	} else if achievement > 105.0 && achievement <= 120.0 {
		status = "di Atas Target"
		code = "above"
		desc = "Asupan kalori pasien melebihi target rekomendasi."
	} else if achievement > 120.0 {
		status = "Asupan Sangat Tinggi"
		code = "excessive"
		desc = "Asupan kalori pasien jauh melebihi target rekomendasi."
	}

	return &CalorieStatusInfo{
		TargetCalories:        target,
		ConsumedCalories:      math.Round(consumed),
		AchievementPercentage: achievement,
		CalorieDifference:     diff,
		CalorieDifferenceStr:  diffStr,
		CalorieStatus:         status,
		CalorieStatusCode:     code,
		CalorieDescription:    desc,
	}
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

	consumed := 0.0
	if summary.TodayConsumedCalories != nil {
		consumed = *summary.TodayConsumedCalories
	}
	res.CalorieStatus = CalculateCalorieStatus(res.DailyCalorieTarget, consumed)
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
		score, label, _ := s.CalculateDynamicCompliance(ctx, items[i].ID, items[i].DailyCalorieTarget, items[i].CreatedAt)
		resp[i].Compliance = score
		resp[i].ComplianceLabel = label
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

	score, label, breakdown := s.CalculateDynamicCompliance(ctx, patient.ID, patient.DailyCalorieTarget, patient.CreatedAt)
	res.Compliance = score
	res.ComplianceLabel = label
	res.ComplianceBreakdown = breakdown

	summary, err := s.repo.GetPatientSummary(ctx, id)
	if err == nil && summary != nil {
		populateSummary(&res.PatientResponse, summary)
	}

	return &res, nil
}

func (s *patientService) GetPatientActivityAnalytics(ctx context.Context, patientID string, days int) (*PatientActivityAnalyticsResponse, error) {
	return s.repo.GetPatientActivityAnalytics(ctx, patientID, days)
}

func (s *patientService) CalculateDynamicCompliance(ctx context.Context, patientID string, dailyTarget int, createdAt time.Time) (int, string, *ComplianceBreakdown) {
	now := time.Now()
	endDate := now
	startDate := now.AddDate(0, 0, -6)

	if dailyTarget <= 0 {
		dailyTarget = 2000
	}

	daysSinceReg := int(now.Sub(createdAt).Hours()/24) + 1
	evalWindow := 7
	if daysSinceReg < evalWindow {
		evalWindow = daysSinceReg
	}
	if evalWindow < 1 {
		evalWindow = 1
	}

	dailyAggs, err := s.repo.GetPatientDailyLogsAggregate(ctx, patientID, startDate, endDate)
	if err != nil {
		s.log.Warn("patient: failed to fetch daily logs aggregate for compliance", zap.Error(err))
		dailyAggs = make(map[string]*DailyLogsAggregate)
	}

	var sumBS, sumFood, sumAct, sumMed float64

	for i := 0; i < evalWindow; i++ {
		d := now.AddDate(0, 0, -i).Format("2006-01-02")
		agg, ok := dailyAggs[d]
		if !ok || agg == nil {
			agg = &DailyLogsAggregate{}
		}

		// 1. Blood Sugar (25 pts max)
		bsScore := 0.0
		if agg.BloodSugarCount > 0 {
			bsScore = 25.0
		}

		// 2. Food & Calories (25 pts max)
		foodScore := 0.0
		if agg.MealCount > 0 {
			ratio := agg.TotalMealCalories / float64(dailyTarget)
			if ratio >= 0.85 && ratio <= 1.15 {
				foodScore = 25.0
			} else if (ratio >= 0.70 && ratio < 0.85) || (ratio > 1.15 && ratio <= 1.30) {
				foodScore = 17.5
			} else if (ratio >= 0.50 && ratio < 0.70) || (ratio > 1.30 && ratio <= 1.50) {
				foodScore = 10.0
			} else if ratio > 0 {
				foodScore = 5.0
			}
		}

		// 3. Activity (25 pts max, 30 min target)
		actScore := 0.0
		if agg.TotalActivityMinutes > 0 {
			actRatio := float64(agg.TotalActivityMinutes) / 30.0
			if actRatio > 1.0 {
				actRatio = 1.0
			}
			actScore = actRatio * 25.0
		}

		// 4. Medication (25 pts max)
		medScore := 0.0
		if agg.MedicationScheduledCount > 0 {
			medRatio := float64(agg.MedicationCompletedCount) / float64(agg.MedicationScheduledCount)
			if medRatio > 1.0 {
				medRatio = 1.0
			}
			medScore = medRatio * 25.0
		} else {
			// No medication reminders configured — redistribute 25 pts equally among active 3 pillars (+8.333 each)
			if bsScore > 0 {
				bsScore += 8.333
			}
			if foodScore > 0 {
				foodScore += (foodScore / 25.0) * 8.333
			}
			if actScore > 0 {
				actScore += (actScore / 25.0) * 8.333
			}
		}

		sumBS += bsScore
		sumFood += foodScore
		sumAct += actScore
		sumMed += medScore
	}

	w := float64(evalWindow)
	avgBS := math.Round((sumBS/w)*10) / 10
	avgFood := math.Round((sumFood/w)*10) / 10
	avgAct := math.Round((sumAct/w)*10) / 10
	avgMed := math.Round((sumMed/w)*10) / 10

	totalAvg := int(math.Round((sumBS + sumFood + sumAct + sumMed) / w))
	if totalAvg > 100 {
		totalAvg = 100
	}
	if totalAvg < 0 {
		totalAvg = 0
	}

	label := "Kurang"
	if totalAvg >= 90 {
		label = "Sangat Patuh"
	} else if totalAvg >= 75 {
		label = "Patuh"
	} else if totalAvg >= 60 {
		label = "Cukup"
	} else if totalAvg >= 40 {
		label = "Kurang"
	} else {
		label = "Tidak Patuh"
	}

	breakdown := &ComplianceBreakdown{
		BloodSugarScore: avgBS,
		FoodScore:       avgFood,
		ActivityScore:   avgAct,
		MedicationScore: avgMed,
	}

	return totalAvg, label, breakdown
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
