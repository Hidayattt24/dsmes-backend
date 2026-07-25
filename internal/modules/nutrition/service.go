package nutrition

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type nutritionService struct {
	repo NutritionRepository
	log  *zap.Logger
}

func NewNutritionService(repo NutritionRepository, log *zap.Logger) NutritionService {
	return &nutritionService{repo: repo, log: log}
}

func (s *nutritionService) SearchFoods(ctx context.Context, patientID string, query string) ([]FoodResponse, error) {
	items, err := s.repo.SearchFoods(ctx, query)
	if err != nil {
		return nil, err
	}

	resp := make([]FoodResponse, len(items))
	for i := range items {
		resp[i] = ToFoodResponse(&items[i])
	}
	return resp, nil
}

func (s *nutritionService) GetRecentFoods(ctx context.Context, patientID string) ([]FoodResponse, error) {
	items, err := s.repo.FindRecentSearches(ctx, patientID)
	if err != nil {
		return nil, err
	}

	resp := make([]FoodResponse, 0, len(items))
	for _, item := range items {
		if item.Food != nil {
			resp = append(resp, ToFoodResponse(item.Food))
		}
	}
	return resp, nil
}

func (s *nutritionService) LogMeal(ctx context.Context, patientID string, req LogMealRequest) (*MealLogResponse, error) {
	food, err := s.repo.FindFoodByID(ctx, req.FoodID)
	if err != nil {
		return nil, err
	}

	log := &domain.MealLog{
		PatientID:         patientID,
		FoodID:            req.FoodID,
		MealType:          req.MealType,
		PortionMultiplier: req.PortionMultiplier,
		LoggedAt:          time.Now(),
	}

	if err = s.repo.CreateMealLog(ctx, log); err != nil {
		return nil, err
	}

	// Async log recent searches to improve analytics
	_ = s.repo.UpsertRecentSearch(ctx, patientID, req.FoodID)

	return &MealLogResponse{
		ID:                log.ID,
		Food:              ToFoodResponse(food),
		MealType:          log.MealType,
		PortionMultiplier: log.PortionMultiplier,
		LoggedAt:          log.LoggedAt.Format(time.RFC3339),
	}, nil
}

func (s *nutritionService) GetDailyNutritionSummary(ctx context.Context, patientID string, dateStr string) (*DailyNutritionSummaryResponse, error) {
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	meals, err := s.repo.FindMealsByPatientAndDate(ctx, patientID, dateStr)
	if err != nil {
		return nil, err
	}

	target, err := s.repo.GetDailyCalorieTarget(ctx, patientID)
	if err != nil {
		target = 2000
	}

	var consumed float64
	var carbs float64
	var protein float64
	var fat float64

	for _, m := range meals {
		if m.Food != nil {
			consumed += m.Food.Calories * m.PortionMultiplier
			carbs += m.Food.CarbsG * m.PortionMultiplier
			protein += m.Food.ProteinG * m.PortionMultiplier
			fat += m.Food.FatG * m.PortionMultiplier
		}
	}

	remaining := float64(target) - consumed
	if remaining < 0 {
		remaining = 0
	}

	return &DailyNutritionSummaryResponse{
		CaloriesConsumed:   consumed,
		DailyCalorieTarget: target,
		CaloriesRemaining:  remaining,
		TotalCarbsG:        carbs,
		TotalProteinG:      protein,
		TotalFatG:          fat,
	}, nil
}

func (s *nutritionService) CreateFood(ctx context.Context, req CreateFoodRequest) (*FoodResponse, error) {
	food := &domain.Food{
		Name:                req.Name,
		DefaultServingUnit:  req.DefaultServingUnit,
		DefaultServingGrams: req.DefaultServingGrams,
		Calories:            req.Calories,
		CarbsG:              req.CarbsG,
		ProteinG:            req.ProteinG,
		FatG:                req.FatG,
	}

	if err := s.repo.CreateFood(ctx, food); err != nil {
		return nil, err
	}

	res := ToFoodResponse(food)
	return &res, nil
}

func (s *nutritionService) UpdateFood(ctx context.Context, id string, req CreateFoodRequest) (*FoodResponse, error) {
	food, err := s.repo.FindFoodByID(ctx, id)
	if err != nil {
		return nil, err
	}

	food.Name = req.Name
	food.DefaultServingUnit = req.DefaultServingUnit
	food.DefaultServingGrams = req.DefaultServingGrams
	food.Calories = req.Calories
	food.CarbsG = req.CarbsG
	food.ProteinG = req.ProteinG
	food.FatG = req.FatG

	if err = s.repo.UpdateFood(ctx, food); err != nil {
		return nil, err
	}

	res := ToFoodResponse(food)
	return &res, nil
}

func (s *nutritionService) GetPatientMealLogs(ctx context.Context, patientID string, dateStr string) ([]MealLogResponse, error) {
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	meals, err := s.repo.FindMealsByPatientAndDate(ctx, patientID, dateStr)
	if err != nil {
		return nil, err
	}

	resp := make([]MealLogResponse, len(meals))
	for i, m := range meals {
		var fResp FoodResponse
		if m.Food != nil {
			fResp = ToFoodResponse(m.Food)
		}
		resp[i] = MealLogResponse{
			ID:                m.ID,
			Food:              fResp,
			MealType:          m.MealType,
			PortionMultiplier: m.PortionMultiplier,
			LoggedAt:          m.LoggedAt.Format(time.RFC3339),
		}
	}
	return resp, nil
}

func (s *nutritionService) CalculateCalories(ctx context.Context, req CalorieCalculationRequest) (*CalorieCalculationResponse, error) {
	req.Normalize()
	return CalculateDailyCalories(req)
}

