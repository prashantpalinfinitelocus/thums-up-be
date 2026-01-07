package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/services"
)

type StateHandler struct {
	stateService services.StateService
}

func NewStateHandler(stateService services.StateService) *StateHandler {
	return &StateHandler{
		stateService: stateService,
	}
}

// GetAllStates godoc
//
//	@Summary		Get all states
//	@Description	Retrieve all active states. Returns a list of all states where is_active is true and is_deleted is false, ordered by name alphabetically.
//	@Tags			Location
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dtos.SuccessResponse	"States retrieved successfully"
//	@Failure		500	{object}	dtos.ErrorResponse	"Failed to fetch states"
//	@Router			/states [get]
func (h *StateHandler) GetAllStates(c *gin.Context) {
	states, err := h.stateService.GetAllStates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   "Failed to fetch states",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    states,
	})
}

