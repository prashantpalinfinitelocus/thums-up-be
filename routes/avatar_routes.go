package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/handlers"
	"github.com/Infinite-Locus-Product/thums_up_backend/middlewares"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
)

func SetupAvatarRoutes(
	api *gin.RouterGroup,
	db *gorm.DB,
	userRepo repository.UserRepository,
	avatarHandler *handlers.AvatarHandler,
) {
	avatarGroup := api.Group("/avatars")
	{
		avatarGroup.GET("", avatarHandler.GetAvatars)
		avatarGroup.GET("/:avatarId", avatarHandler.GetAvatarByID)

		avatarGroup.Use(middlewares.AuthMiddleware(db, userRepo))
		avatarGroup.POST("", avatarHandler.CreateAvatar)
	}
}
