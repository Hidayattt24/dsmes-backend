package dashboard

import (
	"time"
)

type MonthData struct {
	Month string `json:"month"`
	Count int64  `json:"count"`
}

type AdminDashboardResponse struct {
	TotalPatients    int64       `json:"total_patients"`
	ActivePatients   int64       `json:"active_patients"`
	TotalPuskesmas   int64       `json:"total_puskesmas"`
	TotalArticles    int64       `json:"total_articles"`
	TotalQuizzes     int64       `json:"total_quizzes"`
	TotalSugarLogs   int64       `json:"total_sugar_logs"`
	PatientMonthly   []MonthData `json:"patient_monthly"`
	ArticleViewsMonthly []MonthData `json:"article_views_monthly"`
	SugarLogsMonthly []MonthData `json:"sugar_logs_monthly"`
}

type GlucoseDistribution struct {
	NormalCount      int64 `json:"normal_count"`
	TinggiCount      int64 `json:"tinggi_count"`
	SangatTinggiCount int64 `json:"sangat_tinggi_count"`
	RendahCount      int64 `json:"rendah_count"`
}

type PriorityPatient struct {
	ID               string    `json:"id"`
	FullName         string    `json:"full_name"`
	Nickname         string    `json:"nickname"`
	Email            string    `json:"email"`
	WhatsappNumber   string    `json:"whatsapp_number"`
	DiabetesType     string    `json:"diabetes_type"`
	Compliance       int       `json:"compliance"`
	LastActiveAt     *time.Time `json:"last_active_at"`
	PriorityReason   string    `json:"priority_reason"`
	LatestGlucose    *int      `json:"latest_glucose,omitempty"`
	LatestGlucoseStatus string `json:"latest_glucose_status,omitempty"`
}

type StaffDashboardResponse struct {
	TotalAssignedPatients  int64               `json:"total_assigned_patients"`
	ActiveAssignedPatients int64               `json:"active_assigned_patients"`
	TotalAttempts          int64               `json:"total_attempts"`
	GlucoseDistribution    GlucoseDistribution `json:"glucose_distribution"`
	PriorityPatients       []PriorityPatient   `json:"priority_patients"`
	NonCompliantPatients   []PriorityPatient   `json:"non_compliant_patients"`
}
