package education

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type educationService struct {
	repo EducationRepository
	log  *zap.Logger
}

func NewEducationService(repo EducationRepository, log *zap.Logger) EducationService {
	return &educationService{repo: repo, log: log}
}

func (s *educationService) ListCategories(ctx context.Context) ([]CategoryResponse, error) {
	items, err := s.repo.FindAllCategories(ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]CategoryResponse, len(items))
	for i := range items {
		resp[i] = ToCategoryResponse(&items[i])
	}
	return resp, nil
}

func (s *educationService) ListArticles(ctx context.Context, categoryID string, status *domain.ArticleStatus, page, limit int) ([]ArticleListResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	items, total, err := s.repo.FindAllArticles(ctx, categoryID, status, page, limit)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]ArticleListResponse, len(items))
	for i := range items {
		resp[i] = ToArticleListResponse(&items[i])
	}
	return resp, total, nil
}

func (s *educationService) GetArticle(ctx context.Context, id string, patientID *string) (*ArticleDetailResponse, error) {
	a, err := s.repo.FindArticleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Record pageview asynchronously
	view := &domain.ArticleView{
		ArticleID: id,
		PatientID: patientID,
		ViewedAt:  time.Now(),
	}
	_ = s.repo.RecordView(ctx, view)

	res := ToArticleDetailResponse(a)
	return &res, nil
}

func (s *educationService) CreateArticle(ctx context.Context, staffID string, req CreateArticleRequest) (*ArticleDetailResponse, error) {
	// Verify category exists
	_, err := s.repo.FindCategoryByID(ctx, req.CategoryID)
	if err != nil {
		return nil, err
	}

	article := &domain.Article{
		Title:                req.Title,
		CategoryID:           req.CategoryID,
		EstimatedReadMinutes: req.EstimatedReadMinutes,
		AuthorName:           req.AuthorName,
		BannerImageURL:       req.BannerImageURL,
		Summary:              req.Summary,
		Status:               domain.StatusDraft,
		CreatedBy:            &staffID,
		Content:              req.Content,
		YoutubeLink:          req.YoutubeLink,
	}

	if err = s.repo.CreateArticle(ctx, article); err != nil {
		return nil, err
	}

	// Create nested sections/steps
	sections := make([]domain.ArticleSection, len(req.Sections))
	for i, sec := range req.Sections {
		steps := make([]domain.ArticleSectionStep, len(sec.Steps))
		for j, st := range sec.Steps {
			steps[j] = domain.ArticleSectionStep{
				StepOrder: st.StepOrder,
				StepText:  st.StepText,
			}
		}
		sections[i] = domain.ArticleSection{
			SectionOrder: sec.SectionOrder,
			SectionTitle: sec.SectionTitle,
			SectionType:  sec.SectionType,
			ContentText:  sec.ContentText,
			ImageURL:     sec.ImageURL,
			Steps:        steps,
		}
	}

	if err = s.repo.ReplaceSections(ctx, article.ID, sections); err != nil {
		return nil, err
	}

	// Refetch full article
	refetched, err := s.repo.FindArticleByID(ctx, article.ID)
	if err != nil {
		return nil, err
	}

	res := ToArticleDetailResponse(refetched)
	return &res, nil
}

func (s *educationService) UpdateArticle(ctx context.Context, id string, req CreateArticleRequest) (*ArticleDetailResponse, error) {
	article, err := s.repo.FindArticleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.FindCategoryByID(ctx, req.CategoryID)
	if err != nil {
		return nil, err
	}

	article.Title = req.Title
	article.CategoryID = req.CategoryID
	article.EstimatedReadMinutes = req.EstimatedReadMinutes
	article.AuthorName = req.AuthorName
	article.BannerImageURL = req.BannerImageURL
	article.Summary = req.Summary
	article.Content = req.Content
	article.YoutubeLink = req.YoutubeLink

	if err = s.repo.UpdateArticle(ctx, article); err != nil {
		return nil, err
	}

	// Replace nested sections/steps
	sections := make([]domain.ArticleSection, len(req.Sections))
	for i, sec := range req.Sections {
		steps := make([]domain.ArticleSectionStep, len(sec.Steps))
		for j, st := range sec.Steps {
			steps[j] = domain.ArticleSectionStep{
				StepOrder: st.StepOrder,
				StepText:  st.StepText,
			}
		}
		sections[i] = domain.ArticleSection{
			SectionOrder: sec.SectionOrder,
			SectionTitle: sec.SectionTitle,
			SectionType:  sec.SectionType,
			ContentText:  sec.ContentText,
			ImageURL:     sec.ImageURL,
			Steps:        steps,
		}
	}

	if err = s.repo.ReplaceSections(ctx, id, sections); err != nil {
		return nil, err
	}

	// Refetch full
	refetched, err := s.repo.FindArticleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	res := ToArticleDetailResponse(refetched)
	return &res, nil
}

func (s *educationService) PublishArticle(ctx context.Context, id string) error {
	return s.repo.PublishArticle(ctx, id)
}

func (s *educationService) CompleteArticle(ctx context.Context, patientID string, id string) error {
	_, err := s.repo.FindArticleByID(ctx, id)
	if err != nil {
		return err
	}

	completion := &domain.UserArticleCompletion{
		PatientID:   patientID,
		ArticleID:   id,
		CompletedAt: time.Now(),
	}

	return s.repo.MarkCompleted(ctx, completion)
}

func (s *educationService) ToggleSaveArticle(ctx context.Context, patientID string, id string) (bool, error) {
	_, err := s.repo.FindArticleByID(ctx, id)
	if err != nil {
		return false, err
	}

	return s.repo.ToggleSaved(ctx, patientID, id)
}

func (s *educationService) ListSavedArticles(ctx context.Context, patientID string) ([]ArticleListResponse, error) {
	items, err := s.repo.FindSavedArticles(ctx, patientID)
	if err != nil {
		return nil, err
	}

	resp := make([]ArticleListResponse, len(items))
	for i := range items {
		resp[i] = ToArticleListResponse(&items[i])
	}
	return resp, nil
}

func (s *educationService) DeleteArticle(ctx context.Context, id string) error {
	return s.repo.DeleteArticle(ctx, id)
}

func (s *educationService) GetStats(ctx context.Context) (*EducationStats, error) {
	return s.repo.GetStats(ctx)
}
