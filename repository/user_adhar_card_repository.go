package repository

import (
	"context"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"gorm.io/gorm"
)

type UserAadharCardRepository interface {
	FindByUserID(ctx context.Context, db *gorm.DB, userID string) (*entities.UserAadharCard, error)
	Create(ctx context.Context, db *gorm.DB, entity *entities.UserAadharCard) error
	Update(ctx context.Context, db *gorm.DB, entity *entities.UserAadharCard) error
}

type userAadharCardRepository struct {
}

func NewUserAadharCardRepository() UserAadharCardRepository {
	return &userAadharCardRepository{}
}

func (r *userAadharCardRepository) FindByUserID(ctx context.Context, db *gorm.DB, userID string) (*entities.UserAadharCard, error) {
	var card entities.UserAadharCard
	if err := db.WithContext(ctx).Where("user_id = ? AND is_deleted = false", userID).First(&card).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &card, nil
}

func (r *userAadharCardRepository) Create(ctx context.Context, db *gorm.DB, entity *entities.UserAadharCard) error {
	return db.WithContext(ctx).Create(entity).Error
}

func (r *userAadharCardRepository) Update(ctx context.Context, db *gorm.DB, entity *entities.UserAadharCard) error {
	return db.WithContext(ctx).Save(entity).Error
}
