package domain

import "time"

// FoodMaster represents the master food item database.
// Table: food_master
type FoodMaster struct {
	BaseModel

	Name         string `gorm:"type:varchar(255);uniqueIndex:idx_food_unique_entry;not null" json:"name"`
	Manufacturer string `gorm:"type:varchar(255);uniqueIndex:idx_food_unique_entry;not null;default:''" json:"manufacturer"`
	ServingSize  string `gorm:"type:varchar(255);uniqueIndex:idx_food_unique_entry;not null;default:'1 porsi'" json:"serving_size"`

	// Nutrition Values (NUMERIC(10,2))
	Calories      float64 `gorm:"type:numeric(10,2);default:0" json:"calories"`
	EnergyKcal    float64 `gorm:"type:numeric(10,2);not null;default:0" json:"energy_kcal"`
	ProteinG      float64 `gorm:"type:numeric(10,2);not null;default:0" json:"protein_g"`
	CarbohydrateG float64 `gorm:"type:numeric(10,2);not null;default:0" json:"carbohydrate_g"`
	FatG          float64 `gorm:"type:numeric(10,2);not null;default:0" json:"fat_g"`
	SugarG        float64 `gorm:"type:numeric(10,2);default:0" json:"sugar_g"`
	SodiumMg      float64 `gorm:"type:numeric(10,2);default:0" json:"sodium_mg"`
	FiberG        float64 `gorm:"type:numeric(10,2);default:0" json:"fiber_g"`
	SaturatedFatG float64 `gorm:"type:numeric(10,2);default:0" json:"saturated_fat_g"`

	// Nutrition Daily Value %DV (NUMERIC(10,2))
	EnergyPercentageDV       float64 `gorm:"type:numeric(10,2);default:0" json:"energy_percentage_dv"`
	ProteinPercentageDV      float64 `gorm:"type:numeric(10,2);default:0" json:"protein_percentage_dv"`
	CarbohydratePercentageDV float64 `gorm:"type:numeric(10,2);default:0" json:"carbohydrate_percentage_dv"`
	FatPercentageDV          float64 `gorm:"type:numeric(10,2);default:0" json:"fat_percentage_dv"`
	SodiumPercentageDV       float64 `gorm:"type:numeric(10,2);default:0" json:"sodium_percentage_dv"`

	// Additional Label Values (NUMERIC(10,2))
	TotalFat          float64 `gorm:"type:numeric(10,2);default:0" json:"total_fat"`
	SaturatedFat      float64 `gorm:"type:numeric(10,2);default:0" json:"saturated_fat"`
	Sodium            float64 `gorm:"type:numeric(10,2);default:0" json:"sodium"`
	Protein           float64 `gorm:"type:numeric(10,2);default:0" json:"protein"`
	TotalCarbohydrate float64 `gorm:"type:numeric(10,2);default:0" json:"total_carbohydrate"`
	DietaryFiber      float64 `gorm:"type:numeric(10,2);default:0" json:"dietary_fiber"`
	Energy            float64 `gorm:"type:numeric(10,2);default:0" json:"energy"`

	// Metadata
	NutritionBasis string `gorm:"type:varchar(50);index;not null;default:'UNVERIFIED'" json:"nutrition_basis"`
	Source         string `gorm:"type:varchar(100);default:'manual'" json:"source"`
	Barcode        string `gorm:"type:varchar(100);index" json:"barcode"`
	ImageURL       string `gorm:"type:text" json:"image_url"`
	Status         string `gorm:"type:varchar(50);index;not null;default:'active'" json:"status"`
}

func (FoodMaster) TableName() string {
	return "foods"
}

// FoodRequest represents a user's proposal to add a new food item.
// Table: food_requests
type FoodRequest struct {
	BaseModel

	PatientID         string        `gorm:"type:uuid;not null;index" json:"patient_id"`
	Patient           *Patient      `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Name              string        `gorm:"type:varchar(255);not null" json:"name"`
	ServingSize       string        `gorm:"type:varchar(255);not null" json:"serving_size"`
	EstimatedCalories float64       `gorm:"type:numeric(10,2);default:0" json:"estimated_calories"`
	MealType          string        `gorm:"type:varchar(50);default:''" json:"meal_type"`
	Notes             string        `gorm:"type:text;default:''" json:"notes"`
	ImageURL          string        `gorm:"type:text;default:''" json:"image_url"`
	Status            string        `gorm:"type:varchar(50);index;not null;default:'pending'" json:"status"`
	AdminNotes        string        `gorm:"type:text;default:''" json:"admin_notes"`
	ReviewedBy        *string       `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	Reviewer          *StaffAccount `gorm:"foreignKey:ReviewedBy" json:"reviewer,omitempty"`
	ReviewedAt        *time.Time    `json:"reviewed_at,omitempty"`
	ApprovedFoodID    *string       `gorm:"type:uuid" json:"approved_food_id,omitempty"`
	ApprovedFood      *FoodMaster   `gorm:"foreignKey:ApprovedFoodID" json:"approved_food,omitempty"`
}

func (FoodRequest) TableName() string {
	return "food_requests"
}
