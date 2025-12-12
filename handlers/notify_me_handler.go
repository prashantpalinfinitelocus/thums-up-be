package handlers

import (
	"context"
	stderrors "errors"
	"net/http"
	"time"

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
			Error:   errors.ErrValidationFailed,
			Details: validationErrors,
		})
		return
	}

	response, created, err := h.notifyMeService.Subscribe(c.Request.Context(), req)
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to subscribe")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrSubscriptionFailed,
		})
		return
	}

	if created && h.notificationService != nil {
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		go func() {
			defer cancel()
			if err := h.notificationService.PublishNotifyMeMessage(bgCtx, req.PhoneNumber, req.Email); err != nil {
				log.WithError(err).Error("Failed to publish notify me message")
			}
		}()
	}

	statusCode := http.StatusCreated
	message := "Successfully subscribed to notifications"
	if !created {
		statusCode = http.StatusOK
		message = "Phone number is already subscribed to notifications"
	}

	c.JSON(statusCode, dtos.SuccessResponse{
		Success: true,
		Data:    response,
		Message: message,
	})
}

func (h *NotifyMeHandler) GetSubscription(c *gin.Context) {
	phoneNumber := c.Param("phone")
	if phoneNumber == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrPhoneNumberRequired,
		})
		return
	}

	response, err := h.notifyMeService.GetSubscription(c.Request.Context(), phoneNumber)
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to get subscription")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrSubscriptionNotFound,
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
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to get unnotified subscriptions")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrUnnotifiedFetchFailed,
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    responses,
	})
}
