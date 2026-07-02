// Package domain contains shared domain types used across all business modules.
//
// This package is the innermost layer in Clean Architecture — it has ZERO
// dependencies on any framework, database driver, or external library.
// It defines only primitive types and interfaces that describe business rules.
//
// # BaseModel
//
// All GORM entity structs MUST embed BaseModel instead of gorm.Model.
// This enforces the database-architect skill rules:
//
//   - UUID v4 primary key (never auto-increment integer)
//   - created_at / updated_at managed by GORM hooks
//   - deleted_at soft-delete (records are never physically deleted)
//
// # Usage
//
//	type Patient struct {
//	    domain.BaseModel          // ← embed this
//	    Email    string `gorm:"uniqueIndex;not null"`
//	    FullName string `gorm:"not null"`
//	}
package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel is the shared GORM base for every entity in the application.
// Embed this struct (not gorm.Model) in all models to enforce consistent
// primary key strategy and audit timestamps across all tables.
//
// GORM hooks:
//   - BeforeCreate: generates a UUID v4 if ID is empty
//   - UpdatedAt:    automatically set by GORM on every save
//   - DeletedAt:    enables soft-delete; queries auto-exclude deleted rows
type BaseModel struct {
	// ID is a UUID v4 primary key. Using UUID instead of auto-increment integers:
	//   - Prevents ID enumeration attacks (security)
	//   - Allows client-side ID generation
	//   - Works safely across distributed systems and DB replicas
	ID string `gorm:"type:uuid;primaryKey" json:"id"`

	// CreatedAt is set once when the record is first inserted.
	CreatedAt time.Time `gorm:"autoCreateTime;not null" json:"created_at"`

	// UpdatedAt is updated by GORM on every Save / Update call.
	UpdatedAt time.Time `gorm:"autoUpdateTime;not null" json:"updated_at"`

	// DeletedAt enables soft-delete. A non-NULL value means the record is
	// "deleted" — GORM will automatically filter it out of all queries.
	// Use db.Unscoped() to include soft-deleted records.
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate is a GORM hook that runs before every INSERT statement.
// It assigns a new UUID v4 to the record's ID field if it has not been set.
func (b *BaseModel) BeforeCreate(_ *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.NewString()
	}
	return nil
}
