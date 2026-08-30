package facility

import (
	"github.com/dsmes/dsmes-backend/internal/domain"
)

type CreateFacilityRequest struct {
	Name    string `json:"name"    validate:"required,min=2,max=150"`
	Address string `json:"address"`
}

type UpdateFacilityRequest struct {
	Name    string `json:"name"    validate:"required,min=2,max=150"`
	Address string `json:"address"`
}

type FacilityResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

func ToFacilityResponse(f *domain.HealthFacility) FacilityResponse {
	return FacilityResponse{
		ID:        f.ID,
		Name:      f.Name,
		Address:   f.Address,
		IsActive:  f.IsActive,
		CreatedAt: f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
