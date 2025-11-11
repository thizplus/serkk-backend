// +build cgo

package database

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TestModel struct {
	ID   uint   `gorm:"primaryKey"`
	Name string
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Skip("CGO not enabled, skipping database tests")
		return nil
	}

	// Auto migrate
	err = db.AutoMigrate(&TestModel{})
	assert.NoError(t, err)

	return db
}

func TestWithTransaction_Success(t *testing.T) {
	db := setupTestDB(t)

	// Act
	err := WithTransaction(db, func(tx *gorm.DB) error {
		return tx.Create(&TestModel{Name: "Test"}).Error
	})

	// Assert
	assert.NoError(t, err)

	// Verify data was committed
	var count int64
	db.Model(&TestModel{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestWithTransaction_Rollback(t *testing.T) {
	db := setupTestDB(t)

	// Act
	err := WithTransaction(db, func(tx *gorm.DB) error {
		tx.Create(&TestModel{Name: "Test"})
		return errors.New("rollback")
	})

	// Assert
	assert.Error(t, err)

	// Verify data was rolled back
	var count int64
	db.Model(&TestModel{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestWithTransactionContext_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// Act
	err := WithTransactionContext(ctx, db, func(tx *gorm.DB) error {
		return tx.Create(&TestModel{Name: "Test"}).Error
	})

	// Assert
	assert.NoError(t, err)

	var count int64
	db.Model(&TestModel{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestTransactionManager_Execute(t *testing.T) {
	db := setupTestDB(t)
	tm := NewTransactionManager(db)

	// Act
	err := tm.Execute(func(tx *gorm.DB) error {
		return tx.Create(&TestModel{Name: "Test"}).Error
	})

	// Assert
	assert.NoError(t, err)

	var count int64
	db.Model(&TestModel{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestTransactionManager_ExecuteWithContext(t *testing.T) {
	db := setupTestDB(t)
	tm := NewTransactionManager(db)
	ctx := context.Background()

	// Act
	err := tm.ExecuteWithContext(ctx, func(tx *gorm.DB) error {
		return tx.Create(&TestModel{Name: "Test"}).Error
	})

	// Assert
	assert.NoError(t, err)
}

func TestTransactionManager_BeginCommit(t *testing.T) {
	db := setupTestDB(t)
	tm := NewTransactionManager(db)
	ctx := context.Background()

	// Begin transaction
	tx := tm.BeginTx(ctx)
	assert.NotNil(t, tx)

	// Insert data
	err := tx.Create(&TestModel{Name: "Test"}).Error
	assert.NoError(t, err)

	// Commit
	err = tm.CommitTx(tx)
	assert.NoError(t, err)

	// Verify
	var count int64
	db.Model(&TestModel{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestTransactionManager_BeginRollback(t *testing.T) {
	db := setupTestDB(t)
	tm := NewTransactionManager(db)
	ctx := context.Background()

	// Begin transaction
	tx := tm.BeginTx(ctx)

	// Insert data
	tx.Create(&TestModel{Name: "Test"})

	// Rollback
	err := tm.RollbackTx(tx)
	assert.NoError(t, err)

	// Verify - should be 0
	var count int64
	db.Model(&TestModel{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestTransactionError(t *testing.T) {
	err := NewTransactionError("create", errors.New("test error"))

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "create")
	assert.Contains(t, err.Error(), "test error")
	assert.True(t, IsTransactionError(err))
}
