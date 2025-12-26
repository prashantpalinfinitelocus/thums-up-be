package handlers

import (
	stderrors "errors"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/errors"
	"github.com/Infinite-Locus-Product/thums_up_backend/services"
	"github.com/Infinite-Locus-Product/thums_up_backend/utils"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// SendOTP godoc
// @Summary Send OTP to phone number
// @Description Send a one-time password to the provided phone number for authentication. Returns the OTP in the response (for development/testing purposes when SMS service is not configured).
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dtos.SendOTPRequest true "Phone number"
// @Success 200 {object} dtos.SuccessResponse{data=dtos.OTPResponse} "OTP sent successfully"
// @Failure 400 {object} dtos.ErrorResponse "Validation failed"
// @Failure 429 {object} dtos.ErrorResponse "Too many OTP requests"
// @Failure 500 {object} dtos.ErrorResponse "Failed to send OTP"
// @Router /auth/send-otp [post]
func (h *AuthHandler) SendOTP(c *gin.Context) {
	var req dtos.SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := utils.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
		return
	}

	otp, err := h.authService.SendOTP(c.Request.Context(), req.PhoneNumber)
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to send OTP")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrOTPSendFailed,
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Message: "OTP sent successfully",
		Data:    dtos.OTPResponse{OTP: otp},
	})
}

// VerifyOTP godoc
// @Summary Verify OTP
// @Description Verify the OTP sent to the phone number and return authentication token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dtos.VerifyOTPRequest true "Phone number and OTP"
// @Success 200 {object} dtos.SuccessResponse{data=dtos.TokenResponse} "OTP verified successfully"
// @Failure 400 {object} dtos.ErrorResponse "Validation failed"
// @Failure 401 {object} dtos.ErrorResponse "Invalid OTP"
// @Failure 500 {object} dtos.ErrorResponse "Failed to verify OTP"
// @Router /auth/verify-otp [post]
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req dtos.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := utils.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
		return
	}

	tokenResponse, err := h.authService.VerifyOTP(c.Request.Context(), req.PhoneNumber, req.OTP)
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to verify OTP")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrOTPVerifyFailed,
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    tokenResponse,
	})
}

// SignUp godoc
// @Summary User sign up
// @Description Register a new user with phone number, name, and optional email and referral code
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dtos.SignUpRequest true "User registration details"
// @Success 201 {object} dtos.SuccessResponse{data=dtos.TokenResponse} "User registered successfully"
// @Failure 400 {object} dtos.ErrorResponse "Validation failed"
// @Failure 500 {object} dtos.ErrorResponse "Failed to sign up"
// @Router /auth/signup [post]
func (h *AuthHandler) SignUp(c *gin.Context) {
	var req dtos.SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := utils.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
		return
	}

	tokenResponse, err := h.authService.SignUp(c.Request.Context(), req)
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to sign up")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrProfileCreateFailed,
		})
		return
	}

	c.JSON(http.StatusCreated, dtos.SuccessResponse{
		Success: true,
		Data:    tokenResponse,
	})
}

// RefreshToken godoc
// @Summary Refresh authentication token
// @Description Get a new access token using a valid refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dtos.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} dtos.SuccessResponse{data=dtos.TokenResponse} "Token refreshed successfully"
// @Failure 400 {object} dtos.ErrorResponse "Validation failed"
// @Failure 401 {object} dtos.ErrorResponse "Invalid refresh token"
// @Failure 500 {object} dtos.ErrorResponse "Failed to refresh token"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dtos.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := utils.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
		return
	}

	tokenResponse, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			c.JSON(appErr.StatusCode, dtos.ErrorResponse{
				Success: false,
				Error:   appErr.Message,
			})
			return
		}
		log.WithError(err).Error("Failed to refresh token")
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Success: false,
			Error:   errors.ErrTokenRefreshFailed,
		})
		return
	}

	c.JSON(http.StatusOK, dtos.SuccessResponse{
		Success: true,
		Data:    tokenResponse,
	})
}
