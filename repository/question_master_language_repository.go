package repository

import (
	"context"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"gorm.io/gorm"
)

type QuestionMasterLanguageRepository interface {
	FindByID(ctx context.Context, tx *gorm.DB, id int) (*entities.QuestionMasterLanguage, error)
	FindByQuestionMasterID(ctx context.Context, tx *gorm.DB, questionMasterID int) ([]entities.QuestionMasterLanguage, error)
	FindByLanguageID(ctx context.Context, tx *gorm.DB, languageID int) ([]entities.QuestionMasterLanguage, error)
	FindByQuestionMasterIDAndLanguageID(ctx context.Context, tx *gorm.DB, questionMasterID int, languageID int) (*entities.QuestionMasterLanguage, error)
	FindActiveByLanguageID(ctx context.Context, tx *gorm.DB, languageID int) ([]entities.QuestionMasterLanguage, error)
	FindByQuestionTextAndLanguageID(ctx context.Context, tx *gorm.DB, questionText string, languageID int) (*entities.QuestionMasterLanguage, error)
	Create(ctx context.Context, tx *gorm.DB, questionLanguage *entities.QuestionMasterLanguage) error
	Update(ctx context.Context, tx *gorm.DB, questionLanguage *entities.QuestionMasterLanguage) error
	Delete(ctx context.Context, tx *gorm.DB, id int) error
}

type questionMasterLanguageRepository struct {
	db *gorm.DB
}

func NewQuestionMasterLanguageRepository(db *gorm.DB) QuestionMasterLanguageRepository {
	return &questionMasterLanguageRepository{
		db: db,
	}
}

func (r *questionMasterLanguageRepository) FindByID(ctx context.Context, tx *gorm.DB, id int) (*entities.QuestionMasterLanguage, error) {
	var questionLanguage entities.QuestionMasterLanguage
	if err := tx.Where("id = ? AND is_deleted = false", id).First(&questionLanguage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &questionLanguage, nil
}

func (r *questionMasterLanguageRepository) FindByQuestionMasterID(ctx context.Context, tx *gorm.DB, questionMasterID int) ([]entities.QuestionMasterLanguage, error) {
	var questionLanguages []entities.QuestionMasterLanguage
	if err := tx.Where("question_master_id = ? AND is_deleted = false", questionMasterID).Find(&questionLanguages).Error; err != nil {
		return nil, err
	}
	return questionLanguages, nil
}

func (r *questionMasterLanguageRepository) FindByLanguageID(ctx context.Context, tx *gorm.DB, languageID int) ([]entities.QuestionMasterLanguage, error) {
	var questionLanguages []entities.QuestionMasterLanguage
	if err := tx.Where("language_id = ? AND is_deleted = false", languageID).Find(&questionLanguages).Error; err != nil {
		return nil, err
	}
	return questionLanguages, nil
}

func (r *questionMasterLanguageRepository) FindByQuestionMasterIDAndLanguageID(ctx context.Context, tx *gorm.DB, questionMasterID int, languageID int) (*entities.QuestionMasterLanguage, error) {
	var questionLanguage entities.QuestionMasterLanguage
	if err := tx.Where("question_master_id = ? AND language_id = ? AND is_deleted = false", questionMasterID, languageID).First(&questionLanguage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &questionLanguage, nil
}

func (r *questionMasterLanguageRepository) FindActiveByLanguageID(ctx context.Context, tx *gorm.DB, languageID int) ([]entities.QuestionMasterLanguage, error) {
	var questionLanguages []entities.QuestionMasterLanguage
	if err := tx.Where("language_id = ? AND is_active = true AND is_deleted = false", languageID).Find(&questionLanguages).Error; err != nil {
		return nil, err
	}
	return questionLanguages, nil
}

func (r *questionMasterLanguageRepository) FindByQuestionTextAndLanguageID(ctx context.Context, tx *gorm.DB, questionText string, languageID int) (*entities.QuestionMasterLanguage, error) {
	var questionLanguage entities.QuestionMasterLanguage
	if err := tx.Where("question_text = ? AND language_id = ? AND is_deleted = false", questionText, languageID).First(&questionLanguage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &questionLanguage, nil
}

func (r *questionMasterLanguageRepository) Create(ctx context.Context, tx *gorm.DB, questionLanguage *entities.QuestionMasterLanguage) error {
	return tx.Create(questionLanguage).Error
}

func (r *questionMasterLanguageRepository) Update(ctx context.Context, tx *gorm.DB, questionLanguage *entities.QuestionMasterLanguage) error {
	return tx.Save(questionLanguage).Error
}

func (r *questionMasterLanguageRepository) Delete(ctx context.Context, tx *gorm.DB, id int) error {
	return tx.Model(&entities.QuestionMasterLanguage{}).Where("id = ?", id).Update("is_deleted", true).Error
}


