package patient

import (
	"math"
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/modules/nutrition"
)

type RegisterPatientRequest struct {
	Email                 string  `json:"email"           validate:"omitempty,email"`
	Password              string  `json:"password"        validate:"required,min=8"`
	FullName              string  `json:"full_name"       validate:"required,min=3,max=150"`
	Nickname              string  `json:"nickname"`
	WhatsappNumber        string  `json:"whatsapp_number"`
	PhoneNumber           string  `json:"phone_number"`
	Gender                string  `json:"gender"`
	DateOfBirth           string  `json:"date_of_birth"`
	HeightCm              float64 `json:"height_cm"`
	WeightKg              float64 `json:"weight_kg"`
	BloodType             string  `json:"blood_type"`
	ActivityLevel         string  `json:"activity_level"`
	PhysicalActivityLevel string  `json:"physical_activity_level"`
}

func (r *RegisterPatientRequest) GetPhone() string {
	if r.PhoneNumber != "" {
		return r.PhoneNumber
	}
	return r.WhatsappNumber
}

func (r *RegisterPatientRequest) GetActivity() string {
	if r.ActivityLevel != "" {
		return r.ActivityLevel
	}
	return r.PhysicalActivityLevel
}

type SetupHealthProfileRequest struct {
	Gender                string  `json:"gender"        validate:"required"`
	DateOfBirth           string  `json:"date_of_birth" validate:"required"` // format: "YYYY-MM-DD"
	HeightCm              float64 `json:"height_cm"     validate:"required,gt=0"`
	WeightKg              float64 `json:"weight_kg"     validate:"required,gt=0"`
	BloodType             string  `json:"blood_type"    validate:"required"`
	ActivityLevel         string  `json:"activity_level"`
	PhysicalActivityLevel string  `json:"physical_activity_level"`
}

func (r *SetupHealthProfileRequest) GetActivity() string {
	if r.ActivityLevel != "" {
		return r.ActivityLevel
	}
	return r.PhysicalActivityLevel
}

type UpdatePatientProfileRequest struct {
	FullName              string  `json:"full_name"       validate:"required,min=3,max=150"`
	Nickname              string  `json:"nickname"`
	WhatsappNumber        string  `json:"whatsapp_number" validate:"required,numeric,min=10,max=20"`
	HeightCm              float64 `json:"height_cm"       validate:"required,gt=0"`
	WeightKg              float64 `json:"weight_kg"       validate:"required,gt=0"`
	Gender                string  `json:"gender"`
	DateOfBirth           string  `json:"date_of_birth"`
	BloodType             string  `json:"blood_type"`
	ProfilePhotoURL       string  `json:"profile_photo_url"`
	BPJS                  string  `json:"bpjs"`
	NIK                   string  `json:"nik"`
	EmergencyName         string  `json:"emergency_name"`
	EmergencyRelation     string  `json:"emergency_relation"`
	EmergencyPhone        string  `json:"emergency_phone"`
	DiabetesType          string  `json:"diabetes_type"`
	InterventionType      string  `json:"intervention_type"`
	PatientCode           string  `json:"patient_code"`
	Address               string  `json:"address"`
	DiagnosisDate         string  `json:"diagnosis_date"`
	CurrentMedication     string  `json:"current_medication"`
	Allergies             string  `json:"allergies"`
	SmokingStatus         string  `json:"smoking_status"`
	PhysicalActivityLevel string  `json:"physical_activity_level"`
}

type AssignStaffRequest struct {
	StaffID string `json:"staff_id" validate:"required,uuid4"`
}

type PatientResponse struct {
	ID                    string                            `json:"id"`
	PhoneNumber           string                            `json:"phone_number"`
	Email                 string                            `json:"email"`
	FullName              string                            `json:"full_name"`
	Nickname              string                            `json:"nickname"`
	WhatsappNumber        string                            `json:"whatsapp_number"`
	Gender                domain.Gender                     `json:"gender"`
	DateOfBirth           string                            `json:"date_of_birth"`
	HeightCm              float64                           `json:"height_cm"`
	WeightKg              float64                           `json:"weight_kg"`
	BloodType             domain.BloodType                  `json:"blood_type"`
	DailyCalorieTarget    int                               `json:"daily_calorie_target"`
	Recommendations       *nutrition.CalorieRecommendations `json:"recommendations,omitempty"`
	MedicalStatus         string                            `json:"medical_status"`
	ProfilePhotoURL       string                            `json:"profile_photo_url"`
	Status                domain.AccountStatus              `json:"status"`
	CreatedAt             string                            `json:"created_at"`
	BPJS                  string                            `json:"bpjs"`
	NIK                   string                            `json:"nik"`
	EmergencyName         string                            `json:"emergency_name"`
	EmergencyRelation     string                            `json:"emergency_relation"`
	EmergencyPhone        string                            `json:"emergency_phone"`
	DiabetesType          string                            `json:"diabetes_type"`
	Compliance            int                               `json:"compliance"`
	ComplianceLabel       string                            `json:"compliance_label,omitempty"`
	ComplianceBreakdown   *ComplianceBreakdown              `json:"compliance_breakdown,omitempty"`
	InterventionType      string                            `json:"intervention_type"`
	PatientCode           string                            `json:"patient_code"`
	Address               string                            `json:"address"`
	DiagnosisDate         string                            `json:"diagnosis_date"`
	CurrentMedication     string                            `json:"current_medication"`
	Allergies             string                            `json:"allergies"`
	SmokingStatus         string                            `json:"smoking_status"`
	PhysicalActivityLevel string                            `json:"physical_activity_level"`
	LastActiveAt          *string                           `json:"last_active_at,omitempty"`

	// Summary statistics fields
	LatestBloodSugar       *int               `json:"latest_blood_sugar,omitempty"`
	LatestBloodSugarTime   *string            `json:"latest_blood_sugar_time,omitempty"`
	LatestBloodSugarStatus *string            `json:"latest_blood_sugar_status,omitempty"`
	AverageBloodSugar      *float64           `json:"average_blood_sugar,omitempty"`
	LatestWeight           *float64           `json:"latest_weight,omitempty"`
	BMI                    *float64           `json:"bmi,omitempty"`
	BMICategory            *string            `json:"bmi_category,omitempty"`
	LatestMealCalories     *float64           `json:"latest_meal_calories,omitempty"`
	LatestMealType         *string            `json:"latest_meal_type,omitempty"`
	LatestMealName         *string            `json:"latest_meal_name,omitempty"`
	LatestActivityTime     *string            `json:"latest_activity_time,omitempty"`
	LatestActivityName     *string            `json:"latest_activity_name,omitempty"`
	WaistCircumferenceCm   *float64           `json:"waist_circumference_cm,omitempty"`
	CalorieStatus          *CalorieStatusInfo `json:"calorie_status_info,omitempty"`
}

type CalorieStatusInfo struct {
	TargetCalories        int     `json:"target_calories"`
	ConsumedCalories      float64 `json:"consumed_calories"`
	AchievementPercentage float64 `json:"achievement_percentage"`
	CalorieDifference     float64 `json:"calorie_difference"`
	CalorieDifferenceStr  string  `json:"calorie_difference_str"`
	CalorieStatus         string  `json:"calorie_status"`
	CalorieStatusCode     string  `json:"calorie_status_code"`
	CalorieDescription    string  `json:"calorie_description"`
}

type ComplianceBreakdown struct {
	BloodSugarScore float64 `json:"blood_sugar_score"`
	FoodScore       float64 `json:"food_score"`
	ActivityScore   float64 `json:"activity_score"`
	MedicationScore float64 `json:"medication_score"`
}

type ActivityAnalyticsItem struct {
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

type DailyLogsAggregate struct {
	BloodSugarCount          int
	TotalMealCalories        float64
	MealCount                int
	TotalActivityMinutes     int
	MedicationCompletedCount int
	MedicationScheduledCount int
}

type PatientActivityAnalyticsResponse struct {
	TotalRecords     int64                 `json:"total_records"`
	BloodSugar       ActivityAnalyticsItem `json:"blood_sugar"`
	Food             ActivityAnalyticsItem `json:"food"`
	PhysicalActivity ActivityAnalyticsItem `json:"physical_activity"`
	Medication       ActivityAnalyticsItem `json:"medication"`
	MostUsed         string                `json:"most_used"`
	LeastUsed        string                `json:"least_used"`
}

type PatientSummaryData struct {
	LatestBloodSugar       *int       `json:"latest_blood_sugar,omitempty"`
	LatestBloodSugarTime   *time.Time `json:"latest_blood_sugar_time,omitempty"`
	LatestBloodSugarStatus *string    `json:"latest_blood_sugar_status,omitempty"`
	AverageBloodSugar      *float64   `json:"average_blood_sugar,omitempty"`
	LatestWeight           *float64   `json:"latest_weight,omitempty"`
	BMI                    *float64   `json:"bmi,omitempty"`
	LatestMealCalories     *float64   `json:"latest_meal_calories,omitempty"`
	LatestMealType         *string    `json:"latest_meal_type,omitempty"`
	LatestMealName         *string    `json:"latest_meal_name,omitempty"`
	LatestActivityTime     *time.Time `json:"latest_activity_time_raw,omitempty"`
	LatestActivityName     *string    `json:"latest_activity_name,omitempty"`
	TodayConsumedCalories  *float64   `json:"today_consumed_calories,omitempty"`
}

type PatientDetailResponse struct {
	PatientResponse
	AssignedStaff     *StaffInfo                   `json:"assigned_staff,omitempty"`
	LatestMeasurement *PatientMeasurementResponse  `json:"latest_measurement,omitempty"`
	Measurements      []PatientMeasurementResponse `json:"measurements,omitempty"`
}

type StaffInfo struct {
	ID            string `json:"id"`
	FullName      string `json:"full_name"`
	PositionTitle string `json:"position_title"`
}

func ToPatientResponse(p *domain.Patient) PatientResponse {
	diagDate := ""
	if p.DiagnosisDate != nil {
		diagDate = p.DiagnosisDate.Format("2006-01-02")
	}

	var lastActiveAt *string
	if p.LastActiveAt != nil {
		t := p.LastActiveAt.Format("2006-01-02T15:04:05Z07:00")
		lastActiveAt = &t
	}

	var bmiVal *float64
	var bmiCat *string
	if p.HeightCm > 0 && p.WeightKg > 0 {
		hM := p.HeightCm / 100.0
		b := p.WeightKg / (hM * hM)
		b = math.Round(b*10) / 10
		bmiVal = &b

		var cat string
		if b < 18.5 {
			cat = "Kurus"
		} else if b >= 18.5 && b < 23.0 {
			cat = "Normal"
		} else if b >= 23.0 && b < 25.0 {
			cat = "Kelebihan Berat Badan"
		} else {
			cat = "Obesitas"
		}
		bmiCat = &cat
	}

	resp := PatientResponse{
		ID:                    p.ID,
		PhoneNumber:           p.PhoneNumber,
		Email:                 p.GetEmail(),
		FullName:              p.FullName,
		Nickname:              p.Nickname,
		WhatsappNumber:        p.WhatsappNumber,
		Gender:                p.Gender,
		DateOfBirth:           p.DateOfBirth.Format("2006-01-02"),
		HeightCm:              p.HeightCm,
		WeightKg:              p.WeightKg,
		BloodType:             p.BloodType,
		DailyCalorieTarget:    p.DailyCalorieTarget,
		MedicalStatus:         p.MedicalStatus,
		ProfilePhotoURL:       p.ProfilePhotoURL,
		Status:                p.Status,
		CreatedAt:             p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		BPJS:                  p.BPJS,
		NIK:                   p.NIK,
		EmergencyName:         p.EmergencyName,
		EmergencyRelation:     p.EmergencyRelation,
		EmergencyPhone:        p.EmergencyPhone,
		DiabetesType:          p.DiabetesType,
		Compliance:            p.Compliance,
		InterventionType:      p.InterventionType,
		PatientCode:           p.PatientCode,
		Address:               p.Address,
		DiagnosisDate:         diagDate,
		CurrentMedication:     p.CurrentMedication,
		Allergies:             p.Allergies,
		SmokingStatus:         p.SmokingStatus,
		PhysicalActivityLevel: p.PhysicalActivityLevel,
		LastActiveAt:          lastActiveAt,
		BMI:                   bmiVal,
		BMICategory:           bmiCat,
	}

	if p.MaintenanceCalories > 0 {
		resp.Recommendations = &nutrition.CalorieRecommendations{
			Maintain: nutrition.CalorieRecommendationDetail{
				Title:      "Pertahankan Berat Badan",
				Calories:   p.MaintenanceCalories,
				Percentage: p.MaintenancePercentage,
			},
			MildLoss: nutrition.CalorieRecommendationDetail{
				Title:        "Penurunan Berat Ringan",
				WeeklyTarget: "0,25 kg/minggu",
				Calories:     p.MildWeightLossCalories,
				Percentage:   p.MildPercentage,
			},
			WeightLoss: nutrition.CalorieRecommendationDetail{
				Title:        "Turunkan Berat Badan",
				WeeklyTarget: "0,5 kg/minggu",
				Calories:     p.WeightLossCalories,
				Percentage:   p.WeightLossPercentage,
			},
			ExtremeLoss: nutrition.CalorieRecommendationDetail{
				Title:        "Penurunan Berat Intensif",
				WeeklyTarget: "1 kg/minggu",
				Calories:     p.ExtremeWeightLossCalories,
				Percentage:   p.ExtremePercentage,
			},
		}
	}

	return resp
}

func ToPatientDetailResponse(p *domain.Patient) PatientDetailResponse {
	detail := PatientDetailResponse{
		PatientResponse: ToPatientResponse(p),
	}
	if p.AssignedStaff != nil {
		detail.AssignedStaff = &StaffInfo{
			ID:            p.AssignedStaff.ID,
			FullName:      p.AssignedStaff.FullName,
			PositionTitle: p.AssignedStaff.PositionTitle,
		}
	}
	return detail
}

type PatientFilterQuery struct {
	StaffID          string
	Search           string
	Gender           string
	Status           string
	Page             int
	Limit            int
	SortBy           string // name, newest, oldest, latest_record, highest_blood_sugar
	SortOrder        string // asc, desc
	BloodSugarStatus string // normal, tinggi, rendah, sangat_tinggi
	RiskLevel        string // rendah, sedang, tinggi, sangat_tinggi
	ComplianceMin    *int
	ComplianceMax    *int
	AgeMin           *int
	AgeMax           *int
}

type PatientStats struct {
	TotalPatients  int64 `json:"total_patients"`
	ActivePatients int64 `json:"active_patients"`
	YoungestAge    int   `json:"youngest_age"`
	OldestAge      int   `json:"oldest_age"`
}

// ParseDOB parses date of birth string ("YYYY-MM-DD" or ISO 8601)
func ParseDOB(dateStr string) (time.Time, error) {
	if len(dateStr) >= 10 {
		dateStrOnly := dateStr[:10]
		if t, err := time.Parse("2006-01-02", dateStrOnly); err == nil {
			return t, nil
		}
	}
	return time.Parse(time.RFC3339, dateStr)
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password"     validate:"required,min=8"`
}

type CreateMeasurementRequest struct {
	WeightKg               *float64               `json:"weight_kg"`
	HeightCm               *float64               `json:"height_cm"`
	BloodPressureSystolic  *int                   `json:"blood_pressure_systolic"`
	BloodPressureDiastolic *int                   `json:"blood_pressure_diastolic"`
	BloodSugar             *int                   `json:"blood_sugar"`
	BloodSugarTimeType     domain.MeasurementTime `json:"blood_sugar_time_type"`
	WaistCircumferenceCm   *float64               `json:"waist_circumference_cm"`
	DailyCalorieTarget     *int                   `json:"daily_calorie_target"`
	Gender                 string                 `json:"gender"`
	BloodType              string                 `json:"blood_type"`
	PhysicalActivityLevel  string                 `json:"physical_activity_level"`
	Notes                  string                 `json:"notes"`
	MeasuredAt             *string                `json:"measured_at"`
}

type UpdateMeasurementRequest struct {
	WeightKg               *float64 `json:"weight_kg"`
	HeightCm               *float64 `json:"height_cm"`
	BloodPressureSystolic  *int     `json:"blood_pressure_systolic"`
	BloodPressureDiastolic *int     `json:"blood_pressure_diastolic"`
	BloodSugar             *int     `json:"blood_sugar"`
	WaistCircumferenceCm   *float64 `json:"waist_circumference_cm"`
	DailyCalorieTarget     *int     `json:"daily_calorie_target"`
	Notes                  string   `json:"notes"`
}

type UpdatePatientRequest struct {
	FullName              string   `json:"full_name"`
	WhatsappNumber        string   `json:"whatsapp_number"`
	Gender                string   `json:"gender"`
	DateOfBirth           string   `json:"date_of_birth"`
	Address               string   `json:"address"`
	DiabetesType          string   `json:"diabetes_type"`
	BPJS                  string   `json:"bpjs"`
	NIK                   string   `json:"nik"`
	EmergencyName         string   `json:"emergency_name"`
	EmergencyRelation     string   `json:"emergency_relation"`
	EmergencyPhone        string   `json:"emergency_phone"`
	HeightCm              *float64 `json:"height_cm"`
	WeightKg              *float64 `json:"weight_kg"`
	DailyCalorieTarget    *int     `json:"daily_calorie_target"`
	DiagnosisDate         string   `json:"diagnosis_date"`
	CurrentMedication     string   `json:"current_medication"`
	Allergies             string   `json:"allergies"`
	SmokingStatus         string   `json:"smoking_status"`
	PhysicalActivityLevel string   `json:"physical_activity_level"`
}

type PatientMeasurementResponse struct {
	ID                     string   `json:"id"`
	PatientID              string   `json:"patient_id"`
	WeightKg               *float64 `json:"weight_kg,omitempty"`
	HeightCm               *float64 `json:"height_cm,omitempty"`
	BMI                    *float64 `json:"bmi,omitempty"`
	BloodPressureSystolic  *int     `json:"blood_pressure_systolic,omitempty"`
	BloodPressureDiastolic *int     `json:"blood_pressure_diastolic,omitempty"`
	BloodSugar             *int     `json:"blood_sugar,omitempty"`
	WaistCircumferenceCm   *float64 `json:"waist_circumference_cm,omitempty"`
	DailyCalorieTarget     *int     `json:"daily_calorie_target,omitempty"`
	Notes                  string   `json:"notes,omitempty"`
	RecordedByID           *string  `json:"recorded_by_id,omitempty"`
	RecordedByName         string   `json:"recorded_by_name"`
	RecordedByRole         string   `json:"recorded_by_role"`
	MeasuredAt             string   `json:"measured_at"`
	CreatedAt              string   `json:"created_at"`
}
