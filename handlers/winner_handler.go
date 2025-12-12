package handlers

import (
	stderrors "errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
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
