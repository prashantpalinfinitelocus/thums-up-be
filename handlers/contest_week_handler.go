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

type ContestWeekHandler struct {
	contestWeekService services.ContestWeekService
}

func NewContestWeekHandler(contestWeekService services.ContestWeekService) *ContestWeekHandler {
	return &ContestWeekHandler{
		contestWeekService: contestWeekService,
	}
}

// CreateContestWeek godoc
//
//	@Summary		Create a new contest week
//	@Description	Create a new contest week with week number, start date, end date, and winner count. Requires authentication.
//	@Tags			Contest Weeks
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Param			request	body		dtos.ContestWeekRequest								true	"Contest week details"
//	@Success		201		{object}	dtos.SuccessResponse{data=dtos.ContestWeekResponse}	"Contest week created successfully"
//	@Failure		400		{object}	dtos.ErrorResponse									"Validation failed"
//	@Failure		401		{object}	dtos.ErrorResponse									"Unauthorized"
//	@Failure		500		{object}	dtos.ErrorResponse									"Failed to create contest week"
//	@Router			/contest-weeks [post]
func (h *ContestWeekHandler) CreateContestWeek(c *gin.Context) {
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

	var req dtos.ContestWeekRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := utils.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrValidationFailed,
			Details: validationErrors,
		})
		return
	}

	response, err := h.contestWeekService.CreateContestWeek(c.Request.Context(), req, userEntity.ID)
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to create contest week")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to create contest week",
		})
		return
	}

	c.JSON(http.StatusCreated, dtos.SuccessResponse{
		Success: true,
		Data:    response,
		Message: "Contest week created successfully",
	})
}

// GetAllContestWeeks godoc
//
//	@Summary		Get all contest weeks
//	@Description	Retrieve a list of all contest weeks
//	@Tags			Contest Weeks
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dtos.SuccessResponse{data=[]dtos.ContestWeekResponse}	"Contest weeks retrieved successfully"
//	@Failure		500	{object}	dtos.ErrorResponse										"Failed to get contest weeks"
//	@Router			/contest-weeks [get]
func (h *ContestWeekHandler) GetAllContestWeeks(c *gin.Context) {
	responses, err := h.contestWeekService.GetAllContestWeeks(c.Request.Context())
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to get contest weeks")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to get contest weeks",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    responses,
	})
}

// GetContestWeekByNumber godoc
//
//	@Summary		Get contest week by week number
//	@Description	Retrieve a specific contest week by its week number
//	@Tags			Contest Weeks
//	@Accept			json
//	@Produce		json
//	@Param			weekNumber	path		int													true	"Week number"
//	@Success		200			{object}	dtos.SuccessResponse{data=dtos.ContestWeekResponse}	"Contest week retrieved successfully"
//	@Failure		400			{object}	dtos.ErrorResponse									"Invalid week number"
//	@Failure		404			{object}	dtos.ErrorResponse									"Contest week not found"
//	@Failure		500			{object}	dtos.ErrorResponse									"Failed to get contest week"
//	@Router			/contest-weeks/{weekNumber} [get]
func (h *ContestWeekHandler) GetContestWeekByNumber(c *gin.Context) {
	weekNumberStr := c.Param("weekNumber")
	weekNumber, err := strconv.Atoi(weekNumberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Invalid week number",
		})
		return
	}

	response, err := h.contestWeekService.GetContestWeekByNumber(c.Request.Context(), weekNumber)
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to get contest week")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to get contest week",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    response,
	})
}

// ActivateWeek godoc
//
//	@Summary		Activate a contest week
//	@Description	Activate a specific contest week by week number. Only one week can be active at a time. Requires authentication.
//	@Tags			Contest Weeks
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Param			request	body		dtos.ActivateWeekRequest							true	"Week number to activate"
//	@Success		200		{object}	dtos.SuccessResponse{data=dtos.ContestWeekResponse}	"Contest week activated successfully"
//	@Failure		400		{object}	dtos.ErrorResponse									"Validation failed"
//	@Failure		401		{object}	dtos.ErrorResponse									"Unauthorized"
//	@Failure		404		{object}	dtos.ErrorResponse									"Contest week not found"
//	@Failure		500		{object}	dtos.ErrorResponse									"Failed to activate contest week"
//	@Router			/contest-weeks/activate [post]
func (h *ContestWeekHandler) ActivateWeek(c *gin.Context) {
	var req dtos.ActivateWeekRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := utils.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrValidationFailed,
			Details: validationErrors,
		})
		return
	}

	response, err := h.contestWeekService.ActivateWeek(c.Request.Context(), req.WeekNumber)
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to activate contest week")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to activate contest week",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    response,
		Message: "Contest week activated successfully",
	})
}

// GetActiveWeek godoc
//
//	@Summary		Get active contest week
//	@Description	Retrieve the currently active contest week
//	@Tags			Contest Weeks
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dtos.SuccessResponse{data=dtos.ContestWeekResponse}	"Active week retrieved successfully"
//	@Failure		404	{object}	dtos.ErrorResponse									"No active contest week found"
//	@Failure		500	{object}	dtos.ErrorResponse									"Failed to get active week"
//	@Router			/contest-weeks/active [get]
func (h *ContestWeekHandler) GetActiveWeek(c *gin.Context) {
	response, err := h.contestWeekService.GetActiveWeek(c.Request.Context())
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to get active week")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to get active week",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    response,
	})
}
