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
		Preload("Categories", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Preload("Categories.Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Preload("Categories.Questions.Options", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
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
		var existingCatIDs []string
		if err := tx.Model(&domain.QuestionCategory{}).Where("questionnaire_id = ? AND deleted_at IS NULL", q.ID).Pluck("id", &existingCatIDs).Error; err != nil {
			return errs.NewInternal("failed to query existing categories for update", err)
		}

		if len(existingCatIDs) > 0 {
			var existingQuestIDs []string
			if err := tx.Model(&domain.Question{}).Where("category_id IN ? AND deleted_at IS NULL", existingCatIDs).Pluck("id", &existingQuestIDs).Error; err != nil {
				return errs.NewInternal("failed to query existing questions for update", err)
			}
			if len(existingQuestIDs) > 0 {
				if err := tx.Where("question_id IN ?", existingQuestIDs).Delete(&domain.QuestionOption{}).Error; err != nil {
					return errs.NewInternal("failed to delete existing question options during update", err)
				}
				if err := tx.Where("id IN ?", existingQuestIDs).Delete(&domain.Question{}).Error; err != nil {
					return errs.NewInternal("failed to delete existing questions during update", err)
				}
			}
			if err := tx.Where("id IN ?", existingCatIDs).Delete(&domain.QuestionCategory{}).Error; err != nil {
				return errs.NewInternal("failed to delete existing categories during update", err)
			}
		}

		if err := tx.Save(q).Error; err != nil {
			return errs.NewInternal("failed to update questionnaire", err)
		}

		return nil
	})
}

func (r *quizRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var catIDs []string
		if err := tx.Model(&domain.QuestionCategory{}).Where("questionnaire_id = ? AND deleted_at IS NULL", id).Pluck("id", &catIDs).Error; err != nil {
			return errs.NewInternal("failed to fetch category IDs for deletion", err)
		}

		if len(catIDs) > 0 {
			var questIDs []string
			if err := tx.Model(&domain.Question{}).Where("category_id IN ? AND deleted_at IS NULL", catIDs).Pluck("id", &questIDs).Error; err != nil {
				return errs.NewInternal("failed to fetch question IDs for deletion", err)
			}
			if len(questIDs) > 0 {
				if err := tx.Where("question_id IN ?", questIDs).Delete(&domain.QuestionOption{}).Error; err != nil {
					return errs.NewInternal("failed to delete question options", err)
				}
				if err := tx.Where("id IN ?", questIDs).Delete(&domain.Question{}).Error; err != nil {
					return errs.NewInternal("failed to delete questions", err)
				}
			}
			if err := tx.Where("id IN ?", catIDs).Delete(&domain.QuestionCategory{}).Error; err != nil {
				return errs.NewInternal("failed to delete question categories", err)
			}
		}

		if err := tx.Where("quiz_id = ?", id).Delete(&domain.QuizAttempt{}).Error; err != nil {
			return errs.NewInternal("failed to delete questionnaire attempts", err)
		}
		if err := tx.Delete(&domain.Questionnaire{}, "id = ?", id).Error; err != nil {
			return errs.NewInternal("failed to delete questionnaire", err)
		}
		return nil
	})
}

func (r *quizRepository) GetStats(ctx context.Context) (*QuizStats, error) {
	var total, published, draft, attempts int64
	var avgScore int

	if err := r.db.WithContext(ctx).Model(&domain.Questionnaire{}).Where("deleted_at IS NULL").Count(&total).Error; err != nil {
		return nil, errs.NewInternal("failed to count questionnaires", err)
	}

	if err := r.db.WithContext(ctx).Model(&domain.Questionnaire{}).Where("LOWER(status) IN ('aktif', 'terbit') AND deleted_at IS NULL").Count(&published).Error; err != nil {
		return nil, errs.NewInternal("failed to count published questionnaires", err)
	}

	if err := r.db.WithContext(ctx).Model(&domain.Questionnaire{}).Where("LOWER(status) = 'draft' AND deleted_at IS NULL").Count(&draft).Error; err != nil {
		return nil, errs.NewInternal("failed to count draft questionnaires", err)
	}

	if err := r.db.WithContext(ctx).Table("quiz_attempts").
		Joins("JOIN questionnaires ON questionnaires.id = quiz_attempts.quiz_id AND questionnaires.deleted_at IS NULL").
		Where("quiz_attempts.deleted_at IS NULL").
		Count(&attempts).Error; err != nil {
		return nil, errs.NewInternal("failed to count quiz attempts", err)
	}

	err := r.db.WithContext(ctx).Table("quiz_attempts").
		Joins("JOIN questionnaires ON questionnaires.id = quiz_attempts.quiz_id AND questionnaires.deleted_at IS NULL").
		Where("quiz_attempts.deleted_at IS NULL").
		Select("COALESCE(ROUND(AVG(quiz_attempts.score)), 0)").
		Scan(&avgScore).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NewInternal("failed to fetch average score", err)
	}

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

func (r *quizRepository) FindActiveForPatient(ctx context.Context, qType string, patientID string, page, perPage int) ([]PatientQuestionnaireItem, int64, error) {
	var questionnaires []domain.Questionnaire
	var total int64

	base := r.db.WithContext(ctx).
		Model(&domain.Questionnaire{}).
		Where("LOWER(questionnaires.status) IN ('aktif', 'terbit') AND questionnaires.deleted_at IS NULL")

	if qType != "" {
		base = base.Where("questionnaires.type = ?", qType)
	}

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, errs.NewInternal("failed to count patient questionnaires", err)
	}

	offset := (page - 1) * perPage
	q := r.db.WithContext(ctx).
		Preload("Education").
		Preload("Categories", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Preload("Categories.Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Where("LOWER(questionnaires.status) IN ('aktif', 'terbit') AND questionnaires.deleted_at IS NULL")

	if qType != "" {
		q = q.Where("questionnaires.type = ?", qType)
	}

	if err := q.Order("questionnaires.created_at DESC").Offset(offset).Limit(perPage).Find(&questionnaires).Error; err != nil {
		return nil, 0, errs.NewInternal("failed to fetch patient questionnaires", err)
	}

	items := make([]PatientQuestionnaireItem, 0, len(questionnaires))
	if len(questionnaires) == 0 {
		return items, total, nil
	}

	quizIDs := make([]string, len(questionnaires))
	for i, qn := range questionnaires {
		quizIDs[i] = qn.ID
	}

	var attempts []domain.QuizAttempt
	err := r.db.WithContext(ctx).
		Where("patient_id = ? AND quiz_id IN ? AND deleted_at IS NULL", patientID, quizIDs).
		Order("completed_at DESC").
		Find(&attempts).Error
	if err != nil {
		return nil, 0, errs.NewInternal("failed to fetch patient quiz attempts", err)
	}

	latestAttempts := make(map[string]domain.QuizAttempt)
	for _, att := range attempts {
		if _, exists := latestAttempts[att.QuestionnaireID]; !exists {
			latestAttempts[att.QuestionnaireID] = att
		}
	}

	for _, qn := range questionnaires {
		eduTitle := ""
		if qn.Education != nil {
			eduTitle = qn.Education.Title
		}

		totalQuest := 0
		for _, cat := range qn.Categories {
			totalQuest += len(cat.Questions)
		}

		var isCompleted bool
		var score *int
		if att, ok := latestAttempts[qn.ID]; ok {
			isCompleted = true
			sc := att.Score
			score = &sc
		}

		items = append(items, PatientQuestionnaireItem{
			ID:             qn.ID,
			Title:          qn.Title,
			Type:           string(qn.Type),
			Description:    qn.Description,
			EducationID:    qn.EducationID,
			EducationTitle: eduTitle,
			QuestionCount:  totalQuest,
			PassingScore:   qn.PassingScore,
			Difficulty:     qn.Difficulty,
			IsCompleted:    isCompleted,
			Score:          score,
		})
	}

	return items, total, nil
}

func (r *quizRepository) CountAttempts(ctx context.Context, questionnaireID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.QuizAttempt{}).
		Where("quiz_id = ? AND deleted_at IS NULL", questionnaireID).
		Count(&count).Error
	if err != nil {
		return 0, errs.NewInternal("failed to count attempts for questionnaire", err)
	}
	return int(count), nil
}

func (r *quizRepository) GetAverageScore(ctx context.Context, questionnaireID string) (*int, error) {
	var avg int
	err := r.db.WithContext(ctx).Model(&domain.QuizAttempt{}).
		Where("quiz_id = ? AND deleted_at IS NULL", questionnaireID).
		Select("COALESCE(ROUND(AVG(score)), 0)").
		Scan(&avg).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NewInternal("failed to get average score for questionnaire", err)
	}
	return &avg, nil
}

func (r *quizRepository) FindMyAttempt(ctx context.Context, patientID, questionnaireID string) (*domain.QuizAttempt, error) {
	var attempt domain.QuizAttempt
	err := r.db.WithContext(ctx).
		Preload("Questionnaire").
		Preload("Questionnaire.Education").
		Where("quiz_id = ? AND patient_id = ? AND deleted_at IS NULL", questionnaireID, patientID).
		Order("completed_at DESC").
		First(&attempt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("no attempt found for this questionnaire")
		}
		return nil, errs.NewInternal("failed to fetch patient attempt", err)
	}
	return &attempt, nil
}

func (r *quizRepository) FindMyHistory(ctx context.Context, patientID, qType string) ([]domain.QuizAttempt, error) {
	var attempts []domain.QuizAttempt
	q := r.db.WithContext(ctx).
		Preload("Questionnaire").
		Preload("Questionnaire.Education").
		Joins("JOIN questionnaires ON questionnaires.id = quiz_attempts.quiz_id AND questionnaires.deleted_at IS NULL").
		Where("quiz_attempts.patient_id = ? AND quiz_attempts.deleted_at IS NULL", patientID)

	if qType != "" {
		q = q.Where("questionnaires.type = ?", qType)
	}

	if err := q.Order("quiz_attempts.completed_at DESC").Find(&attempts).Error; err != nil {
		return nil, errs.NewInternal("failed to fetch patient history", err)
	}
	return attempts, nil
}
