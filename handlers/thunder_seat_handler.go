package handlers

import (
	stderrors "errors"
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

// SubmitAnswer godoc
//
//	@Summary		Submit Thunder Seat answer
//	@Description	Submit an answer to a Thunder Seat question for the current week with optional media file (audio/video). Requires authentication.
//	@Tags			Thunder Seat
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		Bearer
//	@Param			description	formData	string												true	"Answer text"
//	@Param			social_media	formData	string												false	"Sharing platform (instagram, snapchat, facebook, twitter, tiktok, youtube)"
//	@Param			user_name	formData	string												false	"Platform user name (min 3, max 255 characters)"
//	@Param			media_file	formData	file												false	"Optional media file (audio/video, max 100MB)"
//	@Success		201			{object}	dtos.SuccessResponse{data=dtos.ThunderSeatResponse}	"Answer submitted successfully"
//	@Failure		400			{object}	dtos.ErrorResponse									"Validation failed"
//	@Failure		401			{object}	dtos.ErrorResponse									"Unauthorized"
//	@Failure		500			{object}	dtos.ErrorResponse									"Failed to submit answer"
//	@Router			/thunder-seat [post]
func (h *ThunderSeatHandler) SubmitAnswer(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Success: false, Error: errors.ErrUserNotAuthenticated})
		return
	}

	userEntity, ok := user.(*entities.User)
	if !ok || userEntity == nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Success: false, Error: errors.ErrInvalidUserContext})
		return
	}
	userID := userEntity.ID

	// Parse multipart form with large buffer (150MB) to handle large files
	const maxMultipartMemory = 150 * 1024 * 1024
	if err := utils.ParseMultipartFormWithLargeBuffer(c.Request, maxMultipartMemory); err != nil && err != http.ErrNotMultipart {
		log.WithFields(log.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to parse multipart form with large buffer")
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to process form data",
		})
		return
	}

	// Get media file using large buffer helper
	mediaFile, err := utils.GetFormFileWithLargeBuffer(c.Request, "media_file", maxMultipartMemory)
	if err != nil && err != http.ErrMissingFile {
		log.WithError(err).Error("Failed to get media file from request")
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to process media file",
		})
		return
	}

	if mediaFile != nil {
		if err := utils.ValidateMediaFile(mediaFile); err != nil {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
	}

	// Get form values using large buffer helper
	description, _ := utils.GetPostFormWithLargeBuffer(c.Request, "description", maxMultipartMemory)
	if description == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrValidationFailed,
			Details: map[string]string{"description": "description is required"},
		})
		return
	}

	var req dtos.ThunderSeatSubmitRequest
	req.Answer = description

	// Optional fields
	socialMedia, _ := utils.GetPostFormWithLargeBuffer(c.Request, "social_media", maxMultipartMemory)
	if socialMedia != "" {
		// Validate social media platform
		validPlatforms := []string{"instagram", "snapchat", "facebook", "twitter", "tiktok", "youtube"}
		isValid := false
		for _, platform := range validPlatforms {
			if socialMedia == platform {
				isValid = true
				break
			}
		}
		if !isValid {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Success: false,
				Error:   errors.ErrValidationFailed,
				Details: map[string]string{"social_media": "social_media must be one of: instagram, snapchat, facebook, twitter, tiktok, youtube"},
			})
			return
		}
		req.SharingPlatform = &socialMedia
	}

	userName, _ := utils.GetPostFormWithLargeBuffer(c.Request, "user_name", maxMultipartMemory)
	if userName != "" {
		if len(userName) < 3 || len(userName) > 255 {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Success: false,
				Error:   errors.ErrValidationFailed,
				Details: map[string]string{"user_name": "user_name must be between 3 and 255 characters"},
			})
			return
		}
		req.PlatformUserName = &userName
	}

	response, err := h.thunderSeatService.SubmitAnswer(c.Request.Context(), req, userID, mediaFile)
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			log.WithFields(log.Fields{
				"user_id":     userID,
				"status_code": appErr.StatusCode,
				"error":       appErr.Message,
				"has_media":   mediaFile != nil,
			}).WithError(appErr.Err).Warn("Thunder seat answer submission failed with application error")
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithFields(log.Fields{
			"user_id":   userID,
			"has_media": mediaFile != nil,
		}).WithError(err).Error("Failed to submit thunder seat answer - unexpected error type")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrAnswerSubmitFailed,
		})
		return
	}

	c.JSON(http.StatusCreated, dtos.SuccessResponse{
		Success: true,
		Data:    response,
		Message: "Answer submitted successfully",
	})
}

// GetUserSubmissions godoc
//
//	@Summary		Get user's Thunder Seat submissions
//	@Description	Retrieve all Thunder Seat answer submissions for the authenticated user. Requires authentication.
//	@Tags			Thunder Seat
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Success		200	{object}	dtos.SuccessResponse{data=[]dtos.ThunderSeatResponse}	"Submissions retrieved successfully"
//	@Failure		401	{object}	dtos.ErrorResponse										"Unauthorized"
//	@Failure		500	{object}	dtos.ErrorResponse										"Failed to get submissions"
//	@Router			/thunder-seat/submissions [get]
func (h *ThunderSeatHandler) GetUserSubmissions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrUserNotAuthenticated,
		})
		return
	}

	responses, err := h.thunderSeatService.GetUserSubmissions(c.Request.Context(), userID.(string))
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to get user submissions")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrSubmissionsFetchFailed,
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    responses,
	})
}

// GetCurrentWeek godoc
//
//	@Summary		Get current contest week information
//	@Description	Retrieve the current active contest week details including dates and winner count
//	@Tags			Thunder Seat
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dtos.SuccessResponse{data=dtos.CurrentWeekResponse}	"Current week retrieved successfully"
//	@Failure		404	{object}	dtos.ErrorResponse									"No active contest week"
//	@Failure		500	{object}	dtos.ErrorResponse									"Failed to get active contest week"
//	@Router			/thunder-seat/current-week [get]
func (h *ThunderSeatHandler) GetCurrentWeek(c *gin.Context) {
	response, err := h.thunderSeatService.GetCurrentWeek(c.Request.Context())
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to get active contest week")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to get active contest week",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    response,
	})
}
