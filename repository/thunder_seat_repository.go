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
	GetRandomEntriesByWeek(ctx context.Context, db *gorm.DB, weekNumber int, limit int, excludeUserIDs []string) ([]entities.ThunderSeat, error)
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
	if err := db.WithContext(ctx).
		Preload("User.Avatar").
		Where("user_id = ?", userID).
		Find(&entries).Error; err != nil {
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
	var userIDs []string
	
	subquery := db.WithContext(ctx).
		Model(&entities.ThunderSeat{}).
		Select("user_id").
		Group("user_id")
	
	if len(excludeUserIDs) > 0 {
		subquery = subquery.Where("user_id NOT IN ?", excludeUserIDs)
	}
	
	if err := subquery.Order("RANDOM()").Limit(limit).Pluck("user_id", &userIDs).Error; err != nil {
		return nil, err
	}
	
	if len(userIDs) == 0 {
		return []entities.ThunderSeat{}, nil
	}
	
	var entries []entities.ThunderSeat
	query := db.WithContext(ctx).
		Where("user_id IN ?", userIDs).
		Order("RANDOM()")
	
	if err := query.Find(&entries).Error; err != nil {
		return nil, err
	}
	
	userIDMap := make(map[string]entities.ThunderSeat)
	for _, entry := range entries {
		if _, exists := userIDMap[entry.UserID]; !exists {
			userIDMap[entry.UserID] = entry
		}
	}
	
	result := make([]entities.ThunderSeat, 0, len(userIDMap))
	for _, entry := range userIDMap {
		result = append(result, entry)
	}
	
	return result, nil
}

func (r *thunderSeatRepository) GetRandomEntriesByWeek(ctx context.Context, db *gorm.DB, weekNumber int, limit int, excludeUserIDs []string) ([]entities.ThunderSeat, error) {
	var userIDs []string
	
	subquery := db.WithContext(ctx).
		Model(&entities.ThunderSeat{}).
		Select("user_id").
		Where("week_number = ?", weekNumber).
		Group("user_id")
	
	if len(excludeUserIDs) > 0 {
		subquery = subquery.Where("user_id NOT IN ?", excludeUserIDs)
	}
	
	if err := subquery.Order("RANDOM()").Limit(limit).Pluck("user_id", &userIDs).Error; err != nil {
		return nil, err
	}
	
	if len(userIDs) == 0 {
		return []entities.ThunderSeat{}, nil
	}
	
	var entries []entities.ThunderSeat
	query := db.WithContext(ctx).
		Where("week_number = ? AND user_id IN ?", weekNumber, userIDs).
		Order("RANDOM()")
	
	if err := query.Find(&entries).Error; err != nil {
		return nil, err
	}
	
	userIDMap := make(map[string]entities.ThunderSeat)
	for _, entry := range entries {
		if _, exists := userIDMap[entry.UserID]; !exists {
			userIDMap[entry.UserID] = entry
		}
	}
	
	result := make([]entities.ThunderSeat, 0, len(userIDMap))
	for _, entry := range userIDMap {
		result = append(result, entry)
	}
	
	return result, nil
}
