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

type UpdateMealLogRequest struct {
	MealType          domain.MealType `json:"meal_type"          validate:"omitempty,oneof=sarapan makan_siang makan_malam camilan"`
	PortionMultiplier float64         `json:"portion_multiplier" validate:"omitempty,gt=0"`
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
	TotalFoodMeal      int     `json:"total_food_today"`
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

type CalorieCalculationRequest struct {
	Gender        string  `json:"gender"          validate:"required"`
	DateOfBirth   string  `json:"date_of_birth"`
	Age           int     `json:"age"`
	HeightCm      float64 `json:"height_cm"`
	Height        float64 `json:"height"`
	WeightKg      float64 `json:"weight_kg"`
	Weight        float64 `json:"weight"`
	ActivityLevel string  `json:"activity_level" validate:"required"`
	BloodType     string  `json:"blood_type"`
}

func (r *CalorieCalculationRequest) Normalize() {
	if r.HeightCm <= 0 && r.Height > 0 {
		r.HeightCm = r.Height
	}
	if r.WeightKg <= 0 && r.Weight > 0 {
		r.WeightKg = r.Weight
	}
}

type CalorieRecommendationDetail struct {
	Title        string `json:"title"`
	WeeklyTarget string `json:"weekly_target,omitempty"`
	Calories     int    `json:"calories"`
	Percentage   int    `json:"percentage"`
}

type CalorieRecommendations struct {
	Maintain    CalorieRecommendationDetail `json:"maintain"`
	MildLoss    CalorieRecommendationDetail `json:"mild_loss"`
	WeightLoss  CalorieRecommendationDetail `json:"weight_loss"`
	ExtremeLoss CalorieRecommendationDetail `json:"extreme_loss"`
}

type RecommendedCalories struct {
	WeightLoss  int `json:"weightLoss"`
	Maintenance int `json:"maintenance"`
	WeightGain  int `json:"weightGain"`
}

type CalorieCalculationResponse struct {
	Age                 int                    `json:"age"`
	Gender              string                 `json:"gender"`
	Height              float64                `json:"height"`
	HeightCm            float64                `json:"height_cm"`
	Weight              float64                `json:"weight"`
	WeightKg            float64                `json:"weight_kg"`
	BMI                 float64                `json:"bmi"`
	BMICategory         string                 `json:"bmi_category"`
	BMR                 int                    `json:"bmr"`
	ActivityMultiplier  float64                `json:"activityMultiplier"`
	TDEE                int                    `json:"tdee"`
	RecommendedCalories RecommendedCalories    `json:"recommendedCalories"`
	Recommendations     CalorieRecommendations `json:"recommendations"`
}
