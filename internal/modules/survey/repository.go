package survey

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"gorm.io/gorm"
)

type surveyRepository struct {
	db *gorm.DB
}

func NewSurveyRepository(db *gorm.DB) SurveyRepository {
	return &surveyRepository{db: db}
}

func (r *surveyRepository) Create(ctx context.Context, survey *domain.Survey) error {
	if err := r.db.WithContext(ctx).Create(survey).Error; err != nil {
		return errs.NewInternal("failed to create survey", err)
	}
	return nil
}

func (r *surveyRepository) Update(ctx context.Context, survey *domain.Survey) error {
	if err := r.db.WithContext(ctx).Save(survey).Error; err != nil {
		return errs.NewInternal("failed to update survey", err)
	}
	return nil
}

func (r *surveyRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Model(&domain.Survey{}).Where("id = ?", id).Update("deleted_at", now).Error; err != nil {
			return errs.NewInternal("failed to delete survey", err)
		}
		if err := tx.Model(&domain.SurveyQuestion{}).Where("survey_id = ?", id).Update("deleted_at", now).Error; err != nil {
			return errs.NewInternal("failed to delete survey questions", err)
		}
		return nil
	})
}

func (r *surveyRepository) GetByID(ctx context.Context, id string) (*domain.Survey, error) {
	var s domain.Survey
	err := r.db.WithContext(ctx).
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Where("deleted_at IS NULL").Order("display_order ASC")
		}).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&s).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("survey not found")
		}
		return nil, errs.NewInternal("failed to fetch survey", err)
	}
	return &s, nil
}

func (r *surveyRepository) List(ctx context.Context, surveyType string, status string, page int, limit int) ([]domain.Survey, int64, error) {
	var items []domain.Survey
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.Survey{}).Where("deleted_at IS NULL")

	if surveyType != "" {
		q = q.Where("type = ?", surveyType)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errs.NewInternal("failed to count surveys", err)
	}

	offset := (page - 1) * limit
	err := q.Preload("Questions", "deleted_at IS NULL").
		Preload("Responses", "deleted_at IS NULL").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&items).Error

	if err != nil {
		return nil, 0, errs.NewInternal("failed to list surveys", err)
	}

	return items, total, nil
}

func (r *surveyRepository) GetActiveSurvey(ctx context.Context, surveyType string) (*domain.Survey, error) {
	var s domain.Survey
	q := r.db.WithContext(ctx).
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Where("deleted_at IS NULL").Order("display_order ASC")
		}).
		Where("status = ? AND deleted_at IS NULL", domain.SurveyStatusPublished)

	if surveyType != "" {
		q = q.Where("type = ?", surveyType)
	}

	// Prefer is_active = true first, or latest published
	err := q.Order("is_active DESC, updated_at DESC").First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NewNotFound("no active survey found")
		}
		return nil, errs.NewInternal("failed to fetch active survey", err)
	}
	return &s, nil
}

func (r *surveyRepository) ListActiveSurveys(ctx context.Context, surveyType string) ([]domain.Survey, error) {
	var surveys []domain.Survey
	q := r.db.WithContext(ctx).Model(&domain.Survey{}).
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Where("status = ? AND is_active = ? AND deleted_at IS NULL", domain.SurveyStatusPublished, true)

	if surveyType != "" {
		q = q.Where("type = ?", surveyType)
	}

	err := q.Order("updated_at DESC").Find(&surveys).Error
	if err != nil {
		return nil, errs.NewInternal("failed to fetch active surveys", err)
	}
	return surveys, nil
}

func (r *surveyRepository) SetActive(ctx context.Context, id string, surveyType string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Deactivate other surveys of the same type
		if err := tx.Model(&domain.Survey{}).
			Where("type = ? AND id != ? AND deleted_at IS NULL", surveyType, id).
			Update("is_active", false).Error; err != nil {
			return errs.NewInternal("failed to deactivate other surveys", err)
		}
		// Activate target survey
		if err := tx.Model(&domain.Survey{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"is_active": true,
				"status":    domain.SurveyStatusPublished,
			}).Error; err != nil {
			return errs.NewInternal("failed to activate survey", err)
		}
		return nil
	})
}

func (r *surveyRepository) ReplaceQuestions(ctx context.Context, surveyID string, questions []domain.SurveyQuestion) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Hard delete old questions to replace with fresh ordered list
		if err := tx.Where("survey_id = ?", surveyID).Delete(&domain.SurveyQuestion{}).Error; err != nil {
			return errs.NewInternal("failed to remove old questions", err)
		}
		if len(questions) > 0 {
			for i := range questions {
				questions[i].SurveyID = surveyID
			}
			if err := tx.Create(&questions).Error; err != nil {
				return errs.NewInternal("failed to insert new questions", err)
			}
		}
		return nil
	})
}

func (r *surveyRepository) CreateResponse(ctx context.Context, resp *domain.SurveyResponse, answers []domain.SurveyAnswer) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(resp).Error; err != nil {
			return errs.NewInternal("failed to save survey response", err)
		}
		if len(answers) > 0 {
			for i := range answers {
				answers[i].ResponseID = resp.ID
			}
			if err := tx.Create(&answers).Error; err != nil {
				return errs.NewInternal("failed to save survey answers", err)
			}
		}
		return nil
	})
}

func (r *surveyRepository) GetResponseBySurveyAndPatient(ctx context.Context, surveyID string, patientID string) (*domain.SurveyResponse, error) {
	var resp domain.SurveyResponse
	err := r.db.WithContext(ctx).
		Preload("Answers").
		Where("survey_id = ? AND patient_id = ? AND deleted_at IS NULL", surveyID, patientID).
		First(&resp).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not submitted yet
		}
		return nil, errs.NewInternal("failed to check existing response", err)
	}
	return &resp, nil
}

func (r *surveyRepository) ListResponses(ctx context.Context, surveyID string, page int, limit int) ([]domain.SurveyResponse, int64, error) {
	var items []domain.SurveyResponse
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.SurveyResponse{}).
		Where("survey_id = ? AND deleted_at IS NULL", surveyID)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, errs.NewInternal("failed to count responses", err)
	}

	offset := (page - 1) * limit
	err := q.Preload("Patient").
		Preload("Answers.Question").
		Order("completed_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&items).Error

	if err != nil {
		return nil, 0, errs.NewInternal("failed to list responses", err)
	}
	return items, total, nil
}

func (r *surveyRepository) GetAllResponsesForExport(ctx context.Context, surveyID string) ([]domain.SurveyResponse, error) {
	var items []domain.SurveyResponse
	err := r.db.WithContext(ctx).
		Preload("Patient").
		Preload("Answers.Question").
		Where("survey_id = ? AND deleted_at IS NULL", surveyID).
		Order("completed_at ASC").
		Find(&items).Error

	if err != nil {
		return nil, errs.NewInternal("failed to fetch responses for export", err)
	}
	return items, nil
}

func (r *surveyRepository) GetAnalytics(ctx context.Context, surveyID string) (*SurveyAnalyticsResponse, error) {
	survey, err := r.GetByID(ctx, surveyID)
	if err != nil {
		return nil, err
	}

	var responses []domain.SurveyResponse
	if err := r.db.WithContext(ctx).
		Preload("Answers").
		Where("survey_id = ? AND deleted_at IS NULL", surveyID).
		Find(&responses).Error; err != nil {
		return nil, errs.NewInternal("failed to fetch survey responses", err)
	}

	totalParticipants := len(responses)
	res := &SurveyAnalyticsResponse{
		SurveyID:            survey.ID,
		SurveyTitle:         survey.Title,
		Type:                survey.Type,
		TotalParticipants:   totalParticipants,
		CompletedCount:      totalParticipants,
		CompletionRate:      100.0,
		QuestionStatistics:  make([]QuestionAnalytic, 0, len(survey.Questions)),
	}

	if totalParticipants == 0 {
		// Populate empty stats for questions
		for _, q := range survey.Questions {
			res.QuestionStatistics = append(res.QuestionStatistics, QuestionAnalytic{
				QuestionID:    q.ID,
				QuestionText:  q.QuestionText,
				DisplayOrder:  q.DisplayOrder,
				AverageRating: 0.0,
				RatingCounts:  map[string]int{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0},
			})
		}
		return res, nil
	}

	var totalDurationSecs int
	var sumAvgScore, sumPctScore float64
	var sumSUS float64
	var highestSUS, lowestSUS float64
	var passCount, failCount int
	interpretationCounts := make(map[string]int)

	if survey.Type == domain.SurveyTypeSUS {
		highestSUS = -1.0
		lowestSUS = 101.0
	}

	for _, resp := range responses {
		totalDurationSecs += resp.DurationSeconds

		if resp.AverageScore != nil {
			sumAvgScore += *resp.AverageScore
		}
		if resp.PercentageScore != nil {
			sumPctScore += *resp.PercentageScore
		}
		if resp.SUSScore != nil {
			susVal := *resp.SUSScore
			sumSUS += susVal
			if susVal > highestSUS {
				highestSUS = susVal
			}
			if susVal < lowestSUS {
				lowestSUS = susVal
			}
		}
		if resp.Passed != nil {
			if *resp.Passed {
				passCount++
			} else {
				failCount++
			}
		}
		if resp.Interpretation != nil && *resp.Interpretation != "" {
			interpretationCounts[*resp.Interpretation]++
		}
	}

	avgDur := totalDurationSecs / totalParticipants
	res.AverageDurationSecs = avgDur

	if survey.Type == domain.SurveyTypeUserSatisfaction {
		avgScoreVal := sumAvgScore / float64(totalParticipants)
		avgPctVal := sumPctScore / float64(totalParticipants)
		res.AverageScore = &avgScoreVal
		res.AveragePercentage = &avgPctVal
	} else if survey.Type == domain.SurveyTypeSUS {
		avgSUSVal := sumSUS / float64(totalParticipants)
		res.AverageSUSScore = &avgSUSVal
		if highestSUS >= 0 {
			res.HighestSUSScore = &highestSUS
		}
		if lowestSUS <= 100 {
			res.LowestSUSScore = &lowestSUS
		}
		res.PassCount = &passCount
		res.FailCount = &failCount
		passRateVal := float64(passCount) / float64(totalParticipants) * 100.0
		res.PassRate = &passRateVal

		// Interpretations distribution
		interpLabels := []string{"Excellent", "Good", "OK", "Poor", "Awful"}
		res.Interpretations = make([]DistributionItem, 0, len(interpLabels))
		for _, lbl := range interpLabels {
			cnt := interpretationCounts[lbl]
			pct := float64(cnt) / float64(totalParticipants) * 100.0
			res.Interpretations = append(res.Interpretations, DistributionItem{
				Label:      lbl,
				Count:      cnt,
				Percentage: pct,
			})
		}
	}

	// Question-by-question analytics
	for _, q := range survey.Questions {
		counts := map[string]int{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0}
		var sumRating int
		var ratingCount int

		for _, resp := range responses {
			for _, ans := range resp.Answers {
				if ans.QuestionID == q.ID {
					valStr := strconv.Itoa(ans.RatingValue)
					counts[valStr]++
					sumRating += ans.RatingValue
					ratingCount++
				}
			}
		}

		avgRating := 0.0
		if ratingCount > 0 {
			avgRating = float64(sumRating) / float64(ratingCount)
		}

		res.QuestionStatistics = append(res.QuestionStatistics, QuestionAnalytic{
			QuestionID:    q.ID,
			QuestionText:  q.QuestionText,
			DisplayOrder:  q.DisplayOrder,
			AverageRating: avgRating,
			RatingCounts:  counts,
		})
	}

	sort.Slice(res.QuestionStatistics, func(i, j int) bool {
		return res.QuestionStatistics[i].DisplayOrder < res.QuestionStatistics[j].DisplayOrder
	})

	return res, nil
}
