package education

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/bootstrap"
	"github.com/dsmes/dsmes-backend/internal/config"
	"github.com/dsmes/dsmes-backend/internal/domain"
)

// All article and inbox writes are rolled back. Enable explicitly for the configured DB.
func TestEducationPublicationInboxRollback(t *testing.T) {
	if os.Getenv("RUN_REMOTE_EDUCATION_TEST") != "1" {
		t.Skip("set RUN_REMOTE_EDUCATION_TEST=1 to run the rollback test")
	}
	t.Chdir("../../..")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	logger := zap.NewNop()
	db, err := bootstrap.NewDatabase(cfg, logger)
	if err != nil {
		t.Fatal(err)
	}
	defer bootstrap.CloseDatabase(db, logger)

	var patients int64
	if err := db.Model(&domain.Patient{}).Count(&patients).Error; err != nil {
		t.Fatal(err)
	}
	if patients == 0 {
		t.Skip("database has no active patients")
	}
	rollback := errors.New("intentional rollback")
	err = db.Transaction(func(tx *gorm.DB) error {
		ctx := context.Background()
		repo := NewEducationRepository(tx, logger)
		category := &domain.ArticleCategory{Name: "notif-test-" + uuid.NewString()[:8]}
		if err := tx.Create(category).Error; err != nil {
			return err
		}

		draft := &domain.Article{Title: "notification-test-draft", CategoryID: category.ID, Status: domain.StatusDraft}
		if n, err := repo.CreateArticle(ctx, draft); err != nil || n != 0 {
			t.Fatalf("create draft: count=%d err=%v", n, err)
		}
		draft.Status = domain.StatusPublikasi
		if n, err := repo.UpdateArticle(ctx, draft); err != nil || n != patients {
			t.Fatalf("publish draft: count=%d want=%d err=%v", n, patients, err)
		}
		if n, err := repo.PublishArticle(ctx, draft.ID); err != nil || n != 0 {
			t.Fatalf("repeat publication: count=%d err=%v", n, err)
		}

		published := &domain.Article{Title: "notification-test-published", CategoryID: category.ID, Status: domain.StatusPublikasi}
		if n, err := repo.CreateArticle(ctx, published); err != nil || n != patients {
			t.Fatalf("create published: count=%d want=%d err=%v", n, patients, err)
		}
		var openedPatient domain.Patient
		if err := tx.Select("id").First(&openedPatient).Error; err != nil {
			return err
		}
		if err := repo.RecordView(ctx, &domain.ArticleView{ArticleID: published.ID, PatientID: &openedPatient.ID, ViewedAt: time.Now()}); err != nil {
			return err
		}
		unpublished := &domain.Article{Title: "notification-test-unpublished", CategoryID: category.ID, Status: domain.StatusPublikasi}
		if _, err := repo.CreateArticle(ctx, unpublished); err != nil {
			return err
		}
		if err := tx.Model(unpublished).Update("status", domain.StatusDraft).Error; err != nil {
			return err
		}
		deleted := &domain.Article{Title: "notification-test-deleted", CategoryID: category.ID, Status: domain.StatusPublikasi}
		if _, err := repo.CreateArticle(ctx, deleted); err != nil {
			return err
		}
		if err := tx.Delete(deleted).Error; err != nil {
			return err
		}
		for _, article := range []*domain.Article{draft, published} {
			var count int64
			if err := tx.Model(&domain.NotificationLog{}).Where("article_id = ? AND action_type = ?", article.ID, "education_detail").Count(&count).Error; err != nil {
				return err
			}
			if count != patients {
				t.Fatalf("article %s: inbox=%d want=%d", article.ID, count, patients)
			}
			var initial domain.NotificationLog
			if err := tx.Where("article_id = ? AND education_followup_due_at IS NOT NULL", article.ID).First(&initial).Error; err != nil {
				return err
			}
			if initial.EducationFollowupDueAt.Sub(initial.NotifiedAt) != 48*time.Hour {
				t.Fatalf("article %s: followup due at wrong time", article.ID)
			}
		}
		followups := NewEducationFollowupRepository(tx)
		beforeDue, err := followups.ClaimDue(ctx, time.Now().Add(47*time.Hour), 1000)
		if err != nil {
			return err
		}
		for _, item := range beforeDue {
			if item.ArticleID == draft.ID || item.ArticleID == published.ID {
				t.Fatal("followup claimed before 48 hours")
			}
		}
		claimed, err := followups.ClaimDue(ctx, time.Now().Add(49*time.Hour), 1000)
		if err != nil {
			return err
		}
		counts := map[string]int{}
		for _, item := range claimed {
			counts[item.ArticleID]++
		}
		if counts[draft.ID] != int(patients) || counts[published.ID] != int(patients)-1 {
			t.Fatalf("followup counts: draft=%d published=%d want=%d/%d", counts[draft.ID], counts[published.ID], patients, patients-1)
		}
		if counts[unpublished.ID] != 0 || counts[deleted.ID] != 0 {
			t.Fatalf("unpublished or deleted article received followups: %v", counts)
		}
		again, err := followups.ClaimDue(ctx, time.Now().Add(49*time.Hour), 1000)
		if err != nil {
			return err
		}
		for _, item := range again {
			if item.ArticleID == draft.ID || item.ArticleID == published.ID {
				t.Fatal("followup claimed twice")
			}
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("transaction did not roll back cleanly: %v", err)
	}
}
