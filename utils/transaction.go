package utils

import (
	"context"

	"gorm.io/gorm"
)

type TransactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

func (tm *TransactionManager) ExecuteInTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	tx := tm.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (tm *TransactionManager) GetDB() *gorm.DB {
	return tm.db
}

func (tm *TransactionManager) StartTxn() (*gorm.DB, error) {
	tx := tm.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return tx, nil
}

func (tm *TransactionManager) CommitTxn(tx *gorm.DB) error {
	return tx.Commit().Error
}

func (tm *TransactionManager) AbortTxn(tx *gorm.DB) error {
	return tx.Rollback().Error
}

func (tm *TransactionManager) RollbackOnPanic(tx *gorm.DB) {
	if r := recover(); r != nil {
		tx.Rollback()
		panic(r)
	}
}
