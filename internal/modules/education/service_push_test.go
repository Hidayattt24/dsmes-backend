package education

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type pushTestRepository struct {
	EducationRepository
	mu      sync.Mutex
	devices []domain.DeviceToken
	deleted []string
}

func (r *pushTestRepository) FindAllDeviceTokens(context.Context) ([]domain.DeviceToken, error) {
	return r.devices, nil
}
func (r *pushTestRepository) DeleteDeviceToken(_ context.Context, _ string, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deleted = append(r.deleted, token)
	return nil
}

type pushTestSender struct {
	mu          sync.Mutex
	sent        []string
	sentAt      []string
	goodSent    chan struct{}
	releaseSlow chan struct{}
}

func (s *pushTestSender) Send(ctx context.Context, token, _, _ string, data map[string]string) (string, error) {
	if token == "slow" {
		select {
		case <-s.releaseSlow:
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	if token == "invalid" {
		return "", errors.New("NotRegistered")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sent = append(s.sent, token)
	s.sentAt = append(s.sentAt, data["sent_at"])
	if token == "good" {
		close(s.goodSent)
	}
	return "message-id", nil
}

func TestEducationPushInvalidTokenAndConcurrentDelivery(t *testing.T) {
	repo := &pushTestRepository{devices: []domain.DeviceToken{
		{Token: "slow", PatientID: "one"},
		{Token: "invalid", PatientID: "two"},
		{Token: "good", PatientID: "three"},
	}}
	sender := &pushTestSender{goodSent: make(chan struct{}), releaseSlow: make(chan struct{})}
	service := &educationService{repo: repo, log: zap.NewNop(), push: sender}
	done := make(chan struct{})
	go func() { service.dispatchEducationPush("Artikel Baru", "article-id"); close(done) }()
	select {
	case <-sender.goodSent:
	case <-time.After(2 * time.Second):
		t.Fatal("valid device waited for a slow device")
	}
	close(sender.releaseSlow)
	<-done
	if len(sender.sent) != 2 {
		t.Fatalf("successful sends=%d, want 2", len(sender.sent))
	}
	for _, stamp := range sender.sentAt {
		if _, err := time.Parse(time.RFC3339Nano, stamp); err != nil {
			t.Fatalf("invalid sent_at payload: %v", err)
		}
	}
	if len(repo.deleted) != 1 || repo.deleted[0] != "invalid" {
		t.Fatalf("deleted tokens=%v, want invalid token only", repo.deleted)
	}
}
