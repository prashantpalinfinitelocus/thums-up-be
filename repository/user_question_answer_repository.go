package repository

import (
	"context"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"gorm.io/gorm"
)

type UserQuestionAnswerRepository interface {
	FindByID(ctx context.Context, tx *gorm.DB, id int) (*entities.UserQuestionAnswer, error)
	FindByUserID(ctx context.Context, tx *gorm.DB, userID string) ([]entities.UserQuestionAnswer, error)
	FindSortedByUserID(ctx context.Context, tx *gorm.DB, userID string) ([]entities.UserQuestionAnswer, error)
	FindByQuestionID(ctx context.Context, tx *gorm.DB, questionID int) ([]entities.UserQuestionAnswer, error)
	FindByUserAndQuestionID(ctx context.Context, tx *gorm.DB, userID string, questionID int) (*entities.UserQuestionAnswer, error)
	Create(ctx context.Context, tx *gorm.DB, answer *entities.UserQuestionAnswer) error
	CreateMany(ctx context.Context, tx *gorm.DB, answers []entities.UserQuestionAnswer) error
	Update(ctx context.Context, tx *gorm.DB, answer *entities.UserQuestionAnswer) error
	Delete(ctx context.Context, tx *gorm.DB, id int) error
	DeleteByUserID(ctx context.Context, tx *gorm.DB, userID string) error
	FindAllByUserID(ctx context.Context, tx *gorm.DB, userID string) ([]entities.UserQuestionAnswer, error)
}

type userQuestionAnswerRepository struct {
	db *gorm.DB
}

func NewUserQuestionAnswerRepository(db *gorm.DB) UserQuestionAnswerRepository {
	return &userQuestionAnswerRepository{
		db: db,
	}
}

func (r *userQuestionAnswerRepository) FindByID(ctx context.Context, tx *gorm.DB, id int) (*entities.UserQuestionAnswer, error) {
	var answer entities.UserQuestionAnswer
	if err := tx.Where("id = ? AND is_deleted = false", id).First(&answer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &answer, nil
}

func (r *userQuestionAnswerRepository) FindByUserID(ctx context.Context, tx *gorm.DB, userID string) ([]entities.UserQuestionAnswer, error) {
	var answers []entities.UserQuestionAnswer
	if err := tx.Where("user_id = ? AND is_deleted = false", userID).Find(&answers).Error; err != nil {
		return nil, err
	}
	return answers, nil
}

func (r *userQuestionAnswerRepository) FindSortedByUserID(ctx context.Context, tx *gorm.DB, userID string) ([]entities.UserQuestionAnswer, error) {
	var answers []entities.UserQuestionAnswer
	if err := tx.Where("user_id = ? AND is_deleted = false", userID).Order("created_on asc").Find(&answers).Error; err != nil {
		return nil, err
	}
	return answers, nil
}

func (r *userQuestionAnswerRepository) FindByQuestionID(ctx context.Context, tx *gorm.DB, questionID int) ([]entities.UserQuestionAnswer, error) {
	var answers []entities.UserQuestionAnswer
	if err := tx.Where("question_master_id = ? AND is_deleted = false", questionID).Find(&answers).Error; err != nil {
		return nil, err
	}
	return answers, nil
}

func (r *userQuestionAnswerRepository) FindByUserAndQuestionID(ctx context.Context, tx *gorm.DB, userID string, questionID int) (*entities.UserQuestionAnswer, error) {
	var answer entities.UserQuestionAnswer
	if err := tx.Where("user_id = ? AND question_master_id = ? AND is_deleted = false", userID, questionID).First(&answer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &answer, nil
}

func (r *userQuestionAnswerRepository) Create(ctx context.Context, tx *gorm.DB, answer *entities.UserQuestionAnswer) error {
	return tx.Create(answer).Error
}

func (r *userQuestionAnswerRepository) CreateMany(ctx context.Context, tx *gorm.DB, answers []entities.UserQuestionAnswer) error {
	return tx.Create(&answers).Error
}

func (r *userQuestionAnswerRepository) Update(ctx context.Context, tx *gorm.DB, answer *entities.UserQuestionAnswer) error {
	return tx.Save(answer).Error
}

func (r *userQuestionAnswerRepository) Delete(ctx context.Context, tx *gorm.DB, id int) error {
	return tx.Model(&entities.UserQuestionAnswer{}).Where("id = ?", id).Update("is_deleted", true).Error
}

func (r *userQuestionAnswerRepository) DeleteByUserID(ctx context.Context, tx *gorm.DB, userID string) error {
	return tx.Model(&entities.UserQuestionAnswer{}).Where("user_id = ?", userID).Update("is_deleted", true).Error
}

func (r *userQuestionAnswerRepository) FindAllByUserID(ctx context.Context, tx *gorm.DB, userID string) ([]entities.UserQuestionAnswer, error) {
	var answers []entities.UserQuestionAnswer
	if err := tx.Where("user_id = ? AND is_deleted = false", userID).Find(&answers).Error; err != nil {
		return nil, err
	}
	return answers, nil
}


