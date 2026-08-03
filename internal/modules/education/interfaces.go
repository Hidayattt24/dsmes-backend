package education

import (
	"context"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type EducationRepository interface {
	FindAllCategories(ctx context.Context) ([]domain.ArticleCategory, error)
	FindCategoryByID(ctx context.Context, id string) (*domain.ArticleCategory, error)
	FindOrCreateCategoryByName(ctx context.Context, name string) (*domain.ArticleCategory, error)

	FindAllArticles(ctx context.Context, categoryID string, status *domain.ArticleStatus, page, limit int) ([]domain.Article, int64, error)
	FindArticleByID(ctx context.Context, id string) (*domain.Article, error)
	CreateArticle(ctx context.Context, a *domain.Article) error
	UpdateArticle(ctx context.Context, a *domain.Article) error
	PublishArticle(ctx context.Context, id string) error
	DeleteArticle(ctx context.Context, id string) error
	GetStats(ctx context.Context) (*EducationStats, error)

	RecordView(ctx context.Context, view *domain.ArticleView) error
	MarkCompleted(ctx context.Context, completion *domain.UserArticleCompletion) error
	SaveArticle(ctx context.Context, patientID string, articleID string) error
	UnsaveArticle(ctx context.Context, patientID string, articleID string) error
	FindSavedArticles(ctx context.Context, patientID string) ([]domain.Article, error)
	GetPatientSavedMap(ctx context.Context, patientID string) (map[string]bool, error)
	GetPatientCompletedMap(ctx context.Context, patientID string) (map[string]bool, error)

	// BroadcastEducationNotification inserts an education notification into
	// every active patient's notification_logs inbox.
	BroadcastEducationNotification(ctx context.Context, message string, articleID string) error

	// Transactional updates for sections/steps
	ReplaceSections(ctx context.Context, articleID string, sections []domain.ArticleSection) error
}

type EducationService interface {
	ListCategories(ctx context.Context) ([]CategoryResponse, error)
	ListArticles(ctx context.Context, patientID *string, categoryID string, status *domain.ArticleStatus, page, limit int) ([]ArticleListResponse, int64, error)
	GetArticle(ctx context.Context, id string, patientID *string) (*ArticleDetailResponse, error)
	CreateArticle(ctx context.Context, staffID string, req CreateArticleRequest) (*ArticleDetailResponse, error)
	UpdateArticle(ctx context.Context, id string, req CreateArticleRequest) (*ArticleDetailResponse, error)
	PublishArticle(ctx context.Context, id string) error
	DeleteArticle(ctx context.Context, id string) error
	GetStats(ctx context.Context) (*EducationStats, error)

	CompleteArticle(ctx context.Context, patientID string, id string) error
	SaveArticle(ctx context.Context, patientID string, id string) error
	UnsaveArticle(ctx context.Context, patientID string, id string) error
	ListSavedArticles(ctx context.Context, patientID string) ([]ArticleListResponse, error)
}
