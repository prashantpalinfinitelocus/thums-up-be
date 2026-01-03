package handlers

import (
	"encoding/json"
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

type WinnerHandler struct {
	winnerService services.WinnerService
	gcsService    utils.GCSService
}

func NewWinnerHandler(winnerService services.WinnerService, gcsService utils.GCSService) *WinnerHandler {
	return &WinnerHandler{
		winnerService: winnerService,
		gcsService:    gcsService,
	}
}

// SelectWinners godoc
//
//	@Summary		Select winners for a week
//	@Description	Admin endpoint to select winners for a specific contest week. Requires API key authentication.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Security		APIKey
//	@Param			request	body		dtos.SelectWinnersRequest							true	"Week number"
//	@Success		201		{object}	dtos.SuccessResponse{data=[]dtos.WinnerResponse}	"Winners selected successfully"
//	@Failure		400		{object}	dtos.ErrorResponse									"Validation failed"
//	@Failure		401		{object}	dtos.ErrorResponse									"Unauthorized"
//	@Failure		500		{object}	dtos.ErrorResponse									"Failed to select winners"
//	@Router			/admin/winners/select [post]
func (h *WinnerHandler) SelectWinners(c *gin.Context) {
	var req dtos.SelectWinnersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := utils.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
		return
	}

	responses, err := h.winnerService.SelectWinners(c.Request.Context(), req)
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to select winners")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to select winners",
		})
		return
	}

	c.JSON(http.StatusCreated, dtos.SuccessResponse{
		Success: true,
		Data:    responses,
		Message: "Winners selected successfully",
	})
}

// GetWinnersByWeek godoc
//
//	@Summary		Get winners by week
//	@Description	Retrieve all winners for a specific contest week
//	@Tags			Winners
//	@Accept			json
//	@Produce		json
//	@Param			weekNumber	path		int													true	"Week number"
//	@Success		200			{object}	dtos.SuccessResponse{data=[]dtos.WinnerResponse}	"Winners retrieved successfully"
//	@Failure		400			{object}	dtos.ErrorResponse									"Invalid week number"
//	@Failure		500			{object}	dtos.ErrorResponse									"Failed to get winners"
//	@Router			/winners/week/{weekNumber} [get]
func (h *WinnerHandler) GetWinnersByWeek(c *gin.Context) {
	weekNumberStr := c.Param("weekNumber")
	weekNumber, err := strconv.Atoi(weekNumberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Invalid week number",
		})
		return
	}

	responses, err := h.winnerService.GetWinnersByWeek(c.Request.Context(), weekNumber)
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to get winners")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to get winners",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    responses,
	})
}

// GetAllWinners godoc
//
//	@Summary		Get all winners with pagination
//	@Description	Retrieve all winners with pagination support
//	@Tags			Winners
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int													true	"Number of items per page"	minimum(1)	maximum(100)
//	@Param			offset	query		int													false	"Number of items to skip"	minimum(0)	default(0)
//	@Success		200		{object}	dtos.PaginatedResponse{data=[]dtos.WinnerResponse}	"Winners retrieved successfully"
//	@Failure		400		{object}	dtos.ErrorResponse									"Validation failed"
//	@Failure		500		{object}	dtos.ErrorResponse									"Failed to get all winners"
//	@Router			/winners [get]
func (h *WinnerHandler) GetAllWinners(c *gin.Context) {
	var req dtos.AllWinnersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		validationErrors := utils.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
		return
	}

	responses, total, err := h.winnerService.GetAllWinners(c.Request.Context(), req.Limit, req.Offset)
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to get all winners")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to get all winners",
		})
		return
	}

	totalPages := int(total) / req.Limit
	if int(total)%req.Limit != 0 {
		totalPages++
	}
	currentPage := (req.Offset / req.Limit) + 1

	c.JSON(http.StatusOK, dtos.PaginatedResponse{
		Success: true,
		Data:    responses,
		Meta: dtos.PaginationMeta{
			Page:       currentPage,
			PageSize:   req.Limit,
			TotalPages: totalPages,
			TotalCount: total,
		},
	})
}

// SubmitWinnerKYC godoc
//
//	@Summary		Submit winner KYC details
//	@Description	After being selected as a winner, user submits their details and friends' information.
//	@Tags			Winners
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		Bearer
//	@Param			user_name		formData	string								true	"User name"
//	@Param			user_email		formData	string								true	"User email"
//	@Param			aadhar_number	formData	string								true	"Aadhar number"
//	@Param			aadhar_front	formData	file								true	"Aadhar front image"
//	@Param			aadhar_back		formData	file								true	"Aadhar back image"
//	@Param			friends			formData	string								false	"Friends JSON array"
//	@Success		200				{object}	dtos.SuccessResponse{data=string}	"KYC submitted successfully"
//	@Failure		400				{object}	dtos.ErrorResponse					"Validation failed"
//	@Failure		401				{object}	dtos.ErrorResponse					"Unauthorized"
//	@Failure		500				{object}	dtos.ErrorResponse					"Failed to submit KYC"
//	@Router			/winners/kyc [post]
func (h *WinnerHandler) SubmitWinnerKYC(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrUserNotAuthenticated,
		})
		return
	}

	userEntity, ok := user.(*entities.User)
	if !ok || userEntity == nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrInvalidUserContext,
		})
		return
	}

	userName := c.PostForm("user_name")
	userEmail := c.PostForm("user_email")
	aadharNumber := c.PostForm("aadhar_number")

	if userName == "" || userEmail == "" || aadharNumber == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "user_name, user_email, and aadhar_number are required",
		})
		return
	}

	aadharFrontFile, err := c.FormFile("aadhar_front")
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "aadhar_front file is required",
		})
		return
	}

	aadharBackFile, err := c.FormFile("aadhar_back")
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "aadhar_back file is required",
		})
		return
	}

	if err := utils.ValidateImageFile(aadharFrontFile); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Invalid aadhar_front file: " + err.Error(),
		})
		return
	}

	if err := utils.ValidateImageFile(aadharBackFile); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Invalid aadhar_back file: " + err.Error(),
		})
		return
	}

	aadharFrontURL, _, err := h.gcsService.UploadFile(c.Request.Context(), aadharFrontFile, "winners/kyc/aadhar")
	if err != nil {
		log.WithError(err).Error("Failed to upload aadhar front")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to upload aadhar front image",
		})
		return
	}

	aadharBackURL, _, err := h.gcsService.UploadFile(c.Request.Context(), aadharBackFile, "winners/kyc/aadhar")
	if err != nil {
		log.WithError(err).Error("Failed to upload aadhar back")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to upload aadhar back image",
		})
		return
	}

	var friends []dtos.WinnerFriendDTO
	friendsJSON := c.PostForm("friends")
	if friendsJSON != "" {
		if err := json.Unmarshal([]byte(friendsJSON), &friends); err != nil {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Success: false,
				Error:   "Invalid friends JSON format",
			})
			return
		}
		if len(friends) > 10 {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Success: false,
				Error:   "Maximum 10 friends allowed",
			})
			return
		}

		for i := range friends {
			friendFrontKey := fmt.Sprintf("friend_%d_aadhar_front", i)
			friendBackKey := fmt.Sprintf("friend_%d_aadhar_back", i)

			friendFrontFile, err := c.FormFile(friendFrontKey)
			if err != nil {
				c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
					Success: false,
					Error:   fmt.Sprintf("friend_%d_aadhar_front file is required", i),
				})
				return
			}

			if err := utils.ValidateImageFile(friendFrontFile); err != nil {
				c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
					Success: false,
					Error:   fmt.Sprintf("Invalid friend %d aadhar_front file: %s", i, err.Error()),
				})
				return
			}

			friendFrontURL, _, err := h.gcsService.UploadFile(c.Request.Context(), friendFrontFile, "winners/kyc/friends/aadhar")
			if err != nil {
				log.WithError(err).Errorf("Failed to upload friend %d aadhar front", i)
				c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
					Success: false,
					Error:   fmt.Sprintf("Failed to upload friend %d aadhar front image", i),
				})
				return
			}
			friends[i].AadharFront = friendFrontURL

			friendBackFile, err := c.FormFile(friendBackKey)
			if err != nil {
				c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
					Success: false,
					Error:   fmt.Sprintf("friend_%d_aadhar_back file is required", i),
				})
				return
			}

			if err := utils.ValidateImageFile(friendBackFile); err != nil {
				c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
					Success: false,
					Error:   fmt.Sprintf("Invalid friend %d aadhar_back file: %s", i, err.Error()),
				})
				return
			}

			friendBackURL, _, err := h.gcsService.UploadFile(c.Request.Context(), friendBackFile, "winners/kyc/friends/aadhar")
			if err != nil {
				log.WithError(err).Errorf("Failed to upload friend %d aadhar back", i)
				c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
					Success: false,
					Error:   fmt.Sprintf("Failed to upload friend %d aadhar back image", i),
				})
				return
			}
			friends[i].AadharBack = friendBackURL
		}
	}

	req := dtos.WinnerKYCRequest{
		UserName:     userName,
		UserEmail:    userEmail,
		AadharNumber: aadharNumber,
		AadharFront:  aadharFrontURL,
		AadharBack:   aadharBackURL,
		Friends:      friends,
	}

	if err := h.winnerService.SubmitWinnerKYC(c.Request.Context(), userEntity.ID, req); err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}

		log.WithError(err).Error("Failed to submit winner KYC")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to submit KYC",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    "KYC submitted successfully",
	})
}

// CheckWinnerStatus godoc
//
//	@Summary		Check user winner status
//	@Description	Check if the authenticated user has won and whether they have viewed the congratulations banner
//	@Tags			Winners
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Success		200	{object}	dtos.SuccessResponse{data=dtos.WinnerStatusResponse}	"Winner status retrieved successfully"
//	@Failure		401	{object}	dtos.ErrorResponse										"Unauthorized"
//	@Failure		500	{object}	dtos.ErrorResponse										"Failed to check winner status"
//	@Router			/winners/status [get]
func (h *WinnerHandler) CheckWinnerStatus(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrUserNotAuthenticated,
		})
		return
	}

	userEntity, ok := user.(*entities.User)
	if !ok || userEntity == nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrInvalidUserContext,
		})
		return
	}

	status, err := h.winnerService.CheckUserWinnerStatus(c.Request.Context(), userEntity.ID)
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to check winner status")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to check winner status",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    status,
	})
}

// MarkBannerAsViewed godoc
//
//	@Summary		Mark banner as viewed
//	@Description	Mark the congratulations banner as viewed for the authenticated user
//	@Tags			Winners
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Success		200	{object}	dtos.SuccessResponse{data=string}	"Banner marked as viewed successfully"
//	@Failure		401	{object}	dtos.ErrorResponse					"Unauthorized"
//	@Failure		404	{object}	dtos.ErrorResponse					"User is not a winner"
//	@Failure		500	{object}	dtos.ErrorResponse					"Failed to mark banner as viewed"
//	@Router			/winners/mark-viewed [post]
func (h *WinnerHandler) MarkBannerAsViewed(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrUserNotAuthenticated,
		})
		return
	}

	userEntity, ok := user.(*entities.User)
	if !ok || userEntity == nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrInvalidUserContext,
		})
		return
	}

	err := h.winnerService.MarkBannerAsViewed(c.Request.Context(), userEntity.ID)
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to mark banner as viewed")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to mark banner as viewed",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    "Banner marked as viewed",
	})
}
