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
		FoodName:          food.Name,
		ServingSize:       food.DefaultServingUnit,
		Calories:          food.Calories,
		CarbsG:            food.CarbsG,
		ProteinG:          food.ProteinG,
		FatG:              food.FatG,
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
		target = domain.DefaultDailyCalorieTarget
	}

	var consumed float64
	var carbs float64
	var protein float64
	var fat float64

	for _, m := range meals {
		// Use the nutrition snapshot frozen at logging time. This works for
		// logs created from either the `foods` or the food master dataset.
		cal, c, p, f := m.Calories, m.CarbsG, m.ProteinG, m.FatG
		if m.Food != nil {
			cal, c, p, f = m.Food.Calories, m.Food.CarbsG, m.Food.ProteinG, m.Food.FatG
		}
		consumed += cal * m.PortionMultiplier
		carbs += c * m.PortionMultiplier
		protein += p * m.PortionMultiplier
		fat += f * m.PortionMultiplier
	}

	remaining := float64(target) - consumed
	if remaining < 0 {
		remaining = 0
	}

	return &DailyNutritionSummaryResponse{
		CaloriesConsumed:   consumed,
		DailyCalorieTarget: target,
		CaloriesRemaining:  remaining,
		TotalFoodMeal:      len(meals),
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

func (s *nutritionService) UpdateMealLog(ctx context.Context, patientID string, id string, req UpdateMealLogRequest) (*MealLogResponse, error) {
	log, err := s.repo.FindMealLogByID(ctx, patientID, id)
	if err != nil {
		return nil, err
	}

	if req.MealType != "" {
		log.MealType = req.MealType
	}
	if req.PortionMultiplier > 0 {
		log.PortionMultiplier = req.PortionMultiplier
	}

	if err = s.repo.UpdateMealLog(ctx, log); err != nil {
		return nil, err
	}

	var fResp FoodResponse
	if log.Food != nil {
		fResp = ToFoodResponse(log.Food)
	}

	return &MealLogResponse{
		ID:                log.ID,
		Food:              fResp,
		MealType:          log.MealType,
		PortionMultiplier: log.PortionMultiplier,
		LoggedAt:          log.LoggedAt.Format(time.RFC3339),
	}, nil
}

func (s *nutritionService) DeleteMealLog(ctx context.Context, patientID string, id string) error {
	return s.repo.DeleteMealLog(ctx, patientID, id)
}

func (s *nutritionService) CalculateCalories(ctx context.Context, req CalorieCalculationRequest) (*CalorieCalculationResponse, error) {
	req.Normalize()
	return CalculateDailyCalories(req)
}
