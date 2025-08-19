package models

import (
	"testing"

	"github.com/tr4d3r/ghcp-memory-context/pkg/types"
)

func TestNewTask(t *testing.T) {
	title := "Test Task"
	description := "This is a test task"

	task := NewTask(title, description)

	// Verify basic fields
	if task.Title != title {
		t.Errorf("Expected title %s, got %s", title, task.Title)
	}

	if task.Description != description {
		t.Errorf("Expected description %s, got %s", description, task.Description)
	}

	// Verify defaults
	if task.Status != TaskStatusPending {
		t.Errorf("Expected status %s, got %s", TaskStatusPending, task.Status)
	}

	if task.Priority != TaskPriorityMedium {
		t.Errorf("Expected priority %s, got %s", TaskPriorityMedium, task.Priority)
	}

	if task.Type != types.ContextTypeTask {
		t.Errorf("Expected type %s, got %s", types.ContextTypeTask, task.Type)
	}

	if task.Scope != types.ContextScopeLocal {
		t.Errorf("Expected scope %s, got %s", types.ContextScopeLocal, task.Scope)
	}

	// Verify ID was generated
	if task.ID == "" {
		t.Error("Expected non-empty ID")
	}

	// Verify slices are initialized
	if task.Subtasks == nil {
		t.Error("Expected Subtasks slice to be initialized")
	}

	if task.Dependencies == nil {
		t.Error("Expected Dependencies slice to be initialized")
	}

	if task.CodeRefs == nil {
		t.Error("Expected CodeRefs slice to be initialized")
	}
}

func TestNewTaskWithParent(t *testing.T) {
	parentID := "parent-123"
	task := NewTaskWithParent("Child Task", "Child description", parentID)

	if task.ParentID != parentID {
		t.Errorf("Expected parent ID %s, got %s", parentID, task.ParentID)
	}
}

func TestTaskValidation(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		expectError bool
	}{
		{
			name:        "Valid task",
			title:       "Valid Task",
			description: "Valid description",
			expectError: false,
		},
		{
			name:        "Empty title",
			title:       "",
			description: "Valid description",
			expectError: true,
		},
		{
			name:        "Title too long",
			title:       "This is a very long title that exceeds the maximum allowed length of 200 characters. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.",
			description: "Valid description",
			expectError: true,
		},
		{
			name:        "Description too long",
			title:       "Valid Title",
			description: "This is a very long description that exceeds the maximum allowed length of 1000 characters. Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt. Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem.",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask(tt.title, tt.description)
			err := task.Validate()

			if tt.expectError && err == nil {
				t.Error("Expected validation error, got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}

func TestTaskSubtasks(t *testing.T) {
	parent := NewTask("Parent Task", "Parent description")
	subtask1 := NewTask("Subtask 1", "First subtask")
	subtask2 := NewTask("Subtask 2", "Second subtask")

	// Test adding subtasks
	parent.AddSubtask(subtask1)
	parent.AddSubtask(subtask2)

	if len(parent.Subtasks) != 2 {
		t.Errorf("Expected 2 subtasks, got %d", len(parent.Subtasks))
	}

	if subtask1.ParentID != parent.ID {
		t.Errorf("Expected subtask1 parent ID %s, got %s", parent.ID, subtask1.ParentID)
	}

	// Test HasSubtasks
	if !parent.HasSubtasks() {
		t.Error("Expected parent to have subtasks")
	}

	// Test removing subtask
	removed := parent.RemoveSubtask(subtask1.ID)
	if !removed {
		t.Error("Expected subtask to be removed")
	}

	if len(parent.Subtasks) != 1 {
		t.Errorf("Expected 1 subtask after removal, got %d", len(parent.Subtasks))
	}

	// Test removing non-existent subtask
	removed = parent.RemoveSubtask("non-existent")
	if removed {
		t.Error("Expected removal of non-existent subtask to return false")
	}
}

func TestTaskDependencies(t *testing.T) {
	task := NewTask("Test Task", "Test description")
	dep1 := "dependency-1"
	dep2 := "dependency-2"

	// Test adding dependencies
	task.AddDependency(dep1)
	task.AddDependency(dep2)

	if len(task.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(task.Dependencies))
	}

	// Test adding duplicate dependency
	task.AddDependency(dep1)
	if len(task.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies after duplicate add, got %d", len(task.Dependencies))
	}

	// Test removing dependency
	removed := task.RemoveDependency(dep1)
	if !removed {
		t.Error("Expected dependency to be removed")
	}

	if len(task.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency after removal, got %d", len(task.Dependencies))
	}

	// Test removing non-existent dependency
	removed = task.RemoveDependency("non-existent")
	if removed {
		t.Error("Expected removal of non-existent dependency to return false")
	}
}

func TestTaskStatus(t *testing.T) {
	task := NewTask("Test Task", "Test description")

	// Test initial status
	if task.Status != TaskStatusPending {
		t.Errorf("Expected initial status %s, got %s", TaskStatusPending, task.Status)
	}

	// Test setting status to in-progress
	task.SetStatus(TaskStatusInProgress)
	if task.Status != TaskStatusInProgress {
		t.Errorf("Expected status %s, got %s", TaskStatusInProgress, task.Status)
	}

	if task.StartedAt == nil {
		t.Error("Expected StartedAt to be set when status changes to in-progress")
	}

	// Test setting status to completed
	task.SetStatus(TaskStatusCompleted)
	if task.Status != TaskStatusCompleted {
		t.Errorf("Expected status %s, got %s", TaskStatusCompleted, task.Status)
	}

	if task.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set when status changes to completed")
	}

	// Test status helper methods
	if !task.IsCompleted() {
		t.Error("Expected IsCompleted() to return true")
	}

	if task.IsBlocked() {
		t.Error("Expected IsBlocked() to return false")
	}

	// Test blocked status
	task.SetStatus(TaskStatusBlocked)
	if !task.IsBlocked() {
		t.Error("Expected IsBlocked() to return true")
	}
}

func TestTaskPriority(t *testing.T) {
	task := NewTask("Test Task", "Test description")

	// Test initial priority
	if task.Priority != TaskPriorityMedium {
		t.Errorf("Expected initial priority %s, got %s", TaskPriorityMedium, task.Priority)
	}

	// Test setting priority
	task.SetPriority(TaskPriorityHigh)
	if task.Priority != TaskPriorityHigh {
		t.Errorf("Expected priority %s, got %s", TaskPriorityHigh, task.Priority)
	}
}

func TestTaskCompletion(t *testing.T) {
	parent := NewTask("Parent Task", "Parent description")

	// Test completion percentage with no subtasks
	if parent.GetCompletionPercentage() != 0.0 {
		t.Errorf("Expected 0%% completion for non-completed task with no subtasks, got %.1f%%", parent.GetCompletionPercentage())
	}

	parent.SetStatus(TaskStatusCompleted)
	if parent.GetCompletionPercentage() != 100.0 {
		t.Errorf("Expected 100%% completion for completed task with no subtasks, got %.1f%%", parent.GetCompletionPercentage())
	}

	// Reset status and add subtasks
	parent.SetStatus(TaskStatusPending)
	subtask1 := NewTask("Subtask 1", "First subtask")
	subtask2 := NewTask("Subtask 2", "Second subtask")
	subtask3 := NewTask("Subtask 3", "Third subtask")

	parent.AddSubtask(subtask1)
	parent.AddSubtask(subtask2)
	parent.AddSubtask(subtask3)

	// Test completion with subtasks
	if parent.GetCompletedSubtasks() != 0 {
		t.Errorf("Expected 0 completed subtasks, got %d", parent.GetCompletedSubtasks())
	}

	if parent.GetCompletionPercentage() != 0.0 {
		t.Errorf("Expected 0%% completion with no completed subtasks, got %.1f%%", parent.GetCompletionPercentage())
	}

	// Complete one subtask
	subtask1.SetStatus(TaskStatusCompleted)
	if parent.GetCompletedSubtasks() != 1 {
		t.Errorf("Expected 1 completed subtask, got %d", parent.GetCompletedSubtasks())
	}

	expected := 100.0 / 3.0 // 33.33%
	if actual := parent.GetCompletionPercentage(); actual < expected-0.1 || actual > expected+0.1 {
		t.Errorf("Expected %.1f%% completion, got %.1f%%", expected, actual)
	}

	// Complete all subtasks
	subtask2.SetStatus(TaskStatusCompleted)
	subtask3.SetStatus(TaskStatusCompleted)

	if parent.GetCompletedSubtasks() != 3 {
		t.Errorf("Expected 3 completed subtasks, got %d", parent.GetCompletedSubtasks())
	}

	if parent.GetCompletionPercentage() != 100.0 {
		t.Errorf("Expected 100%% completion with all subtasks completed, got %.1f%%", parent.GetCompletionPercentage())
	}
}

func TestTaskCodeReferences(t *testing.T) {
	task := NewTask("Test Task", "Test description")

	codeRef := CodeReference{
		FilePath:    "src/main.go",
		LineStart:   10,
		LineEnd:     20,
		Description: "Main function implementation",
		CommitHash:  "abcd1234567890abcd1234567890abcd12345678", // pragma: allowlist secret
	}

	task.AddCodeReference(codeRef)

	if len(task.CodeRefs) != 1 {
		t.Errorf("Expected 1 code reference, got %d", len(task.CodeRefs))
	}

	if task.CodeRefs[0].FilePath != codeRef.FilePath {
		t.Errorf("Expected file path %s, got %s", codeRef.FilePath, task.CodeRefs[0].FilePath)
	}
}

func TestTaskJSONSerialization(t *testing.T) {
	task := NewTask("Test Task", "Test description")
	task.SetStatus(TaskStatusInProgress)
	task.SetPriority(TaskPriorityHigh)

	// Test ToJSON
	jsonData, err := task.ToJSON()
	if err != nil {
		t.Errorf("Error serializing task to JSON: %v", err)
	}

	// Test FromJSON
	newTask := &Task{}
	err = newTask.FromJSON(jsonData)
	if err != nil {
		t.Errorf("Error deserializing task from JSON: %v", err)
	}

	// Verify fields were preserved
	if newTask.Title != task.Title {
		t.Errorf("Expected title %s, got %s", task.Title, newTask.Title)
	}

	if newTask.Description != task.Description {
		t.Errorf("Expected description %s, got %s", task.Description, newTask.Description)
	}

	if newTask.Status != task.Status {
		t.Errorf("Expected status %s, got %s", task.Status, newTask.Status)
	}

	if newTask.Priority != task.Priority {
		t.Errorf("Expected priority %s, got %s", task.Priority, newTask.Priority)
	}
}

func TestTaskValidationEdgeCases(t *testing.T) {
	// Test task with invalid UUID in dependencies
	task := NewTask("Test Task", "Test description")
	task.Dependencies = []string{"invalid-uuid"}

	err := task.Validate()
	if err == nil {
		t.Error("Expected validation error for invalid UUID in dependencies")
	}

	// Test task with invalid parent ID
	task2 := NewTask("Test Task 2", "Test description")
	task2.ParentID = "invalid-uuid"

	err = task2.Validate()
	if err == nil {
		t.Error("Expected validation error for invalid parent UUID")
	}
}

func TestContextObjectInterface(t *testing.T) {
	task := NewTask("Test Task", "Test description")

	// Verify task implements ContextObject interface
	var obj types.ContextObject = task

	if obj.GetID() != task.ID {
		t.Errorf("Expected ID %s, got %s", task.ID, obj.GetID())
	}

	if obj.GetType() != types.ContextTypeTask {
		t.Errorf("Expected type %s, got %s", types.ContextTypeTask, obj.GetType())
	}

	if obj.GetVersion() != task.Version {
		t.Errorf("Expected version %s, got %s", task.Version, obj.GetVersion())
	}

	if obj.GetTimestamp() != task.Timestamp {
		t.Errorf("Expected timestamp %d, got %d", task.Timestamp, obj.GetTimestamp())
	}
}
