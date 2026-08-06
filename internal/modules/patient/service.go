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
	"github.com/dsmes/dsmes-backend/internal/modules/nutrition"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	jwtpkg "github.com/dsmes/dsmes-backend/internal/pkg/jwt"
	"github.com/dsmes/dsmes-backend/internal/pkg/phone"
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
	// 1. Phone Number Normalization & Unique Check
	rawPhone := req.GetPhone()
	phoneNum, err := phone.Normalize(rawPhone)
	if err != nil {
		return nil, errs.NewBadRequest(err.Error())
	}

	_, err = s.repo.FindByPhoneNumber(ctx, phoneNum)
	if err == nil {
		return nil, errs.NewConflict("nomor handphone sudah terdaftar")
	}

	// 2. Optional Email Check
	var emailPtr *string
	cleanEmail := strings.TrimSpace(req.Email)
	if cleanEmail != "" {
		_, err := s.repo.FindByEmail(ctx, cleanEmail)
		if err == nil {
			return nil, errs.NewConflict("email sudah terdaftar")
		}
		emailPtr = &cleanEmail
	}

	// 3. Hash Password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errs.NewInternal("failed to hash password", err)
	}

	// 4. Optional Parse DOB
	var dob time.Time
	if req.DateOfBirth != "" {
		dob, _ = ParseDOB(req.DateOfBirth)
	}

	// 5. Normalize Gender & BloodType & Activity
	gender := domain.GenderLakiLaki
	if strings.EqualFold(req.Gender, "perempuan") || strings.EqualFold(req.Gender, "female") {
		gender = domain.GenderPerempuan
	}

	bloodType := domain.BloodType(req.BloodType)
	cleanBlood := strings.ToLower(strings.TrimSpace(req.BloodType))
	if cleanBlood == "" || cleanBlood == "tidak tahu" || cleanBlood == "tidak_tahu" {
		bloodType = domain.BloodTypeTidakTahu
	}

	activityLevel := req.GetActivity()
	if activityLevel == "" {
		activityLevel = "Ringan"
	}

	age := 0
	if !dob.IsZero() {
		age = time.Now().Year() - dob.Year()
	}

	dailyCalorie := domain.DefaultDailyCalorieTarget
	if req.WeightKg > 0 && req.HeightCm > 0 {
		dailyCalorie = CalculateDSMESCalorieTarget(string(gender), req.WeightKg, req.HeightCm, age)
	}

	patient := &domain.Patient{
		PhoneNumber:           phoneNum,
		Email:                 emailPtr,
		PasswordHash:          string(hash),
		FullName:              req.FullName,
		Nickname:              req.Nickname,
		WhatsappNumber:        phoneNum,
		Gender:                gender,
		DateOfBirth:           dob,
		HeightCm:              req.HeightCm,
		WeightKg:              req.WeightKg,
		BloodType:             bloodType,
		PhysicalActivityLevel: activityLevel,
		DailyCalorieTarget:    dailyCalorie,
		Status:                domain.StatusAktif,
	}

	// 6. Seed default routines & times
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
			BaseFrequency:   "Harian",
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
					ScheduledTime:  strPtr("15:00:00"),
					Status:         domain.WaktuSet,
					ReminderActive: true,
				},
				{
					TimeType:       domain.WaktuDefault,
					ScheduledTime:  strPtr("18:00:00"),
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

	// Seed default reminders using system templates inside repo transaction
	var defaultReminders []domain.Reminder

	if err = s.repo.CreateWithOnboarding(ctx, patient, defaultRoutines, defaultReminders); err != nil {
		return nil, err
	}

	// Send welcome email in the background if email was provided
	if patient.Email != nil && *patient.Email != "" {
		targetEmail := *patient.Email
		go func() {
			bgCtx := context.Background()
			if err := s.email.SendWelcomeEmail(bgCtx, targetEmail, patient.FullName); err != nil {
				s.log.Warn("patient: welcome email delivery skipped or restricted by Resend test mode", zap.String("email", targetEmail), zap.Error(err))
			}
		}()
	}

	// Generate JWT tokens for instant mobile login
	tokens, err := s.jwt.GenerateTokenPair(patient.ID, patient.GetEmail(), "user")
	if err != nil {
		return nil, errs.NewInternal("failed to generate tokens on register", err)
	}

	session := &auth.AuthSession{
		OwnerType:    auth.OwnerTypePatient,
		OwnerID:      patient.ID,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    time.Now().Add(s.jwt.RefreshTTL()),
	}
	if err = s.authRepo.CreateSession(ctx, session); err != nil {
		s.log.Warn("patient: failed to persist session on register", zap.Error(err), zap.String("patient_id", patient.ID))
	}

	// Seed initial baseline measurement record in DB if metrics were provided
	if req.WeightKg > 0 && req.HeightCm > 0 {
		hM := req.HeightCm / 100.0
		val := math.Round((req.WeightKg/(hM*hM))*10) / 10
		bmiVal := &val
		initCalTarget := patient.DailyCalorieTarget

		initMeasurement := &domain.PatientMeasurement{
			PatientID:          patient.ID,
			WeightKg:           &req.WeightKg,
			HeightCm:           &req.HeightCm,
			BMI:                bmiVal,
			DailyCalorieTarget: &initCalTarget,
			Notes:              "Pengukuran Awal (Registrasi Akun Pasien)",
			RecordedByID:       &patient.ID,
			RecordedByName:     "Sistem (Registrasi Awal)",
			RecordedByRole:     "admin",
			MeasuredAt:         patient.CreatedAt,
		}
		_ = s.repo.CreateMeasurement(ctx, initMeasurement)
	}

	return &auth.LoginResponse{
		User: auth.AuthUserResponse{
			ID:          patient.ID,
			FullName:    patient.FullName,
			PhoneNumber: patient.PhoneNumber,
			Email:       patient.GetEmail(),
			Role:        "user",
		},
		Tokens: *tokens,
	}, nil
}

func (s *patientService) SetupHealthProfile(ctx context.Context, patientID string, req SetupHealthProfileRequest) (*PatientDetailResponse, error) {
	patient, err := s.repo.FindByID(ctx, patientID)
	if err != nil {
		return nil, err
	}

	dob, err := ParseDOB(req.DateOfBirth)
	if err != nil {
		return nil, errs.NewBadRequest("invalid date of birth format (must be YYYY-MM-DD or ISO string)", err)
	}

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

	age := time.Now().Year() - dob.Year()

	// Calculate using standardized nutrition calculator (Mifflin-St Jeor + TDEE multiplier)
	calcRes, calcErr := nutrition.CalculateDailyCalories(nutrition.CalorieCalculationRequest{
		Gender:        req.Gender,
		DateOfBirth:   req.DateOfBirth,
		HeightCm:      req.HeightCm,
		WeightKg:      req.WeightKg,
		ActivityLevel: activityLevel,
	})

	dailyCalorieTarget := domain.DefaultDailyCalorieTarget
	if calcErr == nil && calcRes != nil {
		dailyCalorieTarget = calcRes.TDEE
		patient.MaintenanceCalories = calcRes.Recommendations.Maintain.Calories
		patient.MildWeightLossCalories = calcRes.Recommendations.MildLoss.Calories
		patient.WeightLossCalories = calcRes.Recommendations.WeightLoss.Calories
		patient.ExtremeWeightLossCalories = calcRes.Recommendations.ExtremeLoss.Calories
		patient.MaintenancePercentage = calcRes.Recommendations.Maintain.Percentage
		patient.MildPercentage = calcRes.Recommendations.MildLoss.Percentage
		patient.WeightLossPercentage = calcRes.Recommendations.WeightLoss.Percentage
		patient.ExtremePercentage = calcRes.Recommendations.ExtremeLoss.Percentage
	} else {
		dailyCalorieTarget = CalculateDSMESCalorieTarget(string(gender), req.WeightKg, req.HeightCm, age)
	}

	patient.Gender = gender
	patient.DateOfBirth = dob
	patient.HeightCm = req.HeightCm
	patient.WeightKg = req.WeightKg
	patient.BloodType = bloodType
	patient.PhysicalActivityLevel = activityLevel
	patient.DailyCalorieTarget = dailyCalorieTarget

	if err := s.repo.Update(ctx, patient); err != nil {
		return nil, err
	}

	// Check if there is an existing initial measurement from registration to update (to avoid duplicate logs)
	latestMeasurement, latestErr := s.repo.GetLatestMeasurement(ctx, patient.ID)
	if latestErr == nil && latestMeasurement != nil &&
		(latestMeasurement.Notes == "Pengukuran Awal (Registrasi Akun Pasien)" || latestMeasurement.Notes == "Pengukuran Awal (Registrasi)") &&
		latestMeasurement.WeightKg != nil && *latestMeasurement.WeightKg == req.WeightKg &&
		latestMeasurement.HeightCm != nil && *latestMeasurement.HeightCm == req.HeightCm {

		latestMeasurement.DailyCalorieTarget = &dailyCalorieTarget
		latestMeasurement.Notes = "Setup Profil Kesehatan (Onboarding Phase 2)"
		latestMeasurement.RecordedByID = &patient.ID
		latestMeasurement.RecordedByName = patient.FullName
		latestMeasurement.RecordedByRole = "user"
		latestMeasurement.MeasuredAt = time.Now()

		_ = s.repo.UpdateMeasurement(ctx, latestMeasurement)
	} else {
		// Create new measurement entry
		hM := req.HeightCm / 100.0
		val := math.Round((req.WeightKg/(hM*hM))*10) / 10
		bmiVal := &val

		measurement := &domain.PatientMeasurement{
			PatientID:          patient.ID,
			WeightKg:           &req.WeightKg,
			HeightCm:           &req.HeightCm,
			BMI:                bmiVal,
			DailyCalorieTarget: &dailyCalorieTarget,
			Notes:              "Setup Profil Kesehatan (Onboarding Phase 2)",
			RecordedByID:       &patient.ID,
			RecordedByName:     patient.FullName,
			RecordedByRole:     "user",
			MeasuredAt:         time.Now(),
		}
		_ = s.repo.CreateMeasurement(ctx, measurement)
	}

	return s.GetPatient(ctx, patient.ID)
}

func CalculateCalorieStatus(target int, consumed float64) *CalorieStatusInfo {
	if target <= 0 {
		target = domain.DefaultDailyCalorieTarget
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

	// Batch fetch daily aggregates once for the whole page (avoids N+1 queries).
	now := time.Now()
	aggregates, aggErr := s.repo.GetPatientDailyLogsAggregates(ctx, patientIDs, now.AddDate(0, 0, -6), now)
	if aggErr != nil {
		s.log.Warn("patient: failed to batch fetch daily aggregates", zap.Error(aggErr))
		aggregates = make(map[string]map[string]*DailyLogsAggregate)
	}

	for i := range items {
		resp[i] = ToPatientResponse(&items[i])
		score, label, _ := complianceFromAggregates(aggregates[items[i].ID], items[i].DailyCalorieTarget, items[i].CreatedAt, now)
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

	measurements, err := s.GetPatientMeasurements(ctx, id)
	if err == nil {
		res.Measurements = measurements
		if len(measurements) > 0 {
			res.LatestMeasurement = &measurements[0]
			if measurements[0].WeightKg != nil && *measurements[0].WeightKg > 0 {
				res.WeightKg = *measurements[0].WeightKg
				res.LatestWeight = measurements[0].WeightKg
			}
			if measurements[0].HeightCm != nil && *measurements[0].HeightCm > 0 {
				res.HeightCm = *measurements[0].HeightCm
			}
			if measurements[0].BMI != nil && *measurements[0].BMI > 0 {
				res.BMI = measurements[0].BMI
			}
			if measurements[0].WaistCircumferenceCm != nil && *measurements[0].WaistCircumferenceCm > 0 {
				res.WaistCircumferenceCm = measurements[0].WaistCircumferenceCm
			}
		}
	}

	return &res, nil
}

func (s *patientService) GetPatientActivityAnalytics(ctx context.Context, patientID string, days int) (*PatientActivityAnalyticsResponse, error) {
	return s.repo.GetPatientActivityAnalytics(ctx, patientID, days)
}

func (s *patientService) CalculateDynamicCompliance(ctx context.Context, patientID string, dailyTarget int, createdAt time.Time) (int, string, *ComplianceBreakdown) {
	now := time.Now()
	dailyAggs, err := s.repo.GetPatientDailyLogsAggregate(ctx, patientID, now.AddDate(0, 0, -6), now)
	if err != nil {
		s.log.Warn("patient: failed to fetch daily logs aggregate for compliance", zap.Error(err))
		dailyAggs = make(map[string]*DailyLogsAggregate)
	}
	return complianceFromAggregates(dailyAggs, dailyTarget, createdAt, now)
}

// complianceFromAggregates scores a patient's compliance from pre-fetched daily
// aggregates. Extracted so ListPatients can score all patients on a page from a
// single batched repository call instead of issuing one query per patient.
func complianceFromAggregates(dailyAggs map[string]*DailyLogsAggregate, dailyTarget int, createdAt time.Time, now time.Time) (int, string, *ComplianceBreakdown) {
	if dailyTarget <= 0 {
		dailyTarget = domain.DefaultDailyCalorieTarget
	}

	daysSinceReg := int(now.Sub(createdAt).Hours()/24) + 1
	evalWindow := 7
	if daysSinceReg < evalWindow {
		evalWindow = daysSinceReg
	}
	if evalWindow < 1 {
		evalWindow = 1
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
	if req.Gender != "" {
		gender := domain.GenderLakiLaki
		if strings.EqualFold(req.Gender, "perempuan") || strings.EqualFold(req.Gender, "female") {
			gender = domain.GenderPerempuan
		}
		patient.Gender = gender
	}
	if req.DateOfBirth != "" {
		if t, err := ParseDOB(req.DateOfBirth); err == nil {
			patient.DateOfBirth = t
		}
	}
	if req.BloodType != "" {
		bloodType := domain.BloodType(req.BloodType)
		cleanBlood := strings.ToLower(strings.TrimSpace(req.BloodType))
		if cleanBlood == "" || cleanBlood == "tidak tahu" || cleanBlood == "tidak_tahu" {
			bloodType = domain.BloodTypeTidakTahu
		}
		patient.BloodType = bloodType
	}
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

	// Always trigger automatic calorie & BMI recalculation on profile update
	dobStr := ""
	if !patient.DateOfBirth.IsZero() {
		dobStr = patient.DateOfBirth.Format("2006-01-02")
	}

	calcRes, calcErr := nutrition.CalculateDailyCalories(nutrition.CalorieCalculationRequest{
		Gender:        string(patient.Gender),
		DateOfBirth:   dobStr,
		HeightCm:      patient.HeightCm,
		WeightKg:      patient.WeightKg,
		ActivityLevel: patient.PhysicalActivityLevel,
	})
	if calcErr != nil {
		return nil, calcErr
	}

	patient.DailyCalorieTarget = calcRes.TDEE
	patient.MaintenanceCalories = calcRes.Recommendations.Maintain.Calories
	patient.MildWeightLossCalories = calcRes.Recommendations.MildLoss.Calories
	patient.WeightLossCalories = calcRes.Recommendations.WeightLoss.Calories
	patient.ExtremeWeightLossCalories = calcRes.Recommendations.ExtremeLoss.Calories
	patient.MaintenancePercentage = calcRes.Recommendations.Maintain.Percentage
	patient.MildPercentage = calcRes.Recommendations.MildLoss.Percentage
	patient.WeightLossPercentage = calcRes.Recommendations.WeightLoss.Percentage
	patient.ExtremePercentage = calcRes.Recommendations.ExtremeLoss.Percentage

	// Insert updated measurement record for tracking history
	hM := patient.HeightCm / 100.0
	bmiVal := math.Round((patient.WeightKg/(hM*hM))*10) / 10
	tdeeVal := calcRes.TDEE

	_ = s.repo.CreateMeasurement(ctx, &domain.PatientMeasurement{
		PatientID:          patient.ID,
		WeightKg:           &patient.WeightKg,
		HeightCm:           &patient.HeightCm,
		BMI:                &bmiVal,
		DailyCalorieTarget: &tdeeVal,
		Notes:              "Pembaruan Profil Kesehatan (Rekalkulasi Otomatis)",
		RecordedByID:       &patient.ID,
		RecordedByName:     patient.FullName,
		RecordedByRole:     "user",
		MeasuredAt:         time.Now(),
	})

	if err = s.repo.Update(ctx, patient); err != nil {
		return nil, err
	}

	res := ToPatientResponse(patient)
	return &res, nil
}

func (s *patientService) ChangePassword(ctx context.Context, patientID string, req ChangePasswordRequest) error {
	patient, err := s.repo.FindByID(ctx, patientID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(patient.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return errs.NewUnauthorized("kata sandi saat ini salah")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errs.NewInternal("failed to hash password", err)
	}

	patient.PasswordHash = string(hash)
	if err := s.repo.Update(ctx, patient); err != nil {
		return err
	}

	// Invalidate every active session so old refresh tokens stop working.
	return s.authRepo.RevokeAllSessions(ctx, auth.OwnerTypePatient, patientID)
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

func ToPatientMeasurementResponse(m *domain.PatientMeasurement) PatientMeasurementResponse {
	recID := ""
	if m.RecordedByID != nil {
		recID = *m.RecordedByID
	}
	return PatientMeasurementResponse{
		ID:                     m.ID,
		PatientID:              m.PatientID,
		WeightKg:               m.WeightKg,
		HeightCm:               m.HeightCm,
		BMI:                    m.BMI,
		BloodPressureSystolic:  m.BloodPressureSystolic,
		BloodPressureDiastolic: m.BloodPressureDiastolic,
		BloodSugar:             m.BloodSugar,
		WaistCircumferenceCm:   m.WaistCircumferenceCm,
		DailyCalorieTarget:     m.DailyCalorieTarget,
		Notes:                  m.Notes,
		RecordedByID:           &recID,
		RecordedByName:         m.RecordedByName,
		RecordedByRole:         m.RecordedByRole,
		MeasuredAt:             m.MeasuredAt.Format(time.RFC3339),
		CreatedAt:              m.CreatedAt.Format(time.RFC3339),
	}
}

func (s *patientService) CreateMeasurement(ctx context.Context, patientID string, req CreateMeasurementRequest, recordedByID, recordedByName, recordedByRole string) (*PatientMeasurementResponse, error) {
	patient, err := s.repo.FindByID(ctx, patientID)
	if err != nil {
		return nil, err
	}

	measuredAt := time.Now()
	if req.MeasuredAt != nil && *req.MeasuredAt != "" {
		if t, err := time.Parse(time.RFC3339, *req.MeasuredAt); err == nil {
			measuredAt = t
		} else if t, err := time.Parse("2006-01-02T15:04", *req.MeasuredAt); err == nil {
			measuredAt = t
		} else if t, err := time.Parse("2006-01-02", *req.MeasuredAt); err == nil {
			measuredAt = t
		}
	}

	var bmi *float64
	weight := req.WeightKg
	if weight == nil || *weight <= 0 {
		w := patient.WeightKg
		weight = &w
	}
	height := req.HeightCm
	if height == nil || *height <= 0 {
		h := patient.HeightCm
		height = &h
	}
	if weight != nil && *weight > 0 && height != nil && *height > 0 {
		hM := *height / 100.0
		val := math.Round((*weight/(hM*hM))*10) / 10
		bmi = &val
	}

	calTarget := req.DailyCalorieTarget
	if calTarget == nil || *calTarget <= 0 {
		c := patient.DailyCalorieTarget
		calTarget = &c
	}

	if calTarget != nil && *calTarget > 0 {
		patient.DailyCalorieTarget = *calTarget
	}
	if req.WeightKg != nil && *req.WeightKg > 0 {
		patient.WeightKg = *req.WeightKg
	}
	if req.HeightCm != nil && *req.HeightCm > 0 {
		patient.HeightCm = *req.HeightCm
	}
	if req.PhysicalActivityLevel != "" {
		patient.PhysicalActivityLevel = req.PhysicalActivityLevel
	}
	if req.BloodType != "" {
		patient.BloodType = domain.BloodType(req.BloodType)
	}
	recRole := recordedByRole
	if recRole == "" {
		recRole = "admin"
	}
	recName := recordedByName
	if recName == "" {
		recName = "Admin"
	}

	m := &domain.PatientMeasurement{
		PatientID:              patient.ID,
		WeightKg:               req.WeightKg,
		HeightCm:               req.HeightCm,
		BMI:                    bmi,
		BloodPressureSystolic:  req.BloodPressureSystolic,
		BloodPressureDiastolic: req.BloodPressureDiastolic,
		BloodSugar:             req.BloodSugar,
		WaistCircumferenceCm:   req.WaistCircumferenceCm,
		DailyCalorieTarget:     calTarget,
		Notes:                  req.Notes,
		RecordedByID:           &recordedByID,
		RecordedByName:         recName,
		RecordedByRole:         recRole,
		MeasuredAt:             measuredAt,
	}

	// Synchronize with global blood_sugar_logs so latest blood sugar & trends
	// update automatically everywhere (created atomically with the measurement).
	var bsLog *domain.BloodSugarLog
	if req.BloodSugar != nil && *req.BloodSugar > 0 {
		bsTimeType := req.BloodSugarTimeType
		if bsTimeType == "" {
			bsTimeType = domain.TimeSewaktu
		}
		medRes := domain.CalculateBloodSugarMedicalResult(*req.BloodSugar, bsTimeType, &patient.DateOfBirth)
		bsLog = &domain.BloodSugarLog{
			BaseModel: domain.BaseModel{
				ID: m.ID,
			},
			PatientID:           patient.ID,
			GlucoseValue:        *req.BloodSugar,
			MeasurementTimeType: bsTimeType,
			MeasuredAt:          measuredAt,
			Category:            medRes.Category,
			Severity:            medRes.Severity,
			ReferenceMin:        medRes.ReferenceMin,
			ReferenceMax:        medRes.ReferenceMax,
			ReferenceRange:      medRes.ReferenceRange,
			Recommendation:      medRes.Recommendation,
			Color:               medRes.Color,
		}
	}

	// Patient profile + measurement + blood sugar log are written in one
	// transaction; a failure rolls all of them back.
	if err := s.repo.CreateMeasurementWithSync(ctx, patient, m, bsLog); err != nil {
		return nil, err
	}

	resp := ToPatientMeasurementResponse(m)
	return &resp, nil
}

func (s *patientService) GetPatientMeasurements(ctx context.Context, patientID string) ([]PatientMeasurementResponse, error) {
	items, err := s.repo.GetPatientMeasurements(ctx, patientID)
	if err != nil {
		return nil, err
	}
	resp := make([]PatientMeasurementResponse, len(items))
	for i := range items {
		resp[i] = ToPatientMeasurementResponse(&items[i])
	}

	// Always ensure the initial baseline registration record (Pengukuran Awal / Minggu 0)
	// is included at the end of the history timeline so account creation track record is NEVER lost!
	patient, pErr := s.repo.FindByID(ctx, patientID)
	if pErr == nil && patient != nil {
		hasBaseline := false
		for _, item := range resp {
			if strings.HasPrefix(item.ID, "baseline-") || item.Notes == "Pengukuran Awal (Registrasi Akun Pasien)" {
				hasBaseline = true
				break
			}
		}

		if !hasBaseline {
			initWeight := patient.WeightKg
			initHeight := patient.HeightCm
			initCalTarget := patient.DailyCalorieTarget

			if len(items) > 0 {
				oldest := items[len(items)-1]
				if oldest.WeightKg != nil && *oldest.WeightKg > 0 {
					initWeight = *oldest.WeightKg
				}
				if oldest.HeightCm != nil && *oldest.HeightCm > 0 {
					initHeight = *oldest.HeightCm
				}
				if oldest.DailyCalorieTarget != nil && *oldest.DailyCalorieTarget > 0 {
					initCalTarget = *oldest.DailyCalorieTarget
				}
			}

			var bmi *float64
			if initWeight > 0 && initHeight > 0 {
				hM := initHeight / 100.0
				val := math.Round((initWeight/(hM*hM))*10) / 10
				bmi = &val
			}
			recID := patient.ID

			baseline := PatientMeasurementResponse{
				ID:                 "baseline-" + patient.ID,
				PatientID:          patient.ID,
				WeightKg:           &initWeight,
				HeightCm:           &initHeight,
				BMI:                bmi,
				DailyCalorieTarget: &initCalTarget,
				Notes:              "Pengukuran Awal (Registrasi Akun Pasien)",
				RecordedByID:       &recID,
				RecordedByName:     "Sistem (Registrasi Awal)",
				RecordedByRole:     "admin",
				MeasuredAt:         patient.CreatedAt.Format(time.RFC3339),
				CreatedAt:          patient.CreatedAt.Format(time.RFC3339),
			}
			resp = append(resp, baseline)
		}
	}

	return resp, nil
}

func (s *patientService) UpdateMeasurement(ctx context.Context, patientID, measurementID string, req UpdateMeasurementRequest) (*PatientMeasurementResponse, error) {
	m, err := s.repo.FindMeasurementByID(ctx, measurementID)
	if err != nil {
		return nil, err
	}
	if m.PatientID != patientID {
		return nil, errs.NewForbidden("measurement record does not belong to this patient")
	}

	if req.WeightKg != nil {
		m.WeightKg = req.WeightKg
	}
	if req.HeightCm != nil {
		m.HeightCm = req.HeightCm
	}
	if req.BloodPressureSystolic != nil {
		m.BloodPressureSystolic = req.BloodPressureSystolic
	}
	if req.BloodPressureDiastolic != nil {
		m.BloodPressureDiastolic = req.BloodPressureDiastolic
	}
	if req.BloodSugar != nil {
		m.BloodSugar = req.BloodSugar
	}
	if req.WaistCircumferenceCm != nil {
		m.WaistCircumferenceCm = req.WaistCircumferenceCm
	}
	if req.DailyCalorieTarget != nil {
		m.DailyCalorieTarget = req.DailyCalorieTarget
	}
	if req.Notes != "" {
		m.Notes = req.Notes
	}

	if m.WeightKg != nil && *m.WeightKg > 0 && m.HeightCm != nil && *m.HeightCm > 0 {
		hM := *m.HeightCm / 100.0
		val := math.Round((*m.WeightKg/(hM*hM))*10) / 10
		m.BMI = &val
	}

	if err := s.repo.UpdateMeasurement(ctx, m); err != nil {
		return nil, err
	}

	resp := ToPatientMeasurementResponse(m)
	return &resp, nil
}

func (s *patientService) UpdatePatientByAdmin(ctx context.Context, patientID string, req UpdatePatientRequest) (*PatientDetailResponse, error) {
	patient, err := s.repo.FindByID(ctx, patientID)
	if err != nil {
		return nil, err
	}

	if req.FullName != "" {
		patient.FullName = req.FullName
	}
	if req.WhatsappNumber != "" {
		patient.WhatsappNumber = req.WhatsappNumber
	}
	if req.Gender != "" {
		patient.Gender = domain.Gender(req.Gender)
	}
	if req.DateOfBirth != "" {
		if dob, err := ParseDOB(req.DateOfBirth); err == nil {
			patient.DateOfBirth = dob
		}
	}
	if req.Address != "" {
		patient.Address = req.Address
	}
	if req.DiabetesType != "" {
		patient.DiabetesType = req.DiabetesType
	}
	if req.BPJS != "" {
		patient.BPJS = req.BPJS
	}
	if req.NIK != "" {
		patient.NIK = req.NIK
	}
	if req.EmergencyName != "" {
		patient.EmergencyName = req.EmergencyName
	}
	if req.EmergencyRelation != "" {
		patient.EmergencyRelation = req.EmergencyRelation
	}
	if req.EmergencyPhone != "" {
		patient.EmergencyPhone = req.EmergencyPhone
	}
	if req.HeightCm != nil && *req.HeightCm > 0 {
		patient.HeightCm = *req.HeightCm
	}
	if req.WeightKg != nil && *req.WeightKg > 0 {
		patient.WeightKg = *req.WeightKg
	}
	if req.DailyCalorieTarget != nil && *req.DailyCalorieTarget > 0 {
		patient.DailyCalorieTarget = *req.DailyCalorieTarget
	}
	if req.DiagnosisDate != "" {
		if diag, err := ParseDOB(req.DiagnosisDate); err == nil {
			patient.DiagnosisDate = &diag
		}
	}
	if req.CurrentMedication != "" {
		patient.CurrentMedication = req.CurrentMedication
	}
	if req.Allergies != "" {
		patient.Allergies = req.Allergies
	}
	if req.SmokingStatus != "" {
		patient.SmokingStatus = req.SmokingStatus
	}
	if req.PhysicalActivityLevel != "" {
		patient.PhysicalActivityLevel = req.PhysicalActivityLevel
	}

	if err := s.repo.Update(ctx, patient); err != nil {
		return nil, err
	}

	return s.GetPatient(ctx, patientID)
}

// Helpers
func strPtr(s string) *string {
	return &s
}

func CalculateDSMESCalorieTarget(gender string, weightKg, heightCm float64, age int) int {
	if weightKg <= 0 {
		weightKg = 60
	}
	if heightCm <= 0 {
		heightCm = 160
	}
	if age <= 0 {
		age = 40
	}

	isMale := strings.ToLower(gender) == "laki-laki" || strings.ToLower(gender) == "laki_laki" || strings.ToLower(gender) == "male"
	var bmr float64
	if isMale {
		bmr = 66.5 + (13.75 * weightKg) + (5.003 * heightCm) - (6.75 * float64(age))
	} else {
		bmr = 655.1 + (9.563 * weightKg) + (1.850 * heightCm) - (4.676 * float64(age))
	}

	if bmr < 800 {
		if isMale {
			bmr = 30 * weightKg
		} else {
			bmr = 25 * weightKg
		}
	}

	return int(math.Round(bmr))
}
