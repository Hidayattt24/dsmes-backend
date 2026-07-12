package patient

import (
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type RegisterPatientRequest struct {
	Email          string          `json:"email"           validate:"required,email"`
	Password       string          `json:"password"        validate:"required,min=6"`
	FullName       string          `json:"full_name"       validate:"required,min=3,max=150"`
	Nickname       string          `json:"nickname"`
	WhatsappNumber string          `json:"whatsapp_number" validate:"required,numeric,min=10,max=20"`
	Gender         domain.Gender   `json:"gender"          validate:"required,oneof=laki_laki perempuan"`
	DateOfBirth    string          `json:"date_of_birth"   validate:"required"` // format: "YYYY-MM-DD"
	HeightCm       float64         `json:"height_cm"       validate:"required,gt=0"`
	WeightKg       float64         `json:"weight_kg"       validate:"required,gt=0"`
	BloodType      domain.BloodType `json:"blood_type"      validate:"required,oneof=A B AB O tidak_tahu"`
}

type UpdatePatientProfileRequest struct {
	FullName          string  `json:"full_name"       validate:"required,min=3,max=150"`
	Nickname          string  `json:"nickname"`
	WhatsappNumber    string  `json:"whatsapp_number" validate:"required,numeric,min=10,max=20"`
	HeightCm          float64 `json:"height_cm"       validate:"required,gt=0"`
	WeightKg          float64 `json:"weight_kg"       validate:"required,gt=0"`
	ProfilePhotoURL   string  `json:"profile_photo_url"`
	BPJS              string  `json:"bpjs"`
	NIK               string  `json:"nik"`
	EmergencyName     string  `json:"emergency_name"`
	EmergencyRelation string  `json:"emergency_relation"`
	EmergencyPhone    string  `json:"emergency_phone"`
	DiabetesType      string  `json:"diabetes_type"`
	InterventionType  string  `json:"intervention_type"`
}

type AssignStaffRequest struct {
	StaffID string `json:"staff_id" validate:"required,uuid4"`
}

type PatientResponse struct {
	ID                 string               `json:"id"`
	Email              string               `json:"email"`
	FullName           string               `json:"full_name"`
	Nickname           string               `json:"nickname"`
	WhatsappNumber     string               `json:"whatsapp_number"`
	Gender             domain.Gender        `json:"gender"`
	DateOfBirth        string               `json:"date_of_birth"`
	HeightCm           float64              `json:"height_cm"`
	WeightKg           float64              `json:"weight_kg"`
	BloodType          domain.BloodType     `json:"blood_type"`
	DailyCalorieTarget int                  `json:"daily_calorie_target"`
	MedicalStatus      string               `json:"medical_status"`
	ProfilePhotoURL    string               `json:"profile_photo_url"`
	Status             domain.AccountStatus `json:"status"`
	CreatedAt          string               `json:"created_at"`
	BPJS               string               `json:"bpjs"`
	NIK                string               `json:"nik"`
	EmergencyName      string               `json:"emergency_name"`
	EmergencyRelation  string               `json:"emergency_relation"`
	EmergencyPhone     string               `json:"emergency_phone"`
	DiabetesType       string               `json:"diabetes_type"`
	Compliance         int                  `json:"compliance"`
	InterventionType   string               `json:"intervention_type"`
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
	return PatientResponse{
		ID:                 p.ID,
		Email:              p.Email,
		FullName:           p.FullName,
		Nickname:           p.Nickname,
		WhatsappNumber:     p.WhatsappNumber,
		Gender:             p.Gender,
		DateOfBirth:        p.DateOfBirth.Format("2006-01-02"),
		HeightCm:           p.HeightCm,
		WeightKg:           p.WeightKg,
		BloodType:          p.BloodType,
		DailyCalorieTarget: p.DailyCalorieTarget,
		MedicalStatus:      p.MedicalStatus,
		ProfilePhotoURL:    p.ProfilePhotoURL,
		Status:             p.Status,
		CreatedAt:          p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		BPJS:               p.BPJS,
		NIK:                p.NIK,
		EmergencyName:      p.EmergencyName,
		EmergencyRelation:  p.EmergencyRelation,
		EmergencyPhone:     p.EmergencyPhone,
		DiabetesType:       p.DiabetesType,
		Compliance:         p.Compliance,
		InterventionType:   p.InterventionType,
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
	StaffID string
	Search      string
	Gender      string
	Status      string
	Page        int
	Limit       int
}

type PatientStats struct {
	TotalPatients  int64 `json:"total_patients"`
	ActivePatients int64 `json:"active_patients"`
	AverageAge     int   `json:"average_age"`
}

// ParseDOB parses date of birth string "YYYY-MM-DD"
func ParseDOB(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}
