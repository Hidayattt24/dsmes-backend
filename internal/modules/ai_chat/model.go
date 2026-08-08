package ai_chat

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AIConversation represents a chat session for a patient.
type AIConversation struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PatientID uuid.UUID      `gorm:"type:uuid;not null;index" json:"patient_id"`
	Title     string         `gorm:"type:varchar(255);not null;default:'Percakapan Baru'" json:"title"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Messages []AIMessage `gorm:"foreignKey:ConversationID;constraint:OnDelete:CASCADE" json:"messages,omitempty"`
}

func (AIConversation) TableName() string {
	return "ai_conversations"
}

// AIMessage represents an individual user or assistant message inside a conversation.
type AIMessage struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConversationID uuid.UUID      `gorm:"type:uuid;not null;index" json:"conversation_id"`
	PatientID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"patient_id"`
	Role           string         `gorm:"type:varchar(20);not null" json:"role"` // 'user', 'assistant', 'system'
	Message        string         `gorm:"type:text;not null" json:"message"`
	TokenCount     int            `gorm:"default:0" json:"token_count"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (AIMessage) TableName() string {
	return "ai_messages"
}

// AIPromptLog audit log for tracking prompt generation and AI provider calls.
type AIPromptLog struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PatientID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"patient_id"`
	ConversationID  *uuid.UUID `gorm:"type:uuid;index" json:"conversation_id,omitempty"`
	GeneratedPrompt string     `gorm:"type:text;not null" json:"generated_prompt"`
	Model           string     `gorm:"type:varchar(100);not null" json:"model"`
	ExecutionTimeMS int        `gorm:"default:0" json:"execution_time_ms"`
	Status          string     `gorm:"type:varchar(20);not null;default:'success'" json:"status"`
	ErrorMessage    string     `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (AIPromptLog) TableName() string {
	return "ai_prompt_logs"
}
