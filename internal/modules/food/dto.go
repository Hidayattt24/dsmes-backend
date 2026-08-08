package food

import "github.com/dsmes/dsmes-backend/internal/domain"

type CreateFoodRequest struct {
	Name         string  `json:"name"          validate:"required,min=2,max=255"`
	Manufacturer string  `json:"manufacturer"  validate:"required"`
	ServingSize  string  `json:"serving_size"  validate:"required"`
	EnergyKcal   float64 `json:"energy_kcal"   validate:"gte=0"`
	ProteinG     float64 `json:"protein_g"     validate:"gte=0"`
	CarbohydrateG float64 `json:"carbohydrate_g" validate:"gte=0"`
	FatG         float64 `json:"fat_g"         validate:"gte=0"`
	SugarG       float64 `json:"sugar_g"       validate:"gte=0"`
	SodiumMg     float64 `json:"sodium_mg"     validate:"gte=0"`
	FiberG       float64 `json:"fiber_g"       validate:"gte=0"`
	SaturatedFatG float64 `json:"saturated_fat_g" validate:"gte=0"`

	EnergyPercentageDV       float64 `json:"energy_percentage_dv" validate:"gte=0"`
	ProteinPercentageDV      float64 `json:"protein_percentage_dv" validate:"gte=0"`
	CarbohydratePercentageDV float64 `json:"carbohydrate_percentage_dv" validate:"gte=0"`
	FatPercentageDV          float64 `json:"fat_percentage_dv" validate:"gte=0"`
	SodiumPercentageDV       float64 `json:"sodium_percentage_dv" validate:"gte=0"`

	TotalFat          float64 `json:"total_fat" validate:"gte=0"`
	SaturatedFat      float64 `json:"saturated_fat" validate:"gte=0"`
	Sodium            float64 `json:"sodium" validate:"gte=0"`
	Protein           float64 `json:"protein" validate:"gte=0"`
	TotalCarbohydrate float64 `json:"total_carbohydrate" validate:"gte=0"`
	DietaryFiber      float64 `json:"dietary_fiber" validate:"gte=0"`
	Energy            float64 `json:"energy" validate:"gte=0"`

	NutritionBasis string `json:"nutrition_basis" validate:"omitempty,oneof=PER_100G PER_SERVING PER_PACKAGE"`
	Source         string `json:"source"`
	Barcode        string `json:"barcode"`
	ImageURL       string `json:"image_url"`
	Status         string `json:"status" validate:"omitempty,oneof=active inactive"`
}

type UpdateFoodRequest struct {
	Name         string  `json:"name"          validate:"omitempty,min=2,max=255"`
	Manufacturer string  `json:"manufacturer"`
	ServingSize  string  `json:"serving_size"`
	EnergyKcal   float64 `json:"energy_kcal"   validate:"gte=0"`
	ProteinG     float64 `json:"protein_g"     validate:"gte=0"`
	CarbohydrateG float64 `json:"carbohydrate_g" validate:"gte=0"`
	FatG         float64 `json:"fat_g"         validate:"gte=0"`
	SugarG       float64 `json:"sugar_g"       validate:"gte=0"`
	SodiumMg     float64 `json:"sodium_mg"     validate:"gte=0"`
	FiberG       float64 `json:"fiber_g"       validate:"gte=0"`
	SaturatedFatG float64 `json:"saturated_fat_g" validate:"gte=0"`

	EnergyPercentageDV       float64 `json:"energy_percentage_dv" validate:"gte=0"`
	ProteinPercentageDV      float64 `json:"protein_percentage_dv" validate:"gte=0"`
	CarbohydratePercentageDV float64 `json:"carbohydrate_percentage_dv" validate:"gte=0"`
	FatPercentageDV          float64 `json:"fat_percentage_dv" validate:"gte=0"`
	SodiumPercentageDV       float64 `json:"sodium_percentage_dv" validate:"gte=0"`

	TotalFat          float64 `json:"total_fat" validate:"gte=0"`
	SaturatedFat      float64 `json:"saturated_fat" validate:"gte=0"`
	Sodium            float64 `json:"sodium" validate:"gte=0"`
	Protein           float64 `json:"protein" validate:"gte=0"`
	TotalCarbohydrate float64 `json:"total_carbohydrate" validate:"gte=0"`
	DietaryFiber      float64 `json:"dietary_fiber" validate:"gte=0"`
	Energy            float64 `json:"energy" validate:"gte=0"`

	NutritionBasis string `json:"nutrition_basis" validate:"omitempty,oneof=PER_100G PER_SERVING PER_PACKAGE"`
	Source         string `json:"source"`
	Barcode        string `json:"barcode"`
	ImageURL       string `json:"image_url"`
	Status         string `json:"status" validate:"omitempty,oneof=active inactive"`
}

type FoodFilterQuery struct {
	Q            string  `query:"q"`
	Manufacturer string  `query:"manufacturer"`
	MinCalories  float64 `query:"min_calories"`
	MaxCalories  float64 `query:"max_calories"`
	Status       string  `query:"status"`
	SortBy       string  `query:"sort_by"`
	SortOrder    string  `query:"sort_order"`
	Page         int     `query:"page"`
	Limit        int     `query:"limit"`
}

type FoodMasterResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Manufacturer string  `json:"manufacturer"`
	ServingSize  string  `json:"serving_size"`

	EnergyKcal    float64 `json:"energy_kcal"`
	ProteinG      float64 `json:"protein_g"`
	CarbohydrateG float64 `json:"carbohydrate_g"`
	FatG          float64 `json:"fat_g"`
	SugarG        float64 `json:"sugar_g"`
	SodiumMg      float64 `json:"sodium_mg"`
	FiberG        float64 `json:"fiber_g"`
	SaturatedFatG float64 `json:"saturated_fat_g"`

	EnergyPercentageDV       float64 `json:"energy_percentage_dv"`
	ProteinPercentageDV      float64 `json:"protein_percentage_dv"`
	CarbohydratePercentageDV float64 `json:"carbohydrate_percentage_dv"`
	FatPercentageDV          float64 `json:"fat_percentage_dv"`
	SodiumPercentageDV       float64 `json:"sodium_percentage_dv"`

	TotalFat          float64 `json:"total_fat"`
	SaturatedFat      float64 `json:"saturated_fat"`
	Sodium            float64 `json:"sodium"`
	Protein           float64 `json:"protein"`
	TotalCarbohydrate float64 `json:"total_carbohydrate"`
	DietaryFiber      float64 `json:"dietary_fiber"`
	Energy            float64 `json:"energy"`

	NutritionBasis string `json:"nutrition_basis"`
	Source         string `json:"source"`
	Barcode        string `json:"barcode"`
	ImageURL       string `json:"image_url"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

func ToFoodMasterResponse(f *domain.FoodMaster) FoodMasterResponse {
	nutritionBasis := f.NutritionBasis
	if nutritionBasis == "" {
		nutritionBasis = "PER_100G"
	}
	return FoodMasterResponse{
		ID:           f.ID,
		Name:         f.Name,
		Manufacturer: f.Manufacturer,
		ServingSize:  f.ServingSize,

		EnergyKcal:    f.EnergyKcal,
		ProteinG:      f.ProteinG,
		CarbohydrateG: f.CarbohydrateG,
		FatG:          f.FatG,
		SugarG:        f.SugarG,
		SodiumMg:      f.SodiumMg,
		FiberG:        f.FiberG,
		SaturatedFatG: f.SaturatedFatG,

		EnergyPercentageDV:       f.EnergyPercentageDV,
		ProteinPercentageDV:      f.ProteinPercentageDV,
		CarbohydratePercentageDV: f.CarbohydratePercentageDV,
		FatPercentageDV:          f.FatPercentageDV,
		SodiumPercentageDV:       f.SodiumPercentageDV,

		TotalFat:          f.TotalFat,
		SaturatedFat:      f.SaturatedFat,
		Sodium:            f.Sodium,
		Protein:           f.Protein,
		TotalCarbohydrate: f.TotalCarbohydrate,
		DietaryFiber:      f.DietaryFiber,
		Energy:            f.Energy,

		NutritionBasis: nutritionBasis,
		Source:         f.Source,
		Barcode:        f.Barcode,
		ImageURL:       f.ImageURL,
		Status:         f.Status,
		CreatedAt:      f.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      f.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

type ExcelImportRow struct {
	RowIndex int               `json:"row_index"`
	IsValid  bool              `json:"is_valid"`
	Errors   []string          `json:"errors"`
	Data     CreateFoodRequest `json:"data"`
}

type ExcelImportPreviewResponse struct {
	TotalRows   int              `json:"total_rows"`
	ValidRows   int              `json:"valid_rows"`
	InvalidRows int              `json:"invalid_rows"`
	Rows        []ExcelImportRow `json:"rows"`
}

type ExcelImportConfirmRequest struct {
	Items []CreateFoodRequest `json:"items" validate:"required,min=1"`
}

type ExcelImportConfirmResponse struct {
	SuccessCount int `json:"success_count"`
	FailedCount  int `json:"failed_count"`
}

type FoodStatsResponse struct {
	TotalFoods         int64 `json:"total_foods"`
	TodayImportedFoods int64 `json:"today_imported_foods"`
	TotalManufacturers int64 `json:"total_manufacturers"`
	ActiveFoods        int64 `json:"active_foods"`
}
