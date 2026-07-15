package dashboard

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type dashboardRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewDashboardRepository(db *gorm.DB, log *zap.Logger) DashboardRepository {
	return &dashboardRepository{db: db, log: log}
}

func (r *dashboardRepository) GetAdminStats(ctx context.Context) (*AdminDashboardResponse, error) {
	var totalPatients int64
	var activePatients int64
	var totalStaff int64
	var totalArticles int64
	var totalQuizzes int64
	var totalSugarLogs int64

	if err := r.db.WithContext(ctx).Model(&domain.Patient{}).Where("deleted_at IS NULL").Count(&totalPatients).Error; err != nil {
		return nil, errs.NewInternal("failed to count patients", err)
	}

	if err := r.db.WithContext(ctx).Model(&domain.Patient{}).Where("status = ? AND deleted_at IS NULL", domain.StatusAktif).Count(&activePatients).Error; err != nil {
		return nil, errs.NewInternal("failed to count active patients", err)
	}

	if err := r.db.WithContext(ctx).Model(&domain.StaffAccount{}).Where("role = ? AND status = ? AND deleted_at IS NULL", domain.RoleStaff, domain.StatusAktif).Count(&totalStaff).Error; err != nil {
		return nil, errs.NewInternal("failed to count staff accounts", err)
	}

	if err := r.db.WithContext(ctx).Model(&domain.Article{}).Where("deleted_at IS NULL").Count(&totalArticles).Error; err != nil {
		return nil, errs.NewInternal("failed to count articles", err)
	}

	if err := r.db.WithContext(ctx).Model(&domain.Quiz{}).Where("deleted_at IS NULL").Count(&totalQuizzes).Error; err != nil {
		return nil, errs.NewInternal("failed to count quizzes", err)
	}

	if err := r.db.WithContext(ctx).Model(&domain.BloodSugarLog{}).Where("deleted_at IS NULL").Count(&totalSugarLogs).Error; err != nil {
		return nil, errs.NewInternal("failed to count blood sugar logs", err)
	}

	var totalMealLogs int64
	var totalActivityLogs int64
	var totalMedicationLogs int64
	_ = r.db.WithContext(ctx).Model(&domain.MealLog{}).Where("deleted_at IS NULL").Count(&totalMealLogs)
	_ = r.db.WithContext(ctx).Model(&domain.RoutineLogEntry{}).Where("deleted_at IS NULL").Count(&totalActivityLogs)
	_ = r.db.WithContext(ctx).Model(&domain.DailyReminderLog{}).Where("deleted_at IS NULL").Count(&totalMedicationLogs)

	var avgBloodSugar float64
	if err := r.db.WithContext(ctx).Model(&domain.BloodSugarLog{}).Where("deleted_at IS NULL").Select("COALESCE(AVG(glucose_value), 0)").Scan(&avgBloodSugar).Error; err != nil {
		r.log.Warn("failed to fetch average blood sugar log stats", zap.Error(err))
	}

	var todaySugarLogs int64
	var todayMealLogs int64
	var todayCheckins int64
	_ = r.db.WithContext(ctx).Model(&domain.BloodSugarLog{}).Where("DATE(measured_at) = CURRENT_DATE AND deleted_at IS NULL").Count(&todaySugarLogs)
	_ = r.db.WithContext(ctx).Model(&domain.MealLog{}).Where("DATE(logged_at) = CURRENT_DATE AND deleted_at IS NULL").Count(&todayMealLogs)
	_ = r.db.WithContext(ctx).Model(&domain.RoutineLogEntry{}).Where("DATE(logged_at) = CURRENT_DATE AND deleted_at IS NULL").Count(&todayCheckins)
	todayRecords := todaySugarLogs + todayMealLogs + todayCheckins

	var newRegistrations int64
	if err := r.db.WithContext(ctx).Model(&domain.Patient{}).Where("DATE(created_at) = CURRENT_DATE AND deleted_at IS NULL").Count(&newRegistrations).Error; err != nil {
		r.log.Warn("failed to fetch today new registrations", zap.Error(err))
	}

	// Fetch registration trends (last 6 months)
	var patientMonthly []MonthData
	err := r.db.WithContext(ctx).Raw(`
		SELECT TO_CHAR(created_at, 'YYYY-MM') as month, COUNT(*) as count 
		FROM patients 
		WHERE created_at >= NOW() - INTERVAL '6 months' AND deleted_at IS NULL
		GROUP BY TO_CHAR(created_at, 'YYYY-MM')
		ORDER BY month ASC
	`).Scan(&patientMonthly).Error
	if err != nil {
		r.log.Warn("failed to fetch monthly patient registration trends", zap.Error(err))
	}

	// Fetch article views trends (last 6 months)
	var articleViewsMonthly []MonthData
	err = r.db.WithContext(ctx).Raw(`
		SELECT TO_CHAR(viewed_at, 'YYYY-MM') as month, COUNT(*) as count 
		FROM article_views 
		WHERE viewed_at >= NOW() - INTERVAL '6 months' AND deleted_at IS NULL
		GROUP BY TO_CHAR(viewed_at, 'YYYY-MM')
		ORDER BY month ASC
	`).Scan(&articleViewsMonthly).Error
	if err != nil {
		r.log.Warn("failed to fetch monthly article views trends", zap.Error(err))
	}

	// Fetch sugar logs trends (last 6 months)
	var sugarLogsMonthly []MonthData
	err = r.db.WithContext(ctx).Raw(`
		SELECT TO_CHAR(measured_at, 'YYYY-MM') as month, COUNT(*) as count 
		FROM blood_sugar_logs 
		WHERE measured_at >= NOW() - INTERVAL '6 months' AND deleted_at IS NULL
		GROUP BY TO_CHAR(measured_at, 'YYYY-MM')
		ORDER BY month ASC
	`).Scan(&sugarLogsMonthly).Error
	if err != nil {
		r.log.Warn("failed to fetch monthly blood sugar logs trends", zap.Error(err))
	}

	return &AdminDashboardResponse{
		TotalPatients:       totalPatients,
		ActivePatients:      activePatients,
		TotalStaff:          totalStaff,
		TotalArticles:       totalArticles,
		TotalQuizzes:        totalQuizzes,
		TotalSugarLogs:      totalSugarLogs,
		TotalMealLogs:       totalMealLogs,
		TotalActivityLogs:   totalActivityLogs,
		TotalMedicationLogs: totalMedicationLogs,
		AverageBloodSugar:   avgBloodSugar,
		TodayRecords:        todayRecords,
		NewRegistrations:    newRegistrations,
		PatientMonthly:      patientMonthly,
		ArticleViewsMonthly: articleViewsMonthly,
		SugarLogsMonthly:    sugarLogsMonthly,
	}, nil
}

func (r *dashboardRepository) GetStaffStats(ctx context.Context, staffID string) (*StaffDashboardResponse, error) {
	var totalAssignedPatients int64
	var activeAssignedPatients int64
	var totalAttempts int64
	var avgBloodSugar float64
	var dist GlucoseDistribution

	if err := r.db.WithContext(ctx).Model(&domain.Patient{}).Where("assigned_staff_id = ? AND deleted_at IS NULL", staffID).Count(&totalAssignedPatients).Error; err != nil {
		return nil, errs.NewInternal("failed to count assigned patients", err)
	}

	if err := r.db.WithContext(ctx).Model(&domain.Patient{}).Where("assigned_staff_id = ? AND status = ? AND deleted_at IS NULL", staffID, domain.StatusAktif).Count(&activeAssignedPatients).Error; err != nil {
		return nil, errs.NewInternal("failed to count active assigned patients", err)
	}

	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(qa.id)
		FROM quiz_attempts qa
		JOIN patients p ON p.id = qa.patient_id
		WHERE p.assigned_staff_id = ? AND qa.deleted_at IS NULL AND p.deleted_at IS NULL
	`, staffID).Scan(&totalAttempts).Error
	if err != nil {
		return nil, errs.NewInternal("failed to count quiz attempts for staff", err)
	}

	var glucoseStats struct {
		AvgBloodSugar     float64
		NormalCount       int64
		TinggiCount       int64
		SangatTinggiCount int64
		RendahCount       int64
	}
	err = r.db.WithContext(ctx).Raw(`
		SELECT 
			COALESCE(AVG(b.glucose_value), 0) AS avg_blood_sugar,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'normal'), 0) AS normal_count,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'tinggi'), 0) AS tinggi_count,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'sangat_tinggi'), 0) AS sangat_tinggi_count,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'rendah'), 0) AS rendah_count
		FROM blood_sugar_logs b
		JOIN patients p ON p.id = b.patient_id
		WHERE p.assigned_staff_id = ? AND b.deleted_at IS NULL AND p.deleted_at IS NULL
	`, staffID).Scan(&glucoseStats).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch glucose stats", err)
	}
	avgBloodSugar = glucoseStats.AvgBloodSugar
	dist = GlucoseDistribution{
		NormalCount:       glucoseStats.NormalCount,
		TinggiCount:       glucoseStats.TinggiCount,
		SangatTinggiCount: glucoseStats.SangatTinggiCount,
		RendahCount:       glucoseStats.RendahCount,
	}

	totalGlucose := dist.NormalCount + dist.TinggiCount + dist.SangatTinggiCount + dist.RendahCount
	var stabilityPct float64
	if totalGlucose > 0 {
		stabilityPct = float64(dist.NormalCount) / float64(totalGlucose) * 100.0
	}

	// Fetch priority patients (sangat_tinggi or rendah in last 7 days)
	var priorityPatientsRaw []struct {
		ID                  string
		FullName            string
		Nickname            string
		Email               string
		WhatsappNumber      string
		DiabetesType        string
		Compliance          int
		LastActiveAt        *gorm.DeletedAt // map to nullable time
		GlucoseValue        int
		LatestGlucoseStatus string
	}

	err = r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (p.id) 
			p.id, p.full_name, p.nickname, p.email, p.whatsapp_number, p.diabetes_type, p.compliance, p.last_active_at,
			b.glucose_value, b.status AS latest_glucose_status
		FROM patients p
		JOIN blood_sugar_logs b ON b.patient_id = p.id
		WHERE p.assigned_staff_id = ? AND b.measured_at >= NOW() - INTERVAL '7 days' AND (b.status = 'sangat_tinggi' OR b.status = 'rendah') AND p.deleted_at IS NULL AND b.deleted_at IS NULL
		ORDER BY p.id, b.measured_at DESC
	`, staffID).Scan(&priorityPatientsRaw).Error

	var priorityPatients []PriorityPatient
	if err == nil {
		for _, p := range priorityPatientsRaw {
			var activeTime *time.Time
			if p.LastActiveAt != nil && p.LastActiveAt.Valid {
				activeTime = &p.LastActiveAt.Time
			}
			val := p.GlucoseValue
			priorityPatients = append(priorityPatients, PriorityPatient{
				ID:                  p.ID,
				FullName:            p.FullName,
				Nickname:            p.Nickname,
				Email:               p.Email,
				WhatsappNumber:      p.WhatsappNumber,
				DiabetesType:        p.DiabetesType,
				Compliance:          p.Compliance,
				LastActiveAt:        activeTime,
				PriorityReason:      "Log kadar gula darah berada di tingkat kritis (" + p.LatestGlucoseStatus + ") dalam 7 hari terakhir",
				LatestGlucose:       &val,
				LatestGlucoseStatus: p.LatestGlucoseStatus,
			})
		}
	}

	// Fetch non-compliant patients (compliance < 50 or inactive for last 3 days)
	var nonCompliantRaw []struct {
		ID             string
		FullName       string
		Nickname       string
		Email          string
		WhatsappNumber string
		DiabetesType   string
		Compliance     int
		LastActiveAt   *gorm.DeletedAt
	}

	err = r.db.WithContext(ctx).Raw(`
		SELECT p.id, p.full_name, p.nickname, p.email, p.whatsapp_number, p.diabetes_type, p.compliance, p.last_active_at
		FROM patients p
		WHERE p.assigned_staff_id = ? AND p.deleted_at IS NULL AND (p.compliance < 50 OR p.last_active_at IS NULL OR p.last_active_at < NOW() - INTERVAL '3 days')
		ORDER BY p.compliance ASC
		LIMIT 15
	`, staffID).Scan(&nonCompliantRaw).Error

	var nonCompliantPatients []PriorityPatient
	if err == nil {
		for _, p := range nonCompliantRaw {
			var activeTime *time.Time
			if p.LastActiveAt != nil && p.LastActiveAt.Valid {
				activeTime = &p.LastActiveAt.Time
			}

			reason := "Tingkat kepatuhan harian rendah (< 50%)"
			if p.LastActiveAt != nil && p.LastActiveAt.Valid && p.LastActiveAt.Time.Before(time.Now().AddDate(0, 0, -3)) {
				reason = "Tidak aktif melakukan pencatatan selama lebih dari 3 hari"
			} else if p.LastActiveAt == nil || !p.LastActiveAt.Valid {
				reason = "Belum pernah melakukan pencatatan di aplikasi"
			}

			nonCompliantPatients = append(nonCompliantPatients, PriorityPatient{
				ID:             p.ID,
				FullName:       p.FullName,
				Nickname:       p.Nickname,
				Email:          p.Email,
				WhatsappNumber: p.WhatsappNumber,
				DiabetesType:   p.DiabetesType,
				Compliance:     p.Compliance,
				LastActiveAt:   activeTime,
				PriorityReason: reason,
			})
		}
	}

	return &StaffDashboardResponse{
		TotalAssignedPatients:  totalAssignedPatients,
		ActiveAssignedPatients: activeAssignedPatients,
		TotalAttempts:          totalAttempts,
		AverageBloodSugar:      avgBloodSugar,
		StabilityPercentage:    stabilityPct,
		GlucoseDistribution:    dist,
		PriorityPatients:       priorityPatients,
		NonCompliantPatients:   nonCompliantPatients,
	}, nil
}

func (r *dashboardRepository) GetTopArticles(ctx context.Context) ([]TopArticleResponse, error) {
	var items []TopArticleResponse
	err := r.db.WithContext(ctx).Raw(`
		SELECT a.id, a.title, c.name as category, COUNT(v.id) as read_count, a.banner_image_url as thumbnail_url
		FROM articles a
		LEFT JOIN article_categories c ON c.id = a.category_id
		LEFT JOIN article_views v ON v.article_id = a.id AND v.deleted_at IS NULL
		WHERE a.deleted_at IS NULL
		GROUP BY a.id, c.name
		ORDER BY read_count DESC
		LIMIT 5
	`).Scan(&items).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch top articles", err)
	}
	return items, nil
}

func (r *dashboardRepository) GetActivityChart(ctx context.Context) ([]ActivityChartResponse, error) {
	type tempRes struct {
		Day   string
		Value int64
	}
	var raw []tempRes
	err := r.db.WithContext(ctx).Raw(`
		SELECT TO_CHAR(d, 'Dy') as day, COALESCE(SUM(cnt), 0) as value
		FROM (
			SELECT generate_series(NOW() - INTERVAL '6 days', NOW(), '1 day')::date as d
		) days
		LEFT JOIN (
			SELECT DATE(measured_at) as dt, COUNT(*) as cnt FROM blood_sugar_logs WHERE deleted_at IS NULL GROUP BY dt
			UNION ALL
			SELECT DATE(logged_at) as dt, COUNT(*) as cnt FROM meal_logs WHERE deleted_at IS NULL GROUP BY dt
			UNION ALL
			SELECT DATE(logged_at) as dt, COUNT(*) as cnt FROM routine_log_entries WHERE deleted_at IS NULL GROUP BY dt
		) activity ON activity.dt = days.d
		GROUP BY d
		ORDER BY d ASC
	`).Scan(&raw).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch activity chart", err)
	}

	dayMap := map[string]string{
		"Mon": "Sen",
		"Tue": "Sel",
		"Wed": "Rab",
		"Thu": "Kam",
		"Fri": "Jum",
		"Sat": "Sab",
		"Sun": "Min",
	}

	items := make([]ActivityChartResponse, len(raw))
	maxVal := int64(1)
	for _, r := range raw {
		if r.Value > maxVal {
			maxVal = r.Value
		}
	}

	for i, r := range raw {
		dayIndo := dayMap[r.Day]
		if dayIndo == "" {
			dayIndo = r.Day
		}
		items[i] = ActivityChartResponse{
			Day:           dayIndo,
			Value:         r.Value,
			HeightPercent: float64(r.Value) / float64(maxVal) * 100.0,
		}
		if items[i].HeightPercent < 10 {
			items[i].HeightPercent = 10
		}
	}
	return items, nil
}

func (r *dashboardRepository) GetPopulationMetrics(ctx context.Context, staffID string, rangeDays int) (*PopulationMetricsResponse, error) {
	type foodRow struct {
		MealType string
		Count    int64
	}
	var foodRows []foodRow
	sqlFood := fmt.Sprintf(`
		SELECT ml.meal_type, COUNT(*) AS count
		FROM meal_logs ml
		JOIN patients p ON p.id = ml.patient_id
		WHERE p.assigned_staff_id = ? AND ml.deleted_at IS NULL AND p.deleted_at IS NULL
		AND ml.logged_at >= NOW() - INTERVAL '%d days'
		GROUP BY ml.meal_type
		ORDER BY ml.meal_type
	`, rangeDays)
	if err := r.db.WithContext(ctx).Raw(sqlFood, staffID).Scan(&foodRows).Error; err != nil {
		return nil, errs.NewInternal("failed to fetch food intake distribution", err)
	}

	colorMap := map[string]string{
		"sarapan":     "#00695C",
		"makan_siang": "#10B981",
		"makan_malam": "#F59E0B",
		"camilan":     "#EF4444",
	}
	labelMap := map[string]string{
		"sarapan":     "Sarapan",
		"makan_siang": "Makan Siang",
		"makan_malam": "Makan Malam",
		"camilan":     "Cemilan",
	}

	var totalFood int64
	for _, f := range foodRows {
		totalFood += f.Count
	}

	foodIntake := make([]FoodIntakeItem, 0, len(foodRows))
	for _, f := range foodRows {
		pct := 0
		if totalFood > 0 {
			pct = int(float64(f.Count) / float64(totalFood) * 100.0)
		}
		label := labelMap[f.MealType]
		if label == "" {
			label = f.MealType
		}
		color := colorMap[f.MealType]
		if color == "" {
			color = "#718096"
		}
		foodIntake = append(foodIntake, FoodIntakeItem{
			Category:   label,
			Percentage: pct,
			Count:      f.Count,
			Color:      color,
		})
	}

	type activityRow struct {
		Level string
		Count int64
	}
	var activityRows []activityRow
	sqlActivity := fmt.Sprintf(`
		SELECT COALESCE(NULLIF(p.physical_activity_level, ''), 'Tidak Diketahui') AS level, COUNT(DISTINCT p.id) AS count
		FROM patients p
		JOIN routine_log_entries rle ON rle.patient_id = p.id
		WHERE p.assigned_staff_id = ? AND p.deleted_at IS NULL AND rle.deleted_at IS NULL
		AND rle.logged_at >= NOW() - INTERVAL '%d days'
		GROUP BY p.physical_activity_level
		ORDER BY count DESC
	`, rangeDays)
	if err := r.db.WithContext(ctx).Raw(sqlActivity, staffID).Scan(&activityRows).Error; err != nil {
		return nil, errs.NewInternal("failed to fetch physical activity distribution", err)
	}

	physicalActivity := make([]PhysicalActivityItem, 0, len(activityRows))
	for _, a := range activityRows {
		physicalActivity = append(physicalActivity, PhysicalActivityItem{
			Level: a.Level,
			Count: a.Count,
		})
	}

	type adhRow struct {
		Status string
		Count  int64
	}
	var adhRows []adhRow
	sqlAdh := fmt.Sprintf(`
		SELECT drl.status, COUNT(*) AS count
		FROM daily_reminder_logs drl
		JOIN reminders r ON r.id = drl.reminder_id
		JOIN patients p ON p.id = r.patient_id
		WHERE p.assigned_staff_id = ? AND drl.deleted_at IS NULL AND p.deleted_at IS NULL
		AND drl.log_date >= CURRENT_DATE - INTERVAL '%d days'
		GROUP BY drl.status
		ORDER BY drl.status
	`, rangeDays)
	if err := r.db.WithContext(ctx).Raw(sqlAdh, staffID).Scan(&adhRows).Error; err != nil {
		return nil, errs.NewInternal("failed to fetch medication adherence", err)
	}

	var totalAdh int64
	for _, a := range adhRows {
		totalAdh += a.Count
	}

	adhColorMap := map[string]string{
		"selesai":  "#00695C",
		"terlewat": "#EF4444",
		"pending":  "#F59E0B",
	}
	adhLabelMap := map[string]string{
		"selesai":  "Patuh",
		"terlewat": "Tidak Patuh",
		"pending":  "Kadang-kadang",
	}

	adherence := make([]AdherenceItem, 0, len(adhRows))
	for _, a := range adhRows {
		pct := 0
		if totalAdh > 0 {
			pct = int(float64(a.Count) / float64(totalAdh) * 100.0)
		}
		label := adhLabelMap[a.Status]
		if label == "" {
			label = a.Status
		}
		color := adhColorMap[a.Status]
		if color == "" {
			color = "#718096"
		}
		adherence = append(adherence, AdherenceItem{
			Label:      label,
			Percentage: pct,
			Count:      a.Count,
			Color:      color,
		})
	}

	return &PopulationMetricsResponse{
		FoodIntake:          foodIntake,
		PhysicalActivity:    physicalActivity,
		MedicationAdherence: adherence,
	}, nil
}

func (r *dashboardRepository) GetPatientTrends(ctx context.Context, staffID string, rangeDays int) ([]TrendPatient, error) {
	type trendRow struct {
		ID          string
		FullName    string
		Nickname    string
		AvgPrevious float64
		AvgCurrent  float64
	}

	var rows []trendRow
	halfDays := rangeDays / 2

	sql := fmt.Sprintf(`
		SELECT
			p.id,
			p.full_name,
			COALESCE(p.nickname, '') AS nickname,
			COALESCE(AVG(b.glucose_value) FILTER (WHERE b.measured_at >= NOW() - INTERVAL '%d days' AND b.measured_at < NOW() - INTERVAL '%d days'), 0) AS avg_previous,
			COALESCE(AVG(b.glucose_value) FILTER (WHERE b.measured_at >= NOW() - INTERVAL '%d days'), 0) AS avg_current
		FROM patients p
		JOIN blood_sugar_logs b ON b.patient_id = p.id
		WHERE p.assigned_staff_id = ? AND b.deleted_at IS NULL AND p.deleted_at IS NULL
		AND b.measured_at >= NOW() - INTERVAL '%d days'
		GROUP BY p.id, p.full_name, p.nickname
		HAVING
			COALESCE(AVG(b.glucose_value) FILTER (WHERE b.measured_at >= NOW() - INTERVAL '%d days'), 0) >
			COALESCE(AVG(b.glucose_value) FILTER (WHERE b.measured_at >= NOW() - INTERVAL '%d days' AND b.measured_at < NOW() - INTERVAL '%d days'), 0)
			AND COALESCE(AVG(b.glucose_value) FILTER (WHERE b.measured_at >= NOW() - INTERVAL '%d days' AND b.measured_at < NOW() - INTERVAL '%d days'), 0) > 0
		ORDER BY (COALESCE(AVG(b.glucose_value) FILTER (WHERE b.measured_at >= NOW() - INTERVAL '%d days'), 0) - COALESCE(AVG(b.glucose_value) FILTER (WHERE b.measured_at >= NOW() - INTERVAL '%d days' AND b.measured_at < NOW() - INTERVAL '%d days'), 0)) DESC
		LIMIT 10
	`,
		rangeDays, halfDays, halfDays, rangeDays, halfDays, rangeDays, halfDays, rangeDays, halfDays, halfDays, rangeDays, halfDays)
	err := r.db.WithContext(ctx).Raw(sql, staffID).Scan(&rows).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch patient trends", err)
	}

	trends := make([]TrendPatient, len(rows))
	for i, r := range rows {
		inc := r.AvgCurrent - r.AvgPrevious
		pct := 0.0
		if r.AvgPrevious > 0 {
			pct = inc / r.AvgPrevious * 100.0
		}
		trends[i] = TrendPatient{
			ID:                 r.ID,
			FullName:           r.FullName,
			Nickname:           r.Nickname,
			AvgStart:           r.AvgPrevious,
			AvgCurrent:         r.AvgCurrent,
			Increase:           inc,
			PercentageIncrease: pct,
		}
	}

	return trends, nil
}
