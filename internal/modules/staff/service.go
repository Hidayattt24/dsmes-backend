package staff

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type staffService struct {
	repo StaffRepository
	log  *zap.Logger
}

func NewStaffService(repo StaffRepository, log *zap.Logger) StaffService {
	return &staffService{repo: repo, log: log}
}

func (s *staffService) ListStaff(ctx context.Context, role *domain.StaffRole, page, limit int) ([]StaffResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	items, total, err := s.repo.FindAll(ctx, role, page, limit)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]StaffResponse, len(items))
	for i := range items {
		resp[i] = ToStaffResponse(&items[i])
	}

	return resp, total, nil
}

func (s *staffService) GetStaff(ctx context.Context, id string) (*StaffResponse, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := ToStaffResponse(item)
	return &res, nil
}

func (s *staffService) CreateStaff(ctx context.Context, req CreateStaffRequest) (*StaffResponse, error) {
	// Check username / email unique
	_, err := s.repo.FindByEmail(ctx, req.Email)
	if err == nil {
		return nil, errs.NewConflict("email already registered")
	}
	_, err = s.repo.FindByUsername(ctx, req.Username)
	if err == nil {
		return nil, errs.NewConflict("username already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errs.NewInternal("failed to hash password", err)
	}

	staff := &domain.StaffAccount{
		FullName:       req.FullName,
		Username:       req.Username,
		Email:          req.Email,
		PasswordHash:   string(hash),
		WhatsappNumber: req.WhatsappNumber,
		Role:           req.Role,
		Status:         domain.StatusAktif,
		PositionTitle:  req.PositionTitle,
		ShortBio:       req.ShortBio,
	}

	if err = s.repo.Create(ctx, staff); err != nil {
		return nil, err
	}

	res := ToStaffResponse(staff)
	return &res, nil
}

func (s *staffService) UpdateStaff(ctx context.Context, id string, req UpdateStaffRequest) (*StaffResponse, error) {
	staff, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	staff.FullName = req.FullName
	staff.WhatsappNumber = req.WhatsappNumber
	staff.PositionTitle = req.PositionTitle
	staff.ShortBio = req.ShortBio

	if err = s.repo.Update(ctx, staff); err != nil {
		return nil, err
	}

	res := ToStaffResponse(staff)
	return &res, nil
}

func (s *staffService) ToggleStatus(ctx context.Context, id string) (*StaffResponse, error) {
	staff, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if staff.Status == domain.StatusAktif {
		staff.Status = domain.StatusNonaktif
	} else {
		staff.Status = domain.StatusAktif
	}

	if err = s.repo.Update(ctx, staff); err != nil {
		return nil, err
	}

	res := ToStaffResponse(staff)
	return &res, nil
}

func (s *staffService) UpdateMyProfile(ctx context.Context, staffID string, req UpdateProfileRequest) (*StaffResponse, error) {
	staff, err := s.repo.FindByID(ctx, staffID)
	if err != nil {
		return nil, err
	}

	staff.FullName = req.FullName
	staff.WhatsappNumber = req.WhatsappNumber
	staff.PositionTitle = req.PositionTitle
	staff.ShortBio = req.ShortBio
	if req.ProfilePhotoURL != "" {
		staff.ProfilePhotoURL = req.ProfilePhotoURL
	}

	if err = s.repo.Update(ctx, staff); err != nil {
		return nil, err
	}

	res := ToStaffResponse(staff)
	return &res, nil
}

func (s *staffService) DeleteStaff(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
