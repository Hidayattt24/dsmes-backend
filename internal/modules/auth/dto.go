package auth

import jwtpkg "github.com/dsmes/dsmes-backend/internal/pkg/jwt"

// ─── Request DTOs ─────────────────────────────────────────────────────────────

// StaffLoginRequest is the body for POST /api/v1/auth/staff/login.
type StaffLoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// PatientLoginRequest is the body for POST /api/v1/auth/patient/login.
type PatientLoginRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	Password    string `json:"password"     validate:"required,min=6"`
}

// ForgotPasswordRequest is the body for POST /api/v1/auth/forgot-password.
type ForgotPasswordRequest struct {
	Email     string `json:"email"      validate:"required,email"`
	OwnerType string `json:"owner_type" validate:"omitempty,oneof=staff patient"`
}

// ForgotPasswordCheckPhoneRequest is the body for POST /api/v1/auth/forgot-password/check-phone.
type ForgotPasswordCheckPhoneRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
}

// ResetPasswordByPhoneRequest is the body for POST /api/v1/auth/reset-password-by-phone.
type ResetPasswordByPhoneRequest struct {
	PhoneNumber     string `json:"phone_number"     validate:"required"`
	NewPassword     string `json:"new_password"     validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

// ForgotPasswordCheckEmailRequest is the body for POST /api/v1/auth/forgot-password/check-email.
type ForgotPasswordCheckEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordByEmailRequest is the body for POST /api/v1/auth/reset-password-by-email.
type ResetPasswordByEmailRequest struct {
	Email           string `json:"email"            validate:"required,email"`
	NewPassword     string `json:"new_password"     validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

// VerifyOTPRequest is the body for POST /api/v1/auth/verify-otp.
type VerifyOTPRequest struct {
	Email     string `json:"email"      validate:"required,email"`
	OTPCode   string `json:"otp_code"   validate:"required,len=6"`
	OwnerType string `json:"owner_type" validate:"required,oneof=staff patient"`
}

// ResetPasswordRequest is the body for POST /api/v1/auth/reset-password.
type ResetPasswordRequest struct {
	Email           string `json:"email"            validate:"required,email"`
	OTPCode         string `json:"otp_code"         validate:"required,len=6"`
	OwnerType       string `json:"owner_type"       validate:"required,oneof=staff patient"`
	NewPassword     string `json:"new_password"     validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

// RefreshTokenRequest is the body for POST /api/v1/auth/refresh.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ─── Response DTOs ────────────────────────────────────────────────────────────

// LoginResponse is returned after a successful login.
type LoginResponse struct {
	User   AuthUserResponse `json:"user"`
	Tokens jwtpkg.TokenPair `json:"tokens"`
}

// AuthUserResponse is the minimal user info included in the login response.
// Never expose password_hash or internal fields.
type AuthUserResponse struct {
	ID          string `json:"id"`
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email,omitempty"`
	Role        string `json:"role"`
}

// TokenResponse wraps just the token pair — used by the refresh endpoint.
type TokenResponse struct {
	Tokens jwtpkg.TokenPair `json:"tokens"`
}
