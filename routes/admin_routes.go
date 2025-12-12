package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Infinite-Locus-Product/thums_up_backend/handlers"
	"github.com/Infinite-Locus-Product/thums_up_backend/middlewares"
)

func SetupAdminRoutes(api *gin.RouterGroup, winnerHandler *handlers.WinnerHandler) {
	admin := api.Group("/admin")
	admin.Use(middlewares.APIKeyMiddleware())
	{
		admin.POST("/winners/select", winnerHandler.SelectWinners)
	}
}

