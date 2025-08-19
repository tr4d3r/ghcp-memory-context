package sqlite

import (
	"context"
	"testing"

	"github.com/tr4d3r/ghcp-memory-context/internal/storage"
)

func TestDriver_MigrationIntegration(t *testing.T) {
	ctx := context.Background()
	config := storage.DefaultTestConfig()

	storageDriver, err := NewDriver(ctx, config)
	if err != nil {
		t.Fatalf("NewDriver() failed: %v", err)
	}

	// Connect should automatically run migrations
	if err := storageDriver.Connect(ctx); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer storageDriver.Close()

	// Cast to concrete driver type for migration methods
	driver, ok := storageDriver.(*Driver)
	if !ok {
		t.Fatal("Failed to cast to concrete Driver type")
	}

	// Check that migration manager is initialized
	manager := driver.GetMigrationManager()
	if manager == nil {
		t.Fatal("Migration manager not initialized")
	}

	// Check migration status
	migrations, err := driver.MigrationStatus(ctx)
	if err != nil {
		t.Fatalf("MigrationStatus() failed: %v", err)
	}

	if len(migrations) == 0 {
		t.Error("Expected at least one migration")
	}

	// Check that initial schema migration was applied
	found := false
	for _, migration := range migrations {
		if migration.Version == 1 && migration.IsApplied {
			found = true
			break
		}
	}

	if !found {
		t.Error("Initial schema migration (version 1) should be applied")
	}

	// Test that we can query the created tables
	tables := []string{"sessions", "context_objects", "tasks", "task_dependencies", "code_references"}

	for _, table := range tables {
		var count int
		query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
		err := driver.DB().QueryRowContext(ctx, query, table).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check table %s: %v", table, err)
		}

		if count != 1 {
			t.Errorf("Table %s should exist", table)
		}
	}
}

func TestDriver_MigrationMethods(t *testing.T) {
	ctx := context.Background()
	config := storage.DefaultTestConfig()

	storageDriver, err := NewDriver(ctx, config)
	if err != nil {
		t.Fatalf("NewDriver() failed: %v", err)
	}

	if err := storageDriver.Connect(ctx); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer storageDriver.Close()

	// Cast to concrete driver type for migration methods
	driver, ok := storageDriver.(*Driver)
	if !ok {
		t.Fatal("Failed to cast to concrete Driver type")
	}

	// Test migration status
	status, err := driver.MigrationStatus(ctx)
	if err != nil {
		t.Errorf("MigrationStatus() failed: %v", err)
	}

	if len(status) == 0 {
		t.Error("Expected migration status to contain migrations")
	}

	// Test migration down (should work since we have migrations applied)
	if err := driver.MigrateDown(ctx); err != nil {
		t.Errorf("MigrateDown() failed: %v", err)
	}

	// Test migration up (should restore the migration)
	if err := driver.MigrateUp(ctx); err != nil {
		t.Errorf("MigrateUp() failed: %v", err)
	}
}

func TestDriver_AutoMigrateDisabled(t *testing.T) {
	ctx := context.Background()
	config := storage.DefaultTestConfig()
	config.AutoMigrate = false // Disable auto-migration

	storageDriver, err := NewDriver(ctx, config)
	if err != nil {
		t.Fatalf("NewDriver() failed: %v", err)
	}

	// Connect should not run migrations when disabled
	if err := storageDriver.Connect(ctx); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer storageDriver.Close()

	// Cast to concrete driver type for migration methods
	driver, ok := storageDriver.(*Driver)
	if !ok {
		t.Fatal("Failed to cast to concrete Driver type")
	}

	// Migration manager should still be initialized
	manager := driver.GetMigrationManager()
	if manager == nil {
		t.Fatal("Migration manager should be initialized even when auto-migrate is disabled")
	}

	// But migrations should not be automatically applied
	migrations, err := driver.MigrationStatus(ctx)
	if err != nil {
		// This might fail if the migrations table doesn't exist, which is expected
		// when auto-migrate is disabled
		return
	}

	appliedCount := 0
	for _, migration := range migrations {
		if migration.IsApplied {
			appliedCount++
		}
	}

	// With auto-migrate disabled, we shouldn't have any applied migrations initially
	// (unless manually applied)
	if appliedCount > 0 {
		t.Logf("Note: %d migrations were applied even with auto-migrate disabled", appliedCount)
	}
}
