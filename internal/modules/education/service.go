package education

import (
	"context"
	"fmt"
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

func (s *educationService) ListArticles(ctx context.Context, patientID *string, categoryID string, status *domain.ArticleStatus, page, limit int) ([]ArticleListResponse, int64, error) {
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

	var savedMap map[string]bool
	var completedMap map[string]bool
	if patientID != nil && *patientID != "" {
		savedMap, _ = s.repo.GetPatientSavedMap(ctx, *patientID)
		completedMap, _ = s.repo.GetPatientCompletedMap(ctx, *patientID)
	}

	resp := make([]ArticleListResponse, len(items))
	for i := range items {
		resp[i] = ToArticleListResponse(&items[i])
		if savedMap != nil {
			resp[i].IsBookmarked = savedMap[items[i].ID]
		}
		if completedMap != nil {
			resp[i].IsCompleted = completedMap[items[i].ID]
		}
	}
	return resp, total, nil
}

func (s *educationService) GetArticle(ctx context.Context, id string, patientID *string) (*ArticleDetailResponse, error) {
	a, err := s.repo.FindArticleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Record pageview asynchronously only if viewed by a patient
	if patientID != nil {
		view := &domain.ArticleView{
			ArticleID: id,
			PatientID: patientID,
			ViewedAt:  time.Now(),
		}
		_ = s.repo.RecordView(ctx, view)
	}

	res := ToArticleDetailResponse(a)
	if patientID != nil && *patientID != "" {
		savedMap, _ := s.repo.GetPatientSavedMap(ctx, *patientID)
		completedMap, _ := s.repo.GetPatientCompletedMap(ctx, *patientID)
		if savedMap != nil {
			res.IsBookmarked = savedMap[a.ID]
		}
		if completedMap != nil {
			res.IsCompleted = completedMap[a.ID]
		}
	}
	return &res, nil
}

func (s *educationService) CreateArticle(ctx context.Context, staffID string, req CreateArticleRequest) (*ArticleDetailResponse, error) {
	// Verify and resolve category
	var categoryID string
	var err error

	if req.CategoryName != "" {
		var category *domain.ArticleCategory
		category, err = s.repo.FindOrCreateCategoryByName(ctx, req.CategoryName)
		if err != nil {
			return nil, err
		}
		categoryID = category.ID
	} else {
		_, err = s.repo.FindCategoryByID(ctx, req.CategoryID)
		if err != nil {
			return nil, err
		}
		categoryID = req.CategoryID
	}

	status := domain.StatusDraft
	if req.Status != "" {
		status = domain.ArticleStatus(req.Status)
	}

	article := &domain.Article{
		Title:                req.Title,
		CategoryID:           categoryID,
		EstimatedReadMinutes: req.EstimatedReadMinutes,
		AuthorName:           req.AuthorName,
		BannerImageURL:       req.BannerImageURL,
		Summary:              req.Summary,
		Status:               status,
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
	var refetched *domain.Article
	refetched, err = s.repo.FindArticleByID(ctx, article.ID)
	if err != nil {
		return nil, err
	}

	// Broadcast new-education notification to all patients when article is
	// created directly as published (active).
	if article.Status == domain.StatusPublikasi {
		s.broadcastEducationNotif(ctx, article.Title, article.ID)
	}

	res := ToArticleDetailResponse(refetched)
	return &res, nil
}

func (s *educationService) UpdateArticle(ctx context.Context, id string, req CreateArticleRequest) (*ArticleDetailResponse, error) {
	article, err := s.repo.FindArticleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify and resolve category
	var categoryID string
	var categoryObj *domain.ArticleCategory
	if req.CategoryName != "" {
		categoryObj, err = s.repo.FindOrCreateCategoryByName(ctx, req.CategoryName)
		if err != nil {
			return nil, err
		}
		categoryID = categoryObj.ID
	} else if req.CategoryID != "" {
		categoryObj, err = s.repo.FindCategoryByID(ctx, req.CategoryID)
		if err != nil {
			return nil, err
		}
		categoryID = req.CategoryID
	}

	article.Title = req.Title
	article.CategoryID = categoryID
	article.Category = categoryObj
	article.EstimatedReadMinutes = req.EstimatedReadMinutes
	article.AuthorName = req.AuthorName
	article.BannerImageURL = req.BannerImageURL
	article.Summary = req.Summary
	article.Content = req.Content
	article.YoutubeLink = req.YoutubeLink

	if req.Status != "" {
		article.Status = domain.ArticleStatus(req.Status)
	}

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
	var refetched *domain.Article
	refetched, err = s.repo.FindArticleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	res := ToArticleDetailResponse(refetched)
	return &res, nil
}

func (s *educationService) PublishArticle(ctx context.Context, id string) error {
	article, err := s.repo.FindArticleByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.PublishArticle(ctx, id); err != nil {
		return err
	}

	// Broadcast only when transitioning from draft to published to avoid
	// duplicate notifications on re-publish.
	if article.Status != domain.StatusPublikasi {
		s.broadcastEducationNotif(ctx, article.Title, article.ID)
	}

	return nil
}

// broadcastEducationNotif asynchronously pushes a notification_log entry to
// every active patient. It runs in a goroutine so it never blocks the request.
func (s *educationService) broadcastEducationNotif(ctx context.Context, title string, articleID string) {
	message := fmt.Sprintf("Materi edukasi baru: %s", title)
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := s.repo.BroadcastEducationNotification(bgCtx, message, articleID); err != nil {
			s.log.Warn("failed to broadcast education notification",
				zap.String("article_id", articleID),
				zap.Error(err),
			)
		}
	}()
}

func (s *educationService) CompleteArticle(ctx context.Context, patientID string, id string) error {
	_, err := s.repo.FindArticleByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	completion := &domain.UserArticleCompletion{
		PatientID:     patientID,
		ArticleID:     id,
		ArticleRead:   true,
		ArticleReadAt: &now,
		CompletedAt:   &now,
	}

	return s.repo.MarkCompleted(ctx, completion)
}

func (s *educationService) SaveArticle(ctx context.Context, patientID string, id string) error {
	_, err := s.repo.FindArticleByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.SaveArticle(ctx, patientID, id)
}

func (s *educationService) UnsaveArticle(ctx context.Context, patientID string, id string) error {
	_, err := s.repo.FindArticleByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.UnsaveArticle(ctx, patientID, id)
}

func (s *educationService) ListSavedArticles(ctx context.Context, patientID string) ([]ArticleListResponse, error) {
	items, err := s.repo.FindSavedArticles(ctx, patientID)
	if err != nil {
		return nil, err
	}

	completedMap, _ := s.repo.GetPatientCompletedMap(ctx, patientID)

	resp := make([]ArticleListResponse, len(items))
	for i := range items {
		resp[i] = ToArticleListResponse(&items[i])
		resp[i].IsBookmarked = true
		if completedMap != nil {
			resp[i].IsCompleted = completedMap[items[i].ID]
		}
	}
	return resp, nil
}

func (s *educationService) DeleteArticle(ctx context.Context, id string) error {
	return s.repo.DeleteArticle(ctx, id)
}

func (s *educationService) GetStats(ctx context.Context) (*EducationStats, error) {
	return s.repo.GetStats(ctx)
}

func (s *educationService) SubmitReview(ctx context.Context, patientID string, educationID string, req CreateReviewRequest) (*EducationReviewResponse, error) {
	if _, err := s.repo.FindArticleByID(ctx, educationID); err != nil {
		return nil, err
	}

	review := &domain.EducationReview{
		EducationID: educationID,
		PatientID:   patientID,
		Rating:      req.Rating,
		Note:        req.Note,
	}

	if err := s.repo.UpsertReview(ctx, review); err != nil {
		return nil, err
	}

	saved, err := s.repo.GetReviewByPatientAndArticle(ctx, patientID, educationID)
	if err != nil || saved == nil {
		saved = review
	}

	res := &EducationReviewResponse{
		ID:          saved.ID,
		EducationID: saved.EducationID,
		PatientID:   saved.PatientID,
		Rating:      saved.Rating,
		Note:        saved.Note,
		CreatedAt:   saved.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   saved.UpdatedAt.Format(time.RFC3339),
	}
	return res, nil
}

func (s *educationService) GetPatientReview(ctx context.Context, patientID string, educationID string) (*EducationReviewResponse, error) {
	rev, err := s.repo.GetReviewByPatientAndArticle(ctx, patientID, educationID)
	if err != nil {
		return nil, err
	}
	if rev == nil {
		return nil, nil
	}
	res := &EducationReviewResponse{
		ID:          rev.ID,
		EducationID: rev.EducationID,
		PatientID:   rev.PatientID,
		Rating:      rev.Rating,
		Note:        rev.Note,
		CreatedAt:   rev.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   rev.UpdatedAt.Format(time.RFC3339),
	}
	return res, nil
}

func (s *educationService) GetRatingSummary(ctx context.Context, educationID string, patientID *string) (*ArticleRatingResponse, error) {
	if _, err := s.repo.FindArticleByID(ctx, educationID); err != nil {
		return nil, err
	}

	avg, total, dist, err := s.repo.GetRatingSummary(ctx, educationID)
	if err != nil {
		return nil, err
	}

	var userReviewResp *EducationReviewResponse
	if patientID != nil && *patientID != "" {
		rev, err := s.repo.GetReviewByPatientAndArticle(ctx, *patientID, educationID)
		if err == nil && rev != nil {
			userReviewResp = &EducationReviewResponse{
				ID:          rev.ID,
				EducationID: rev.EducationID,
				PatientID:   rev.PatientID,
				Rating:      rev.Rating,
				Note:        rev.Note,
				CreatedAt:   rev.CreatedAt.Format(time.RFC3339),
				UpdatedAt:   rev.UpdatedAt.Format(time.RFC3339),
			}
		}
	}

	return &ArticleRatingResponse{
		AverageRating:      avg,
		TotalReviews:       total,
		RatingDistribution: dist,
		CurrentUserReview:  userReviewResp,
	}, nil
}

func (s *educationService) GetAdminReviews(ctx context.Context, educationID string) (*AdminArticleReviewsResponse, error) {
	if _, err := s.repo.FindArticleByID(ctx, educationID); err != nil {
		return nil, err
	}

	avg, total, dist, err := s.repo.GetRatingSummary(ctx, educationID)
	if err != nil {
		return nil, err
	}

	reviews, patientNames, completionDates, err := s.repo.GetAdminReviews(ctx, educationID)
	if err != nil {
		return nil, err
	}

	list := make([]EducationReviewResponse, len(reviews))
	for i, r := range reviews {
		name := patientNames[r.PatientID]
		if name == "" {
			name = "Pasien"
		}
		var compStr *string
		if cDate, ok := completionDates[r.PatientID]; ok && cDate != nil {
			formatted := cDate.Format(time.RFC3339)
			compStr = &formatted
		}

		list[i] = EducationReviewResponse{
			ID:             r.ID,
			EducationID:    r.EducationID,
			PatientID:      r.PatientID,
			PatientName:    name,
			Rating:         r.Rating,
			Note:           r.Note,
			CompletionDate: compStr,
			CreatedAt:      r.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      r.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &AdminArticleReviewsResponse{
		AverageRating:      avg,
		TotalReviews:       total,
		RatingDistribution: dist,
		Reviews:            list,
	}, nil
}

