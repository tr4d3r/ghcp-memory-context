package storage

import (
	"context"
	"fmt"
)

// Factory is a function that creates a Storage instance
type Factory func(ctx context.Context, config *Config) (Storage, error)

// Registry holds registered storage factories
var registry = make(map[string]Factory)

// Register registers a storage factory for a given driver
func Register(driver string, factory Factory) {
	if factory == nil {
		panic("storage: Register factory is nil")
	}
	if _, dup := registry[driver]; dup {
		panic(fmt.Sprintf("storage: Register called twice for driver %s", driver))
	}
	registry[driver] = factory
}

// New creates a new Storage instance based on the configuration
func New(ctx context.Context, config *Config) (Storage, error) {
	if config == nil {
		return nil, NewStorageError("new", "storage", "", ErrInvalidInput)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Get factory for the driver
	factory, ok := registry[config.Driver]
	if !ok {
		return nil, NewStorageError("new", "storage", "",
			fmt.Errorf("unknown driver: %s", config.Driver))
	}

	// Create storage instance
	storage, err := factory(ctx, config)
	if err != nil {
		return nil, NewStorageError("new", "storage", "", err)
	}

	// Connect to the database
	if err := storage.Connect(ctx); err != nil {
		return nil, NewStorageError("connect", "storage", "", err)
	}

	return storage, nil
}

// MustNew creates a new Storage instance and panics on error
func MustNew(ctx context.Context, config *Config) Storage {
	storage, err := New(ctx, config)
	if err != nil {
		panic(err)
	}
	return storage
}

// Drivers returns a list of registered drivers
func Drivers() []string {
	drivers := make([]string, 0, len(registry))
	for driver := range registry {
		drivers = append(drivers, driver)
	}
	return drivers
}

// IsDriverRegistered checks if a driver is registered
func IsDriverRegistered(driver string) bool {
	_, ok := registry[driver]
	return ok
}
