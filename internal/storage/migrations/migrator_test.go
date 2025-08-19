package migrations

import (
	"context"
	"database/sql"
	"embed"
	"testing"
	"time"

	"github.com/tr4d3r/ghcp-memory-context/internal/storage"
	_ "modernc.org/sqlite"
)

// testMigrations contains embedded test migration files
// The go:embed directive below embeds all .sql files from testdata/
//
//go:embed testdata/*.sql
var testMigrations embed.FS

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	return db
}

func TestMigrator_Initialize(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	migrator := NewMigrator(MigratorConfig{
		DB:        db,
		TableName: "test_migrations",
	})

	ctx := context.Background()
	err := migrator.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Check that table was created
	var tableName string
	err = db.QueryRowContext(ctx,
		"SELECT name FROM sqlite_master WHERE type='table' AND name='test_migrations'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Migration table not created: %v", err)
	}

	if tableName != "test_migrations" {
		t.Errorf("Expected table name 'test_migrations', got '%s'", tableName)
	}
}

func TestMigrator_LoadMigrations(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	migrator := NewMigrator(MigratorConfig{
		DB:             db,
		TableName:      "schema_migrations",
		MigrationsFS:   testMigrations,
		MigrationsPath: "testdata",
	})

	ctx := context.Background()
	if err := migrator.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	migrations, err := migrator.LoadMigrations(ctx)
	if err != nil {
		t.Fatalf("LoadMigrations failed: %v", err)
	}

	if len(migrations) == 0 {
		t.Error("Expected migrations to be loaded")
	}

	// Check that migrations are sorted by version
	for i := 1; i < len(migrations); i++ {
		if migrations[i].Version <= migrations[i-1].Version {
			t.Errorf("Migrations not sorted by version: %d <= %d",
				migrations[i].Version, migrations[i-1].Version)
		}
	}
}

func TestMigrator_UpDown(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	migrator := NewMigrator(MigratorConfig{
		DB:             db,
		TableName:      "schema_migrations",
		MigrationsFS:   testMigrations,
		MigrationsPath: "testdata",
	})

	ctx := context.Background()
	if err := migrator.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test Up migration
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("Up migration failed: %v", err)
	}

	// Check migration status
	migrations, err := migrator.Status(ctx)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	appliedCount := 0
	for _, migration := range migrations {
		if migration.IsApplied {
			appliedCount++
		}
	}

	if appliedCount == 0 {
		t.Error("Expected at least one migration to be applied")
	}

	// Test Down migration
	if err := migrator.Down(ctx); err != nil {
		t.Fatalf("Down migration failed: %v", err)
	}

	// Check that one migration was rolled back
	migrations, err = migrator.Status(ctx)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	newAppliedCount := 0
	for _, migration := range migrations {
		if migration.IsApplied {
			newAppliedCount++
		}
	}

	if newAppliedCount != appliedCount-1 {
		t.Errorf("Expected %d applied migrations after rollback, got %d",
			appliedCount-1, newAppliedCount)
	}
}

func TestMigrator_UpTo(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	migrator := NewMigrator(MigratorConfig{
		DB:             db,
		TableName:      "schema_migrations",
		MigrationsFS:   testMigrations,
		MigrationsPath: "testdata",
	})

	ctx := context.Background()
	if err := migrator.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Migrate up to version 1
	if err := migrator.UpTo(ctx, 1); err != nil {
		t.Fatalf("UpTo failed: %v", err)
	}

	// Check that only version 1 is applied
	migrations, err := migrator.Status(ctx)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	for _, migration := range migrations {
		if migration.Version == 1 && !migration.IsApplied {
			t.Error("Migration version 1 should be applied")
		}
		if migration.Version > 1 && migration.IsApplied {
			t.Errorf("Migration version %d should not be applied", migration.Version)
		}
	}
}

func TestMigrator_Validate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	migrator := NewMigrator(MigratorConfig{
		DB:             db,
		TableName:      "schema_migrations",
		MigrationsFS:   testMigrations,
		MigrationsPath: "testdata",
	})

	ctx := context.Background()
	if err := migrator.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Validate should pass for valid test migrations
	if err := migrator.Validate(ctx); err != nil {
		t.Errorf("Validate failed: %v", err)
	}
}

func TestManager_AutoMigrate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	config := &storage.Config{
		AutoMigrate:    true,
		MigrationTable: "schema_migrations",
	}

	manager := NewManager(db, config, WithMigrationsFS(testMigrations), WithMigrationsPath("testdata"))

	ctx := context.Background()
	if err := manager.AutoMigrate(ctx); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	// Check that migrations were applied
	migrations, err := manager.Status(ctx)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	appliedCount := 0
	for _, migration := range migrations {
		if migration.IsApplied {
			appliedCount++
		}
	}

	if appliedCount == 0 {
		t.Error("Expected migrations to be auto-applied")
	}
}

func TestManager_AutoMigrateDisabled(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	config := &storage.Config{
		AutoMigrate:    false,
		MigrationTable: "schema_migrations",
	}

	manager := NewManager(db, config, WithMigrationsFS(testMigrations), WithMigrationsPath("testdata"))

	ctx := context.Background()
	if err := manager.AutoMigrate(ctx); err != nil {
		t.Fatalf("AutoMigrate should not fail when disabled: %v", err)
	}

	// Migrations table should not exist when auto-migrate is disabled
	var count int
	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check for migrations table: %v", err)
	}

	if count > 0 {
		t.Error("Migrations table should not exist when auto-migrate is disabled")
	}
}

func TestMigrationRecord(t *testing.T) {
	now := time.Now()
	record := MigrationRecord{
		Version:   1,
		Filename:  "001_test.sql",
		AppliedAt: now,
		Checksum:  "test-checksum",
	}

	if record.Version != 1 {
		t.Errorf("Expected version 1, got %d", record.Version)
	}

	if record.Filename != "001_test.sql" {
		t.Errorf("Expected filename '001_test.sql', got '%s'", record.Filename)
	}

	if !record.AppliedAt.Equal(now) {
		t.Errorf("Expected applied time %v, got %v", now, record.AppliedAt)
	}

	if record.Checksum != "test-checksum" {
		t.Errorf("Expected checksum 'test-checksum', got '%s'", record.Checksum)
	}
}

func TestCalculateChecksum(t *testing.T) {
	content := "CREATE TABLE test (id INTEGER PRIMARY KEY);"
	checksum1 := calculateChecksum(content)
	checksum2 := calculateChecksum(content)

	if checksum1 != checksum2 {
		t.Error("Checksum should be deterministic")
	}

	if len(checksum1) != 64 { // SHA256 hex = 64 chars
		t.Errorf("Expected checksum length 64, got %d", len(checksum1))
	}

	// Different content should produce different checksum
	differentContent := "CREATE TABLE test2 (id INTEGER PRIMARY KEY);"
	checksum3 := calculateChecksum(differentContent)

	if checksum1 == checksum3 {
		t.Error("Different content should produce different checksums")
	}
}
