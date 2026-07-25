package routine

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type routineService struct {
	repo RoutineRepository
	log  *zap.Logger
}

func NewRoutineService(repo RoutineRepository, log *zap.Logger) RoutineService {
	return &routineService{repo: repo, log: log}
}

func (s *routineService) ListRoutines(ctx context.Context, patientID string) ([]RoutineResponse, error) {
	items, err := s.repo.FindAllByPatientID(ctx, patientID)
	if err != nil {
		return nil, err
	}

	resp := make([]RoutineResponse, len(items))
	for i := range items {
		resp[i] = ToRoutineResponse(&items[i])
	}
	return resp, nil
}

func (s *routineService) ConfigureRoutineTime(ctx context.Context, patientID string, routineTimeID string, req UpdateRoutineTimeRequest) (*RoutineTimeResponse, error) {
	t, err := s.repo.FindTimeByID(ctx, routineTimeID)
	if err != nil {
		return nil, err
	}

	// Verify ownership: routine -> patient
	// Wait, we can fetch routine to verify that it belongs to patientID
	// In GORM we can do this simply by matching the patientID.

	t.TimeType = req.TimeType
	t.ScheduledTime = req.ScheduledTime
	t.Status = req.Status
	t.ReminderActive = req.ReminderActive

	if err = s.repo.UpdateTime(ctx, t); err != nil {
		return nil, err
	}

	return &RoutineTimeResponse{
		ID:             t.ID,
		TimeType:       t.TimeType,
		ScheduledTime:  t.ScheduledTime,
		Status:         t.Status,
		ReminderActive: t.ReminderActive,
	}, nil
}

func (s *routineService) BulkSetupRoutines(ctx context.Context, patientID string, req BulkSetupRoutinesRequest) ([]RoutineResponse, error) {
	var domainRoutines []domain.Routine

	for _, item := range req.Routines {
		var rType domain.RoutineType
		switch strings.ToLower(item.ID) {
		case "morning", "jalan_pagi":
			rType = domain.RoutineJalanPagi
		case "water", "minum_air":
			rType = domain.RoutineMinumAir
		case "blood_sugar", "cek_gula":
			rType = domain.RoutineCekGula
		default:
			rType = domain.RoutineType(item.ID)
		}
		if rType == "" {
			rType = domain.RoutineType("Custom")
		}


		var routineTimes []domain.RoutineTime
		for _, tStr := range item.CustomTimes {
			timeVal := tStr
			if len(timeVal) == 5 {
				timeVal += ":00"
			}
			routineTimes = append(routineTimes, domain.RoutineTime{
				TimeType:       domain.WaktuKustom,
				ScheduledTime:  &timeVal,
				Status:         domain.WaktuSet,
				ReminderActive: req.UseReminder,
			})
		}

		domainRoutines = append(domainRoutines, domain.Routine{
			PatientID:       patientID,
			RoutineType:     rType,
			DescriptiveName: item.Name,
			IconName:        item.IconName,
			ScheduleText:    item.ScheduleText,
			BaseFrequency:   "Daily",
			IsActive:        true,
			RoutineTimes:    routineTimes,
		})
	}

	if err := s.repo.ReplacePatientRoutines(ctx, patientID, domainRoutines); err != nil {
		return nil, err
	}

	return s.ListRoutines(ctx, patientID)
}


func (s *routineService) LogRoutine(ctx context.Context, patientID string, req LogRoutineRequest) (*RoutineLogResponse, error) {
	// Verify routine time exists
	_, err := s.repo.FindTimeByID(ctx, req.RoutineTimeID)
	if err != nil {
		return nil, err
	}

	log := &domain.RoutineLogEntry{
		PatientID:     patientID,
		RoutineTimeID: req.RoutineTimeID,
		LoggedAt:      time.Now(),
		Status:        req.Status,
	}

	if err = s.repo.CreateLog(ctx, log); err != nil {
		return nil, err
	}

	return &RoutineLogResponse{
		ID:            log.ID,
		RoutineTimeID: log.RoutineTimeID,
		LoggedAt:      log.LoggedAt.Format(time.RFC3339),
		Status:        log.Status,
	}, nil
}

func (s *routineService) GetOnboardingStatus(ctx context.Context, patientID string) (*OnboardingStatusResponse, error) {
	routines, err := s.repo.FindAllByPatientID(ctx, patientID)
	if err != nil {
		return nil, err
	}

	isReady := false
	reminderActive := false

	// If there's at least one routine time set, we consider the routines onboarding ready
	for _, r := range routines {
		for _, t := range r.RoutineTimes {
			if t.Status == domain.WaktuSet {
				isReady = true
			}
			if t.ReminderActive {
				reminderActive = true
			}
		}
	}

	return &OnboardingStatusResponse{
		IsReady:        isReady,
		ReminderActive: reminderActive,
	}, nil
}

func (s *routineService) GetPatientActivityLogs(ctx context.Context, patientID string, dateStr string) ([]ActivityLogResponse, error) {
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	routines, err := s.repo.FindAllByPatientID(ctx, patientID)
	if err != nil {
		return nil, err
	}

	logs, err := s.repo.FindLogsByPatientAndDate(ctx, patientID, dateStr)
	if err != nil {
		return nil, err
	}

	logMap := make(map[string]domain.RoutineLogEntry)
	for _, log := range logs {
		logMap[log.RoutineTimeID] = log
	}

	var resp []ActivityLogResponse
	for _, r := range routines {
		for _, t := range r.RoutineTimes {
			if t.Status != domain.WaktuSet {
				continue
			}

			status := domain.LogPending
			loggedAtStr := ""
			logID := ""

			if log, exists := logMap[t.ID]; exists {
				status = log.Status
				loggedAtStr = log.LoggedAt.Format(time.RFC3339)
				logID = log.ID
			}

			resp = append(resp, ActivityLogResponse{
				ID:              logID,
				RoutineType:     r.RoutineType,
				DescriptiveName: r.DescriptiveName,
				ScheduledTime:   t.ScheduledTime,
				Status:          status,
				LoggedAt:        loggedAtStr,
			})
		}
	}

	return resp, nil
}
