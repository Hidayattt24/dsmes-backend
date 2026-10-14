package blood_sugar

import (
	"context"
	"testing"
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"go.uber.org/zap"
)

type mockBSRepo struct {
	items    map[string]*domain.BloodSugarLog // id -> log
	byClient map[string]*domain.BloodSugarLog // patientId:clientId -> log
}

func newMockBSRepo() *mockBSRepo {
	return &mockBSRepo{
		items:    make(map[string]*domain.BloodSugarLog),
		byClient: make(map[string]*domain.BloodSugarLog),
	}
}

func (m *mockBSRepo) Create(ctx context.Context, log *domain.BloodSugarLog) error {
	log.ID = "generated-server-id-" + log.MeasuredAt.Format("150405")
	m.items[log.ID] = log
	if log.ClientID != nil && *log.ClientID != "" {
		m.byClient[log.PatientID+":"+*log.ClientID] = log
	}
	return nil
}

func (m *mockBSRepo) FindByID(ctx context.Context, id string) (*domain.BloodSugarLog, error) {
	log, ok := m.items[id]
	if !ok {
		return nil, nil
	}
	return log, nil
}

func (m *mockBSRepo) FindByClientID(ctx context.Context, patientID, clientID string) (*domain.BloodSugarLog, error) {
	log, ok := m.byClient[patientID+":"+clientID]
	if !ok {
		return nil, nil
	}
	return log, nil
}

func (m *mockBSRepo) FindAllByPatientID(ctx context.Context, patientID string, page, limit int) ([]domain.BloodSugarLog, int64, error) {
	return nil, 0, nil
}

func (m *mockBSRepo) Update(ctx context.Context, log *domain.BloodSugarLog) error {
	return nil
}

func (m *mockBSRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockBSRepo) GetDistributionForStaff(ctx context.Context, staffID string) (*GlucoseDistributionResponse, error) {
	return nil, nil
}

func (m *mockBSRepo) GetPatientDOB(ctx context.Context, patientID string) (*time.Time, error) {
	t := time.Date(1985, 5, 20, 0, 0, 0, 0, time.UTC)
	return &t, nil
}

func TestLogBloodSugar_Idempotency(t *testing.T) {
	repo := newMockBSRepo()
	svc := NewBloodSugarService(repo, zap.NewNop())

	clientID := "client-uuid-bs-12345"
	req := LogBloodSugarRequest{
		GlucoseValue:        120,
		MeasurementTimeType: domain.TimeFasting,
		MeasuredAt:          time.Now().Format(time.RFC3339),
		ClientID:            &clientID,
	}

	// 1. First submission (simulating original request)
	resp1, err := svc.LogBloodSugar(context.Background(), "patient-1", req)
	if err != nil {
		t.Fatalf("first submission failed: %v", err)
	}
	if resp1.GlucoseValue != 120 {
		t.Errorf("expected 120, got %d", resp1.GlucoseValue)
	}

	initialServerID := resp1.ID

	// 2. Second submission with the exact same client_id (simulating lost response retry)
	resp2, err := svc.LogBloodSugar(context.Background(), "patient-1", req)
	if err != nil {
		t.Fatalf("second submission failed: %v", err)
	}

	// Must return the exact same server ID, not creating a duplicate!
	if resp2.ID != initialServerID {
		t.Errorf("idempotency violation: expected same server ID %s, got new ID %s", initialServerID, resp2.ID)
	}
	if len(repo.items) != 1 {
		t.Errorf("expected exactly 1 record in repository, found %d", len(repo.items))
	}
}
