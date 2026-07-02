package reminder

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type reminderService struct {
	repo ReminderRepository
	log  *zap.Logger
}

func NewReminderService(repo ReminderRepository, log *zap.Logger) ReminderService {
	return &reminderService{repo: repo, log: log}
}

func (s *reminderService) ListReminders(ctx context.Context, patientID string) ([]ReminderResponse, error) {
	items, err := s.repo.FindAllByPatientID(ctx, patientID)
	if err != nil {
		return nil, err
	}

	resp := make([]ReminderResponse, len(items))
	for i := range items {
		resp[i] = ToReminderResponse(&items[i])
	}
	return resp, nil
}

func (s *reminderService) CreateReminder(ctx context.Context, patientID string, req CreateReminderRequest) (*ReminderResponse, error) {
	rem := &domain.Reminder{
		PatientID:          patientID,
		ActivityName:       req.ActivityName,
		ReminderType:       domain.ReminderPersonal,
		Category:           req.Category,
		ScheduledTime:      req.ScheduledTime,
		IsActive:           true,
		Notes:              req.Notes,
		RepeatIntervalDays: req.RepeatIntervalDays,
	}

	if err := s.repo.Create(ctx, rem, req.ActiveDays); err != nil {
		return nil, err
	}

	// Refetch to get loaded active days relations
	refetched, err := s.repo.FindByID(ctx, rem.ID)
	if err != nil {
		return nil, err
	}

	res := ToReminderResponse(refetched)
	return &res, nil
}

func (s *reminderService) UpdateReminder(ctx context.Context, patientID string, id string, req CreateReminderRequest) (*ReminderResponse, error) {
	rem, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if rem.PatientID != patientID {
		return nil, errs.NewForbidden("unauthorized access to reminder")
	}

	rem.ActivityName = req.ActivityName
	rem.Category = req.Category
	rem.ScheduledTime = req.ScheduledTime
	rem.Notes = req.Notes
	rem.RepeatIntervalDays = req.RepeatIntervalDays

	if err = s.repo.Update(ctx, rem, req.ActiveDays); err != nil {
		return nil, err
	}

	// Refetch
	refetched, err := s.repo.FindByID(ctx, rem.ID)
	if err != nil {
		return nil, err
	}

	res := ToReminderResponse(refetched)
	return &res, nil
}

func (s *reminderService) DeleteReminder(ctx context.Context, patientID string, id string) error {
	rem, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if rem.PatientID != patientID {
		return errs.NewForbidden("unauthorized access to reminder")
	}

	return s.repo.Delete(ctx, id)
}

func (s *reminderService) ToggleReminder(ctx context.Context, patientID string, id string) (*ReminderResponse, error) {
	rem, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if rem.PatientID != patientID {
		return nil, errs.NewForbidden("unauthorized access to reminder")
	}

	rem.IsActive = !rem.IsActive

	// Update without altering active days
	var days []int
	for _, d := range rem.ActiveDays {
		days = append(days, d.DayOfWeek)
	}

	if err = s.repo.Update(ctx, rem, days); err != nil {
		return nil, err
	}

	res := ToReminderResponse(rem)
	return &res, nil
}

func (s *reminderService) GetNotifications(ctx context.Context, patientID string) ([]NotificationResponse, error) {
	items, err := s.repo.FindNotificationsByPatientID(ctx, patientID)
	if err != nil {
		return nil, err
	}

	resp := make([]NotificationResponse, len(items))
	for i := range items {
		resp[i] = NotificationResponse{
			ID:          items[i].ID,
			ReminderID:  items[i].ReminderID,
			MessageText: items[i].MessageText,
			NotifiedAt:  items[i].NotifiedAt.Format(time.RFC3339),
			IsRead:      items[i].IsRead,
		}
	}
	return resp, nil
}

func (s *reminderService) MarkAllRead(ctx context.Context, patientID string) error {
	return s.repo.MarkNotificationsAsRead(ctx, patientID)
}
