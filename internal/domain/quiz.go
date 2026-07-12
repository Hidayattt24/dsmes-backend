package domain

import (
	"time"
)

// Quiz represents an evaluation questionnaire linked to an educational article.
type Quiz struct {
	BaseModel

	Title           string `gorm:"type:varchar(200);not null" json:"title"`
	LinkedArticleID string `gorm:"type:uuid;not null" json:"linked_article_id"`
	Difficulty      string `gorm:"type:varchar(50);not null;default:'Sedang'" json:"difficulty"`
	PassingScore    int    `gorm:"type:int;not null;default:80" json:"passing_score"`
	Status          string `gorm:"type:varchar(50);not null;default:'draft'" json:"status"`
	CreatedBy       *string `gorm:"type:uuid" json:"created_by"`

	// Relations
	LinkedArticle *Article       `gorm:"foreignKey:LinkedArticleID" json:"linked_article,omitempty"`
	Questions     []QuizQuestion `gorm:"foreignKey:QuizID;constraint:OnDelete:CASCADE" json:"questions,omitempty"`
}

func (Quiz) TableName() string { return "quizzes" }

// QuizQuestion represents a multiple choice question in a Quiz.
type QuizQuestion struct {
	BaseModel

	QuizID         string `gorm:"type:uuid;not null" json:"quiz_id"`
	QuestionText   string `gorm:"type:text;not null" json:"question_text"`
	OptionA        string `gorm:"type:text;not null" json:"option_a"`
	OptionB        string `gorm:"type:text;not null" json:"option_b"`
	OptionC        string `gorm:"type:text;not null" json:"option_c"`
	OptionD        string `gorm:"type:text;not null" json:"option_d"`
	CorrectOption  string `gorm:"type:varchar(10);not null" json:"correct_option"`
	Explanation    string `gorm:"type:text" json:"explanation"`
}

func (QuizQuestion) TableName() string { return "quiz_questions" }

// QuizAttempt represents a patient's submission/try of a Quiz.
type QuizAttempt struct {
	BaseModel

	QuizID          string    `gorm:"type:uuid;not null" json:"quiz_id"`
	PatientID       string    `gorm:"type:uuid;not null" json:"patient_id"`
	Score           int       `gorm:"type:int;not null" json:"score"`
	Passed          bool      `gorm:"type:boolean;not null" json:"passed"`
	DurationSeconds int       `gorm:"type:int;not null" json:"duration_seconds"`
	CompletedAt     time.Time `gorm:"type:timestamptz;not null;default:now()" json:"completed_at"`

	// Relations
	Quiz    *Quiz         `gorm:"foreignKey:QuizID" json:"quiz,omitempty"`
	Patient *Patient      `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Answers []QuizAnswer  `gorm:"foreignKey:AttemptID;constraint:OnDelete:CASCADE" json:"answers,omitempty"`
}

func (QuizAttempt) TableName() string { return "quiz_attempts" }

// QuizAnswer records the option selected for a question in a quiz attempt.
type QuizAnswer struct {
	BaseModel

	AttemptID      string `gorm:"type:uuid;not null" json:"attempt_id"`
	QuestionID     string `gorm:"type:uuid;not null" json:"question_id"`
	SelectedOption string `gorm:"type:varchar(10);not null" json:"selected_option"`
	IsCorrect      bool   `gorm:"type:boolean;not null" json:"is_correct"`

	// Relations
	Question *QuizQuestion `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}

func (QuizAnswer) TableName() string { return "quiz_answers" }
