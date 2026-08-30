package domain

// HealthFacility represents a Puskesmas (primary health care facility) master
// record. Patients are grouped under a facility via Patient.HealthFacility, and
// staff accounts are scoped to a facility via StaffAccount.HealthFacilityID.
type HealthFacility struct {
	BaseModel

	Name     string `gorm:"type:varchar(150);uniqueIndex:idx_health_facilities_name;not null" json:"name"`
	Address  string `gorm:"type:varchar(255)" json:"address"`
	IsActive bool   `gorm:"type:boolean;not null;default:true" json:"is_active"`
}

func (HealthFacility) TableName() string { return "health_facilities" }
