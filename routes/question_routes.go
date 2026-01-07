package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/handlers"
	"github.com/Infinite-Locus-Product/thums_up_backend/middlewares"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
)

func SetupQuestionRoutes(
	api *gin.RouterGroup,
	db *gorm.DB,
	userRepo repository.UserRepository,
	questionHandler *handlers.QuestionHandler,
) {
	questions := api.Group("/questions")
	{
		questions.GET("/active", questionHandler.GetActiveQuestions)

		questionsAuth := questions.Group("")
		questionsAuth.Use(middlewares.AuthMiddleware(db, userRepo))
		{
			questionsAuth.POST("", questionHandler.SubmitQuestion)
		}
	}
}
