package routine

import (
	"context"
	"testing"
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"go.uber.org/zap"
)

type mockRoutineRepo struct {
	activityLogs map[string]*domain.PatientActivityLog
	byClient     map[string]*domain.PatientActivityLog
}

func newMockRoutineRepo() *mockRoutineRepo {
	return &mockRoutineRepo{
		activityLogs: make(map[string]*domain.PatientActivityLog),
		byClient:     make(map[string]*domain.PatientActivityLog),
	}
}

func (m *mockRoutineRepo) FindAllByPatientID(ctx context.Context, patientID string) ([]domain.Routine, error) {
	return nil, nil
}
func (m *mockRoutineRepo) FindTimeByID(ctx context.Context, id string) (*domain.RoutineTime, error) {
	return nil, nil
}
func (m *mockRoutineRepo) UpdateTime(ctx context.Context, t *domain.RoutineTime) error {
	return nil
}
func (m *mockRoutineRepo) CreateLog(ctx context.Context, log *domain.RoutineLogEntry) error {
	return nil
}
func (m *mockRoutineRepo) FindLogsByPatientAndDate(ctx context.Context, patientID string, dateStr string) ([]domain.RoutineLogEntry, error) {
	return nil, nil
}
func (m *mockRoutineRepo) ReplacePatientRoutines(ctx context.Context, patientID string, routines []domain.Routine) error {
	return nil
}
func (m *mockRoutineRepo) CreateActivityLog(ctx context.Context, log *domain.PatientActivityLog) error {
	log.ID = "generated-activity-" + time.Now().Format("150405.000")
	m.activityLogs[log.ID] = log
	if log.ClientID != nil && *log.ClientID != "" {
		m.byClient[log.PatientID+":"+*log.ClientID] = log
	}
	return nil
}
func (m *mockRoutineRepo) FindActivityLogByClientID(ctx context.Context, patientID string, clientID string) (*domain.PatientActivityLog, error) {
	log, ok := m.byClient[patientID+":"+clientID]
	if !ok {
		return nil, nil
	}
	return log, nil
}
func (m *mockRoutineRepo) FindFreeActivityLogsByPatientAndDate(ctx context.Context, patientID string, dateStr string) ([]domain.PatientActivityLog, error) {
	return nil, nil
}

func TestLogActivity_Idempotency(t *testing.T) {
	repo := newMockRoutineRepo()
	svc := NewRoutineService(repo, zap.NewNop())

	clientID := "client-act-uuid-777"
	req := LogActivityRequest{
		ActivityName:    "Jalan Santai",
		DurationMinutes: 30,
		Intensity:       "Ringan",
		ClientID:        &clientID,
	}

	// First request
	resp1, err := svc.LogActivity(context.Background(), "patient-1", req)
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}

	// Second request with same client_id (lost response / retry simulation)
	resp2, err := svc.LogActivity(context.Background(), "patient-1", req)
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}

	if resp2.ID != resp1.ID {
		t.Errorf("idempotency violated: expected same ID %s, got %s", resp1.ID, resp2.ID)
	}
	if len(repo.activityLogs) != 1 {
		t.Errorf("expected 1 record, got %d", len(repo.activityLogs))
	}
}
