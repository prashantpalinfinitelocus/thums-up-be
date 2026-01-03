package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/handlers"
	"github.com/Infinite-Locus-Product/thums_up_backend/middlewares"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
)

func SetupAuthRoutes(api *gin.RouterGroup, authHandler *handlers.AuthHandler, db *gorm.DB, userRepo repository.UserRepository) {
	auth := api.Group("/auth")
	{
		auth.POST("/send-otp", authHandler.SendOTP)
		auth.POST("/verify-otp", authHandler.VerifyOTP)
		auth.POST("/signup", authHandler.SignUp)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.GET("/login-count", middlewares.AuthMiddleware(db, userRepo), authHandler.GetLoginCount)
	}
}
