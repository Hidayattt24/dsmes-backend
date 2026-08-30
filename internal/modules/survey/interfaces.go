package survey

import (
	"context"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type SurveyRepository interface {
	Create(ctx context.Context, survey *domain.Survey) error
	Update(ctx context.Context, survey *domain.Survey) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.Survey, error)
	List(ctx context.Context, surveyType string, status string, page int, limit int, facilityName string) ([]domain.Survey, int64, error)
	GetActiveSurvey(ctx context.Context, surveyType string) (*domain.Survey, error)
	ListActiveSurveys(ctx context.Context, surveyType string) ([]domain.Survey, error)
	SetActive(ctx context.Context, id string, surveyType string) error

	// Question methods
	ReplaceQuestions(ctx context.Context, surveyID string, questions []domain.SurveyQuestion) error

	// Response methods
	CreateResponse(ctx context.Context, resp *domain.SurveyResponse, answers []domain.SurveyAnswer) error
	GetResponseBySurveyAndPatient(ctx context.Context, surveyID string, patientID string) (*domain.SurveyResponse, error)
	ListResponses(ctx context.Context, surveyID string, page int, limit int, facilityName string) ([]domain.SurveyResponse, int64, error)
	GetAnalytics(ctx context.Context, surveyID string, facilityName string) (*SurveyAnalyticsResponse, error)
	GetAllResponsesForExport(ctx context.Context, surveyID string, facilityName string) ([]domain.SurveyResponse, error)
}

type SurveyService interface {
	CreateSurvey(ctx context.Context, adminID string, req CreateSurveyRequest) (*SurveyDetailResponse, error)
	UpdateSurvey(ctx context.Context, id string, req UpdateSurveyRequest) (*SurveyDetailResponse, error)
	DeleteSurvey(ctx context.Context, id string) error
	GetSurveyByID(ctx context.Context, id string, isPatient bool) (*SurveyDetailResponse, error)
	ListSurveys(ctx context.Context, surveyType string, status string, page int, limit int) ([]SurveyListItemResponse, int64, error)
	ListSurveysForStaff(ctx context.Context, staffID string, surveyType string, status string, page int, limit int) ([]SurveyListItemResponse, int64, error)
	UpdateStatus(ctx context.Context, id string, req UpdateSurveyStatusRequest) (*SurveyDetailResponse, error)
	DuplicateSurvey(ctx context.Context, id string, adminID string) (*SurveyDetailResponse, error)

	// Participant / Patient
	GetActiveSurveysForPatient(ctx context.Context, surveyType string, patientID string) ([]SurveyDetailResponse, error)
	SubmitSurvey(ctx context.Context, surveyID string, patientID string, req SubmitSurveyRequest) (*SubmitSurveyResponse, error)

	// Analytics & Export
	GetSurveyResponses(ctx context.Context, surveyID string, page int, limit int) ([]SurveyResponseItemResponse, int64, error)
	GetSurveyResponsesForStaff(ctx context.Context, staffID string, surveyID string, page int, limit int) ([]SurveyResponseItemResponse, int64, error)
	GetSurveyAnalytics(ctx context.Context, surveyID string) (*SurveyAnalyticsResponse, error)
	GetSurveyAnalyticsForStaff(ctx context.Context, staffID string, surveyID string) (*SurveyAnalyticsResponse, error)
	ExportResponsesCSV(ctx context.Context, surveyID string) ([]byte, string, error)
	ExportResponsesCSVForStaff(ctx context.Context, staffID string, surveyID string) ([]byte, string, error)
}
