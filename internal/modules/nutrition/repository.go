package nutrition

import (
	"context"
	"errors"
	"strings"
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
	trimmed := strings.TrimSpace(query)
	db := r.db.WithContext(ctx).Model(&domain.Food{}).Where("deleted_at IS NULL")

	if trimmed != "" {
		words := strings.Fields(trimmed)
		if len(words) <= 1 {
			pattern := "%" + trimmed + "%"
			db = db.Where("name ILIKE ?", pattern)
		} else {
			for _, w := range words {
				pattern := "%" + w + "%"
				db = db.Where("name ILIKE ?", pattern)
			}
		}

		exact := strings.ToLower(trimmed)
		prefix := strings.ToLower(trimmed) + "%"
		wordPrefix := "% " + strings.ToLower(trimmed) + "%"
		db = db.Order(gorm.Expr(`
			CASE 
				WHEN LOWER(name) = ? THEN 1
				WHEN LOWER(name) LIKE ? THEN 2
				WHEN LOWER(name) LIKE ? THEN 3
				ELSE 4
			END ASC,
			LENGTH(name) ASC,
			name ASC
		`, exact, prefix, wordPrefix))
	} else {
		db = db.Order("name ASC")
	}

	err := db.Limit(30).Find(&items).Error
	if err != nil {
		return nil, errs.NewInternal("failed to search foods", err)
	}
	return items, nil
}

func (r *nutritionRepository) FindFoodByID(ctx context.Context, id string) (*domain.Food, error) {
	var f domain.Food
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&f).Error
	if err == nil {
		return &f, nil
	}

	var fm domain.FoodMaster
	errMaster := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&fm).Error
	if errMaster == nil {
		servingGrams := 100.0
		calories := fm.EnergyKcal
		if calories == 0 && fm.Calories > 0 {
			calories = fm.Calories
		}
		carbs := fm.CarbohydrateG
		if carbs == 0 && fm.TotalCarbohydrate > 0 {
			carbs = fm.TotalCarbohydrate
		}
		protein := fm.ProteinG
		if protein == 0 && fm.Protein > 0 {
			protein = fm.Protein
		}
		fat := fm.FatG
		if fat == 0 && fm.TotalFat > 0 {
			fat = fm.TotalFat
		}

		return &domain.Food{
			BaseModel:           fm.BaseModel,
			Name:                fm.Name,
			DefaultServingUnit:  fm.ServingSize,
			DefaultServingGrams: servingGrams,
			Calories:            calories,
			CarbsG:              carbs,
			ProteinG:            protein,
			FatG:                fat,
			SugarG:              fm.SugarG,
			SodiumMg:            fm.SodiumMg,
			FiberG:              fm.FiberG,
		}, nil
	}

	return nil, errs.NewNotFound("food item not found")
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

func (r *nutritionRepository) FindMealLogByID(ctx context.Context, patientID string, id string) (*domain.MealLog, error) {
	var log domain.MealLog
	err := r.db.WithContext(ctx).
		Preload("Food").
		Where("id = ? AND patient_id = ? AND deleted_at IS NULL", id, patientID).
		First(&log).Error
	if err != nil {
		return nil, errs.NewNotFound("meal log not found")
	}
	return &log, nil
}

func (r *nutritionRepository) FindMealLogByClientID(ctx context.Context, patientID string, clientID string) (*domain.MealLog, error) {
	var log domain.MealLog
	err := r.db.WithContext(ctx).
		Preload("Food").
		Where("client_id = ? AND patient_id = ? AND deleted_at IS NULL", clientID, patientID).
		First(&log).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("meal log not found")
		}
		return nil, errs.NewInternal("failed to fetch meal log by client_id", err)
	}
	return &log, nil
}

func (r *nutritionRepository) UpdateMealLog(ctx context.Context, log *domain.MealLog) error {
	result := r.db.WithContext(ctx).Save(log)
	if result.Error != nil {
		return errs.NewInternal("failed to update meal log", result.Error)
	}
	return nil
}

func (r *nutritionRepository) DeleteMealLog(ctx context.Context, patientID string, id string) error {
	result := r.db.WithContext(ctx).Where("id = ? AND patient_id = ?", id, patientID).Delete(&domain.MealLog{})
	if result.Error != nil {
		return errs.NewInternal("failed to delete meal log", result.Error)
	}
	return nil
}

func (r *nutritionRepository) FindMealsByPatientAndDate(ctx context.Context, patientID string, dateStr string) ([]domain.MealLog, error) {
	var items []domain.MealLog
	q := r.db.WithContext(ctx).Preload("Food").Preload("FoodMaster").Where("patient_id = ? AND deleted_at IS NULL", patientID)
	if dateStr != "" {
		// Range predicate so the (patient_id, logged_at) index can be used.
		q = q.Where("logged_at >= ?::date AND logged_at < (?::date + INTERVAL '1 day')", dateStr, dateStr)
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
		return domain.DefaultDailyCalorieTarget, nil // default fallback
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
