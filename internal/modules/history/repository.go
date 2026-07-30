package history

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type historyRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewHistoryRepository(db *gorm.DB, log *zap.Logger) HistoryRepository {
	return &historyRepository{db: db, log: log}
}

const historyUnionSQL = `
SELECT
	bs.id,
	bs.patient_id,
	'blood_sugar'::text AS activity_type,
	'Pemeriksaan Gula Darah'::text AS title,
	COALESCE(bs.recommendation, '')::text AS subtitle,
	'blood_sugar'::text AS category,
	CAST(bs.glucose_value AS VARCHAR) AS value,
	'mg/dL'::text AS unit,
	CAST(bs.status AS VARCHAR) AS status,
	''::text AS notes,
	to_char(bs.measured_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD"T"HH24:MI:SS"+07:00"') AS measured_at,
	to_char(bs.created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD"T"HH24:MI:SS"+07:00"') AS created_at,
	to_char(bs.updated_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD"T"HH24:MI:SS"+07:00"') AS updated_at,
	''::text AS recorded_by,
	'water_drop'::text AS icon,
	COALESCE(bs.color_indicator, '#00695C')::text AS color,
	bs.glucose_value::int AS glucose_value,
	CAST(bs.measurement_time_type AS VARCHAR) AS measurement_type,
	''::text AS meal_type,
	NULL::double precision AS calories,
	NULL::double precision AS carbs_g,
	NULL::double precision AS protein_g,
	NULL::double precision AS fat_g,
	NULL::int AS activity_minutes,
	''::text AS routine_type,
	''::text AS log_date,
	''::text AS reminder_name
FROM blood_sugar_logs bs
WHERE bs.patient_id = ? AND bs.deleted_at IS NULL

UNION ALL

SELECT
	ml.id,
	ml.patient_id,
	'meal'::text AS activity_type,
	COALESCE(f.name, 'Makanan')::text AS title,
	CAST(ml.meal_type AS VARCHAR) AS subtitle,
	CAST(ml.meal_type AS VARCHAR) AS category,
	CAST(ROUND(ml.portion_multiplier * COALESCE(f.calories, 0)) AS VARCHAR) AS value,
	'kcal'::text AS unit,
	'Selesai'::text AS status,
	''::text AS notes,
	to_char(ml.logged_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD"T"HH24:MI:SS"+07:00"') AS measured_at,
	to_char(ml.created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD"T"HH24:MI:SS"+07:00"') AS created_at,
	to_char(ml.updated_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD"T"HH24:MI:SS"+07:00"') AS updated_at,
	''::text AS recorded_by,
	'restaurant'::text AS icon,
	'#E65100'::text AS color,
	NULL::int AS glucose_value,
	''::text AS measurement_type,
	CAST(ml.meal_type AS VARCHAR) AS meal_type,
	(ml.portion_multiplier * COALESCE(f.calories, 0))::double precision AS calories,
	(ml.portion_multiplier * COALESCE(f.carbs_g, 0))::double precision AS carbs_g,
	(ml.portion_multiplier * COALESCE(f.protein_g, 0))::double precision AS protein_g,
	(ml.portion_multiplier * COALESCE(f.fat_g, 0))::double precision AS fat_g,
	NULL::int AS activity_minutes,
	''::text AS routine_type,
	''::text AS log_date,
	''::text AS reminder_name
FROM meal_logs ml
LEFT JOIN foods f ON f.id = ml.food_id
WHERE ml.patient_id = ? AND ml.deleted_at IS NULL

UNION ALL

SELECT
	rle.id,
	rle.patient_id,
	'activity'::text AS activity_type,
	COALESCE(r.descriptive_name, CAST(r.routine_type AS VARCHAR), 'Aktivitas Fisik')::text AS title,
	CAST(rle.status AS VARCHAR) AS subtitle,
	CAST(r.routine_type AS VARCHAR) AS category,
	'1'::text AS value,
	'aktivitas'::text AS unit,
	CAST(rle.status AS VARCHAR) AS status,
	''::text AS notes,
	to_char(rle.logged_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD"T"HH24:MI:SS"+07:00"') AS measured_at,
	to_char(rle.created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD"T"HH24:MI:SS"+07:00"') AS created_at,
	to_char(rle.updated_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD"T"HH24:MI:SS"+07:00"') AS updated_at,
	''::text AS recorded_by,
	'directions_run'::text AS icon,
	'#0284C7'::text AS color,
	NULL::int AS glucose_value,
	''::text AS measurement_type,
	''::text AS meal_type,
	NULL::double precision AS calories,
	NULL::double precision AS carbs_g,
	NULL::double precision AS protein_g,
	NULL::double precision AS fat_g,
	30::int AS activity_minutes,
	CAST(r.routine_type AS VARCHAR) AS routine_type,
	''::text AS log_date,
	''::text AS reminder_name
FROM routine_log_entries rle
LEFT JOIN routine_times rt ON rt.id = rle.routine_time_id
LEFT JOIN routines r ON r.id = rt.routine_id
WHERE rle.patient_id = ? AND rle.deleted_at IS NULL

UNION ALL

SELECT
	drl.id,
	rm.patient_id,
	'medication'::text AS activity_type,
	COALESCE(rm.activity_name, 'Obat-obatan')::text AS title,
	COALESCE(NULLIF(rm.notes, ''), CAST(drl.status AS VARCHAR))::text AS subtitle,
	CAST(rm.category AS VARCHAR) AS category,
	'1'::text AS value,
	'dosis'::text AS unit,
	CAST(drl.status AS VARCHAR) AS status,
	COALESCE(rm.notes, '')::text AS notes,
	to_char((drl.log_date::text || ' ' || COALESCE(NULLIF(rm.scheduled_time::text, ''), to_char(drl.created_at AT TIME ZONE 'Asia/Jakarta', 'HH24:MI:SS')))::timestamp, 'YYYY-MM-DD"T"HH24:MI:SS"+07:00"') AS measured_at,
	to_char(drl.created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD"T"HH24:MI:SS"+07:00"') AS created_at,
	to_char(drl.updated_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD"T"HH24:MI:SS"+07:00"') AS updated_at,
	''::text AS recorded_by,
	'medication'::text AS icon,
	'#6B21A8'::text AS color,
	NULL::int AS glucose_value,
	''::text AS measurement_type,
	''::text AS meal_type,
	NULL::double precision AS calories,
	NULL::double precision AS carbs_g,
	NULL::double precision AS protein_g,
	NULL::double precision AS fat_g,
	NULL::int AS activity_minutes,
	''::text AS routine_type,
	to_char(drl.log_date, 'YYYY-MM-DD') AS log_date,
	rm.activity_name::text AS reminder_name
FROM daily_reminder_logs drl
JOIN reminders rm ON rm.id = drl.reminder_id
WHERE rm.patient_id = ? AND drl.deleted_at IS NULL AND rm.deleted_at IS NULL

UNION ALL

SELECT
	pm.id,
	pm.patient_id,
	'measurement'::text AS activity_type,
	'Pengukuran Tubuh'::text AS title,
	COALESCE('BB: ' || CAST(ROUND(pm.weight_kg::numeric, 1) AS VARCHAR) || ' kg', '')::text AS subtitle,
	'measurement'::text AS category,
	CASE WHEN pm.blood_sugar IS NOT NULL AND pm.blood_sugar > 0 THEN CAST(ROUND(pm.blood_sugar::numeric, 0) AS VARCHAR) ELSE '-' END AS value,
	CASE WHEN pm.blood_sugar IS NOT NULL AND pm.blood_sugar > 0 THEN 'mg/dL'::text ELSE ''::text END AS unit,
	'Selesai'::text AS status,
	COALESCE(pm.notes, '')::text AS notes,
	to_char(pm.measured_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD"T"HH24:MI:SS"+07:00"') AS measured_at,
	to_char(pm.created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD"T"HH24:MI:SS"+07:00"') AS created_at,
	to_char(pm.updated_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD"T"HH24:MI:SS"+07:00"') AS updated_at,
	COALESCE(pm.recorded_by_name, '')::text AS recorded_by,
	'monitor_heart'::text AS icon,
	'#475569'::text AS color,
	CASE WHEN pm.blood_sugar IS NOT NULL AND pm.blood_sugar > 0 THEN CAST(ROUND(pm.blood_sugar::numeric, 0) AS INTEGER) ELSE NULL::int END AS glucose_value,
	''::text AS measurement_type,
	''::text AS meal_type,
	NULL::double precision AS calories,
	NULL::double precision AS carbs_g,
	NULL::double precision AS protein_g,
	NULL::double precision AS fat_g,
	NULL::int AS activity_minutes,
	''::text AS routine_type,
	''::text AS log_date,
	''::text AS reminder_name
FROM patient_measurements pm
WHERE pm.patient_id = ? AND pm.deleted_at IS NULL

UNION ALL

SELECT
	pal.id,
	pal.patient_id,
	'activity'::text AS activity_type,
	pal.activity_name::text AS title,
	pal.intensity::text AS subtitle,
	'activity'::text AS category,
	CAST(pal.duration_minutes AS VARCHAR) AS value,
	'menit'::text AS unit,
	'Selesai'::text AS status,
	COALESCE(pal.notes, '')::text AS notes,
	to_char(pal.logged_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD"T"HH24:MI:SS"+07:00"') AS measured_at,
	to_char(pal.created_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD"T"HH24:MI:SS"+07:00"') AS created_at,
	to_char(pal.updated_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD"T"HH24:MI:SS"+07:00"') AS updated_at,
	''::text AS recorded_by,
	'directions_walk'::text AS icon,
	'#388E3C'::text AS color,
	NULL::int AS glucose_value,
	''::text AS measurement_type,
	''::text AS meal_type,
	NULL::double precision AS calories,
	NULL::double precision AS carbs_g,
	NULL::double precision AS protein_g,
	NULL::double precision AS fat_g,
	pal.duration_minutes::int AS activity_minutes,
	''::text AS routine_type,
	''::text AS log_date,
	''::text AS reminder_name
FROM patient_activity_logs pal
WHERE pal.patient_id = ? AND pal.deleted_at IS NULL
`

func (r *historyRepository) FindAll(ctx context.Context, patientID string, page, limit int) ([]historyRawItem, int64, error) {
	offset := (page - 1) * limit

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS combined", historyUnionSQL)
	if err := r.db.WithContext(ctx).Raw(countSQL, patientID, patientID, patientID, patientID, patientID, patientID).Scan(&total).Error; err != nil {
		return nil, 0, errs.NewInternal("failed to count patient history", err)
	}

	var items []historyRawItem
	paginatedSQL := fmt.Sprintf(`
		SELECT * FROM (%s) AS combined
		ORDER BY measured_at DESC, created_at DESC
		OFFSET ? LIMIT ?
	`, historyUnionSQL)

	if err := r.db.WithContext(ctx).Raw(paginatedSQL, patientID, patientID, patientID, patientID, patientID, patientID, offset, limit).Scan(&items).Error; err != nil {
		return nil, 0, errs.NewInternal("failed to fetch patient history", err)
	}

	return items, total, nil
}

func (r *historyRepository) DeleteHistoryItem(ctx context.Context, patientID string, activityType string, id string) error {
	switch activityType {
	case "blood_sugar":
		result := r.db.WithContext(ctx).Exec("DELETE FROM blood_sugar_logs WHERE id = ? AND patient_id = ?", id, patientID)
		if result.Error != nil {
			return errs.NewInternal("failed to delete blood sugar log", result.Error)
		}
	case "meal":
		result := r.db.WithContext(ctx).Exec("DELETE FROM meal_logs WHERE id = ? AND patient_id = ?", id, patientID)
		if result.Error != nil {
			return errs.NewInternal("failed to delete meal log", result.Error)
		}
	case "activity":
		r.db.WithContext(ctx).Exec("DELETE FROM patient_activity_logs WHERE id = ? AND patient_id = ?", id, patientID)
		r.db.WithContext(ctx).Exec("DELETE FROM routine_log_entries WHERE id = ? AND patient_id = ?", id, patientID)
	case "medication":
		result := r.db.WithContext(ctx).Exec("DELETE FROM daily_reminder_logs WHERE id = ? AND reminder_id IN (SELECT id FROM reminders WHERE patient_id = ?)", id, patientID)
		if result.Error != nil {
			return errs.NewInternal("failed to delete medication log", result.Error)
		}
	case "measurement":
		result := r.db.WithContext(ctx).Exec("DELETE FROM patient_measurements WHERE id = ? AND patient_id = ?", id, patientID)
		if result.Error != nil {
			return errs.NewInternal("failed to delete measurement", result.Error)
		}
	default:
		return errs.NewBadRequest("invalid activity type")
	}
	return nil
}
