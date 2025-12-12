package handlers

import (
	"context"
	stderrors "errors"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/errors"
	"github.com/Infinite-Locus-Product/thums_up_backend/pkg/queue"
	"github.com/Infinite-Locus-Product/thums_up_backend/services"
	"github.com/Infinite-Locus-Product/thums_up_backend/utils"
)

type NotifyMeHandler struct {
	notifyMeService     services.NotifyMeService
	notificationService services.NotificationService
	workerPool          *queue.WorkerPool
}

func NewNotifyMeHandler(
	notifyMeService services.NotifyMeService,
	notificationService services.NotificationService,
	workerPool *queue.WorkerPool,
) *NotifyMeHandler {
	return &NotifyMeHandler{
		notifyMeService:     notifyMeService,
		notificationService: notificationService,
		workerPool:          workerPool,
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

	// Submit background task to worker pool instead of spawning goroutine
	if created && h.notificationService != nil && h.workerPool != nil {
		phoneNumber := req.PhoneNumber
		email := req.Email

		task := func(ctx context.Context) error {
			if err := h.notificationService.PublishNotifyMeMessage(ctx, phoneNumber, email); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"phone_number": phoneNumber,
					"email":        email,
				}).Error("Failed to publish notify me message")
				return err
			}
			log.WithFields(log.Fields{
				"phone_number": phoneNumber,
			}).Info("Successfully published notify me message")
			return nil
		}

		if err := h.workerPool.Submit(task); err != nil {
			log.WithError(err).Warn("Failed to submit notification task to worker pool")
			// Don't fail the request if background task submission fails
		}
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
