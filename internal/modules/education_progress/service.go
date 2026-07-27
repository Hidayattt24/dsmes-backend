package education_progress

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/modules/education"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type educationProgressService struct {
	repo    EducationProgressRepository
	eduRepo education.EducationRepository
	log     *zap.Logger
}

func NewEducationProgressService(
	repo EducationProgressRepository,
	eduRepo education.EducationRepository,
	log *zap.Logger,
) EducationProgressService {
	return &educationProgressService{
		repo:    repo,
		eduRepo: eduRepo,
		log:     log,
	}
}

func fmtTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05Z07:00")
}

func (s *educationProgressService) GetArticleProgress(ctx context.Context, articleID string) ([]PatientProgressItem, error) {
	_, err := s.eduRepo.FindArticleByID(ctx, articleID)
	if err != nil {
		return nil, err
	}

	records, err := s.repo.FindAllByArticle(ctx, articleID)
	if err != nil {
		return nil, err
	}

	items := make([]PatientProgressItem, 0, len(records))
	for _, r := range records {
		item := toProgressItem(r)
		name, puskesmas := s.repo.FindPatientNameAndPuskesmas(ctx, r.PatientID)
		item.PatientName = name
		item.Puskesmas = puskesmas
		items = append(items, item)
	}
	return items, nil
}

func (s *educationProgressService) GetArticleAnalytics(ctx context.Context, articleID string) (*ProgressAnalytics, error) {
	_, err := s.eduRepo.FindArticleByID(ctx, articleID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetAnalytics(ctx, articleID)
}

func (s *educationProgressService) MarkArticleRead(ctx context.Context, patientID, articleID string, readingDuration, lastScroll int) error {
	_, err := s.eduRepo.FindArticleByID(ctx, articleID)
	if err != nil {
		return err
	}

	existing, err := s.repo.FindByPatientAndArticle(ctx, patientID, articleID)
	if err != nil {
		return err
	}

	now := time.Now()
	isCompleting := true

	if existing != nil {
		if existing.ArticleReadingDuration == 0 {
			_ = s.repo.LogActivity(ctx, patientID, articleID, "OPEN_ARTICLE", `{"status":"started"}`)
		}

		existing.ArticleReadingDuration += readingDuration
		if lastScroll > existing.ArticleLastScrollPosition {
			existing.ArticleLastScrollPosition = lastScroll
		}

		if !existing.ArticleRead {
			existing.ArticleRead = true
			existing.ArticleReadAt = &now
			existing.ArticleFinishedAt = &now
			_ = s.repo.LogActivity(ctx, patientID, articleID, "COMPLETE_READ", `{"status":"completed"}`)
		}

		if (existing.ArticleRead || existing.YouTubeWatched) && existing.CompletedAt == nil {
			existing.CompletedAt = &now
			if existing.CompletionSource == "" {
				if existing.ArticleRead {
					existing.CompletionSource = "ARTICLE"
				} else {
					existing.CompletionSource = "VIDEO"
				}
			}
		}

		return s.repo.Upsert(ctx, existing)
	}

	_ = s.repo.LogActivity(ctx, patientID, articleID, "OPEN_ARTICLE", `{"status":"started"}`)

	var completedAt *time.Time
	var readAt *time.Time
	var finishedAt *time.Time
	var source string

	if isCompleting {
		readAt = &now
		finishedAt = &now
		completedAt = &now
		source = "ARTICLE"
		_ = s.repo.LogActivity(ctx, patientID, articleID, "COMPLETE_READ", `{"status":"completed"}`)
	}

	progress := &domain.UserArticleCompletion{
		PatientID:                 patientID,
		ArticleID:                 articleID,
		ArticleRead:               isCompleting,
		ArticleReadAt:             readAt,
		ArticleStartedAt:          &now,
		ArticleFinishedAt:         finishedAt,
		ArticleReadingDuration:    readingDuration,
		ArticleLastScrollPosition: lastScroll,
		CompletedAt:               completedAt,
		CompletionSource:          source,
	}

	return s.repo.Upsert(ctx, progress)
}

func (s *educationProgressService) MarkVideoWatched(ctx context.Context, patientID, articleID string, watchDuration, lastTimestamp int) error {
	_, err := s.eduRepo.FindArticleByID(ctx, articleID)
	if err != nil {
		return err
	}

	existing, err := s.repo.FindByPatientAndArticle(ctx, patientID, articleID)
	if err != nil {
		return err
	}

	now := time.Now()
	isCompleting := true

	if existing != nil {
		if existing.VideoWatchDuration == 0 {
			_ = s.repo.LogActivity(ctx, patientID, articleID, "OPEN_VIDEO", `{"status":"started"}`)
		}

		existing.VideoWatchDuration += watchDuration
		existing.VideoLastTimestamp = lastTimestamp

		if !existing.YouTubeWatched {
			existing.YouTubeWatched = true
			existing.YouTubeWatchedAt = &now
			existing.VideoFinishedAt = &now
			_ = s.repo.LogActivity(ctx, patientID, articleID, "COMPLETE_WATCH", `{"status":"completed"}`)
		}

		if (existing.ArticleRead || existing.YouTubeWatched) && existing.CompletedAt == nil {
			existing.CompletedAt = &now
			if existing.CompletionSource == "" {
				if existing.YouTubeWatched {
					existing.CompletionSource = "VIDEO"
				} else {
					existing.CompletionSource = "ARTICLE"
				}
			}
		}

		return s.repo.Upsert(ctx, existing)
	}

	_ = s.repo.LogActivity(ctx, patientID, articleID, "OPEN_VIDEO", `{"status":"started"}`)

	var completedAt *time.Time
	var watchedAt *time.Time
	var finishedAt *time.Time
	var source string

	if isCompleting {
		watchedAt = &now
		finishedAt = &now
		completedAt = &now
		source = "VIDEO"
		_ = s.repo.LogActivity(ctx, patientID, articleID, "COMPLETE_WATCH", `{"status":"completed"}`)
	}

	progress := &domain.UserArticleCompletion{
		PatientID:          patientID,
		ArticleID:          articleID,
		YouTubeWatched:     isCompleting,
		YouTubeWatchedAt:   watchedAt,
		VideoStartedAt:     &now,
		VideoFinishedAt:    finishedAt,
		VideoWatchDuration: watchDuration,
		VideoLastTimestamp: lastTimestamp,
		CompletedAt:        completedAt,
		CompletionSource:   source,
	}

	return s.repo.Upsert(ctx, progress)
}

func (s *educationProgressService) GetPatientProgress(ctx context.Context, patientID, articleID string) (*PatientProgressItem, error) {
	record, err := s.repo.FindByPatientAndArticle(ctx, patientID, articleID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, errs.NewNotFound("education progress not found for this patient")
	}

	item := toProgressItem(*record)
	name, puskesmas := s.repo.FindPatientNameAndPuskesmas(ctx, patientID)
	item.PatientName = name
	item.Puskesmas = puskesmas
	return &item, nil
}

func (s *educationProgressService) LogActivity(ctx context.Context, patientID, articleID, activityType, metadata string) error {
	return s.repo.LogActivity(ctx, patientID, articleID, activityType, metadata)
}

func toProgressItem(r domain.UserArticleCompletion) PatientProgressItem {
	item := PatientProgressItem{
		PatientID:                 r.PatientID,
		ArticleRead:               r.ArticleRead,
		ArticleReadingDuration:    r.ArticleReadingDuration,
		ArticleLastScrollPosition: r.ArticleLastScrollPosition,
		YouTubeWatched:            r.YouTubeWatched,
		VideoWatchDuration:        r.VideoWatchDuration,
		VideoLastTimestamp:        r.VideoLastTimestamp,
		Completed:                 r.CompletedAt != nil,
		CompletionSource:          r.CompletionSource,
	}

	if r.ArticleReadAt != nil {
		val := fmtTime(*r.ArticleReadAt)
		item.ArticleReadAt = &val
	}
	if r.ArticleStartedAt != nil {
		val := fmtTime(*r.ArticleStartedAt)
		item.ArticleStartedAt = &val
	}
	if r.ArticleFinishedAt != nil {
		val := fmtTime(*r.ArticleFinishedAt)
		item.ArticleFinishedAt = &val
	}
	if r.YouTubeWatchedAt != nil {
		val := fmtTime(*r.YouTubeWatchedAt)
		item.YouTubeWatchedAt = &val
	}
	if r.VideoStartedAt != nil {
		val := fmtTime(*r.VideoStartedAt)
		item.VideoStartedAt = &val
	}
	if r.VideoFinishedAt != nil {
		val := fmtTime(*r.VideoFinishedAt)
		item.VideoFinishedAt = &val
	}
	if r.CompletedAt != nil {
		val := fmtTime(*r.CompletedAt)
		item.CompletedAt = &val
	}

	var lastTime *time.Time
	if r.ArticleReadAt != nil {
		lastTime = r.ArticleReadAt
	}
	if r.YouTubeWatchedAt != nil {
		if lastTime == nil || r.YouTubeWatchedAt.After(*lastTime) {
			lastTime = r.YouTubeWatchedAt
		}
	}
	if r.UpdatedAt.After(r.CreatedAt) {
		if lastTime == nil || r.UpdatedAt.After(*lastTime) {
			lastTime = &r.UpdatedAt
		}
	}

	if lastTime != nil {
		val := fmtTime(*lastTime)
		item.LastActivityAt = &val
	}

	return item
}
