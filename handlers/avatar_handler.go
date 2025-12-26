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
	"github.com/Infinite-Locus-Product/thums_up_backend/utils"
)

type AvatarHandler struct {
	avatarService services.AvatarService
}

func NewAvatarHandler(avatarService services.AvatarService) *AvatarHandler {
	return &AvatarHandler{
		avatarService: avatarService,
	}
}

// CreateAvatar godoc
// @Summary Create a new avatar
// @Description Create a new avatar with name and image file. Requires authentication.
// @Tags Avatars
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param name formData string true "Avatar name"
// @Param image formData file true "Avatar image file (jpg, jpeg, png, gif, webp, svg, bmp, ico)"
// @Param is_published formData bool false "Whether the avatar is published"
// @Success 201 {object} dtos.AvatarResponseDTO "Avatar created successfully"
// @Failure 400 {object} map[string]string "Validation failed"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Failed to create avatar"
// @Router /avatars [post]
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

	// Bind form data
	var req dtos.CreateAvatarRequestDTO
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %v", errors.ErrInvalidRequestBody, err)})
		return
	}

	// Get image file
	imageFile, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
		return
	}

	// Validate image file
	if err := utils.ValidateImageFile(imageFile); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	avatar, err := h.avatarService.CreateAvatar(ctx, req, imageFile, userEntity.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create avatar: %v", err)})
		return
	}

	ctx.JSON(http.StatusCreated, avatar)
}

// GetAvatars godoc
// @Summary Get all avatars
// @Description Retrieve all avatars, optionally filtered by publication status
// @Tags Avatars
// @Accept json
// @Produce json
// @Param is_published query bool false "Filter by publication status"
// @Success 200 {object} map[string][]dtos.AvatarResponseDTO "Avatars retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid is_published parameter"
// @Failure 500 {object} map[string]string "Failed to fetch avatars"
// @Router /avatars [get]
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

// GetAvatarByID godoc
// @Summary Get avatar by ID
// @Description Retrieve a specific avatar by its ID
// @Tags Avatars
// @Accept json
// @Produce json
// @Param avatarId path int true "Avatar ID"
// @Success 200 {object} dtos.AvatarResponseDTO "Avatar retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid avatar ID"
// @Failure 404 {object} map[string]string "Avatar not found"
// @Failure 500 {object} map[string]string "Failed to fetch avatar"
// @Router /avatars/{avatarId} [get]
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
