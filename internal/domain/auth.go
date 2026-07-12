package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StaffRole represents the role field on a staff account.
type StaffRole string

const (
	RoleAdmin StaffRole = "admin"
	RoleStaff StaffRole = "staff"
)

// AccountStatus represents the active/inactive status of any account.
type AccountStatus string

const (
	StatusAktif    AccountStatus = "aktif"
	StatusNonaktif AccountStatus = "nonaktif"
)

// OwnerType identifies whether a polymorphic record belongs to a staff or patient.
type OwnerType string

const (
	OwnerTypeStaff   OwnerType = "staff"
	OwnerTypePatient OwnerType = "patient"
)

// StaffAccount represents an admin or staff monitoring account.
type StaffAccount struct {
	BaseModel

	FullName        string        `gorm:"type:varchar(150);not null" json:"full_name"`
	Username        string        `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email           string        `gorm:"type:varchar(150);uniqueIndex;not null" json:"email"`
	PasswordHash    string        `gorm:"type:varchar(255);not null" json:"-"`
	WhatsappNumber  string        `gorm:"type:varchar(20)" json:"whatsapp_number"`
	Role            StaffRole     `gorm:"type:staff_role_enum;not null" json:"role"`
	Status          AccountStatus `gorm:"type:account_status_enum;not null;default:aktif" json:"status"`
	PositionTitle   string        `gorm:"type:varchar(100)" json:"position_title"`
	ShortBio        string        `gorm:"type:text" json:"short_bio"`
	ProfilePhotoURL string        `gorm:"type:text" json:"profile_photo_url"`
}

func (StaffAccount) TableName() string { return "staff_accounts" }

// PasswordResetToken is a one-time OTP token for password reset.
type PasswordResetToken struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime;not null" json:"created_at"`

	OwnerType OwnerType `gorm:"type:owner_type_enum;not null" json:"owner_type"`
	OwnerID   string    `gorm:"type:uuid;not null" json:"owner_id"`
	Email     string    `gorm:"type:varchar(150);not null" json:"email"`
	OTPCode   string    `gorm:"type:varchar(10);not null" json:"otp_code"`
	IsUsed    bool      `gorm:"not null;default:false" json:"is_used"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
}

func (PasswordResetToken) TableName() string { return "password_reset_tokens" }

func (p *PasswordResetToken) BeforeCreate(_ *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	return nil
}

// AuthSession tracks active refresh tokens for logout/revocation support.
type AuthSession struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime;not null" json:"created_at"`

	OwnerType    OwnerType `gorm:"type:owner_type_enum;not null" json:"owner_type"`
	OwnerID      string    `gorm:"type:uuid;not null" json:"owner_id"`
	DeviceInfo   string    `gorm:"type:varchar(255)" json:"device_info"`
	RefreshToken string    `gorm:"type:varchar(500);uniqueIndex;not null" json:"refresh_token"`
	ExpiresAt    time.Time `gorm:"not null" json:"expires_at"`
}

func (AuthSession) TableName() string { return "auth_sessions" }

func (a *AuthSession) BeforeCreate(_ *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	return nil
}
