package storage

import (
	"context"
	"time"

	"github.com/tr4d3r/ghcp-memory-context/internal/models"
	"github.com/tr4d3r/ghcp-memory-context/pkg/types"
)

// Storage defines the main interface for all storage operations
// This abstraction allows us to swap storage backends (file-based, database, etc.)
type Storage interface {
	// Lifecycle methods
	Connect(ctx context.Context) error
	Close() error
	Ping(ctx context.Context) error

	// Entity operations
	EntityStore

	// Context operations
	ContextStore

	// Session operations
	SessionStore

	// Transaction support (may not be applicable for file-based storage)
	BeginTx(ctx context.Context) (Transaction, error)
}

// EntityStore defines operations for Entity storage
type EntityStore interface {
	// CreateEntity creates a new entity
	CreateEntity(ctx context.Context, entity *models.Entity) error

	// GetEntity retrieves an entity by name
	GetEntity(ctx context.Context, name string) (*models.Entity, error)

	// UpdateEntity updates an existing entity
	UpdateEntity(ctx context.Context, entity *models.Entity) error

	// DeleteEntity removes an entity by name
	DeleteEntity(ctx context.Context, name string) error

	// ListEntities retrieves entities, optionally filtered by type
	ListEntities(ctx context.Context, entityType string) ([]*models.Entity, error)

	// EntityExists checks if an entity exists
	EntityExists(name string) bool

	// SearchObservations searches for observations across entities
	SearchObservations(ctx context.Context, query string, entityType string) ([]SearchResult, error)

	// GetRelations retrieves all relations
	GetRelations(ctx context.Context) (*models.RelationSet, error)

	// SaveRelations saves the relation set
	SaveRelations(ctx context.Context, relations *models.RelationSet) error
}

// SearchResult represents a search result from observation queries
type SearchResult struct {
	EntityName  string             `json:"entityName"`
	EntityType  string             `json:"entityType"`
	Observation models.Observation `json:"observation"`
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

// Transaction represents a storage transaction (may be no-op for file storage)
type Transaction interface {
	// Commit commits the transaction
	Commit() error

	// Rollback rolls back the transaction
	Rollback() error

	// Entity operations within transaction
	EntityStore

	// Context operations within transaction
	ContextStore

	// Session operations within transaction
	SessionStore
}

// EntityFilter defines filtering options for entity queries
type EntityFilter struct {
	// Filter by entity fields
	EntityType *string
	Owner      *string

	// Filter by time ranges
	CreatedAfter  *time.Time
	CreatedBefore *time.Time

	// Search in observations
	SearchQuery *string

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
