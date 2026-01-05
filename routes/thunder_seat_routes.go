package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/handlers"
	"github.com/Infinite-Locus-Product/thums_up_backend/middlewares"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
)

func SetupThunderSeatRoutes(
	api *gin.RouterGroup,
	db *gorm.DB,
	userRepo repository.UserRepository,
	thunderSeatHandler *handlers.ThunderSeatHandler,
) {
	thunderSeat := api.Group("/thunder-seat")
	{
		thunderSeat.GET("/current-week", thunderSeatHandler.GetCurrentWeek)

		thunderSeatAuth := thunderSeat.Group("")
		thunderSeatAuth.Use(middlewares.AuthMiddleware(db, userRepo))
		{
			thunderSeatAuth.GET("/submissions", thunderSeatHandler.GetUserSubmissions)
			thunderSeatAuth.POST("", thunderSeatHandler.SubmitAnswer)
			thunderSeatAuth.POST("/", thunderSeatHandler.SubmitAnswer)
		}
	}
}
