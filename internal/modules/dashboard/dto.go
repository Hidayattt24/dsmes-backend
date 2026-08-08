package dashboard

import (
	"time"
)

type MonthData struct {
	Month string `json:"month"`
	Count int64  `json:"count"`
}

type AdminDashboardResponse struct {
	TotalPatients       int64       `json:"total_patients"`
	ActivePatients      int64       `json:"active_patients"`
	TotalStaff          int64       `json:"total_staff"`
	TotalArticles       int64       `json:"total_articles"`
	TotalQuizzes        int64       `json:"total_quizzes"`
	TotalSugarLogs      int64       `json:"total_sugar_logs"`
	TotalMealLogs       int64       `json:"total_meal_logs"`
	TotalActivityLogs   int64       `json:"total_activity_logs"`
	TotalMedicationLogs int64       `json:"total_medication_logs"`
	AverageBloodSugar   float64     `json:"average_blood_sugar"`
	TodayRecords        int64       `json:"today_records"`
	NewRegistrations    int64       `json:"new_registrations"`
	PatientMonthly      []MonthData `json:"patient_monthly"`
	ArticleViewsMonthly []MonthData `json:"article_views_monthly"`
	SugarLogsMonthly    []MonthData `json:"sugar_logs_monthly"`
}

type GlucoseDistribution struct {
	HypoglycemiaCount  int64 `json:"hypoglycemia_count"`
	NormalCount        int64 `json:"normal_count"`
	PrediabetesCount   int64 `json:"prediabetes_count"`
	ElevatedCount      int64 `json:"elevated_count"`
	HyperglycemiaCount int64 `json:"hyperglycemia_count"`
}

type PriorityPatient struct {
	ID                  string     `json:"id"`
	FullName            string     `json:"full_name"`
	Nickname            string     `json:"nickname"`
	Email               string     `json:"email"`
	WhatsappNumber      string     `json:"whatsapp_number"`
	DiabetesType        string     `json:"diabetes_type"`
	Compliance          int        `json:"compliance"`
	LastActiveAt        *time.Time `json:"last_active_at"`
	PriorityReason      string     `json:"priority_reason"`
	LatestGlucose       *int       `json:"latest_glucose,omitempty"`
	LatestGlucoseStatus string     `json:"latest_glucose_status,omitempty"`
}

type StaffDashboardResponse struct {
	TotalAssignedPatients  int64               `json:"total_assigned_patients"`
	ActiveAssignedPatients int64               `json:"active_assigned_patients"`
	TotalAttempts          int64               `json:"total_attempts"`
	AverageBloodSugar      float64             `json:"average_blood_sugar"`
	StabilityPercentage    float64             `json:"stability_percentage"`
	GlucoseDistribution    GlucoseDistribution `json:"glucose_distribution"`
	PriorityPatients       []PriorityPatient   `json:"priority_patients"`
	NonCompliantPatients   []PriorityPatient   `json:"non_compliant_patients"`
	TotalSugarLogs         int64               `json:"total_sugar_logs"`
	TotalMealLogs          int64               `json:"total_meal_logs"`
	TotalActivityLogs      int64               `json:"total_activity_logs"`
	TotalMedicationLogs    int64               `json:"total_medication_logs"`
}

type TopArticleResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Category     string `json:"category"`
	ReadCount    int64  `json:"read_count"`
	ThumbnailURL string `json:"thumbnail_url"`
}

type ActivityChartResponse struct {
	Day           string  `json:"day"`
	Value         int64   `json:"value"`
	HeightPercent float64 `json:"height_percent"`
}

type FoodIntakeItem struct {
	Category   string `json:"category"`
	Percentage int    `json:"percentage"`
	Count      int64  `json:"count"`
	Color      string `json:"color"`
}

type PhysicalActivityItem struct {
	Level string `json:"level"`
	Count int64  `json:"count"`
}

type AdherenceItem struct {
	Label      string `json:"label"`
	Percentage int    `json:"percentage"`
	Count      int64  `json:"count"`
	Color      string `json:"color"`
}

type PopulationMetricsResponse struct {
	FoodIntake          []FoodIntakeItem       `json:"food_intake"`
	PhysicalActivity    []PhysicalActivityItem `json:"physical_activity"`
	MedicationAdherence []AdherenceItem        `json:"medication_adherence"`
	FoodPatients        []PatientContribution  `json:"food_patients"`
	ActivityPatients    []PatientContribution  `json:"activity_patients"`
	MedicationPatients  []PatientContribution  `json:"medication_patients"`
}

// PatientContribution lists a patient and how many logs they contributed within
// the dashboard range. Used to show "which patients have data" under each card.
type PatientContribution struct {
	PatientID string `json:"patient_id"`
	FullName  string `json:"full_name"`
	Count     int64  `json:"count"`
}

type TrendPatient struct {
	ID                 string  `json:"id"`
	FullName           string  `json:"full_name"`
	Nickname           string  `json:"nickname"`
	AvgStart           float64 `json:"avg_start"`
	AvgCurrent         float64 `json:"avg_current"`
	Increase           float64 `json:"increase"`
	PercentageIncrease float64 `json:"percentage_increase"`
}
