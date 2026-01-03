package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/services"
)

type WebsiteStatusHandler struct {
	websiteStatusService services.WebsiteStatusService
}

func NewWebsiteStatusHandler(websiteStatusService services.WebsiteStatusService) *WebsiteStatusHandler {
	return &WebsiteStatusHandler{
		websiteStatusService: websiteStatusService,
	}
}

// GetStatus godoc
//
//	@Summary		Get website status
//	@Description	Get the current website status (live, live_soon, or coming_soon) based on launch date
//	@Tags			Website
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dtos.SuccessResponse{data=dtos.WebsiteStatusResponse}	"Website status retrieved successfully"
//	@Failure		500	{object}	dtos.ErrorResponse										"Failed to get website status"
//	@Router			/website-status [get]
func (h *WebsiteStatusHandler) GetStatus(c *gin.Context) {
	status := h.websiteStatusService.GetStatus()

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    status,
	})
}
