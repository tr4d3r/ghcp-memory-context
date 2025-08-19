package storage

import (
	"context"
	"time"

	"github.com/tr4d3r/ghcp-memory-context/internal/models"
	"github.com/tr4d3r/ghcp-memory-context/pkg/types"
)

// Storage defines the main interface for all storage operations
// This abstraction allows us to swap storage backends (SQLite, PostgreSQL, etc.)
type Storage interface {
	// Lifecycle methods
	Connect(ctx context.Context) error
	Close() error
	Ping(ctx context.Context) error

	// Task operations
	TaskStore

	// Context operations
	ContextStore

	// Session operations
	SessionStore

	// Transaction support
	BeginTx(ctx context.Context) (Transaction, error)
}

// TaskStore defines operations for Task entities
type TaskStore interface {
	// CreateTask creates a new task in the database
	CreateTask(ctx context.Context, task *models.Task) error

	// GetTask retrieves a task by ID
	GetTask(ctx context.Context, id string) (*models.Task, error)

	// UpdateTask updates an existing task
	UpdateTask(ctx context.Context, task *models.Task) error

	// DeleteTask removes a task by ID
	DeleteTask(ctx context.Context, id string) error

	// ListTasks retrieves tasks with optional filters
	ListTasks(ctx context.Context, filter TaskFilter) ([]*models.Task, error)

	// GetTasksByParent retrieves all subtasks for a parent task
	GetTasksByParent(ctx context.Context, parentID string) ([]*models.Task, error)

	// GetTaskDependencies retrieves all tasks that depend on the given task
	GetTaskDependencies(ctx context.Context, taskID string) ([]*models.Task, error)
}

// ContextStore defines operations for generic context objects
type ContextStore interface {
	// CreateContext creates a new context object
	CreateContext(ctx context.Context, obj types.ContextObject) error

	// GetContext retrieves a context object by ID
	GetContext(ctx context.Context, id string) (types.ContextObject, error)

	// UpdateContext updates an existing context object
	UpdateContext(ctx context.Context, obj types.ContextObject) error

	// DeleteContext removes a context object by ID
	DeleteContext(ctx context.Context, id string) error

	// ListContexts retrieves context objects with optional filters
	ListContexts(ctx context.Context, filter ContextFilter) ([]types.ContextObject, error)
}

// SessionStore defines operations for session management
type SessionStore interface {
	// CreateSession creates a new session
	CreateSession(ctx context.Context, session *Session) error

	// GetSession retrieves a session by ID
	GetSession(ctx context.Context, id string) (*Session, error)

	// UpdateSession updates session information
	UpdateSession(ctx context.Context, session *Session) error

	// DeleteSession removes a session
	DeleteSession(ctx context.Context, id string) error

	// ListSessions retrieves sessions with optional filters
	ListSessions(ctx context.Context, filter SessionFilter) ([]*Session, error)

	// CleanupExpiredSessions removes sessions older than the specified duration
	CleanupExpiredSessions(ctx context.Context, olderThan time.Duration) error
}

// Transaction represents a database transaction
type Transaction interface {
	// Commit commits the transaction
	Commit() error

	// Rollback rolls back the transaction
	Rollback() error

	// Task operations within transaction
	TaskStore

	// Context operations within transaction
	ContextStore

	// Session operations within transaction
	SessionStore
}

// TaskFilter defines filtering options for task queries
type TaskFilter struct {
	// Filter by task fields
	Status   *models.TaskStatus
	Priority *models.TaskPriority
	Owner    *string
	Assignee *string

	// Filter by relationships
	ParentID        *string
	HasSubtasks     *bool
	HasDependencies *bool

	// Filter by time ranges
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	UpdatedAfter  *time.Time
	UpdatedBefore *time.Time
	DueAfter      *time.Time
	DueBefore     *time.Time

	// Pagination
	Limit  int
	Offset int

	// Sorting
	SortBy    string // field name to sort by
	SortOrder string // "asc" or "desc"
}

// ContextFilter defines filtering options for context queries
type ContextFilter struct {
	Type      *types.ContextType
	Scope     *types.ContextScope
	SessionID *string
	ProjectID *string
	Owner     *string

	// Time filters
	CreatedAfter  *time.Time
	CreatedBefore *time.Time

	// Pagination
	Limit  int
	Offset int
}

// SessionFilter defines filtering options for session queries
type SessionFilter struct {
	UserID    *string
	ProjectID *string
	Active    *bool

	// Time filters
	CreatedAfter       *time.Time
	CreatedBefore      *time.Time
	LastAccessedAfter  *time.Time
	LastAccessedBefore *time.Time

	// Pagination
	Limit  int
	Offset int
}

// Session represents a user session with context
type Session struct {
	ID             string            `json:"id"`
	UserID         string            `json:"user_id"`
	ProjectID      string            `json:"project_id,omitempty"`
	Name           string            `json:"name,omitempty"`
	Description    string            `json:"description,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	Active         bool              `json:"active"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	LastAccessedAt time.Time         `json:"last_accessed_at"`
	ExpiresAt      *time.Time        `json:"expires_at,omitempty"`
}
