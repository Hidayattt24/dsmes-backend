package summary

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type summaryRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewSummaryRepository(db *gorm.DB, log *zap.Logger) SummaryRepository {
	return &summaryRepository{db: db, log: log}
}

func (r *summaryRepository) FindLatest(ctx context.Context, patientID string) (*domain.WeeklyHealthSummary, error) {
	var s domain.WeeklyHealthSummary
	err := r.db.WithContext(ctx).
		Where("patient_id = ? AND deleted_at IS NULL", patientID).
		Order("week_start_date DESC").
		First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("weekly summary not found")
		}
		return nil, errs.NewInternal("failed to fetch latest weekly summary", err)
	}
	return &s, nil
}

func (r *summaryRepository) Upsert(ctx context.Context, s *domain.WeeklyHealthSummary) error {
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "patient_id"}, {Name: "week_start_date"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"week_end_date", "blood_sugar_trend", "avg_blood_sugar",
			"articles_read_count", "articles_target_count", "generated_at", "updated_at",
		}),
	}).Create(s).Error
	if err != nil {
		return errs.NewInternal("failed to upsert weekly summary", err)
	}
	return nil
}

func (r *summaryRepository) CalculateWeeklyAverage(ctx context.Context, patientID string, start, end time.Time) (float64, error) {
	var avg float64
	err := r.db.WithContext(ctx).
		Model(&domain.BloodSugarLog{}).
		Where("patient_id = ? AND measured_at BETWEEN ? AND ? AND deleted_at IS NULL", patientID, start, end).
		Select("COALESCE(AVG(glucose_value), 0)").
		Row().
		Scan(&avg)
	if err != nil {
		return 0, errs.NewInternal("failed to calculate avg blood sugar", err)
	}
	return avg, nil
}

func (r *summaryRepository) CalculateCompletionsCount(ctx context.Context, patientID string, start, end time.Time) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.UserArticleCompletion{}).
		Where("patient_id = ? AND completed_at BETWEEN ? AND ? AND deleted_at IS NULL", patientID, start, end).
		Count(&count).Error
	if err != nil {
		return 0, errs.NewInternal("failed to count weekly completions", err)
	}
	return int(count), nil
}
