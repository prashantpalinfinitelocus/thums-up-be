package dtos

type QuestionSubmitRequest struct {
	QuestionText string `json:"question_text" binding:"required"`
	LanguageID   int    `json:"language_id" binding:"required"`
}

type QuestionResponse struct {
	ID           int    `json:"id"`
	QuestionText string `json:"question_text"`
	LanguageID   int    `json:"language_id"`
	IsActive     bool   `json:"is_active"`
}
