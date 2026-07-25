package quiz

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type quizRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewQuizRepository(db *gorm.DB, log *zap.Logger) QuizRepository {
	return &quizRepository{db: db, log: log}
}

func (r *quizRepository) FindAll(ctx context.Context, search, qType, status, sortBy, sortOrder string, page, limit int) ([]domain.Questionnaire, int64, error) {
	var items []domain.Questionnaire
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.Questionnaire{}).Where("questionnaires.deleted_at IS NULL")

	if search != "" {
		pattern := "%" + search + "%"
		q = q.Where("questionnaires.title ILIKE ? OR questionnaires.description ILIKE ?", pattern, pattern)
	}

	if qType != "" && qType != "Semua" {
		q = q.Where("questionnaires.type = ?", qType)
	}

	if status != "" && status != "Semua" {
		q = q.Where("LOWER(questionnaires.status) = ?", normalizeStatus(status))
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errs.NewInternal("failed to count questionnaires", err)
	}

	order := "DESC"
	if sortOrder == "asc" {
		order = "ASC"
	}

	switch sortBy {
	case "title", "name":
		q = q.Order("questionnaires.title " + order)
	case "oldest":
		q = q.Order("questionnaires.created_at ASC")
	default:
		q = q.Order("questionnaires.created_at " + order)
	}

	offset := (page - 1) * limit
	err := q.Preload("Education").
		Preload("Categories.Questions.Options").
		Offset(offset).Limit(limit).
		Find(&items).Error
	if err != nil {
		return nil, 0, errs.NewInternal("failed to fetch questionnaires", err)
	}

	return items, total, nil
}

func (r *quizRepository) FindByID(ctx context.Context, id string) (*domain.Questionnaire, error) {
	var item domain.Questionnaire
	err := r.db.WithContext(ctx).
		Preload("Education").
		Preload("Categories", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Preload("Categories.Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Preload("Categories.Questions.Options", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("questionnaire not found")
		}
		return nil, errs.NewInternal("failed to fetch questionnaire", err)
	}
	return &item, nil
}

func (r *quizRepository) GetActivePreTest(ctx context.Context) (*domain.Questionnaire, error) {
	var item domain.Questionnaire
	err := r.db.WithContext(ctx).
		Preload("Categories", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Preload("Categories.Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Preload("Categories.Questions.Options", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Where("type = ? AND LOWER(status) IN ('aktif', 'terbit') AND deleted_at IS NULL", domain.TypePreTest).
		Order("created_at DESC").
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("active Pre-Test not found")
		}
		return nil, errs.NewInternal("failed to fetch active Pre-Test", err)
	}
	return &item, nil
}

func (r *quizRepository) GetPostTestByEducation(ctx context.Context, educationID string) (*domain.Questionnaire, error) {
	var item domain.Questionnaire
	err := r.db.WithContext(ctx).
		Preload("Education").
		Preload("Categories", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Preload("Categories.Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Preload("Categories.Questions.Options", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Where("type = ? AND education_id = ? AND LOWER(status) IN ('aktif', 'terbit') AND deleted_at IS NULL", domain.TypePostTest, educationID).
		Order("created_at DESC").
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("Post-Test for this education material not found")
		}
		return nil, errs.NewInternal("failed to fetch Post-Test", err)
	}
	return &item, nil
}

func (r *quizRepository) Create(ctx context.Context, q *domain.Questionnaire) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(q).Error; err != nil {
			return errs.NewInternal("failed to create questionnaire", err)
		}
		return nil
	})
}

func (r *quizRepository) Update(ctx context.Context, q *domain.Questionnaire) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Remove existing categories and questions under this questionnaire
		var existingCatIDs []string
		tx.Model(&domain.QuestionCategory{}).Where("questionnaire_id = ?", q.ID).Pluck("id", &existingCatIDs)

		if len(existingCatIDs) > 0 {
			var existingQuestIDs []string
			tx.Model(&domain.Question{}).Where("category_id IN ?", existingCatIDs).Pluck("id", &existingQuestIDs)
			if len(existingQuestIDs) > 0 {
				_ = tx.Where("question_id IN ?", existingQuestIDs).Delete(&domain.QuestionOption{}).Error
				_ = tx.Where("id IN ?", existingQuestIDs).Delete(&domain.Question{}).Error
			}
			_ = tx.Where("id IN ?", existingCatIDs).Delete(&domain.QuestionCategory{}).Error
		}

		// Save updated questionnaire record
		if err := tx.Save(q).Error; err != nil {
			return errs.NewInternal("failed to update questionnaire", err)
		}

		return nil
	})
}

func (r *quizRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Questionnaire{}, "id = ?", id).Error; err != nil {
		return errs.NewInternal("failed to delete questionnaire", err)
	}
	return nil
}

func (r *quizRepository) GetStats(ctx context.Context) (*QuizStats, error) {
	var total, published, draft, attempts int64
	var avgScore int

	_ = r.db.WithContext(ctx).Model(&domain.Questionnaire{}).Where("deleted_at IS NULL").Count(&total)
	_ = r.db.WithContext(ctx).Model(&domain.Questionnaire{}).Where("LOWER(status) IN ('aktif', 'terbit') AND deleted_at IS NULL").Count(&published)
	_ = r.db.WithContext(ctx).Model(&domain.Questionnaire{}).Where("LOWER(status) = 'draft' AND deleted_at IS NULL").Count(&draft)
	_ = r.db.WithContext(ctx).Model(&domain.QuizAttempt{}).Where("deleted_at IS NULL").Count(&attempts)

	_ = r.db.WithContext(ctx).Model(&domain.QuizAttempt{}).
		Where("deleted_at IS NULL").
		Select("COALESCE(ROUND(AVG(score)), 0)").
		Scan(&avgScore)

	return &QuizStats{
		TotalQuizzes:     total,
		PublishedQuizzes: published,
		DraftQuizzes:     draft,
		TotalAttempts:    attempts,
		AverageScore:     avgScore,
	}, nil
}

func (r *quizRepository) SaveAttempt(ctx context.Context, attempt *domain.QuizAttempt) error {
	if err := r.db.WithContext(ctx).Create(attempt).Error; err != nil {
		return errs.NewInternal("failed to save attempt", err)
	}
	return nil
}

func (r *quizRepository) FindAttemptsByQuestionnaireID(ctx context.Context, questionnaireID string) ([]domain.QuizAttempt, error) {
	var attempts []domain.QuizAttempt
	err := r.db.WithContext(ctx).
		Preload("Patient").
		Where("quiz_id = ? AND deleted_at IS NULL", questionnaireID).
		Order("completed_at DESC").
		Find(&attempts).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch attempts", err)
	}
	return attempts, nil
}

func (r *quizRepository) FindAttemptByID(ctx context.Context, questionnaireID string, participantID string) (*domain.QuizAttempt, error) {
	var attempt domain.QuizAttempt
	err := r.db.WithContext(ctx).
		Preload("Patient").
		Preload("Answers.Question.Options").
		Where("quiz_id = ? AND patient_id = ? AND deleted_at IS NULL", questionnaireID, participantID).
		Order("completed_at DESC").
		First(&attempt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("participant attempt not found")
		}
		return nil, errs.NewInternal("failed to fetch participant attempt", err)
	}
	return &attempt, nil
}

func (r *quizRepository) CountAttempts(ctx context.Context, questionnaireID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.QuizAttempt{}).
		Where("quiz_id = ? AND deleted_at IS NULL", questionnaireID).
		Count(&count).Error
	return int(count), err
}

func (r *quizRepository) GetAverageScore(ctx context.Context, questionnaireID string) (*int, error) {
	var avg int
	err := r.db.WithContext(ctx).Model(&domain.QuizAttempt{}).
		Where("quiz_id = ? AND deleted_at IS NULL", questionnaireID).
		Select("COALESCE(ROUND(AVG(score)), 0)").
		Scan(&avg).Error
	if err != nil {
		return nil, err
	}
	return &avg, nil
}
