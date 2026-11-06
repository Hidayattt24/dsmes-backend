package education

import (
	"context"
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type PushSender interface {
	Send(ctx context.Context, token, title, body string, data map[string]string) (string, error)
}

type EducationRepository interface {
	FindAllCategories(ctx context.Context) ([]domain.ArticleCategory, error)
	FindCategoryByID(ctx context.Context, id string) (*domain.ArticleCategory, error)
	FindOrCreateCategoryByName(ctx context.Context, name string) (*domain.ArticleCategory, error)

	FindAllArticles(ctx context.Context, categoryID string, status *domain.ArticleStatus, page, limit int) ([]domain.Article, int64, error)
	FindArticleByID(ctx context.Context, id string) (*domain.Article, error)
	CreateArticle(ctx context.Context, a *domain.Article) (int64, error)
	UpdateArticle(ctx context.Context, a *domain.Article) (int64, error)
	PublishArticle(ctx context.Context, id string) (int64, error)
	DeleteArticle(ctx context.Context, id string) error
	GetStats(ctx context.Context) (*EducationStats, error)

	RecordView(ctx context.Context, view *domain.ArticleView) error
	MarkCompleted(ctx context.Context, completion *domain.UserArticleCompletion) error
	SaveArticle(ctx context.Context, patientID string, articleID string) error
	UnsaveArticle(ctx context.Context, patientID string, articleID string) error
	FindSavedArticles(ctx context.Context, patientID string) ([]domain.Article, error)
	GetPatientSavedMap(ctx context.Context, patientID string, articleIDs ...string) (map[string]bool, error)
	GetPatientCompletedMap(ctx context.Context, patientID string, articleIDs ...string) (map[string]bool, error)

	FindAllDeviceTokens(ctx context.Context) ([]domain.DeviceToken, error)
	DeleteDeviceToken(ctx context.Context, patientID string, token string) error

	// Transactional updates for sections/steps
	ReplaceSections(ctx context.Context, articleID string, sections []domain.ArticleSection) error

	// Reviews & Ratings
	UpsertReview(ctx context.Context, review *domain.EducationReview) error
	GetReviewByPatientAndArticle(ctx context.Context, patientID string, educationID string) (*domain.EducationReview, error)
	GetRatingSummary(ctx context.Context, educationID string) (average float64, count int64, dist RatingDistribution, err error)
	GetAdminReviews(ctx context.Context, educationID string) ([]domain.EducationReview, map[string]string, map[string]*time.Time, error)
}

type EducationService interface {
	ListCategories(ctx context.Context) ([]CategoryResponse, error)
	ListArticles(ctx context.Context, patientID *string, categoryID string, status *domain.ArticleStatus, page, limit int) ([]ArticleListResponse, int64, error)
	GetArticle(ctx context.Context, id string, patientID *string) (*ArticleDetailResponse, error)
	CreateArticle(ctx context.Context, staffID string, req CreateArticleRequest) (*ArticleDetailResponse, error)
	UpdateArticle(ctx context.Context, id string, req UpdateArticleRequest) (*ArticleDetailResponse, error)
	PublishArticle(ctx context.Context, id string) error
	DeleteArticle(ctx context.Context, id string) error
	GetStats(ctx context.Context) (*EducationStats, error)

	CompleteArticle(ctx context.Context, patientID string, id string) error
	SaveArticle(ctx context.Context, patientID string, id string) error
	UnsaveArticle(ctx context.Context, patientID string, id string) error
	ListSavedArticles(ctx context.Context, patientID string) ([]ArticleListResponse, error)

	// Reviews & Ratings
	SubmitReview(ctx context.Context, patientID string, educationID string, req CreateReviewRequest) (*EducationReviewResponse, error)
	GetPatientReview(ctx context.Context, patientID string, educationID string) (*EducationReviewResponse, error)
	GetRatingSummary(ctx context.Context, educationID string, patientID *string) (*ArticleRatingResponse, error)
	GetAdminReviews(ctx context.Context, educationID string) (*AdminArticleReviewsResponse, error)
}
