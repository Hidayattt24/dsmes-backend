package history

import "context"

// HistoryRepository defines the data access contract for patient history.
type HistoryRepository interface {
	FindAll(ctx context.Context, patientID string, page, limit int) ([]historyRawItem, int64, error)
	DeleteHistoryItem(ctx context.Context, patientID string, activityType string, id string) error
}

// HistoryService defines the business logic contract for the history module.
type HistoryService interface {
	GetPatientHistory(ctx context.Context, patientID string, page, limit int) ([]HistoryItemResponse, int64, error)
	DeleteHistoryItem(ctx context.Context, patientID string, activityType string, id string) error
}

// historyRawItem is the raw database row returned by the UNION ALL query.
type historyRawItem struct {
	ID              string   `gorm:"column:id"`
	PatientID       string   `gorm:"column:patient_id"`
	ActivityType    string   `gorm:"column:activity_type"`
	Title           string   `gorm:"column:title"`
	Subtitle        string   `gorm:"column:subtitle"`
	Category        string   `gorm:"column:category"`
	Value           string   `gorm:"column:value"`
	Unit            string   `gorm:"column:unit"`
	Status          string   `gorm:"column:status"`
	Notes           string   `gorm:"column:notes"`
	MeasuredAt      string   `gorm:"column:measured_at"`
	CreatedAt       string   `gorm:"column:created_at"`
	UpdatedAt       string   `gorm:"column:updated_at"`
	RecordedBy      string   `gorm:"column:recorded_by"`
	Icon            string   `gorm:"column:icon"`
	Color           string   `gorm:"column:color"`
	GlucoseValue    *int     `gorm:"column:glucose_value"`
	MeasurementType string   `gorm:"column:measurement_type"`
	MealType        string   `gorm:"column:meal_type"`
	Calories        *float64 `gorm:"column:calories"`
	CarbsG          *float64 `gorm:"column:carbs_g"`
	ProteinG        *float64 `gorm:"column:protein_g"`
	FatG            *float64 `gorm:"column:fat_g"`
	ActivityMinutes *int     `gorm:"column:activity_minutes"`
	RoutineType     string   `gorm:"column:routine_type"`
	LogDate         string   `gorm:"column:log_date"`
	ReminderName    string   `gorm:"column:reminder_name"`
}
