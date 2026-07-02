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
	FullName        string  `json:"full_name"       validate:"required,min=3,max=150"`
	Nickname        string  `json:"nickname"`
	WhatsappNumber  string  `json:"whatsapp_number" validate:"required,numeric,min=10,max=20"`
	HeightCm        float64 `json:"height_cm"       validate:"required,gt=0"`
	WeightKg        float64 `json:"weight_kg"       validate:"required,gt=0"`
	ProfilePhotoURL string  `json:"profile_photo_url"`
}

type AssignPuskesmasRequest struct {
	PuskesmasID string `json:"puskesmas_id" validate:"required,uuid4"`
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
}

type PatientDetailResponse struct {
	PatientResponse
	AssignedPuskesmas *PuskesmasInfo `json:"assigned_puskesmas,omitempty"`
}

type PuskesmasInfo struct {
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
	}
}

func ToPatientDetailResponse(p *domain.Patient) PatientDetailResponse {
	detail := PatientDetailResponse{
		PatientResponse: ToPatientResponse(p),
	}
	if p.AssignedPuskesmas != nil {
		detail.AssignedPuskesmas = &PuskesmasInfo{
			ID:            p.AssignedPuskesmas.ID,
			FullName:      p.AssignedPuskesmas.FullName,
			PositionTitle: p.AssignedPuskesmas.PositionTitle,
		}
	}
	return detail
}

type PatientFilterQuery struct {
	PuskesmasID string
	Search      string
	Page        int
	Limit       int
}

// ParseDOB parses date of birth string "YYYY-MM-DD"
func ParseDOB(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}
