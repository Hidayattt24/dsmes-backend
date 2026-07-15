package staff

import (
	"context"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type StaffRepository interface {
	FindAll(ctx context.Context, search, status string, role *domain.StaffRole, page, limit int) ([]domain.StaffAccount, int64, error)
	FindByID(ctx context.Context, id string) (*domain.StaffAccount, error)
	FindByEmail(ctx context.Context, email string) (*domain.StaffAccount, error)
	FindByUsername(ctx context.Context, username string) (*domain.StaffAccount, error)
	Create(ctx context.Context, s *domain.StaffAccount) error
	Update(ctx context.Context, s *domain.StaffAccount) error
	Delete(ctx context.Context, id string) error
}

type StaffService interface {
	ListStaff(ctx context.Context, search, status string, role *domain.StaffRole, page, limit int) ([]StaffResponse, int64, error)
	GetStaff(ctx context.Context, id string) (*StaffResponse, error)
	CreateStaff(ctx context.Context, req CreateStaffRequest) (*StaffResponse, error)
	UpdateStaff(ctx context.Context, id string, req UpdateStaffRequest) (*StaffResponse, error)
	ToggleStatus(ctx context.Context, id string) (*StaffResponse, error)
	UpdateMyProfile(ctx context.Context, staffID string, req UpdateProfileRequest) (*StaffResponse, error)
	ChangePassword(ctx context.Context, staffID string, req ChangePasswordRequest) error
	DeleteStaff(ctx context.Context, id string) error
}
