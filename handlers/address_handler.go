package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"github.com/Infinite-Locus-Product/thums_up_backend/services"
)

type AddressHandler struct {
	userService services.UserService
}

func NewAddressHandler(userService services.UserService) *AddressHandler {
	return &AddressHandler{
		userService: userService,
	}
}

func (h *AddressHandler) GetAddresses(ctx *gin.Context) {
	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userEntity, ok := user.(*entities.User)
	if !ok || userEntity == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
		return
	}
	addresses, err := h.userService.GetUserAddresses(ctx, userEntity.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch addresses: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, addresses)
}

func (h *AddressHandler) AddAddress(ctx *gin.Context) {
	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userEntity, ok := user.(*entities.User)
	if !ok || userEntity == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
		return
	}
	var req dtos.AddressRequestDTO

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request body: %v", err)})
		return
	}

	address, err := h.userService.AddUserAddress(ctx, userEntity.ID, req)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "state"):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "city"):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "pincode"):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "not deliverable"):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case err.Error() == "user not found":
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to add address: %v", err)})
		}
		return
	}

	if address == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Failed to create address"})
		return
	}

	ctx.JSON(http.StatusCreated, address)
}

func (h *AddressHandler) UpdateAddress(ctx *gin.Context) {
	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userEntity, ok := user.(*entities.User)
	if !ok || userEntity == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
		return
	}

	addressID := ctx.Param("addressId")
	if addressID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Address ID is required"})
		return
	}

	var req dtos.AddressRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request body: %v", err)})
		return
	}

	address, err := h.userService.UpdateUserAddress(ctx, userEntity.ID, addressID, req)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "state"):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "city"):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "pincode"):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case strings.Contains(err.Error(), "not deliverable"):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case err.Error() == "user not found":
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		case err.Error() == "address not found":
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
		case err.Error() == "address does not belong to user":
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Address does not belong to user"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update address: %v", err)})
		}
		return
	}

	if address == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
		return
	}

	ctx.JSON(http.StatusOK, address)
}

func (h *AddressHandler) DeleteAddress(ctx *gin.Context) {
	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userEntity, ok := user.(*entities.User)
	if !ok || userEntity == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
		return
	}

	addressID := ctx.Param("addressId")
	if addressID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Address ID is required"})
		return
	}

	err := h.userService.DeleteUserAddress(ctx, userEntity.ID, addressID)
	if err != nil {
		if err.Error() == "user not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if err.Error() == "address not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
			return
		}
		if err.Error() == "address does not belong to user" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Address does not belong to user"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete address: %v", err)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Address deleted successfully"})
}
