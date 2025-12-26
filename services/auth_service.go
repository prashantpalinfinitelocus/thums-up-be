package services

import (
	"context"
	stderrors "errors"
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
	SendOTP(ctx context.Context, phoneNumber string) (string, error)
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

func (s *authService) SendOTP(ctx context.Context, phoneNumber string) (string, error) {
	// formattedPhone := utils.FormatPhoneNumber(phoneNumber)

	count, err := s.otpRepo.CountRecentAttempts(ctx, s.txnManager.GetDB(), phoneNumber,
		time.Duration(constants.OTP_COOLDOWN_MINUTES)*time.Minute)
	if err == nil && count >= constants.MAX_OTP_ATTEMPTS {
		return "", errors.NewTooManyRequestsError(errors.ErrOTPTooManyRequests, nil)
	}

	otp, err := utils.GenerateOTP(constants.OTP_LENGTH)
	if err != nil {
		log.WithError(err).Error("Failed to generate OTP")
		return "", errors.NewInternalServerError(errors.ErrOTPSendFailed, err)
	}
	expiresAt := time.Now().Add(time.Duration(constants.OTP_EXPIRY_MINUTES) * time.Minute)

	otpLog := &entities.OTPLog{
		PhoneNumber: phoneNumber,
		OTP:         otp,
		ExpiresAt:   expiresAt,
		IsVerified:  false,
	}

	if err := s.otpRepo.Create(ctx, s.txnManager.GetDB(), otpLog); err != nil {
		log.WithError(err).Error("Failed to save OTP")
		return "", errors.NewInternalServerError(errors.ErrOTPSendFailed, err)
	}

	// message := fmt.Sprintf("Your Thums Up verification code is: %s. Valid for %d minutes.",
	// 	otp, constants.OTP_EXPIRY_MINUTES)

	// if err := s.infobipClient.SendSMS(ctx, formattedPhone, message); err != nil {
	// 	log.WithError(err).Error("Failed to send SMS")
	// 	return "", errors.NewInternalServerError(errors.ErrOTPSMSFailed, err)
	// }

	return otp, nil
}

func (s *authService) VerifyOTP(ctx context.Context, phoneNumber string, otp string) (*dtos.TokenResponse, error) {
	// Start main transaction
	var tokenResponse *dtos.TokenResponse
	err := s.txnManager.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		// Step 1: Check verification rate limiting
		rateLimitCount, err := s.otpRepo.CountRecentAttempts(ctx, tx, phoneNumber,
			time.Duration(constants.VERIFY_OTP_RATE_LIMIT_DURATION_MINUTES)*time.Minute)
		if err != nil {
			log.WithError(err).Error("Failed to check verification rate limit")
			return errors.NewInternalServerError("Failed to check rate limit", err)
		}
		if rateLimitCount >= constants.VERIFY_OTP_MAX_ATTEMPTS {
			return errors.NewTooManyRequestsError(
				"Too many verification attempts. Please try again after 5 minutes.", nil)
		}

		// Step 2: Verify OTP
		isValid, err := s.otpRepo.VerifyOTP(ctx, tx, phoneNumber, otp)
		if err != nil || !isValid {
			// Handle failed verification in a separate goroutine with its own transaction
			go s.handleFailedOTPAttempt(context.Background(), phoneNumber)

			if err != nil {
				return errors.NewUnauthorizedError(errors.ErrOTPInvalidOrExpired, err)
			}
			return errors.NewUnauthorizedError("Invalid OTP", nil)
		}

		// Step 3: Mark OTP as verified
		if err := tx.Model(&entities.OTPLog{}).
			Where("phone_number = ? AND otp = ? AND is_verified = ?", phoneNumber, otp, false).
			Updates(map[string]interface{}{
				"is_verified": true,
				"verified_at": time.Now(),
			}).Error; err != nil {
			log.WithError(err).Error("Failed to mark OTP as verified")
			return errors.NewInternalServerError(errors.ErrOTPVerifyFailed, err)
		}

		// Step 4: Check if user exists
		user, err := s.userRepo.FindByPhoneNumber(ctx, tx, phoneNumber)
		if err != nil {
			if stderrors.Is(err, gorm.ErrRecordNotFound) {
				// User doesn't exist - generate temporary token for signup flow
				tempToken, err := s.generateTempAccessToken(phoneNumber)
				if err != nil {
					return errors.NewInternalServerError("Failed to generate temp access token", err)
				}

				tokenResponse = &dtos.TokenResponse{
					AccessToken:  tempToken,
					RefreshToken: "",
					ExpiresIn:    int64(5 * 60), // 5 minutes in seconds
					TokenType:    "temp",
					PhoneNumber:  phoneNumber,
				}
				return nil
			}
			return errors.NewInternalServerError("Failed to fetch user", err)
		}

		// Step 5: User exists - generate full token pair
		accessToken, err := s.generateAccessToken(user)
		if err != nil {
			return errors.NewInternalServerError(errors.ErrTokenGenerationFailed, err)
		}

		// Generate refresh token
		refreshTokenString := uuid.New().String()
		expiresAt := time.Now().Add(time.Duration(s.cfg.JwtConfig.RefreshTokenExpiry) * time.Second)

		refreshToken := &entities.RefreshToken{
			UserID:    user.ID,
			Token:     refreshTokenString,
			ExpiresAt: expiresAt,
			IsRevoked: false,
		}

		// Step 6: Store refresh token
		if err := s.refreshTokenRepo.Create(ctx, tx, refreshToken); err != nil {
			return errors.NewInternalServerError("Failed to store refresh token", err)
		}

		// Step 7: Track login count (optional - log error but don't fail)
		if err := s.createOrIncrementLoginCount(ctx, tx, user.ID, user.PhoneNumber); err != nil {
			log.WithError(err).Error("Failed to create/increment login count")
		}

		// Build token response
		name := ""
		if user.Name != nil {
			name = *user.Name
		}

		email := ""
		if user.Email != nil {
			email = *user.Email
		}

		tokenResponse = &dtos.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshTokenString,
			ExpiresIn:    int64(s.cfg.JwtConfig.AccessTokenExpiry),
			TokenType:    "Bearer",
			UserID:       user.ID,
			PhoneNumber:  user.PhoneNumber,
			Name:         name,
			Email:        email,
		}

		return nil
	})

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

	referralCode, err := utils.GenerateReferralCode()
	if err != nil {
		log.WithError(err).Error("Failed to generate referral code")
		return nil, errors.NewInternalServerError(errors.ErrProfileCreateFailed, err)
	}
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
		return nil, errors.NewInternalServerError("User not found", err)
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

	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	return &dtos.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(s.cfg.JwtConfig.AccessTokenExpiry),
		TokenType:    "Bearer",
		UserID:       user.ID,
		PhoneNumber:  user.PhoneNumber,
		Name:         name,
		Email:        email,
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

// handleFailedOTPAttempt increments the OTP attempts counter in a separate transaction
// This ensures attempt tracking even if the main transaction fails
func (s *authService) handleFailedOTPAttempt(ctx context.Context, phoneNumber string) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Panic recovered in handleFailedOTPAttempt: %v", r)
		}
	}()

	err := s.txnManager.ExecuteInTransaction(ctx, func(tx *gorm.DB) error {
		return s.otpRepo.IncrementAttempts(ctx, tx, phoneNumber)
	})

	if err != nil {
		log.WithError(err).Errorf("Failed to increment OTP attempts for phone: %s", phoneNumber)
	} else {
		log.Infof("Successfully incremented OTP attempts for phone: %s", phoneNumber)
	}
}

// generateTempAccessToken creates a temporary JWT token for users who haven't completed signup
func (s *authService) generateTempAccessToken(phoneNumber string) (string, error) {
	claims := jwt.MapClaims{
		"phone":      phoneNumber,
		"token_type": "temp",
		"purpose":    "signup",
		"exp":        time.Now().Add(5 * time.Minute).Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JwtConfig.SecretKey))
}

// createOrIncrementLoginCount tracks user login activity
// This is optional and doesn't fail the main flow if it encounters errors
func (s *authService) createOrIncrementLoginCount(ctx context.Context, tx *gorm.DB, userID, phoneNumber string) error {
	// Check if login count record exists
	type LoginCount struct {
		ID          string    `gorm:"primaryKey"`
		UserID      string    `gorm:"column:user_id"`
		PhoneNumber string    `gorm:"column:phone_number"`
		Count       int       `gorm:"column:count"`
		LastLogin   time.Time `gorm:"column:last_login"`
		CreatedAt   time.Time `gorm:"column:created_at"`
		UpdatedAt   time.Time `gorm:"column:updated_at"`
	}

	var loginCount LoginCount
	err := tx.WithContext(ctx).Table("login_counts").
		Where("user_id = ?", userID).
		First(&loginCount).Error

	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			// Create new login count record
			newLoginCount := LoginCount{
				ID:          uuid.New().String(),
				UserID:      userID,
				PhoneNumber: phoneNumber,
				Count:       1,
				LastLogin:   time.Now(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			return tx.WithContext(ctx).Table("login_counts").Create(&newLoginCount).Error
		}
		return err
	}

	// Update existing login count
	return tx.WithContext(ctx).Table("login_counts").
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"count":      gorm.Expr("count + ?", 1),
			"last_login": time.Now(),
			"updated_at": time.Now(),
		}).Error
}
