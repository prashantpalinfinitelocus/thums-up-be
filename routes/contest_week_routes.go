package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/handlers"
	"github.com/Infinite-Locus-Product/thums_up_backend/middlewares"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
)

func SetupContestWeekRoutes(api *gin.RouterGroup, db *gorm.DB, userRepo repository.UserRepository, contestWeekHandler *handlers.ContestWeekHandler) {
	contestWeeks := api.Group("/contest-weeks")
	{
		contestWeeks.GET("", contestWeekHandler.GetAllContestWeeks)
		contestWeeks.GET("/active", contestWeekHandler.GetActiveWeek)
		contestWeeks.GET("/:weekNumber", contestWeekHandler.GetContestWeekByNumber)

		authRequired := contestWeeks.Group("")
		authRequired.Use(middlewares.AuthMiddleware(db, userRepo))
		{
			authRequired.POST("", contestWeekHandler.CreateContestWeek)
			authRequired.POST("/activate", contestWeekHandler.ActivateWeek)
		}
	}
}
