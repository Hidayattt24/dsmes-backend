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
	BloodTypeA         BloodType = "A"
	BloodTypeB         BloodType = "B"
	BloodTypeAB        BloodType = "AB"
	BloodTypeO         BloodType = "O"
	BloodTypeTidakTahu BloodType = "tidak_tahu"
)

// DefaultDailyCalorieTarget is the fallback calorie target used when a patient
// has no computed recommendation yet.
const DefaultDailyCalorieTarget = 2000

// Patient represents the patient entity in the DSMES Aceh system.
type Patient struct {
	BaseModel

	PhoneNumber               string        `gorm:"type:varchar(20);uniqueIndex:idx_patients_phone_number;not null" json:"phone_number"`
	Email                     *string       `gorm:"type:varchar(150);uniqueIndex:idx_patients_email" json:"email"`
	PasswordHash              string        `gorm:"type:varchar(255);not null" json:"-"`
	FullName                  string        `gorm:"type:varchar(150);not null" json:"full_name"`
	Nickname                  string        `gorm:"type:varchar(50)" json:"nickname"`
	WhatsappNumber            string        `gorm:"type:varchar(20)" json:"whatsapp_number"`
	Gender                    Gender        `gorm:"type:gender_enum;not null" json:"gender"`
	DateOfBirth               time.Time     `gorm:"type:date;not null" json:"date_of_birth"`
	HeightCm                  float64       `gorm:"type:numeric(5,2)" json:"height_cm"`
	WeightKg                  float64       `gorm:"type:numeric(5,2)" json:"weight_kg"`
	BloodType                 BloodType     `gorm:"type:blood_type_enum;default:tidak_tahu" json:"blood_type"`
	DailyCalorieTarget        int           `gorm:"type:int;not null;default:2000" json:"daily_calorie_target"`
	MaintenanceCalories       int           `gorm:"type:int" json:"maintenance_calories"`
	MildWeightLossCalories    int           `gorm:"type:int" json:"mild_weight_loss_calories"`
	WeightLossCalories        int           `gorm:"type:int" json:"weight_loss_calories"`
	ExtremeWeightLossCalories int           `gorm:"type:int" json:"extreme_weight_loss_calories"`
	MaintenancePercentage     int           `gorm:"type:int" json:"maintenance_percentage"`
	MildPercentage            int           `gorm:"type:int" json:"mild_percentage"`
	WeightLossPercentage      int           `gorm:"type:int" json:"weight_loss_percentage"`
	ExtremePercentage         int           `gorm:"type:int" json:"extreme_percentage"`
	MedicalStatus             string        `gorm:"type:varchar(100)" json:"medical_status"`
	ProfilePhotoURL           string        `gorm:"type:text" json:"profile_photo_url"`
	Status                    AccountStatus `gorm:"type:account_status_enum;not null;default:aktif" json:"status"`
	AssignedStaffID           *string       `gorm:"type:uuid;column:assigned_staff_id" json:"assigned_staff_id"`
	LastActiveAt              *time.Time    `json:"last_active_at"`
	BPJS                      string        `gorm:"type:varchar(50)" json:"bpjs"`
	NIK                       string        `gorm:"type:varchar(50)" json:"nik"`
	EmergencyName             string        `gorm:"type:varchar(150)" json:"emergency_name"`
	EmergencyRelation         string        `gorm:"type:varchar(100)" json:"emergency_relation"`
	EmergencyPhone            string        `gorm:"type:varchar(20)" json:"emergency_phone"`
	DiabetesType              string        `gorm:"type:varchar(50)" json:"diabetes_type"`
	Compliance                int           `gorm:"type:int;default:0" json:"compliance"`
	InterventionType          string        `gorm:"type:varchar(50)" json:"intervention_type"`
	PatientCode               string        `gorm:"type:varchar(50)" json:"patient_code"`
	Address                   string        `gorm:"type:text" json:"address"`
	DiagnosisDate             *time.Time    `gorm:"type:date" json:"diagnosis_date"`
	CurrentMedication         string        `gorm:"type:text" json:"current_medication"`
	Allergies                 string        `gorm:"type:text" json:"allergies"`
	SmokingStatus             string        `gorm:"type:varchar(50)" json:"smoking_status"`
	PhysicalActivityLevel     string        `gorm:"type:varchar(50)" json:"physical_activity_level"`

	// Relations
	AssignedStaff *StaffAccount `gorm:"foreignKey:AssignedStaffID" json:"assigned_staff,omitempty"`
}

func (Patient) TableName() string { return "patients" }

func (p *Patient) GetEmail() string {
	if p.Email != nil {
		return *p.Email
	}
	return ""
}
