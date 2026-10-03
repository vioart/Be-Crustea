package model

import "time"

type FAQCategory string

const (
	FAQCategoryProduct     FAQCategory = "product"
	FAQCategoryService     FAQCategory = "service"
	FAQCategoryPartnership FAQCategory = "partnership"
)

type FAQ struct {
	ID          int64       `json:"id"`
	Category    FAQCategory `json:"category"`
	QuestionIDN string      `json:"question_idn"`
	QuestionEN  string      `json:"question_en"`
	AnswerIDN   string      `json:"answer_idn"`
	AnswerEN    string      `json:"answer_en"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
