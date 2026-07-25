package quiz

import (
	"fmt"
	"strings"
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type ChoiceRequest struct {
	ID           string `json:"id,omitempty"`
	OptionText   string `json:"option_text" validate:"required"`
	IsCorrect    bool   `json:"is_correct"`
	DisplayOrder int    `json:"display_order"`
}

type QuestionRequest struct {
	ID           string          `json:"id,omitempty"`
	QuestionText string          `json:"question_text" validate:"required"`
	Explanation  string          `json:"explanation"`
	DisplayOrder int             `json:"display_order"`
	Choices      []ChoiceRequest `json:"choices" validate:"required,min=2"`
}

type QuestionCategoryRequest struct {
	ID           string            `json:"id,omitempty"`
	Title        string            `json:"title" validate:"required,min=2,max=200"`
	Description  string            `json:"description"`
	DisplayOrder int               `json:"display_order"`
	Questions    []QuestionRequest `json:"questions" validate:"required,min=1"`
}

type CreateQuestionnaireRequest struct {
	Title        string                    `json:"title" validate:"required,min=3,max=200"`
	Type         string                    `json:"type" validate:"required"` // PRE_TEST | POST_TEST
	Description  string                    `json:"description"`
	EducationID  *string                   `json:"education_id"`
	PassingScore *int                      `json:"passing_score"`
	Difficulty   *string                   `json:"difficulty"`
	Status       string                    `json:"status"` // aktif | draft | nonaktif
	Categories   []QuestionCategoryRequest `json:"categories" validate:"required,min=1"`
}

type SubmitQuestionnaireRequest struct {
	DurationSeconds int `json:"duration_seconds"`
	Answers         []struct {
		QuestionID string `json:"question_id" validate:"required"`
		OptionID   string `json:"option_id" validate:"required"`
	} `json:"answers" validate:"required,min=1"`
}

type ChoiceResponse struct {
	ID           string `json:"id"`
	OptionText   string `json:"option_text"`
	IsCorrect    bool   `json:"is_correct,omitempty"`
	DisplayOrder int    `json:"display_order"`
}

type QuestionResponse struct {
	ID           string           `json:"id"`
	QuestionText string           `json:"question_text"`
	Explanation  string           `json:"explanation,omitempty"`
	DisplayOrder int              `json:"display_order"`
	Choices      []ChoiceResponse `json:"choices"`
}

type QuestionCategoryResponse struct {
	ID           string             `json:"id"`
	Title        string             `json:"title"`
	Description  string             `json:"description,omitempty"`
	DisplayOrder int                `json:"display_order"`
	Questions    []QuestionResponse `json:"questions"`
}

type QuestionnaireResponse struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Type             string    `json:"type"` // "PRE_TEST" | "POST_TEST"
	Description      string    `json:"description,omitempty"`
	EducationID      *string   `json:"education_id,omitempty"`
	EducationTitle   string    `json:"education_title,omitempty"`
	PassingScore     *int      `json:"passing_score,omitempty"`
	Difficulty       *string   `json:"difficulty,omitempty"`
	Status           string    `json:"status"`
	CategoryCount    int       `json:"category_count"`
	QuestionCount    int       `json:"question_count"`
	ParticipantCount int       `json:"participant_count"`
	AverageScore     *int      `json:"average_score"`
	CreatedBy        string    `json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type QuestionnaireDetailResponse struct {
	QuestionnaireResponse
	Categories []QuestionCategoryResponse `json:"categories"`
}

type QuizStats struct {
	TotalQuizzes     int64 `json:"total_quizzes"`
	PublishedQuizzes int64 `json:"published_quizzes"`
	DraftQuizzes     int64 `json:"draft_quizzes"`
	TotalAttempts    int64 `json:"total_attempts"`
	AverageScore     int   `json:"average_score"`
}

type ParticipantResponse struct {
	ID             string    `json:"id"`
	PatientID      string    `json:"patient_id"`
	PatientName    string    `json:"patient_name"`
	PatientAvatar  string    `json:"patient_avatar,omitempty"`
	Puskesmas      string    `json:"puskesmas"`
	CompletionDate time.Time `json:"completion_date"`
	Score          int       `json:"score"`
	Passed         bool      `json:"passed"`
	Duration       string    `json:"duration"`
}

type QuestionAnalysisResponse struct {
	ID             string `json:"id"`
	QuestionNumber int    `json:"question_number"`
	QuestionText   string `json:"question_text"`
	PatientAnswer  string `json:"patient_answer"`
	CorrectAnswer  string `json:"correct_answer"`
	IsCorrect      bool   `json:"is_correct"`
	Explanation    string `json:"explanation"`
}

type ParticipantDetailResponse struct {
	Participant      ParticipantResponse        `json:"participant"`
	QuizTitle        string                     `json:"quiz_title"`
	QuestionAnalysis []QuestionAnalysisResponse `json:"question_analysis"`
}

type SubmitResultResponse struct {
	AttemptID       string `json:"attempt_id"`
	QuestionnaireID string `json:"questionnaire_id"`
	Score           int    `json:"score"`
	Passed          bool   `json:"passed"`
	TotalQuestions  int    `json:"total_questions"`
	CorrectCount    int    `json:"correct_count"`
}

func formatDuration(seconds int) string {
	m := seconds / 60
	s := seconds % 60
	if m == 0 {
		return fmt.Sprintf("%ds", s)
	}
	if s == 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%dm %ds", m, s)
}

func normalizeType(t string) domain.QuestionnaireType {
	upper := strings.ToUpper(strings.TrimSpace(t))
	if strings.Contains(upper, "PRE") {
		return domain.TypePreTest
	}
	return domain.TypePostTest
}

func normalizeStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "terbit", "aktif", "active":
		return "aktif"
	case "nonaktif", "inactive":
		return "nonaktif"
	default:
		return "draft"
	}
}

func displayStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "aktif", "active", "terbit":
		return "Aktif"
	case "nonaktif":
		return "Nonaktif"
	default:
		return "Draft"
	}
}

func ToQuestionnaireResponse(q *domain.Questionnaire, categoryCount, questionCount, participantCount int, avgScore *int) QuestionnaireResponse {
	eduTitle := ""
	if q.Education != nil {
		eduTitle = q.Education.Title
	}

	createdBy := "-"
	if q.CreatedBy != nil {
		createdBy = *q.CreatedBy
	}

	return QuestionnaireResponse{
		ID:               q.ID,
		Title:            q.Title,
		Type:             string(q.Type),
		Description:      q.Description,
		EducationID:      q.EducationID,
		EducationTitle:   eduTitle,
		PassingScore:     q.PassingScore,
		Difficulty:       q.Difficulty,
		Status:           displayStatus(q.Status),
		CategoryCount:    categoryCount,
		QuestionCount:    questionCount,
		ParticipantCount: participantCount,
		AverageScore:     avgScore,
		CreatedBy:        createdBy,
		CreatedAt:        q.CreatedAt,
		UpdatedAt:        q.UpdatedAt,
	}
}

func ToQuestionnaireDetailResponse(q *domain.Questionnaire, participantCount int, avgScore *int, isAdminOrStaff bool) QuestionnaireDetailResponse {
	totalCat := len(q.Categories)
	totalQuest := 0
	catResps := make([]QuestionCategoryResponse, len(q.Categories))

	for i, cat := range q.Categories {
		qResps := make([]QuestionResponse, len(cat.Questions))
		totalQuest += len(cat.Questions)

		for j, quest := range cat.Questions {
			cResps := make([]ChoiceResponse, len(quest.Options))
			for k, opt := range quest.Options {
				cResps[k] = ChoiceResponse{
					ID:           opt.ID,
					OptionText:   opt.OptionText,
					IsCorrect:    isAdminOrStaff && opt.IsCorrect,
					DisplayOrder: opt.DisplayOrder,
				}
			}
			qResps[j] = QuestionResponse{
				ID:           quest.ID,
				QuestionText: quest.QuestionText,
				Explanation:  quest.Explanation,
				DisplayOrder: quest.DisplayOrder,
				Choices:      cResps,
			}
		}

		catResps[i] = QuestionCategoryResponse{
			ID:           cat.ID,
			Title:        cat.Title,
			Description:  cat.Description,
			DisplayOrder: cat.DisplayOrder,
			Questions:    qResps,
		}
	}

	return QuestionnaireDetailResponse{
		QuestionnaireResponse: ToQuestionnaireResponse(q, totalCat, totalQuest, participantCount, avgScore),
		Categories:            catResps,
	}
}
