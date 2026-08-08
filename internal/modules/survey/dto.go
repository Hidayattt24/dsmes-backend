package survey

import (
	"encoding/json"
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

// CreateSurveyRequest represents payload to create a new survey.
type CreateSurveyRequest struct {
	Title       string            `json:"title" validate:"required,max=200"`
	Description string            `json:"description,omitempty"`
	Type        domain.SurveyType `json:"type" validate:"required,oneof=USER_SATISFACTION SUS"`
	StartDate   *time.Time        `json:"start_date,omitempty"`
	EndDate     *time.Time        `json:"end_date,omitempty"`
	Questions   []QuestionRequest `json:"questions" validate:"required,min=1,dive"`
}

// UpdateSurveyRequest represents payload to update an existing survey.
type UpdateSurveyRequest struct {
	Title       string            `json:"title" validate:"required,max=200"`
	Description string            `json:"description,omitempty"`
	Type        domain.SurveyType `json:"type" validate:"required,oneof=USER_SATISFACTION SUS"`
	StartDate   *time.Time        `json:"start_date,omitempty"`
	EndDate     *time.Time        `json:"end_date,omitempty"`
	Questions   []QuestionRequest `json:"questions" validate:"required,min=1,dive"`
}

// QuestionRequest represents a single question in survey creation/update.
type QuestionRequest struct {
	ID              *string  `json:"id,omitempty"`
	QuestionText    string   `json:"question_text" validate:"required"`
	Description     string   `json:"description,omitempty"`
	ImageURL        *string  `json:"image_url,omitempty"`
	SVGIllustration *string  `json:"svg_illustration,omitempty"`
	LikertLabels    []string `json:"likert_labels,omitempty"`
	IsRequired      *bool    `json:"is_required,omitempty"`
	DisplayOrder    int      `json:"display_order"`
}

// UpdateSurveyStatusRequest payload for changing status or active state.
type UpdateSurveyStatusRequest struct {
	Status   domain.SurveyStatus `json:"status,omitempty" validate:"omitempty,oneof=draft published archived"`
	IsActive *bool               `json:"is_active,omitempty"`
}

// QuestionResponseDTO represents question output.
type QuestionResponseDTO struct {
	ID              string   `json:"id"`
	SurveyID        string   `json:"survey_id"`
	QuestionText    string   `json:"question_text"`
	Description     string   `json:"description,omitempty"`
	ImageURL        *string  `json:"image_url,omitempty"`
	SVGIllustration *string  `json:"svg_illustration,omitempty"`
	LikertLabels    []string `json:"likert_labels"`
	IsRequired      bool     `json:"is_required"`
	DisplayOrder    int      `json:"display_order"`
}

// SurveyDetailResponse represents survey output detail with questions.
type SurveyDetailResponse struct {
	ID            string                `json:"id"`
	Title         string                `json:"title"`
	Description   string                `json:"description,omitempty"`
	Type          domain.SurveyType     `json:"type"`
	Status        domain.SurveyStatus   `json:"status"`
	IsActive      bool                  `json:"is_active"`
	HasSubmitted  bool                  `json:"has_submitted"`
	StartDate     *time.Time            `json:"start_date,omitempty"`
	EndDate       *time.Time            `json:"end_date,omitempty"`
	QuestionCount int                   `json:"question_count"`
	ResponseCount int                   `json:"response_count"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
	Questions     []QuestionResponseDTO `json:"questions"`
}

// SurveyListItemResponse represents survey item in list view.
type SurveyListItemResponse struct {
	ID            string              `json:"id"`
	Title         string              `json:"title"`
	Description   string              `json:"description,omitempty"`
	Type          domain.SurveyType   `json:"type"`
	Status        domain.SurveyStatus `json:"status"`
	IsActive      bool                `json:"is_active"`
	StartDate     *time.Time          `json:"start_date,omitempty"`
	EndDate       *time.Time          `json:"end_date,omitempty"`
	QuestionCount int                 `json:"question_count"`
	ResponseCount int                 `json:"response_count"`
	CreatedAt     time.Time           `json:"created_at"`
}

// SubmitAnswerRequest represents single answer in submission.
type SubmitAnswerRequest struct {
	QuestionID  string `json:"question_id" validate:"required"`
	RatingValue int    `json:"rating_value" validate:"required,min=1,max=5"`
}

// SubmitSurveyRequest represents survey submission by participant.
type SubmitSurveyRequest struct {
	Answers         []SubmitAnswerRequest `json:"answers" validate:"required,min=1,dive"`
	DurationSeconds int                   `json:"duration_seconds" validate:"gte=0"`
	StartedAt       *time.Time            `json:"started_at,omitempty"`
}

// SubmitSurveyResponse represents output after successful submission.
type SubmitSurveyResponse struct {
	ResponseID      string    `json:"response_id"`
	SurveyID        string    `json:"survey_id"`
	SurveyTitle     string    `json:"survey_title"`
	Type            string    `json:"type"`
	TotalScore      *float64  `json:"total_score,omitempty"`
	AverageScore    *float64  `json:"average_score,omitempty"`
	PercentageScore *float64  `json:"percentage_score,omitempty"`
	RawScore        *float64  `json:"raw_score,omitempty"`
	SUSScore        *float64  `json:"sus_score,omitempty"`
	Interpretation  *string   `json:"interpretation,omitempty"`
	Passed          *bool     `json:"passed,omitempty"`
	CompletedAt     time.Time `json:"completed_at"`
	Message         string    `json:"message"`
}

// SurveyResponseItemResponse represents a participant's response item for Admin/Staff.
type SurveyResponseItemResponse struct {
	ID              string             `json:"id"`
	SurveyID        string             `json:"survey_id"`
	PatientID       string             `json:"patient_id"`
	PatientName     string             `json:"patient_name"`
	PatientEmail    string             `json:"patient_email"`
	PatientPhone    string             `json:"patient_phone"`
	StartedAt       *time.Time         `json:"started_at,omitempty"`
	CompletedAt     time.Time          `json:"completed_at"`
	DurationSeconds int                `json:"duration_seconds"`
	TotalScore      *float64           `json:"total_score,omitempty"`
	AverageScore    *float64           `json:"average_score,omitempty"`
	PercentageScore *float64           `json:"percentage_score,omitempty"`
	RawScore        *float64           `json:"raw_score,omitempty"`
	SUSScore        *float64           `json:"sus_score,omitempty"`
	Interpretation  *string            `json:"interpretation,omitempty"`
	Passed          *bool              `json:"passed,omitempty"`
	Answers         []AnswerDetailDTO `json:"answers,omitempty"`
}

// AnswerDetailDTO detailed answer record.
type AnswerDetailDTO struct {
	QuestionID    string   `json:"question_id"`
	QuestionText  string   `json:"question_text"`
	RatingValue   int      `json:"rating_value"`
	AdjustedScore *float64 `json:"adjusted_score,omitempty"`
}

// QuestionAnalytic breakdown per question.
type QuestionAnalytic struct {
	QuestionID    string         `json:"question_id"`
	QuestionText  string         `json:"question_text"`
	DisplayOrder  int            `json:"display_order"`
	AverageRating float64        `json:"average_rating"`
	RatingCounts  map[string]int `json:"rating_counts"` // "1": 5, "2": 3, etc.
}

// DistributionItem breakdown for SUS interpretations or rating percentages.
type DistributionItem struct {
	Label      string  `json:"label"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// SurveyAnalyticsResponse represents complete analytics report for a survey.
type SurveyAnalyticsResponse struct {
	SurveyID             string             `json:"survey_id"`
	SurveyTitle          string             `json:"survey_title"`
	Type                 domain.SurveyType  `json:"type"`
	TotalParticipants    int                `json:"total_participants"`
	CompletedCount       int                `json:"completed_count"`
	CompletionRate       float64            `json:"completion_rate"`
	AverageDurationSecs  int                `json:"average_duration_secs"`

	// Satisfaction Metrics (if USER_SATISFACTION)
	AverageScore        *float64           `json:"average_score,omitempty"`
	AveragePercentage   *float64           `json:"average_percentage,omitempty"`

	// SUS Metrics (if SUS)
	AverageSUSScore     *float64           `json:"average_sus_score,omitempty"`
	HighestSUSScore     *float64           `json:"highest_sus_score,omitempty"`
	LowestSUSScore      *float64           `json:"lowest_sus_score,omitempty"`
	PassCount           *int               `json:"pass_count,omitempty"`
	FailCount           *int               `json:"fail_count,omitempty"`
	PassRate            *float64           `json:"pass_rate,omitempty"`

	Interpretations     []DistributionItem `json:"interpretations,omitempty"`
	QuestionStatistics  []QuestionAnalytic `json:"question_statistics"`
}

// Helper to convert json.RawMessage labels to []string
func ParseLikertLabels(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return []string{"Sangat Tidak Setuju", "Tidak Setuju", "Netral", "Setuju", "Sangat Setuju"}
	}
	var labels []string
	if err := json.Unmarshal(raw, &labels); err == nil && len(labels) > 0 {
		return labels
	}
	return []string{"Sangat Tidak Setuju", "Tidak Setuju", "Netral", "Setuju", "Sangat Setuju"}
}
