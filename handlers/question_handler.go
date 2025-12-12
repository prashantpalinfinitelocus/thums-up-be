package handlers

import (
	stderrors "errors"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/errors"
	"github.com/Infinite-Locus-Product/thums_up_backend/services"
	"github.com/Infinite-Locus-Product/thums_up_backend/utils"
)

type QuestionHandler struct {
	questionService services.QuestionService
}

func NewQuestionHandler(questionService services.QuestionService) *QuestionHandler {
	return &QuestionHandler{
		questionService: questionService,
	}
}

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

