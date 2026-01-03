package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Infinite-Locus-Product/thums_up_backend/handlers"
)

func SetupWebsiteStatusRoutes(api *gin.RouterGroup, websiteStatusHandler *handlers.WebsiteStatusHandler) {
	api.GET("/website-status", websiteStatusHandler.GetStatus)
}
