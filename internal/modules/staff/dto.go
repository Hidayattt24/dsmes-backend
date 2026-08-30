package staff

import "github.com/dsmes/dsmes-backend/internal/domain"

type CreateStaffRequest struct {
	FullName         string           `json:"full_name"        validate:"required,min=3,max=150"`
	Username         string           `json:"username"         validate:"required,min=3,max=50"`
	Email            string           `json:"email"            validate:"required,email"`
	Password         string           `json:"password"         validate:"required,min=6"`
	WhatsappNumber   string           `json:"whatsapp_number"  validate:"required,numeric,min=10,max=20"`
	Role             domain.StaffRole `json:"role"             validate:"required,oneof=admin staff"`
	PositionTitle    string           `json:"position_title"`
	ShortBio         string           `json:"short_bio"`
	ProfilePhotoURL  string           `json:"profile_photo_url"`
	HealthFacilityID string           `json:"health_facility_id" validate:"omitempty,uuid4"`
}

type UpdateStaffRequest struct {
	FullName         string `json:"full_name"        validate:"required,min=3,max=150"`
	Username         string `json:"username"         validate:"required,min=3,max=50"`
	Email            string `json:"email"            validate:"required,email"`
	WhatsappNumber   string `json:"whatsapp_number"  validate:"required,numeric,min=10,max=20"`
	PositionTitle    string `json:"position_title"`
	ShortBio         string `json:"short_bio"`
	ProfilePhotoURL  string `json:"profile_photo_url"`
	HealthFacilityID string `json:"health_facility_id" validate:"omitempty,uuid4"`
}

type UpdateProfileRequest struct {
	FullName        string `json:"full_name"       validate:"required,min=3,max=150"`
	Username        string `json:"username"        validate:"required,min=3,max=50"`
	Email           string `json:"email"           validate:"required,email"`
	WhatsappNumber  string `json:"whatsapp_number" validate:"required,numeric,min=10,max=20"`
	PositionTitle   string `json:"position_title"`
	ShortBio        string `json:"short_bio"`
	ProfilePhotoURL string `json:"profile_photo_url"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password"     validate:"required,min=8"`
}

type StaffResponse struct {
	ID                 string               `json:"id"`
	FullName           string               `json:"full_name"`
	Username           string               `json:"username"`
	Email              string               `json:"email"`
	WhatsappNumber     string               `json:"whatsapp_number"`
	Role               domain.StaffRole     `json:"role"`
	Status             domain.AccountStatus `json:"status"`
	PositionTitle      string               `json:"position_title"`
	ShortBio           string               `json:"short_bio"`
	ProfilePhotoURL    string               `json:"profile_photo_url"`
	HealthFacilityID   *string              `json:"health_facility_id,omitempty"`
	HealthFacilityName string               `json:"health_facility_name,omitempty"`
	CreatedAt          string               `json:"created_at"`
}

func ToStaffResponse(s *domain.StaffAccount) StaffResponse {
	resp := StaffResponse{
		ID:               s.ID,
		FullName:         s.FullName,
		Username:         s.Username,
		Email:            s.Email,
		WhatsappNumber:   s.WhatsappNumber,
		Role:             s.Role,
		Status:           s.Status,
		PositionTitle:    s.PositionTitle,
		ShortBio:         s.ShortBio,
		ProfilePhotoURL:  s.ProfilePhotoURL,
		HealthFacilityID: s.HealthFacilityID,
		CreatedAt:        s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if s.HealthFacility != nil {
		resp.HealthFacilityName = s.HealthFacility.Name
	}
	return resp
}
