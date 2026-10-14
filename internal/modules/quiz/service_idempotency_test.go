package quiz

import (
	"context"
	"testing"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"go.uber.org/zap"
)

type mockQuizRepo struct {
	attempts map[string]*domain.QuizAttempt
	byClient map[string]*domain.QuizAttempt
}

func newMockQuizRepo() *mockQuizRepo {
	return &mockQuizRepo{
		attempts: make(map[string]*domain.QuizAttempt),
		byClient: make(map[string]*domain.QuizAttempt),
	}
}

func (m *mockQuizRepo) FindAll(ctx context.Context, search, qType, status, sortBy, sortOrder string, page, limit int) ([]domain.Questionnaire, int64, error) {
	return nil, 0, nil
}
func (m *mockQuizRepo) FindByID(ctx context.Context, id string) (*domain.Questionnaire, error) {
	return &domain.Questionnaire{
		BaseModel: domain.BaseModel{ID: id},
		Title:     "Pre-Test DMSES",
		Type:      domain.TypePreTest,
		Status:    "aktif",
		Categories: []domain.QuestionCategory{
			{
				Questions: []domain.Question{
					{BaseModel: domain.BaseModel{ID: "q-1"}, QuestionText: "Soal 1"},
				},
			},
		},
	}, nil
}
func (m *mockQuizRepo) GetActivePreTest(ctx context.Context) (*domain.Questionnaire, error) {
	return nil, nil
}
func (m *mockQuizRepo) GetAnyPreTest(ctx context.Context) (*domain.Questionnaire, error) {
	return nil, nil
}
func (m *mockQuizRepo) GetPostTestByEducation(ctx context.Context, educationID string) (*domain.Questionnaire, error) {
	return nil, nil
}
func (m *mockQuizRepo) Create(ctx context.Context, q *domain.Questionnaire) error { return nil }
func (m *mockQuizRepo) Update(ctx context.Context, q *domain.Questionnaire) error { return nil }
func (m *mockQuizRepo) Delete(ctx context.Context, id string) error               { return nil }
func (m *mockQuizRepo) GetStats(ctx context.Context, facilityName string) (*QuizStats, error) {
	return nil, nil
}
func (m *mockQuizRepo) SaveAttempt(ctx context.Context, attempt *domain.QuizAttempt) error {
	attempt.ID = "generated-attempt-1"
	m.attempts[attempt.ID] = attempt
	if attempt.ClientID != nil && *attempt.ClientID != "" {
		m.byClient[attempt.PatientID+":"+*attempt.ClientID] = attempt
	}
	return nil
}
func (m *mockQuizRepo) FindAttemptsByQuestionnaireID(ctx context.Context, questionnaireID string, facilityName string) ([]domain.QuizAttempt, error) {
	return nil, nil
}
func (m *mockQuizRepo) FindAttemptByID(ctx context.Context, questionnaireID string, participantID string, facilityName string) (*domain.QuizAttempt, error) {
	return nil, nil
}
func (m *mockQuizRepo) FindAttemptByClientID(ctx context.Context, patientID string, clientID string) (*domain.QuizAttempt, error) {
	return m.byClient[patientID+":"+clientID], nil
}
func (m *mockQuizRepo) FindMyAttempt(ctx context.Context, patientID, questionnaireID string) (*domain.QuizAttempt, error) {
	return nil, nil
}
func (m *mockQuizRepo) FindMyAttemptByID(ctx context.Context, patientID, questionnaireID, attemptID string) (*domain.QuizAttempt, error) {
	return nil, nil
}
func (m *mockQuizRepo) FindMyHistory(ctx context.Context, patientID, qType string) ([]domain.QuizAttempt, error) {
	return nil, nil
}
func (m *mockQuizRepo) FindActiveForPatient(ctx context.Context, qType string, patientID string, page, perPage int) ([]PatientQuestionnaireItem, int64, error) {
	return nil, 0, nil
}
func (m *mockQuizRepo) CountAttempts(ctx context.Context, questionnaireID string, facilityName string) (int, error) {
	return 0, nil
}
func (m *mockQuizRepo) GetAverageScore(ctx context.Context, questionnaireID string, facilityName string) (*int, error) {
	return nil, nil
}

func TestSubmitQuestionnaire_Idempotency(t *testing.T) {
	repo := newMockQuizRepo()
	svc := NewQuizService(repo, nil, nil, zap.NewNop())

	clientID := "client-quiz-uuid-999"
	val := 4
	req := SubmitQuestionnaireRequest{
		DurationSeconds: 120,
		Answers: []SubmitAnswerItem{
			{QuestionID: "q-1", SelectedValue: &val},
		},
		ClientID: &clientID,
	}

	// 1. First submission
	resp1, err := svc.SubmitQuestionnaire(context.Background(), "patient-1", "questionnaire-1", req)
	if err != nil {
		t.Fatalf("first submission failed: %v", err)
	}

	// 2. Retry with same client_id
	resp2, err := svc.SubmitQuestionnaire(context.Background(), "patient-1", "questionnaire-1", req)
	if err != nil {
		t.Fatalf("second submission failed: %v", err)
	}

	if resp2.AttemptID != resp1.AttemptID {
		t.Errorf("idempotency violated: expected same AttemptID %s, got %s", resp1.AttemptID, resp2.AttemptID)
	}
	if len(repo.attempts) != 1 {
		t.Errorf("expected 1 attempt record, got %d", len(repo.attempts))
	}
}
