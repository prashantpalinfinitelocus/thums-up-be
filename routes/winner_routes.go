package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/handlers"
	"github.com/Infinite-Locus-Product/thums_up_backend/middlewares"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
)

func SetupWinnerRoutes(api *gin.RouterGroup, db *gorm.DB, userRepo repository.UserRepository, winnerHandler *handlers.WinnerHandler) {
	winners := api.Group("/winners")
	{
		winners.GET("/", winnerHandler.GetAllWinners)
		winners.GET("/week/:weekNumber", winnerHandler.GetWinnersByWeek)

		winnersAuth := winners.Group("")
		winnersAuth.Use(middlewares.AuthMiddleware(db, userRepo))
		{
			winnersAuth.GET("/status", winnerHandler.CheckWinnerStatus)
			winnersAuth.POST("/mark-viewed", winnerHandler.MarkBannerAsViewed)
			winnersAuth.POST("/kyc", winnerHandler.SubmitWinnerKYC)
		}
	}
}
