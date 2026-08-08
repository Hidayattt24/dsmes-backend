package auth

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type authRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewAuthRepository(db *gorm.DB, log *zap.Logger) AuthRepository {
	return &authRepository{db: db, log: log}
}

// ── Staff ─────────────────────────────────────────────────────────────────────

func (r *authRepository) FindStaffByEmail(ctx context.Context, email string) (*StaffAccount, error) {
	var s StaffAccount
	err := r.db.WithContext(ctx).
		Where("email = ? AND deleted_at IS NULL", email).
		First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("staff account not found")
		}
		return nil, errs.NewInternal("failed to find staff account", err)
	}
	return &s, nil
}

func (r *authRepository) FindStaffByID(ctx context.Context, id string) (*StaffAccount, error) {
	var s StaffAccount
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("staff account not found")
		}
		return nil, errs.NewInternal("failed to find staff account", err)
	}
	return &s, nil
}

func (r *authRepository) UpdateStaffPassword(ctx context.Context, id, hash string) error {
	result := r.db.WithContext(ctx).
		Model(&StaffAccount{}).
		Where("id = ?", id).
		Update("password_hash", hash)
	if result.Error != nil {
		return errs.NewInternal("failed to update password", result.Error)
	}
	if result.RowsAffected == 0 {
		return errs.NewNotFound("staff account not found")
	}
	return nil
}

// ── Patient ───────────────────────────────────────────────────────────────────

func (r *authRepository) FindPatientByEmail(ctx context.Context, email string) (*domain.Patient, error) {
	var p domain.Patient
	err := r.db.WithContext(ctx).
		Where("email = ? AND deleted_at IS NULL", email).
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("patient account not found")
		}
		return nil, errs.NewInternal("failed to find patient account", err)
	}
	return &p, nil
}

func (r *authRepository) FindPatientByPhoneNumber(ctx context.Context, phone string) (*domain.Patient, error) {
	var p domain.Patient
	err := r.db.WithContext(ctx).
		Where("phone_number = ? AND deleted_at IS NULL", phone).
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("patient account not found")
		}
		return nil, errs.NewInternal("failed to find patient account by phone number", err)
	}
	return &p, nil
}

func (r *authRepository) FindPatientByID(ctx context.Context, id string) (*domain.Patient, error) {
	var p domain.Patient
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("patient account not found")
		}
		return nil, errs.NewInternal("failed to find patient account", err)
	}
	return &p, nil
}

func (r *authRepository) UpdatePatientPassword(ctx context.Context, id, hash string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Patient{}).
		Where("id = ?", id).
		Update("password_hash", hash)
	if result.Error != nil {
		return errs.NewInternal("failed to update patient password", result.Error)
	}
	if result.RowsAffected == 0 {
		return errs.NewNotFound("patient not found")
	}
	return nil
}

// ── OTP / Reset ───────────────────────────────────────────────────────────────

func (r *authRepository) CreateResetToken(ctx context.Context, token *PasswordResetToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return errs.NewInternal("failed to create reset token", err)
	}
	return nil
}

func (r *authRepository) FindActiveResetTokens(ctx context.Context, email string, ownerType OwnerType) ([]PasswordResetToken, error) {
	var tokens []PasswordResetToken
	err := r.db.WithContext(ctx).
		Where("email = ? AND owner_type = ? AND is_used = false AND expires_at > ?",
			email, ownerType, time.Now()).
		Order("expires_at DESC").
		Find(&tokens).Error
	if err != nil {
		return nil, errs.NewInternal("failed to find active reset tokens", err)
	}
	return tokens, nil
}

func (r *authRepository) MarkTokenUsed(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).
		Model(&PasswordResetToken{}).
		Where("id = ?", id).
		Update("is_used", true)
	if result.Error != nil {
		return errs.NewInternal("failed to mark token used", result.Error)
	}
	return nil
}

// ── Session / Refresh ─────────────────────────────────────────────────────────

func (r *authRepository) CreateSession(ctx context.Context, session *AuthSession) error {
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return errs.NewInternal("failed to create auth session", err)
	}
	return nil
}

func (r *authRepository) FindSession(ctx context.Context, refreshToken string) (*AuthSession, error) {
	var s AuthSession
	err := r.db.WithContext(ctx).
		Where("refresh_token = ? AND expires_at > ?", refreshToken, time.Now()).
		First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("session not found or expired")
		}
		return nil, errs.NewInternal("failed to find session", err)
	}
	return &s, nil
}

func (r *authRepository) RevokeSession(ctx context.Context, refreshToken string) error {
	result := r.db.WithContext(ctx).
		Where("refresh_token = ?", refreshToken).
		Delete(&AuthSession{})
	if result.Error != nil {
		return errs.NewInternal("failed to revoke session", result.Error)
	}
	return nil
}

func (r *authRepository) RevokeAllSessions(ctx context.Context, ownerType OwnerType, ownerID string) error {
	result := r.db.WithContext(ctx).
		Where("owner_type = ? AND owner_id = ?", ownerType, ownerID).
		Delete(&AuthSession{})
	if result.Error != nil {
		return errs.NewInternal("failed to revoke all sessions", result.Error)
	}
	return nil
}
