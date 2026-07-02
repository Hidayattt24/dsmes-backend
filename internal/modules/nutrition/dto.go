package nutrition

import "github.com/dsmes/dsmes-backend/internal/domain"

type CreateFoodRequest struct {
	Name                string  `json:"name"                  validate:"required,min=2,max=150"`
	DefaultServingUnit  string  `json:"default_serving_unit"`
	DefaultServingGrams float64 `json:"default_serving_grams" validate:"required,gt=0"`
	Calories            float64 `json:"calories"              validate:"required,gte=0"`
	CarbsG              float64 `json:"carbs_g"               validate:"gte=0"`
	ProteinG            float64 `json:"protein_g"             validate:"gte=0"`
	FatG                float64 `json:"fat_g"                 validate:"gte=0"`
}

type LogMealRequest struct {
	FoodID            string          `json:"food_id"            validate:"required,uuid4"`
	MealType          domain.MealType `json:"meal_type"          validate:"required,oneof=sarapan makan_siang makan_malam camilan"`
	PortionMultiplier float64         `json:"portion_multiplier" validate:"required,gt=0"`
}

type FoodResponse struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	DefaultServingUnit  string  `json:"default_serving_unit"`
	DefaultServingGrams float64 `json:"default_serving_grams"`
	Calories            float64 `json:"calories"`
	CarbsG              float64 `json:"carbs_g"`
	ProteinG            float64 `json:"protein_g"`
	FatG                float64 `json:"fat_g"`
}

type MealLogResponse struct {
	ID                string          `json:"id"`
	Food              FoodResponse    `json:"food"`
	MealType          domain.MealType `json:"meal_type"`
	PortionMultiplier float64         `json:"portion_multiplier"`
	LoggedAt          string          `json:"logged_at"`
}

type DailyNutritionSummaryResponse struct {
	CaloriesConsumed   float64 `json:"calories_consumed"`
	DailyCalorieTarget int     `json:"daily_calorie_target"`
	CaloriesRemaining  float64 `json:"calories_remaining"`
	TotalCarbsG        float64 `json:"total_carbs_g"`
	TotalProteinG      float64 `json:"total_protein_g"`
	TotalFatG          float64 `json:"total_fat_g"`
}

func ToFoodResponse(f *domain.Food) FoodResponse {
	return FoodResponse{
		ID:                  f.ID,
		Name:                f.Name,
		DefaultServingUnit:  f.DefaultServingUnit,
		DefaultServingGrams: f.DefaultServingGrams,
		Calories:            f.Calories,
		CarbsG:              f.CarbsG,
		ProteinG:            f.ProteinG,
		FatG:                f.FatG,
	}
}
