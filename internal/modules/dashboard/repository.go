package dashboard

import (
	"context"
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
		PatientMonthly:      patientMonthly,
		ArticleViewsMonthly: articleViewsMonthly,
		SugarLogsMonthly:    sugarLogsMonthly,
	}, nil
}

func (r *dashboardRepository) GetStaffStats(ctx context.Context, staffID string) (*StaffDashboardResponse, error) {
	var totalAssignedPatients int64
	var activeAssignedPatients int64
	var totalAttempts int64
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

	err = r.db.WithContext(ctx).Raw(`
		SELECT 
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'normal'), 0) AS normal_count,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'tinggi'), 0) AS tinggi_count,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'sangat_tinggi'), 0) AS sangat_tinggi_count,
			COALESCE(COUNT(*) FILTER (WHERE b.status = 'rendah'), 0) AS rendah_count
		FROM blood_sugar_logs b
		JOIN patients p ON p.id = b.patient_id
		WHERE p.assigned_staff_id = ? AND b.deleted_at IS NULL AND p.deleted_at IS NULL
	`, staffID).Scan(&dist).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch glucose distribution", err)
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
			if p.LastActiveAt.Valid {
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
			if p.LastActiveAt.Valid {
				activeTime = &p.LastActiveAt.Time
			}

			reason := "Tingkat kepatuhan harian rendah (< 50%)"
			if p.LastActiveAt.Valid && p.LastActiveAt.Time.Before(time.Now().AddDate(0, 0, -3)) {
				reason = "Tidak aktif melakukan pencatatan selama lebih dari 3 hari"
			} else if !p.LastActiveAt.Valid {
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
		GlucoseDistribution:    dist,
		PriorityPatients:       priorityPatients,
		NonCompliantPatients:   nonCompliantPatients,
	}, nil
}
