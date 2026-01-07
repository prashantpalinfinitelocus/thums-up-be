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

type AnswerQuestionsRequestDTO struct {
	QuestionID     int `json:"question_id"`
	AnswerID       int `json:"answer_id"`
	QuestionNumber int `json:"question_number"`
}

type OptionDTO struct {
	ID           int    `json:"id"`
	OptionText   string `json:"option_text"`
	DisplayOrder int    `json:"display_order"`
}

type QuestionResponseDTO struct {
	ID             int         `json:"id"`
	QuestionText   string      `json:"question_text"`
	LanguageID     int         `json:"language_id"`
	QuesPoint      int         `json:"ques_point"`
	Options        []OptionDTO `json:"options"`
	SelectedOption *int        `json:"selected_option,omitempty"`
}

type GetQuestionByTextRequestDTO struct {
	QuestionText string `json:"question_text" binding:"required"`
	LanguageID   int    `json:"language_id" binding:"required"`
}

type CreateOptionDTO struct {
	ID           *int   `json:"id"` // Nullable for creation
	OptionText   string `json:"option_text"`
	DisplayOrder int    `json:"display_order"`
	IsActive     *bool  `json:"is_active"`
}

type CreateQuestionDTO struct {
	ID           *int              `json:"id"`
	QuestionText string            `json:"question_text"`
	QuesPoint    int               `json:"ques_point"`
	LanguageID   int               `json:"language_id"`
	IsActive     *bool             `json:"is_active"`
	Options      []CreateOptionDTO `json:"options"`
}

type CreateQuestionsRequestDTO struct {
	Questions []CreateQuestionDTO `json:"questions"`
}
