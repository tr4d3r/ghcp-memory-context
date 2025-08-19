package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/tr4d3r/ghcp-memory-context/internal/storage"
	"github.com/tr4d3r/ghcp-memory-context/internal/storage/migrations"
	_ "modernc.org/sqlite" // SQLite driver
)

// Driver implements the SQLite storage driver
type Driver struct {
	db               *sql.DB
	config           *storage.Config
	dsn              string
	migrationManager *migrations.Manager
}

// init registers the SQLite driver with the storage factory
func init() {
	storage.Register("sqlite", NewDriver)
}

// NewDriver creates a new SQLite driver instance
func NewDriver(ctx context.Context, config *storage.Config) (storage.Storage, error) {
	if config == nil {
		return nil, storage.NewStorageError("new_driver", "sqlite", "", storage.ErrInvalidInput)
	}

	// Build DSN if not provided
	dsn := config.DSN
	if dsn == "" {
		dsn = buildDSN(config)
	}

	driver := &Driver{
		config: config,
		dsn:    dsn,
	}

	return driver, nil
}

// buildDSN constructs a SQLite DSN from configuration
func buildDSN(config *storage.Config) string {
	// Use DSN if provided directly
	if config.DSN != "" {
		return config.DSN
	}

	// Build default DSN with SQLite parameters
	dsn := "file:data.db"

	// Add query parameters for SQLite configuration
	params := []string{}

	if config.SQLite.WALEnabled {
		params = append(params, "journal_mode=WAL")
	} else if config.SQLite.JournalMode != "" {
		params = append(params, fmt.Sprintf("journal_mode=%s", config.SQLite.JournalMode))
	}

	if config.SQLite.ForeignKeys {
		params = append(params, "foreign_keys=on")
	}

	if config.SQLite.Synchronous != "" {
		params = append(params, fmt.Sprintf("synchronous=%s", config.SQLite.Synchronous))
	}

	if config.SQLite.CacheSize != 0 {
		params = append(params, fmt.Sprintf("cache_size=%d", config.SQLite.CacheSize))
	}

	if config.SQLite.BusyTimeout > 0 {
		params = append(params, fmt.Sprintf("busy_timeout=%d", config.SQLite.BusyTimeout))
	}

	// Append parameters to DSN
	if len(params) > 0 {
		dsn += "?"
		for i, param := range params {
			if i > 0 {
				dsn += "&"
			}
			dsn += param
		}
	}

	return dsn
}

// Connect establishes a connection to the SQLite database
func (d *Driver) Connect(ctx context.Context) error {
	var err error

	// Open database connection
	d.db, err = sql.Open("sqlite", d.dsn)
	if err != nil {
		return storage.NewStorageError("connect", "sqlite", "", err)
	}

	// Configure connection pool
	if d.config.MaxOpenConns > 0 {
		d.db.SetMaxOpenConns(d.config.MaxOpenConns)
	} else {
		// SQLite default: single connection for writes, multiple for reads
		d.db.SetMaxOpenConns(10)
	}

	// SQLite doesn't need many idle connections
	d.db.SetMaxIdleConns(2)
	d.db.SetConnMaxLifetime(time.Hour)

	// Test the connection
	if err := d.db.PingContext(ctx); err != nil {
		d.db.Close()
		return storage.NewStorageError("ping", "sqlite", "", err)
	}

	// Enable foreign keys (if not already set in DSN)
	if d.config.SQLite.ForeignKeys {
		if _, err := d.db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
			d.db.Close()
			return storage.NewStorageError("pragma", "sqlite", "", err)
		}
	}

	// Enable WAL mode (if not already set in DSN)
	if d.config.SQLite.WALEnabled {
		if _, err := d.db.ExecContext(ctx, "PRAGMA journal_mode = WAL"); err != nil {
			d.db.Close()
			return storage.NewStorageError("pragma", "sqlite", "", err)
		}
	}

	// Initialize migration manager
	d.migrationManager = migrations.NewManager(d.db, d.config)

	// Run auto-migrations if enabled
	if err := d.migrationManager.AutoMigrate(ctx); err != nil {
		d.db.Close()
		return storage.NewStorageError("auto_migrate", "sqlite", "", err)
	}

	return nil
}

// Close closes the database connection
func (d *Driver) Close() error {
	if d.db == nil {
		return nil
	}

	err := d.db.Close()
	d.db = nil

	if err != nil {
		return storage.NewStorageError("close", "sqlite", "", err)
	}

	return nil
}

// Ping tests the database connection
func (d *Driver) Ping(ctx context.Context) error {
	if d.db == nil {
		return storage.NewStorageError("ping", "sqlite", "", fmt.Errorf("database not connected"))
	}

	if err := d.db.PingContext(ctx); err != nil {
		return storage.NewStorageError("ping", "sqlite", "", err)
	}

	return nil
}

// BeginTx starts a new database transaction
func (d *Driver) BeginTx(ctx context.Context) (storage.Transaction, error) {
	if d.db == nil {
		return nil, storage.NewStorageError("begin_tx", "sqlite", "", fmt.Errorf("database not connected"))
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, storage.NewStorageError("begin_tx", "sqlite", "", err)
	}

	return &Transaction{
		tx:     tx,
		driver: d,
	}, nil
}

// DB returns the underlying *sql.DB for advanced operations
// This method is not part of the Storage interface but may be useful for testing
func (d *Driver) DB() *sql.DB {
	return d.db
}

// GetConfig returns the driver configuration
func (d *Driver) GetConfig() *storage.Config {
	return d.config
}

// GetDSN returns the data source name
func (d *Driver) GetDSN() string {
	return d.dsn
}

// GetMigrationManager returns the migration manager for advanced migration operations
func (d *Driver) GetMigrationManager() *migrations.Manager {
	return d.migrationManager
}

// MigrateUp migrates to the latest version
func (d *Driver) MigrateUp(ctx context.Context) error {
	if d.migrationManager == nil {
		return storage.NewStorageError("migrate_up", "sqlite", "", fmt.Errorf("migration manager not initialized"))
	}
	return d.migrationManager.Up(ctx)
}

// MigrateDown rolls back one migration
func (d *Driver) MigrateDown(ctx context.Context) error {
	if d.migrationManager == nil {
		return storage.NewStorageError("migrate_down", "sqlite", "", fmt.Errorf("migration manager not initialized"))
	}
	return d.migrationManager.Down(ctx)
}

// MigrationStatus returns the current migration status
func (d *Driver) MigrationStatus(ctx context.Context) ([]*migrations.Migration, error) {
	if d.migrationManager == nil {
		return nil, storage.NewStorageError("migration_status", "sqlite", "", fmt.Errorf("migration manager not initialized"))
	}
	return d.migrationManager.Status(ctx)
}
