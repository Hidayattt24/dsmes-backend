package domain

import "time"

type MealType string

const (
	MealSarapan    MealType = "sarapan"
	MealMakanSiang MealType = "makan_siang"
	MealMakanMalam MealType = "makan_malam"
	MealCamilan    MealType = "camilan"
)

// Food represents a food item database.
type Food struct {
	BaseModel

	Name                string  `gorm:"type:varchar(150);index;not null" json:"name"`
	DefaultServingUnit  string  `gorm:"type:varchar(50)" json:"default_serving_unit"`
	DefaultServingGrams float64 `gorm:"type:numeric(6,2)" json:"default_serving_grams"`
	Calories            float64 `gorm:"type:numeric(6,2);not null" json:"calories"`
	CarbsG              float64 `gorm:"type:numeric(6,2)" json:"carbs_g"`
	ProteinG            float64 `gorm:"type:numeric(6,2)" json:"protein_g"`
	FatG                float64 `gorm:"type:numeric(6,2)" json:"fat_g"`
}

func (Food) TableName() string { return "foods" }

// MealLog represents a patient's meal log.
type MealLog struct {
	BaseModel

	PatientID         string    `gorm:"type:uuid;not null" json:"patient_id"`
	FoodID            string    `gorm:"type:uuid;not null" json:"food_id"`
	MealType          MealType  `gorm:"type:meal_type_enum;not null" json:"meal_type"`
	PortionMultiplier float64   `gorm:"type:numeric(4,2);not null;default:1.0" json:"portion_multiplier"`
	LoggedAt          time.Time `gorm:"not null;default:now()" json:"logged_at"`

	// Nutrition Snapshot (frozen at logging time)
	FoodName     string  `gorm:"type:varchar(255)" json:"food_name"`
	ServingSize  string  `gorm:"type:varchar(100)" json:"serving_size"`
	Calories     float64 `gorm:"type:numeric(8,2)" json:"calories"`
	CarbsG       float64 `gorm:"type:numeric(8,2)" json:"carbs_g"`
	ProteinG     float64 `gorm:"type:numeric(8,2)" json:"protein_g"`
	FatG         float64 `gorm:"type:numeric(8,2)" json:"fat_g"`
	SugarG       float64 `gorm:"type:numeric(8,2)" json:"sugar_g"`
	SodiumMg     float64 `gorm:"type:numeric(8,2)" json:"sodium_mg"`
	FiberG       float64 `gorm:"type:numeric(8,2)" json:"fiber_g"`

	// Relations
	Food       *Food       `gorm:"foreignKey:FoodID" json:"food,omitempty"`
	FoodMaster *FoodMaster `gorm:"foreignKey:FoodID" json:"food_master,omitempty"`
}

func (MealLog) TableName() string { return "meal_logs" }

// RecentFoodSearch tracks a patient's food search history.
type RecentFoodSearch struct {
	BaseModel

	PatientID  string    `gorm:"type:uuid;not null;uniqueIndex:idx_patient_food" json:"patient_id"`
	FoodID     string    `gorm:"type:uuid;not null;uniqueIndex:idx_patient_food" json:"food_id"`
	UsageCount int       `gorm:"not null;default:1" json:"usage_count"`
	LastUsedAt time.Time `gorm:"not null;default:now()" json:"last_used_at"`

	// Relations
	Food       *Food       `gorm:"foreignKey:FoodID" json:"food,omitempty"`
	FoodMaster *FoodMaster `gorm:"foreignKey:FoodID" json:"food_master,omitempty"`
}

func (RecentFoodSearch) TableName() string { return "recent_food_searches" }
