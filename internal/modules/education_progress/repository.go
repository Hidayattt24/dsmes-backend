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
	type progressRow struct {
		ID               string     `gorm:"column:id"`
		PatientID        string     `gorm:"column:patient_id"`
		ArticleID        string     `gorm:"column:article_id"`
		ArticleRead      bool       `gorm:"column:article_read"`
		ArticleReadAt    *time.Time `gorm:"column:article_read_at"`
		YouTubeWatched   bool       `gorm:"column:youtube_watched"`
		YouTubeWatchedAt *time.Time `gorm:"column:youtube_watched_at"`
		CompletedAt      *time.Time `gorm:"column:completed_at"`
		CompletionSource string     `gorm:"column:completion_source"`
		CreatedAt        time.Time  `gorm:"column:created_at"`
		PatientName      string     `gorm:"column:patient_name"`
	}

	var rows []progressRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			uac.id,
			uac.patient_id,
			uac.article_id,
			uac.article_read,
			uac.article_read_at,
			uac.youtube_watched,
			uac.youtube_watched_at,
			uac.completed_at,
			uac.completion_source,
			uac.created_at,
			p.full_name AS patient_name
		FROM user_article_completions uac
		JOIN patients p ON p.id = uac.patient_id AND p.deleted_at IS NULL
		WHERE uac.article_id = ? AND uac.deleted_at IS NULL
		ORDER BY COALESCE(uac.completed_at, uac.created_at) DESC
	`, articleID).Scan(&rows).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch education progress", err)
	}

	items := make([]domain.UserArticleCompletion, len(rows))
	for i, r := range rows {
		items[i] = domain.UserArticleCompletion{
			BaseModel: domain.BaseModel{
				ID:        r.ID,
				CreatedAt: r.CreatedAt,
			},
			PatientID:        r.PatientID,
			ArticleID:        r.ArticleID,
			ArticleRead:      r.ArticleRead,
			ArticleReadAt:    r.ArticleReadAt,
			YouTubeWatched:   r.YouTubeWatched,
			YouTubeWatchedAt: r.YouTubeWatchedAt,
			CompletedAt:      r.CompletedAt,
			CompletionSource: r.CompletionSource,
		}
	}
	return items, nil
}

func (r *educationProgressRepository) FindByPatientAndArticle(ctx context.Context, patientID, articleID string) (*domain.UserArticleCompletion, error) {
	var item domain.UserArticleCompletion
	err := r.db.WithContext(ctx).
		Where("patient_id = ? AND article_id = ? AND deleted_at IS NULL", patientID, articleID).
		First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errs.NewInternal("failed to fetch patient progress", err)
	}
	return &item, nil
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
			COALESCE(SUM(CASE WHEN article_read = TRUE AND youtube_watched = TRUE THEN 1 ELSE 0 END), 0) AS read_and_video,
			COALESCE(SUM(CASE WHEN completed_at IS NOT NULL THEN 1 ELSE 0 END), 0)    AS completed,
			COALESCE(SUM(CASE WHEN article_read = FALSE AND youtube_watched = FALSE THEN 1 ELSE 0 END), 0) AS not_started
		FROM user_article_completions
		WHERE article_id = ? AND deleted_at IS NULL
	`, articleID).Scan(&row).Error
	if err != nil {
		return nil, errs.NewInternal("failed to compute education progress analytics", err)
	}

	return &ProgressAnalytics{
		TotalPatients:     row.TotalPatients,
		CompletedCount:    row.Completed,
		ReadArticleCount:  row.ArticleRead,
		WatchedVideoCount: row.YouTubeWatched,
		ReadAndVideoCount: row.ReadAndVideo,
		NotStartedCount:   row.NotStarted,
	}, nil
}
