package repository

import (
	"context"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"gorm.io/gorm"
)

type UserAdditionalInfoRepository interface {
	GenericRepository[entities.UserAdditionalInfo]
	FindByUserID(ctx context.Context, db *gorm.DB, userID string) (*entities.UserAdditionalInfo, error)
}

type userAdditionalInfoRepository struct {
	*GormRepository[entities.UserAdditionalInfo]
}

func NewUserAdditionalInfoRepository() UserAdditionalInfoRepository {
	return &userAdditionalInfoRepository{
		GormRepository: NewGormRepository[entities.UserAdditionalInfo](),
	}
}

func (r *userAdditionalInfoRepository) FindByUserID(ctx context.Context, db *gorm.DB, userID string) (*entities.UserAdditionalInfo, error) {
	var info entities.UserAdditionalInfo
	if err := db.WithContext(ctx).
		Where("user_id = ? AND is_deleted = false", userID).
		First(&info).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &info, nil
}

