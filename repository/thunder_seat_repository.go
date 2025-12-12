package repository

import (
	"context"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"gorm.io/gorm"
)

type ThunderSeatRepository interface {
	GenericRepository[entities.ThunderSeat]
	FindByUserID(ctx context.Context, db *gorm.DB, userID string) ([]entities.ThunderSeat, error)
	FindByQuestionID(ctx context.Context, db *gorm.DB, questionID int) ([]entities.ThunderSeat, error)
	CheckUserSubmission(ctx context.Context, db *gorm.DB, userID string, questionID int) (*entities.ThunderSeat, error)
	GetRandomEntries(ctx context.Context, db *gorm.DB, limit int, excludeUserIDs []string) ([]entities.ThunderSeat, error)
}

type thunderSeatRepository struct {
	*GormRepository[entities.ThunderSeat]
}

func NewThunderSeatRepository() ThunderSeatRepository {
	return &thunderSeatRepository{
		GormRepository: NewGormRepository[entities.ThunderSeat](),
	}
}

func (r *thunderSeatRepository) FindByUserID(ctx context.Context, db *gorm.DB, userID string) ([]entities.ThunderSeat, error) {
	var entries []entities.ThunderSeat
	if err := db.WithContext(ctx).Where("user_id = ?", userID).Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func (r *thunderSeatRepository) FindByQuestionID(ctx context.Context, db *gorm.DB, questionID int) ([]entities.ThunderSeat, error) {
	var entries []entities.ThunderSeat
	if err := db.WithContext(ctx).Where("question_id = ?", questionID).Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func (r *thunderSeatRepository) CheckUserSubmission(ctx context.Context, db *gorm.DB, userID string, questionID int) (*entities.ThunderSeat, error) {
	var entry entities.ThunderSeat
	if err := db.WithContext(ctx).Where("user_id = ? AND question_id = ?", userID, questionID).First(&entry).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entry, nil
}

func (r *thunderSeatRepository) GetRandomEntries(ctx context.Context, db *gorm.DB, limit int, excludeUserIDs []string) ([]entities.ThunderSeat, error) {
	var entries []entities.ThunderSeat
	query := db.WithContext(ctx).Order("RANDOM()").Limit(limit)

	if len(excludeUserIDs) > 0 {
		query = query.Where("user_id NOT IN ?", excludeUserIDs)
	}

	if err := query.Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}
