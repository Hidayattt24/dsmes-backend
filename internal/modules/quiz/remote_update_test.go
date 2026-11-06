package quiz

import (
	"context"
	"errors"
	"os"
	"testing"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/bootstrap"
	"github.com/dsmes/dsmes-backend/internal/config"
	"github.com/dsmes/dsmes-backend/internal/domain"
)

// The transaction rolls back its fixture. Enable explicitly for the configured DB.
func TestUpdateEducationQuestionnaireDoesNotSavePartialArticle(t *testing.T) {
	if os.Getenv("RUN_REMOTE_QUIZ_UPDATE_TEST") != "1" {
		t.Skip("set RUN_REMOTE_QUIZ_UPDATE_TEST=1 to run the rollback test")
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

	var article domain.Article
	if err := db.Select("id", "title", "category_id").First(&article).Error; err != nil {
		t.Skipf("no education article available: %v", err)
	}
	rollback := errors.New("intentional rollback")
	err = db.Transaction(func(tx *gorm.DB) error {
		ctx := context.Background()
		repo := NewQuizRepository(tx, logger)
		questionnaire := &domain.Questionnaire{
			Title: "Regression Pre-Test", Type: domain.TypePreTest,
			Description: "Deskripsi uji yang harus tampil kembali saat edit",
			Scope: domain.ScopeEducation, EducationID: &article.ID, Status: "draft",
			Categories: []domain.QuestionCategory{{
				Title: "Original", InstrumentCategory: "PRE_TEST",
				Questions: []domain.Question{{QuestionText: "Original question"}},
			}},
		}
		if err := repo.Create(ctx, questionnaire); err != nil {
			return err
		}
		loaded, err := repo.FindByID(ctx, questionnaire.ID)
		if err != nil {
			return err
		}
		if loaded.Description != questionnaire.Description {
			t.Fatalf("description did not survive create and reload: %q", loaded.Description)
		}
		if loaded.Education == nil || loaded.Education.CategoryID != "" {
			t.Fatalf("expected a partially preloaded article, got %+v", loaded.Education)
		}
		loaded.Title = "Updated Pre-Test"
		loaded.Description = "Deskripsi setelah edit"
		loaded.Categories = []domain.QuestionCategory{{
			Title: "Updated", InstrumentCategory: "PRE_TEST",
			Questions: []domain.Question{{QuestionText: "Updated question"}},
		}}
		if err := repo.Update(ctx, loaded); err != nil {
			return err
		}
		updated, err := repo.FindByID(ctx, questionnaire.ID)
		if err != nil {
			return err
		}
		if updated.Title != "Updated Pre-Test" || updated.Description != "Deskripsi setelah edit" ||
			updated.EducationID == nil || *updated.EducationID != article.ID ||
			len(updated.Categories) != 1 || len(updated.Categories[0].Questions) != 1 ||
			updated.Categories[0].Questions[0].QuestionText != "Updated question" {
			t.Fatalf("questionnaire update was not persisted correctly: %+v", updated)
		}
		var unchanged domain.Article
		if err := tx.Select("id", "title", "category_id").First(&unchanged, "id = ?", article.ID).Error; err != nil {
			return err
		}
		if unchanged.Title != article.Title || unchanged.CategoryID != article.CategoryID {
			t.Fatalf("linked article changed: %+v", unchanged)
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("rollback test failed: %v", err)
	}
}
