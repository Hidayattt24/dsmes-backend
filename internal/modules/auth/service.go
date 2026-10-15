package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/infrastructure/email"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	jwtpkg "github.com/dsmes/dsmes-backend/internal/pkg/jwt"
	"github.com/dsmes/dsmes-backend/internal/pkg/phone"
)

type rotationReplayEntry struct {
	tokenResponse TokenResponse
	ownerID       string
	expiresAt     time.Time
}

type authService struct {
	repo         AuthRepository
	jwt          *jwtpkg.Manager
	email        email.EmailService
	log          *zap.Logger
	replayMu     sync.RWMutex
	replayBuffer map[string]rotationReplayEntry
}

func NewAuthService(repo AuthRepository, jwt *jwtpkg.Manager, email email.EmailService, log *zap.Logger) AuthService {
	return &authService{
		repo:         repo,
		jwt:          jwt,
		email:        email,
		log:          log,
		replayBuffer: make(map[string]rotationReplayEntry),
	}
}

// ── StaffLogin ────────────────────────────────────────────────────────────────

func (s *authService) StaffLogin(ctx context.Context, req StaffLoginRequest) (*LoginResponse, error) {
	staff, err := s.repo.FindStaffByEmail(ctx, req.Email)
	if err != nil {
		return nil, errs.NewUnauthorized("Email tidak terdaftar")
	}

	if staff.Status == StatusNonaktif {
		return nil, errs.NewForbidden("Akun Anda telah dinonaktifkan — hubungi administrator")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errs.NewUnauthorized("Kata sandi yang Anda masukkan salah")
	}

	tokens, err := s.jwt.GenerateTokenPair(staff.ID, staff.Email, string(staff.Role))
	if err != nil {
		return nil, errs.NewInternal("failed to generate tokens", err)
	}

	session := &AuthSession{
		OwnerType:    OwnerTypeStaff,
		OwnerID:      staff.ID,
		RefreshToken: HashRefreshToken(tokens.RefreshToken),
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
		RefreshToken: HashRefreshToken(tokens.RefreshToken),
		ExpiresAt:    time.Now().Add(s.jwt.RefreshTTL()),
	}
	if err = s.repo.CreateSession(ctx, session); err != nil {
		s.log.Warn("auth: failed to persist session", zap.Error(err), zap.String("patient_id", patient.ID))
	}

	// Record patient login event for "Login Hari Ini" metric
	loginRecord := &domain.PatientLogin{
		PatientID: patient.ID,
		LoginAt:   time.Now(),
	}
	if err = s.repo.RecordPatientLogin(ctx, loginRecord); err != nil {
		s.log.Warn("auth: failed to record patient login", zap.Error(err), zap.String("patient_id", patient.ID))
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
	return s.repo.RevokeSession(ctx, HashRefreshToken(refreshToken))
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

// ── CheckPhoneNumber ──────────────────────────────────────────────────────────

func (s *authService) CheckPhoneNumber(ctx context.Context, req ForgotPasswordCheckPhoneRequest) error {
	phoneNum, err := phone.Normalize(req.PhoneNumber)
	if err != nil {
		return errs.NewBadRequest(err.Error())
	}

	patient, err := s.repo.FindPatientByPhoneNumber(ctx, phoneNum)
	if err != nil {
		if errs.IsNotFound(err) {
			return errs.NewNotFound("nomor handphone tidak terdaftar")
		}
		return err
	}

	if patient.Status == domain.StatusNonaktif {
		return errs.NewForbidden("account is deactivated")
	}

	return nil
}

// ── ResetPasswordByPhone ──────────────────────────────────────────────────────

func (s *authService) ResetPasswordByPhone(ctx context.Context, req ResetPasswordByPhoneRequest) error {
	phoneNum, err := phone.Normalize(req.PhoneNumber)
	if err != nil {
		return errs.NewBadRequest(err.Error())
	}

	patient, err := s.repo.FindPatientByPhoneNumber(ctx, phoneNum)
	if err != nil {
		if errs.IsNotFound(err) {
			return errs.NewNotFound("nomor handphone tidak terdaftar")
		}
		return err
	}

	if patient.Status == domain.StatusNonaktif {
		return errs.NewForbidden("account is deactivated")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errs.NewInternal("failed to hash password", err)
	}

	if err := s.repo.UpdatePatientPassword(ctx, patient.ID, string(hash)); err != nil {
		return err
	}

	// Invalidate every active session so old refresh tokens stop working
	// after a password reset.
	return s.repo.RevokeAllSessions(ctx, OwnerTypePatient, patient.ID)
}

// ── CheckEmail ────────────────────────────────────────────────────────────────

func (s *authService) CheckEmail(ctx context.Context, req ForgotPasswordCheckEmailRequest) error {
	staff, err := s.repo.FindStaffByEmail(ctx, req.Email)
	if err != nil {
		if errs.IsNotFound(err) {
			return errs.NewNotFound("email tidak terdaftar")
		}
		return err
	}

	if staff.Status == StatusNonaktif {
		return errs.NewForbidden("account is deactivated")
	}

	return nil
}

// ── ResetPasswordByEmail ──────────────────────────────────────────────────────

func (s *authService) ResetPasswordByEmail(ctx context.Context, req ResetPasswordByEmailRequest) error {
	staff, err := s.repo.FindStaffByEmail(ctx, req.Email)
	if err != nil {
		if errs.IsNotFound(err) {
			return errs.NewNotFound("email tidak terdaftar")
		}
		return err
	}

	if staff.Status == StatusNonaktif {
		return errs.NewForbidden("account is deactivated")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errs.NewInternal("failed to hash password", err)
	}

	if err := s.repo.UpdateStaffPassword(ctx, staff.ID, string(hash)); err != nil {
		return err
	}

	// Invalidate every active session so old refresh tokens stop working
	// after a password reset.
	return s.repo.RevokeAllSessions(ctx, OwnerTypeStaff, staff.ID)
}

// ── RefreshToken ──────────────────────────────────────────────────────────────

func (s *authService) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*TokenResponse, error) {
	refreshHash := HashRefreshToken(req.RefreshToken)

	// 1. Check ephemeral replay buffer under read lock (fast path).
	// NOTE: This buffer is strictly in-memory (RAM) with a 15-second TTL.
	// It resets on server restart and is local to this backend process.
	// It safely recovers lost HTTP responses or concurrent retries without creating orphan sessions.
	s.replayMu.RLock()
	if entry, ok := s.replayBuffer[refreshHash]; ok && time.Now().Before(entry.expiresAt) {
		s.replayMu.RUnlock()
		s.log.Info("auth: replaying rotated token from ephemeral buffer", zap.String("owner_id", entry.ownerID))
		return &entry.tokenResponse, nil
	}
	s.replayMu.RUnlock()

	// 2. Parse and validate the JWT signature and claims before database rotation.
	claims, err := s.jwt.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errs.NewUnauthorized("refresh token is invalid")
	}

	// 3. Acquire write lock to synchronize concurrent rotations for the same token.
	s.replayMu.Lock()
	defer s.replayMu.Unlock()

	// Double-check replay buffer after acquiring write lock.
	if entry, ok := s.replayBuffer[refreshHash]; ok && time.Now().Before(entry.expiresAt) {
		return &entry.tokenResponse, nil
	}

	// 4. Find existing active session in database.
	session, err := s.repo.FindSession(ctx, refreshHash)
	if err != nil {
		if errs.IsNotFound(err) {
			return nil, errs.NewUnauthorized("refresh token is invalid or expired")
		}
		// Return true database error (HTTP 500) — DO NOT mask database failures as 401!
		return nil, err
	}

	// 5. Generate fresh token pair.
	tokens, err := s.jwt.GenerateTokenPair(claims.UserID, claims.Email, claims.Role)
	if err != nil {
		return nil, errs.NewInternal("failed to generate tokens", err)
	}

	newSession := &AuthSession{
		OwnerType:    session.OwnerType,
		OwnerID:      session.OwnerID,
		DeviceInfo:   session.DeviceInfo,
		RefreshToken: HashRefreshToken(tokens.RefreshToken),
		ExpiresAt:    time.Now().Add(s.jwt.RefreshTTL()),
	}

	// 6. Atomically delete old session and insert new session in single DB transaction.
	if err = s.repo.RotateSession(ctx, refreshHash, newSession); err != nil {
		if errs.IsNotFound(err) {
			return nil, errs.NewUnauthorized("refresh token is invalid or expired")
		}
		return nil, err
	}

	tokenResp := TokenResponse{Tokens: *tokens}

	// 7. Record into ephemeral replay buffer (15-second TTL).
	s.replayBuffer[refreshHash] = rotationReplayEntry{
		tokenResponse: tokenResp,
		ownerID:       session.OwnerID,
		expiresAt:     time.Now().Add(15 * time.Second),
	}

	// Prune stale entries if map grows
	if len(s.replayBuffer) > 100 {
		now := time.Now()
		for k, v := range s.replayBuffer {
			if now.After(v.expiresAt) {
				delete(s.replayBuffer, k)
			}
		}
	}

	return &tokenResp, nil
}

// HashRefreshToken stores a one-way SHA-256 digest of the refresh token in the
// database so a leaked auth_sessions table cannot be replayed as valid tokens.
func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func generateOTP() string {
	return fmt.Sprintf("%06d", rand.IntN(1_000_000))
}
