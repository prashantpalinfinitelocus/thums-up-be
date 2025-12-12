package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Infinite-Locus-Product/thums_up_backend/handlers"
)

func SetupWinnerRoutes(api *gin.RouterGroup, winnerHandler *handlers.WinnerHandler) {
	winners := api.Group("/winners")
	{
		winners.GET("/", winnerHandler.GetAllWinners)
		winners.GET("/week/:weekNumber", winnerHandler.GetWinnersByWeek)
	}
}
