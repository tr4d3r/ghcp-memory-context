package storage

import (
	"errors"
	"fmt"
)

// Common storage errors
var (
	// ErrNotFound is returned when a requested entity doesn't exist
	ErrNotFound = errors.New("entity not found")

	// ErrAlreadyExists is returned when trying to create an entity that already exists
	ErrAlreadyExists = errors.New("entity already exists")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")

	// ErrDatabaseClosed is returned when attempting operations on a closed database
	ErrDatabaseClosed = errors.New("database connection is closed")

	// ErrTransactionClosed is returned when attempting operations on a closed transaction
	ErrTransactionClosed = errors.New("transaction is closed")

	// ErrConstraintViolation is returned when a database constraint is violated
	ErrConstraintViolation = errors.New("constraint violation")

	// ErrConcurrentUpdate is returned when a concurrent update conflict occurs
	ErrConcurrentUpdate = errors.New("concurrent update conflict")

	// ErrMigrationFailed is returned when a database migration fails
	ErrMigrationFailed = errors.New("migration failed")

	// ErrUnsupportedOperation is returned when an operation is not supported
	ErrUnsupportedOperation = errors.New("unsupported operation")
)

// StorageError wraps storage-specific errors with additional context
type StorageError struct {
	Op   string // Operation that failed
	Type string // Type of entity (e.g., "task", "context", "session")
	ID   string // Entity ID if applicable
	Err  error  // Underlying error
}

// Error implements the error interface
func (e *StorageError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("storage: %s %s %s: %v", e.Op, e.Type, e.ID, e.Err)
	}
	return fmt.Sprintf("storage: %s %s: %v", e.Op, e.Type, e.Err)
}

// Unwrap returns the underlying error
func (e *StorageError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches the target error
func (e *StorageError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

// NewStorageError creates a new storage error
func NewStorageError(op, entityType, id string, err error) *StorageError {
	return &StorageError{
		Op:   op,
		Type: entityType,
		ID:   id,
		Err:  err,
	}
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExists checks if an error is an already exists error
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// IsInvalidInput checks if an error is an invalid input error
func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsConstraintViolation checks if an error is a constraint violation
func IsConstraintViolation(err error) bool {
	return errors.Is(err, ErrConstraintViolation)
}
