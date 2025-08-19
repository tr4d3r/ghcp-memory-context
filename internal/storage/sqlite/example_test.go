package sqlite_test

import (
	"context"
	"fmt"
	"log"

	"github.com/tr4d3r/ghcp-memory-context/internal/storage"
	_ "github.com/tr4d3r/ghcp-memory-context/internal/storage/sqlite" // Register SQLite driver
)

func Example() {
	ctx := context.Background()

	// Create configuration for in-memory SQLite database
	config := &storage.Config{
		Driver: "sqlite",
		DSN:    ":memory:",
		SQLite: storage.SQLiteConfig{
			ForeignKeys: true,
			WALEnabled:  false, // Not needed for in-memory
			JournalMode: "MEMORY",
			Synchronous: "FULL",
		},
	}

	// Create storage instance
	store, err := storage.New(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Test connection
	if err := store.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Successfully connected to SQLite database")

	// Create a transaction
	tx, err := store.BeginTx(ctx)
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	// Commit transaction (empty transaction for demo)
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	fmt.Println("Transaction completed successfully")

	// Output:
	// Successfully connected to SQLite database
	// Transaction completed successfully
}
