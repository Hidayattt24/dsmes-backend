package staff

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/modules/auth"
	"github.com/dsmes/dsmes-backend/internal/modules/facility"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type staffService struct {
	repo         StaffRepository
	authRepo     auth.AuthRepository
	facilityRepo facility.FacilityRepository
	log          *zap.Logger
}

func NewStaffService(repo StaffRepository, authRepo auth.AuthRepository, facilityRepo facility.FacilityRepository, log *zap.Logger) StaffService {
	return &staffService{repo: repo, authRepo: authRepo, facilityRepo: facilityRepo, log: log}
}

func (s *staffService) ListStaff(ctx context.Context, search, status string, role *domain.StaffRole, page, limit int) ([]StaffResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	items, total, err := s.repo.FindAll(ctx, search, status, role, page, limit)
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

	var healthFacilityID *string
	if req.HealthFacilityID != "" {
		if err := s.validateFacility(ctx, req.HealthFacilityID); err != nil {
			return nil, err
		}
		healthFacilityID = &req.HealthFacilityID
	}

	staff := &domain.StaffAccount{
		FullName:         req.FullName,
		Username:         req.Username,
		Email:            req.Email,
		PasswordHash:     string(hash),
		WhatsappNumber:   req.WhatsappNumber,
		Role:             req.Role,
		Status:           domain.StatusAktif,
		PositionTitle:    req.PositionTitle,
		ShortBio:         req.ShortBio,
		ProfilePhotoURL:  req.ProfilePhotoURL,
		HealthFacilityID: healthFacilityID,
	}

	if err = s.repo.Create(ctx, staff); err != nil {
		return nil, err
	}

	// Reload so the HealthFacility relation is populated for the response.
	staff, err = s.repo.FindByID(ctx, staff.ID)
	if err != nil {
		return nil, err
	}

	res := ToStaffResponse(staff)
	return &res, nil
}

// validateFacility ensures the referenced facility exists.
func (s *staffService) validateFacility(ctx context.Context, id string) error {
	if _, err := s.facilityRepo.FindByID(ctx, id); err != nil {
		return errs.NewBadRequest("puskesmas tidak ditemukan")
	}
	return nil
}

func (s *staffService) UpdateStaff(ctx context.Context, id string, req UpdateStaffRequest) (*StaffResponse, error) {
	staff, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Username != staff.Username {
		existing, err := s.repo.FindByUsername(ctx, req.Username)
		if err == nil && existing.ID != id {
			return nil, errs.NewConflict("username already exists")
		}
	}

	if req.Email != staff.Email {
		existing, err := s.repo.FindByEmail(ctx, req.Email)
		if err == nil && existing.ID != id {
			return nil, errs.NewConflict("email already registered")
		}
	}

	staff.FullName = req.FullName
	staff.Username = req.Username
	staff.Email = req.Email
	staff.WhatsappNumber = req.WhatsappNumber
	staff.PositionTitle = req.PositionTitle
	staff.ShortBio = req.ShortBio
	staff.ProfilePhotoURL = req.ProfilePhotoURL

	if req.HealthFacilityID != "" {
		if err := s.validateFacility(ctx, req.HealthFacilityID); err != nil {
			return nil, err
		}
		staff.HealthFacilityID = &req.HealthFacilityID
	} else {
		staff.HealthFacilityID = nil
	}

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

	// Validate duplicate username if changed
	if req.Username != staff.Username {
		existing, err := s.repo.FindByUsername(ctx, req.Username)
		if err == nil && existing.ID != staffID {
			return nil, errs.NewConflict("username already exists")
		}
	}

	// Validate duplicate email if changed
	if req.Email != staff.Email {
		existing, err := s.repo.FindByEmail(ctx, req.Email)
		if err == nil && existing.ID != staffID {
			return nil, errs.NewConflict("email already registered")
		}
	}

	staff.FullName = req.FullName
	staff.Username = req.Username
	staff.Email = req.Email
	staff.WhatsappNumber = req.WhatsappNumber
	staff.PositionTitle = req.PositionTitle
	staff.ShortBio = req.ShortBio
	staff.ProfilePhotoURL = req.ProfilePhotoURL

	if err = s.repo.Update(ctx, staff); err != nil {
		return nil, err
	}

	res := ToStaffResponse(staff)
	return &res, nil
}

func (s *staffService) ChangePassword(ctx context.Context, staffID string, req ChangePasswordRequest) error {
	staff, err := s.repo.FindByID(ctx, staffID)
	if err != nil {
		return err
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return errs.NewUnauthorized("kata sandi saat ini salah")
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errs.NewInternal("failed to hash password", err)
	}

	staff.PasswordHash = string(hash)
	if err := s.repo.Update(ctx, staff); err != nil {
		return err
	}

	// Invalidate every active session so old refresh tokens stop working.
	return s.authRepo.RevokeAllSessions(ctx, auth.OwnerTypeStaff, staffID)
}

func (s *staffService) DeleteStaff(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
