package facility

import (
	"context"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type FacilityRepository interface {
	FindAll(ctx context.Context, search string, page, limit int) ([]domain.HealthFacility, int64, error)
	FindByID(ctx context.Context, id string) (*domain.HealthFacility, error)
	FindByName(ctx context.Context, name string) (*domain.HealthFacility, error)
	Create(ctx context.Context, f *domain.HealthFacility) error
	Update(ctx context.Context, f *domain.HealthFacility) error
	Delete(ctx context.Context, id string) error
}

type FacilityService interface {
	ListFacilities(ctx context.Context, search string, page, limit int) ([]FacilityResponse, int64, error)
	GetFacility(ctx context.Context, id string) (*FacilityResponse, error)
	CreateFacility(ctx context.Context, req CreateFacilityRequest) (*FacilityResponse, error)
	UpdateFacility(ctx context.Context, id string, req UpdateFacilityRequest) (*FacilityResponse, error)
	DeleteFacility(ctx context.Context, id string) error
}
