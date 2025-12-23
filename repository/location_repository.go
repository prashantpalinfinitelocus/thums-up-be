package repository

import (
	"context"

	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"gorm.io/gorm"
)

type StateRepository interface {
	FindByName(ctx context.Context, tx *gorm.DB, name string) (*entities.State, error)
	FindByID(ctx context.Context, tx *gorm.DB, id int) (*entities.State, error)
	FindByIDs(ctx context.Context, tx *gorm.DB, ids []int) ([]entities.State, error)
}

type CityRepository interface {
	FindByNameAndStateID(ctx context.Context, tx *gorm.DB, name string, stateID int) (*entities.City, error)
	FindByID(ctx context.Context, tx *gorm.DB, id int) (*entities.City, error)
	FindByIDs(ctx context.Context, tx *gorm.DB, ids []int) ([]entities.City, error)
}

type PinCodeRepository interface {
	FindByPincodeAndCityID(ctx context.Context, tx *gorm.DB, pincode int, cityID int) (*entities.PinCode, error)
	FindByID(ctx context.Context, tx *gorm.DB, id int) (*entities.PinCode, error)
	IsDeliverable(ctx context.Context, tx *gorm.DB, pincode int, cityID int) (bool, error)
	FindByIDs(ctx context.Context, tx *gorm.DB, ids []int) ([]entities.PinCode, error)
	FindByPincode(ctx context.Context, tx *gorm.DB, pincode int) (*entities.PinCode, error)
}

type stateRepository struct {
	db *gorm.DB
}

type cityRepository struct {
	db *gorm.DB
}

type pinCodeRepository struct {
	db *gorm.DB
}

func NewStateRepository() StateRepository {
	return &stateRepository{}
}

func NewCityRepository() CityRepository {
	return &cityRepository{}
}

func NewPinCodeRepository() PinCodeRepository {
	return &pinCodeRepository{}
}

func (r *stateRepository) FindByName(ctx context.Context, tx *gorm.DB, name string) (*entities.State, error) {
	var state entities.State
	err := tx.WithContext(ctx).Where("name = ?", name).First(&state).Error
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *stateRepository) FindByID(ctx context.Context, tx *gorm.DB, id int) (*entities.State, error) {
	var state entities.State
	err := tx.WithContext(ctx).Where("id = ?", id).First(&state).Error
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *stateRepository) FindByIDs(ctx context.Context, tx *gorm.DB, ids []int) ([]entities.State, error) {
	var states []entities.State
	err := tx.WithContext(ctx).Where("id IN ?", ids).Find(&states).Error
	if err != nil {
		return nil, err
	}
	return states, nil
}

func (r *cityRepository) FindByNameAndStateID(ctx context.Context, tx *gorm.DB, name string, stateID int) (*entities.City, error) {
	var city entities.City
	err := tx.WithContext(ctx).Where("name = ? AND state_id = ?", name, stateID).First(&city).Error
	if err != nil {
		return nil, err
	}
	return &city, nil
}

func (r *cityRepository) FindByID(ctx context.Context, tx *gorm.DB, id int) (*entities.City, error) {
	var city entities.City
	err := tx.WithContext(ctx).Where("id = ?", id).First(&city).Error
	if err != nil {
		return nil, err
	}
	return &city, nil
}

func (r *cityRepository) FindByIDs(ctx context.Context, tx *gorm.DB, ids []int) ([]entities.City, error) {
	var cities []entities.City
	err := tx.WithContext(ctx).Where("id IN ?", ids).Find(&cities).Error
	if err != nil {
		return nil, err
	}
	return cities, nil
}

func (r *pinCodeRepository) FindByPincodeAndCityID(ctx context.Context, tx *gorm.DB, pincode int, cityID int) (*entities.PinCode, error) {
	var pinCode entities.PinCode
	err := tx.WithContext(ctx).Table("pin_codes").Where("pincode = ? AND city_id = ?", pincode, cityID).First(&pinCode).Error
	if err != nil {
		return nil, err
	}
	return &pinCode, nil
}

func (r *pinCodeRepository) FindByID(ctx context.Context, tx *gorm.DB, id int) (*entities.PinCode, error) {
	var pinCode entities.PinCode
	err := tx.WithContext(ctx).Table("pin_codes").Where("id = ?", id).First(&pinCode).Error
	if err != nil {
		return nil, err
	}
	return &pinCode, nil
}

func (r *pinCodeRepository) IsDeliverable(ctx context.Context, tx *gorm.DB, pincode int, cityID int) (bool, error) {
	var count int64
	err := tx.WithContext(ctx).Table("pin_codes").
		Where("pincode = ? AND city_id = ? AND is_deliverable = true", pincode, cityID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *pinCodeRepository) FindByIDs(ctx context.Context, tx *gorm.DB, ids []int) ([]entities.PinCode, error) {
	var pincodes []entities.PinCode
	err := tx.WithContext(ctx).Table("pin_codes").Where("id IN ?", ids).Find(&pincodes).Error
	if err != nil {
		return nil, err
	}
	return pincodes, nil
}

func (r *pinCodeRepository) FindByPincode(ctx context.Context, tx *gorm.DB, pincode int) (*entities.PinCode, error) {
	var pinCode entities.PinCode
	err := tx.WithContext(ctx).Table("pin_codes").Where("pincode = ?", pincode).First(&pinCode).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &pinCode, nil
}
