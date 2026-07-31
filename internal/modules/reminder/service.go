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
		IconName:           req.IconName,
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
	rem.IconName = req.IconName
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

func (s *reminderService) LogMedication(ctx context.Context, patientID string, req LogMedicationRequest) (*MedicationLogResponse, error) {
	rem, err := s.repo.FindByID(ctx, req.ReminderID)
	if err != nil {
		return nil, err
	}
	if rem.PatientID != patientID {
		return nil, errs.NewForbidden("unauthorized access to reminder")
	}

	logDate := time.Now()
	if req.LogDate != "" {
		if t, err := time.Parse("2006-01-02", req.LogDate); err == nil {
			logDate = t
		}
	}

	status := domain.ReminderLogStatus(req.Status)
	if status == "" {
		status = domain.ReminderSelesai
	}

	log := &domain.DailyReminderLog{
		ReminderID: req.ReminderID,
		LogDate:    logDate,
		Status:     status,
	}

	if err := s.repo.UpsertLog(ctx, log); err != nil {
		return nil, err
	}

	return &MedicationLogResponse{
		ID:            log.ID,
		ReminderID:    req.ReminderID,
		ActivityName:  rem.ActivityName,
		Category:      rem.Category,
		ScheduledTime: rem.ScheduledTime,
		Status:        status,
		LoggedDate:    logDate.Format("2006-01-02"),
	}, nil
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

func (s *reminderService) MarkReadByID(ctx context.Context, patientID string, notifID string) error {
	return s.repo.MarkNotificationReadByID(ctx, patientID, notifID)
}

func (s *reminderService) DeleteNotificationByID(ctx context.Context, patientID string, notifID string) error {
	return s.repo.DeleteNotificationByID(ctx, patientID, notifID)
}

func (s *reminderService) GetPatientMedicationLogs(ctx context.Context, patientID string, dateStr string) ([]MedicationLogResponse, error) {
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	reminders, err := s.repo.FindAllByPatientID(ctx, patientID)
	if err != nil {
		return nil, err
	}

	logs, err := s.repo.FindLogsByPatientAndDate(ctx, patientID, dateStr)
	if err != nil {
		return nil, err
	}

	logMap := make(map[string]domain.DailyReminderLog)
	for _, log := range logs {
		logMap[log.ReminderID] = log
	}

	var resp []MedicationLogResponse
	for _, r := range reminders {
		if r.Category != domain.CategoryMedisObat {
			continue
		}

		if log, exists := logMap[r.ID]; exists {
			resp = append(resp, MedicationLogResponse{
				ID:            log.ID,
				ReminderID:    r.ID,
				ActivityName:  r.ActivityName,
				Category:      r.Category,
				ScheduledTime: r.ScheduledTime,
				Status:        log.Status,
				LoggedDate:    log.LogDate.Format("2006-01-02"),
			})
		}
	}

	return resp, nil
}
