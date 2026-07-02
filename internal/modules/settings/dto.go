package settings

import (
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type FAQResponse struct {
	ID       string `json:"id"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type CreateTicketRequest struct {
	Subject string `json:"subject" validate:"required,min=5,max=200"`
	Message string `json:"message" validate:"required,min=10"`
}

type TicketResponse struct {
	ID         string              `json:"id"`
	Subject    string              `json:"subject"`
	Message    string              `json:"message"`
	Status     domain.TicketStatus `json:"status"`
	CreatedAt  string              `json:"created_at"`
	ResolvedAt *string             `json:"resolved_at,omitempty"`
}

func ToFAQResponse(f *domain.FAQ) FAQResponse {
	return FAQResponse{
		ID:       f.ID,
		Question: f.Question,
		Answer:   f.Answer,
	}
}

func ToTicketResponse(t *domain.SupportTicket) TicketResponse {
	var resolvedStr *string
	if t.ResolvedAt != nil {
		s := t.ResolvedAt.Format(time.RFC3339)
		resolvedStr = &s
	}
	return TicketResponse{
		ID:         t.ID,
		Subject:    t.Subject,
		Message:    t.Message,
		Status:     t.Status,
		CreatedAt:  t.CreatedAt.Format(time.RFC3339),
		ResolvedAt: resolvedStr,
	}
}
