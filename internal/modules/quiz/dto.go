package quiz

import (
	"fmt"
	"strings"
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type ChoiceRequest struct {
	ID           string   `json:"id,omitempty"`
	OptionText   string   `json:"option_text"`
	ScoreValue   *float64 `json:"score_value,omitempty"`
	IsCorrect    bool     `json:"is_correct"`
	DisplayOrder int      `json:"display_order"`
}

type QuestionRequest struct {
	ID               string          `json:"id,omitempty"`
	QuestionText     string          `json:"question_text" validate:"required"`
	QuestionImageURL string          `json:"question_image_url,omitempty"`
	Explanation      string          `json:"explanation,omitempty"`
	DisplayOrder     int             `json:"display_order"`
	Choices          []ChoiceRequest `json:"choices,omitempty"`
}

type QuestionCategoryRequest struct {
	ID                 string            `json:"id,omitempty"`
	Title              string            `json:"title"`
	Description        string            `json:"description,omitempty"`
	DisplayOrder       int               `json:"display_order"`
	InstrumentCategory string            `json:"instrument_category,omitempty"` // PRE_TEST | DSMQR | DSMES
	Questions          []QuestionRequest `json:"questions" validate:"required,min=1"`
}

type CreateQuestionnaireRequest struct {
	Title        string                    `json:"title" validate:"required,min=3,max=200"`
	Type         string                    `json:"type" validate:"required"` // PRE_TEST | POST_TEST
	Scope        *string                   `json:"scope,omitempty"`          // APPLICATION | EDUCATION
	Description  string                    `json:"description"`
	EducationID  *string                   `json:"education_id"`
	PassingScore *int                      `json:"passing_score"`
	Difficulty   *string                   `json:"difficulty"`
	Status       string                    `json:"status"` // aktif | draft | nonaktif
	Categories   []QuestionCategoryRequest `json:"categories,omitempty"`
	Questions    []QuestionRequest         `json:"questions,omitempty"` // For sequential questions without categories

	// Post-Test is created separately and explicitly linked to a Pre-Test.
	SourcePreTestID          *string `json:"source_pre_test_id,omitempty"`
	CopyQuestionsFromPreTest bool    `json:"copy_questions_from_pre_test,omitempty"`
}

type UpdateQuestionnaireStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

type PairQuestionnaireRequest struct {
	SourcePreTestID string `json:"source_pre_test_id" validate:"required"`
}

type SubmitAnswerItem struct {
	QuestionID    string  `json:"question_id" validate:"required"`
	OptionID      *string `json:"option_id,omitempty"`
	SelectedValue *int    `json:"selected_value,omitempty"` // 1-5 integer scale for PRE_TEST
}

type SubmitQuestionnaireRequest struct {
	DurationSeconds int                `json:"duration_seconds"`
	Answers         []SubmitAnswerItem `json:"answers" validate:"required,min=1"`
	ClientID        *string            `json:"client_id,omitempty"`
}

type ChoiceResponse struct {
	ID           string   `json:"id"`
	OptionText   string   `json:"option_text"`
	ScoreValue   *float64 `json:"score_value,omitempty"`
	IsCorrect    bool     `json:"is_correct,omitempty"`
	DisplayOrder int      `json:"display_order"`
}

type QuestionResponse struct {
	ID               string           `json:"id"`
	QuestionText     string           `json:"question_text"`
	QuestionImageURL string           `json:"question_image_url,omitempty"`
	Explanation      string           `json:"explanation,omitempty"`
	DisplayOrder     int              `json:"display_order"`
	Choices          []ChoiceResponse `json:"choices"`
}

type QuestionCategoryResponse struct {
	ID                 string             `json:"id"`
	Title              string             `json:"title"`
	Description        string             `json:"description,omitempty"`
	DisplayOrder       int                `json:"display_order"`
	InstrumentCategory string             `json:"instrument_category,omitempty"`
	Questions          []QuestionResponse `json:"questions"`
}

type QuestionnaireResponse struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Type             string    `json:"type"`  // "PRE_TEST" | "POST_TEST"
	Scope            string    `json:"scope"` // "APPLICATION" | "EDUCATION"
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

	// Pre/Post-Test Linkage
	SourcePreTestID          *string `json:"source_pre_test_id,omitempty"`
	IsLinkedToPreTest        bool    `json:"is_linked_to_pre_test"`
	PairedQuestionnaireID    *string `json:"paired_questionnaire_id,omitempty"`
	PairedQuestionnaireTitle *string `json:"paired_questionnaire_title,omitempty"`
}

type QuestionnaireDetailResponse struct {
	QuestionnaireResponse
	Categories    []QuestionCategoryResponse `json:"categories"`
	HasAttempted  bool                       `json:"has_attempted"`
	LatestAttempt *MyAttemptResponse         `json:"latest_attempt,omitempty"`
}

type QuestionnairePairResponse struct {
	PreTest  *QuestionnaireDetailResponse `json:"pre_test"`
	PostTest *QuestionnaireDetailResponse `json:"post_test,omitempty"`
}

type PairAttemptResponse struct {
	AttemptID     string    `json:"attempt_id"`
	Score         int       `json:"score"`
	Passed        bool      `json:"passed"`
	ScoreDSMES    *int      `json:"score_dsmes,omitempty"`
	ScoreDSMQR    *float64  `json:"score_dsmqr,omitempty"`
	DSMQRCategory *string   `json:"dsmqr_category,omitempty"`
	CompletedAt   time.Time `json:"completed_at"`
}

type PairComparisonResponse struct {
	PatientID   string               `json:"patient_id"`
	PatientName string               `json:"patient_name"`
	Puskesmas   string               `json:"puskesmas"`
	PreAttempt  *PairAttemptResponse `json:"pre_attempt,omitempty"`
	PostAttempt *PairAttemptResponse `json:"post_attempt,omitempty"`
	ScoreChange *int                 `json:"score_change,omitempty"`
	DSMESChange *int                 `json:"dsmes_change,omitempty"`
	DSMQRChange *float64             `json:"dsmqr_change,omitempty"`
	Status      string               `json:"status"`
}

type PairComparisonRecord struct {
	PatientID         string
	PatientName       string
	Puskesmas         string
	PreAttemptID      *string
	PreScore          *int
	PrePassed         *bool
	PreCompletedAt    *time.Time
	PreScoreDSMES     *int
	PreScoreDSMQR     *float64
	PreDSMQRCategory  *string
	PostAttemptID     *string
	PostScore         *int
	PostPassed        *bool
	PostCompletedAt   *time.Time
	PostScoreDSMES    *int
	PostScoreDSMQR    *float64
	PostDSMQRCategory *string
}

type QuizStats struct {
	TotalQuizzes     int64 `json:"total_quizzes"`
	PublishedQuizzes int64 `json:"published_quizzes"`
	DraftQuizzes     int64 `json:"draft_quizzes"`
	TotalAttempts    int64 `json:"total_attempts"`
	AverageScore     int   `json:"average_score"`
}

type ParticipantResponse struct {
	ID                   string    `json:"id"`
	PatientID            string    `json:"patient_id"`
	PatientName          string    `json:"patient_name"`
	PatientAvatar        string    `json:"patient_avatar,omitempty"`
	Puskesmas            string    `json:"puskesmas"`
	CompletionDate       time.Time `json:"completion_date"`
	Score                int       `json:"score"`
	Passed               bool      `json:"passed"`
	Duration             string    `json:"duration"`
	SelfEfficacyCategory string    `json:"self_efficacy_category,omitempty"`

	// Category-Specific Scores
	ScoreDSMES    *int     `json:"score_dsmes,omitempty"`
	ScoreDSMQR    *float64 `json:"score_dsmqr,omitempty"`
	DSMQRCategory *string  `json:"dsmqr_category,omitempty"`
	ScorePreTest  *int     `json:"score_pre_test,omitempty"`
}

type QuestionAnalysisResponse struct {
	ID                 string `json:"id"`
	QuestionNumber     int    `json:"question_number"`
	QuestionText       string `json:"question_text"`
	InstrumentCategory string `json:"instrument_category,omitempty"` // PRE_TEST | DSMQR | DSMES
	PatientAnswer      string `json:"patient_answer"`
	CorrectAnswer      string `json:"correct_answer"`
	IsCorrect          bool   `json:"is_correct"`
	Explanation        string `json:"explanation"`
}

type ParticipantDetailResponse struct {
	Participant      ParticipantResponse        `json:"participant"`
	QuizTitle        string                     `json:"quiz_title"`
	Type             string                     `json:"type"`
	Scope            string                     `json:"scope"`
	QuestionAnalysis []QuestionAnalysisResponse `json:"question_analysis"`
}

// PatientQuestionnaireItem is the patient-facing questionnaire list item.
// It includes the questionnaire status and whether the patient has completed it.
type PatientQuestionnaireItem struct {
	ID             string  `json:"id"`
	Title          string  `json:"title"`
	Type           string  `json:"type"`
	Description    string  `json:"description,omitempty"`
	EducationID    *string `json:"education_id,omitempty"`
	EducationTitle string  `json:"education_title,omitempty"`
	QuestionCount  int     `json:"question_count"`
	PassingScore   *int    `json:"passing_score,omitempty"`
	Difficulty     *string `json:"difficulty,omitempty"`
	Status         string  `json:"status"` // draft | aktif | nonaktif
	IsCompleted    bool    `json:"is_completed"`
	Score          *int    `json:"score,omitempty"`
}

var DefaultLikertChoices = []ChoiceResponse{
	{ID: "1", OptionText: "Tidak Yakin", DisplayOrder: 1},
	{ID: "2", OptionText: "Kurang Yakin", DisplayOrder: 2},
	{ID: "3", OptionText: "Cukup Yakin", DisplayOrder: 3},
	{ID: "4", OptionText: "Yakin", DisplayOrder: 4},
	{ID: "5", OptionText: "Sangat Yakin", DisplayOrder: 5},
}

func DetermineSelfEfficacyCategory(score int) string {
	switch {
	case score <= 40:
		return "Low Self-Efficacy"
	case score <= 60:
		return "Moderate Self-Efficacy"
	case score <= 80:
		return "Good Self-Efficacy"
	default:
		return "Very High Self-Efficacy"
	}
}

// DetermineDSMQRCategory calculates the category for DSMQR based on score and question count.
// Can be customized with precise cutoffs once formula is provided.
func DetermineDSMQRCategory(score float64, count int) string {
	if count <= 0 {
		return "-"
	}
	avg := score / float64(count)
	switch {
	case avg < 2.0:
		return "Kurang Baik"
	case avg <= 3.5:
		return "Cukup Baik"
	default:
		return "Sangat Baik"
	}
}

type SubmitResultResponse struct {
	AttemptID            string   `json:"attempt_id"`
	QuestionnaireID      string   `json:"questionnaire_id"`
	Type                 string   `json:"type,omitempty"`
	Message              string   `json:"message,omitempty"`
	Score                *int     `json:"score,omitempty"`
	Passed               *bool    `json:"passed,omitempty"`
	TotalQuestions       int      `json:"total_questions,omitempty"`
	CorrectCount         *int     `json:"correct_count,omitempty"`
	SelfEfficacyCategory string   `json:"self_efficacy_category,omitempty"`
	ScoreDSMES           *int     `json:"score_dsmes,omitempty"`
	ScoreDSMQR           *float64 `json:"score_dsmqr,omitempty"`
	DSMQRCategory        *string  `json:"dsmqr_category,omitempty"`
	ScorePreTest         *int     `json:"score_pre_test,omitempty"`
}

// MyAttemptResponse is the patient-facing view of their own attempt for a specific questionnaire.
type MyAttemptResponse struct {
	AttemptID       string    `json:"attempt_id"`
	QuestionnaireID string    `json:"questionnaire_id"`
	Score           int       `json:"score"`
	Passed          bool      `json:"passed"`
	TotalQuestions  int       `json:"total_questions"`
	CorrectCount    int       `json:"correct_count"`
	IncorrectCount  int       `json:"incorrect_count"`
	Percentage      int       `json:"percentage"`
	CompletedAt     time.Time `json:"completed_at"`
}

// MyHistoryItemResponse is a single questionnaire attempt in the patient's history list.
type MyHistoryItemResponse struct {
	AttemptID          string    `json:"attempt_id"`
	QuestionnaireID    string    `json:"questionnaire_id"`
	QuestionnaireTitle string    `json:"questionnaire_title"`
	Type               string    `json:"type"`
	Scope              string    `json:"scope"`
	Score              int       `json:"score"`
	Passed             bool      `json:"passed"`
	TotalQuestions     int       `json:"total_questions"`
	CorrectCount       int       `json:"correct_count"`
	IncorrectCount     int       `json:"incorrect_count"`
	Percentage         int       `json:"percentage"`
	CompletedAt        time.Time `json:"completed_at"`
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

	scopeStr := string(q.Scope)
	if scopeStr == "" {
		if q.EducationID != nil {
			scopeStr = string(domain.ScopeEducation)
		} else {
			scopeStr = string(domain.ScopeApplication)
		}
	}

	return QuestionnaireResponse{
		ID:                q.ID,
		Title:             q.Title,
		Type:              string(q.Type),
		Scope:             scopeStr,
		Description:       q.Description,
		EducationID:       q.EducationID,
		EducationTitle:    eduTitle,
		PassingScore:      q.PassingScore,
		Difficulty:        q.Difficulty,
		Status:            displayStatus(q.Status),
		CategoryCount:     categoryCount,
		QuestionCount:     questionCount,
		ParticipantCount:  participantCount,
		AverageScore:      avgScore,
		CreatedBy:         createdBy,
		CreatedAt:         q.CreatedAt,
		UpdatedAt:         q.UpdatedAt,
		SourcePreTestID:   q.SourcePreTestID,
		IsLinkedToPreTest: q.IsLinkedToPreTest,
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
			var cResps []ChoiceResponse
			if len(quest.Options) > 0 {
				cResps = make([]ChoiceResponse, len(quest.Options))
				for k, opt := range quest.Options {
					cResps[k] = ChoiceResponse{
						ID:           opt.ID,
						OptionText:   opt.OptionText,
						ScoreValue:   opt.ScoreValue,
						IsCorrect:    isAdminOrStaff && opt.IsCorrect,
						DisplayOrder: opt.DisplayOrder,
					}
				}
			} else if q.Type == domain.TypePreTest {
				cResps = DefaultLikertChoices
			}

			imageURL := ""
			if quest.QuestionImageURL != nil {
				imageURL = *quest.QuestionImageURL
			}

			qResps[j] = QuestionResponse{
				ID:               quest.ID,
				QuestionText:     quest.QuestionText,
				QuestionImageURL: imageURL,
				Explanation:      quest.Explanation,
				DisplayOrder:     quest.DisplayOrder,
				Choices:          cResps,
			}
		}

		catResps[i] = QuestionCategoryResponse{
			ID:                 cat.ID,
			Title:              cat.Title,
			Description:        cat.Description,
			DisplayOrder:       cat.DisplayOrder,
			InstrumentCategory: cat.InstrumentCategory,
			Questions:          qResps,
		}
	}

	return QuestionnaireDetailResponse{
		QuestionnaireResponse: ToQuestionnaireResponse(q, totalCat, totalQuest, participantCount, avgScore),
		Categories:            catResps,
	}
}
