package handlers

import (
	stderrors "errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"github.com/Infinite-Locus-Product/thums_up_backend/errors"
	"github.com/Infinite-Locus-Product/thums_up_backend/services"
)

type AvatarHandler struct {
	avatarService services.AvatarService
}

func NewAvatarHandler(avatarService services.AvatarService) *AvatarHandler {
	return &AvatarHandler{
		avatarService: avatarService,
	}
}

func (h *AvatarHandler) CreateAvatar(ctx *gin.Context) {
	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": errors.ErrUserNotAuthenticated})
		return
	}

	userEntity, ok := user.(*entities.User)
	if !ok || userEntity == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidUserContext})
		return
	}

	var req dtos.CreateAvatarRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %v", errors.ErrInvalidRequestBody, err)})
		return
	}

	avatar, err := h.avatarService.CreateAvatar(ctx, req, userEntity.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create avatar: %v", err)})
		return
	}

	ctx.JSON(http.StatusCreated, avatar)
}

func (h *AvatarHandler) GetAvatars(ctx *gin.Context) {
	var isPublished *bool
	if publishedParam := ctx.Query("is_published"); publishedParam != "" {
		published, err := strconv.ParseBool(publishedParam)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid is_published parameter"})
			return
		}
		isPublished = &published
	}

	avatars, err := h.avatarService.GetAllAvatars(ctx, isPublished)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch avatars: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"avatars": avatars})
}

func (h *AvatarHandler) GetAvatarByID(ctx *gin.Context) {
	avatarIDStr := ctx.Param("avatarId")
	avatarID, err := strconv.Atoi(avatarIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid avatar ID"})
		return
	}

	avatar, err := h.avatarService.GetAvatarByID(ctx, avatarID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "avatar not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch avatar: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, avatar)
}
