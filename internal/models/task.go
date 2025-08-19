package models

import (
	"encoding/json"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/tr4d3r/ghcp-memory-context/pkg/types"
)

var validate *validator.Validate

// validateTaskStatus validates TaskStatus values
func validateTaskStatus(fl validator.FieldLevel) bool {
	status := TaskStatus(fl.Field().String())
	return status == TaskStatusPending || status == TaskStatusInProgress ||
		status == TaskStatusCompleted || status == TaskStatusBlocked ||
		status == TaskStatusCancelled
}

// validateTaskPriority validates TaskPriority values
func validateTaskPriority(fl validator.FieldLevel) bool {
	priority := TaskPriority(fl.Field().String())
	return priority == TaskPriorityLow || priority == TaskPriorityMedium ||
		priority == TaskPriorityHigh || priority == TaskPriorityCritical
}

func init() {
	validate = validator.New()
	// Register custom validations if needed
	if err := validate.RegisterValidation("taskstatus", validateTaskStatus); err != nil {
		// Log error in production, panic in development
		panic(err)
	}
	if err := validate.RegisterValidation("taskpriority", validateTaskPriority); err != nil {
		// Log error in production, panic in development
		panic(err)
	}
}

// TaskStatus represents the current status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusBlocked    TaskStatus = "blocked"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// TaskPriority represents the priority level of a task
type TaskPriority string

const (
	TaskPriorityLow      TaskPriority = "low"
	TaskPriorityMedium   TaskPriority = "medium"
	TaskPriorityHigh     TaskPriority = "high"
	TaskPriorityCritical TaskPriority = "critical"
)

// Task represents a task context object that extends BaseContext
// This is the primary context object for developer workflows
type Task struct {
	// Embed BaseContext for MCP compliance
	types.BaseContext

	// Task-specific fields as defined in PRD
	Title       string       `json:"title" validate:"required,min=1,max=200"`
	Description string       `json:"description" validate:"max=1000"`
	Status      TaskStatus   `json:"status" validate:"required,taskstatus"`
	Priority    TaskPriority `json:"priority" validate:"required,taskpriority"`

	// Hierarchical structure
	ParentID string  `json:"parent_id,omitempty" validate:"omitempty,uuid"`
	Subtasks []*Task `json:"subtasks,omitempty"`

	// Dependencies and relationships
	Dependencies []string `json:"dependencies,omitempty" validate:"dive,uuid"`

	// Code references for linking to implementation
	CodeRefs []CodeReference `json:"code_refs,omitempty"`

	// Assignment and estimation
	Assignee       string  `json:"assignee,omitempty"`
	EstimatedHours float64 `json:"estimated_hours,omitempty" validate:"min=0"`
	ActualHours    float64 `json:"actual_hours,omitempty" validate:"min=0"`

	// Dates
	DueDate     *time.Time `json:"due_date,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// CodeReference represents a reference to code files or locations
type CodeReference struct {
	FilePath    string `json:"file_path" validate:"required"`
	LineStart   int    `json:"line_start,omitempty" validate:"min=0"`
	LineEnd     int    `json:"line_end,omitempty" validate:"min=0,gtefield=LineStart"`
	Description string `json:"description,omitempty" validate:"max=500"`
	CommitHash  string `json:"commit_hash,omitempty" validate:"omitempty,len=40"`
}

// Implement ContextObject interface methods for Task

// Validate validates the Task struct and calls the base validation
func (t *Task) Validate() error {
	// Set task type
	t.Type = types.ContextTypeTask

	// Set default status if not provided
	if t.Status == "" {
		t.Status = TaskStatusPending
	}

	// Set default priority if not provided
	if t.Priority == "" {
		t.Priority = TaskPriorityMedium
	}

	// Set the Data field to a map representation for MCP compliance
	t.Data = map[string]interface{}{
		"title":        t.Title,
		"description":  t.Description,
		"status":       t.Status,
		"priority":     t.Priority,
		"parent_id":    t.ParentID,
		"dependencies": t.Dependencies,
		"code_refs":    t.CodeRefs,
		"assignee":     t.Assignee,
		"due_date":     t.DueDate,
	}

	// Validate base context first
	if err := t.BaseContext.Validate(); err != nil {
		return err
	}

	// Use the global validator for task-specific validation
	return validate.Struct(t)
}

// ToJSON marshals the Task to JSON
func (t *Task) ToJSON() ([]byte, error) {
	return json.Marshal(t)
}

// FromJSON unmarshals JSON data into the Task
func (t *Task) FromJSON(data []byte) error {
	return json.Unmarshal(data, t)
}

// Factory and Helper Functions

// NewTask creates a new Task with the given title and description
func NewTask(title, description string) *Task {
	task := &Task{
		BaseContext: types.BaseContext{
			Type:    types.ContextTypeTask,
			Scope:   types.ContextScopeLocal,
			Version: "1.0.0",
		},
		Title:        title,
		Description:  description,
		Status:       TaskStatusPending,
		Priority:     TaskPriorityMedium,
		Subtasks:     make([]*Task, 0),
		Dependencies: make([]string, 0),
		CodeRefs:     make([]CodeReference, 0),
	}

	// Validate and set defaults
	_ = task.Validate() // Ignore error - validation will be called again later

	return task
}

// NewTaskWithParent creates a new subtask under a parent task
func NewTaskWithParent(title, description, parentID string) *Task {
	task := NewTask(title, description)
	task.ParentID = parentID
	return task
}

// AddSubtask adds a subtask to this task
func (t *Task) AddSubtask(subtask *Task) {
	if subtask == nil {
		return
	}

	subtask.ParentID = t.ID
	t.Subtasks = append(t.Subtasks, subtask)
	t.UpdatedAt = time.Now()
}

// RemoveSubtask removes a subtask by ID
func (t *Task) RemoveSubtask(subtaskID string) bool {
	for i, subtask := range t.Subtasks {
		if subtask.ID == subtaskID {
			// Remove subtask from slice
			t.Subtasks = append(t.Subtasks[:i], t.Subtasks[i+1:]...)
			t.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// AddDependency adds a dependency to this task
func (t *Task) AddDependency(dependencyID string) {
	// Check if dependency already exists
	for _, dep := range t.Dependencies {
		if dep == dependencyID {
			return
		}
	}

	t.Dependencies = append(t.Dependencies, dependencyID)
	t.UpdatedAt = time.Now()
}

// RemoveDependency removes a dependency from this task
func (t *Task) RemoveDependency(dependencyID string) bool {
	for i, dep := range t.Dependencies {
		if dep == dependencyID {
			t.Dependencies = append(t.Dependencies[:i], t.Dependencies[i+1:]...)
			t.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// AddCodeReference adds a code reference to this task
func (t *Task) AddCodeReference(codeRef CodeReference) {
	t.CodeRefs = append(t.CodeRefs, codeRef)
	t.UpdatedAt = time.Now()
}

// SetStatus updates the task status and related timestamps
func (t *Task) SetStatus(status TaskStatus) {
	oldStatus := t.Status
	t.Status = status
	t.UpdatedAt = time.Now()

	// Update status-specific timestamps
	now := time.Now()
	switch status {
	case TaskStatusInProgress:
		if oldStatus == TaskStatusPending && t.StartedAt == nil {
			t.StartedAt = &now
		}
	case TaskStatusCompleted:
		if t.CompletedAt == nil {
			t.CompletedAt = &now
		}
	}
}

// SetPriority updates the task priority
func (t *Task) SetPriority(priority TaskPriority) {
	t.Priority = priority
	t.UpdatedAt = time.Now()
}

// IsCompleted returns true if the task is completed
func (t *Task) IsCompleted() bool {
	return t.Status == TaskStatusCompleted
}

// IsBlocked returns true if the task is blocked
func (t *Task) IsBlocked() bool {
	return t.Status == TaskStatusBlocked
}

// HasSubtasks returns true if the task has subtasks
func (t *Task) HasSubtasks() bool {
	return len(t.Subtasks) > 0
}

// GetCompletedSubtasks returns the number of completed subtasks
func (t *Task) GetCompletedSubtasks() int {
	count := 0
	for _, subtask := range t.Subtasks {
		if subtask.IsCompleted() {
			count++
		}
	}
	return count
}

// GetCompletionPercentage returns the completion percentage based on subtasks
func (t *Task) GetCompletionPercentage() float64 {
	if len(t.Subtasks) == 0 {
		if t.IsCompleted() {
			return 100.0
		}
		return 0.0
	}

	completed := t.GetCompletedSubtasks()
	return float64(completed) / float64(len(t.Subtasks)) * 100.0
}
