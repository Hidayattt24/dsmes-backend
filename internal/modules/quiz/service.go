package quiz

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/modules/facility"
	"github.com/dsmes/dsmes-backend/internal/modules/staff"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type quizService struct {
	repo         QuizRepository
	staffRepo    staff.StaffRepository
	facilityRepo facility.FacilityRepository
	log          *zap.Logger
}

func NewQuizService(repo QuizRepository, staffRepo staff.StaffRepository, facilityRepo facility.FacilityRepository, log *zap.Logger) QuizService {
	return &quizService{repo: repo, staffRepo: staffRepo, facilityRepo: facilityRepo, log: log}
}

// resolveFacilityName maps a staff account to its assigned Puskesmas name.
func (s *quizService) resolveFacilityName(ctx context.Context, staffID string) (string, error) {
	if staffID == "" {
		return "", errs.NewUnauthorized("staff not identified")
	}
	sa, err := s.staffRepo.FindByID(ctx, staffID)
	if err != nil {
		return "", err
	}
	if sa.HealthFacilityID == nil || *sa.HealthFacilityID == "" {
		return "", errs.NewForbidden("staff belum memiliki puskesmas")
	}
	f, err := s.facilityRepo.FindByID(ctx, *sa.HealthFacilityID)
	if err != nil {
		return "", err
	}
	return f.Name, nil
}

func (s *quizService) ListQuestionnaires(ctx context.Context, search, qType, status, scope, sortBy, sortOrder string, page, limit int) ([]QuestionnaireResponse, int64, error) {
	return s.listQuestionnaires(ctx, search, qType, status, scope, sortBy, sortOrder, page, limit, "")
}

func (s *quizService) ListQuestionnairesForStaff(ctx context.Context, staffID string, search, qType, status, scope, sortBy, sortOrder string, page, limit int) ([]QuestionnaireResponse, int64, error) {
	name, err := s.resolveFacilityName(ctx, staffID)
	if err != nil {
		return nil, 0, err
	}
	return s.listQuestionnaires(ctx, search, qType, status, scope, sortBy, sortOrder, page, limit, name)
}

func (s *quizService) listQuestionnaires(ctx context.Context, search, qType, status, scope, sortBy, sortOrder string, page, limit int, facilityName string) ([]QuestionnaireResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	items, total, err := s.repo.FindAll(ctx, search, qType, status, scope, sortBy, sortOrder, page, limit)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]QuestionnaireResponse, len(items))
	for i, q := range items {
		catCount := len(q.Categories)
		questCount := 0
		for _, cat := range q.Categories {
			questCount += len(cat.Questions)
		}
		pCount, _ := s.repo.CountAttempts(ctx, q.ID, facilityName)
		avgScore, _ := s.repo.GetAverageScore(ctx, q.ID, facilityName)

		resp[i] = ToQuestionnaireResponse(&q, catCount, questCount, pCount, avgScore)
		if err := s.enrichPairResponse(ctx, &resp[i], &q); err != nil {
			return nil, 0, err
		}
	}

	return resp, total, nil
}

func (s *quizService) GetQuestionnaire(ctx context.Context, id string, isAdminOrStaff bool) (*QuestionnaireDetailResponse, error) {
	return s.getQuestionnaire(ctx, id, isAdminOrStaff, "")
}

func (s *quizService) GetQuestionnaireForStaff(ctx context.Context, staffID string, id string) (*QuestionnaireDetailResponse, error) {
	name, err := s.resolveFacilityName(ctx, staffID)
	if err != nil {
		return nil, err
	}
	return s.getQuestionnaire(ctx, id, true, name)
}

func (s *quizService) getQuestionnaire(ctx context.Context, id string, isAdminOrStaff bool, facilityName string) (*QuestionnaireDetailResponse, error) {
	q, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	pCount, _ := s.repo.CountAttempts(ctx, q.ID, facilityName)
	avgScore, _ := s.repo.GetAverageScore(ctx, q.ID, facilityName)
	resp := ToQuestionnaireDetailResponse(q, pCount, avgScore, isAdminOrStaff)
	if err := s.enrichPairResponse(ctx, &resp.QuestionnaireResponse, q); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *quizService) enrichPairResponse(ctx context.Context, response *QuestionnaireResponse, q *domain.Questionnaire) error {
	paired, err := s.repo.FindPairedQuestionnaire(ctx, q)
	if err != nil || paired == nil {
		return err
	}
	response.PairedQuestionnaireID = &paired.ID
	response.PairedQuestionnaireTitle = &paired.Title
	return nil
}

func (s *quizService) GetActivePreTest(ctx context.Context, isAdminOrStaff bool) (*QuestionnaireDetailResponse, error) {
	q, err := s.repo.GetActivePreTest(ctx)
	if err != nil {
		return nil, err
	}
	pCount, _ := s.repo.CountAttempts(ctx, q.ID, "")
	avgScore, _ := s.repo.GetAverageScore(ctx, q.ID, "")
	resp := ToQuestionnaireDetailResponse(q, pCount, avgScore, isAdminOrStaff)
	return &resp, nil
}

func (s *quizService) GetPreTestByEducation(ctx context.Context, educationID string, isAdminOrStaff bool) (*QuestionnaireDetailResponse, error) {
	if strings.TrimSpace(educationID) == "" {
		return nil, errs.NewBadRequest("education_id parameter is required")
	}
	q, err := s.repo.GetPreTestByEducation(ctx, educationID)
	if err != nil {
		return nil, err
	}
	pCount, _ := s.repo.CountAttempts(ctx, q.ID, "")
	avgScore, _ := s.repo.GetAverageScore(ctx, q.ID, "")
	resp := ToQuestionnaireDetailResponse(q, pCount, avgScore, isAdminOrStaff)
	return &resp, nil
}

func (s *quizService) GetPostTestByEducation(ctx context.Context, educationID string, isAdminOrStaff bool) (*QuestionnaireDetailResponse, error) {
	if strings.TrimSpace(educationID) == "" {
		return nil, errs.NewBadRequest("education_id parameter is required")
	}
	q, err := s.repo.GetPostTestByEducation(ctx, educationID)
	if err != nil {
		return nil, err
	}
	pCount, _ := s.repo.CountAttempts(ctx, q.ID, "")
	avgScore, _ := s.repo.GetAverageScore(ctx, q.ID, "")
	resp := ToQuestionnaireDetailResponse(q, pCount, avgScore, isAdminOrStaff)
	return &resp, nil
}

func (s *quizService) validateQuestionnairePayload(ctx context.Context, questionnaireID string, req CreateQuestionnaireRequest) (domain.QuestionnaireType, domain.QuestionnaireScope, string, error) {
	qType := normalizeType(req.Type)
	status := normalizeStatus(req.Status)

	scope := domain.ScopeEducation
	if req.Scope != nil && strings.ToUpper(strings.TrimSpace(*req.Scope)) == string(domain.ScopeApplication) {
		scope = domain.ScopeApplication
	} else if req.EducationID == nil || strings.TrimSpace(*req.EducationID) == "" {
		scope = domain.ScopeApplication
	}

	if scope == domain.ScopeApplication {
		// Kuesioner Aplikasi (Evaluasi Intervensi Platform: DSMQR & DSMES)
		// No education_id, no passing score, no difficulty required.
		if len(req.Categories) == 0 && len(req.Questions) == 0 && !(qType == domain.TypePostTest && req.CopyQuestionsFromPreTest) {
			return qType, scope, status, errs.NewBadRequest("Kuesioner wajib memiliki minimal 1 Pertanyaan")
		}
	} else {
		// Kuesioner Modul Edukasi (Tes Pengetahuan Artikel)
		if req.EducationID == nil || strings.TrimSpace(*req.EducationID) == "" {
			return qType, scope, status, errs.NewBadRequest("Kuesioner Edukasi wajib terhubung dengan 1 Materi Edukasi")
		}

		if qType == domain.TypePostTest {
			if req.PassingScore == nil {
				return qType, scope, status, errs.NewBadRequest("Post-Test wajib memiliki Nilai Kelulusan (Passing Score)")
			}
			if req.Difficulty == nil || strings.TrimSpace(*req.Difficulty) == "" {
				return qType, scope, status, errs.NewBadRequest("Post-Test wajib memiliki Tingkat Kesulitan")
			}
		}

		if len(req.Categories) == 0 && len(req.Questions) == 0 && !(qType == domain.TypePostTest && req.CopyQuestionsFromPreTest) {
			return qType, scope, status, errs.NewBadRequest("Kuisioner wajib memiliki minimal 1 Kategori atau Pertanyaan")
		}
	}

	return qType, scope, status, nil
}

func sameStringPointers(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func cloneCategories(categories []domain.QuestionCategory) []domain.QuestionCategory {
	cloned := make([]domain.QuestionCategory, len(categories))
	for i, category := range categories {
		cloned[i] = domain.QuestionCategory{
			Title: category.Title, Description: category.Description, DisplayOrder: category.DisplayOrder,
			InstrumentCategory: category.InstrumentCategory,
			Questions:          make([]domain.Question, len(category.Questions)),
		}
		for j, question := range category.Questions {
			cloned[i].Questions[j] = domain.Question{
				QuestionText: question.QuestionText, QuestionImageURL: question.QuestionImageURL,
				Explanation: question.Explanation, DisplayOrder: question.DisplayOrder,
				Options: make([]domain.QuestionOption, len(question.Options)),
			}
			for k, option := range question.Options {
				cloned[i].Questions[j].Options[k] = domain.QuestionOption{
					OptionText: option.OptionText, ScoreValue: option.ScoreValue, IsCorrect: option.IsCorrect, DisplayOrder: option.DisplayOrder,
				}
			}
		}
	}
	return cloned
}

func (s *quizService) validatePostTestPair(ctx context.Context, questionnaireID string, scope domain.QuestionnaireScope, educationID, sourceID *string) (*domain.Questionnaire, error) {
	if sourceID == nil || strings.TrimSpace(*sourceID) == "" {
		return nil, errs.NewBadRequest("Post-Test wajib memiliki source_pre_test_id")
	}
	source, err := s.repo.FindByID(ctx, *sourceID)
	if err != nil {
		return nil, err
	}
	if source.Type != domain.TypePreTest {
		return nil, errs.NewBadRequest("source_pre_test_id harus merujuk ke Pre-Test")
	}
	if source.Scope != scope {
		return nil, errs.NewBadRequest("Post-Test dan Pre-Test harus memiliki scope yang sama")
	}
	if scope == domain.ScopeEducation && !sameStringPointers(source.EducationID, educationID) {
		return nil, errs.NewBadRequest("Post-Test dan Pre-Test harus memiliki education_id yang sama")
	}
	existing, err := s.repo.FindPostTestBySource(ctx, source.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.ID != questionnaireID {
		return nil, errs.NewBadRequest("Pre-Test sudah memiliki Post-Test")
	}
	return source, nil
}

func buildCategoriesFromRequest(reqTitle string, qType domain.QuestionnaireType, reqCategories []QuestionCategoryRequest, reqQuestions []QuestionRequest) []domain.QuestionCategory {
	// If PRE_TEST questions are sent directly at root level without categories:
	if qType == domain.TypePreTest && len(reqQuestions) > 0 {
		questions := make([]domain.Question, len(reqQuestions))
		for j, qReq := range reqQuestions {
			var imgURL *string
			if strings.TrimSpace(qReq.QuestionImageURL) != "" {
				urlVal := qReq.QuestionImageURL
				imgURL = &urlVal
			}

			// Pre-test questions do not store answer choices in DB
			questions[j] = domain.Question{
				QuestionText:     qReq.QuestionText,
				QuestionImageURL: imgURL,
				Explanation:      qReq.Explanation,
				DisplayOrder:     j,
				Options:          nil,
			}
		}

		catTitle := "Pertanyaan DMSES"
		if strings.TrimSpace(reqTitle) != "" {
			catTitle = reqTitle
		}

		return []domain.QuestionCategory{
			{
				Title:        catTitle,
				Description:  "Kuesioner Efikasi Diri Manajemen Diabetes",
				DisplayOrder: 0,
				Questions:    questions,
			},
		}
	}

	categories := make([]domain.QuestionCategory, len(reqCategories))
	for i, cReq := range reqCategories {
		title := cReq.Title
		if strings.TrimSpace(title) == "" {
			title = reqTitle
			if strings.TrimSpace(title) == "" {
				title = "Soal Kuesioner"
			}
		}

		instCategory := strings.ToUpper(strings.TrimSpace(cReq.InstrumentCategory))
		if instCategory == "" {
			instCategory = "PRE_TEST"
		}

		questions := make([]domain.Question, len(cReq.Questions))
		for j, qReq := range cReq.Questions {
			var imgURL *string
			if strings.TrimSpace(qReq.QuestionImageURL) != "" {
				urlVal := qReq.QuestionImageURL
				imgURL = &urlVal
			}

			var options []domain.QuestionOption
			if len(qReq.Choices) > 0 {
				options = make([]domain.QuestionOption, len(qReq.Choices))
				for k, optReq := range qReq.Choices {
					options[k] = domain.QuestionOption{
						OptionText:   optReq.OptionText,
						ScoreValue:   optReq.ScoreValue,
						IsCorrect:    optReq.IsCorrect,
						DisplayOrder: k,
					}
				}
			}

			questions[j] = domain.Question{
				QuestionText:     qReq.QuestionText,
				QuestionImageURL: imgURL,
				Explanation:      qReq.Explanation,
				DisplayOrder:     j,
				Options:          options,
			}
		}
		categories[i] = domain.QuestionCategory{
			Title:              title,
			Description:        cReq.Description,
			DisplayOrder:       i,
			InstrumentCategory: instCategory,
			Questions:          questions,
		}
	}
	return categories
}

func (s *quizService) CreateQuestionnaire(ctx context.Context, staffID string, req CreateQuestionnaireRequest) (*QuestionnaireDetailResponse, error) {
	qType, scope, status, err := s.validateQuestionnairePayload(ctx, "", req)
	if err != nil {
		return nil, err
	}

	var eduID *string
	var passingScore *int
	var difficulty *string

	if scope == domain.ScopeEducation {
		eduID = req.EducationID
		if qType == domain.TypePostTest {
			passingScore = req.PassingScore
			difficulty = req.Difficulty
		}
	}

	var source *domain.Questionnaire
	if qType == domain.TypePostTest {
		source, err = s.validatePostTestPair(ctx, "", scope, eduID, req.SourcePreTestID)
		if err != nil {
			return nil, err
		}
	}
	categories := buildCategoriesFromRequest(req.Title, qType, req.Categories, req.Questions)
	if source != nil && req.CopyQuestionsFromPreTest {
		categories = cloneCategories(source.Categories)
	}

	createdBy := staffID
	q := &domain.Questionnaire{
		Title:             req.Title,
		Type:              qType,
		Scope:             scope,
		Description:       req.Description,
		EducationID:       eduID,
		PassingScore:      passingScore,
		Difficulty:        difficulty,
		Status:            status,
		CreatedBy:         &createdBy,
		Categories:        categories,
		SourcePreTestID:   req.SourcePreTestID,
		IsLinkedToPreTest: qType == domain.TypePostTest,
	}

	if err := s.repo.Create(ctx, q); err != nil {
		return nil, err
	}

	return s.GetQuestionnaire(ctx, q.ID, true)
}

func (s *quizService) UpdateQuestionnaire(ctx context.Context, id string, req CreateQuestionnaireRequest) (*QuestionnaireDetailResponse, error) {
	q, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	qType, scope, status, err := s.validateQuestionnairePayload(ctx, id, req)
	if err != nil {
		return nil, err
	}

	var eduID *string
	var passingScore *int
	var difficulty *string

	if scope == domain.ScopeEducation {
		eduID = req.EducationID
		if qType == domain.TypePostTest && !q.IsLinkedToPreTest {
			passingScore = req.PassingScore
			difficulty = req.Difficulty
		}
	}

	var source *domain.Questionnaire
	if qType == domain.TypePostTest {
		source, err = s.validatePostTestPair(ctx, id, scope, eduID, req.SourcePreTestID)
		if err != nil {
			return nil, err
		}
	}
	categories := buildCategoriesFromRequest(req.Title, qType, req.Categories, req.Questions)
	if source != nil && req.CopyQuestionsFromPreTest {
		categories = cloneCategories(source.Categories)
	}

	q.Title = req.Title
	q.Type = qType
	q.Scope = scope
	q.Description = req.Description
	q.EducationID = eduID
	q.PassingScore = passingScore
	q.Difficulty = difficulty
	q.Status = status
	q.Categories = categories
	if qType == domain.TypePostTest {
		q.SourcePreTestID = req.SourcePreTestID
		q.IsLinkedToPreTest = true
	} else {
		q.SourcePreTestID = nil
		q.IsLinkedToPreTest = false
	}

	if err := s.repo.Update(ctx, q); err != nil {
		return nil, err
	}

	return s.GetQuestionnaire(ctx, id, true)
}

func (s *quizService) UpdateQuestionnaireStatus(ctx context.Context, id, status string) error {
	normalized := normalizeStatus(status)
	if normalized != "aktif" && normalized != "draft" && normalized != "nonaktif" {
		return errs.NewBadRequest("status must be aktif, draft, or nonaktif")
	}
	return s.repo.UpdateStatus(ctx, id, normalized)
}

func (s *quizService) PairQuestionnaire(ctx context.Context, postTestID, sourcePreTestID string) error {
	post, err := s.repo.FindByID(ctx, postTestID)
	if err != nil {
		return err
	}
	if post.Type != domain.TypePostTest {
		return errs.NewBadRequest("hanya Post-Test yang dapat dipasangkan")
	}

	_, err = s.validatePostTestPair(ctx, postTestID, post.Scope, post.EducationID, &sourcePreTestID)
	if err != nil {
		return err
	}

	post.SourcePreTestID = &sourcePreTestID
	post.IsLinkedToPreTest = true
	return s.repo.Update(ctx, post)
}

func (s *quizService) GetEligiblePreTests(ctx context.Context, scopeValue string, educationID *string) ([]QuestionnaireResponse, error) {
	scope := domain.QuestionnaireScope(strings.ToUpper(strings.TrimSpace(scopeValue)))
	if scope != domain.ScopeApplication && scope != domain.ScopeEducation {
		return nil, errs.NewBadRequest("scope must be APPLICATION or EDUCATION")
	}
	if scope == domain.ScopeEducation && (educationID == nil || strings.TrimSpace(*educationID) == "") {
		return nil, errs.NewBadRequest("education_id is required for EDUCATION scope")
	}
	items, err := s.repo.FindEligiblePreTests(ctx, scope, educationID)
	if err != nil {
		return nil, err
	}
	result := make([]QuestionnaireResponse, len(items))
	for i := range items {
		questionCount := 0
		for _, category := range items[i].Categories {
			questionCount += len(category.Questions)
		}
		result[i] = ToQuestionnaireResponse(&items[i], len(items[i].Categories), questionCount, 0, nil)
	}
	return result, nil
}

func (s *quizService) resolvePair(ctx context.Context, id string) (*domain.Questionnaire, *domain.Questionnaire, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if item.Type == domain.TypePostTest {
		if item.SourcePreTestID == nil {
			return nil, nil, errs.NewNotFound("questionnaire pair not found")
		}
		pre, err := s.repo.FindByID(ctx, *item.SourcePreTestID)
		return pre, item, err
	}
	post, err := s.repo.FindPostTestBySource(ctx, item.ID)
	return item, post, err
}

func (s *quizService) GetQuestionnairePair(ctx context.Context, id string, staffID string) (*QuestionnairePairResponse, error) {
	pre, post, err := s.resolvePair(ctx, id)
	if err != nil {
		return nil, err
	}
	facilityName := ""
	if staffID != "" {
		facilityName, err = s.resolveFacilityName(ctx, staffID)
		if err != nil {
			return nil, err
		}
	}
	preDetail, err := s.getQuestionnaire(ctx, pre.ID, true, facilityName)
	if err != nil {
		return nil, err
	}
	result := &QuestionnairePairResponse{PreTest: preDetail}
	if post != nil {
		result.PostTest, err = s.getQuestionnaire(ctx, post.ID, true, facilityName)
	}
	return result, err
}

func (s *quizService) GetPairComparison(ctx context.Context, id string, staffID string) ([]PairComparisonResponse, error) {
	pre, post, err := s.resolvePair(ctx, id)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, errs.NewNotFound("linked Post-Test not found")
	}
	facilityName := ""
	if staffID != "" {
		facilityName, err = s.resolveFacilityName(ctx, staffID)
		if err != nil {
			return nil, err
		}
	}
	rows, err := s.repo.FindPairComparisons(ctx, pre.ID, post.ID, facilityName)
	if err != nil {
		return nil, err
	}
	result := make([]PairComparisonResponse, len(rows))
	for i, row := range rows {
		result[i] = PairComparisonResponse{PatientID: row.PatientID, PatientName: row.PatientName, Puskesmas: row.Puskesmas}
		if row.PreAttemptID != nil {
			result[i].PreAttempt = &PairAttemptResponse{AttemptID: *row.PreAttemptID, Score: *row.PreScore, Passed: *row.PrePassed, ScoreDSMES: row.PreScoreDSMES, ScoreDSMQR: row.PreScoreDSMQR, DSMQRCategory: row.PreDSMQRCategory, CompletedAt: *row.PreCompletedAt}
		}
		switch {
		case row.PreAttemptID != nil && row.PostAttemptID != nil:
			result[i].Status = "Pasangan lengkap"
		case row.PreAttemptID != nil:
			result[i].Status = "Pre-Test selesai"
		case row.PostAttemptID != nil:
			result[i].Status = "Post-Test selesai"
		default:
			result[i].Status = "Belum dikerjakan"
		}
		if row.PostAttemptID != nil {
			result[i].PostAttempt = &PairAttemptResponse{AttemptID: *row.PostAttemptID, Score: *row.PostScore, Passed: *row.PostPassed, ScoreDSMES: row.PostScoreDSMES, ScoreDSMQR: row.PostScoreDSMQR, DSMQRCategory: row.PostDSMQRCategory, CompletedAt: *row.PostCompletedAt}
		}
		if row.PreScore != nil && row.PostScore != nil {
			change := *row.PostScore - *row.PreScore
			result[i].ScoreChange = &change
		}
		if row.PreScoreDSMES != nil && row.PostScoreDSMES != nil {
			change := *row.PostScoreDSMES - *row.PreScoreDSMES
			result[i].DSMESChange = &change
		}
		if row.PreScoreDSMQR != nil && row.PostScoreDSMQR != nil {
			change := *row.PostScoreDSMQR - *row.PreScoreDSMQR
			result[i].DSMQRChange = &change
		}
	}
	return result, nil
}

func (s *quizService) GetParticipantPairComparison(ctx context.Context, id, participantID, staffID string) (*PairComparisonResponse, error) {
	items, err := s.GetPairComparison(ctx, id, staffID)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].PatientID == participantID {
			return &items[i], nil
		}
	}
	return nil, errs.NewNotFound("participant comparison not found")
}

func (s *quizService) DeleteQuestionnaire(ctx context.Context, id string) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *quizService) GetStats(ctx context.Context) (*QuizStats, error) {
	return s.repo.GetStats(ctx, "")
}

func (s *quizService) GetStatsForStaff(ctx context.Context, staffID string) (*QuizStats, error) {
	name, err := s.resolveFacilityName(ctx, staffID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetStats(ctx, name)
}

func (s *quizService) SubmitQuestionnaire(ctx context.Context, patientID string, questionnaireID string, req SubmitQuestionnaireRequest) (*SubmitResultResponse, error) {
	q, err := s.repo.FindByID(ctx, questionnaireID)
	if err != nil {
		return nil, err
	}
	if status := strings.ToLower(strings.TrimSpace(q.Status)); status != "aktif" && status != "terbit" {
		return nil, errs.NewBadRequest("Kuisioner belum aktif")
	}
	questionnaireScope := q.Scope
	if questionnaireScope == "" {
		if q.EducationID != nil {
			questionnaireScope = domain.ScopeEducation
		} else if q.Type == domain.TypePreTest {
			questionnaireScope = domain.ScopeApplication
		} else {
			questionnaireScope = domain.ScopeEducation
		}
	}

	totalQuestions := 0
	questionIDs := make(map[string]domain.Question)
	optionIDs := make(map[string]map[string]bool)
	for _, cat := range q.Categories {
		for _, question := range cat.Questions {
			totalQuestions++
			questionIDs[question.ID] = question
			optionIDs[question.ID] = make(map[string]bool)
			for _, option := range question.Options {
				optionIDs[question.ID][option.ID] = true
			}
		}
	}

	// Idempotency check: if client provided client_id, check if already recorded
	if req.ClientID != nil && *req.ClientID != "" {
		existing, err := s.repo.FindAttemptByClientID(ctx, patientID, *req.ClientID)
		if err == nil && existing != nil {
			var selfEffCat string
			if existing.SelfEfficacyCategory != nil {
				selfEffCat = *existing.SelfEfficacyCategory
			}
			return &SubmitResultResponse{
				AttemptID:            existing.ID,
				QuestionnaireID:      existing.QuestionnaireID,
				Type:                 string(q.Type),
				Message:              "Jawaban Anda telah berhasil disimpan.",
				Score:                &existing.Score,
				Passed:               &existing.Passed,
				TotalQuestions:       totalQuestions,
				SelfEfficacyCategory: selfEffCat,
			}, nil
		}
	}

	if totalQuestions == 0 {
		return nil, errs.NewBadRequest("Kuisioner tidak memiliki pertanyaan")
	}
	if len(req.Answers) != totalQuestions {
		return nil, errs.NewBadRequest("Semua pertanyaan wajib dijawab")
	}
	// Map each question to its instrument category
	questionCategoryMap := make(map[string]string)
	for _, cat := range q.Categories {
		instCat := strings.ToUpper(strings.TrimSpace(cat.InstrumentCategory))
		if instCat == "" {
			instCat = "PRE_TEST"
		}
		for _, question := range cat.Questions {
			questionCategoryMap[question.ID] = instCat
		}
	}

	seen := make(map[string]bool, len(req.Answers))
	for _, answer := range req.Answers {
		if _, exists := questionIDs[answer.QuestionID]; !exists {
			return nil, errs.NewBadRequest("Jawaban mengandung pertanyaan yang tidak valid")
		}
		if seen[answer.QuestionID] {
			return nil, errs.NewBadRequest("Jawaban pertanyaan tidak boleh duplikat")
		}
		seen[answer.QuestionID] = true
		if questionnaireScope == domain.ScopeApplication {
			// Application questionnaires accept either a direct scale value or a dynamic option.
			if answer.SelectedValue == nil && (answer.OptionID == nil || *answer.OptionID == "") {
				return nil, errs.NewBadRequest("Jawaban kuesioner aplikasi harus dipilih")
			}
		} else {
			if answer.OptionID == nil || !optionIDs[answer.QuestionID][*answer.OptionID] {
				return nil, errs.NewBadRequest("Pilihan jawaban tidak valid")
			}
		}
	}

	if questionnaireScope == domain.ScopeApplication {
		// ── Application questionnaire scoring (DSMQR & DSMES) ───────────────
		totalDMSESScore := 0
		dmsesCount := 0

		totalDSMQRScore := 0.0
		dsmqrCount := 0

		totalPreScore := 0
		preCount := 0

		answers := make([]domain.QuizAnswer, len(req.Answers))

		for i, ans := range req.Answers {
			val := 1
			if ans.OptionID != nil {
				for _, option := range questionIDs[ans.QuestionID].Options {
					if option.ID == *ans.OptionID {
						if option.ScoreValue != nil {
							val = int(math.Round(*option.ScoreValue))
						} else {
							val = option.DisplayOrder + 1
						}
						break
					}
				}
			} else if ans.SelectedValue != nil && *ans.SelectedValue >= 1 && *ans.SelectedValue <= 5 {
				val = *ans.SelectedValue
			}

			// If OptionID was provided from dynamic custom options
			var optID *string
			if ans.OptionID != nil && *ans.OptionID != "" {
				optID = ans.OptionID
			}

			instCat := questionCategoryMap[ans.QuestionID]
			switch instCat {
			case "DSMQR":
				totalDSMQRScore += float64(val)
				dsmqrCount++
			case "DSMES":
				totalDMSESScore += val
				dmsesCount++
			default:
				// PRE_TEST general knowledge
				totalPreScore += val
				preCount++
			}

			selectedVal := val
			answers[i] = domain.QuizAnswer{
				QuestionID:     ans.QuestionID,
				SelectedOption: fmt.Sprintf("%d", val),
				SelectedValue:  &selectedVal,
				OptionID:       optID,
				IsCorrect:      true,
			}
		}

		selfEfficacyCat := "Belum dapat diinterpretasikan"
		if dmsesCount == 20 && totalDMSESScore >= 20 && totalDMSESScore <= 100 {
			selfEfficacyCat = DetermineSelfEfficacyCategory(totalDMSESScore)
		}
		dsmqrCat := DetermineDSMQRCategory(totalDSMQRScore, dsmqrCount)

		overallScore := totalDMSESScore
		if dmsesCount == 0 && dsmqrCount > 0 {
			overallScore = int(math.Round(totalDSMQRScore))
		} else if dmsesCount == 0 && preCount > 0 {
			overallScore = totalPreScore
		}

		attempt := &domain.QuizAttempt{
			ClientID:             req.ClientID,
			QuestionnaireID:      q.ID,
			PatientID:            patientID,
			Score:                overallScore,
			SelfEfficacyCategory: &selfEfficacyCat,
			ScoreDSMES:           &totalDMSESScore,
			ScoreDSMQR:           &totalDSMQRScore,
			DSMQRCategory:        &dsmqrCat,
			ScorePreTest:         &totalPreScore,
			Passed:               true,
			DurationSeconds:      req.DurationSeconds,
			Answers:              answers,
		}

		if err := s.repo.SaveAttempt(ctx, attempt); err != nil {
			if req.ClientID != nil && *req.ClientID != "" {
				if existing, findErr := s.repo.FindAttemptByClientID(ctx, patientID, *req.ClientID); findErr == nil && existing != nil {
					var selfEffCat string
					if existing.SelfEfficacyCategory != nil {
						selfEffCat = *existing.SelfEfficacyCategory
					}
					return &SubmitResultResponse{
						AttemptID:            existing.ID,
						QuestionnaireID:      existing.QuestionnaireID,
						Type:                 string(q.Type),
						Message:              "Jawaban Anda telah berhasil disimpan dan akan digunakan untuk membantu memahami tingkat keyakinan Anda dalam mengelola diabetes.",
						Score:                &existing.Score,
						Passed:               &existing.Passed,
						TotalQuestions:       totalQuestions,
						SelfEfficacyCategory: selfEffCat,
						ScoreDSMES:           existing.ScoreDSMES,
						ScoreDSMQR:           existing.ScoreDSMQR,
						DSMQRCategory:        existing.DSMQRCategory,
						ScorePreTest:         existing.ScorePreTest,
					}, nil
				}
			}
			return nil, err
		}

		// Return PRE_TEST completion response with breakdown
		return &SubmitResultResponse{
			AttemptID:            attempt.ID,
			QuestionnaireID:      q.ID,
			Type:                 string(q.Type),
			Message:              "Jawaban Anda telah berhasil disimpan dan akan digunakan untuk membantu memahami tingkat keyakinan Anda dalam mengelola diabetes.",
			Score:                &overallScore,
			TotalQuestions:       totalQuestions,
			SelfEfficacyCategory: selfEfficacyCat,
			ScoreDSMES:           &totalDMSESScore,
			ScoreDSMQR:           &totalDSMQRScore,
			DSMQRCategory:        &dsmqrCat,
			ScorePreTest:         &totalPreScore,
		}, nil
	}

	// ── POST_TEST Submission (Knowledge Assessment) ────────────────────────
	correctOptionMap := make(map[string]string) // question_id -> correct option_id
	for _, cat := range q.Categories {
		for _, quest := range cat.Questions {
			for _, opt := range quest.Options {
				if opt.IsCorrect {
					correctOptionMap[quest.ID] = opt.ID
				}
			}
		}
	}

	correctCount := 0
	answers := make([]domain.QuizAnswer, len(req.Answers))

	for i, ans := range req.Answers {
		optID := ""
		if ans.OptionID != nil {
			optID = *ans.OptionID
		}
		correctOptID, exists := correctOptionMap[ans.QuestionID]
		isCorrect := exists && correctOptID == optID
		if isCorrect {
			correctCount++
		}
		optionIDVal := optID
		answers[i] = domain.QuizAnswer{
			QuestionID:     ans.QuestionID,
			SelectedOption: optID,
			OptionID:       &optionIDVal,
			IsCorrect:      isCorrect,
		}
	}

	score := int((float64(correctCount) / float64(totalQuestions)) * 100)
	passed := true
	if q.PassingScore != nil {
		passed = score >= *q.PassingScore
	}

	attempt := &domain.QuizAttempt{
		ClientID:        req.ClientID,
		QuestionnaireID: q.ID,
		PatientID:       patientID,
		Score:           score,
		Passed:          passed,
		DurationSeconds: req.DurationSeconds,
		Answers:         answers,
	}

	if err := s.repo.SaveAttempt(ctx, attempt); err != nil {
		if req.ClientID != nil && *req.ClientID != "" {
			if existing, findErr := s.repo.FindAttemptByClientID(ctx, patientID, *req.ClientID); findErr == nil && existing != nil {
				scoreVal := existing.Score
				passedVal := existing.Passed
				return &SubmitResultResponse{
					AttemptID:       existing.ID,
					QuestionnaireID: q.ID,
					Type:            "POST_TEST",
					Score:           &scoreVal,
					Passed:          &passedVal,
					TotalQuestions:  totalQuestions,
				}, nil
			}
		}
		return nil, err
	}

	scoreVal := score
	passedVal := passed
	correctCountVal := correctCount

	return &SubmitResultResponse{
		AttemptID:       attempt.ID,
		QuestionnaireID: q.ID,
		Type:            "POST_TEST",
		Score:           &scoreVal,
		Passed:          &passedVal,
		TotalQuestions:  totalQuestions,
		CorrectCount:    &correctCountVal,
	}, nil
}

func (s *quizService) ListParticipants(ctx context.Context, questionnaireID string) ([]ParticipantResponse, error) {
	return s.listParticipants(ctx, questionnaireID, "")
}

func (s *quizService) ListParticipantsForStaff(ctx context.Context, staffID string, questionnaireID string) ([]ParticipantResponse, error) {
	name, err := s.resolveFacilityName(ctx, staffID)
	if err != nil {
		return nil, err
	}
	return s.listParticipants(ctx, questionnaireID, name)
}

func (s *quizService) listParticipants(ctx context.Context, questionnaireID string, facilityName string) ([]ParticipantResponse, error) {
	attempts, err := s.repo.FindAttemptsByQuestionnaireID(ctx, questionnaireID, facilityName)
	if err != nil {
		return nil, err
	}

	resps := make([]ParticipantResponse, len(attempts))
	for i, a := range attempts {
		name := "-"
		avatar := ""
		puskesmas := "DSMES Platform"
		if a.Patient != nil {
			name = a.Patient.FullName
			avatar = a.Patient.ProfilePhotoURL
			if a.Patient.HealthFacility != "" {
				puskesmas = a.Patient.HealthFacility
			}
		}
		catStr := ""
		if a.SelfEfficacyCategory != nil {
			catStr = *a.SelfEfficacyCategory
		}
		resps[i] = ParticipantResponse{
			ID:                   a.ID,
			PatientID:            a.PatientID,
			PatientName:          name,
			PatientAvatar:        avatar,
			Puskesmas:            puskesmas,
			CompletionDate:       a.CompletedAt,
			Score:                a.Score,
			Passed:               a.Passed,
			Duration:             formatDuration(a.DurationSeconds),
			SelfEfficacyCategory: catStr,
			ScoreDSMES:           a.ScoreDSMES,
			ScoreDSMQR:           a.ScoreDSMQR,
			DSMQRCategory:        a.DSMQRCategory,
			ScorePreTest:         a.ScorePreTest,
		}
	}
	return resps, nil
}

func (s *quizService) GetParticipantDetail(ctx context.Context, questionnaireID string, participantID string) (*ParticipantDetailResponse, error) {
	return s.getParticipantDetail(ctx, questionnaireID, participantID, "")
}

func (s *quizService) GetParticipantDetailForStaff(ctx context.Context, staffID string, questionnaireID string, participantID string) (*ParticipantDetailResponse, error) {
	name, err := s.resolveFacilityName(ctx, staffID)
	if err != nil {
		return nil, err
	}
	return s.getParticipantDetail(ctx, questionnaireID, participantID, name)
}

func (s *quizService) getParticipantDetail(ctx context.Context, questionnaireID string, participantID string, facilityName string) (*ParticipantDetailResponse, error) {
	q, err := s.repo.FindByID(ctx, questionnaireID)
	if err != nil {
		return nil, err
	}
	questionnaireScope := q.Scope
	if questionnaireScope == "" {
		if q.EducationID != nil {
			questionnaireScope = domain.ScopeEducation
		} else if q.Type == domain.TypePreTest {
			questionnaireScope = domain.ScopeApplication
		} else {
			questionnaireScope = domain.ScopeEducation
		}
	}

	attempt, err := s.repo.FindAttemptByID(ctx, questionnaireID, participantID, facilityName)
	if err != nil {
		return nil, err
	}

	patientName := "-"
	patientAvatar := ""
	puskesmas := "DSMES Platform"
	if attempt.Patient != nil {
		patientName = attempt.Patient.FullName
		patientAvatar = attempt.Patient.ProfilePhotoURL
		if attempt.Patient.HealthFacility != "" {
			puskesmas = attempt.Patient.HealthFacility
		}
	}

	attemptCatStr := ""
	if attempt.SelfEfficacyCategory != nil {
		attemptCatStr = *attempt.SelfEfficacyCategory
	}

	partResp := ParticipantResponse{
		ID:                   attempt.ID,
		PatientID:            attempt.PatientID,
		PatientName:          patientName,
		PatientAvatar:        patientAvatar,
		Puskesmas:            puskesmas,
		CompletionDate:       attempt.CompletedAt,
		Score:                attempt.Score,
		Passed:               attempt.Passed,
		Duration:             formatDuration(attempt.DurationSeconds),
		SelfEfficacyCategory: attemptCatStr,
		ScoreDSMES:           attempt.ScoreDSMES,
		ScoreDSMQR:           attempt.ScoreDSMQR,
		DSMQRCategory:        attempt.DSMQRCategory,
		ScorePreTest:         attempt.ScorePreTest,
	}

	ansMap := make(map[string]domain.QuizAnswer)
	for _, ans := range attempt.Answers {
		ansMap[ans.QuestionID] = ans
	}

	var analysis []QuestionAnalysisResponse
	qNum := 1

	for _, cat := range q.Categories {
		instCat := strings.ToUpper(strings.TrimSpace(cat.InstrumentCategory))
		if instCat == "" {
			instCat = "PRE_TEST"
		}

		for _, quest := range cat.Questions {
			userAnsText := "-"
			correctAnsText := "-"
			isCorr := false

			if questionnaireScope == domain.ScopeApplication {
				correctAnsText = ""
				isCorr = true
				if userAns, ok := ansMap[quest.ID]; ok {
					// Check if answered with a custom option
					foundCustom := false
					for _, opt := range quest.Options {
						if userAns.OptionID != nil && *userAns.OptionID == opt.ID {
							userAnsText = opt.OptionText
							foundCustom = true
							break
						}
					}
					if !foundCustom && userAns.SelectedValue != nil {
						val := *userAns.SelectedValue
						label := "Cukup Yakin"
						for _, choice := range DefaultLikertChoices {
							if choice.ID == strconv.Itoa(val) {
								label = choice.OptionText
								break
							}
						}
						userAnsText = strconv.Itoa(val) + ". " + label
					}
				}
			} else {
				for _, opt := range quest.Options {
					if opt.IsCorrect {
						correctAnsText = opt.OptionText
					}
					if userAns, ok := ansMap[quest.ID]; ok {
						if (userAns.OptionID != nil && *userAns.OptionID == opt.ID) || userAns.SelectedOption == opt.ID {
							userAnsText = opt.OptionText
							isCorr = userAns.IsCorrect
						}
					}
				}
			}

			analysis = append(analysis, QuestionAnalysisResponse{
				ID:                 quest.ID,
				QuestionNumber:     qNum,
				QuestionText:       quest.QuestionText,
				InstrumentCategory: instCat,
				PatientAnswer:      userAnsText,
				CorrectAnswer:      correctAnsText,
				IsCorrect:          isCorr,
				Explanation:        quest.Explanation,
			})
			qNum++
		}
	}

	return &ParticipantDetailResponse{
		Participant:      partResp,
		QuizTitle:        q.Title,
		Type:             string(q.Type),
		Scope:            string(questionnaireScope),
		QuestionAnalysis: analysis,
	}, nil
}

func (s *quizService) ListPatientQuestionnaires(ctx context.Context, qType string, patientID string, page, perPage int) ([]PatientQuestionnaireItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.repo.FindActiveForPatient(ctx, qType, patientID, page, perPage)
}

func (s *quizService) GetMyAttempt(ctx context.Context, patientID, questionnaireID string) (*MyAttemptResponse, error) {
	attempt, err := s.repo.FindMyAttempt(ctx, patientID, questionnaireID)
	if err != nil {
		return nil, err
	}

	q, err := s.repo.FindByID(ctx, questionnaireID)
	if err != nil {
		return nil, err
	}

	totalQuestions := 0
	for _, cat := range q.Categories {
		totalQuestions += len(cat.Questions)
	}

	correctCount := 0
	for _, ans := range attempt.Answers {
		if ans.IsCorrect {
			correctCount++
		}
	}
	if len(attempt.Answers) == 0 && totalQuestions > 0 {
		correctCount = (attempt.Score * totalQuestions) / 100
	}
	incorrectCount := totalQuestions - correctCount

	return &MyAttemptResponse{
		AttemptID:       attempt.ID,
		QuestionnaireID: questionnaireID,
		Score:           attempt.Score,
		Passed:          attempt.Passed,
		TotalQuestions:  totalQuestions,
		CorrectCount:    correctCount,
		IncorrectCount:  incorrectCount,
		Percentage:      attempt.Score,
		CompletedAt:     attempt.CompletedAt,
	}, nil
}

func (s *quizService) GetMyAttemptDetail(ctx context.Context, patientID, questionnaireID string) (*ParticipantDetailResponse, error) {
	return s.GetParticipantDetail(ctx, questionnaireID, patientID)
}

func (s *quizService) GetMyAttemptDetailByID(ctx context.Context, patientID, questionnaireID, attemptID string) (*ParticipantDetailResponse, error) {
	q, err := s.repo.FindByID(ctx, questionnaireID)
	if err != nil {
		return nil, err
	}
	questionnaireScope := q.Scope
	if questionnaireScope == "" {
		if q.EducationID != nil {
			questionnaireScope = domain.ScopeEducation
		} else {
			questionnaireScope = domain.ScopeApplication
		}
	}
	attempt, err := s.repo.FindMyAttemptByID(ctx, patientID, questionnaireID, attemptID)
	if err != nil {
		return nil, err
	}

	patientName := "-"
	puskesmas := "DSMES Platform"
	if attempt.Patient != nil {
		patientName = attempt.Patient.FullName
		if attempt.Patient.HealthFacility != "" {
			puskesmas = attempt.Patient.HealthFacility
		}
	}
	category := ""
	if attempt.SelfEfficacyCategory != nil {
		category = *attempt.SelfEfficacyCategory
	}
	answers := make(map[string]domain.QuizAnswer)
	for _, answer := range attempt.Answers {
		answers[answer.QuestionID] = answer
	}
	analysis := make([]QuestionAnalysisResponse, 0)
	questionNumber := 1
	for _, cat := range q.Categories {
		for _, question := range cat.Questions {
			answerText := "-"
			correctText := "-"
			isCorrect := false
			if questionnaireScope == domain.ScopeApplication {
				correctText = ""
				isCorrect = true
				if answer, ok := answers[question.ID]; ok {
					if answer.OptionID != nil {
						for _, option := range question.Options {
							if option.ID == *answer.OptionID {
								answerText = option.OptionText
								break
							}
						}
					}
					if answerText == "-" && answer.SelectedValue != nil {
						value := *answer.SelectedValue
						label := "Cukup Yakin"
						for _, choice := range DefaultLikertChoices {
							if choice.ID == strconv.Itoa(value) {
								label = choice.OptionText
								break
							}
						}
						answerText = strconv.Itoa(value) + ". " + label
					}
				}
			} else if answer, ok := answers[question.ID]; ok {
				for _, option := range question.Options {
					if option.IsCorrect {
						correctText = option.OptionText
					}
					if answer.OptionID != nil && *answer.OptionID == option.ID {
						answerText = option.OptionText
					}
				}
				isCorrect = answer.IsCorrect
			}
			analysis = append(analysis, QuestionAnalysisResponse{
				ID: question.ID, QuestionNumber: questionNumber, QuestionText: question.QuestionText,
				PatientAnswer: answerText, CorrectAnswer: correctText, IsCorrect: isCorrect,
				Explanation: question.Explanation,
			})
			questionNumber++
		}
	}
	return &ParticipantDetailResponse{
		Participant: ParticipantResponse{
			ID: attempt.ID, PatientID: attempt.PatientID, PatientName: patientName,
			Puskesmas: puskesmas, CompletionDate: attempt.CompletedAt, Score: attempt.Score,
			Passed: attempt.Passed, Duration: formatDuration(attempt.DurationSeconds),
			SelfEfficacyCategory: category,
		},
		QuizTitle: q.Title, Type: string(q.Type), Scope: string(questionnaireScope), QuestionAnalysis: analysis,
	}, nil
}

func (s *quizService) GetMyHistory(ctx context.Context, patientID, qType, scope string) ([]MyHistoryItemResponse, error) {
	attempts, err := s.repo.FindMyHistory(
		ctx,
		patientID,
		strings.ToUpper(strings.TrimSpace(qType)),
		strings.ToUpper(strings.TrimSpace(scope)),
	)
	if err != nil {
		return nil, err
	}

	result := make([]MyHistoryItemResponse, 0, len(attempts))
	for _, a := range attempts {
		title := ""
		qTypeStr := ""
		scopeStr := ""
		if a.Questionnaire != nil {
			title = a.Questionnaire.Title
			qTypeStr = string(a.Questionnaire.Type)
			scopeStr = string(a.Questionnaire.Scope)
		}

		correctCount := 0
		for _, ans := range a.Answers {
			if ans.IsCorrect {
				correctCount++
			}
		}

		totalQuestions := 0
		if a.Questionnaire != nil {
			for _, cat := range a.Questionnaire.Categories {
				totalQuestions += len(cat.Questions)
			}
		}

		if len(a.Answers) == 0 && totalQuestions > 0 {
			correctCount = (a.Score * totalQuestions) / 100
		}
		incorrectCount := totalQuestions - correctCount

		result = append(result, MyHistoryItemResponse{
			AttemptID:          a.ID,
			QuestionnaireID:    a.QuestionnaireID,
			QuestionnaireTitle: title,
			Type:               qTypeStr,
			Scope:              scopeStr,
			Score:              a.Score,
			Passed:             a.Passed,
			TotalQuestions:     totalQuestions,
			CorrectCount:       correctCount,
			IncorrectCount:     incorrectCount,
			Percentage:         a.Score,
			CompletedAt:        a.CompletedAt,
		})
	}
	return result, nil
}
