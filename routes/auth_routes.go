package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Infinite-Locus-Product/thums_up_backend/handlers"
)

func SetupAuthRoutes(api *gin.RouterGroup, authHandler *handlers.AuthHandler) {
	auth := api.Group("/auth")
	{
		auth.POST("/send-otp", authHandler.SendOTP)
		auth.POST("/verify-otp", authHandler.VerifyOTP)
		auth.POST("/signup", authHandler.SignUp)
		auth.POST("/refresh", authHandler.RefreshToken)
	}
}
