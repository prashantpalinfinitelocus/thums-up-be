package repository

import (
	"context"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"gorm.io/gorm"
)

type QuestionRepository interface {
	GenericRepository[entities.QuestionMaster]
	FindActiveQuestions(ctx context.Context, db *gorm.DB) ([]entities.QuestionMaster, error)
	FindByLanguageID(ctx context.Context, db *gorm.DB, languageID int) ([]entities.QuestionMaster, error)
}

type questionRepository struct {
	*GormRepository[entities.QuestionMaster]
}

func NewQuestionRepository() QuestionRepository {
	return &questionRepository{
		GormRepository: NewGormRepository[entities.QuestionMaster](),
	}
}

func (r *questionRepository) FindActiveQuestions(ctx context.Context, db *gorm.DB) ([]entities.QuestionMaster, error) {
	var questions []entities.QuestionMaster
	if err := db.WithContext(ctx).Where("is_active = ? AND is_deleted = ?", true, false).Find(&questions).Error; err != nil {
		return nil, err
	}
	return questions, nil
}

func (r *questionRepository) FindByLanguageID(ctx context.Context, db *gorm.DB, languageID int) ([]entities.QuestionMaster, error) {
	var questions []entities.QuestionMaster
	if err := db.WithContext(ctx).Where("language_id = ? AND is_active = ? AND is_deleted = ?", languageID, true, false).Find(&questions).Error; err != nil {
		return nil, err
	}
	return questions, nil
}
