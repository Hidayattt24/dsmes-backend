package quiz

import (
	"context"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type QuizRepository interface {
	FindAll(ctx context.Context, search, qType, status, sortBy, sortOrder string, page, limit int) ([]domain.Questionnaire, int64, error)
	FindByID(ctx context.Context, id string) (*domain.Questionnaire, error)
	GetActivePreTest(ctx context.Context) (*domain.Questionnaire, error)
	GetPostTestByEducation(ctx context.Context, educationID string) (*domain.Questionnaire, error)
	Create(ctx context.Context, q *domain.Questionnaire) error
	Update(ctx context.Context, q *domain.Questionnaire) error
	Delete(ctx context.Context, id string) error
	GetStats(ctx context.Context, facilityName string) (*QuizStats, error)

	SaveAttempt(ctx context.Context, attempt *domain.QuizAttempt) error
	FindAttemptsByQuestionnaireID(ctx context.Context, questionnaireID string, facilityName string) ([]domain.QuizAttempt, error)
	FindAttemptByID(ctx context.Context, questionnaireID string, participantID string, facilityName string) (*domain.QuizAttempt, error)
	FindMyAttempt(ctx context.Context, patientID, questionnaireID string) (*domain.QuizAttempt, error)
	FindMyAttemptByID(ctx context.Context, patientID, questionnaireID, attemptID string) (*domain.QuizAttempt, error)
	FindMyHistory(ctx context.Context, patientID, qType string) ([]domain.QuizAttempt, error)
	FindActiveForPatient(ctx context.Context, qType string, patientID string, page, perPage int) ([]PatientQuestionnaireItem, int64, error)
	CountAttempts(ctx context.Context, questionnaireID string, facilityName string) (int, error)
	GetAverageScore(ctx context.Context, questionnaireID string, facilityName string) (*int, error)
}

type QuizService interface {
	ListQuestionnaires(ctx context.Context, search, qType, status, sortBy, sortOrder string, page, limit int) ([]QuestionnaireResponse, int64, error)
	ListQuestionnairesForStaff(ctx context.Context, staffID string, search, qType, status, sortBy, sortOrder string, page, limit int) ([]QuestionnaireResponse, int64, error)
	GetQuestionnaire(ctx context.Context, id string, isAdminOrStaff bool) (*QuestionnaireDetailResponse, error)
	GetQuestionnaireForStaff(ctx context.Context, staffID string, id string) (*QuestionnaireDetailResponse, error)
	GetActivePreTest(ctx context.Context, isAdminOrStaff bool) (*QuestionnaireDetailResponse, error)
	GetPostTestByEducation(ctx context.Context, educationID string, isAdminOrStaff bool) (*QuestionnaireDetailResponse, error)
	CreateQuestionnaire(ctx context.Context, staffID string, req CreateQuestionnaireRequest) (*QuestionnaireDetailResponse, error)
	UpdateQuestionnaire(ctx context.Context, id string, req CreateQuestionnaireRequest) (*QuestionnaireDetailResponse, error)
	DeleteQuestionnaire(ctx context.Context, id string) error
	GetStats(ctx context.Context) (*QuizStats, error)
	GetStatsForStaff(ctx context.Context, staffID string) (*QuizStats, error)

	SubmitQuestionnaire(ctx context.Context, patientID string, questionnaireID string, req SubmitQuestionnaireRequest) (*SubmitResultResponse, error)
	ListParticipants(ctx context.Context, questionnaireID string) ([]ParticipantResponse, error)
	ListParticipantsForStaff(ctx context.Context, staffID string, questionnaireID string) ([]ParticipantResponse, error)
	GetParticipantDetail(ctx context.Context, questionnaireID string, participantID string) (*ParticipantDetailResponse, error)
	GetParticipantDetailForStaff(ctx context.Context, staffID string, questionnaireID string, participantID string) (*ParticipantDetailResponse, error)
	GetMyAttempt(ctx context.Context, patientID, questionnaireID string) (*MyAttemptResponse, error)
	GetMyAttemptDetail(ctx context.Context, patientID, questionnaireID string) (*ParticipantDetailResponse, error)
	GetMyAttemptDetailByID(ctx context.Context, patientID, questionnaireID, attemptID string) (*ParticipantDetailResponse, error)
	GetMyHistory(ctx context.Context, patientID, qType string) ([]MyHistoryItemResponse, error)
	ListPatientQuestionnaires(ctx context.Context, qType string, patientID string, page, perPage int) ([]PatientQuestionnaireItem, int64, error)
}
