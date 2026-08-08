package history

import (
	"context"

	"go.uber.org/zap"
)

type historyService struct {
	repo HistoryRepository
	log  *zap.Logger
}

func NewHistoryService(repo HistoryRepository, log *zap.Logger) HistoryService {
	return &historyService{repo: repo, log: log}
}

func (s *historyService) GetPatientHistory(ctx context.Context, patientID string, page, limit int) ([]HistoryItemResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	items, total, err := s.repo.FindAll(ctx, patientID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]HistoryItemResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toResponse(item))
	}

	return responses, total, nil
}

func (s *historyService) DeleteHistoryItem(ctx context.Context, patientID string, activityType string, id string) error {
	return s.repo.DeleteHistoryItem(ctx, patientID, activityType, id)
}

func toResponse(item historyRawItem) HistoryItemResponse {
	metadata := make(map[string]any)

	switch item.ActivityType {
	case "blood_sugar":
		if item.GlucoseValue != nil {
			metadata["glucose_value"] = *item.GlucoseValue
		}
		if item.MeasurementType != "" {
			metadata["measurement_type"] = item.MeasurementType
		}
	case "meal":
		if item.Calories != nil {
			metadata["calories"] = *item.Calories
		}
		if item.CarbsG != nil {
			metadata["carbs_g"] = *item.CarbsG
		}
		if item.ProteinG != nil {
			metadata["protein_g"] = *item.ProteinG
		}
		if item.FatG != nil {
			metadata["fat_g"] = *item.FatG
		}
		if item.MealType != "" {
			metadata["meal_type"] = item.MealType
		}
	case "activity":
		if item.ActivityMinutes != nil {
			metadata["activity_minutes"] = *item.ActivityMinutes
		}
		if item.RoutineType != "" {
			metadata["routine_type"] = item.RoutineType
		}
	case "medication":
		if item.ReminderName != "" {
			metadata["reminder_name"] = item.ReminderName
		}
		if item.LogDate != "" {
			metadata["log_date"] = item.LogDate
		}
	}

	return HistoryItemResponse{
		ID:           item.ID,
		PatientID:    item.PatientID,
		ActivityType: item.ActivityType,
		Title:        item.Title,
		Subtitle:     item.Subtitle,
		Category:     item.Category,
		Value:        item.Value,
		Unit:         item.Unit,
		Status:       item.Status,
		Notes:        item.Notes,
		MeasuredAt:   item.MeasuredAt,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
		RecordedBy:   item.RecordedBy,
		Icon:         item.Icon,
		Color:        item.Color,
		Metadata:     metadata,
	}
}
