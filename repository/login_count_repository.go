package repository

import (
	"context"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"gorm.io/gorm"
)

type LoginCountRepository interface {
	GenericRepository[entities.LoginCount]
	FindByUserID(ctx context.Context, db *gorm.DB, userID string) (*entities.LoginCount, error)
}

type loginCountRepository struct {
	*GormRepository[entities.LoginCount]
}

func NewLoginCountRepository() LoginCountRepository {
	return &loginCountRepository{
		GormRepository: NewGormRepository[entities.LoginCount](),
	}
}

func (r *loginCountRepository) FindByUserID(ctx context.Context, db *gorm.DB, userID string) (*entities.LoginCount, error) {
	var loginCount entities.LoginCount
	if err := db.WithContext(ctx).Where("user_id = ?", userID).First(&loginCount).Error; err != nil {
		return nil, err
	}
	return &loginCount, nil
}
