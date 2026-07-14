package quiz

import (
	"fmt"
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type QuestionRequest struct {
	QuestionText  string `json:"question_text"  validate:"required"`
	OptionA       string `json:"option_a"       validate:"required"`
	OptionB       string `json:"option_b"       validate:"required"`
	OptionC       string `json:"option_c"`
	OptionD       string `json:"option_d"`
	CorrectOption string `json:"correct_option" validate:"required,oneof=A B C D"`
	Explanation   string `json:"explanation"`
}

type CreateQuizRequest struct {
	Title           string            `json:"title"             validate:"required,min=5,max=200"`
	LinkedArticleID string            `json:"linked_article_id" validate:"required,uuid4"`
	Difficulty      string            `json:"difficulty"        validate:"required,oneof=Mudah Sedang Sulit"`
	PassingScore    int               `json:"passing_score"     validate:"min=0,max=100"`
	Status          string            `json:"status"            validate:"required,oneof=Draft Terbit draft terbit"`
	Questions       []QuestionRequest `json:"questions"         validate:"required,min=1"`
}

type QuestionResponse struct {
	ID            string `json:"id"`
	QuestionText  string `json:"question_text"`
	OptionA       string `json:"option_a"`
	OptionB       string `json:"option_b"`
	OptionC       string `json:"option_c"`
	OptionD       string `json:"option_d"`
	CorrectOption string `json:"correct_option,omitempty"` // hide for patient, show for admin/staff
	Explanation   string `json:"explanation,omitempty"`
}

type QuizResponse struct {
	ID                 string    `json:"id"`
	Title              string    `json:"title"`
	LinkedArticleID    string    `json:"linked_article_id"`
	LinkedArticleTitle string    `json:"linked_article_title"`
	Difficulty         string    `json:"difficulty"`
	PassingScore       int       `json:"passing_score"`
	Status             string    `json:"status"`
	QuestionCount      int       `json:"question_count"`
	ParticipantCount   int       `json:"participant_count"`
	AverageScore       *int      `json:"average_score"`
	CreatedBy          string    `json:"created_by"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type QuizDetailResponse struct {
	QuizResponse
	Questions []QuestionResponse `json:"questions"`
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

// formatDuration converts duration in seconds to human-readable form e.g. "4m 12s"
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

// normalizeStatus converts frontend-sent status ("Terbit"/"Draft") to lowercase for storage.
func normalizeStatus(status string) string {
	switch status {
	case "Terbit", "terbit":
		return "terbit"
	case "Draft", "draft":
		return "draft"
	default:
		return status
	}
}

// displayStatus converts stored lowercase status to frontend display case.
func displayStatus(status string) string {
	switch status {
	case "terbit":
		return "Terbit"
	case "draft":
		return "Draft"
	default:
		return status
	}
}

func ToQuizResponse(q *domain.Quiz, questionCount int, participantCount int, avgScore *int) QuizResponse {
	articleTitle := ""
	if q.LinkedArticle != nil {
		articleTitle = q.LinkedArticle.Title
	}

	createdBy := "-"
	if q.CreatedBy != nil {
		createdBy = *q.CreatedBy
	}

	return QuizResponse{
		ID:                 q.ID,
		Title:              q.Title,
		LinkedArticleID:    q.LinkedArticleID,
		LinkedArticleTitle: articleTitle,
		Difficulty:         q.Difficulty,
		PassingScore:       q.PassingScore,
		Status:             displayStatus(q.Status),
		QuestionCount:      questionCount,
		ParticipantCount:   participantCount,
		AverageScore:       avgScore,
		CreatedBy:          createdBy,
		CreatedAt:          q.CreatedAt,
		UpdatedAt:          q.UpdatedAt,
	}
}

func ToQuizDetailResponse(q *domain.Quiz, questionCount int, participantCount int, avgScore *int, isAdminOrStaff bool) QuizDetailResponse {
	resp := QuizDetailResponse{
		QuizResponse: ToQuizResponse(q, questionCount, participantCount, avgScore),
		Questions:    make([]QuestionResponse, len(q.Questions)),
	}

	for i, quest := range q.Questions {
		qResp := QuestionResponse{
			ID:           quest.ID,
			QuestionText: quest.QuestionText,
			OptionA:      quest.OptionA,
			OptionB:      quest.OptionB,
			OptionC:      quest.OptionC,
			OptionD:      quest.OptionD,
		}
		if isAdminOrStaff {
			qResp.CorrectOption = quest.CorrectOption
			qResp.Explanation = quest.Explanation
		}
		resp.Questions[i] = qResp
	}

	return resp
}
