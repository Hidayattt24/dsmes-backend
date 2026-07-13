package nutrition

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type nutritionRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewNutritionRepository(db *gorm.DB, log *zap.Logger) NutritionRepository {
	return &nutritionRepository{db: db, log: log}
}

func (r *nutritionRepository) SearchFoods(ctx context.Context, query string) ([]domain.Food, error) {
	var items []domain.Food
	pattern := "%" + query + "%"
	err := r.db.WithContext(ctx).
		Where("name ILIKE ? AND deleted_at IS NULL", pattern).
		Limit(30).
		Find(&items).Error
	if err != nil {
		return nil, errs.NewInternal("failed to search foods", err)
	}
	return items, nil
}

func (r *nutritionRepository) FindFoodByID(ctx context.Context, id string) (*domain.Food, error) {
	var f domain.Food
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&f).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("food item not found")
		}
		return nil, errs.NewInternal("failed to fetch food item", err)
	}
	return &f, nil
}

func (r *nutritionRepository) CreateFood(ctx context.Context, f *domain.Food) error {
	if err := r.db.WithContext(ctx).Create(f).Error; err != nil {
		return errs.NewInternal("failed to create food item", err)
	}
	return nil
}

func (r *nutritionRepository) UpdateFood(ctx context.Context, f *domain.Food) error {
	result := r.db.WithContext(ctx).Save(f)
	if result.Error != nil {
		return errs.NewInternal("failed to update food item", result.Error)
	}
	return nil
}

func (r *nutritionRepository) CreateMealLog(ctx context.Context, log *domain.MealLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return errs.NewInternal("failed to log meal", err)
	}
	return nil
}

func (r *nutritionRepository) FindMealsByPatientAndDate(ctx context.Context, patientID string, dateStr string) ([]domain.MealLog, error) {
	var items []domain.MealLog
	q := r.db.WithContext(ctx).Preload("Food").Where("patient_id = ? AND deleted_at IS NULL", patientID)
	if dateStr != "" {
		q = q.Where("DATE(logged_at) = ?", dateStr)
	} else {
		q = q.Where("logged_at >= NOW() - INTERVAL '30 days'")
	}
	err := q.Order("logged_at ASC").Find(&items).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch meal logs", err)
	}
	return items, nil
}

func (r *nutritionRepository) GetDailyCalorieTarget(ctx context.Context, patientID string) (int, error) {
	var target int
	err := r.db.WithContext(ctx).
		Model(&domain.Patient{}).
		Where("id = ? AND deleted_at IS NULL", patientID).
		Select("daily_calorie_target").
		Row().
		Scan(&target)
	if err != nil {
		return 2000, nil // default fallback
	}
	return target, nil
}

func (r *nutritionRepository) UpsertRecentSearch(ctx context.Context, patientID string, foodID string) error {
	var s domain.RecentFoodSearch
	err := r.db.WithContext(ctx).
		Where("patient_id = ? AND food_id = ?", patientID, foodID).
		First(&s).Error
	if err == nil {
		s.UsageCount++
		s.LastUsedAt = time.Now()
		return r.db.WithContext(ctx).Save(&s).Error
	}
	s = domain.RecentFoodSearch{
		PatientID:  patientID,
		FoodID:     foodID,
		UsageCount: 1,
		LastUsedAt: time.Now(),
	}
	return r.db.WithContext(ctx).Create(&s).Error
}

func (r *nutritionRepository) FindRecentSearches(ctx context.Context, patientID string) ([]domain.RecentFoodSearch, error) {
	var items []domain.RecentFoodSearch
	err := r.db.WithContext(ctx).
		Preload("Food").
		Where("patient_id = ? AND deleted_at IS NULL", patientID).
		Order("last_used_at DESC").
		Limit(10).
		Find(&items).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch recent searches", err)
	}
	return items, nil
}
