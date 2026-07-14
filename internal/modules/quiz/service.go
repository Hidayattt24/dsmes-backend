package quiz

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type quizService struct {
	repo QuizRepository
	log  *zap.Logger
}

func NewQuizService(repo QuizRepository, log *zap.Logger) QuizService {
	return &quizService{repo: repo, log: log}
}

func (s *quizService) ListQuizzes(ctx context.Context, search string, status string, page, limit int) ([]QuizResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	items, total, err := s.repo.FindAll(ctx, search, status, page, limit)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]QuizResponse, len(items))
	for i := range items {
		count, _ := s.repo.CountAttempts(ctx, items[i].ID)
		avg, _ := s.repo.GetAverageScore(ctx, items[i].ID)
		resp[i] = ToQuizResponse(&items[i], len(items[i].Questions), count, avg)
	}

	return resp, total, nil
}

func (s *quizService) GetQuiz(ctx context.Context, id string, isAdminOrStaff bool) (*QuizDetailResponse, error) {
	q, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	count, _ := s.repo.CountAttempts(ctx, q.ID)
	avg, _ := s.repo.GetAverageScore(ctx, q.ID)

	res := ToQuizDetailResponse(q, len(q.Questions), count, avg, isAdminOrStaff)
	return &res, nil
}

func (s *quizService) CreateQuiz(ctx context.Context, staffID string, req CreateQuizRequest) (*QuizDetailResponse, error) {
	quiz := &domain.Quiz{
		Title:           req.Title,
		LinkedArticleID: req.LinkedArticleID,
		Difficulty:      req.Difficulty,
		PassingScore:    req.PassingScore,
		Status:          normalizeStatus(req.Status),
		CreatedBy:       &staffID,
	}

	if err := s.repo.Create(ctx, quiz); err != nil {
		return nil, err
	}

	questions := make([]domain.QuizQuestion, len(req.Questions))
	for i, q := range req.Questions {
		questions[i] = domain.QuizQuestion{
			QuestionText:  q.QuestionText,
			OptionA:       q.OptionA,
			OptionB:       q.OptionB,
			OptionC:       q.OptionC,
			OptionD:       q.OptionD,
			CorrectOption: q.CorrectOption,
			Explanation:   q.Explanation,
		}
	}

	if err := s.repo.ReplaceQuestions(ctx, quiz.ID, questions); err != nil {
		return nil, err
	}

	// Fetch full details
	return s.GetQuiz(ctx, quiz.ID, true)
}

func (s *quizService) UpdateQuiz(ctx context.Context, id string, req CreateQuizRequest) (*QuizDetailResponse, error) {
	quiz, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	quiz.Title = req.Title
	quiz.LinkedArticleID = req.LinkedArticleID
	quiz.Difficulty = req.Difficulty
	quiz.PassingScore = req.PassingScore
	quiz.Status = normalizeStatus(req.Status)

	if err := s.repo.Update(ctx, quiz); err != nil {
		return nil, err
	}

	questions := make([]domain.QuizQuestion, len(req.Questions))
	for i, q := range req.Questions {
		questions[i] = domain.QuizQuestion{
			QuestionText:  q.QuestionText,
			OptionA:       q.OptionA,
			OptionB:       q.OptionB,
			OptionC:       q.OptionC,
			OptionD:       q.OptionD,
			CorrectOption: q.CorrectOption,
			Explanation:   q.Explanation,
		}
	}

	if err := s.repo.ReplaceQuestions(ctx, id, questions); err != nil {
		return nil, err
	}

	return s.GetQuiz(ctx, id, true)
}

func (s *quizService) DeleteQuiz(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *quizService) GetStats(ctx context.Context) (*QuizStats, error) {
	return s.repo.GetStats(ctx)
}

func (s *quizService) ListParticipants(ctx context.Context, quizID string) ([]ParticipantResponse, error) {
	attempts, err := s.repo.FindAttemptsByQuizID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	resp := make([]ParticipantResponse, len(attempts))
	for i, a := range attempts {
		pName := "-"
		pAvatar := ""
		puskesmas := "-"
		if a.Patient != nil {
			pName = a.Patient.FullName
			pAvatar = a.Patient.ProfilePhotoURL
			if a.Patient.AssignedStaff != nil {
				// Use PositionTitle as puskesmas (e.g. "Puskesmas Kuta Alam")
				puskesmas = a.Patient.AssignedStaff.PositionTitle
				if strings.TrimSpace(puskesmas) == "" {
					puskesmas = a.Patient.AssignedStaff.FullName
				}
			}
		}

		resp[i] = ParticipantResponse{
			ID:             a.ID,
			PatientID:      a.PatientID,
			PatientName:    pName,
			PatientAvatar:  pAvatar,
			Puskesmas:      puskesmas,
			CompletionDate: a.CompletedAt,
			Score:          a.Score,
			Passed:         a.Passed,
			Duration:       formatDuration(a.DurationSeconds),
		}
	}

	return resp, nil
}

func (s *quizService) GetParticipantDetail(ctx context.Context, quizID string, participantID string) (*ParticipantDetailResponse, error) {
	attempt, err := s.repo.FindAttemptByID(ctx, quizID, participantID)
	if err != nil {
		return nil, err
	}

	quiz, err := s.repo.FindByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	pName := "-"
	pAvatar := ""
	puskesmas := "-"
	if attempt.Patient != nil {
		pName = attempt.Patient.FullName
		pAvatar = attempt.Patient.ProfilePhotoURL
		if attempt.Patient.AssignedStaff != nil {
			puskesmas = attempt.Patient.AssignedStaff.PositionTitle
			if strings.TrimSpace(puskesmas) == "" {
				puskesmas = attempt.Patient.AssignedStaff.FullName
			}
		}
	}

	partResp := ParticipantResponse{
		ID:             attempt.ID,
		PatientID:      attempt.PatientID,
		PatientName:    pName,
		PatientAvatar:  pAvatar,
		Puskesmas:      puskesmas,
		CompletionDate: attempt.CompletedAt,
		Score:          attempt.Score,
		Passed:         attempt.Passed,
		Duration:       formatDuration(attempt.DurationSeconds),
	}

	analysis := make([]QuestionAnalysisResponse, len(attempt.Answers))
	for i, ans := range attempt.Answers {
		qText := ""
		correctAnswerText := ""
		patientAnswerText := ""
		explanation := ""

		if ans.Question != nil {
			qText = ans.Question.QuestionText
			explanation = ans.Question.Explanation

			// Map answer letter to full text
			switch ans.Question.CorrectOption {
			case "A":
				correctAnswerText = "A. " + ans.Question.OptionA
			case "B":
				correctAnswerText = "B. " + ans.Question.OptionB
			case "C":
				correctAnswerText = "C. " + ans.Question.OptionC
			case "D":
				correctAnswerText = "D. " + ans.Question.OptionD
			}

			switch ans.SelectedOption {
			case "A":
				patientAnswerText = "A. " + ans.Question.OptionA
			case "B":
				patientAnswerText = "B. " + ans.Question.OptionB
			case "C":
				patientAnswerText = "C. " + ans.Question.OptionC
			case "D":
				patientAnswerText = "D. " + ans.Question.OptionD
			}
		}

		analysis[i] = QuestionAnalysisResponse{
			ID:             ans.ID,
			QuestionNumber: i + 1,
			QuestionText:   qText,
			PatientAnswer:  patientAnswerText,
			CorrectAnswer:  correctAnswerText,
			IsCorrect:      ans.IsCorrect,
			Explanation:    explanation,
		}
	}

	return &ParticipantDetailResponse{
		Participant:      partResp,
		QuizTitle:        quiz.Title,
		QuestionAnalysis: analysis,
	}, nil
}
