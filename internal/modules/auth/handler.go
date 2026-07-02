package auth

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"github.com/dsmes/dsmes-backend/internal/pkg/response"
	"github.com/dsmes/dsmes-backend/internal/pkg/validator"
)

// AuthHandler handles all HTTP requests for the auth module.
// It contains ZERO business logic — only HTTP parsing, validation, and response writing.
type AuthHandler struct {
	svc AuthService
	log *zap.Logger
}

// NewAuthHandler creates a new handler with the given service.
func NewAuthHandler(svc AuthService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, log: log}
}

// StaffLogin handles POST /api/v1/auth/staff/login
//
// @Summary      Staff login
// @Description  Authenticates an admin or puskesmas staff member and returns a JWT token pair.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      StaffLoginRequest  true  "Login credentials"
// @Success      200   {object}  map[string]any
// @Failure      400   {object}  map[string]any
// @Failure      401   {object}  map[string]any
// @Failure      422   {object}  map[string]any
// @Router       /auth/staff/login [post]
func (h *AuthHandler) StaffLogin(c fiber.Ctx) error {
	var req StaffLoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}
	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.StaffLogin(c.Context(), req)
	if err != nil {
		return err
	}
	return response.Success(c, "login successful", res)
}

// PatientLogin handles POST /api/v1/auth/patient/login
//
// @Summary      Patient login
// @Description  Authenticates a patient and returns a JWT token pair.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      PatientLoginRequest  true  "Login credentials"
// @Success      200   {object}  map[string]any
// @Failure      400   {object}  map[string]any
// @Failure      401   {object}  map[string]any
// @Failure      422   {object}  map[string]any
// @Router       /auth/patient/login [post]
func (h *AuthHandler) PatientLogin(c fiber.Ctx) error {
	var req PatientLoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}
	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	res, err := h.svc.PatientLogin(c.Context(), req)
	if err != nil {
		return err
	}
	return response.Success(c, "login successful", res)
}


// Logout handles POST /api/v1/auth/logout
//
// @Summary      Logout
// @Description  Revokes the current refresh token session. Requires a valid JWT in the Authorization header.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      RefreshTokenRequest  true  "Refresh token to revoke"
// @Success      200   {object}  map[string]any
// @Failure      401   {object}  map[string]any
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	var req RefreshTokenRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}
	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	if err := h.svc.Logout(c.Context(), req.RefreshToken); err != nil {
		return err
	}
	return response.Success(c, "logged out successfully", nil)
}

// ForgotPassword handles POST /api/v1/auth/forgot-password
//
// @Summary      Forgot password
// @Description  Sends a 6-digit OTP to the registered email address.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      ForgotPasswordRequest  true  "Email and owner type"
// @Success      200   {object}  map[string]any
// @Failure      422   {object}  map[string]any
// @Router       /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}
	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	if err := h.svc.ForgotPassword(c.Context(), req); err != nil {
		return err
	}
	// Always return 200 — prevents email enumeration
	return response.Success(c, "if the email is registered, an OTP has been sent", nil)
}

// VerifyOTP handles POST /api/v1/auth/verify-otp
//
// @Summary      Verify OTP
// @Description  Validates a password reset OTP. Must be called before ResetPassword.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      VerifyOTPRequest  true  "OTP verification payload"
// @Success      200   {object}  map[string]any
// @Failure      400   {object}  map[string]any
// @Failure      422   {object}  map[string]any
// @Router       /auth/verify-otp [post]
func (h *AuthHandler) VerifyOTP(c fiber.Ctx) error {
	var req VerifyOTPRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}
	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	if err := h.svc.VerifyOTP(c.Context(), req); err != nil {
		return err
	}
	return response.Success(c, "OTP is valid", nil)
}

// ResetPassword handles POST /api/v1/auth/reset-password
//
// @Summary      Reset password
// @Description  Resets the user's password after OTP verification.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      ResetPasswordRequest  true  "Reset password payload"
// @Success      200   {object}  map[string]any
// @Failure      400   {object}  map[string]any
// @Failure      422   {object}  map[string]any
// @Router       /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return errs.NewBadRequest("invalid request body")
	}
	if fieldErrs := validator.Validate(&req); fieldErrs != nil {
		return response.ValidationError(c, fieldErrs)
	}

	if err := h.svc.ResetPassword(c.Context(), req); err != nil {
		return err
	}
	return response.Success(c, "password reset successfully", nil)
}

// RefreshToken handles POST /api/v1/auth/refresh
//
// @Summary      Refresh token
// @Description  Exchanges a valid refresh token for a new access + refresh token pair.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      RefreshTokenRequest  true  "Refresh token"
// @Success      200   {object}  map[string]any
// @Failure      401   {object}  map[string]any
// @Failure      422   {object}  map[string]any
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c fiber.Ctx) error {
	// Support both JSON body and Authorization header for the refresh token
	var req RefreshTokenRequest

	// Try Authorization header first: "Bearer <token>"
	authHeader := c.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		req.RefreshToken = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		if err := c.Bind().Body(&req); err != nil {
			return errs.NewBadRequest("invalid request body")
		}
		if fieldErrs := validator.Validate(&req); fieldErrs != nil {
			return response.ValidationError(c, fieldErrs)
		}
	}

	res, err := h.svc.RefreshToken(c.Context(), req)
	if err != nil {
		return err
	}
	return response.Success(c, "token refreshed", res)
}
