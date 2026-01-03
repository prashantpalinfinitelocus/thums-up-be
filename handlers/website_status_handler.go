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

func (h *WebsiteStatusHandler) GetStatus(c *gin.Context) {
	status := h.websiteStatusService.GetStatus()

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    status,
	})
}
