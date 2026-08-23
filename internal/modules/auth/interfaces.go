package auth

import (
	"context"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type AuthRepository interface {
	// ── Staff ──────────────────────────────────────────────────────────────────
	FindStaffByEmail(ctx context.Context, email string) (*StaffAccount, error)
	FindStaffByID(ctx context.Context, id string) (*StaffAccount, error)
	UpdateStaffPassword(ctx context.Context, id, hash string) error

	// ── Patient ────────────────────────────────────────────────────────────────
	FindPatientByEmail(ctx context.Context, email string) (*domain.Patient, error)
	FindPatientByPhoneNumber(ctx context.Context, phone string) (*domain.Patient, error)
	FindPatientByID(ctx context.Context, id string) (*domain.Patient, error)
	UpdatePatientPassword(ctx context.Context, id, hash string) error

	// ── OTP / Reset ────────────────────────────────────────────────────────────
	CreateResetToken(ctx context.Context, token *PasswordResetToken) error
	FindActiveResetTokens(ctx context.Context, email string, ownerType OwnerType) ([]PasswordResetToken, error)
	MarkTokenUsed(ctx context.Context, id string) error

	// ── Session / Refresh ──────────────────────────────────────────────────────
	CreateSession(ctx context.Context, session *AuthSession) error
	FindSession(ctx context.Context, refreshToken string) (*AuthSession, error)
	RevokeSession(ctx context.Context, refreshToken string) error
	RevokeAllSessions(ctx context.Context, ownerType OwnerType, ownerID string) error
}

type AuthService interface {
	StaffLogin(ctx context.Context, req StaffLoginRequest) (*LoginResponse, error)
	PatientLogin(ctx context.Context, req PatientLoginRequest) (*LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error
	VerifyOTP(ctx context.Context, req VerifyOTPRequest) error
	ResetPassword(ctx context.Context, req ResetPasswordRequest) error
	CheckPhoneNumber(ctx context.Context, req ForgotPasswordCheckPhoneRequest) error
	ResetPasswordByPhone(ctx context.Context, req ResetPasswordByPhoneRequest) error
	CheckEmail(ctx context.Context, req ForgotPasswordCheckEmailRequest) error
	ResetPasswordByEmail(ctx context.Context, req ResetPasswordByEmailRequest) error
	RefreshToken(ctx context.Context, req RefreshTokenRequest) (*TokenResponse, error)
}
