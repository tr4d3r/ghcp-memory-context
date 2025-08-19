package storage

import (
	"time"
)

// Config holds configuration for the storage layer
type Config struct {
	// Driver specifies the storage backend (e.g., "sqlite", "postgres")
	Driver string `json:"driver" validate:"required,oneof=sqlite postgres memory"`

	// DSN is the data source name/connection string
	DSN string `json:"dsn" validate:"required"`

	// Connection pool settings
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`

	// Query timeout settings
	QueryTimeout time.Duration `json:"query_timeout"`

	// Migration settings
	MigrationsPath string `json:"migrations_path"`
	AutoMigrate    bool   `json:"auto_migrate"`
	MigrationTable string `json:"migration_table"`

	// Performance settings
	EnableQueryLog     bool          `json:"enable_query_log"`
	SlowQueryThreshold time.Duration `json:"slow_query_threshold"`

	// SQLite specific settings
	SQLite SQLiteConfig `json:"sqlite,omitempty"`

	// PostgreSQL specific settings (future)
	Postgres PostgresConfig `json:"postgres,omitempty"`
}

// SQLiteConfig holds SQLite-specific configuration
type SQLiteConfig struct {
	// Journal mode (DELETE, TRUNCATE, PERSIST, MEMORY, WAL, OFF)
	JournalMode string `json:"journal_mode"`

	// Synchronous mode (OFF, NORMAL, FULL, EXTRA)
	Synchronous string `json:"synchronous"`

	// Cache size in pages (-2000 means 2MB)
	CacheSize int `json:"cache_size"`

	// Enable foreign key constraints
	ForeignKeys bool `json:"foreign_keys"`

	// Busy timeout in milliseconds
	BusyTimeout int `json:"busy_timeout"`

	// Enable WAL mode optimizations
	WALEnabled bool `json:"wal_enabled"`
}

// PostgresConfig holds PostgreSQL-specific configuration (for future use)
type PostgresConfig struct {
	// SSL mode (disable, require, verify-ca, verify-full)
	SSLMode string `json:"ssl_mode"`

	// Schema name
	Schema string `json:"schema"`

	// Search path
	SearchPath string `json:"search_path"`

	// Application name
	ApplicationName string `json:"application_name"`

	// Statement timeout in milliseconds
	StatementTimeout int `json:"statement_timeout"`

	// Lock timeout in milliseconds
	LockTimeout int `json:"lock_timeout"`
}

// DefaultConfig returns a default configuration for SQLite
func DefaultConfig() *Config {
	return &Config{
		Driver: "sqlite",
		DSN:    "file:memory.db?cache=shared&mode=rwc",

		// Connection pool (SQLite typically uses 1 connection)
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: 0, // No limit for SQLite
		ConnMaxIdleTime: 0, // No limit for SQLite

		// Query settings
		QueryTimeout:       30 * time.Second,
		SlowQueryThreshold: 1 * time.Second,

		// Migration settings
		MigrationsPath: "internal/storage/migrations",
		AutoMigrate:    true,
		MigrationTable: "schema_migrations",

		// Performance settings
		EnableQueryLog: false,

		// SQLite specific
		SQLite: SQLiteConfig{
			JournalMode: "WAL",    // Write-Ahead Logging for better concurrency
			Synchronous: "NORMAL", // Good balance of safety and speed
			CacheSize:   -2000,    // 2MB cache
			ForeignKeys: true,     // Enable foreign key constraints
			BusyTimeout: 5000,     // 5 seconds
			WALEnabled:  true,     // Enable WAL mode
		},
	}
}

// DefaultTestConfig returns a configuration suitable for testing
func DefaultTestConfig() *Config {
	config := DefaultConfig()
	config.DSN = "file::memory:?cache=shared&mode=memory"
	config.EnableQueryLog = true
	config.AutoMigrate = true
	return config
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Driver == "" {
		return NewStorageError("validate", "config", "", ErrInvalidInput)
	}

	if c.DSN == "" {
		return NewStorageError("validate", "config", "", ErrInvalidInput)
	}

	// Set defaults if not specified
	if c.MaxOpenConns == 0 {
		if c.Driver == "sqlite" {
			c.MaxOpenConns = 1
		} else {
			c.MaxOpenConns = 25
		}
	}

	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = c.MaxOpenConns
	}

	if c.QueryTimeout == 0 {
		c.QueryTimeout = 30 * time.Second
	}

	if c.MigrationTable == "" {
		c.MigrationTable = "schema_migrations"
	}

	return nil
}
