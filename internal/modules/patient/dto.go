package patient

import (
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type RegisterPatientRequest struct {
	Email                 string  `json:"email"                  validate:"required,email"`
	Password              string  `json:"password"               validate:"required,min=8"`
	FullName              string  `json:"full_name"              validate:"required,min=3,max=150"`
	Nickname              string  `json:"nickname"`
	WhatsappNumber        string  `json:"whatsapp_number"`
	PhoneNumber           string  `json:"phone_number"`
	Gender                string  `json:"gender"                 validate:"required"`
	DateOfBirth           string  `json:"date_of_birth"          validate:"required"` // format: "YYYY-MM-DD"
	HeightCm              float64 `json:"height_cm"              validate:"required,gt=0"`
	WeightKg              float64 `json:"weight_kg"              validate:"required,gt=0"`
	BloodType             string  `json:"blood_type"             validate:"required"`
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


type UpdatePatientProfileRequest struct {
	FullName              string  `json:"full_name"       validate:"required,min=3,max=150"`
	Nickname              string  `json:"nickname"`
	WhatsappNumber        string  `json:"whatsapp_number" validate:"required,numeric,min=10,max=20"`
	HeightCm              float64 `json:"height_cm"       validate:"required,gt=0"`
	WeightKg              float64 `json:"weight_kg"       validate:"required,gt=0"`
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
	ID                    string               `json:"id"`
	Email                 string               `json:"email"`
	FullName              string               `json:"full_name"`
	Nickname              string               `json:"nickname"`
	WhatsappNumber        string               `json:"whatsapp_number"`
	Gender                domain.Gender        `json:"gender"`
	DateOfBirth           string               `json:"date_of_birth"`
	HeightCm              float64              `json:"height_cm"`
	WeightKg              float64              `json:"weight_kg"`
	BloodType             domain.BloodType     `json:"blood_type"`
	DailyCalorieTarget    int                  `json:"daily_calorie_target"`
	MedicalStatus         string               `json:"medical_status"`
	ProfilePhotoURL       string               `json:"profile_photo_url"`
	Status                domain.AccountStatus `json:"status"`
	CreatedAt             string               `json:"created_at"`
	BPJS                  string               `json:"bpjs"`
	NIK                   string               `json:"nik"`
	EmergencyName         string               `json:"emergency_name"`
	EmergencyRelation     string               `json:"emergency_relation"`
	EmergencyPhone        string               `json:"emergency_phone"`
	DiabetesType          string               `json:"diabetes_type"`
	Compliance            int                  `json:"compliance"`
	InterventionType      string               `json:"intervention_type"`
	PatientCode           string               `json:"patient_code"`
	Address               string               `json:"address"`
	DiagnosisDate         string               `json:"diagnosis_date"`
	CurrentMedication     string               `json:"current_medication"`
	Allergies             string               `json:"allergies"`
	SmokingStatus         string               `json:"smoking_status"`
	PhysicalActivityLevel string               `json:"physical_activity_level"`
	LastActiveAt          *string              `json:"last_active_at,omitempty"`

	// Summary statistics fields
	LatestBloodSugar       *int     `json:"latest_blood_sugar,omitempty"`
	LatestBloodSugarTime   *string  `json:"latest_blood_sugar_time,omitempty"`
	LatestBloodSugarStatus *string  `json:"latest_blood_sugar_status,omitempty"`
	AverageBloodSugar      *float64 `json:"average_blood_sugar,omitempty"`
	LatestWeight           *float64 `json:"latest_weight,omitempty"`
	BMI                    *float64 `json:"bmi,omitempty"`
	LatestMealCalories     *float64 `json:"latest_meal_calories,omitempty"`
	LatestMealType         *string  `json:"latest_meal_type,omitempty"`
	LatestActivityTime     *string  `json:"latest_activity_time,omitempty"`
	LatestActivityName     *string  `json:"latest_activity_name,omitempty"`
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
	LatestActivityTime     *time.Time `json:"latest_activity_time_raw,omitempty"`
	LatestActivityName     *string    `json:"latest_activity_name,omitempty"`
}

type PatientDetailResponse struct {
	PatientResponse
	AssignedStaff *StaffInfo `json:"assigned_staff,omitempty"`
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

	return PatientResponse{
		ID:                    p.ID,
		Email:                 p.Email,
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
	}
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
	AverageAge     int   `json:"average_age"`
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

