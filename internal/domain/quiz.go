package domain

import (
	"time"
)

type QuestionnaireType string

const (
	TypePreTest  QuestionnaireType = "PRE_TEST"
	TypePostTest QuestionnaireType = "POST_TEST"
)

// Questionnaire represents a Pre-Test or Post-Test evaluation questionnaire.
type Questionnaire struct {
	BaseModel

	Title        string            `gorm:"type:varchar(200);not null" json:"title"`
	Type         QuestionnaireType `gorm:"type:questionnaire_type_enum;not null;default:'POST_TEST'" json:"type"`
	Description  string            `gorm:"type:text" json:"description,omitempty"`
	EducationID  *string           `gorm:"type:uuid" json:"education_id,omitempty"`
	PassingScore *int              `gorm:"type:int" json:"passing_score,omitempty"`
	Difficulty   *string           `gorm:"type:varchar(50)" json:"difficulty,omitempty"`
	Status       string            `gorm:"type:varchar(50);not null;default:'draft'" json:"status"`
	CreatedBy    *string           `gorm:"type:uuid" json:"created_by,omitempty"`

	// Relations
	Education  *Article           `gorm:"foreignKey:EducationID" json:"education,omitempty"`
	Categories []QuestionCategory `gorm:"foreignKey:QuestionnaireID;constraint:OnDelete:CASCADE" json:"categories,omitempty"`
}

func (Questionnaire) TableName() string { return "questionnaires" }

// QuestionCategory represents a section/learning category grouping questions.
type QuestionCategory struct {
	BaseModel

	QuestionnaireID string `gorm:"type:uuid;not null" json:"questionnaire_id"`
	Title           string `gorm:"type:varchar(200);not null" json:"title"`
	Description     string `gorm:"type:text" json:"description,omitempty"`
	DisplayOrder    int    `gorm:"type:int;not null;default:0" json:"display_order"`

	// Relations
	Questions []Question `gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE" json:"questions,omitempty"`
}

func (QuestionCategory) TableName() string { return "question_categories" }

// Question represents a single question item inside a category.
type Question struct {
	BaseModel

	CategoryID       *string `gorm:"type:uuid" json:"category_id,omitempty"`
	QuestionText     string  `gorm:"type:text;not null" json:"question_text"`
	QuestionImageURL *string `gorm:"type:text;column:question_image_url" json:"question_image_url,omitempty"`
	Explanation      string  `gorm:"type:text" json:"explanation,omitempty"`
	DisplayOrder     int     `gorm:"type:int;not null;default:0" json:"display_order"`

	// Relations
	Options []QuestionOption `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE" json:"options,omitempty"`
}

func (Question) TableName() string { return "questions" }

// QuestionOption represents an answer option for a question.
type QuestionOption struct {
	BaseModel

	QuestionID   string `gorm:"type:uuid;not null" json:"question_id"`
	OptionText   string `gorm:"type:text;not null" json:"option_text"`
	IsCorrect    bool   `gorm:"type:boolean;not null;default:false" json:"is_correct"`
	DisplayOrder int    `gorm:"type:int;not null;default:0" json:"display_order"`
}

func (QuestionOption) TableName() string { return "question_options" }

// QuizAttempt represents a patient's submission of a Questionnaire.
type QuizAttempt struct {
	BaseModel

	QuestionnaireID      string    `gorm:"column:quiz_id;type:uuid;not null" json:"questionnaire_id"`
	PatientID            string    `gorm:"type:uuid;not null" json:"patient_id"`
	Score                int       `gorm:"type:int;not null" json:"score"`
	SelfEfficacyCategory *string   `gorm:"type:varchar(50);column:self_efficacy_category" json:"self_efficacy_category,omitempty"`
	Passed               bool      `gorm:"type:boolean;not null" json:"passed"`
	DurationSeconds      int       `gorm:"type:int;not null" json:"duration_seconds"`
	CompletedAt          time.Time `gorm:"type:timestamptz;not null;default:now()" json:"completed_at"`

	// Relations
	Questionnaire *Questionnaire `gorm:"foreignKey:QuestionnaireID" json:"questionnaire,omitempty"`
	Patient       *Patient       `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Answers       []QuizAnswer   `gorm:"foreignKey:AttemptID;constraint:OnDelete:CASCADE" json:"answers,omitempty"`
}

func (QuizAttempt) TableName() string { return "quiz_attempts" }

// QuizAnswer records the option selected for a question in a quiz attempt.
type QuizAnswer struct {
	BaseModel

	AttemptID      string  `gorm:"type:uuid;not null" json:"attempt_id"`
	QuestionID     string  `gorm:"type:uuid;not null" json:"question_id"`
	SelectedOption string  `gorm:"type:varchar(50)" json:"selected_option,omitempty"`
	SelectedValue  *int    `gorm:"type:int;column:selected_value" json:"selected_value,omitempty"`
	OptionID       *string `gorm:"type:uuid" json:"option_id,omitempty"`
	IsCorrect      bool    `gorm:"type:boolean;not null" json:"is_correct"`

	// Relations
	Question *Question `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}

func (QuizAnswer) TableName() string { return "quiz_answers" }
