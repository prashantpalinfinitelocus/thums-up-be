package repository

import (
	"context"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"gorm.io/gorm"
)

type OptionMasterRepository interface {
	FindByID(ctx context.Context, tx *gorm.DB, id int) (*entities.OptionMaster, error)
	FindByQuestionID(ctx context.Context, tx *gorm.DB, questionID int) ([]entities.OptionMaster, error)
	FindActiveByQuestionID(ctx context.Context, tx *gorm.DB, questionID int) ([]entities.OptionMaster, error)
	Create(ctx context.Context, tx *gorm.DB, option *entities.OptionMaster) error
	CreateMany(ctx context.Context, tx *gorm.DB, options []entities.OptionMaster) error
	Update(ctx context.Context, tx *gorm.DB, option *entities.OptionMaster) error
	Delete(ctx context.Context, tx *gorm.DB, id int) error
	DeleteByQuestionID(ctx context.Context, tx *gorm.DB, questionID int) error
	FindByQuestionIDs(ctx context.Context, tx *gorm.DB, questionIDs []int) ([]entities.OptionMaster, error)
}

type optionMasterRepository struct {
	db *gorm.DB
}

func NewOptionMasterRepository(db *gorm.DB) OptionMasterRepository {
	return &optionMasterRepository{
		db: db,
	}
}

func (r *optionMasterRepository) FindByID(ctx context.Context, tx *gorm.DB, id int) (*entities.OptionMaster, error) {
	var option entities.OptionMaster
	if err := tx.Where("id = ? AND is_deleted = false", id).First(&option).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &option, nil
}

func (r *optionMasterRepository) FindByQuestionID(ctx context.Context, tx *gorm.DB, questionID int) ([]entities.OptionMaster, error) {
	var options []entities.OptionMaster
	if err := tx.Where("question_master_id = ? AND is_deleted = false", questionID).
		Order("display_order ASC").
		Find(&options).Error; err != nil {
		return nil, err
	}
	return options, nil
}

func (r *optionMasterRepository) FindActiveByQuestionID(ctx context.Context, tx *gorm.DB, questionID int) ([]entities.OptionMaster, error) {
	var options []entities.OptionMaster
	if err := tx.Where("question_master_id = ? AND is_active = true AND is_deleted = false", questionID).
		Order("display_order ASC").
		Find(&options).Error; err != nil {
		return nil, err
	}
	return options, nil
}

func (r *optionMasterRepository) Create(ctx context.Context, tx *gorm.DB, option *entities.OptionMaster) error {
	return tx.Create(option).Error
}

func (r *optionMasterRepository) CreateMany(ctx context.Context, tx *gorm.DB, options []entities.OptionMaster) error {
	return tx.Create(&options).Error
}

func (r *optionMasterRepository) Update(ctx context.Context, tx *gorm.DB, option *entities.OptionMaster) error {
	return tx.Save(option).Error
}

func (r *optionMasterRepository) Delete(ctx context.Context, tx *gorm.DB, id int) error {
	return tx.Model(&entities.OptionMaster{}).Where("id = ?", id).Update("is_deleted", true).Error
}

func (r *optionMasterRepository) DeleteByQuestionID(ctx context.Context, tx *gorm.DB, questionID int) error {
	return tx.Model(&entities.OptionMaster{}).Where("question_master_id = ?", questionID).Update("is_deleted", true).Error
}

func (r *optionMasterRepository) FindByQuestionIDs(ctx context.Context, tx *gorm.DB, questionIDs []int) ([]entities.OptionMaster, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	var options []entities.OptionMaster
	err := db.WithContext(ctx).Where("question_master_id IN ?", questionIDs).Find(&options).Error
	return options, err
}
