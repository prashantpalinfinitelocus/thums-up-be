package repository

import (
	"context"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"gorm.io/gorm"
)

type WinnerRepository interface {
	GenericRepository[entities.ThunderSeatWinner]
	FindByWeekNumber(ctx context.Context, db *gorm.DB, weekNumber int) ([]entities.ThunderSeatWinner, error)
	FindByUserID(ctx context.Context, db *gorm.DB, userID string) ([]entities.ThunderSeatWinner, error)
	GetWinnerUserIDs(ctx context.Context, db *gorm.DB, weekNumber int) ([]string, error)
	CheckUserWinner(ctx context.Context, db *gorm.DB, userID string, weekNumber int) (bool, error)
	FindAllWithPagination(ctx context.Context, db *gorm.DB, limit, offset int) ([]entities.ThunderSeatWinner, int64, error)
	FindLatestByUserID(ctx context.Context, db *gorm.DB, userID string) (*entities.ThunderSeatWinner, error)
	UpdateHasViewed(ctx context.Context, db *gorm.DB, winnerID int) error
}

type winnerRepository struct {
	*GormRepository[entities.ThunderSeatWinner]
}

func NewWinnerRepository() WinnerRepository {
	return &winnerRepository{
		GormRepository: NewGormRepository[entities.ThunderSeatWinner](),
	}
}

func (r *winnerRepository) FindByWeekNumber(ctx context.Context, db *gorm.DB, weekNumber int) ([]entities.ThunderSeatWinner, error) {
	var winners []entities.ThunderSeatWinner
	if err := db.WithContext(ctx).Where("week_number = ?", weekNumber).Find(&winners).Error; err != nil {
		return nil, err
	}
	return winners, nil
}

func (r *winnerRepository) FindByUserID(ctx context.Context, db *gorm.DB, userID string) ([]entities.ThunderSeatWinner, error) {
	var winners []entities.ThunderSeatWinner
	if err := db.WithContext(ctx).Where("user_id = ?", userID).Find(&winners).Error; err != nil {
		return nil, err
	}
	return winners, nil
}

func (r *winnerRepository) GetWinnerUserIDs(ctx context.Context, db *gorm.DB, weekNumber int) ([]string, error) {
	var userIDs []string
	if err := db.WithContext(ctx).Model(&entities.ThunderSeatWinner{}).Where("week_number = ?", weekNumber).Pluck("user_id", &userIDs).Error; err != nil {
		return nil, err
	}
	return userIDs, nil
}

func (r *winnerRepository) CheckUserWinner(ctx context.Context, db *gorm.DB, userID string, weekNumber int) (bool, error) {
	var count int64
	if err := db.WithContext(ctx).Model(&entities.ThunderSeatWinner{}).Where("user_id = ? AND week_number = ?", userID, weekNumber).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *winnerRepository) FindAllWithPagination(ctx context.Context, db *gorm.DB, limit, offset int) ([]entities.ThunderSeatWinner, int64, error) {
	var winners []entities.ThunderSeatWinner
	var total int64

	if err := db.WithContext(ctx).Model(&entities.ThunderSeatWinner{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.WithContext(ctx).Order("created_on DESC").Limit(limit).Offset(offset).Find(&winners).Error; err != nil {
		return nil, 0, err
	}

	return winners, total, nil
}

func (r *winnerRepository) FindLatestByUserID(ctx context.Context, db *gorm.DB, userID string) (*entities.ThunderSeatWinner, error) {
	var winner entities.ThunderSeatWinner
	if err := db.WithContext(ctx).Where("user_id = ?", userID).Order("created_on DESC").First(&winner).Error; err != nil {
		return nil, err
	}
	return &winner, nil
}

func (r *winnerRepository) UpdateHasViewed(ctx context.Context, db *gorm.DB, winnerID int) error {
	return db.WithContext(ctx).Model(&entities.ThunderSeatWinner{}).Where("id = ?", winnerID).Update("has_viewed", true).Error
}
