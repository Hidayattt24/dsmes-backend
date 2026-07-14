package education

import "github.com/dsmes/dsmes-backend/internal/domain"

type StepRequest struct {
	StepOrder int    `json:"step_order" validate:"required"`
	StepText  string `json:"step_text"  validate:"required"`
}

type SectionRequest struct {
	SectionOrder int                `json:"section_order" validate:"required"`
	SectionTitle string             `json:"section_title"`
	SectionType  domain.SectionType `json:"section_type"  validate:"required,oneof=paragraf langkah info_penting"`
	ContentText  string             `json:"content_text"`
	ImageURL     string             `json:"image_url"`
	Steps        []StepRequest      `json:"steps"`
}

type CreateArticleRequest struct {
	Title                string           `json:"title"                  validate:"required,max=200"`
	CategoryID           string           `json:"category_id"            validate:"omitempty,uuid4"`
	CategoryName         string           `json:"category_name"          validate:"required"`
	EstimatedReadMinutes int              `json:"estimated_read_minutes" validate:"required,min=1"`
	AuthorName           string           `json:"author_name"`
	BannerImageURL       string           `json:"banner_image_url"`
	Summary              string           `json:"summary"                validate:"required"`
	Content              string           `json:"content"`
	YoutubeLink          string           `json:"youtube_link"`
	Status               string           `json:"status"                 validate:"omitempty,oneof=draft publikasi"`
	Sections             []SectionRequest `json:"sections"`
}

type StepResponse struct {
	ID        string `json:"id"`
	StepOrder int    `json:"step_order"`
	StepText  string `json:"step_text"`
}

type SectionResponse struct {
	ID           string             `json:"id"`
	SectionOrder int                `json:"section_order"`
	SectionTitle string             `json:"section_title"`
	SectionType  domain.SectionType `json:"section_type"`
	ContentText  string             `json:"content_text"`
	ImageURL     string             `json:"image_url"`
	Steps        []StepResponse     `json:"steps"`
}

type ArticleListResponse struct {
	ID                   string               `json:"id"`
	Title                string               `json:"title"`
	CategoryID           string               `json:"category_id"`
	CategoryName         string               `json:"category_name"`
	EstimatedReadMinutes int                  `json:"estimated_read_minutes"`
	AuthorName           string               `json:"author_name"`
	BannerImageURL       string               `json:"banner_image_url"`
	Summary              string               `json:"summary"`
	Status               domain.ArticleStatus `json:"status"`
	CreatedAt            string               `json:"created_at"`
	Content              string               `json:"content"`
	YoutubeLink          string               `json:"youtube_link"`
	ReadCount            int64                `json:"read_count"`
}

type ArticleDetailResponse struct {
	ArticleListResponse
	Sections []SectionResponse `json:"sections"`
}

type CategoryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type EducationStats struct {
	TotalEducation    int64 `json:"total_education"`
	TotalCategories   int64 `json:"total_categories"`
	PublishedArticles int64 `json:"published_articles"`
	TotalReads        int64 `json:"total_reads"`
}

func ToCategoryResponse(c *domain.ArticleCategory) CategoryResponse {
	return CategoryResponse{
		ID:   c.ID,
		Name: c.Name,
	}
}

func ToArticleListResponse(a *domain.Article) ArticleListResponse {
	catName := ""
	if a.Category != nil {
		catName = a.Category.Name
	}
	return ArticleListResponse{
		ID:                   a.ID,
		Title:                a.Title,
		CategoryID:           a.CategoryID,
		CategoryName:         catName,
		EstimatedReadMinutes: a.EstimatedReadMinutes,
		AuthorName:           a.AuthorName,
		BannerImageURL:       a.BannerImageURL,
		Summary:              a.Summary,
		Status:               a.Status,
		CreatedAt:            a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Content:              a.Content,
		YoutubeLink:          a.YoutubeLink,
		ReadCount:            a.ReadCount,
	}
}

func ToArticleDetailResponse(a *domain.Article) ArticleDetailResponse {
	sections := make([]SectionResponse, len(a.ArticleSections))
	for i, s := range a.ArticleSections {
		steps := make([]StepResponse, len(s.Steps))
		for j, st := range s.Steps {
			steps[j] = StepResponse{
				ID:        st.ID,
				StepOrder: st.StepOrder,
				StepText:  st.StepText,
			}
		}
		sections[i] = SectionResponse{
			ID:           s.ID,
			SectionOrder: s.SectionOrder,
			SectionTitle: s.SectionTitle,
			SectionType:  s.SectionType,
			ContentText:  s.ContentText,
			ImageURL:     s.ImageURL,
			Steps:        steps,
		}
	}

	return ArticleDetailResponse{
		ArticleListResponse: ToArticleListResponse(a),
		Sections:            sections,
	}
}
