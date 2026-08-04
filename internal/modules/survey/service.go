package survey

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/dsmes/dsmes-backend/internal/domain"
	"github.com/dsmes/dsmes-backend/internal/pkg/errs"
	"go.uber.org/zap"
)

type surveyService struct {
	repo SurveyRepository
	log  *zap.Logger
}

func NewSurveyService(repo SurveyRepository, log *zap.Logger) SurveyService {
	return &surveyService{repo: repo, log: log}
}

func (s *surveyService) CreateSurvey(ctx context.Context, adminID string, req CreateSurveyRequest) (*SurveyDetailResponse, error) {
	if len(req.Questions) == 0 {
		return nil, errs.NewBadRequest("at least one question is required to create a survey")
	}

	survey := &domain.Survey{
		Title:             req.Title,
		Description:       req.Description,
		Type:              req.Type,
		Status:            domain.SurveyStatusDraft,
		IsActive:          false,
		StartDate:         req.StartDate,
		EndDate:           req.EndDate,
		CreatedBy:         &adminID,
	}

	if err := s.repo.Create(ctx, survey); err != nil {
		return nil, err
	}

	questions := make([]domain.SurveyQuestion, len(req.Questions))
	for i, qReq := range req.Questions {
		isRequired := true
		if qReq.IsRequired != nil {
			isRequired = *qReq.IsRequired
		}
		labelsJSON, _ := json.Marshal(qReq.LikertLabels)
		if len(qReq.LikertLabels) == 0 {
			labelsJSON = json.RawMessage(`["Sangat Tidak Setuju", "Tidak Setuju", "Netral", "Setuju", "Sangat Setuju"]`)
		}

		questions[i] = domain.SurveyQuestion{
			SurveyID:        survey.ID,
			QuestionText:    qReq.QuestionText,
			Description:     qReq.Description,
			ImageURL:        qReq.ImageURL,
			SVGIllustration: qReq.SVGIllustration,
			LikertLabels:    labelsJSON,
			IsRequired:      isRequired,
			DisplayOrder:    i + 1,
		}
	}

	if err := s.repo.ReplaceQuestions(ctx, survey.ID, questions); err != nil {
		return nil, err
	}

	return s.GetSurveyByID(ctx, survey.ID, false)
}

func (s *surveyService) UpdateSurvey(ctx context.Context, id string, req UpdateSurveyRequest) (*SurveyDetailResponse, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(req.Questions) == 0 {
		return nil, errs.NewBadRequest("at least one question is required")
	}

	existing.Title = req.Title
	existing.Description = req.Description
	existing.Type = req.Type
	existing.StartDate = req.StartDate
	existing.EndDate = req.EndDate

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	questions := make([]domain.SurveyQuestion, len(req.Questions))
	for i, qReq := range req.Questions {
		isRequired := true
		if qReq.IsRequired != nil {
			isRequired = *qReq.IsRequired
		}
		labelsJSON, _ := json.Marshal(qReq.LikertLabels)
		if len(qReq.LikertLabels) == 0 {
			labelsJSON = json.RawMessage(`["Sangat Tidak Setuju", "Tidak Setuju", "Netral", "Setuju", "Sangat Setuju"]`)
		}

		questions[i] = domain.SurveyQuestion{
			SurveyID:        id,
			QuestionText:    qReq.QuestionText,
			Description:     qReq.Description,
			ImageURL:        qReq.ImageURL,
			SVGIllustration: qReq.SVGIllustration,
			LikertLabels:    labelsJSON,
			IsRequired:      isRequired,
			DisplayOrder:    i + 1,
		}
	}

	if err := s.repo.ReplaceQuestions(ctx, id, questions); err != nil {
		return nil, err
	}

	return s.GetSurveyByID(ctx, id, false)
}

func (s *surveyService) DeleteSurvey(ctx context.Context, id string) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *surveyService) GetSurveyByID(ctx context.Context, id string, isPatient bool) (*SurveyDetailResponse, error) {
	survey, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if isPatient && survey.Status != domain.SurveyStatusPublished {
		return nil, errs.NewNotFound("survey is not published")
	}

	qDTOs := make([]QuestionResponseDTO, len(survey.Questions))
	for i, q := range survey.Questions {
		qDTOs[i] = QuestionResponseDTO{
			ID:              q.ID,
			SurveyID:        q.SurveyID,
			QuestionText:    q.QuestionText,
			Description:     q.Description,
			ImageURL:        q.ImageURL,
			SVGIllustration: q.SVGIllustration,
			LikertLabels:    ParseLikertLabels(q.LikertLabels),
			IsRequired:      q.IsRequired,
			DisplayOrder:    q.DisplayOrder,
		}
	}

	return &SurveyDetailResponse{
		ID:                survey.ID,
		Title:             survey.Title,
		Description:       survey.Description,
		Type:              survey.Type,
		Status:            survey.Status,
		IsActive:          survey.IsActive,
		StartDate:         survey.StartDate,
		EndDate:           survey.EndDate,
		QuestionCount:     len(survey.Questions),
		ResponseCount:     len(survey.Responses),
		CreatedAt:         survey.CreatedAt,
		UpdatedAt:         survey.UpdatedAt,
		Questions:         qDTOs,
	}, nil
}

func (s *surveyService) ListSurveys(ctx context.Context, surveyType string, status string, page int, limit int) ([]SurveyListItemResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	items, total, err := s.repo.List(ctx, surveyType, status, page, limit)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]SurveyListItemResponse, len(items))
	for i, item := range items {
		dtos[i] = SurveyListItemResponse{
			ID:            item.ID,
			Title:         item.Title,
			Description:   item.Description,
			Type:          item.Type,
			Status:        item.Status,
			IsActive:      item.IsActive,
			StartDate:     item.StartDate,
			EndDate:       item.EndDate,
			QuestionCount: len(item.Questions),
			ResponseCount: len(item.Responses),
			CreatedAt:     item.CreatedAt,
		}
	}

	return dtos, total, nil
}

func (s *surveyService) UpdateStatus(ctx context.Context, id string, req UpdateSurveyStatusRequest) (*SurveyDetailResponse, error) {
	survey, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Status == domain.SurveyStatusPublished && len(survey.Questions) == 0 {
		return nil, errs.NewBadRequest("cannot publish survey without questions")
	}

	if req.Status != "" {
		survey.Status = req.Status
	}
	if req.IsActive != nil {
		survey.IsActive = *req.IsActive
		if *req.IsActive {
			survey.Status = domain.SurveyStatusPublished
			if err := s.repo.SetActive(ctx, id, string(survey.Type)); err != nil {
				return nil, err
			}
		} else {
			survey.Status = domain.SurveyStatusDraft
		}
	}

	if err := s.repo.Update(ctx, survey); err != nil {
		return nil, err
	}

	return s.GetSurveyByID(ctx, id, false)
}

func (s *surveyService) DuplicateSurvey(ctx context.Context, id string, adminID string) (*SurveyDetailResponse, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	newSurvey := &domain.Survey{
		Title:       fmt.Sprintf("%s (Salinan)", existing.Title),
		Description: existing.Description,
		Type:        existing.Type,
		Status:      domain.SurveyStatusDraft,
		IsActive:    false,
		CreatedBy:   &adminID,
	}

	if err := s.repo.Create(ctx, newSurvey); err != nil {
		return nil, err
	}

	questions := make([]domain.SurveyQuestion, len(existing.Questions))
	for i, q := range existing.Questions {
		questions[i] = domain.SurveyQuestion{
			SurveyID:        newSurvey.ID,
			QuestionText:    q.QuestionText,
			Description:     q.Description,
			ImageURL:        q.ImageURL,
			SVGIllustration: q.SVGIllustration,
			LikertLabels:    q.LikertLabels,
			IsRequired:      q.IsRequired,
			DisplayOrder:    q.DisplayOrder,
		}
	}

	if err := s.repo.ReplaceQuestions(ctx, newSurvey.ID, questions); err != nil {
		return nil, err
	}

	return s.GetSurveyByID(ctx, newSurvey.ID, false)
}

func (s *surveyService) GetActiveSurveysForPatient(ctx context.Context, surveyType string, patientID string) ([]SurveyDetailResponse, error) {
	activeSurveys, err := s.repo.ListActiveSurveys(ctx, surveyType)
	if err != nil {
		return nil, err
	}

	dtos := make([]SurveyDetailResponse, 0, len(activeSurveys))
	for _, activeSurvey := range activeSurveys {
		detail, err := s.GetSurveyByID(ctx, activeSurvey.ID, true)
		if err != nil {
			continue
		}
		if patientID != "" {
			existing, _ := s.repo.GetResponseBySurveyAndPatient(ctx, activeSurvey.ID, patientID)
			if existing != nil {
				detail.HasSubmitted = true
			}
		}
		dtos = append(dtos, *detail)
	}

	return dtos, nil
}

func (s *surveyService) SubmitSurvey(ctx context.Context, surveyID string, patientID string, req SubmitSurveyRequest) (*SubmitSurveyResponse, error) {
	survey, err := s.repo.GetByID(ctx, surveyID)
	if err != nil {
		return nil, err
	}

	if survey.Status != domain.SurveyStatusPublished {
		return nil, errs.NewBadRequest("cannot submit response for unpublished survey")
	}

	// Check duplicate submission
	existing, err := s.repo.GetResponseBySurveyAndPatient(ctx, surveyID, patientID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errs.NewBadRequest("Anda telah menyelesaikan survei ini sebelumnya")
	}

	// Build map of required questions
	qMap := make(map[string]domain.SurveyQuestion)
	for _, q := range survey.Questions {
		qMap[q.ID] = q
	}

	ansMap := make(map[string]int)
	for _, ans := range req.Answers {
		if _, ok := qMap[ans.QuestionID]; !ok {
			return nil, errs.NewBadRequest(fmt.Sprintf("question ID %s does not belong to this survey", ans.QuestionID))
		}
		if ans.RatingValue < 1 || ans.RatingValue > 5 {
			return nil, errs.NewBadRequest("rating value must be between 1 and 5")
		}
		ansMap[ans.QuestionID] = ans.RatingValue
	}

	// Verify required questions
	for _, q := range survey.Questions {
		if q.IsRequired {
			if _, answered := ansMap[q.ID]; !answered {
				return nil, errs.NewBadRequest(fmt.Sprintf("pertanyaan '%s' wajib diisi", q.QuestionText))
			}
		}
	}

	now := time.Now()
	startedAt := req.StartedAt
	if startedAt == nil {
		t := now.Add(-time.Duration(req.DurationSeconds) * time.Second)
		startedAt = &t
	}

	resp := &domain.SurveyResponse{
		SurveyID:        surveyID,
		PatientID:       patientID,
		StartedAt:       startedAt,
		CompletedAt:     now,
		DurationSeconds: req.DurationSeconds,
	}

	answers := make([]domain.SurveyAnswer, 0, len(req.Answers))

	if survey.Type == domain.SurveyTypeUserSatisfaction {
		totalScore := 0.0
		for _, q := range survey.Questions {
			if rating, ok := ansMap[q.ID]; ok {
				val := float64(rating)
				totalScore += val
				answers = append(answers, domain.SurveyAnswer{
					QuestionID:  q.ID,
					RatingValue: rating,
				})
			}
		}

		qCount := float64(len(survey.Questions))
		if qCount > 0 {
			avgScore := totalScore / qCount
			pctScore := (totalScore / (qCount * 5.0)) * 100.0

			resp.TotalScore = &totalScore
			resp.AverageScore = &avgScore
			resp.PercentageScore = &pctScore
		}
	} else if survey.Type == domain.SurveyTypeSUS {
		// Sort questions by display order to guarantee 1-based index (odd/even)
		sortedQuestions := make([]domain.SurveyQuestion, len(survey.Questions))
		copy(sortedQuestions, survey.Questions)
		sort.Slice(sortedQuestions, func(i, j int) bool {
			return sortedQuestions[i].DisplayOrder < sortedQuestions[j].DisplayOrder
		})

		rawScoreSum := 0.0
		for idx, q := range sortedQuestions {
			rating := ansMap[q.ID]
			itemNum := idx + 1 // 1-based index
			var adjusted float64

			if itemNum%2 != 0 {
				// Odd question (1, 3, 5, 7, 9): Score - 1
				adjusted = float64(rating - 1)
			} else {
				// Even question (2, 4, 6, 8, 10): 5 - Score
				adjusted = float64(5 - rating)
			}

			rawScoreSum += adjusted
			answers = append(answers, domain.SurveyAnswer{
				QuestionID:    q.ID,
				RatingValue:   rating,
				AdjustedScore: &adjusted,
			})
		}

		susScore := rawScoreSum * 2.5 // Official SUS formula (0 - 100)
		resp.RawScore = &rawScoreSum
		resp.SUSScore = &susScore

		// Interpretation
		var interp string
		switch {
		case susScore >= 85.0:
			interp = "Excellent"
		case susScore >= 70.0:
			interp = "Good"
		case susScore >= 50.0:
			interp = "OK"
		case susScore >= 35.0:
			interp = "Poor"
		default:
			interp = "Awful"
		}
		resp.Interpretation = &interp

		passed := susScore >= 68.0
		resp.Passed = &passed
	}

	if err := s.repo.CreateResponse(ctx, resp, answers); err != nil {
		return nil, err
	}

	return &SubmitSurveyResponse{
		ResponseID:      resp.ID,
		SurveyID:        surveyID,
		SurveyTitle:     survey.Title,
		Type:            string(survey.Type),
		TotalScore:      resp.TotalScore,
		AverageScore:    resp.AverageScore,
		PercentageScore: resp.PercentageScore,
		RawScore:        resp.RawScore,
		SUSScore:        resp.SUSScore,
		Interpretation:  resp.Interpretation,
		Passed:          resp.Passed,
		CompletedAt:     resp.CompletedAt,
		Message:         "Terima kasih telah menyelesaikan survei. Masukan Anda sangat membantu penelitian dan pengembangan aplikasi DSMES Aceh.",
	}, nil
}

func (s *surveyService) GetSurveyResponses(ctx context.Context, surveyID string, page int, limit int) ([]SurveyResponseItemResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	responses, total, err := s.repo.ListResponses(ctx, surveyID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]SurveyResponseItemResponse, len(responses))
	for i, r := range responses {
		patientName := "Anonim"
		patientEmail := ""
		patientPhone := ""
		if r.Patient != nil {
			patientName = r.Patient.FullName
			if r.Patient.Email != nil {
				patientEmail = *r.Patient.Email
			}
			patientPhone = r.Patient.WhatsappNumber
		}

		ansDTOs := make([]AnswerDetailDTO, len(r.Answers))
		for j, a := range r.Answers {
			qText := ""
			if a.Question != nil {
				qText = a.Question.QuestionText
			}
			ansDTOs[j] = AnswerDetailDTO{
				QuestionID:    a.QuestionID,
				QuestionText:  qText,
				RatingValue:   a.RatingValue,
				AdjustedScore: a.AdjustedScore,
			}
		}

		dtos[i] = SurveyResponseItemResponse{
			ID:              r.ID,
			SurveyID:        r.SurveyID,
			PatientID:       r.PatientID,
			PatientName:     patientName,
			PatientEmail:    patientEmail,
			PatientPhone:    patientPhone,
			StartedAt:       r.StartedAt,
			CompletedAt:     r.CompletedAt,
			DurationSeconds: r.DurationSeconds,
			TotalScore:      r.TotalScore,
			AverageScore:    r.AverageScore,
			PercentageScore: r.PercentageScore,
			RawScore:        r.RawScore,
			SUSScore:        r.SUSScore,
			Interpretation:  r.Interpretation,
			Passed:          r.Passed,
			Answers:         ansDTOs,
		}
	}

	return dtos, total, nil
}

func (s *surveyService) GetSurveyAnalytics(ctx context.Context, surveyID string) (*SurveyAnalyticsResponse, error) {
	return s.repo.GetAnalytics(ctx, surveyID)
}

func (s *surveyService) ExportResponsesCSV(ctx context.Context, surveyID string) ([]byte, string, error) {
	survey, err := s.repo.GetByID(ctx, surveyID)
	if err != nil {
		return nil, "", err
	}

	responses, err := s.repo.GetAllResponsesForExport(ctx, surveyID)
	if err != nil {
		return nil, "", err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Build CSV header dynamically based on survey type and questions
	header := []string{
		"No",
		"ID Respon",
		"Nama Pasien",
		"No WhatsApp",
		"Tanggal Selesai",
		"Durasi (detik)",
	}

	if survey.Type == domain.SurveyTypeUserSatisfaction {
		header = append(header, "Total Skor", "Rata-rata Skor", "Persentase (%)")
	} else if survey.Type == domain.SurveyTypeSUS {
		header = append(header, "Raw Score", "Skor SUS (0-100)", "Interpretasi", "Status Kelulusan")
	}

	for idx, q := range survey.Questions {
		header = append(header, fmt.Sprintf("Q%d: %s", idx+1, q.QuestionText))
	}

	if err := writer.Write(header); err != nil {
		return nil, "", errs.NewInternal("failed to write CSV header", err)
	}

	for i, r := range responses {
		patientName := "Anonim"
		phone := ""
		if r.Patient != nil {
			patientName = r.Patient.FullName
			phone = r.Patient.WhatsappNumber
		}

		ansMap := make(map[string]int)
		for _, a := range r.Answers {
			ansMap[a.QuestionID] = a.RatingValue
		}

		row := []string{
			strconv.Itoa(i + 1),
			r.ID,
			patientName,
			phone,
			r.CompletedAt.Format("2006-01-02 15:04:05"),
			strconv.Itoa(r.DurationSeconds),
		}

		if survey.Type == domain.SurveyTypeUserSatisfaction {
			totStr, avgStr, pctStr := "-", "-", "-"
			if r.TotalScore != nil {
				totStr = fmt.Sprintf("%.2f", *r.TotalScore)
			}
			if r.AverageScore != nil {
				avgStr = fmt.Sprintf("%.2f", *r.AverageScore)
			}
			if r.PercentageScore != nil {
				pctStr = fmt.Sprintf("%.2f%%", *r.PercentageScore)
			}
			row = append(row, totStr, avgStr, pctStr)
		} else if survey.Type == domain.SurveyTypeSUS {
			rawStr, susStr, interpStr, passStr := "-", "-", "-", "-"
			if r.RawScore != nil {
				rawStr = fmt.Sprintf("%.2f", *r.RawScore)
			}
			if r.SUSScore != nil {
				susStr = fmt.Sprintf("%.2f", *r.SUSScore)
			}
			if r.Interpretation != nil {
				interpStr = *r.Interpretation
			}
			if r.Passed != nil {
				if *r.Passed {
					passStr = "PASS"
				} else {
					passStr = "FAIL"
				}
			}
			row = append(row, rawStr, susStr, interpStr, passStr)
		}

		for _, q := range survey.Questions {
			val := "-"
			if rating, ok := ansMap[q.ID]; ok {
				val = strconv.Itoa(rating)
			}
			row = append(row, val)
		}

		if err := writer.Write(row); err != nil {
			return nil, "", errs.NewInternal("failed to write CSV row", err)
		}
	}

	writer.Flush()
	filename := fmt.Sprintf("survey_responses_%s_%s.csv", survey.ID, time.Now().Format("20060102_150405"))
	return buf.Bytes(), filename, nil
}
