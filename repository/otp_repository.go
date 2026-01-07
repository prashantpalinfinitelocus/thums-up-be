package repository

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/Infinite-Locus-Product/thums_up_backend/constants"
	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"github.com/Infinite-Locus-Product/thums_up_backend/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type OTPRepository interface {
	GenericRepository[entities.OTPLog]
	FindLatestByPhoneNumber(ctx context.Context, db *gorm.DB, phoneNumber string) (*entities.OTPLog, error)
	VerifyOTP(ctx context.Context, db *gorm.DB, phoneNumber string, otp string) (bool, error)
	CountRecentAttempts(ctx context.Context, db *gorm.DB, phoneNumber string, duration time.Duration) (int64, error)
	IncrementAttempts(ctx context.Context, db *gorm.DB, phoneNumber string) error
	CheckVerificationRateLimit(ctx context.Context, db *gorm.DB, phoneNumber string) error
}

type otpRepository struct {
	*GormRepository[entities.OTPLog]
}

func NewOTPRepository() OTPRepository {
	return &otpRepository{
		GormRepository: NewGormRepository[entities.OTPLog](),
	}
}

func (r *otpRepository) FindLatestByPhoneNumber(ctx context.Context, db *gorm.DB, phoneNumber string) (*entities.OTPLog, error) {
	var otpLog entities.OTPLog
	if err := db.WithContext(ctx).Where("phone_number = ?", phoneNumber).
		Order("created_at DESC").First(&otpLog).Error; err != nil {
		return nil, err
	}
	return &otpLog, nil
}

func (r *otpRepository) VerifyOTP(ctx context.Context, db *gorm.DB, phoneNumber string, otp string) (bool, error) {
	var otpLog entities.OTPLog
	if err := db.WithContext(ctx).Where("phone_number = ? AND otp = ? AND is_verified = ? AND expires_at > ?",
		phoneNumber, otp, false, time.Now()).First(&otpLog).Error; err != nil {
		return false, err
	}
	return true, nil
}

func (r *otpRepository) CountRecentAttempts(ctx context.Context, db *gorm.DB, phoneNumber string, duration time.Duration) (int64, error) {
	var count int64
	cutoffTime := time.Now().Add(-duration)
	if err := db.WithContext(ctx).Model(&entities.OTPLog{}).
		Where("phone_number = ? AND created_at > ?", phoneNumber, cutoffTime).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *otpRepository) IncrementAttempts(ctx context.Context, db *gorm.DB, phoneNumber string) error {
	var otpLog entities.OTPLog
	cutoffTime := time.Now().Add(-time.Duration(constants.OTP_EXPIRY_MINUTES) * time.Minute)

	err := db.WithContext(ctx).Model(&entities.OTPLog{}).
		Where("phone_number = ? AND is_verified = ? AND expires_at > ?", phoneNumber, false, cutoffTime).
		Order("created_at DESC").
		First(&otpLog).Error

	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			logrus.Warnf("No active OTP found for phone number: %s", phoneNumber)
			return errors.NewNotFoundError("No active OTP found", err)
		}
		logrus.Errorf("Error finding OTP log: %v", err)
		return errors.NewInternalServerError("Error finding OTP log", err)
	}

	logrus.Infof("Found OTP log with ID: %v and current attempts: %d", otpLog.ID, otpLog.Attempts)

	otpLog.Attempts += 1
	err = db.WithContext(ctx).Save(&otpLog).Error

	if err != nil {
		logrus.Errorf("Error incrementing OTP attempts: %v", err)
		return errors.NewInternalServerError("Error incrementing OTP attempts", err)
	}

	logrus.Infof("Successfully incremented attempts for OTP log ID: %v", otpLog.ID)
	return nil
}

func (r *otpRepository) CheckVerificationRateLimit(ctx context.Context, db *gorm.DB, phoneNumber string) error {
	cutoffTime := time.Now().Add(-time.Duration(constants.VERIFY_OTP_RATE_LIMIT_DURATION_MINUTES) * time.Minute)

	var totalAttempts int64
	if err := db.WithContext(ctx).Model(&entities.OTPLog{}).
		Where("phone_number = ? AND created_at >= ?", phoneNumber, cutoffTime).
		Select("COALESCE(SUM(attempts), 0)").
		Scan(&totalAttempts).Error; err != nil {
		logrus.Errorf("Error counting verification attempts: %v", err)
		return errors.NewInternalServerError("Error counting verification attempts", err)
	}

	if totalAttempts >= int64(constants.VERIFY_OTP_MAX_ATTEMPTS) {
		return errors.NewTooManyRequestsError(
			"Too many verification attempts. Please try again after 5 minutes.",
			nil,
		)
	}

	return nil
}
