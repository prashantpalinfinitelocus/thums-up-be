package handlers

import (
	stderrors "errors"
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
//	@Description	After being selected as a winner, user submits their KYC details including name, email, optional Aadhar card images, and up to three cities for additional information. Aadhar card number and images are optional. Cities are optional text fields.
//	@Tags			Winners
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		Bearer
//	@Param			user_name		formData	string								true	"User name"
//	@Param			user_email		formData	string								true	"User email"
//	@Param			aadhar_number	formData	string								false	"Aadhar number (optional)"
//	@Param			aadhar_front	formData	file								false	"Aadhar front image (optional)"
//	@Param			aadhar_back		formData	file								false	"Aadhar back image (optional)"
//	@Param			city1			formData	string								false	"City 1 (optional)"
//	@Param			city2			formData	string								false	"City 2 (optional)"
//	@Param			city3			formData	string								false	"City 3 (optional)"
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

	const maxMultipartMemory = 150 * 1024 * 1024 // 150MB to accommodate large images + form fields
	if err := c.Request.ParseMultipartForm(maxMultipartMemory); err != nil {
		log.WithError(err).Warn("Failed to parse multipart form")
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to process form data",
		})
		return
	}

	userName := c.PostForm("user_name")
	userEmail := c.PostForm("user_email")

	if userName == "" || userEmail == "" {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "user_name and user_email are required",
		})
		return
	}

	var aadharFrontURL *string
	var aadharBackURL *string
	var aadharNumber *string

	aadharNumberStr := c.PostForm("aadhar_number")
	if aadharNumberStr != "" {
		aadharNumber = &aadharNumberStr
	}

	aadharFrontFile, err := c.FormFile("aadhar_front")
	if err == nil && aadharFrontFile != nil {
		if err := utils.ValidateImageFile(aadharFrontFile); err != nil {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Success: false,
				Error:   "Invalid aadhar_front file: " + err.Error(),
			})
			return
		}

		url, _, err := h.gcsService.UploadFile(c.Request.Context(), aadharFrontFile, "winners/kyc/aadhar")
		if err != nil {
			log.WithError(err).Error("Failed to upload aadhar front")
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Success: false,
				Error:   "Failed to upload aadhar front image",
			})
			return
		}
		aadharFrontURL = &url
	}

	aadharBackFile, err := c.FormFile("aadhar_back")
	if err == nil && aadharBackFile != nil {
		if err := utils.ValidateImageFile(aadharBackFile); err != nil {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Success: false,
				Error:   "Invalid aadhar_back file: " + err.Error(),
			})
			return
		}

		url, _, err := h.gcsService.UploadFile(c.Request.Context(), aadharBackFile, "winners/kyc/aadhar")
		if err != nil {
			log.WithError(err).Error("Failed to upload aadhar back")
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Success: false,
				Error:   "Failed to upload aadhar back image",
			})
			return
		}
		aadharBackURL = &url
	}

	city1 := c.PostForm("city1")
	city2 := c.PostForm("city2")
	city3 := c.PostForm("city3")

	req := dtos.WinnerKYCRequest{
		UserName:     userName,
		UserEmail:    userEmail,
		AadharNumber: aadharNumber,
		AadharFront:  aadharFrontURL,
		AadharBack:   aadharBackURL,
		City1:        city1,
		City2:        city2,
		City3:        city3,
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
//	@Description	Check if the authenticated user has won, whether they have viewed the congratulations banner, and if they have participated in the contest. Returns has_won (true if user is a winner), has_viewed (true if banner was viewed), has_participated (true if user has submitted any answers in thunder_seat), week_number (if won), and qr_code_url (if won).
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
