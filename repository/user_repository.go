package repository

import (
	"context"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"gorm.io/gorm"
)

type UserRepository interface {
	GenericRepository[entities.User]
	FindByPhoneNumber(ctx context.Context, db *gorm.DB, phoneNumber string) (*entities.User, error)
	FindByEmail(ctx context.Context, db *gorm.DB, email string) (*entities.User, error)
	FindByReferralCode(ctx context.Context, db *gorm.DB, referralCode string) (*entities.User, error)
}

type userRepository struct {
	*GormRepository[entities.User]
}

func NewUserRepository() UserRepository {
	return &userRepository{
		GormRepository: NewGormRepository[entities.User](),
	}
}

func (r *userRepository) FindByPhoneNumber(ctx context.Context, db *gorm.DB, phoneNumber string) (*entities.User, error) {
	var user entities.User
	if err := db.WithContext(ctx).Where("phone_number = ?", phoneNumber).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, db *gorm.DB, email string) (*entities.User, error) {
	var user entities.User
	if err := db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByReferralCode(ctx context.Context, db *gorm.DB, referralCode string) (*entities.User, error) {
	var user entities.User
	if err := db.WithContext(ctx).Where("referral_code = ?", referralCode).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
