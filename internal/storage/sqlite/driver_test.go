package sqlite

import (
	"context"
	"testing"

	"github.com/tr4d3r/ghcp-memory-context/internal/storage"
)

func TestDriver_NewDriver(t *testing.T) {
	tests := []struct {
		name    string
		config  *storage.Config
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name:    "default config",
			config:  storage.DefaultTestConfig(),
			wantErr: false,
		},
		{
			name: "custom config",
			config: &storage.Config{
				Driver: "sqlite",
				DSN:    ":memory:",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			storage, err := NewDriver(ctx, tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewDriver() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NewDriver() unexpected error: %v", err)
				return
			}

			if storage == nil {
				t.Errorf("NewDriver() returned nil storage")
			}
		})
	}
}

func TestDriver_Connect(t *testing.T) {
	ctx := context.Background()
	config := storage.DefaultTestConfig()

	driver, err := NewDriver(ctx, config)
	if err != nil {
		t.Fatalf("NewDriver() failed: %v", err)
	}

	// Test connection
	if err := driver.Connect(ctx); err != nil {
		t.Errorf("Connect() failed: %v", err)
	}

	// Test ping
	if err := driver.Ping(ctx); err != nil {
		t.Errorf("Ping() failed: %v", err)
	}

	// Test close
	if err := driver.Close(); err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	// Test ping after close should fail
	if err := driver.Ping(ctx); err == nil {
		t.Errorf("Ping() should fail after Close()")
	}
}

func TestDriver_Transaction(t *testing.T) {
	ctx := context.Background()
	config := storage.DefaultTestConfig()

	driver, err := NewDriver(ctx, config)
	if err != nil {
		t.Fatalf("NewDriver() failed: %v", err)
	}

	if err := driver.Connect(ctx); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer driver.Close()

	// Test transaction creation
	tx, err := driver.BeginTx(ctx)
	if err != nil {
		t.Errorf("BeginTx() failed: %v", err)
	}

	// Test commit
	if err := tx.Commit(); err != nil {
		t.Errorf("Commit() failed: %v", err)
	}

	// Test rollback transaction
	tx2, err := driver.BeginTx(ctx)
	if err != nil {
		t.Errorf("BeginTx() failed: %v", err)
	}

	if err := tx2.Rollback(); err != nil {
		t.Errorf("Rollback() failed: %v", err)
	}
}
