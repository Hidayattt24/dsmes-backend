package quiz

import (
	"context"
	"errors"
	"time"

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

func (r *quizRepository) FindAll(ctx context.Context, search, qType, status, scope, sortBy, sortOrder string, page, limit int) ([]domain.Questionnaire, int64, error) {
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

	if scope != "" && scope != "Semua" {
		q = q.Where("questionnaires.scope = ?", scope)
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
		Preload("Education", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "title")
		}).
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
		Where("type = ? AND scope = ? AND education_id IS NULL AND LOWER(status) IN ('aktif', 'terbit') AND deleted_at IS NULL", domain.TypePreTest, domain.ScopeApplication).
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

func (r *quizRepository) GetPreTestByEducation(ctx context.Context, educationID string) (*domain.Questionnaire, error) {
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
		Where("type = ? AND scope = ? AND education_id = ? AND LOWER(status) IN ('aktif', 'terbit') AND deleted_at IS NULL", domain.TypePreTest, domain.ScopeEducation, educationID).
		Order("created_at DESC").
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("Pre-Test for this education material not found")
		}
		return nil, errs.NewInternal("failed to fetch Pre-Test", err)
	}
	return &item, nil
}

func (r *quizRepository) GetAnyPreTest(ctx context.Context) (*domain.Questionnaire, error) {
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
		Where("type = ? AND deleted_at IS NULL", domain.TypePreTest).
		Order("created_at DESC").
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("Pre-Test not found")
		}
		return nil, errs.NewInternal("failed to fetch Pre-Test", err)
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
		Where("type = ? AND scope = ? AND education_id = ? AND LOWER(status) IN ('aktif', 'terbit') AND deleted_at IS NULL", domain.TypePostTest, domain.ScopeEducation, educationID).
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

func (r *quizRepository) FindEligiblePreTests(ctx context.Context, scope domain.QuestionnaireScope, educationID *string) ([]domain.Questionnaire, error) {
	var items []domain.Questionnaire
	q := r.db.WithContext(ctx).
		Preload("Education", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "title")
		}).
		Preload("Categories.Questions").
		Where("questionnaires.type = ? AND questionnaires.scope = ? AND questionnaires.deleted_at IS NULL", domain.TypePreTest, scope).
		Where("NOT EXISTS (SELECT 1 FROM questionnaires post WHERE post.source_pre_test_id = questionnaires.id AND post.deleted_at IS NULL)")
	if scope == domain.ScopeEducation {
		q = q.Where("questionnaires.education_id = ?", *educationID)
	} else {
		q = q.Where("questionnaires.education_id IS NULL")
	}
	if err := q.Order("questionnaires.created_at DESC").Find(&items).Error; err != nil {
		return nil, errs.NewInternal("failed to fetch eligible Pre-Tests", err)
	}
	return items, nil
}

func (r *quizRepository) FindPostTestBySource(ctx context.Context, sourcePreTestID string) (*domain.Questionnaire, error) {
	var item domain.Questionnaire
	err := r.db.WithContext(ctx).
		Preload("Education").
		Preload("Categories", func(db *gorm.DB) *gorm.DB { return db.Order("display_order ASC") }).
		Preload("Categories.Questions", func(db *gorm.DB) *gorm.DB { return db.Order("display_order ASC") }).
		Preload("Categories.Questions.Options", func(db *gorm.DB) *gorm.DB { return db.Order("display_order ASC") }).
		Where("source_pre_test_id = ? AND deleted_at IS NULL", sourcePreTestID).
		First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, errs.NewInternal("failed to fetch linked Post-Test", err)
	}
	return &item, nil
}

func (r *quizRepository) FindPairedQuestionnaire(ctx context.Context, q *domain.Questionnaire) (*domain.Questionnaire, error) {
	var paired domain.Questionnaire
	query := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if q.Type == domain.TypePostTest && q.SourcePreTestID != nil {
		query = query.Where("id = ?", *q.SourcePreTestID)
	} else {
		query = query.Where("source_pre_test_id = ?", q.ID)
	}
	if err := query.First(&paired).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errs.NewInternal("failed to fetch paired questionnaire", err)
	}
	return &paired, nil
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

func (r *quizRepository) UpdateStatus(ctx context.Context, id, status string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Questionnaire{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]interface{}{"status": status, "updated_at": time.Now()})
	if result.Error != nil {
		return errs.NewInternal("failed to update questionnaire status", result.Error)
	}
	if result.RowsAffected == 0 {
		return errs.NewNotFound("questionnaire not found")
	}
	return nil
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

func (r *quizRepository) GetStats(ctx context.Context, facilityName string) (*QuizStats, error) {
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

	qAttempts := r.db.WithContext(ctx).Table("quiz_attempts").
		Joins("JOIN questionnaires ON questionnaires.id = quiz_attempts.quiz_id AND questionnaires.deleted_at IS NULL").
		Joins("JOIN patients p ON p.id = quiz_attempts.patient_id").
		Where("quiz_attempts.deleted_at IS NULL")
	if facilityName != "" {
		qAttempts = qAttempts.Where("p.health_facility = ?", facilityName)
	}
	if err := qAttempts.Count(&attempts).Error; err != nil {
		return nil, errs.NewInternal("failed to count quiz attempts", err)
	}

	qAvg := r.db.WithContext(ctx).Table("quiz_attempts").
		Joins("JOIN questionnaires ON questionnaires.id = quiz_attempts.quiz_id AND questionnaires.deleted_at IS NULL").
		Joins("JOIN patients p ON p.id = quiz_attempts.patient_id").
		Where("quiz_attempts.deleted_at IS NULL")
	if facilityName != "" {
		qAvg = qAvg.Where("p.health_facility = ?", facilityName)
	}
	err := qAvg.Select("COALESCE(ROUND(AVG(quiz_attempts.score)), 0)").
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

func (r *quizRepository) FindAttemptsByQuestionnaireID(ctx context.Context, questionnaireID string, facilityName string) ([]domain.QuizAttempt, error) {
	var attempts []domain.QuizAttempt
	q := r.db.WithContext(ctx).
		Preload("Patient").
		Joins("JOIN patients p ON p.id = quiz_attempts.patient_id AND p.deleted_at IS NULL").
		Where("quiz_attempts.quiz_id = ? AND quiz_attempts.deleted_at IS NULL", questionnaireID)
	if facilityName != "" {
		q = q.Where("p.health_facility = ?", facilityName)
	}
	err := q.Order("quiz_attempts.completed_at DESC").
		Limit(500). // server-side cap to bound unbounded admin queries
		Find(&attempts).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch attempts", err)
	}
	return attempts, nil
}

func (r *quizRepository) FindAttemptByID(ctx context.Context, questionnaireID string, participantID string, facilityName string) (*domain.QuizAttempt, error) {
	var attempt domain.QuizAttempt
	q := r.db.WithContext(ctx).
		Preload("Questionnaire").
		Preload("Questionnaire.Education").
		Preload("Questionnaire.Categories.Questions.Options").
		Preload("Patient").
		Preload("Answers").
		Where("quiz_attempts.quiz_id = ? AND quiz_attempts.patient_id = ? AND quiz_attempts.deleted_at IS NULL", questionnaireID, participantID)

	if facilityName != "" {
		q = q.Joins("JOIN patients ON patients.id = quiz_attempts.patient_id").
			Where("patients.treatment_facility = ?", facilityName)
	}

	err := q.Order("quiz_attempts.completed_at DESC").
		First(&attempt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("participant attempt not found")
		}
		return nil, errs.NewInternal("failed to fetch participant attempt", err)
	}
	return &attempt, nil
}

func (r *quizRepository) FindAttemptByClientID(ctx context.Context, patientID string, clientID string) (*domain.QuizAttempt, error) {
	var attempt domain.QuizAttempt
	err := r.db.WithContext(ctx).
		Preload("Questionnaire").
		Preload("Questionnaire.Education").
		Preload("Answers").
		Where("patient_id = ? AND client_id = ? AND deleted_at IS NULL", patientID, clientID).
		First(&attempt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("attempt not found")
		}
		return nil, errs.NewInternal("failed to fetch attempt by client_id", err)
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
	type questionCount struct {
		QuestionnaireID string `gorm:"column:questionnaire_id"`
		Count           int    `gorm:"column:count"`
	}
	var counts []questionCount
	if err := r.db.WithContext(ctx).
		Table("questions q").
		Select("c.questionnaire_id, COUNT(q.id) AS count").
		Joins("JOIN question_categories c ON c.id = q.category_id AND c.deleted_at IS NULL").
		Where("q.category_id IS NOT NULL AND q.deleted_at IS NULL AND c.questionnaire_id IN ?", quizIDs).
		Group("c.questionnaire_id").
		Scan(&counts).Error; err != nil {
		return nil, 0, errs.NewInternal("failed to count questionnaire questions", err)
	}
	questionCounts := make(map[string]int, len(counts))
	for _, count := range counts {
		questionCounts[count.QuestionnaireID] = count.Count
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

		totalQuest := questionCounts[qn.ID]

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
			Status:         normalizeStatus(qn.Status),
			IsCompleted:    isCompleted,
			Score:          score,
		})
	}

	return items, total, nil
}

func (r *quizRepository) CountAttempts(ctx context.Context, questionnaireID string, facilityName string) (int, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&domain.QuizAttempt{}).
		Joins("JOIN patients p ON p.id = quiz_attempts.patient_id AND p.deleted_at IS NULL").
		Where("quiz_attempts.quiz_id = ? AND quiz_attempts.deleted_at IS NULL", questionnaireID)
	if facilityName != "" {
		q = q.Where("p.health_facility = ?", facilityName)
	}
	err := q.Count(&count).Error
	if err != nil {
		return 0, errs.NewInternal("failed to count attempts for questionnaire", err)
	}
	return int(count), nil
}

func (r *quizRepository) GetAverageScore(ctx context.Context, questionnaireID string, facilityName string) (*int, error) {
	var avg int
	q := r.db.WithContext(ctx).Model(&domain.QuizAttempt{}).
		Joins("JOIN patients p ON p.id = quiz_attempts.patient_id AND p.deleted_at IS NULL").
		Where("quiz_attempts.quiz_id = ? AND quiz_attempts.deleted_at IS NULL", questionnaireID)
	if facilityName != "" {
		q = q.Where("p.health_facility = ?", facilityName)
	}
	err := q.Select("COALESCE(ROUND(AVG(quiz_attempts.score)), 0)").
		Scan(&avg).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NewInternal("failed to get average score for questionnaire", err)
	}
	return &avg, nil
}

func (r *quizRepository) FindPairComparisons(ctx context.Context, preTestID, postTestID, facilityName string) ([]PairComparisonRecord, error) {
	var rows []PairComparisonRecord
	query := r.db.WithContext(ctx).Raw(`
WITH first_pre AS (
    SELECT DISTINCT ON (patient_id) patient_id, id, score, passed, score_dsmes, score_dsmqr, dsmqr_category, completed_at
    FROM quiz_attempts
    WHERE quiz_id = ? AND deleted_at IS NULL
    ORDER BY patient_id, completed_at ASC, created_at ASC
), latest_post AS (
    SELECT DISTINCT ON (patient_id) patient_id, id, score, passed, score_dsmes, score_dsmqr, dsmqr_category, completed_at
    FROM quiz_attempts
    WHERE quiz_id = ? AND deleted_at IS NULL
    ORDER BY patient_id, completed_at DESC, created_at DESC
), paired_patients AS (
    SELECT patient_id FROM first_pre
    UNION
    SELECT patient_id FROM latest_post
)
SELECT p.id AS patient_id, p.full_name AS patient_name, p.health_facility AS puskesmas,
       pre.id AS pre_attempt_id, pre.score AS pre_score, pre.passed AS pre_passed,
       pre.score_dsmes AS pre_score_dsm_es, pre.score_dsmqr AS pre_score_dsm_qr, pre.dsmqr_category AS pre_dsm_qr_category,
       pre.completed_at AS pre_completed_at,
       post.id AS post_attempt_id, post.score AS post_score, post.passed AS post_passed,
       post.score_dsmes AS post_score_dsm_es, post.score_dsmqr AS post_score_dsm_qr, post.dsmqr_category AS post_dsm_qr_category,
       post.completed_at AS post_completed_at
FROM paired_patients pp
JOIN patients p ON p.id = pp.patient_id AND p.deleted_at IS NULL
LEFT JOIN first_pre pre ON pre.patient_id = pp.patient_id
LEFT JOIN latest_post post ON post.patient_id = pp.patient_id
WHERE (? = '' OR p.health_facility = ?)
ORDER BY p.full_name`, preTestID, postTestID, facilityName, facilityName)
	if err := query.Scan(&rows).Error; err != nil {
		return nil, errs.NewInternal("failed to fetch questionnaire comparison", err)
	}
	return rows, nil
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

func (r *quizRepository) FindMyAttemptByID(ctx context.Context, patientID, questionnaireID, attemptID string) (*domain.QuizAttempt, error) {
	var attempt domain.QuizAttempt
	err := r.db.WithContext(ctx).
		Preload("Questionnaire").
		Preload("Questionnaire.Education").
		Preload("Answers.Question.Options").
		Where("id = ? AND quiz_id = ? AND patient_id = ? AND deleted_at IS NULL", attemptID, questionnaireID, patientID).
		First(&attempt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("attempt not found")
		}
		return nil, errs.NewInternal("failed to fetch attempt", err)
	}
	return &attempt, nil
}

func (r *quizRepository) FindMyHistory(ctx context.Context, patientID, qType, scope string) ([]domain.QuizAttempt, error) {
	var attempts []domain.QuizAttempt
	q := r.db.WithContext(ctx).
		Preload("Questionnaire").
		Preload("Questionnaire.Education").
		Joins("JOIN questionnaires ON questionnaires.id = quiz_attempts.quiz_id AND questionnaires.deleted_at IS NULL").
		Where("quiz_attempts.patient_id = ? AND quiz_attempts.deleted_at IS NULL", patientID)

	if qType != "" {
		q = q.Where("questionnaires.type = ?", qType)
	}
	if scope != "" {
		q = q.Where("questionnaires.scope = ?", scope)
	}

	if err := q.Order("quiz_attempts.completed_at DESC").Limit(200).Find(&attempts).Error; err != nil {
		return nil, errs.NewInternal("failed to fetch patient history", err)
	}
	return attempts, nil
}
