package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/infrastructure/email"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	jwtpkg "github.com/dsmes/dsmes-backend/internal/pkg/jwt"
	"github.com/dsmes/dsmes-backend/internal/pkg/phone"
)

type authService struct {
	repo  AuthRepository
	jwt   *jwtpkg.Manager
	email email.EmailService
	log   *zap.Logger
}

func NewAuthService(repo AuthRepository, jwt *jwtpkg.Manager, email email.EmailService, log *zap.Logger) AuthService {
	return &authService{repo: repo, jwt: jwt, email: email, log: log}
}

// ── StaffLogin ────────────────────────────────────────────────────────────────

func (s *authService) StaffLogin(ctx context.Context, req StaffLoginRequest) (*LoginResponse, error) {
	staff, err := s.repo.FindStaffByEmail(ctx, req.Email)
	if err != nil {
		return nil, errs.NewUnauthorized("invalid email or password")
	}

	if staff.Status == StatusNonaktif {
		return nil, errs.NewForbidden("account is deactivated — contact administrator")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errs.NewUnauthorized("invalid email or password")
	}

	tokens, err := s.jwt.GenerateTokenPair(staff.ID, staff.Email, string(staff.Role))
	if err != nil {
		return nil, errs.NewInternal("failed to generate tokens", err)
	}

	session := &AuthSession{
		OwnerType:    OwnerTypeStaff,
		OwnerID:      staff.ID,
		RefreshToken: hashRefreshToken(tokens.RefreshToken),
		ExpiresAt:    time.Now().Add(s.jwt.RefreshTTL()),
	}
	if err = s.repo.CreateSession(ctx, session); err != nil {
		s.log.Warn("auth: failed to persist session", zap.Error(err), zap.String("staff_id", staff.ID))
	}

	return &LoginResponse{
		User: AuthUserResponse{
			ID:       staff.ID,
			FullName: staff.FullName,
			Email:    staff.Email,
			Role:     string(staff.Role),
		},
		Tokens: *tokens,
	}, nil
}

// ── PatientLogin ──────────────────────────────────────────────────────────────

func (s *authService) PatientLogin(ctx context.Context, req PatientLoginRequest) (*LoginResponse, error) {
	phoneNum, err := phone.Normalize(req.PhoneNumber)
	if err != nil {
		return nil, errs.NewBadRequest(err.Error())
	}

	patient, err := s.repo.FindPatientByPhoneNumber(ctx, phoneNum)
	if err != nil {
		return nil, errs.NewUnauthorized("nomor handphone atau kata sandi tidak valid")
	}

	if patient.Status == domain.StatusNonaktif {
		return nil, errs.NewForbidden("account is deactivated")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(patient.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errs.NewUnauthorized("nomor handphone atau kata sandi tidak valid")
	}

	tokens, err := s.jwt.GenerateTokenPair(patient.ID, patient.GetEmail(), "user")
	if err != nil {
		return nil, errs.NewInternal("failed to generate tokens", err)
	}

	session := &AuthSession{
		OwnerType:    OwnerTypePatient,
		OwnerID:      patient.ID,
		RefreshToken: hashRefreshToken(tokens.RefreshToken),
		ExpiresAt:    time.Now().Add(s.jwt.RefreshTTL()),
	}
	if err = s.repo.CreateSession(ctx, session); err != nil {
		s.log.Warn("auth: failed to persist session", zap.Error(err), zap.String("patient_id", patient.ID))
	}

	return &LoginResponse{
		User: AuthUserResponse{
			ID:          patient.ID,
			FullName:    patient.FullName,
			PhoneNumber: patient.PhoneNumber,
			Email:       patient.GetEmail(),
			Role:        "user",
		},
		Tokens: *tokens,
	}, nil
}

// ── Logout ────────────────────────────────────────────────────────────────────

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	return s.repo.RevokeSession(ctx, hashRefreshToken(refreshToken))
}

// ── ForgotPassword ────────────────────────────────────────────────────────────

func (s *authService) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error {
	ownerType := OwnerType(req.OwnerType)
	if ownerType == "" {
		ownerType = OwnerTypePatient
	}

	var ownerID string
	if ownerType == OwnerTypeStaff {
		staff, err := s.repo.FindStaffByEmail(ctx, req.Email)
		if err != nil {
			return nil // Generic response — don't leak account existence
		}
		ownerID = staff.ID
	} else {
		patient, err := s.repo.FindPatientByEmail(ctx, req.Email)
		if err != nil {
			return nil // Generic response
		}
		ownerID = patient.ID
	}

	otp := generateOTP()
	// Store only hashed OTP in database
	otpHash, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return errs.NewInternal("failed to hash OTP", err)
	}

	token := &PasswordResetToken{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		Email:     req.Email,
		OTPCode:   string(otpHash),
		IsUsed:    false,
		ExpiresAt: time.Now().Add(5 * time.Minute), // Expire in 5 minutes
	}

	if err := s.repo.CreateResetToken(ctx, token); err != nil {
		return err
	}

	// Send OTP email using Resend
	if err := s.email.SendPasswordResetOTP(ctx, req.Email, otp); err != nil {
		s.log.Error("auth: failed to send password reset OTP email", zap.String("email", req.Email), zap.Error(err))
		return errs.NewInternal("failed to send OTP email", err)
	}

	s.log.Info("auth: OTP generated and sent successfully",
		zap.String("email", req.Email),
		zap.Time("expires_at", token.ExpiresAt),
	)

	return nil
}

// ── VerifyOTP ─────────────────────────────────────────────────────────────────

func (s *authService) VerifyOTP(ctx context.Context, req VerifyOTPRequest) error {
	ownerType := OwnerType(req.OwnerType)
	tokens, err := s.repo.FindActiveResetTokens(ctx, req.Email, ownerType)
	if err != nil {
		return errs.NewBadRequest("OTP is invalid or has expired")
	}

	var matched bool
	for _, token := range tokens {
		if err := bcrypt.CompareHashAndPassword([]byte(token.OTPCode), []byte(req.OTPCode)); err == nil {
			matched = true
			break
		}
	}

	if !matched {
		return errs.NewBadRequest("OTP is invalid or has expired")
	}

	return nil
}

// ── ResetPassword ─────────────────────────────────────────────────────────────

func (s *authService) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	ownerType := OwnerType(req.OwnerType)

	tokens, err := s.repo.FindActiveResetTokens(ctx, req.Email, ownerType)
	if err != nil {
		return errs.NewBadRequest("OTP is invalid or has expired")
	}

	var matchedToken *PasswordResetToken
	for _, t := range tokens {
		if err := bcrypt.CompareHashAndPassword([]byte(t.OTPCode), []byte(req.OTPCode)); err == nil {
			matchedToken = &t
			break
		}
	}

	if matchedToken == nil {
		return errs.NewBadRequest("OTP is invalid or has expired")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errs.NewInternal("failed to hash password", err)
	}

	if ownerType == OwnerTypeStaff {
		if err = s.repo.UpdateStaffPassword(ctx, matchedToken.OwnerID, string(hash)); err != nil {
			return err
		}
	} else {
		if err = s.repo.UpdatePatientPassword(ctx, matchedToken.OwnerID, string(hash)); err != nil {
			return err
		}
	}

	if err := s.repo.MarkTokenUsed(ctx, matchedToken.ID); err != nil {
		return err
	}

	// Invalidate every active session so old refresh tokens stop working
	// after a password reset.
	return s.repo.RevokeAllSessions(ctx, ownerType, matchedToken.OwnerID)
}

// ── RefreshToken ──────────────────────────────────────────────────────────────

func (s *authService) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*TokenResponse, error) {
	refreshHash := hashRefreshToken(req.RefreshToken)

	session, err := s.repo.FindSession(ctx, refreshHash)
	if err != nil {
		return nil, errs.NewUnauthorized("refresh token is invalid or expired")
	}

	claims, err := s.jwt.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errs.NewUnauthorized("refresh token is invalid")
	}

	_ = s.repo.RevokeSession(ctx, refreshHash)

	tokens, err := s.jwt.GenerateTokenPair(claims.UserID, claims.Email, claims.Role)
	if err != nil {
		return nil, errs.NewInternal("failed to generate tokens", err)
	}

	newSession := &AuthSession{
		OwnerType:    session.OwnerType,
		OwnerID:      session.OwnerID,
		DeviceInfo:   session.DeviceInfo,
		RefreshToken: hashRefreshToken(tokens.RefreshToken),
		ExpiresAt:    time.Now().Add(s.jwt.RefreshTTL()),
	}
	if err = s.repo.CreateSession(ctx, newSession); err != nil {
		s.log.Warn("auth: failed to persist rotated session", zap.Error(err))
	}

	return &TokenResponse{Tokens: *tokens}, nil
}

// hashRefreshToken stores a one-way SHA-256 digest of the refresh token in the
// database so a leaked auth_sessions table cannot be replayed as valid tokens.
func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func generateOTP() string {
	return fmt.Sprintf("%06d", rand.IntN(1_000_000))
}
