package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/handlers"
	"github.com/Infinite-Locus-Product/thums_up_backend/middlewares"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
)

func SetupProfileRoutes(
	api *gin.RouterGroup,
	db *gorm.DB,
	userRepo repository.UserRepository,
	profileHandler *handlers.ProfileHandler,
	addressHandler *handlers.AddressHandler,
	questionHandler *handlers.QuestionHandler,
) {
	profileGroup := api.Group("/profile")
	profileGroup.Use(middlewares.AuthMiddleware(db, userRepo))
	{
		profileGroup.GET("/", profileHandler.GetProfile)
		profileGroup.PATCH("/", profileHandler.UpdateProfile)

		profileGroup.POST("/address", addressHandler.AddAddress)
		profileGroup.GET("/address", addressHandler.GetAddresses)
		profileGroup.PUT("/address/:addressId", addressHandler.UpdateAddress)
		profileGroup.DELETE("/address/:addressId", addressHandler.DeleteAddress)

		// Question operations
		profileGroup.GET("/questions", questionHandler.GetQuestions)
		profileGroup.POST("/questions/text", questionHandler.GetQuestionByID)
		profileGroup.POST("/questions", questionHandler.AnswerQuestions)
		profileGroup.POST("/questions/create", questionHandler.CreateQuestions)
	}
}
