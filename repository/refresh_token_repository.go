package repository

import (
	"context"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	GenericRepository[entities.RefreshToken]
	FindByToken(ctx context.Context, db *gorm.DB, token string) (*entities.RefreshToken, error)
	RevokeByUserID(ctx context.Context, db *gorm.DB, userID string) error
	RevokeByToken(ctx context.Context, db *gorm.DB, token string) error
}

type refreshTokenRepository struct {
	*GormRepository[entities.RefreshToken]
}

func NewRefreshTokenRepository() RefreshTokenRepository {
	return &refreshTokenRepository{
		GormRepository: NewGormRepository[entities.RefreshToken](),
	}
}

func (r *refreshTokenRepository) FindByToken(ctx context.Context, db *gorm.DB, token string) (*entities.RefreshToken, error) {
	var refreshToken entities.RefreshToken
	if err := db.WithContext(ctx).Where("token = ?", token).First(&refreshToken).Error; err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *refreshTokenRepository) RevokeByUserID(ctx context.Context, db *gorm.DB, userID string) error {
	return db.WithContext(ctx).Model(&entities.RefreshToken{}).
		Where("user_id = ?", userID).
		Update("is_revoked", true).Error
}

func (r *refreshTokenRepository) RevokeByToken(ctx context.Context, db *gorm.DB, token string) error {
	return db.WithContext(ctx).Model(&entities.RefreshToken{}).
		Where("token = ?", token).
		Update("is_revoked", true).Error
}
