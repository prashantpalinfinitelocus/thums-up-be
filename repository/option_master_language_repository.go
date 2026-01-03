package repository

import (
	"context"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"gorm.io/gorm"
)

type OptionMasterLanguageRepository interface {
	FindByID(ctx context.Context, tx *gorm.DB, id int) (*entities.OptionMasterLanguage, error)
	FindByOptionMasterID(ctx context.Context, tx *gorm.DB, optionMasterID int) ([]entities.OptionMasterLanguage, error)
	FindByLanguageID(ctx context.Context, tx *gorm.DB, languageID int) ([]entities.OptionMasterLanguage, error)
	FindByOptionMasterIDAndLanguageID(ctx context.Context, tx *gorm.DB, optionMasterID int, languageID int) (*entities.OptionMasterLanguage, error)
	FindActiveByLanguageID(ctx context.Context, tx *gorm.DB, languageID int) ([]entities.OptionMasterLanguage, error)
	FindByOptionMasterIDsAndLanguageID(ctx context.Context, tx *gorm.DB, optionMasterIDs []int, languageID int) ([]entities.OptionMasterLanguage, error)
	Create(ctx context.Context, tx *gorm.DB, optionLanguage *entities.OptionMasterLanguage) error
	Update(ctx context.Context, tx *gorm.DB, optionLanguage *entities.OptionMasterLanguage) error
	Delete(ctx context.Context, tx *gorm.DB, id int) error
}

type optionMasterLanguageRepository struct {
	db *gorm.DB
}

func NewOptionMasterLanguageRepository(db *gorm.DB) OptionMasterLanguageRepository {
	return &optionMasterLanguageRepository{
		db: db,
	}
}

func (r *optionMasterLanguageRepository) FindByID(ctx context.Context, tx *gorm.DB, id int) (*entities.OptionMasterLanguage, error) {
	var optionLanguage entities.OptionMasterLanguage
	if err := tx.Where("id = ? AND is_deleted = false", id).First(&optionLanguage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &optionLanguage, nil
}

func (r *optionMasterLanguageRepository) FindByOptionMasterID(ctx context.Context, tx *gorm.DB, optionMasterID int) ([]entities.OptionMasterLanguage, error) {
	var optionLanguages []entities.OptionMasterLanguage
	if err := tx.Where("option_master_id = ? AND is_deleted = false", optionMasterID).Find(&optionLanguages).Error; err != nil {
		return nil, err
	}
	return optionLanguages, nil
}

func (r *optionMasterLanguageRepository) FindByLanguageID(ctx context.Context, tx *gorm.DB, languageID int) ([]entities.OptionMasterLanguage, error) {
	var optionLanguages []entities.OptionMasterLanguage
	if err := tx.Where("language_id = ? AND is_deleted = false", languageID).Find(&optionLanguages).Error; err != nil {
		return nil, err
	}
	return optionLanguages, nil
}

func (r *optionMasterLanguageRepository) FindByOptionMasterIDAndLanguageID(ctx context.Context, tx *gorm.DB, optionMasterID int, languageID int) (*entities.OptionMasterLanguage, error) {
	var optionLanguage entities.OptionMasterLanguage
	if err := tx.Where("option_master_id = ? AND language_id = ? AND is_deleted = false", optionMasterID, languageID).First(&optionLanguage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &optionLanguage, nil
}

func (r *optionMasterLanguageRepository) FindActiveByLanguageID(ctx context.Context, tx *gorm.DB, languageID int) ([]entities.OptionMasterLanguage, error) {
	var optionLanguages []entities.OptionMasterLanguage
	if err := tx.Where("language_id = ? AND is_active = true AND is_deleted = false", languageID).Find(&optionLanguages).Error; err != nil {
		return nil, err
	}
	return optionLanguages, nil
}

func (r *optionMasterLanguageRepository) FindByOptionMasterIDsAndLanguageID(ctx context.Context, tx *gorm.DB, optionMasterIDs []int, languageID int) ([]entities.OptionMasterLanguage, error) {
	var optionLanguages []entities.OptionMasterLanguage
	if err := tx.Where("option_master_id IN ? AND language_id = ? AND is_deleted = false", optionMasterIDs, languageID).Find(&optionLanguages).Error; err != nil {
		return nil, err
	}
	return optionLanguages, nil
}

func (r *optionMasterLanguageRepository) Create(ctx context.Context, tx *gorm.DB, optionLanguage *entities.OptionMasterLanguage) error {
	return tx.Create(optionLanguage).Error
}

func (r *optionMasterLanguageRepository) Update(ctx context.Context, tx *gorm.DB, optionLanguage *entities.OptionMasterLanguage) error {
	return tx.Save(optionLanguage).Error
}

func (r *optionMasterLanguageRepository) Delete(ctx context.Context, tx *gorm.DB, id int) error {
	return tx.Model(&entities.OptionMasterLanguage{}).Where("id = ?", id).Update("is_deleted", true).Error
}
