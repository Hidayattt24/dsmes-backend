package domain

import (
	"time"
)

type Gender string

const (
	GenderLakiLaki  Gender = "laki_laki"
	GenderPerempuan Gender = "perempuan"
)

type BloodType string

const (
	BloodTypeA        BloodType = "A"
	BloodTypeB        BloodType = "B"
	BloodTypeAB       BloodType = "AB"
	BloodTypeO        BloodType = "O"
	BloodTypeTidakTahu BloodType = "tidak_tahu"
)

// Patient represents the patient entity in the DSMES Aceh system.
type Patient struct {
	BaseModel

	Email               string        `gorm:"type:varchar(150);uniqueIndex;not null" json:"email"`
	PasswordHash        string        `gorm:"type:varchar(255);not null" json:"-"`
	FullName            string        `gorm:"type:varchar(150);not null" json:"full_name"`
	Nickname            string        `gorm:"type:varchar(50)" json:"nickname"`
	WhatsappNumber      string        `gorm:"type:varchar(20);not null" json:"whatsapp_number"`
	Gender              Gender        `gorm:"type:gender_enum;not null" json:"gender"`
	DateOfBirth         time.Time     `gorm:"type:date;not null" json:"date_of_birth"`
	HeightCm            float64       `gorm:"type:numeric(5,2)" json:"height_cm"`
	WeightKg            float64       `gorm:"type:numeric(5,2)" json:"weight_kg"`
	BloodType           BloodType     `gorm:"type:blood_type_enum;default:tidak_tahu" json:"blood_type"`
	DailyCalorieTarget  int           `gorm:"type:int;not null;default:2000" json:"daily_calorie_target"`
	MedicalStatus       string        `gorm:"type:varchar(100)" json:"medical_status"`
	ProfilePhotoURL     string        `gorm:"type:text" json:"profile_photo_url"`
	Status              AccountStatus `gorm:"type:account_status_enum;not null;default:aktif" json:"status"`
	AssignedPuskesmasID *string       `gorm:"type:uuid" json:"assigned_puskesmas_id"`
	LastActiveAt        *time.Time    `json:"last_active_at"`
	BPJS                string        `gorm:"type:varchar(50)" json:"bpjs"`
	NIK                 string        `gorm:"type:varchar(50)" json:"nik"`
	EmergencyName       string        `gorm:"type:varchar(150)" json:"emergency_name"`
	EmergencyRelation   string        `gorm:"type:varchar(100)" json:"emergency_relation"`
	EmergencyPhone      string        `gorm:"type:varchar(20)" json:"emergency_phone"`
	DiabetesType        string        `gorm:"type:varchar(50)" json:"diabetes_type"`
	Compliance          int           `gorm:"type:int;default:0" json:"compliance"`
	InterventionType    string        `gorm:"type:varchar(50)" json:"intervention_type"`

	// Relations
	AssignedPuskesmas *StaffAccount `gorm:"foreignKey:AssignedPuskesmasID" json:"assigned_puskesmas,omitempty"`
}

func (Patient) TableName() string { return "patients" }
