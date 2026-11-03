package education

import (
	"context"
	"errors"
	"os"
	"testing"

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
		for _, article := range []*domain.Article{draft, published} {
			var count int64
			if err := tx.Model(&domain.NotificationLog{}).Where("article_id = ? AND action_type = ?", article.ID, "education_detail").Count(&count).Error; err != nil {
				return err
			}
			if count != patients {
				t.Fatalf("article %s: inbox=%d want=%d", article.ID, count, patients)
			}
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("transaction did not roll back cleanly: %v", err)
	}
}
