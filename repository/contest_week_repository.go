package repository

import (
	"context"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"gorm.io/gorm"
)

type ContestWeekRepository interface {
	GenericRepository[entities.ContestWeek]
	FindByWeekNumber(ctx context.Context, db *gorm.DB, weekNumber int) (*entities.ContestWeek, error)
	FindActiveWeek(ctx context.Context, db *gorm.DB) (*entities.ContestWeek, error)
	FindAll(ctx context.Context, db *gorm.DB) ([]entities.ContestWeek, error)
	DeactivateAll(ctx context.Context, db *gorm.DB) error
}

type contestWeekRepository struct {
	*GormRepository[entities.ContestWeek]
}

func NewContestWeekRepository() ContestWeekRepository {
	return &contestWeekRepository{
		GormRepository: NewGormRepository[entities.ContestWeek](),
	}
}

func (r *contestWeekRepository) FindByWeekNumber(ctx context.Context, db *gorm.DB, weekNumber int) (*entities.ContestWeek, error) {
	var week entities.ContestWeek
	if err := db.WithContext(ctx).Where("week_number = ?", weekNumber).First(&week).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &week, nil
}

func (r *contestWeekRepository) FindActiveWeek(ctx context.Context, db *gorm.DB) (*entities.ContestWeek, error) {
	var week entities.ContestWeek
	if err := db.WithContext(ctx).Where("is_active = ?", true).First(&week).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &week, nil
}

func (r *contestWeekRepository) FindAll(ctx context.Context, db *gorm.DB) ([]entities.ContestWeek, error) {
	var weeks []entities.ContestWeek
	if err := db.WithContext(ctx).Order("week_number ASC").Find(&weeks).Error; err != nil {
		return nil, err
	}
	return weeks, nil
}

func (r *contestWeekRepository) DeactivateAll(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Model(&entities.ContestWeek{}).Where("is_active = ?", true).Update("is_active", false).Error
}
