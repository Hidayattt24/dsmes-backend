package settings

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type settingsService struct {
	repo SettingsRepository
	log  *zap.Logger
}

func NewSettingsService(repo SettingsRepository, log *zap.Logger) SettingsService {
	return &settingsService{repo: repo, log: log}
}

func (s *settingsService) GetFAQs(ctx context.Context) ([]FAQResponse, error) {
	items, err := s.repo.FindAllFAQs(ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]FAQResponse, len(items))
	for i := range items {
		resp[i] = ToFAQResponse(&items[i])
	}
	return resp, nil
}

func (s *settingsService) SubmitTicket(ctx context.Context, patientID string, req CreateTicketRequest) (*TicketResponse, error) {
	t := &domain.SupportTicket{
		PatientID: patientID,
		Subject:   req.Subject,
		Message:   req.Message,
		Status:    domain.TicketOpen,
	}

	if err := s.repo.CreateTicket(ctx, t); err != nil {
		return nil, err
	}

	res := ToTicketResponse(t)
	return &res, nil
}

func (s *settingsService) GetPatientTickets(ctx context.Context, patientID string) ([]TicketResponse, error) {
	items, err := s.repo.FindTicketsByPatientID(ctx, patientID)
	if err != nil {
		return nil, err
	}

	resp := make([]TicketResponse, len(items))
	for i := range items {
		resp[i] = ToTicketResponse(&items[i])
	}
	return resp, nil
}

func (s *settingsService) GetAllTickets(ctx context.Context) ([]TicketResponse, error) {
	items, err := s.repo.FindAllTickets(ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]TicketResponse, len(items))
	for i := range items {
		resp[i] = ToTicketResponse(&items[i])
	}
	return resp, nil
}

func (s *settingsService) ResolveTicket(ctx context.Context, id string) (*TicketResponse, error) {
	t, err := s.repo.FindTicketByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	t.Status = domain.TicketClosed
	t.ResolvedAt = &now

	if err = s.repo.UpdateTicket(ctx, t); err != nil {
		return nil, err
	}

	res := ToTicketResponse(t)
	return &res, nil
}
