# Storage Package

The storage package provides a database abstraction layer for the GHCP Memory Context Server. It defines interfaces and common utilities for persistent storage of tasks, context objects, and sessions.

## Architecture

The storage package follows a repository pattern with the following key components:

```
internal/storage/
├── interface.go      # Core storage interfaces
├── config.go        # Configuration structures
├── errors.go        # Error definitions and handling
├── factory.go       # Storage factory for driver registration
├── utils.go         # Utility functions
├── sqlite/          # SQLite implementation
│   ├── storage.go   # SQLite storage implementation
│   ├── tasks.go     # Task-specific operations
│   ├── context.go   # Context operations
│   ├── session.go   # Session operations
│   └── migrations/  # Database migrations
└── migrations/      # Shared migration utilities
```

## Usage

### Basic Usage

```go
import (
    "context"
    "github.com/tr4d3r/ghcp-memory-context/internal/storage"
    _ "github.com/tr4d3r/ghcp-memory-context/internal/storage/sqlite" // Register SQLite driver
)

// Create configuration
config := storage.DefaultConfig()
config.DSN = "file:memory.db?cache=shared&mode=rwc"

// Create storage instance
ctx := context.Background()
store, err := storage.New(ctx, config)
if err != nil {
    log.Fatal(err)
}
defer store.Close()

// Create a task
task := &models.Task{
    Title:       "Example Task",
    Description: "This is an example",
}
err = store.CreateTask(ctx, task)
```

### With Transactions

```go
// Begin transaction
tx, err := store.BeginTx(ctx)
if err != nil {
    return err
}
defer tx.Rollback() // Rollback if not committed

// Perform operations
err = tx.CreateTask(ctx, task1)
if err != nil {
    return err
}

err = tx.CreateTask(ctx, task2)
if err != nil {
    return err
}

// Commit transaction
return tx.Commit()
```

### Filtering and Querying

```go
// Query tasks with filters
filter := storage.TaskFilter{
    Status:   &models.TaskStatusPending,
    Owner:    storage.StringPtr("user123"),
    Limit:    10,
    SortBy:   "created_at",
    SortOrder: "desc",
}

tasks, err := store.ListTasks(ctx, filter)
```

## Interfaces

### Storage Interface

The main `Storage` interface provides:
- Connection lifecycle methods (`Connect`, `Close`, `Ping`)
- Task operations via `TaskStore`
- Context operations via `ContextStore`
- Session operations via `SessionStore`
- Transaction support via `BeginTx`

### TaskStore Interface

Operations for task management:
- `CreateTask` - Create a new task
- `GetTask` - Retrieve a task by ID
- `UpdateTask` - Update an existing task
- `DeleteTask` - Delete a task
- `ListTasks` - List tasks with filtering
- `GetTasksByParent` - Get subtasks
- `GetTaskDependencies` - Get dependent tasks

### ContextStore Interface

Operations for generic context objects:
- `CreateContext` - Create a context object
- `GetContext` - Retrieve a context object
- `UpdateContext` - Update a context object
- `DeleteContext` - Delete a context object
- `ListContexts` - List contexts with filtering

### SessionStore Interface

Operations for session management:
- `CreateSession` - Create a new session
- `GetSession` - Retrieve a session
- `UpdateSession` - Update session data
- `DeleteSession` - Delete a session
- `ListSessions` - List sessions
- `CleanupExpiredSessions` - Remove old sessions

## Configuration

The package supports configuration through the `Config` struct:

```go
config := &storage.Config{
    Driver: "sqlite",
    DSN:    "file:memory.db",

    // Connection pool settings
    MaxOpenConns:    10,
    MaxIdleConns:    5,
    ConnMaxLifetime: time.Hour,

    // Query settings
    QueryTimeout: 30 * time.Second,

    // Migration settings
    AutoMigrate:    true,
    MigrationsPath: "internal/storage/migrations",

    // SQLite specific
    SQLite: storage.SQLiteConfig{
        JournalMode: "WAL",
        ForeignKeys: true,
    },
}
```

## Error Handling

The package defines common storage errors:
- `ErrNotFound` - Entity not found
- `ErrAlreadyExists` - Entity already exists
- `ErrInvalidInput` - Invalid input data
- `ErrDatabaseClosed` - Database connection closed
- `ErrConstraintViolation` - Database constraint violated

Use the helper functions to check error types:
```go
if storage.IsNotFound(err) {
    // Handle not found error
}
```

## Testing

For testing, use the in-memory configuration:

```go
config := storage.DefaultTestConfig()
store, err := storage.New(ctx, config)
```

This creates an in-memory SQLite database that's perfect for unit tests.

## Adding New Storage Backends

To add a new storage backend:

1. Create a new package (e.g., `internal/storage/postgres`)
2. Implement the `Storage` interface
3. Register the driver in an init function:

```go
func init() {
    storage.Register("postgres", NewPostgresStorage)
}
```

4. Import the package to register the driver:

```go
import _ "github.com/tr4d3r/ghcp-memory-context/internal/storage/postgres"
```

## Migration Management

The storage package supports automatic migrations. Place migration files in the configured migrations path with the naming convention:

```
001_initial_schema.up.sql
001_initial_schema.down.sql
002_add_indexes.up.sql
002_add_indexes.down.sql
```

Migrations are automatically applied when `AutoMigrate` is enabled in the configuration.
