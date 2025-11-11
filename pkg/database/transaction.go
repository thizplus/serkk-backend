package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// TxFunc is a function that executes within a transaction
type TxFunc func(*gorm.DB) error

// WithTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, the transaction is committed
func WithTransaction(db *gorm.DB, fn TxFunc) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// WithTransactionContext executes a function within a transaction with context
func WithTransactionContext(ctx context.Context, db *gorm.DB, fn TxFunc) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// TransactionManager manages database transactions
type TransactionManager struct {
	db *gorm.DB
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// Execute runs a function within a transaction
func (tm *TransactionManager) Execute(fn TxFunc) error {
	return WithTransaction(tm.db, fn)
}

// ExecuteWithContext runs a function within a transaction with context
func (tm *TransactionManager) ExecuteWithContext(ctx context.Context, fn TxFunc) error {
	return WithTransactionContext(ctx, tm.db, fn)
}

// BeginTx explicitly begins a new transaction and returns it
// Useful when you need more control over the transaction
func (tm *TransactionManager) BeginTx(ctx context.Context) *gorm.DB {
	return tm.db.WithContext(ctx).Begin()
}

// CommitTx commits the transaction
func (tm *TransactionManager) CommitTx(tx *gorm.DB) error {
	return tx.Commit().Error
}

// RollbackTx rolls back the transaction
func (tm *TransactionManager) RollbackTx(tx *gorm.DB) error {
	return tx.Rollback().Error
}

// SavepointTx creates a savepoint within a transaction
func (tm *TransactionManager) SavepointTx(tx *gorm.DB, name string) error {
	return tx.SavePoint(name).Error
}

// RollbackToSavepoint rolls back to a specific savepoint
func (tm *TransactionManager) RollbackToSavepoint(tx *gorm.DB, name string) error {
	return tx.RollbackTo(name).Error
}

// TransactionError represents a transaction error with additional context
type TransactionError struct {
	Operation string
	Err       error
}

func (e *TransactionError) Error() string {
	return fmt.Sprintf("transaction error in %s: %v", e.Operation, e.Err)
}

// NewTransactionError creates a new transaction error
func NewTransactionError(operation string, err error) *TransactionError {
	return &TransactionError{
		Operation: operation,
		Err:       err,
	}
}

// IsTransactionError checks if an error is a transaction error
func IsTransactionError(err error) bool {
	_, ok := err.(*TransactionError)
	return ok
}
