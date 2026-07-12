package auth

import "github.com/dsmes/dsmes-backend/internal/domain"

// Type aliases linking the Auth module to the shared internal/domain package.
// This prevents circular dependencies and allows other modules to reference
// staff accounts and sessions via the domain layer.

type StaffAccount = domain.StaffAccount
type PasswordResetToken = domain.PasswordResetToken
type AuthSession = domain.AuthSession

type StaffRole = domain.StaffRole

const (
	RoleAdmin = domain.RoleAdmin
	RoleStaff = domain.RoleStaff
)

type AccountStatus = domain.AccountStatus

const (
	StatusAktif    = domain.StatusAktif
	StatusNonaktif = domain.StatusNonaktif
)

type OwnerType = domain.OwnerType

const (
	OwnerTypeStaff   = domain.OwnerTypeStaff
	OwnerTypePatient = domain.OwnerTypePatient
)
