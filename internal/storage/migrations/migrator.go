package migrations

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tr4d3r/ghcp-memory-context/internal/storage"
)

// Migration represents a single database migration
type Migration struct {
	Version     int       `json:"version"`
	Name        string    `json:"name"`
	Filename    string    `json:"filename"`
	UpSQL       string    `json:"up_sql"`
	DownSQL     string    `json:"down_sql"`
	Checksum    string    `json:"checksum"`
	AppliedAt   time.Time `json:"applied_at,omitempty"`
	IsApplied   bool      `json:"is_applied"`
	Description string    `json:"description,omitempty"`
}

// MigrationRecord represents a migration record in the database
type MigrationRecord struct {
	Version   int       `json:"version"`
	Filename  string    `json:"filename"`
	AppliedAt time.Time `json:"applied_at"`
	Checksum  string    `json:"checksum"`
}

// Migrator handles database migrations
type Migrator struct {
	db             *sql.DB
	tableName      string
	migrationsFS   fs.FS
	migrationsPath string
}

// MigratorConfig holds configuration for the migrator
type MigratorConfig struct {
	DB             *sql.DB
	TableName      string
	MigrationsFS   fs.FS
	MigrationsPath string
}

// NewMigrator creates a new migrator instance
func NewMigrator(config MigratorConfig) *Migrator {
	tableName := config.TableName
	if tableName == "" {
		tableName = "schema_migrations"
	}

	migrationsPath := config.MigrationsPath
	if migrationsPath == "" {
		migrationsPath = "."
	}

	return &Migrator{
		db:             config.DB,
		tableName:      tableName,
		migrationsFS:   config.MigrationsFS,
		migrationsPath: migrationsPath,
	}
}

// Initialize creates the migrations table if it doesn't exist
func (m *Migrator) Initialize(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version INTEGER PRIMARY KEY,
			filename TEXT NOT NULL,
			applied_at DATETIME NOT NULL DEFAULT (datetime('now', 'utc')),
			checksum TEXT
		)`, m.tableName)

	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return storage.NewStorageError("initialize_migrations", "migration", "", err)
	}

	return nil
}

// LoadMigrations reads migration files from the filesystem
func (m *Migrator) LoadMigrations(ctx context.Context) ([]*Migration, error) {
	if m.migrationsFS == nil {
		return nil, storage.NewStorageError("load_migrations", "migration", "",
			fmt.Errorf("migrations filesystem not configured"))
	}

	var migrations []*Migration
	filenameRegex := regexp.MustCompile(`^(\d+)_([^.]+)\.(up|down)\.sql$`)

	err := fs.WalkDir(m.migrationsFS, m.migrationsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".sql") {
			return nil
		}

		matches := filenameRegex.FindStringSubmatch(d.Name())
		if len(matches) != 4 {
			// Skip files that don't match the pattern
			return nil
		}

		version, err := strconv.Atoi(matches[1])
		if err != nil {
			return fmt.Errorf("invalid migration version in %s: %w", d.Name(), err)
		}

		name := matches[2]
		direction := matches[3]

		// Read file content
		content, err := fs.ReadFile(m.migrationsFS, path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", path, err)
		}

		sql := string(content)
		checksum := calculateChecksum(sql)

		// Find or create migration
		var migration *Migration
		for _, m := range migrations {
			if m.Version == version && m.Name == name {
				migration = m
				break
			}
		}

		if migration == nil {
			migration = &Migration{
				Version:  version,
				Name:     name,
				Filename: fmt.Sprintf("%03d_%s", version, name),
				Checksum: checksum,
			}
			migrations = append(migrations, migration)
		}

		// Set SQL content based on direction
		if direction == "up" {
			migration.UpSQL = sql
		} else if direction == "down" {
			migration.DownSQL = sql
		}

		return nil
	})

	if err != nil {
		return nil, storage.NewStorageError("load_migrations", "migration", "", err)
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	// Load applied migrations from database
	appliedMigrations, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	// Mark migrations as applied
	for _, migration := range migrations {
		if record, exists := appliedMigrations[migration.Version]; exists {
			migration.IsApplied = true
			migration.AppliedAt = record.AppliedAt

			// Verify checksum if available
			if record.Checksum != "" && migration.Checksum != record.Checksum {
				return nil, storage.NewStorageError("load_migrations", "migration",
					migration.Filename, fmt.Errorf("checksum mismatch for migration %d", migration.Version))
			}
		}
	}

	return migrations, nil
}

// getAppliedMigrations retrieves applied migrations from the database
func (m *Migrator) getAppliedMigrations(ctx context.Context) (map[int]MigrationRecord, error) {
	query := fmt.Sprintf("SELECT version, filename, applied_at, checksum FROM %s", m.tableName)
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, storage.NewStorageError("get_applied_migrations", "migration", "", err)
	}
	defer rows.Close()

	applied := make(map[int]MigrationRecord)
	for rows.Next() {
		var record MigrationRecord
		var checksum sql.NullString

		err := rows.Scan(&record.Version, &record.Filename, &record.AppliedAt, &checksum)
		if err != nil {
			return nil, storage.NewStorageError("get_applied_migrations", "migration", "", err)
		}

		if checksum.Valid {
			record.Checksum = checksum.String
		}

		applied[record.Version] = record
	}

	if err := rows.Err(); err != nil {
		return nil, storage.NewStorageError("get_applied_migrations", "migration", "", err)
	}

	return applied, nil
}

// Up migrates the database to the latest version
func (m *Migrator) Up(ctx context.Context) error {
	migrations, err := m.LoadMigrations(ctx)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if !migration.IsApplied {
			if err := m.applyMigration(ctx, migration); err != nil {
				return err
			}
		}
	}

	return nil
}

// UpTo migrates the database to a specific version
func (m *Migrator) UpTo(ctx context.Context, targetVersion int) error {
	migrations, err := m.LoadMigrations(ctx)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if migration.Version > targetVersion {
			break
		}

		if !migration.IsApplied {
			if err := m.applyMigration(ctx, migration); err != nil {
				return err
			}
		}
	}

	return nil
}

// Down rolls back the database by one migration
func (m *Migrator) Down(ctx context.Context) error {
	migrations, err := m.LoadMigrations(ctx)
	if err != nil {
		return err
	}

	// Find the last applied migration
	var lastApplied *Migration
	for i := len(migrations) - 1; i >= 0; i-- {
		if migrations[i].IsApplied {
			lastApplied = migrations[i]
			break
		}
	}

	if lastApplied == nil {
		return storage.NewStorageError("down", "migration", "", fmt.Errorf("no migrations to roll back"))
	}

	return m.rollbackMigration(ctx, lastApplied)
}

// DownTo rolls back the database to a specific version
func (m *Migrator) DownTo(ctx context.Context, targetVersion int) error {
	migrations, err := m.LoadMigrations(ctx)
	if err != nil {
		return err
	}

	// Roll back migrations in reverse order
	for i := len(migrations) - 1; i >= 0; i-- {
		migration := migrations[i]
		if migration.Version <= targetVersion {
			break
		}

		if migration.IsApplied {
			if err := m.rollbackMigration(ctx, migration); err != nil {
				return err
			}
		}
	}

	return nil
}

// applyMigration applies a single migration
func (m *Migrator) applyMigration(ctx context.Context, migration *Migration) error {
	if migration.UpSQL == "" {
		return storage.NewStorageError("apply_migration", "migration", migration.Filename,
			fmt.Errorf("up migration SQL not found"))
	}

	// Start transaction
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return storage.NewStorageError("apply_migration", "migration", migration.Filename, err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	_, err = tx.ExecContext(ctx, migration.UpSQL)
	if err != nil {
		return storage.NewStorageError("apply_migration", "migration", migration.Filename, err)
	}

	// Record migration in migrations table
	query := fmt.Sprintf(`
		INSERT INTO %s (version, filename, applied_at, checksum)
		VALUES (?, ?, ?, ?)`, m.tableName)

	_, err = tx.ExecContext(ctx, query, migration.Version, migration.Filename,
		time.Now().UTC(), migration.Checksum)
	if err != nil {
		return storage.NewStorageError("apply_migration", "migration", migration.Filename, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return storage.NewStorageError("apply_migration", "migration", migration.Filename, err)
	}

	migration.IsApplied = true
	migration.AppliedAt = time.Now().UTC()

	return nil
}

// rollbackMigration rolls back a single migration
func (m *Migrator) rollbackMigration(ctx context.Context, migration *Migration) error {
	if migration.DownSQL == "" {
		return storage.NewStorageError("rollback_migration", "migration", migration.Filename,
			fmt.Errorf("down migration SQL not found"))
	}

	// Start transaction
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return storage.NewStorageError("rollback_migration", "migration", migration.Filename, err)
	}
	defer tx.Rollback()

	// Execute rollback SQL
	_, err = tx.ExecContext(ctx, migration.DownSQL)
	if err != nil {
		return storage.NewStorageError("rollback_migration", "migration", migration.Filename, err)
	}

	// Remove migration record from migrations table
	query := fmt.Sprintf("DELETE FROM %s WHERE version = ?", m.tableName)
	_, err = tx.ExecContext(ctx, query, migration.Version)
	if err != nil {
		return storage.NewStorageError("rollback_migration", "migration", migration.Filename, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return storage.NewStorageError("rollback_migration", "migration", migration.Filename, err)
	}

	migration.IsApplied = false
	migration.AppliedAt = time.Time{}

	return nil
}

// Status returns the current migration status
func (m *Migrator) Status(ctx context.Context) ([]*Migration, error) {
	return m.LoadMigrations(ctx)
}

// Validate checks migration integrity
func (m *Migrator) Validate(ctx context.Context) error {
	migrations, err := m.LoadMigrations(ctx)
	if err != nil {
		return err
	}

	// Check for missing up/down migrations
	for _, migration := range migrations {
		if migration.UpSQL == "" {
			return storage.NewStorageError("validate", "migration", migration.Filename,
				fmt.Errorf("missing up migration file"))
		}
		if migration.DownSQL == "" {
			return storage.NewStorageError("validate", "migration", migration.Filename,
				fmt.Errorf("missing down migration file"))
		}
	}

	// Check for version gaps
	for i, migration := range migrations {
		expectedVersion := i + 1
		if migration.Version != expectedVersion {
			return storage.NewStorageError("validate", "migration", migration.Filename,
				fmt.Errorf("version gap: expected %d, got %d", expectedVersion, migration.Version))
		}
	}

	return nil
}

// calculateChecksum calculates SHA256 checksum of migration content
func calculateChecksum(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// Reset drops all tables and resets the database (DANGEROUS)
func (m *Migrator) Reset(ctx context.Context) error {
	// This is a dangerous operation - require explicit confirmation
	migrations, err := m.LoadMigrations(ctx)
	if err != nil {
		return err
	}

	// Roll back all migrations in reverse order
	for i := len(migrations) - 1; i >= 0; i-- {
		migration := migrations[i]
		if migration.IsApplied {
			if err := m.rollbackMigration(ctx, migration); err != nil {
				return err
			}
		}
	}

	return nil
}
