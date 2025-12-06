package repository

import (
	"context"
	"time"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"gorm.io/gorm"
)

type OTPRepository interface {
	GenericRepository[entities.OTPLog]
	FindLatestByPhoneNumber(ctx context.Context, db *gorm.DB, phoneNumber string) (*entities.OTPLog, error)
	VerifyOTP(ctx context.Context, db *gorm.DB, phoneNumber string, otp string) (bool, error)
	CountRecentAttempts(ctx context.Context, db *gorm.DB, phoneNumber string, duration time.Duration) (int64, error)
	IncrementAttempts(ctx context.Context, db *gorm.DB, phoneNumber string) error
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
	return db.WithContext(ctx).Model(&entities.OTPLog{}).
		Where("phone_number = ? AND is_verified = ?", phoneNumber, false).
		Update("attempts", gorm.Expr("attempts + ?", 1)).Error
}

