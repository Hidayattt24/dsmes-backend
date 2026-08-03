package quiz

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
)

type quizService struct {
	repo QuizRepository
	log  *zap.Logger
}

func NewQuizService(repo QuizRepository, log *zap.Logger) QuizService {
	return &quizService{repo: repo, log: log}
}

func (s *quizService) ListQuestionnaires(ctx context.Context, search, qType, status, sortBy, sortOrder string, page, limit int) ([]QuestionnaireResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	items, total, err := s.repo.FindAll(ctx, search, qType, status, sortBy, sortOrder, page, limit)
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
		pCount, _ := s.repo.CountAttempts(ctx, q.ID)
		avgScore, _ := s.repo.GetAverageScore(ctx, q.ID)

		resp[i] = ToQuestionnaireResponse(&q, catCount, questCount, pCount, avgScore)
	}

	return resp, total, nil
}

func (s *quizService) GetQuestionnaire(ctx context.Context, id string, isAdminOrStaff bool) (*QuestionnaireDetailResponse, error) {
	q, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	pCount, _ := s.repo.CountAttempts(ctx, q.ID)
	avgScore, _ := s.repo.GetAverageScore(ctx, q.ID)
	resp := ToQuestionnaireDetailResponse(q, pCount, avgScore, isAdminOrStaff)
	return &resp, nil
}

func (s *quizService) GetActivePreTest(ctx context.Context, isAdminOrStaff bool) (*QuestionnaireDetailResponse, error) {
	q, err := s.repo.GetActivePreTest(ctx)
	if err != nil {
		return nil, err
	}
	pCount, _ := s.repo.CountAttempts(ctx, q.ID)
	avgScore, _ := s.repo.GetAverageScore(ctx, q.ID)
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
	pCount, _ := s.repo.CountAttempts(ctx, q.ID)
	avgScore, _ := s.repo.GetAverageScore(ctx, q.ID)
	resp := ToQuestionnaireDetailResponse(q, pCount, avgScore, isAdminOrStaff)
	return &resp, nil
}

func (s *quizService) validateQuestionnairePayload(ctx context.Context, questionnaireID string, req CreateQuestionnaireRequest) (domain.QuestionnaireType, string, error) {
	qType := normalizeType(req.Type)
	status := normalizeStatus(req.Status)

	if qType == domain.TypePreTest {
		// Pre-Test rules
		if status == "aktif" {
			activePre, err := s.repo.GetActivePreTest(ctx)
			if err == nil && activePre != nil && activePre.ID != questionnaireID {
				return qType, status, errs.NewBadRequest("Sistem hanya memperbolehkan 1 Pre-Test aktif")
			}
		}
		if len(req.Categories) == 0 && len(req.Questions) == 0 {
			return qType, status, errs.NewBadRequest("Pre-Test wajib memiliki minimal 1 Pertanyaan")
		}
	} else {
		// Post-Test rules
		if req.EducationID == nil || strings.TrimSpace(*req.EducationID) == "" {
			return qType, status, errs.NewBadRequest("Post-Test wajib terhubung dengan 1 Materi Edukasi")
		}
		if req.PassingScore == nil {
			return qType, status, errs.NewBadRequest("Post-Test wajib memiliki Nilai Kelulusan (Passing Score)")
		}
		if req.Difficulty == nil || strings.TrimSpace(*req.Difficulty) == "" {
			return qType, status, errs.NewBadRequest("Post-Test wajib memiliki Tingkat Kesulitan")
		}

		// Check if duplicate Post-Test exists for this education material
		existingPost, err := s.repo.GetPostTestByEducation(ctx, *req.EducationID)
		if err == nil && existingPost != nil && existingPost.ID != questionnaireID {
			return qType, status, errs.NewBadRequest("Materi Edukasi ini sudah memiliki Post-Test")
		}

		if len(req.Categories) == 0 {
			return qType, status, errs.NewBadRequest("Kuisioner wajib memiliki minimal 1 Kategori")
		}
	}

	return qType, status, nil
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
		questions := make([]domain.Question, len(cReq.Questions))
		for j, qReq := range cReq.Questions {
			var imgURL *string
			if strings.TrimSpace(qReq.QuestionImageURL) != "" {
				urlVal := qReq.QuestionImageURL
				imgURL = &urlVal
			}

			var options []domain.QuestionOption
			if qType == domain.TypePostTest && len(qReq.Choices) > 0 {
				options = make([]domain.QuestionOption, len(qReq.Choices))
				for k, optReq := range qReq.Choices {
					options[k] = domain.QuestionOption{
						OptionText:   optReq.OptionText,
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
			Title:        title,
			Description:  cReq.Description,
			DisplayOrder: i,
			Questions:    questions,
		}
	}
	return categories
}

func (s *quizService) CreateQuestionnaire(ctx context.Context, staffID string, req CreateQuestionnaireRequest) (*QuestionnaireDetailResponse, error) {
	qType, status, err := s.validateQuestionnairePayload(ctx, "", req)
	if err != nil {
		return nil, err
	}

	var eduID *string
	var passingScore *int
	var difficulty *string

	if qType == domain.TypePostTest {
		eduID = req.EducationID
		passingScore = req.PassingScore
		difficulty = req.Difficulty
	}

	createdBy := staffID
	q := &domain.Questionnaire{
		Title:        req.Title,
		Type:         qType,
		Description:  req.Description,
		EducationID:  eduID,
		PassingScore: passingScore,
		Difficulty:   difficulty,
		Status:       status,
		CreatedBy:    &createdBy,
		Categories:   buildCategoriesFromRequest(req.Title, qType, req.Categories, req.Questions),
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

	qType, status, err := s.validateQuestionnairePayload(ctx, id, req)
	if err != nil {
		return nil, err
	}

	var eduID *string
	var passingScore *int
	var difficulty *string

	if qType == domain.TypePostTest {
		eduID = req.EducationID
		passingScore = req.PassingScore
		difficulty = req.Difficulty
	}

	q.Title = req.Title
	q.Type = qType
	q.Description = req.Description
	q.EducationID = eduID
	q.PassingScore = passingScore
	q.Difficulty = difficulty
	q.Status = status
	q.Categories = buildCategoriesFromRequest(req.Title, qType, req.Categories, req.Questions)

	if err := s.repo.Update(ctx, q); err != nil {
		return nil, err
	}

	return s.GetQuestionnaire(ctx, id, true)
}

func (s *quizService) DeleteQuestionnaire(ctx context.Context, id string) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *quizService) GetStats(ctx context.Context) (*QuizStats, error) {
	return s.repo.GetStats(ctx)
}

func (s *quizService) SubmitQuestionnaire(ctx context.Context, patientID string, questionnaireID string, req SubmitQuestionnaireRequest) (*SubmitResultResponse, error) {
	q, err := s.repo.FindByID(ctx, questionnaireID)
	if err != nil {
		return nil, err
	}

	totalQuestions := 0
	for _, cat := range q.Categories {
		totalQuestions += len(cat.Questions)
	}

	if totalQuestions == 0 {
		return nil, errs.NewBadRequest("Kuisioner tidak memiliki pertanyaan")
	}

	if q.Type == domain.TypePreTest {
		// ── DMSES Pre-Test Submission (Likert Scale 1-5) ────────────────────
		totalDMSESScore := 0
		answers := make([]domain.QuizAnswer, len(req.Answers))

		for i, ans := range req.Answers {
			val := 1
			if ans.SelectedValue != nil && *ans.SelectedValue >= 1 && *ans.SelectedValue <= 5 {
				val = *ans.SelectedValue
			}
			totalDMSESScore += val

			selectedVal := val
			answers[i] = domain.QuizAnswer{
				QuestionID:     ans.QuestionID,
				SelectedOption: fmt.Sprintf("%d", val),
				SelectedValue:  &selectedVal,
				IsCorrect:      true,
			}
		}

		selfEfficacyCat := DetermineSelfEfficacyCategory(totalDMSESScore)

		attempt := &domain.QuizAttempt{
			QuestionnaireID:      q.ID,
			PatientID:            patientID,
			Score:                totalDMSESScore,
			SelfEfficacyCategory: &selfEfficacyCat,
			Passed:               true,
			DurationSeconds:      req.DurationSeconds,
			Answers:              answers,
		}

		if err := s.repo.SaveAttempt(ctx, attempt); err != nil {
			return nil, err
		}

		// Return PRE_TEST completion response (Score & Passed omitted from user display)
		return &SubmitResultResponse{
			AttemptID:       attempt.ID,
			QuestionnaireID: q.ID,
			Type:            "PRE_TEST",
			Message:         "Jawaban Anda telah berhasil disimpan dan akan digunakan untuk membantu memahami tingkat keyakinan Anda dalam mengelola diabetes.",
			TotalQuestions:  totalQuestions,
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
		QuestionnaireID: q.ID,
		PatientID:       patientID,
		Score:           score,
		Passed:          passed,
		DurationSeconds: req.DurationSeconds,
		Answers:         answers,
	}

	if err := s.repo.SaveAttempt(ctx, attempt); err != nil {
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
	attempts, err := s.repo.FindAttemptsByQuestionnaireID(ctx, questionnaireID)
	if err != nil {
		return nil, err
	}

	resps := make([]ParticipantResponse, len(attempts))
	for i, a := range attempts {
		name := "-"
		avatar := ""
		if a.Patient != nil {
			name = a.Patient.FullName
			avatar = a.Patient.ProfilePhotoURL
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
			Puskesmas:            "DSMES Platform",
			CompletionDate:       a.CompletedAt,
			Score:                a.Score,
			Passed:               a.Passed,
			Duration:             formatDuration(a.DurationSeconds),
			SelfEfficacyCategory: catStr,
		}
	}
	return resps, nil
}

func (s *quizService) GetParticipantDetail(ctx context.Context, questionnaireID string, participantID string) (*ParticipantDetailResponse, error) {
	q, err := s.repo.FindByID(ctx, questionnaireID)
	if err != nil {
		return nil, err
	}

	attempt, err := s.repo.FindAttemptByID(ctx, questionnaireID, participantID)
	if err != nil {
		return nil, err
	}

	patientName := "-"
	patientAvatar := ""
	if attempt.Patient != nil {
		patientName = attempt.Patient.FullName
		patientAvatar = attempt.Patient.ProfilePhotoURL
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
		Puskesmas:            "DSMES Platform",
		CompletionDate:       attempt.CompletedAt,
		Score:                attempt.Score,
		Passed:               attempt.Passed,
		Duration:             formatDuration(attempt.DurationSeconds),
		SelfEfficacyCategory: attemptCatStr,
	}

	ansMap := make(map[string]domain.QuizAnswer)
	for _, ans := range attempt.Answers {
		ansMap[ans.QuestionID] = ans
	}

	var analysis []QuestionAnalysisResponse
	qNum := 1

	for _, cat := range q.Categories {
		for _, quest := range cat.Questions {
			userAnsText := "-"
			correctAnsText := "-"
			isCorr := false

			if q.Type == "PRE_TEST" {
				correctAnsText = ""
				isCorr = true
				if userAns, ok := ansMap[quest.ID]; ok {
					if userAns.SelectedValue != nil {
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
				ID:             quest.ID,
				QuestionNumber: qNum,
				QuestionText:   quest.QuestionText,
				PatientAnswer:  userAnsText,
				CorrectAnswer:  correctAnsText,
				IsCorrect:      isCorr,
				Explanation:    quest.Explanation,
			})
			qNum++
		}
	}

	return &ParticipantDetailResponse{
		Participant:      partResp,
		QuizTitle:        q.Title,
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

func (s *quizService) GetMyHistory(ctx context.Context, patientID, qType string) ([]MyHistoryItemResponse, error) {
	attempts, err := s.repo.FindMyHistory(ctx, patientID, strings.ToUpper(strings.TrimSpace(qType)))
	if err != nil {
		return nil, err
	}

	result := make([]MyHistoryItemResponse, 0, len(attempts))
	for _, a := range attempts {
		title := ""
		qTypeStr := ""
		if a.Questionnaire != nil {
			title = a.Questionnaire.Title
			qTypeStr = string(a.Questionnaire.Type)
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
