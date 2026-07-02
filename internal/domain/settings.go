package domain

import (
	"time"
)

type TicketStatus string

const (
	TicketOpen       TicketStatus = "open"
	TicketInProgress TicketStatus = "in_progress"
	TicketClosed     TicketStatus = "closed"
)

// FAQ represents a frequently asked question.
type FAQ struct {
	BaseModel

	Question     string `gorm:"type:text;not null" json:"question"`
	Answer       string `gorm:"type:text;not null" json:"answer"`
	DisplayOrder int    `gorm:"default:0" json:"display_order"`
	IsActive     bool   `gorm:"default:true" json:"is_active"`
}

func (FAQ) TableName() string { return "faqs" }

// SupportTicket represents a support inquiry from a patient.
type SupportTicket struct {
	BaseModel

	PatientID  string       `gorm:"type:uuid;not null" json:"patient_id"`
	Subject    string       `gorm:"type:varchar(200);not null" json:"subject"`
	Message    string       `gorm:"type:text;not null" json:"message"`
	Status     TicketStatus `gorm:"type:ticket_status_enum;not null;default:open" json:"status"`
	ResolvedAt *time.Time   `json:"resolved_at,omitempty"`
}

func (SupportTicket) TableName() string { return "support_tickets" }
