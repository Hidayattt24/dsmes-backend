package ai_chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AIChatRepository interface {
	CreateConversation(ctx context.Context, conv *AIConversation) error
	GetConversationByID(ctx context.Context, id uuid.UUID, patientID uuid.UUID) (*AIConversation, error)
	ListConversationsByPatient(ctx context.Context, patientID uuid.UUID) ([]AIConversation, error)
	DeleteConversation(ctx context.Context, id uuid.UUID, patientID uuid.UUID) error
	UpdateConversationTitle(ctx context.Context, id uuid.UUID, title string) error

	CreateMessage(ctx context.Context, msg *AIMessage) error
	GetMessagesByConversation(ctx context.Context, convID uuid.UUID, patientID uuid.UUID, limit int) ([]AIMessage, error)

	CreatePromptLog(ctx context.Context, log *AIPromptLog) error
	GetPatientHealthContext(ctx context.Context, patientID uuid.UUID) (*PatientHealthContext, error)
}

type aiChatRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewAIChatRepository(db *gorm.DB, logger *zap.Logger) AIChatRepository {
	// AutoMigrate AI tables if they don't exist yet
	if err := db.AutoMigrate(&AIConversation{}, &AIMessage{}, &AIPromptLog{}); err != nil {
		logger.Error("Failed to auto migrate AI Chat tables", zap.Error(err))
	}

	return &aiChatRepository{
		db:     db,
		logger: logger,
	}
}

func (r *aiChatRepository) CreateConversation(ctx context.Context, conv *AIConversation) error {
	return r.db.WithContext(ctx).Create(conv).Error
}

func (r *aiChatRepository) GetConversationByID(ctx context.Context, id uuid.UUID, patientID uuid.UUID) (*AIConversation, error) {
	var conv AIConversation
	err := r.db.WithContext(ctx).
		Where("id = ? AND patient_id = ?", id, patientID).
		First(&conv).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &conv, nil
}

func (r *aiChatRepository) ListConversationsByPatient(ctx context.Context, patientID uuid.UUID) ([]AIConversation, error) {
	var conversations []AIConversation
	err := r.db.WithContext(ctx).
		Where("patient_id = ?", patientID).
		Order("updated_at DESC").
		Find(&conversations).Error
	return conversations, err
}

func (r *aiChatRepository) DeleteConversation(ctx context.Context, id uuid.UUID, patientID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND patient_id = ?", id, patientID).
		Delete(&AIConversation{}).Error
}

func (r *aiChatRepository) UpdateConversationTitle(ctx context.Context, id uuid.UUID, title string) error {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if strings.TrimSpace(title) != "" {
		updates["title"] = strings.TrimSpace(title)
	}
	return r.db.WithContext(ctx).
		Model(&AIConversation{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *aiChatRepository) CreateMessage(ctx context.Context, msg *AIMessage) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *aiChatRepository) GetMessagesByConversation(ctx context.Context, convID uuid.UUID, patientID uuid.UUID, limit int) ([]AIMessage, error) {
	var messages []AIMessage
	query := r.db.WithContext(ctx).
		Where("conversation_id = ? AND patient_id = ?", convID, patientID).
		Order("created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&messages).Error
	return messages, err
}

func (r *aiChatRepository) CreatePromptLog(ctx context.Context, log *AIPromptLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *aiChatRepository) GetPatientHealthContext(ctx context.Context, patientID uuid.UUID) (*PatientHealthContext, error) {
	pContext := &PatientHealthContext{}

	// 1. Fetch patient profile
	var patientData struct {
		FullName           string
		DateOfBirth        time.Time
		Gender             string
		DiabetesType       string
		HeightCm           float64
		WeightKg           float64
		DailyCalorieTarget int
	}
	err := r.db.WithContext(ctx).Table("patients").
		Select("full_name, date_of_birth, gender, diabetes_type, height_cm, weight_kg, daily_calorie_target").
		Where("id = ? AND deleted_at IS NULL", patientID).
		Scan(&patientData).Error

	if err == nil {
		pContext.Name = patientData.FullName
		if !patientData.DateOfBirth.IsZero() {
			now := time.Now()
			age := now.Year() - patientData.DateOfBirth.Year()
			if now.YearDay() < patientData.DateOfBirth.YearDay() {
				age--
			}
			pContext.Age = age
		}
		pContext.Gender = patientData.Gender
		pContext.DiabetesType = patientData.DiabetesType
		pContext.DailyCalorieTarget = patientData.DailyCalorieTarget

		// Calculate BMI if height and weight exist
		if patientData.HeightCm > 0 && patientData.WeightKg > 0 {
			heightM := patientData.HeightCm / 100.0
			pContext.BMI = patientData.WeightKg / (heightM * heightM)
		}
	}

	// 2. Fetch latest blood sugar & average
	var latestBs struct {
		GlucoseValue        float64
		MeasurementTimeType string
		MeasuredAt          time.Time
	}
	errBs := r.db.WithContext(ctx).Table("blood_sugar_logs").
		Select("glucose_value, measurement_time_type, measured_at").
		Where("patient_id = ? AND deleted_at IS NULL", patientID).
		Order("measured_at DESC").
		Limit(1).
		Scan(&latestBs).Error

	if errBs == nil && latestBs.GlucoseValue > 0 {
		label := domain.GetMeasurementTypeLabel(domain.MeasurementTime(latestBs.MeasurementTimeType))
		pContext.LatestBloodSugar = fmt.Sprintf("%.0f mg/dL (%s)", latestBs.GlucoseValue, label)
	}

	var avgBs struct {
		AvgGlucose float64
	}
	_ = r.db.WithContext(ctx).Table("blood_sugar_logs").
		Select("AVG(glucose_value) as avg_glucose").
		Where("patient_id = ? AND deleted_at IS NULL", patientID).
		Scan(&avgBs).Error

	if avgBs.AvgGlucose > 0 {
		pContext.AverageBloodSugar = fmt.Sprintf("%.0f mg/dL", avgBs.AvgGlucose)
	}

	// 3. Fetch active medications
	var meds []string
	_ = r.db.WithContext(ctx).Table("reminders").
		Where("patient_id = ? AND is_active = true AND deleted_at IS NULL", patientID).
		Pluck("activity_name", &meds).Error
	pContext.ActiveMedications = meds

	// 4. Fetch recent activity logs
	var activities []string
	_ = r.db.WithContext(ctx).Table("patient_activity_logs").
		Where("patient_id = ? AND deleted_at IS NULL", patientID).
		Order("logged_at DESC").
		Limit(3).
		Pluck("activity_name", &activities).Error
	pContext.RecentActivities = activities

	// 5. Fetch recent meal logs
	var meals []string
	_ = r.db.WithContext(ctx).Table("meal_logs ml").
		Select("f.name").
		Joins("JOIN foods f ON f.id = ml.food_id").
		Where("ml.patient_id = ? AND ml.deleted_at IS NULL", patientID).
		Order("ml.created_at DESC").
		Limit(3).
		Pluck("f.name", &meals).Error
	pContext.RecentMeals = meals

	return pContext, nil
}
