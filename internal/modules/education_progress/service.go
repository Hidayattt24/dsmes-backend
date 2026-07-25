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

func (s *educationProgressService) MarkArticleRead(ctx context.Context, patientID, articleID string) error {
	_, err := s.eduRepo.FindArticleByID(ctx, articleID)
	if err != nil {
		return err
	}

	existing, err := s.repo.FindByPatientAndArticle(ctx, patientID, articleID)
	if err != nil {
		return err
	}

	now := time.Now()

	if existing != nil {
		existing.ArticleRead = true
		existing.ArticleReadAt = &now
		if existing.YouTubeWatched && existing.CompletedAt == nil {
			existing.CompletedAt = &now
		} else if !existing.YouTubeWatched {
			existing.CompletedAt = &now
		}
		return s.repo.Upsert(ctx, existing)
	}

	completedAt := &now
	progress := &domain.UserArticleCompletion{
		PatientID:     patientID,
		ArticleID:     articleID,
		ArticleRead:   true,
		ArticleReadAt: &now,
		CompletedAt:   completedAt,
	}

	return s.repo.Upsert(ctx, progress)
}

func (s *educationProgressService) MarkVideoWatched(ctx context.Context, patientID, articleID string) error {
	_, err := s.eduRepo.FindArticleByID(ctx, articleID)
	if err != nil {
		return err
	}

	existing, err := s.repo.FindByPatientAndArticle(ctx, patientID, articleID)
	if err != nil {
		return err
	}

	now := time.Now()

	if existing != nil {
		existing.YouTubeWatched = true
		existing.YouTubeWatchedAt = &now
		if existing.ArticleRead && existing.CompletedAt == nil {
			existing.CompletedAt = &now
		} else if !existing.ArticleRead {
			existing.CompletedAt = &now
		}
		return s.repo.Upsert(ctx, existing)
	}

	completedAt := &now
	progress := &domain.UserArticleCompletion{
		PatientID:        patientID,
		ArticleID:        articleID,
		YouTubeWatched:   true,
		YouTubeWatchedAt: &now,
		CompletedAt:      completedAt,
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

func toProgressItem(r domain.UserArticleCompletion) PatientProgressItem {
	item := PatientProgressItem{
		PatientID:      r.PatientID,
		ArticleRead:    r.ArticleRead,
		YouTubeWatched: r.YouTubeWatched,
		Completed:      r.CompletedAt != nil,
	}

	if r.ArticleReadAt != nil {
		val := fmtTime(*r.ArticleReadAt)
		item.ArticleReadAt = &val
	}
	if r.YouTubeWatchedAt != nil {
		val := fmtTime(*r.YouTubeWatchedAt)
		item.YouTubeWatchedAt = &val
	}
	if r.CompletedAt != nil {
		val := fmtTime(*r.CompletedAt)
		item.CompletedAt = &val
	}

	// Last activity is the most recent among the two
	if r.ArticleReadAt != nil && r.YouTubeWatchedAt != nil {
		if r.ArticleReadAt.After(*r.YouTubeWatchedAt) {
			val := fmtTime(*r.ArticleReadAt)
			item.LastActivityAt = &val
		} else {
			val := fmtTime(*r.YouTubeWatchedAt)
			item.LastActivityAt = &val
		}
	} else if r.ArticleReadAt != nil {
		val := fmtTime(*r.ArticleReadAt)
		item.LastActivityAt = &val
	} else if r.YouTubeWatchedAt != nil {
		val := fmtTime(*r.YouTubeWatchedAt)
		item.LastActivityAt = &val
	}

	return item
}
