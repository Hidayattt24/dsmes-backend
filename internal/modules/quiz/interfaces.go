package quiz

import (
	"context"

	"github.com/dsmes/dsmes-backend/internal/domain"
)

type QuizRepository interface {
	FindAll(ctx context.Context, search string, status string, page, limit int) ([]domain.Quiz, int64, error)
	FindByID(ctx context.Context, id string) (*domain.Quiz, error)
	Create(ctx context.Context, q *domain.Quiz) error
	Update(ctx context.Context, q *domain.Quiz) error
	Delete(ctx context.Context, id string) error
	GetStats(ctx context.Context) (*QuizStats, error)

	ReplaceQuestions(ctx context.Context, quizID string, questions []domain.QuizQuestion) error

	// Participant Attempts
	FindAttemptsByQuizID(ctx context.Context, quizID string) ([]domain.QuizAttempt, error)
	FindAttemptByID(ctx context.Context, quizID string, participantID string) (*domain.QuizAttempt, error)
	CountAttempts(ctx context.Context, quizID string) (int, error)
	GetAverageScore(ctx context.Context, quizID string) (*int, error)
}

type QuizService interface {
	ListQuizzes(ctx context.Context, search string, status string, page, limit int) ([]QuizResponse, int64, error)
	GetQuiz(ctx context.Context, id string, isAdminOrStaff bool) (*QuizDetailResponse, error)
	CreateQuiz(ctx context.Context, staffID string, req CreateQuizRequest) (*QuizDetailResponse, error)
	UpdateQuiz(ctx context.Context, id string, req CreateQuizRequest) (*QuizDetailResponse, error)
	DeleteQuiz(ctx context.Context, id string) error
	GetStats(ctx context.Context) (*QuizStats, error)

	ListParticipants(ctx context.Context, quizID string) ([]ParticipantResponse, error)
	GetParticipantDetail(ctx context.Context, quizID string, participantID string) (*ParticipantDetailResponse, error)
}
