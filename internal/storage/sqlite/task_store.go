package sqlite

import (
	"context"

	"github.com/tr4d3r/ghcp-memory-context/internal/models"
	"github.com/tr4d3r/ghcp-memory-context/internal/storage"
)

// TaskStore implementation methods for Driver
// TODO: Implement full task store functionality in next subtask

func (d *Driver) CreateTask(ctx context.Context, task *models.Task) error {
	// TODO: Implement task creation - for now return not implemented
	return storage.NewStorageError("create_task", "task", task.ID, storage.ErrUnsupportedOperation)
}

func (d *Driver) GetTask(ctx context.Context, id string) (*models.Task, error) {
	// TODO: Implement task retrieval - for now return not implemented
	return nil, storage.NewStorageError("get_task", "task", id, storage.ErrUnsupportedOperation)
}

func (d *Driver) UpdateTask(ctx context.Context, task *models.Task) error {
	// TODO: Implement task update - for now return not implemented
	return storage.NewStorageError("update_task", "task", task.ID, storage.ErrUnsupportedOperation)
}

func (d *Driver) DeleteTask(ctx context.Context, id string) error {
	// TODO: Implement task deletion - for now return not implemented
	return storage.NewStorageError("delete_task", "task", id, storage.ErrUnsupportedOperation)
}

func (d *Driver) ListTasks(ctx context.Context, filter storage.TaskFilter) ([]*models.Task, error) {
	// TODO: Implement task listing - for now return not implemented
	return nil, storage.NewStorageError("list_tasks", "task", "", storage.ErrUnsupportedOperation)
}

func (d *Driver) GetTasksByParent(ctx context.Context, parentID string) ([]*models.Task, error) {
	// TODO: Implement subtask retrieval - for now return not implemented
	return nil, storage.NewStorageError("get_tasks_by_parent", "task", parentID, storage.ErrUnsupportedOperation)
}

func (d *Driver) GetTaskDependencies(ctx context.Context, taskID string) ([]*models.Task, error) {
	// TODO: Implement dependency retrieval - for now return not implemented
	return nil, storage.NewStorageError("get_task_dependencies", "task", taskID, storage.ErrUnsupportedOperation)
}
