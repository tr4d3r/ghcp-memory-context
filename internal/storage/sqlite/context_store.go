package sqlite

import (
	"context"

	"github.com/tr4d3r/ghcp-memory-context/internal/storage"
	"github.com/tr4d3r/ghcp-memory-context/pkg/types"
)

// ContextStore implementation methods for Driver
// TODO: Implement full context store functionality in next subtask

func (d *Driver) CreateContext(ctx context.Context, obj types.ContextObject) error {
	// TODO: Implement context creation - for now return not implemented
	return storage.NewStorageError("create_context", "context", obj.GetID(), storage.ErrUnsupportedOperation)
}

func (d *Driver) GetContext(ctx context.Context, id string) (types.ContextObject, error) {
	// TODO: Implement context retrieval - for now return not implemented
	return nil, storage.NewStorageError("get_context", "context", id, storage.ErrUnsupportedOperation)
}

func (d *Driver) UpdateContext(ctx context.Context, obj types.ContextObject) error {
	// TODO: Implement context update - for now return not implemented
	return storage.NewStorageError("update_context", "context", obj.GetID(), storage.ErrUnsupportedOperation)
}

func (d *Driver) DeleteContext(ctx context.Context, id string) error {
	// TODO: Implement context deletion - for now return not implemented
	return storage.NewStorageError("delete_context", "context", id, storage.ErrUnsupportedOperation)
}

func (d *Driver) ListContexts(ctx context.Context, filter storage.ContextFilter) ([]types.ContextObject, error) {
	// TODO: Implement context listing - for now return not implemented
	return nil, storage.NewStorageError("list_contexts", "context", "", storage.ErrUnsupportedOperation)
}
