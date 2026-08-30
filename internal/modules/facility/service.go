package facility

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type facilityService struct {
	repo FacilityRepository
	log  *zap.Logger
}

func NewFacilityService(repo FacilityRepository, log *zap.Logger) FacilityService {
	return &facilityService{repo: repo, log: log}
}

func (s *facilityService) ListFacilities(ctx context.Context, search string, page, limit int) ([]FacilityResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	items, total, err := s.repo.FindAll(ctx, search, page, limit)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]FacilityResponse, len(items))
	for i := range items {
		resp[i] = ToFacilityResponse(&items[i])
	}

	return resp, total, nil
}

func (s *facilityService) GetFacility(ctx context.Context, id string) (*FacilityResponse, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := ToFacilityResponse(item)
	return &res, nil
}

func (s *facilityService) CreateFacility(ctx context.Context, req CreateFacilityRequest) (*FacilityResponse, error) {
	name := strings.TrimSpace(req.Name)
	if _, err := s.repo.FindByName(ctx, name); err == nil {
		return nil, errs.NewConflict("puskesmas sudah terdaftar")
	}

	item := &domain.HealthFacility{
		Name:     name,
		Address:  strings.TrimSpace(req.Address),
		IsActive: true,
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}

	res := ToFacilityResponse(item)
	return &res, nil
}

func (s *facilityService) UpdateFacility(ctx context.Context, id string, req UpdateFacilityRequest) (*FacilityResponse, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if !strings.EqualFold(item.Name, name) {
		if _, err := s.repo.FindByName(ctx, name); err == nil {
			return nil, errs.NewConflict("puskesmas sudah terdaftar")
		}
	}

	item.Name = name
	item.Address = strings.TrimSpace(req.Address)

	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}

	res := ToFacilityResponse(item)
	return &res, nil
}

func (s *facilityService) DeleteFacility(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
