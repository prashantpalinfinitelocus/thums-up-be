package repository

import (
	"context"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"gorm.io/gorm"
)

type UserFriendRepository interface {
	FindByUserID(ctx context.Context, db *gorm.DB, userID string) ([]entities.UserFriend, error)
	DeleteByUserID(ctx context.Context, db *gorm.DB, userID string) error
	CreateMany(ctx context.Context, db *gorm.DB, friends []entities.UserFriend) error
}

type userFriendRepository struct {
}

func NewUserFriendRepository() UserFriendRepository {
	return &userFriendRepository{}
}

func (r *userFriendRepository) FindByUserID(ctx context.Context, db *gorm.DB, userID string) ([]entities.UserFriend, error) {
	var friends []entities.UserFriend
	if err := db.WithContext(ctx).
		Where("user_id = ? AND is_deleted = false", userID).
		Find(&friends).Error; err != nil {
		return nil, err
	}
	return friends, nil
}

func (r *userFriendRepository) DeleteByUserID(ctx context.Context, db *gorm.DB, userID string) error {
	return db.WithContext(ctx).
		Model(&entities.UserFriend{}).
		Where("user_id = ?", userID).
		Update("is_deleted", true).Error
}

func (r *userFriendRepository) CreateMany(ctx context.Context, db *gorm.DB, friends []entities.UserFriend) error {
	if len(friends) == 0 {
		return nil
	}
	return db.WithContext(ctx).Create(&friends).Error
}



