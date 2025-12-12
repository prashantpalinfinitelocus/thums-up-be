package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"github.com/Infinite-Locus-Product/thums_up_backend/errors"
	"github.com/Infinite-Locus-Product/thums_up_backend/services"
	"github.com/Infinite-Locus-Product/thums_up_backend/utils"
)

type ThunderSeatHandler struct {
	thunderSeatService services.ThunderSeatService
}

func NewThunderSeatHandler(thunderSeatService services.ThunderSeatService) *ThunderSeatHandler {
	return &ThunderSeatHandler{
		thunderSeatService: thunderSeatService,
	}
}

func (h *ThunderSeatHandler) SubmitAnswer(c *gin.Context) {
	user, _ := c.Get("user")
	userID := user.(*entities.User).ID

	var req dtos.ThunderSeatSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := utils.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
		return
	}

	response, err := h.thunderSeatService.SubmitAnswer(c.Request.Context(), req, userID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to submit thunder seat answer")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to submit answer",
		})
		return
	}

	c.JSON(http.StatusCreated, dtos.SuccessResponse{
		Success: true,
		Data:    response,
		Message: "Answer submitted successfully",
	})
}

func (h *ThunderSeatHandler) GetUserSubmissions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	responses, err := h.thunderSeatService.GetUserSubmissions(c.Request.Context(), userID.(string))
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to get user submissions")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to get user submissions",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    responses,
	})
}

func (h *ThunderSeatHandler) GetCurrentWeek(c *gin.Context) {
	response, err := h.thunderSeatService.GetCurrentWeek(c.Request.Context())
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to get current week")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to get current week",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    response,
	})
}
