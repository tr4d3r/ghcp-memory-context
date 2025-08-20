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

	// ErrFileLocked is returned when a file is locked by another process
	ErrFileLocked = errors.New("file is locked")

	// ErrConcurrentUpdate is returned when a concurrent update conflict occurs
	ErrConcurrentUpdate = errors.New("concurrent update conflict")

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
