package education_progress

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type educationProgressRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewEducationProgressRepository(db *gorm.DB, log *zap.Logger) EducationProgressRepository {
	return &educationProgressRepository{db: db, log: log}
}

func (r *educationProgressRepository) FindAllByArticle(ctx context.Context, articleID string) ([]domain.UserArticleCompletion, error) {
	var items []domain.UserArticleCompletion
	err := r.db.WithContext(ctx).
		Where("article_id = ? AND deleted_at IS NULL", articleID).
		Order("COALESCE(completed_at, created_at) DESC").
		Find(&items).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch education progress", err)
	}
	return items, nil
}

func (r *educationProgressRepository) FindByPatientAndArticle(ctx context.Context, patientID, articleID string) (*domain.UserArticleCompletion, error) {
	var items []domain.UserArticleCompletion
	err := r.db.WithContext(ctx).
		Where("patient_id = ? AND article_id = ? AND deleted_at IS NULL", patientID, articleID).
		Limit(1).
		Find(&items).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch patient progress", err)
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func (r *educationProgressRepository) FindAllByPatient(ctx context.Context, patientID string) ([]domain.UserArticleCompletion, error) {
	var items []domain.UserArticleCompletion
	err := r.db.WithContext(ctx).
		Where("patient_id = ? AND deleted_at IS NULL", patientID).
		Order("COALESCE(updated_at, created_at) DESC").
		Find(&items).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch patient education activities", err)
	}
	return items, nil
}

func (r *educationProgressRepository) FindPatientNameAndPuskesmas(ctx context.Context, patientID string) (name string, puskesmas string) {
	var row struct {
		PatientName string `gorm:"column:patient_name"`
		Puskesmas   string `gorm:"column:puskesmas"`
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT p.full_name AS patient_name, COALESCE(sa.position_title, '-') AS puskesmas
		FROM patients p
		LEFT JOIN staff_accounts sa ON sa.id = p.assigned_staff_id AND sa.deleted_at IS NULL
		WHERE p.id = ? AND p.deleted_at IS NULL
	`, patientID).Scan(&row).Error
	if err != nil {
		return "", "-"
	}
	return row.PatientName, row.Puskesmas
}

func (r *educationProgressRepository) Upsert(ctx context.Context, progress *domain.UserArticleCompletion) error {
	progress.UpdatedAt = time.Now()

	if progress.ID != "" {
		err := r.db.WithContext(ctx).Save(progress).Error
		if err != nil {
			return errs.NewInternal("failed to update education progress", err)
		}
		return nil
	}

	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "patient_id"}, {Name: "article_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"article_read", "article_read_at",
			"article_started_at", "article_finished_at",
			"article_reading_duration", "article_last_scroll_position",
			"youtube_watched", "youtube_watched_at",
			"video_started_at", "video_finished_at",
			"video_watch_duration", "video_last_timestamp",
			"audio_listened", "audio_listened_at",
			"audio_started_at", "audio_finished_at",
			"audio_listen_duration", "audio_last_timestamp",
			"completed_at", "completion_source", "updated_at",
		}),
	}).Create(progress).Error
	if err != nil {
		return errs.NewInternal("failed to upsert education progress", err)
	}
	return nil
}

func (r *educationProgressRepository) LogActivity(ctx context.Context, patientID, articleID, activityType, metadata string) error {
	activity := &domain.PatientEducationActivity{
		PatientID:    patientID,
		ArticleID:    articleID,
		ActivityType: activityType,
		Metadata:     metadata,
	}
	if err := r.db.WithContext(ctx).Create(activity).Error; err != nil {
		return errs.NewInternal("failed to create patient education activity log", err)
	}
	return nil
}

func (r *educationProgressRepository) GetAnalytics(ctx context.Context, articleID string) (*ProgressAnalytics, error) {
	type analyticsRow struct {
		TotalPatients  int64
		ArticleRead    int64
		YouTubeWatched int64
		AudioListened  int64
		ReadAndVideo   int64
		Completed      int64
		NotStarted     int64
	}

	var row analyticsRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*)                                                                  AS total_patients,
			COALESCE(SUM(CASE WHEN article_read = TRUE THEN 1 ELSE 0 END), 0)         AS article_read,
			COALESCE(SUM(CASE WHEN youtube_watched = TRUE THEN 1 ELSE 0 END), 0)      AS youtube_watched,
			COALESCE(SUM(CASE WHEN audio_listened = TRUE THEN 1 ELSE 0 END), 0)       AS audio_listened,
			COALESCE(SUM(CASE WHEN article_read = TRUE AND youtube_watched = TRUE THEN 1 ELSE 0 END), 0) AS read_and_video,
			COALESCE(SUM(CASE WHEN completed_at IS NOT NULL THEN 1 ELSE 0 END), 0)    AS completed,
			COALESCE(SUM(CASE WHEN article_read = FALSE AND youtube_watched = FALSE AND audio_listened = FALSE THEN 1 ELSE 0 END), 0) AS not_started
		FROM user_article_completions
		WHERE article_id = ? AND deleted_at IS NULL
	`, articleID).Scan(&row).Error
	if err != nil {
		return nil, errs.NewInternal("failed to compute education progress analytics", err)
	}

	return &ProgressAnalytics{
		TotalPatients:      row.TotalPatients,
		CompletedCount:     row.Completed,
		ReadArticleCount:   row.ArticleRead,
		WatchedVideoCount:  row.YouTubeWatched,
		ListenedAudioCount: row.AudioListened,
		ReadAndVideoCount:  row.ReadAndVideo,
		NotStartedCount:    row.NotStarted,
	}, nil
}
