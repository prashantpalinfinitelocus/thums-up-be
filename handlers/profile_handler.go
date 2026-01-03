package handlers

import (
	stderrors "errors"
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

// GetProfile godoc
//
//	@Summary		Get user profile
//	@Description	Retrieve the authenticated user's profile information. Requires authentication.
//	@Tags			Profile
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Success		200	{object}	dtos.ProfileResponseDTO	"Profile retrieved successfully"
//	@Failure		401	{object}	map[string]string		"Unauthorized"
//	@Failure		404	{object}	map[string]string		"User not found"
//	@Failure		500	{object}	map[string]string		"Failed to fetch profile"
//	@Router			/profile [get]
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
	userProfile, avatarImageURL, err := h.userService.GetUser(ctx, userEntity.ID)
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
			AvatarImage:  avatarImageURL,
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

// UpdateProfile godoc
//
//	@Summary		Update user profile
//	@Description	Update the authenticated user's profile information (name, email, avatar, sharing_platform, platform_user_name). Requires authentication.
//	@Tags			Profile
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Param			request	body		dtos.UpdateProfileRequestDTO	true	"Profile update data"
//	@Success		200		{object}	dtos.UserProfileDTO				"Profile updated successfully"
//	@Failure		400		{object}	map[string]string				"Validation failed or email already in use"
//	@Failure		401		{object}	map[string]string				"Unauthorized"
//	@Failure		404		{object}	map[string]string				"User not found"
//	@Failure		500		{object}	map[string]string				"Failed to update profile"
//	@Router			/profile [patch]
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
		if stderrors.Is(err, errors.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
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
