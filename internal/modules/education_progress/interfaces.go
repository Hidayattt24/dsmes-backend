package education_progress

import (
	"context"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type EducationProgressRepository interface {
	FindAllByArticle(ctx context.Context, articleID string) ([]domain.UserArticleCompletion, error)
	FindByPatientAndArticle(ctx context.Context, patientID, articleID string) (*domain.UserArticleCompletion, error)
	FindPatientNameAndPuskesmas(ctx context.Context, patientID string) (name string, puskesmas string)
	Upsert(ctx context.Context, progress *domain.UserArticleCompletion) error
	GetAnalytics(ctx context.Context, articleID string) (*ProgressAnalytics, error)
	LogActivity(ctx context.Context, patientID, articleID, activityType, metadata string) error
}

type EducationProgressService interface {
	GetArticleProgress(ctx context.Context, articleID string) ([]PatientProgressItem, error)
	GetArticleAnalytics(ctx context.Context, articleID string) (*ProgressAnalytics, error)
	MarkArticleRead(ctx context.Context, patientID, articleID string, readingDuration, lastScroll int) error
	MarkVideoWatched(ctx context.Context, patientID, articleID string, watchDuration, lastTimestamp int) error
	GetPatientProgress(ctx context.Context, patientID, articleID string) (*PatientProgressItem, error)
	LogActivity(ctx context.Context, patientID, articleID, activityType, metadata string) error
}
