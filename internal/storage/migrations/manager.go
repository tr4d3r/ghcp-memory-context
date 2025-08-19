package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/tr4d3r/ghcp-memory-context/internal/storage"
)

// Manager provides high-level migration management
type Manager struct {
	migrator *Migrator
	config   *storage.Config
}

// ManagerOption configures the migration manager
type ManagerOption func(*Manager)

// WithMigrationsFS sets a custom filesystem for migrations
func WithMigrationsFS(fsys fs.FS) ManagerOption {
	return func(m *Manager) {
		m.migrator.migrationsFS = fsys
	}
}

// WithMigrationsPath sets a custom path within the filesystem
func WithMigrationsPath(path string) ManagerOption {
	return func(m *Manager) {
		m.migrator.migrationsPath = path
	}
}

// NewManager creates a new migration manager
func NewManager(db *sql.DB, config *storage.Config, opts ...ManagerOption) *Manager {
	migrator := NewMigrator(MigratorConfig{
		DB:             db,
		TableName:      config.MigrationTable,
		MigrationsFS:   GetEmbeddedFS(),
		MigrationsPath: ".",
	})

	manager := &Manager{
		migrator: migrator,
		config:   config,
	}

	// Apply options
	for _, opt := range opts {
		opt(manager)
	}

	return manager
}

// AutoMigrate runs migrations automatically if enabled in config
func (m *Manager) AutoMigrate(ctx context.Context) error {
	if !m.config.AutoMigrate {
		return nil
	}

	// Initialize migrations table
	if err := m.migrator.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	// Run up migrations
	if err := m.migrator.Up(ctx); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

// Up migrates to the latest version
func (m *Manager) Up(ctx context.Context) error {
	if err := m.migrator.Initialize(ctx); err != nil {
		return err
	}
	return m.migrator.Up(ctx)
}

// UpTo migrates to a specific version
func (m *Manager) UpTo(ctx context.Context, version int) error {
	if err := m.migrator.Initialize(ctx); err != nil {
		return err
	}
	return m.migrator.UpTo(ctx, version)
}

// Down rolls back one migration
func (m *Manager) Down(ctx context.Context) error {
	return m.migrator.Down(ctx)
}

// DownTo rolls back to a specific version
func (m *Manager) DownTo(ctx context.Context, version int) error {
	return m.migrator.DownTo(ctx, version)
}

// Status returns current migration status
func (m *Manager) Status(ctx context.Context) ([]*Migration, error) {
	return m.migrator.Status(ctx)
}

// Validate checks migration integrity
func (m *Manager) Validate(ctx context.Context) error {
	return m.migrator.Validate(ctx)
}

// Reset resets the database (DANGEROUS)
func (m *Manager) Reset(ctx context.Context) error {
	return m.migrator.Reset(ctx)
}

// GetMigrator returns the underlying migrator for advanced operations
func (m *Manager) GetMigrator() *Migrator {
	return m.migrator
}
