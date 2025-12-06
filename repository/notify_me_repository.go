package repository

import (
	"context"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"gorm.io/gorm"
)

type NotifyMeRepository interface {
	GenericRepository[entities.NotifyMe]
	FindByPhoneNumber(ctx context.Context, db *gorm.DB, phoneNumber string) (*entities.NotifyMe, error)
	FindUnnotified(ctx context.Context, db *gorm.DB) ([]entities.NotifyMe, error)
	MarkAsNotified(ctx context.Context, db *gorm.DB, id string) error
}

type notifyMeRepository struct {
	*GormRepository[entities.NotifyMe]
}

func NewNotifyMeRepository() NotifyMeRepository {
	return &notifyMeRepository{
		GormRepository: NewGormRepository[entities.NotifyMe](),
	}
}

func (r *notifyMeRepository) FindByPhoneNumber(ctx context.Context, db *gorm.DB, phoneNumber string) (*entities.NotifyMe, error) {
	var notifyMe entities.NotifyMe
	if err := db.WithContext(ctx).Where("phone_number = ?", phoneNumber).First(&notifyMe).Error; err != nil {
		return nil, err
	}
	return &notifyMe, nil
}

func (r *notifyMeRepository) FindUnnotified(ctx context.Context, db *gorm.DB) ([]entities.NotifyMe, error) {
	var records []entities.NotifyMe
	if err := db.WithContext(ctx).Where("is_notified = ?", false).Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *notifyMeRepository) MarkAsNotified(ctx context.Context, db *gorm.DB, id string) error {
	return db.WithContext(ctx).Model(&entities.NotifyMe{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_notified": true,
			"notified_at": gorm.Expr("NOW()"),
		}).Error
}

