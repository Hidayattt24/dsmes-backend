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

type followupTestRepository struct {
	mu      sync.Mutex
	dueAt   time.Time
	claimed bool
	deleted []string
}

func (r *followupTestRepository) ClaimDue(_ context.Context, now time.Time, _ int) ([]EducationFollowup, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if now.Before(r.dueAt) || r.claimed {
		return nil, nil
	}
	r.claimed = true
	return []EducationFollowup{{PatientID: "patient", ArticleID: "article", ArticleTitle: "Porsi Makan", OccurrenceID: "education-followup:article:patient"}}, nil
}

func (r *followupTestRepository) FindDeviceTokens(context.Context, string) ([]domain.DeviceToken, error) {
	return []domain.DeviceToken{
		{Token: "valid", PatientID: "patient", NotificationName: "Dayat"},
		{Token: "valid", PatientID: "patient", NotificationName: "Dayat"},
		{Token: "invalid", PatientID: "patient"},
	}, nil
}

func (r *followupTestRepository) DeleteDeviceToken(_ context.Context, _, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deleted = append(r.deleted, token)
	return nil
}

type followupTestSender struct {
	mu      sync.Mutex
	sent    int
	title   string
	body    string
	payload map[string]string
}

type failingViewRepository struct{ EducationRepository }

func (failingViewRepository) FindArticleByID(context.Context, string) (*domain.Article, error) {
	return &domain.Article{Title: "Porsi Makan"}, nil
}

func (failingViewRepository) RecordView(context.Context, *domain.ArticleView) error {
	return errors.New("view storage unavailable")
}

func TestEducationDetailDoesNotSucceedWithoutViewRecord(t *testing.T) {
	service := &educationService{repo: failingViewRepository{}, log: zap.NewNop()}
	patientID := "patient"
	article, err := service.GetArticle(context.Background(), "article", &patientID)
	if err == nil || article != nil {
		t.Fatalf("detail succeeded without a persisted view: article=%v error=%v", article, err)
	}
}

func (s *followupTestSender) Send(_ context.Context, token, title, body string, data map[string]string) (string, error) {
	if token == "invalid" {
		return "", errors.New("NotRegistered")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sent++
	s.title, s.body, s.payload = title, body, data
	return "message-id", nil
}

func TestEducationFollowupDispatchOnceAfterDue(t *testing.T) {
	due := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	repo := &followupTestRepository{dueAt: due}
	sender := &followupTestSender{}
	if err := DispatchDueEducationFollowups(context.Background(), repo, sender, due.Add(-time.Second), zap.NewNop()); err != nil {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	for range 2 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			if err := DispatchDueEducationFollowups(context.Background(), repo, sender, due, zap.NewNop()); err != nil {
				t.Errorf("concurrent worker: %v", err)
			}
		}()
	}
	workers.Wait()
	if err := DispatchDueEducationFollowups(context.Background(), repo, sender, due.Add(time.Minute), zap.NewNop()); err != nil {
		t.Fatal(err)
	}
	if sender.sent != 1 || sender.title != "Pengingat Edukasi: Porsi Makan 📖" ||
		sender.body != "Hai, Dayat! Kamu belum membuka materi ini. Yuk pelajari saat sempat." {
		t.Fatalf("unexpected followup: sends=%d title=%q body=%q", sender.sent, sender.title, sender.body)
	}
	if sender.payload["type"] != "education" || sender.payload["action_type"] != "education_detail" ||
		sender.payload["article_id"] != "article" || sender.payload["occurrence_id"] != "education-followup:article:patient" {
		t.Fatalf("incorrect followup payload: %v", sender.payload)
	}
	if len(repo.deleted) != 1 || repo.deleted[0] != "invalid" {
		t.Fatalf("invalid token was not cleaned: %v", repo.deleted)
	}
}
