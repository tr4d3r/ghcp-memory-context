# Database Migration System

A comprehensive database migration system for the GHCP Memory Context Server, providing version-controlled schema management with automatic application and rollback capabilities.

## Features

- **Version-Controlled Migrations**: Sequential, numbered migration files with up/down SQL
- **Automatic Application**: Auto-migrate on database connection if enabled  
- **Rollback Support**: Safe rollback to previous versions with down migrations
- **Integrity Checking**: SHA256 checksums and validation of migration files
- **Embedded Migrations**: Compile-time embedding of migration files into binary
- **Transaction Safety**: All migrations run within database transactions
- **Flexible Configuration**: Support for custom migration tables and paths

## Architecture

### Core Components

- **Migrator**: Core migration engine that handles loading and executing migrations
- **Manager**: High-level interface that integrates with storage configuration
- **Migration**: Represents a single migration with up/down SQL and metadata
- **MigrationRecord**: Database record tracking applied migrations

### Integration

The migration system is automatically integrated with the SQLite storage driver:

- Migrations run automatically on database connection (if `AutoMigrate: true`)
- Migration table is created automatically when needed
- All schema changes are version-controlled and reproducible

## Usage

### Basic Configuration

```go
// Auto-migrate enabled (default)
config := &storage.Config{
    AutoMigrate:    true,
    MigrationTable: "schema_migrations", // default
}

// Auto-migrate disabled
config := &storage.Config{
    AutoMigrate: false,
}
```

### Manual Migration Management

```go
// Get migration manager from SQLite driver
store, _ := storage.New(ctx, config)
driver := store.(*sqlite.Driver)
manager := driver.GetMigrationManager()

// Apply all pending migrations
err := manager.Up(ctx)

// Rollback last migration
err := manager.Down(ctx)

// Get migration status
migrations, err := manager.Status(ctx)

// Validate migration integrity
err := manager.Validate(ctx)
```

### Migration Status and Control

```go
// Check current status
status, err := driver.MigrationStatus(ctx)
for _, migration := range status {
    fmt.Printf("Migration %d: %s (Applied: %v)\n",
        migration.Version, migration.Name, migration.IsApplied)
}

// Manual migration control
err := driver.MigrateUp(ctx)   // Apply pending migrations
err := driver.MigrateDown(ctx) // Rollback last migration
```

## Migration File Format

### File Naming Convention

```
{version}_{name}.{direction}.sql
```

- `version`: Sequential integer (001, 002, etc.)
- `name`: Descriptive name using snake_case
- `direction`: Either `up` or `down`

### Examples

```
001_initial_schema.up.sql
001_initial_schema.down.sql
002_add_user_status.up.sql  
002_add_user_status.down.sql
003_create_indexes.up.sql
003_create_indexes.down.sql
```

### File Location

Migrations are stored in `internal/storage/migrations/` and embedded at compile time:

```
internal/storage/migrations/
├── 001_initial_schema.up.sql
├── 001_initial_schema.down.sql
├── 002_add_indexes.up.sql
├── 002_add_indexes.down.sql
└── README.md
```

## Writing Migrations

### Up Migration (Apply Changes)

```sql
-- 002_add_user_preferences.up.sql

-- Add user preferences table
CREATE TABLE user_preferences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now', 'utc')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now', 'utc')),

    UNIQUE(user_id, key)
);

-- Create indexes
CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);
CREATE INDEX idx_user_preferences_key ON user_preferences(key);
```

### Down Migration (Rollback Changes)

```sql
-- 002_add_user_preferences.down.sql

-- Remove user preferences table and indexes
DROP INDEX IF EXISTS idx_user_preferences_key;
DROP INDEX IF EXISTS idx_user_preferences_user_id;
DROP TABLE IF EXISTS user_preferences;
```

### Best Practices

1. **Always Provide Down Migrations**: Every up migration must have a corresponding down migration
2. **Use Transactions**: Migrations run in transactions automatically
3. **Test Rollbacks**: Ensure down migrations work correctly
4. **Avoid Data Loss**: Be careful with destructive operations
5. **Use IF EXISTS/IF NOT EXISTS**: Make migrations idempotent when possible

```sql
-- Good: Safe and idempotent
CREATE TABLE IF NOT EXISTS new_table (...);
DROP TABLE IF EXISTS old_table;

-- Good: Preserve data during schema changes
CREATE TABLE new_table AS SELECT * FROM old_table;
DROP TABLE old_table;
ALTER TABLE new_table RENAME TO old_table;
```

## Migration States

Each migration can be in one of several states:

- **Pending**: Migration file exists but not yet applied
- **Applied**: Migration has been successfully applied to database
- **Invalid**: Migration file has checksum mismatch or validation errors

## Error Handling

The migration system provides comprehensive error handling:

```go
// Migration-specific errors
if storage.IsNotFound(err) {
    // Migration file not found
}

// Get detailed error information
if storageErr, ok := err.(*storage.StorageError); ok {
    fmt.Printf("Operation: %s, Type: %s, ID: %s, Error: %v",
        storageErr.Op, storageErr.Type, storageErr.ID, storageErr.Err)
}
```

## Validation and Integrity

### Checksum Verification

Each migration file is checksummed (SHA256) to detect modifications:

```go
// Validate all migrations
err := manager.Validate(ctx)
if err != nil {
    // Handle validation errors (checksum mismatches, missing files, etc.)
}
```

### Common Validation Errors

- **Checksum Mismatch**: Migration file has been modified after application
- **Missing Down Migration**: Up migration exists without corresponding down migration
- **Version Gaps**: Non-sequential migration versions (e.g., 001, 003 missing 002)
- **Duplicate Versions**: Multiple migrations with same version number

## Advanced Usage

### Custom Migration Filesystem

```go
// Use custom embedded filesystem
//go:embed custom_migrations/*.sql
var customMigrations embed.FS

manager := migrations.NewManager(db, config,
    migrations.WithMigrationsFS(customMigrations),
    migrations.WithMigrationsPath("custom_migrations"))
```

### Programmatic Migration Control

```go
// Migrate to specific version
err := manager.UpTo(ctx, 5)

// Rollback to specific version  
err := manager.DownTo(ctx, 3)

// Reset database (DANGEROUS - removes all data)
err := manager.Reset(ctx)
```

### Migration Status Reporting

```go
migrations, err := manager.Status(ctx)
for _, migration := range migrations {
    status := "Pending"
    if migration.IsApplied {
        status = fmt.Sprintf("Applied %v", migration.AppliedAt.Format(time.RFC3339))
    }

    fmt.Printf("%-3d %-30s %s\n",
        migration.Version, migration.Name, status)
}
```

## Testing

### Unit Tests

The migration system includes comprehensive unit tests:

```bash
# Run migration tests
go test ./internal/storage/migrations -v

# Run integration tests  
go test ./internal/storage/sqlite -run Migration -v
```

### Test Migrations

Test migrations are provided in `testdata/` for testing:

```
testdata/
├── 001_create_users.up.sql
├── 001_create_users.down.sql  
├── 002_add_user_status.up.sql
└── 002_add_user_status.down.sql
```

## Performance Considerations

- **Embedded Migrations**: Zero-cost embedding at compile time
- **Transaction Safety**: Each migration runs in its own transaction
- **Checksumming**: SHA256 hashing for integrity verification
- **Lazy Loading**: Migrations loaded only when needed

## Security

- **Read-Only Filesystem**: Migration files are embedded and immutable
- **Checksum Verification**: Prevents tampering with migration files
- **Transaction Isolation**: Failed migrations are automatically rolled back
- **Input Validation**: Strict validation of migration file format and content

## Troubleshooting

### Common Issues

1. **Migration Already Applied**:
   - Check migration status with `Status()`
   - Verify database state manually

2. **Checksum Mismatch**:
   - Migration file was modified after application
   - Either revert changes or create new migration

3. **Missing Down Migration**:
   - Create corresponding `.down.sql` file
   - Ensure rollback logic is correct

4. **Version Gaps**:
   - Ensure sequential numbering (001, 002, 003...)
   - No skipped version numbers

### Debug Information

Enable detailed logging by setting config options:

```go
config := &storage.Config{
    EnableQueryLog:     true,
    SlowQueryThreshold: 100 * time.Millisecond,
}
```

## Future Enhancements

- **Schema Diffing**: Automatic generation of migration files
- **Multi-Database Support**: PostgreSQL, MySQL driver support
- **Migration Dependencies**: Complex dependency management
- **Parallel Migrations**: Concurrent migration execution
- **Migration Linting**: Static analysis of migration files
