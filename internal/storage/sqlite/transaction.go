package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/tr4d3r/ghcp-memory-context/internal/models"
	"github.com/tr4d3r/ghcp-memory-context/internal/storage"
	"github.com/tr4d3r/ghcp-memory-context/pkg/types"
)

// Transaction implements the storage.Transaction interface for SQLite
type Transaction struct {
	tx     *sql.Tx
	driver *Driver
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	if t.tx == nil {
		return storage.NewStorageError("commit", "sqlite", "", storage.ErrTransactionClosed)
	}

	err := t.tx.Commit()
	t.tx = nil // Mark as closed

	if err != nil {
		return storage.NewStorageError("commit", "sqlite", "", err)
	}

	return nil
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	if t.tx == nil {
		return storage.NewStorageError("rollback", "sqlite", "", storage.ErrTransactionClosed)
	}

	err := t.tx.Rollback()
	t.tx = nil // Mark as closed

	if err != nil {
		return storage.NewStorageError("rollback", "sqlite", "", err)
	}

	return nil
}

// Tx returns the underlying *sql.Tx for operations
// This is not part of the interface but used by the store implementations
func (t *Transaction) Tx() *sql.Tx {
	return t.tx
}

// TaskStore methods - TODO: Implement in future subtask
func (t *Transaction) CreateTask(ctx context.Context, task *models.Task) error {
	return storage.NewStorageError("create_task", "task", task.ID, storage.ErrUnsupportedOperation)
}

func (t *Transaction) GetTask(ctx context.Context, id string) (*models.Task, error) {
	return nil, storage.NewStorageError("get_task", "task", id, storage.ErrUnsupportedOperation)
}

func (t *Transaction) UpdateTask(ctx context.Context, task *models.Task) error {
	return storage.NewStorageError("update_task", "task", task.ID, storage.ErrUnsupportedOperation)
}

func (t *Transaction) DeleteTask(ctx context.Context, id string) error {
	return storage.NewStorageError("delete_task", "task", id, storage.ErrUnsupportedOperation)
}

func (t *Transaction) ListTasks(ctx context.Context, filter storage.TaskFilter) ([]*models.Task, error) {
	return nil, storage.NewStorageError("list_tasks", "task", "", storage.ErrUnsupportedOperation)
}

func (t *Transaction) GetTasksByParent(ctx context.Context, parentID string) ([]*models.Task, error) {
	return nil, storage.NewStorageError("get_tasks_by_parent", "task", parentID, storage.ErrUnsupportedOperation)
}

func (t *Transaction) GetTaskDependencies(ctx context.Context, taskID string) ([]*models.Task, error) {
	return nil, storage.NewStorageError("get_task_dependencies", "task", taskID, storage.ErrUnsupportedOperation)
}

// ContextStore methods - TODO: Implement in future subtask
func (t *Transaction) CreateContext(ctx context.Context, obj types.ContextObject) error {
	return storage.NewStorageError("create_context", "context", obj.GetID(), storage.ErrUnsupportedOperation)
}

func (t *Transaction) GetContext(ctx context.Context, id string) (types.ContextObject, error) {
	return nil, storage.NewStorageError("get_context", "context", id, storage.ErrUnsupportedOperation)
}

func (t *Transaction) UpdateContext(ctx context.Context, obj types.ContextObject) error {
	return storage.NewStorageError("update_context", "context", obj.GetID(), storage.ErrUnsupportedOperation)
}

func (t *Transaction) DeleteContext(ctx context.Context, id string) error {
	return storage.NewStorageError("delete_context", "context", id, storage.ErrUnsupportedOperation)
}

func (t *Transaction) ListContexts(ctx context.Context, filter storage.ContextFilter) ([]types.ContextObject, error) {
	return nil, storage.NewStorageError("list_contexts", "context", "", storage.ErrUnsupportedOperation)
}

// SessionStore methods - TODO: Implement in future subtask
func (t *Transaction) CreateSession(ctx context.Context, session *storage.Session) error {
	return storage.NewStorageError("create_session", "session", session.ID, storage.ErrUnsupportedOperation)
}

func (t *Transaction) GetSession(ctx context.Context, id string) (*storage.Session, error) {
	return nil, storage.NewStorageError("get_session", "session", id, storage.ErrUnsupportedOperation)
}

func (t *Transaction) UpdateSession(ctx context.Context, session *storage.Session) error {
	return storage.NewStorageError("update_session", "session", session.ID, storage.ErrUnsupportedOperation)
}

func (t *Transaction) DeleteSession(ctx context.Context, id string) error {
	return storage.NewStorageError("delete_session", "session", id, storage.ErrUnsupportedOperation)
}

func (t *Transaction) ListSessions(ctx context.Context, filter storage.SessionFilter) ([]*storage.Session, error) {
	return nil, storage.NewStorageError("list_sessions", "session", "", storage.ErrUnsupportedOperation)
}

func (t *Transaction) CleanupExpiredSessions(ctx context.Context, olderThan time.Duration) error {
	return storage.NewStorageError("cleanup_sessions", "session", "", storage.ErrUnsupportedOperation)
}
