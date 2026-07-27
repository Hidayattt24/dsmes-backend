package domain

import "time"

type ArticleStatus string

const (
	StatusDraft     ArticleStatus = "draft"
	StatusPublikasi ArticleStatus = "publikasi"
)

type SectionType string

const (
	SectionParagraf    SectionType = "paragraf"
	SectionLangkah     SectionType = "langkah"
	SectionInfoPenting SectionType = "info_penting"
)

// ArticleCategory represents categorizations for articles.
type ArticleCategory struct {
	BaseModel

	Name string `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
}

func (ArticleCategory) TableName() string { return "article_categories" }

// Article represents an educational article.
type Article struct {
	BaseModel

	Title                string        `gorm:"type:varchar(200);not null" json:"title"`
	CategoryID           string        `gorm:"type:uuid;not null" json:"category_id"`
	EstimatedReadMinutes int           `json:"estimated_read_minutes"`
	AuthorName           string        `gorm:"type:varchar(150)" json:"author_name"`
	BannerImageURL       string        `gorm:"type:text" json:"banner_image_url"`
	Summary              string        `gorm:"type:text" json:"summary"`
	Status               ArticleStatus `gorm:"type:article_status_enum;not null;default:draft" json:"status"`
	CreatedBy            *string       `gorm:"type:uuid" json:"created_by"`
	Content              string        `gorm:"type:text" json:"content"`
	YoutubeLink          string        `gorm:"type:varchar(255)" json:"youtube_link"`

	// Virtual fields (read-only)
	ReadCount int64 `gorm:"->"`

	// Relations
	Category        *ArticleCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	ArticleSections []ArticleSection `gorm:"foreignKey:ArticleID;constraint:OnDelete:CASCADE" json:"sections,omitempty"`
}

func (Article) TableName() string { return "articles" }

// ArticleSection represents sub-contents of an Article.
type ArticleSection struct {
	BaseModel

	ArticleID    string      `gorm:"type:uuid;not null" json:"article_id"`
	SectionOrder int         `gorm:"not null" json:"section_order"`
	SectionTitle string      `gorm:"type:varchar(200)" json:"section_title"`
	SectionType  SectionType `gorm:"type:section_type_enum;not null" json:"section_type"`
	ContentText  string      `gorm:"type:text" json:"content_text"`
	ImageURL     string      `gorm:"type:text" json:"image_url"`

	// Relations
	Steps []ArticleSectionStep `gorm:"foreignKey:SectionID;constraint:OnDelete:CASCADE" json:"steps,omitempty"`
}

func (ArticleSection) TableName() string { return "article_sections" }

// ArticleSectionStep represents sequential steps inside an ArticleSection.
type ArticleSectionStep struct {
	BaseModel

	SectionID string `gorm:"type:uuid;not null" json:"section_id"`
	StepOrder int    `gorm:"not null" json:"step_order"`
	StepText  string `gorm:"type:text;not null" json:"step_text"`
}

func (ArticleSectionStep) TableName() string { return "article_section_steps" }

// UserArticleCompletion tracks which educational content a patient has
// read or watched. Completion uses OR logic: if either article_read OR
// youtube_watched is true, completed_at is set.
type UserArticleCompletion struct {
	BaseModel

	PatientID   string     `gorm:"type:uuid;not null;uniqueIndex:idx_patient_article_completion" json:"patient_id"`
	ArticleID   string     `gorm:"type:uuid;not null;uniqueIndex:idx_patient_article_completion" json:"article_id"`
	CompletedAt *time.Time `json:"completed_at"`

	// Article read tracking
	ArticleRead               bool       `gorm:"not null;default:false" json:"article_read"`
	ArticleReadAt             *time.Time `json:"article_read_at"`
	ArticleStartedAt          *time.Time `json:"article_started_at"`
	ArticleFinishedAt         *time.Time `json:"article_finished_at"`
	ArticleReadingDuration    int        `gorm:"not null;default:0" json:"article_reading_duration"`
	ArticleLastScrollPosition int        `gorm:"not null;default:0" json:"article_last_scroll_position"`

	// YouTube video watch tracking
	YouTubeWatched     bool       `gorm:"column:youtube_watched;not null;default:false" json:"youtube_watched"`
	YouTubeWatchedAt   *time.Time `gorm:"column:youtube_watched_at" json:"youtube_watched_at"`
	VideoStartedAt     *time.Time `gorm:"column:video_started_at" json:"video_started_at"`
	VideoFinishedAt    *time.Time `gorm:"column:video_finished_at" json:"video_finished_at"`
	VideoWatchDuration int        `gorm:"column:video_watch_duration;not null;default:0" json:"video_watch_duration"`
	VideoLastTimestamp int        `gorm:"column:video_last_timestamp;not null;default:0" json:"video_last_timestamp"`

	// Research analytics tracking
	CompletionSource string `gorm:"type:varchar(20)" json:"completion_source"`
}

func (UserArticleCompletion) TableName() string { return "user_article_completions" }

// ArticleView tracks pageviews on articles.
type ArticleView struct {
	BaseModel

	ArticleID string    `gorm:"type:uuid;not null" json:"article_id"`
	PatientID *string   `gorm:"type:uuid" json:"patient_id"`
	ViewedAt  time.Time `gorm:"not null;default:now()" json:"viewed_at"`
}

func (ArticleView) TableName() string { return "article_views" }

// UserSavedArticle holds bookmarked articles.
type UserSavedArticle struct {
	BaseModel

	PatientID string    `gorm:"type:uuid;not null;uniqueIndex:idx_patient_article_saved" json:"patient_id"`
	ArticleID string    `gorm:"type:uuid;not null;uniqueIndex:idx_patient_article_saved" json:"article_id"`
	SavedAt   time.Time `gorm:"not null;default:now()" json:"saved_at"`
}

func (UserSavedArticle) TableName() string { return "user_saved_articles" }

// PatientEducationActivity records granular actions taken by a patient
// on educational resources for research audit trails.
type PatientEducationActivity struct {
	BaseModel

	PatientID    string `gorm:"type:uuid;not null;index:idx_pea_patient" json:"patient_id"`
	ArticleID    string `gorm:"type:uuid;not null;index:idx_pea_article" json:"article_id"`
	ActivityType string `gorm:"type:varchar(50);not null" json:"activity_type"`
	Metadata     string `gorm:"type:jsonb" json:"metadata"` // Raw jsonb string
}

func (PatientEducationActivity) TableName() string { return "patient_education_activities" }
