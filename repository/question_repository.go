package repository

import (
	"context"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"gorm.io/gorm"
)

type QuestionRepository interface {
	GenericRepository[entities.QuestionMaster]
	FindActiveQuestions(ctx context.Context, db *gorm.DB, limit, offset int) ([]entities.QuestionMaster, error)
	FindByLanguageID(ctx context.Context, db *gorm.DB, languageID int, limit, offset int) ([]entities.QuestionMaster, error)
	FindByIDTx(ctx context.Context, tx *gorm.DB, id int) (*entities.QuestionMaster, error)
	FindActive(ctx context.Context, tx *gorm.DB) ([]entities.QuestionMaster, error)
	FindByQuestionTextAndLanguageID(ctx context.Context, tx *gorm.DB, questionText string, languageID int) (*entities.QuestionMaster, error)
}

type questionRepository struct {
	*GormRepository[entities.QuestionMaster]
}

func NewQuestionRepository() QuestionRepository {
	return &questionRepository{
		GormRepository: NewGormRepository[entities.QuestionMaster](),
	}
}

func (r *questionRepository) FindActiveQuestions(ctx context.Context, db *gorm.DB, limit, offset int) ([]entities.QuestionMaster, error) {
	var questions []entities.QuestionMaster
	query := db.WithContext(ctx).Where("is_active = ? AND is_deleted = ?", true, false)
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}
	if err := query.Find(&questions).Error; err != nil {
		return nil, err
	}
	return questions, nil
}

func (r *questionRepository) FindByLanguageID(ctx context.Context, db *gorm.DB, languageID int, limit, offset int) ([]entities.QuestionMaster, error) {
	var questions []entities.QuestionMaster
	query := db.WithContext(ctx).Where("language_id = ? AND is_active = ? AND is_deleted = ?", languageID, true, false)
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}
	if err := query.Find(&questions).Error; err != nil {
		return nil, err
	}
	return questions, nil
}

func (r *questionRepository) FindByIDTx(ctx context.Context, tx *gorm.DB, id int) (*entities.QuestionMaster, error) {
	var question entities.QuestionMaster
	if err := tx.Where("id = ? AND is_deleted = false", id).First(&question).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &question, nil
}

func (r *questionRepository) FindActive(ctx context.Context, tx *gorm.DB) ([]entities.QuestionMaster, error) {
	var questions []entities.QuestionMaster
	if err := tx.Where("is_active = true AND is_deleted = false AND profile_only = true").Find(&questions).Error; err != nil {
		return nil, err
	}
	return questions, nil
}

func (r *questionRepository) FindByQuestionTextAndLanguageID(ctx context.Context, tx *gorm.DB, questionText string, languageID int) (*entities.QuestionMaster, error) {
	var question entities.QuestionMaster
	if err := tx.Where("question_text = ? AND language_id = ? AND is_deleted = false", questionText, languageID).First(&question).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &question, nil
}
