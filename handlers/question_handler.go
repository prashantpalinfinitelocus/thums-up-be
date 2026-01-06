package handlers

import (
	stderrors "errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"github.com/Infinite-Locus-Product/thums_up_backend/errors"
	"github.com/Infinite-Locus-Product/thums_up_backend/services"
	"github.com/Infinite-Locus-Product/thums_up_backend/utils"
)

type QuestionHandler struct {
	questionService services.QuestionService
	userService     services.UserService
}

func NewQuestionHandler(questionService services.QuestionService, userService services.UserService) *QuestionHandler {
	return &QuestionHandler{
		questionService: questionService,
		userService:     userService,
	}
}

// SubmitQuestion godoc
//
//	@Summary		Submit a new question
//	@Description	Submit a new question with text and language. Requires authentication.
//	@Tags			Questions
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Param			request	body		dtos.QuestionSubmitRequest							true	"Question text and language ID"
//	@Success		201		{object}	dtos.SuccessResponse{data=dtos.QuestionResponse}	"Question submitted successfully"
//	@Failure		400		{object}	dtos.ErrorResponse									"Validation failed"
//	@Failure		401		{object}	dtos.ErrorResponse									"Unauthorized"
//	@Failure		500		{object}	dtos.ErrorResponse									"Failed to submit question"
//	@Router			/questions [post]
func (h *QuestionHandler) SubmitQuestion(c *gin.Context) {
	var req dtos.QuestionSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := utils.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
		return
	}
    //fixed the question
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	response, err := h.questionService.SubmitQuestion(c.Request.Context(), req, userID.(string))
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to submit question")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to submit question",
		})
		return
	}

	c.JSON(http.StatusCreated, dtos.SuccessResponse{
		Success: true,
		Data:    response,
		Message: "Question submitted successfully",
	})
}

// GetActiveQuestions godoc
//
//	@Summary		Get active questions
//	@Description	Retrieve all active questions
//	@Tags			Questions
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dtos.SuccessResponse{data=[]dtos.QuestionResponse}	"Questions retrieved successfully"
//	@Failure		500	{object}	dtos.ErrorResponse									"Failed to get active questions"
//	@Router			/questions/active [get]
func (h *QuestionHandler) GetActiveQuestions(c *gin.Context) {
	responses, err := h.questionService.GetActiveQuestions(c.Request.Context())
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to get active questions")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to get active questions",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    responses,
	})
}

// GetQuestions godoc
//
//	@Summary		Get Questions
//	@Description	Get all active questions with user's answers
//	@Tags			Questions
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Param			language_id	query		int														true	"Language ID"
//	@Success		200			{object}	dtos.SuccessResponse{data=[]dtos.QuestionResponseDTO}	"Questions retrieved successfully"
//	@Failure		400			{object}	dtos.ErrorResponse										"Invalid request"
//	@Failure		401			{object}	dtos.ErrorResponse										"Unauthorized"
//	@Failure		500			{object}	dtos.ErrorResponse										"Failed to fetch questions"
//	@Router			/profile/questions [get]
func (h *QuestionHandler) GetQuestions(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	userEntity := user.(*entities.User)

	languageIDStr, exists := c.GetQuery("language_id")
	if !exists {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Language ID is required",
		})
		return
	}

	languageID, err := strconv.Atoi(languageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Invalid language_id",
		})
		return
	}

	questions, err := h.userService.GetQuestions(c.Request.Context(), userEntity.ID, languageID)
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to fetch questions")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch questions",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    questions,
	})
}

// AnswerQuestions godoc
//
//	@Summary		Answer Questions
//	@Description	Submit answers to multiple questions
//	@Tags			Questions
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Param			request	body		[]dtos.AnswerQuestionsRequestDTO	true	"Answer Questions Request"
//	@Success		200		{object}	dtos.SuccessResponse{data=string}	"Answers submitted successfully"
//	@Failure		400		{object}	dtos.ErrorResponse					"Invalid request"
//	@Failure		401		{object}	dtos.ErrorResponse					"Unauthorized"
//	@Failure		500		{object}	dtos.ErrorResponse					"Failed to submit answers"
//	@Router			/profile/questions [post]
func (h *QuestionHandler) AnswerQuestions(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	userEntity := user.(*entities.User)

	var answers []dtos.AnswerQuestionsRequestDTO
	if err := c.ShouldBindJSON(&answers); err != nil {
		validationErrors := utils.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
		return
	}

	// Validate that at least one answer is provided
	if len(answers) == 0 {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "At least one answer must be provided",
		})
		return
	}

	// Validate each answer
	for _, answer := range answers {
		if answer.QuestionID == 0 {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Success: false,
				Error:   "Question ID is required",
			})
			return
		}
		if answer.AnswerID == 0 {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Success: false,
				Error:   "Answer ID is required",
			})
			return
		}
	}

	err := h.userService.AnswerQuestions(c.Request.Context(), userEntity.ID, answers)
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		if err.Error() == "question not found" {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{
				Success: false,
				Error:   "Question not found",
			})
			return
		}
		log.WithError(err).Error("Failed to submit answers")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to submit answers",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    "Answers submitted successfully",
	})
}

// GetQuestionByID godoc
//
//	@Summary		Get Question By Text
//	@Description	Get a question by its text and language ID
//	@Tags			Questions
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Param			request	body		dtos.GetQuestionByTextRequestDTO					true	"Get Question By Text Request"
//	@Success		200		{object}	dtos.SuccessResponse{data=dtos.QuestionResponseDTO}	"Question retrieved successfully"
//	@Failure		400		{object}	dtos.ErrorResponse									"Invalid request body"
//	@Failure		404		{object}	dtos.ErrorResponse									"Question not found"
//	@Failure		500		{object}	dtos.ErrorResponse									"Internal server error"
//	@Router			/profile/questions/text [post]
func (h *QuestionHandler) GetQuestionByID(c *gin.Context) {
	var req dtos.GetQuestionByTextRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := utils.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
		return
	}

	questionID, err := h.userService.GetQuestionIDByText(c.Request.Context(), req.QuestionText, req.LanguageID)
	if err != nil {
		if err.Error() == "question not found" {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{
				Success: false,
				Error:   "Question not found",
			})
			return
		}
		log.WithError(err).Error("Failed to find question")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to find question",
		})
		return
	}

	question, err := h.userService.GetQuestionByID(c.Request.Context(), questionID, req.LanguageID)
	if err != nil {
		if err.Error() == "question not found" {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{
				Success: false,
				Error:   "Question not found",
			})
			return
		}
		log.WithError(err).Error("Failed to fetch question")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to fetch question %d", questionID),
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    question,
	})
}

// CreateQuestions godoc
//
//	@Summary		Create Questions
//	@Description	Create or update questions and options
//	@Tags			Questions
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Param			request	body		dtos.CreateQuestionsRequestDTO		true	"Create Questions Request"
//	@Success		200		{object}	dtos.SuccessResponse{data=string}	"Questions created successfully"
//	@Failure		400		{object}	dtos.ErrorResponse					"Invalid request"
//	@Failure		401		{object}	dtos.ErrorResponse					"Unauthorized"
//	@Failure		500		{object}	dtos.ErrorResponse					"Failed to create questions"
//	@Router			/profile/questions/create [post]
func (h *QuestionHandler) CreateQuestions(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	userEntity := user.(*entities.User)

	var req dtos.CreateQuestionsRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := utils.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
		return
	}

	if err := h.questionService.CreateQuestions(c.Request.Context(), userEntity.ID, req); err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to create questions")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to create questions",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    "Questions created successfully",
	})
}
