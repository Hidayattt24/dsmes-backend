package nutrition

import (
	"context"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type NutritionRepository interface {
	SearchFoods(ctx context.Context, query string) ([]domain.Food, error)
	FindFoodByID(ctx context.Context, id string) (*domain.Food, error)
	CreateFood(ctx context.Context, f *domain.Food) error
	UpdateFood(ctx context.Context, f *domain.Food) error

	CreateMealLog(ctx context.Context, log *domain.MealLog) error
	FindMealsByPatientAndDate(ctx context.Context, patientID string, dateStr string) ([]domain.MealLog, error)
	GetDailyCalorieTarget(ctx context.Context, patientID string) (int, error)

	UpsertRecentSearch(ctx context.Context, patientID string, foodID string) error
	FindRecentSearches(ctx context.Context, patientID string) ([]domain.RecentFoodSearch, error)
}

type NutritionService interface {
	SearchFoods(ctx context.Context, patientID string, query string) ([]FoodResponse, error)
	GetRecentFoods(ctx context.Context, patientID string) ([]FoodResponse, error)
	LogMeal(ctx context.Context, patientID string, req LogMealRequest) (*MealLogResponse, error)
	GetDailyNutritionSummary(ctx context.Context, patientID string, dateStr string) (*DailyNutritionSummaryResponse, error)
	GetPatientMealLogs(ctx context.Context, patientID string, dateStr string) ([]MealLogResponse, error)
	CreateFood(ctx context.Context, req CreateFoodRequest) (*FoodResponse, error)
	UpdateFood(ctx context.Context, id string, req CreateFoodRequest) (*FoodResponse, error)
	CalculateCalories(ctx context.Context, req CalorieCalculationRequest) (*CalorieCalculationResponse, error)
}

