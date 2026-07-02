package settings

import (
	"context"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type SettingsRepository interface {
	FindAllFAQs(ctx context.Context) ([]domain.FAQ, error)
	CreateTicket(ctx context.Context, t *domain.SupportTicket) error
	FindTicketsByPatientID(ctx context.Context, patientID string) ([]domain.SupportTicket, error)
	FindAllTickets(ctx context.Context) ([]domain.SupportTicket, error)
	FindTicketByID(ctx context.Context, id string) (*domain.SupportTicket, error)
	UpdateTicket(ctx context.Context, t *domain.SupportTicket) error
}

type SettingsService interface {
	GetFAQs(ctx context.Context) ([]FAQResponse, error)
	SubmitTicket(ctx context.Context, patientID string, req CreateTicketRequest) (*TicketResponse, error)
	GetPatientTickets(ctx context.Context, patientID string) ([]TicketResponse, error)
	GetAllTickets(ctx context.Context) ([]TicketResponse, error)
	ResolveTicket(ctx context.Context, id string) (*TicketResponse, error)
}
