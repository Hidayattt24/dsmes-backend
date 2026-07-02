package domain

import (
	"time"
)

type TrendStatus string

const (
	TrendStabil   TrendStatus = "stabil"
	TrendMeningkat TrendStatus = "meningkat"
	TrendMenurun   TrendStatus = "menurun"
)

// WeeklyHealthSummary represents analytical weekly health reports.
type WeeklyHealthSummary struct {
	BaseModel

	PatientID           string      `gorm:"type:uuid;not null;uniqueIndex:idx_patient_week" json:"patient_id"`
	WeekStartDate       time.Time   `gorm:"type:date;not null;uniqueIndex:idx_patient_week" json:"week_start_date"`
	WeekEndDate         time.Time   `gorm:"type:date;not null;uniqueIndex:idx_patient_week" json:"week_end_date"`
	BloodSugarTrend     TrendStatus `gorm:"type:trend_status_enum" json:"blood_sugar_trend"`
	AvgBloodSugar       float64     `gorm:"type:numeric(6,2)" json:"avg_blood_sugar"`
	ArticlesReadCount   int         `gorm:"default:0" json:"articles_read_count"`
	ArticlesTargetCount int         `gorm:"default:7" json:"articles_target_count"`
	GeneratedAt         time.Time   `gorm:"not null;default:now()" json:"generated_at"`
}

func (WeeklyHealthSummary) TableName() string { return "weekly_health_summaries" }
