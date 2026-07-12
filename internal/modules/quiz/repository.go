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

func (r *quizRepository) FindAll(ctx context.Context, search string, status string, page, limit int) ([]domain.Quiz, int64, error) {
	var items []domain.Quiz
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.Quiz{}).Where("deleted_at IS NULL")

	if search != "" {
		searchPattern := "%" + search + "%"
		q = q.Where("title ILIKE ?", searchPattern)
	}

	if status != "" && status != "Semua" {
		q = q.Where("status = ?", status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errs.NewInternal("failed to count quizzes", err)
	}

	offset := (page - 1) * limit
	err := q.Preload("LinkedArticle").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&items).Error
	if err != nil {
		return nil, 0, errs.NewInternal("failed to fetch quizzes", err)
	}

	return items, total, nil
}

func (r *quizRepository) FindByID(ctx context.Context, id string) (*domain.Quiz, error) {
	var q domain.Quiz
	err := r.db.WithContext(ctx).
		Preload("LinkedArticle").
		Preload("Questions").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&q).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("quiz not found")
		}
		return nil, errs.NewInternal("failed to fetch quiz", err)
	}
	return &q, nil
}

func (r *quizRepository) Create(ctx context.Context, q *domain.Quiz) error {
	if err := r.db.WithContext(ctx).Create(q).Error; err != nil {
		return errs.NewInternal("failed to create quiz", err)
	}
	return nil
}

func (r *quizRepository) Update(ctx context.Context, q *domain.Quiz) error {
	result := r.db.WithContext(ctx).Save(q)
	if result.Error != nil {
		return errs.NewInternal("failed to update quiz", result.Error)
	}
	return nil
}

func (r *quizRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Model(&domain.Quiz{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()"))
	if result.Error != nil {
		return errs.NewInternal("failed to soft delete quiz", result.Error)
	}
	if result.RowsAffected == 0 {
		return errs.NewNotFound("quiz not found")
	}
	return nil
}

func (r *quizRepository) ReplaceQuestions(ctx context.Context, quizID string, questions []domain.QuizQuestion) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Clean existing questions
		tx.Where("quiz_id = ?", quizID).Delete(&domain.QuizQuestion{})

		// Create new questions
		for _, q := range questions {
			q.QuizID = quizID
			if err := tx.Create(&q).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *quizRepository) GetStats(ctx context.Context) (*QuizStats, error) {
	var totalQuizzes int64
	var publishedQuizzes int64
	var draftQuizzes int64
	var totalAttempts int64
	var averageScore float64

	if err := r.db.WithContext(ctx).Model(&domain.Quiz{}).Where("deleted_at IS NULL").Count(&totalQuizzes).Error; err != nil {
		return nil, errs.NewInternal("failed to count quizzes", err)
	}

	if err := r.db.WithContext(ctx).Model(&domain.Quiz{}).Where("status = ? AND deleted_at IS NULL", "terbit").Count(&publishedQuizzes).Error; err != nil {
		return nil, errs.NewInternal("failed to count published quizzes", err)
	}

	draftQuizzes = totalQuizzes - publishedQuizzes

	if err := r.db.WithContext(ctx).Model(&domain.QuizAttempt{}).Where("deleted_at IS NULL").Count(&totalAttempts).Error; err != nil {
		return nil, errs.NewInternal("failed to count attempts", err)
	}

	var scoreResult struct {
		AvgScore float64
	}
	err := r.db.WithContext(ctx).Model(&domain.QuizAttempt{}).
		Where("deleted_at IS NULL").
		Select("COALESCE(AVG(score), 0) as avg_score").
		Scan(&scoreResult).Error
	if err != nil {
		return nil, errs.NewInternal("failed to calculate average score", err)
	}
	averageScore = scoreResult.AvgScore

	return &QuizStats{
		TotalQuizzes:     totalQuizzes,
		PublishedQuizzes: publishedQuizzes,
		DraftQuizzes:     draftQuizzes,
		TotalAttempts:    totalAttempts,
		AverageScore:     int(averageScore),
	}, nil
}

func (r *quizRepository) FindAttemptsByQuizID(ctx context.Context, quizID string) ([]domain.QuizAttempt, error) {
	var attempts []domain.QuizAttempt
	err := r.db.WithContext(ctx).
		Preload("Patient").
		Preload("Patient.AssignedStaff").
		Where("quiz_id = ? AND deleted_at IS NULL", quizID).
		Order("completed_at DESC").
		Find(&attempts).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch attempts", err)
	}
	return attempts, nil
}

func (r *quizRepository) FindAttemptByID(ctx context.Context, quizID string, participantID string) (*domain.QuizAttempt, error) {
	var attempt domain.QuizAttempt
	err := r.db.WithContext(ctx).
		Preload("Patient").
		Preload("Patient.AssignedStaff").
		Preload("Answers").
		Preload("Answers.Question").
		Where("quiz_id = ? AND patient_id = ? AND deleted_at IS NULL", quizID, participantID).
		First(&attempt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("quiz attempt not found")
		}
		return nil, errs.NewInternal("failed to fetch attempt details", err)
	}
	return &attempt, nil
}

func (r *quizRepository) CountAttempts(ctx context.Context, quizID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.QuizAttempt{}).
		Where("quiz_id = ? AND deleted_at IS NULL", quizID).
		Count(&count).Error
	if err != nil {
		return 0, errs.NewInternal("failed to count attempts for quiz", err)
	}
	return int(count), nil
}

func (r *quizRepository) GetAverageScore(ctx context.Context, quizID string) (*int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.QuizAttempt{}).Where("quiz_id = ? AND deleted_at IS NULL", quizID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}

	var scoreResult struct {
		AvgScore float64
	}
	err := r.db.WithContext(ctx).Model(&domain.QuizAttempt{}).
		Where("quiz_id = ? AND deleted_at IS NULL", quizID).
		Select("AVG(score) as avg_score").
		Scan(&scoreResult).Error
	if err != nil {
		return nil, errs.NewInternal("failed to calculate average score for quiz", err)
	}
	res := int(scoreResult.AvgScore)
	return &res, nil
}
