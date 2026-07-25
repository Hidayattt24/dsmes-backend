package education

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type educationRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewEducationRepository(db *gorm.DB, log *zap.Logger) EducationRepository {
	return &educationRepository{db: db, log: log}
}

func (r *educationRepository) FindAllCategories(ctx context.Context) ([]domain.ArticleCategory, error) {
	var items []domain.ArticleCategory
	err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Find(&items).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch categories", err)
	}
	return items, nil
}

func (r *educationRepository) FindCategoryByID(ctx context.Context, id string) (*domain.ArticleCategory, error) {
	var c domain.ArticleCategory
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&c).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("category not found")
		}
		return nil, errs.NewInternal("failed to fetch category", err)
	}
	return &c, nil
}
func (r *educationRepository) FindOrCreateCategoryByName(ctx context.Context, name string) (*domain.ArticleCategory, error) {
	var c domain.ArticleCategory
	err := r.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", name).First(&c).Error
	if err == nil {
		return &c, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NewInternal("failed to query category by name", err)
	}

	// Create new category
	c = domain.ArticleCategory{
		Name: name,
	}
	if err := r.db.WithContext(ctx).Create(&c).Error; err != nil {
		return nil, errs.NewInternal("failed to create category dynamically", err)
	}
	return &c, nil
}

func (r *educationRepository) FindAllArticles(ctx context.Context, categoryID string, status *domain.ArticleStatus, page, limit int) ([]domain.Article, int64, error) {
	var items []domain.Article
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.Article{}).Where("deleted_at IS NULL")

	if categoryID != "" {
		q = q.Where("category_id = ?", categoryID)
	}

	if status != nil {
		q = q.Where("status = ?", *status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errs.NewInternal("failed to count articles", err)
	}

	offset := (page - 1) * limit
	err := q.Select("articles.*, COALESCE(views.count, 0) as read_count").
		Joins("LEFT JOIN (SELECT article_id, COUNT(*) as count FROM article_views WHERE deleted_at IS NULL GROUP BY article_id) views ON views.article_id = articles.id").
		Preload("Category").
		Offset(offset).Limit(limit).
		Order("articles.created_at DESC").
		Find(&items).Error
	if err != nil {
		return nil, 0, errs.NewInternal("failed to fetch articles", err)
	}

	return items, total, nil
}

func (r *educationRepository) FindArticleByID(ctx context.Context, id string) (*domain.Article, error) {
	var a domain.Article
	err := r.db.WithContext(ctx).
		Select("articles.*, COALESCE(views.count, 0) as read_count").
		Joins("LEFT JOIN (SELECT article_id, COUNT(*) as count FROM article_views WHERE deleted_at IS NULL GROUP BY article_id) views ON views.article_id = articles.id").
		Preload("Category").
		Preload("ArticleSections", func(db *gorm.DB) *gorm.DB {
			return db.Order("article_sections.section_order ASC")
		}).
		Preload("ArticleSections.Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("article_section_steps.step_order ASC")
		}).
		Where("articles.id = ? AND articles.deleted_at IS NULL", id).
		First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("article not found")
		}
		return nil, errs.NewInternal("failed to fetch article", err)
	}
	return &a, nil
}

func (r *educationRepository) CreateArticle(ctx context.Context, a *domain.Article) error {
	if err := r.db.WithContext(ctx).Create(a).Error; err != nil {
		return errs.NewInternal("failed to create article", err)
	}
	return nil
}

func (r *educationRepository) UpdateArticle(ctx context.Context, a *domain.Article) error {
	result := r.db.WithContext(ctx).Save(a)
	if result.Error != nil {
		return errs.NewInternal("failed to update article", result.Error)
	}
	return nil
}

func (r *educationRepository) PublishArticle(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Model(&domain.Article{}).Where("id = ?", id).Update("status", domain.StatusPublikasi)
	if result.Error != nil {
		return errs.NewInternal("failed to publish article", result.Error)
	}
	return nil
}

func (r *educationRepository) RecordView(ctx context.Context, view *domain.ArticleView) error {
	return r.db.WithContext(ctx).Create(view).Error
}

func (r *educationRepository) MarkCompleted(ctx context.Context, completion *domain.UserArticleCompletion) error {
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "patient_id"}, {Name: "article_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"article_read", "article_read_at", "completed_at", "updated_at"}),
	}).Create(completion).Error
	if err != nil {
		return errs.NewInternal("failed to record article completion", err)
	}
	return nil
}

func (r *educationRepository) ToggleSaved(ctx context.Context, patientID string, articleID string) (bool, error) {
	var saved domain.UserSavedArticle
	err := r.db.WithContext(ctx).
		Where("patient_id = ? AND article_id = ?", patientID, articleID).
		First(&saved).Error
	if err == nil {
		// Already exists -> delete (unsave)
		if err := r.db.WithContext(ctx).Delete(&saved).Error; err != nil {
			return false, errs.NewInternal("failed to unsave article", err)
		}
		return false, nil
	}

	// Create bookmark
	saved = domain.UserSavedArticle{
		PatientID: patientID,
		ArticleID: articleID,
	}
	if err := r.db.WithContext(ctx).Create(&saved).Error; err != nil {
		return false, errs.NewInternal("failed to save article", err)
	}
	return true, nil
}

func (r *educationRepository) FindSavedArticles(ctx context.Context, patientID string) ([]domain.Article, error) {
	var items []domain.Article
	err := r.db.WithContext(ctx).
		Preload("Category").
		Joins("JOIN user_saved_articles usa ON usa.article_id = articles.id").
		Where("usa.patient_id = ? AND articles.deleted_at IS NULL", patientID).
		Order("usa.saved_at DESC").
		Find(&items).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch saved articles", err)
	}
	return items, nil
}

func (r *educationRepository) ReplaceSections(ctx context.Context, articleID string, sections []domain.ArticleSection) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Clean steps
		var existingSecs []domain.ArticleSection
		if err := tx.Where("article_id = ?", articleID).Find(&existingSecs).Error; err == nil {
			for _, sec := range existingSecs {
				tx.Where("section_id = ?", sec.ID).Delete(&domain.ArticleSectionStep{})
			}
		}
		// Clean sections
		tx.Where("article_id = ?", articleID).Delete(&domain.ArticleSection{})

		// Create
		for _, sec := range sections {
			sec.ArticleID = articleID
			if err := tx.Create(&sec).Error; err != nil {
				return err
			}
			for _, step := range sec.Steps {
				step.SectionID = sec.ID
				if err := tx.Create(&step).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *educationRepository) DeleteArticle(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Model(&domain.Article{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()"))
	if result.Error != nil {
		return errs.NewInternal("failed to soft delete article", result.Error)
	}
	if result.RowsAffected == 0 {
		return errs.NewNotFound("article not found")
	}
	return nil
}

func (r *educationRepository) GetStats(ctx context.Context) (*EducationStats, error) {
	var totalArticles int64
	var totalCategories int64
	var publishedArticles int64
	var totalReads int64

	if err := r.db.WithContext(ctx).Model(&domain.Article{}).Where("deleted_at IS NULL").Count(&totalArticles).Error; err != nil {
		return nil, errs.NewInternal("failed to count articles", err)
	}

	if err := r.db.WithContext(ctx).Model(&domain.Article{}).Where("deleted_at IS NULL").Distinct("category_id").Count(&totalCategories).Error; err != nil {
		return nil, errs.NewInternal("failed to count categories", err)
	}

	if err := r.db.WithContext(ctx).Model(&domain.Article{}).Where("status = ? AND deleted_at IS NULL", domain.StatusPublikasi).Count(&publishedArticles).Error; err != nil {
		return nil, errs.NewInternal("failed to count published articles", err)
	}

	if err := r.db.WithContext(ctx).Model(&domain.ArticleView{}).
		Joins("JOIN articles ON article_views.article_id = articles.id").
		Where("articles.deleted_at IS NULL").
		Count(&totalReads).Error; err != nil {
		return nil, errs.NewInternal("failed to count article views", err)
	}

	return &EducationStats{
		TotalEducation:    totalArticles,
		TotalCategories:   totalCategories,
		PublishedArticles: publishedArticles,
		TotalReads:        totalReads,
	}, nil
}
