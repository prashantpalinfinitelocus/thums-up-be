package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"github.com/Infinite-Locus-Product/thums_up_backend/errors"
	"github.com/Infinite-Locus-Product/thums_up_backend/services"
)

type ProfileHandler struct {
	userService services.UserService
}

func NewProfileHandler(userService services.UserService) *ProfileHandler {
	return &ProfileHandler{
		userService: userService,
	}
}

func (h *ProfileHandler) GetProfile(ctx *gin.Context) {
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
	userProfile, err := h.userService.GetUser(ctx, userEntity.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %v", errors.ErrProfileFetchFailed, err)})
		return
	}

	if userProfile == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": errors.ErrUserNotFound})
		return
	}

	response := dtos.ProfileResponseDTO{
		User: dtos.UserProfileDTO{
			ID:           userProfile.ID,
			PhoneNumber:  userProfile.PhoneNumber,
			Name:         userProfile.Name,
			Email:        userProfile.Email,
			IsActive:     userProfile.IsActive,
			IsVerified:   userProfile.IsVerified,
			ReferralCode: userProfile.ReferralCode,
			ReferredBy:   userProfile.ReferredBy,
			CreatedAt:    userProfile.CreatedAt,
			UpdatedAt:    userProfile.UpdatedAt,
		},
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *ProfileHandler) UpdateProfile(ctx *gin.Context) {
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

	var req dtos.UpdateProfileRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %v", errors.ErrInvalidRequestBody, err)})
		return
	}

	updatedUser, err := h.userService.UpdateUser(ctx, userEntity.ID, req)
	if err != nil {
		if err.Error() == errors.ErrUserNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": errors.ErrUserNotFound})
			return
		}
		if err.Error() == errors.ErrEmailAlreadyInUse {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrEmailAlreadyInUse})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %v", errors.ErrProfileUpdateFailed, err)})
		return
	}

	ctx.JSON(http.StatusOK, updatedUser)
}
