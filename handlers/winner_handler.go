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
}

func NewWinnerHandler(winnerService services.WinnerService) *WinnerHandler {
	return &WinnerHandler{
		winnerService: winnerService,
	}
}

// SelectWinners godoc
// @Summary Select winners for a week
// @Description Admin endpoint to select winners for a specific contest week. Requires API key authentication.
// @Tags Admin
// @Accept json
// @Produce json
// @Security APIKey
// @Param request body dtos.SelectWinnersRequest true "Week number"
// @Success 201 {object} dtos.SuccessResponse{data=[]dtos.WinnerResponse} "Winners selected successfully"
// @Failure 400 {object} dtos.ErrorResponse "Validation failed"
// @Failure 401 {object} dtos.ErrorResponse "Unauthorized"
// @Failure 500 {object} dtos.ErrorResponse "Failed to select winners"
// @Router /admin/winners/select [post]
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
// @Summary Get winners by week
// @Description Retrieve all winners for a specific contest week
// @Tags Winners
// @Accept json
// @Produce json
// @Param weekNumber path int true "Week number"
// @Success 200 {object} dtos.SuccessResponse{data=[]dtos.WinnerResponse} "Winners retrieved successfully"
// @Failure 400 {object} dtos.ErrorResponse "Invalid week number"
// @Failure 500 {object} dtos.ErrorResponse "Failed to get winners"
// @Router /winners/week/{weekNumber} [get]
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
// @Summary Get all winners with pagination
// @Description Retrieve all winners with pagination support
// @Tags Winners
// @Accept json
// @Produce json
// @Param limit query int true "Number of items per page" minimum(1) maximum(100)
// @Param offset query int false "Number of items to skip" minimum(0) default(0)
// @Success 200 {object} dtos.PaginatedResponse{data=[]dtos.WinnerResponse} "Winners retrieved successfully"
// @Failure 400 {object} dtos.ErrorResponse "Validation failed"
// @Failure 500 {object} dtos.ErrorResponse "Failed to get all winners"
// @Router /winners [get]
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
// @Summary Submit winner KYC details
// @Description After being selected as a winner, user submits their details and friends' information.
// @Tags Winners
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dtos.WinnerKYCRequest true "Winner KYC details"
// @Success 200 {object} dtos.SuccessResponse{data=string} "KYC submitted successfully"
// @Failure 400 {object} dtos.ErrorResponse "Validation failed"
// @Failure 401 {object} dtos.ErrorResponse "Unauthorized"
// @Failure 500 {object} dtos.ErrorResponse "Failed to submit KYC"
// @Router /winners/kyc [post]
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

	var req dtos.WinnerKYCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := utils.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
		return
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

