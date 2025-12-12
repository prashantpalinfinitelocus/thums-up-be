package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GenericRepository[T any] interface {
	Create(ctx context.Context, db *gorm.DB, entity *T) error
	Update(ctx context.Context, db *gorm.DB, entity *T) error
	Delete(ctx context.Context, db *gorm.DB, id interface{}) error
	FindByID(ctx context.Context, db *gorm.DB, id interface{}) (*T, error)
	FindById(ctx context.Context, db *gorm.DB, id uuid.UUID) (*T, error)
	FindAll(ctx context.Context, db *gorm.DB) ([]T, error)
	FindByCondition(ctx context.Context, db *gorm.DB, condition interface{}, args ...interface{}) ([]T, error)
	FindWithConditions(ctx context.Context, db *gorm.DB, conditions map[string]interface{}) ([]T, error)
	UpdateFields(ctx context.Context, db *gorm.DB, id interface{}, fields map[string]interface{}) error
}

type GormRepository[T any] struct{}

func NewGormRepository[T any]() *GormRepository[T] {
	return &GormRepository[T]{}
}

func (r *GormRepository[T]) Create(ctx context.Context, db *gorm.DB, entity *T) error {
	return db.WithContext(ctx).Create(entity).Error
}

func (r *GormRepository[T]) Update(ctx context.Context, db *gorm.DB, entity *T) error {
	return db.WithContext(ctx).Save(entity).Error
}

func (r *GormRepository[T]) Delete(ctx context.Context, db *gorm.DB, id interface{}) error {
	var entity T
	return db.WithContext(ctx).Delete(&entity, id).Error
}

func (r *GormRepository[T]) FindByID(ctx context.Context, db *gorm.DB, id interface{}) (*T, error) {
	var entity T
	if err := db.WithContext(ctx).First(&entity, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

func (r *GormRepository[T]) FindById(ctx context.Context, db *gorm.DB, id uuid.UUID) (*T, error) {
	var entity T
	if err := db.WithContext(ctx).First(&entity, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

func (r *GormRepository[T]) FindAll(ctx context.Context, db *gorm.DB) ([]T, error) {
	var entities []T
	if err := db.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *GormRepository[T]) FindByCondition(ctx context.Context, db *gorm.DB, condition interface{}, args ...interface{}) ([]T, error) {
	var entities []T
	if err := db.WithContext(ctx).Where(condition, args...).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *GormRepository[T]) FindWithConditions(ctx context.Context, db *gorm.DB, conditions map[string]interface{}) ([]T, error) {
	var entities []T
	query := db.WithContext(ctx)
	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}
	if err := query.Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *GormRepository[T]) UpdateFields(ctx context.Context, db *gorm.DB, id interface{}, fields map[string]interface{}) error {
	var entity T
	return db.WithContext(ctx).Model(&entity).Where("id = ?", id).Updates(fields).Error
}
