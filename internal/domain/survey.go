package domain

import (
	"encoding/json"
	"time"
)

type SurveyType string
type SurveyStatus string

const (
	SurveyTypeUserSatisfaction SurveyType = "USER_SATISFACTION"
	SurveyTypeSUS              SurveyType = "SUS"

	SurveyStatusDraft     SurveyStatus = "draft"
	SurveyStatusPublished SurveyStatus = "published"
	SurveyStatusArchived  SurveyStatus = "archived"
)

// Survey represents a research evaluation survey instrument (User Satisfaction or SUS).
type Survey struct {
	BaseModel

	Title             string       `gorm:"type:varchar(200);not null" json:"title"`
	Description       string       `gorm:"type:text" json:"description,omitempty"`
	Type              SurveyType   `gorm:"type:varchar(50);not null;default:'USER_SATISFACTION'" json:"type"`
	Status            SurveyStatus `gorm:"type:varchar(50);not null;default:'draft'" json:"status"`
	IsActive          bool         `gorm:"type:boolean;not null;default:false" json:"is_active"`
	StartDate         *time.Time   `gorm:"type:timestamptz" json:"start_date,omitempty"`
	EndDate           *time.Time   `gorm:"type:timestamptz" json:"end_date,omitempty"`
	CreatedBy         *string      `gorm:"type:uuid" json:"created_by,omitempty"`

	// Relations
	Questions []SurveyQuestion `gorm:"foreignKey:SurveyID;constraint:OnDelete:CASCADE" json:"questions,omitempty"`
	Responses []SurveyResponse `gorm:"foreignKey:SurveyID;constraint:OnDelete:CASCADE" json:"responses,omitempty"`
}

func (Survey) TableName() string { return "surveys" }

// SurveyQuestion represents a single Likert-scale question within a survey.
type SurveyQuestion struct {
	BaseModel

	SurveyID        string          `gorm:"type:uuid;not null" json:"survey_id"`
	QuestionText    string          `gorm:"type:text;not null" json:"question_text"`
	Description     string          `gorm:"type:text" json:"description,omitempty"`
	ImageURL        *string         `gorm:"type:text" json:"image_url,omitempty"`
	SVGIllustration *string         `gorm:"type:text" json:"svg_illustration,omitempty"`
	LikertLabels    json.RawMessage `gorm:"type:jsonb;default:'[\"Sangat Tidak Setuju\", \"Tidak Setuju\", \"Netral\", \"Setuju\", \"Sangat Setuju\"]'" json:"likert_labels,omitempty"`
	IsRequired      bool            `gorm:"type:boolean;not null;default:true" json:"is_required"`
	DisplayOrder    int             `gorm:"type:int;not null;default:0" json:"display_order"`
}

func (SurveyQuestion) TableName() string { return "survey_questions" }

// SurveyResponse represents a participant's submission of a Survey.
type SurveyResponse struct {
	BaseModel

	SurveyID        string     `gorm:"type:uuid;not null" json:"survey_id"`
	PatientID       string     `gorm:"type:uuid;not null" json:"patient_id"`
	StartedAt       *time.Time `gorm:"type:timestamptz" json:"started_at,omitempty"`
	CompletedAt     time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"completed_at"`
	DurationSeconds int        `gorm:"type:int;not null;default:0" json:"duration_seconds"`

	// Scoring - User Satisfaction
	TotalScore      *float64 `gorm:"type:double precision" json:"total_score,omitempty"`
	AverageScore    *float64 `gorm:"type:double precision" json:"average_score,omitempty"`
	PercentageScore *float64 `gorm:"type:double precision" json:"percentage_score,omitempty"`

	// Scoring - SUS
	RawScore       *float64 `gorm:"type:double precision" json:"raw_score,omitempty"`
	SUSScore       *float64 `gorm:"type:double precision" json:"sus_score,omitempty"`
	Interpretation *string  `gorm:"type:varchar(50)" json:"interpretation,omitempty"`
	Passed         *bool    `gorm:"type:boolean" json:"passed,omitempty"`

	// Relations
	Survey  *Survey        `gorm:"foreignKey:SurveyID" json:"survey,omitempty"`
	Patient *Patient       `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Answers []SurveyAnswer `gorm:"foreignKey:ResponseID;constraint:OnDelete:CASCADE" json:"answers,omitempty"`
}

func (SurveyResponse) TableName() string { return "survey_responses" }

// SurveyAnswer records a patient's rating selection for a specific question.
type SurveyAnswer struct {
	BaseModel

	ResponseID    string   `gorm:"type:uuid;not null" json:"response_id"`
	QuestionID    string   `gorm:"type:uuid;not null" json:"question_id"`
	RatingValue   int      `gorm:"type:int;not null" json:"rating_value"`
	AdjustedScore *float64 `gorm:"type:double precision" json:"adjusted_score,omitempty"`

	// Relations
	Question *SurveyQuestion `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}

func (SurveyAnswer) TableName() string { return "survey_answers" }
