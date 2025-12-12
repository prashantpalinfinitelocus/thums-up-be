package services

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/config"
	"github.com/Infinite-Locus-Product/thums_up_backend/constants"
	"github.com/Infinite-Locus-Product/thums_up_backend/dtos"
	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"github.com/Infinite-Locus-Product/thums_up_backend/errors"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
	"github.com/Infinite-Locus-Product/thums_up_backend/utils"
	"github.com/Infinite-Locus-Product/thums_up_backend/vendors"
)

type AuthService interface {
	SendOTP(ctx context.Context, phoneNumber string) error
	VerifyOTP(ctx context.Context, phoneNumber string, otp string) (*dtos.TokenResponse, error)
	SignUp(ctx context.Context, req dtos.SignUpRequest) (*dtos.TokenResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dtos.TokenResponse, error)
}

type authService struct {
	txnManager       *utils.TransactionManager
	userRepo         repository.UserRepository
	otpRepo          repository.OTPRepository
	refreshTokenRepo repository.RefreshTokenRepository
	infobipClient    *vendors.InfobipClient
	cfg              *config.Config
}

func NewAuthService(
	txnManager *utils.TransactionManager,
	userRepo repository.UserRepository,
	otpRepo repository.OTPRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	infobipClient *vendors.InfobipClient,
) AuthService {
	return &authService{
		txnManager:       txnManager,
		userRepo:         userRepo,
		otpRepo:          otpRepo,
		refreshTokenRepo: refreshTokenRepo,
		infobipClient:    infobipClient,
		cfg:              config.GetConfig(),
	}
}

func (s *authService) SendOTP(ctx context.Context, phoneNumber string) error {
	formattedPhone := utils.FormatPhoneNumber(phoneNumber)

	count, err := s.otpRepo.CountRecentAttempts(ctx, s.txnManager.GetDB(), phoneNumber,
		time.Duration(constants.OTP_COOLDOWN_MINUTES)*time.Minute)
	if err == nil && count >= constants.MAX_OTP_ATTEMPTS {
		return errors.NewTooManyRequestsError(errors.ErrOTPTooManyRequests, nil)
	}

	otp := utils.GenerateOTP(constants.OTP_LENGTH)
	expiresAt := time.Now().Add(time.Duration(constants.OTP_EXPIRY_MINUTES) * time.Minute)

	otpLog := &entities.OTPLog{
		PhoneNumber: phoneNumber,
		OTP:         otp,
		ExpiresAt:   expiresAt,
		IsVerified:  false,
	}

	if err := s.otpRepo.Create(ctx, s.txnManager.GetDB(), otpLog); err != nil {
		log.WithError(err).Error("Failed to save OTP")
		return errors.NewInternalServerError(errors.ErrOTPSendFailed, err)
	}

	message := fmt.Sprintf("Your Thums Up verification code is: %s. Valid for %d minutes.",
		otp, constants.OTP_EXPIRY_MINUTES)

	if err := s.infobipClient.SendSMS(formattedPhone, message); err != nil {
		log.WithError(err).Error("Failed to send SMS")
		return errors.NewInternalServerError(errors.ErrOTPSMSFailed, err)
	}

	return nil
}

func (s *authService) VerifyOTP(ctx context.Context, phoneNumber string, otp string) (*dtos.TokenResponse, error) {
	valid, err := s.otpRepo.VerifyOTP(ctx, s.txnManager.GetDB(), phoneNumber, otp)
	if err != nil || !valid {
		s.otpRepo.IncrementAttempts(ctx, s.txnManager.GetDB(), phoneNumber)
		return nil, errors.NewUnauthorizedError(errors.ErrOTPInvalidOrExpired, err)
	}

	user, err := s.userRepo.FindByPhoneNumber(ctx, s.txnManager.GetDB(), phoneNumber)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NewNotFoundError(errors.ErrUserNotFound, err)
		}
		return nil, errors.NewInternalServerError(errors.ErrUserNotFound, err)
	}

	err = s.txnManager.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		return tx.Model(&entities.OTPLog{}).
			Where("phone_number = ? AND otp = ?", phoneNumber, otp).
			Updates(map[string]interface{}{
				"is_verified": true,
				"verified_at": time.Now(),
			}).Error
	})
	if err != nil {
		return nil, errors.NewInternalServerError(errors.ErrOTPVerifyFailed, err)
	}

	tokenResponse, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return tokenResponse, nil
}

func (s *authService) SignUp(ctx context.Context, req dtos.SignUpRequest) (*dtos.TokenResponse, error) {
	existing, err := s.userRepo.FindByPhoneNumber(ctx, s.txnManager.GetDB(), req.PhoneNumber)
	if err != nil && !stderrors.Is(err, gorm.ErrRecordNotFound) {
		log.WithError(err).Error("Failed to check existing phone number")
		return nil, errors.NewInternalServerError(errors.ErrPhoneNumberCheck, err)
	}
	if existing != nil {
		return nil, errors.NewConflictError(errors.ErrUserAlreadyExists, nil)
	}

	if req.Email != nil {
		existingEmail, err := s.userRepo.FindByEmail(ctx, s.txnManager.GetDB(), *req.Email)
		if err != nil && !stderrors.Is(err, gorm.ErrRecordNotFound) {
			log.WithError(err).Error("Failed to check existing email")
			return nil, errors.NewInternalServerError(errors.ErrEmailCheck, err)
		}
		if existingEmail != nil {
			return nil, errors.NewConflictError(errors.ErrEmailAlreadyInUse, nil)
		}
	}

	referralCode := utils.GenerateReferralCode()
	user := &entities.User{
		PhoneNumber:  req.PhoneNumber,
		Name:         &req.Name,
		Email:        req.Email,
		ReferralCode: &referralCode,
		ReferredBy:   req.ReferralCode,
		DeviceToken:  req.DeviceToken,
		IsActive:     true,
		IsVerified:   false,
	}

	if req.ReferralCode != nil {
		_, err := s.userRepo.FindByReferralCode(ctx, s.txnManager.GetDB(), *req.ReferralCode)
		if err != nil {
			log.WithError(err).Warn("Invalid referral code provided")
		}
	}

	err = s.txnManager.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		return s.userRepo.Create(ctx, tx, user)
	})
	if err != nil {
		log.WithError(err).Error("Failed to create user")
		return nil, errors.NewInternalServerError(errors.ErrProfileCreateFailed, err)
	}

	tokenResponse, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return tokenResponse, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*dtos.TokenResponse, error) {
	token, err := s.refreshTokenRepo.FindByToken(ctx, s.txnManager.GetDB(), refreshToken)
	if err != nil {
		return nil, errors.NewUnauthorizedError(errors.ErrRefreshTokenInvalid, err)
	}

	if token.IsRevoked {
		return nil, errors.NewUnauthorizedError(errors.ErrRefreshTokenRevoked, nil)
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, errors.NewUnauthorizedError(errors.ErrRefreshTokenExpired, nil)
	}

	user, err := s.userRepo.FindByID(ctx, s.txnManager.GetDB(), token.UserID)
	if err != nil {
		return nil, errors.NewInternalServerError(errors.ErrUserNotFound, err)
	}

	err = s.txnManager.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		return s.refreshTokenRepo.RevokeByToken(ctx, tx, refreshToken)
	})
	if err != nil {
		log.WithError(err).Error("Failed to revoke old refresh token")
	}

	tokenResponse, err := s.generateTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return tokenResponse, nil
}

func (s *authService) generateTokens(ctx context.Context, user *entities.User) (*dtos.TokenResponse, error) {
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, errors.NewInternalServerError(errors.ErrTokenGenerationFailed, err)
	}

	refreshTokenString := uuid.New().String()
	expiresAt := time.Now().Add(time.Duration(s.cfg.JwtConfig.RefreshTokenExpiry) * time.Second)

	refreshToken := &entities.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenString,
		ExpiresAt: expiresAt,
		IsRevoked: false,
	}

	err = s.txnManager.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		return s.refreshTokenRepo.Create(ctx, tx, refreshToken)
	})
	if err != nil {
		return nil, errors.NewInternalServerError(errors.ErrTokenRefreshFailed, err)
	}

	name := ""
	if user.Name != nil {
		name = *user.Name
	}

	return &dtos.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(s.cfg.JwtConfig.AccessTokenExpiry),
		TokenType:    "Bearer",
		UserID:       user.ID,
		PhoneNumber:  user.PhoneNumber,
		Name:         name,
	}, nil
}

func (s *authService) generateAccessToken(user *entities.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"phone":   user.PhoneNumber,
		"exp":     time.Now().Add(time.Duration(s.cfg.JwtConfig.AccessTokenExpiry) * time.Second).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JwtConfig.SecretKey))
}
