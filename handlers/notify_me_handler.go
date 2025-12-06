package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/errors"
	"github.com/Infinite-Locus-Product/thums_up_backend/services"
	"github.com/Infinite-Locus-Product/thums_up_backend/utils"
)

type NotifyMeHandler struct {
	notifyMeService     services.NotifyMeService
	notificationService services.NotificationService
}

func NewNotifyMeHandler(
	notifyMeService services.NotifyMeService,
	notificationService services.NotificationService,
) *NotifyMeHandler {
	return &NotifyMeHandler{
		notifyMeService:     notifyMeService,
		notificationService: notificationService,
	}
}

func (h *NotifyMeHandler) Subscribe(c *gin.Context) {
	var req dtos.NotifyMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := utils.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
		return
	}

	response, err := h.notifyMeService.Subscribe(c.Request.Context(), req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to subscribe")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to subscribe",
		})
		return
	}

	email := ""
	if req.Email != nil {
		email = *req.Email
	}

	if h.notificationService != nil {
		go h.notificationService.PublishNotifyMeMessage(c.Request.Context(), req.PhoneNumber, email)
	}

	c.JSON(http.StatusCreated, dtos.SuccessResponse{
		Success: true,
		Data:    response,
		Message: "Successfully subscribed to notifications",
	})
}

func (h *NotifyMeHandler) GetSubscription(c *gin.Context) {
	phoneNumber := c.Param("phone")
	if phoneNumber == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Phone number is required",
		})
		return
	}

	response, err := h.notifyMeService.GetSubscription(c.Request.Context(), phoneNumber)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to get subscription")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to get subscription",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    response,
	})
}

func (h *NotifyMeHandler) GetAllUnnotified(c *gin.Context) {
	responses, err := h.notifyMeService.GetAllUnnotified(c.Request.Context())
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to get unnotified subscriptions")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to get unnotified subscriptions",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    responses,
	})
}
