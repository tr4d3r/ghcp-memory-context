package filestore

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/tr4d3r/ghcp-memory-context/internal/models"
)

func setupTestFileStore(t *testing.T) (*FileStore, string) {
	tempDir, err := os.MkdirTemp("", "filestore_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	fs := NewFileStore(tempDir)
	if err := fs.Initialize(); err != nil {
		t.Fatalf("Failed to initialize FileStore: %v", err)
	}

	return fs, tempDir
}

func cleanup(tempDir string) {
	os.RemoveAll(tempDir)
}

func TestFileStoreInitialize(t *testing.T) {
	fs, tempDir := setupTestFileStore(t)
	defer cleanup(tempDir)

	// Check that directories were created
	if _, err := os.Stat(fs.entitiesDir); os.IsNotExist(err) {
		t.Error("Entities directory was not created")
	}

	if _, err := os.Stat(fs.relationsFile); os.IsNotExist(err) {
		t.Error("Relations file was not created")
	}
}

func TestEntityCRUD(t *testing.T) {
	fs, tempDir := setupTestFileStore(t)
	defer cleanup(tempDir)

	ctx := context.Background()

	// Create entity
	entity := models.NewEntity("project_standards", "guideline")
	entity.AddObservation("use conventional commits")

	err := fs.CreateEntity(ctx, entity)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Check that file was created
	expectedFile := filepath.Join(fs.entitiesDir, "project_standards.json")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Error("Entity file was not created")
	}

	// Read entity
	retrievedEntity, err := fs.GetEntity(ctx, "project_standards")
	if err != nil {
		t.Fatalf("Failed to get entity: %v", err)
	}

	if retrievedEntity.Name != entity.Name {
		t.Errorf("Expected entity name '%s', got '%s'", entity.Name, retrievedEntity.Name)
	}

	if retrievedEntity.GetObservationCount() != 1 {
		t.Errorf("Expected 1 observation, got %d", retrievedEntity.GetObservationCount())
	}

	// Update entity
	retrievedEntity.AddObservation("format: type(scope): description")
	err = fs.UpdateEntity(ctx, retrievedEntity)
	if err != nil {
		t.Fatalf("Failed to update entity: %v", err)
	}

	// Read updated entity
	updatedEntity, err := fs.GetEntity(ctx, "project_standards")
	if err != nil {
		t.Fatalf("Failed to get updated entity: %v", err)
	}

	if updatedEntity.GetObservationCount() != 2 {
		t.Errorf("Expected 2 observations after update, got %d", updatedEntity.GetObservationCount())
	}

	// Delete entity
	err = fs.DeleteEntity(ctx, "project_standards")
	if err != nil {
		t.Fatalf("Failed to delete entity: %v", err)
	}

	// Check that file was deleted
	if _, err := os.Stat(expectedFile); !os.IsNotExist(err) {
		t.Error("Entity file was not deleted")
	}

	// Try to get deleted entity
	_, err = fs.GetEntity(ctx, "project_standards")
	if err == nil {
		t.Error("Expected error when getting deleted entity")
	}
}

func TestEntityExists(t *testing.T) {
	fs, tempDir := setupTestFileStore(t)
	defer cleanup(tempDir)

	ctx := context.Background()

	// Check non-existent entity
	if fs.EntityExists("nonexistent") {
		t.Error("EntityExists returned true for non-existent entity")
	}

	// Create entity
	entity := models.NewEntity("test_entity", "test")
	err := fs.CreateEntity(ctx, entity)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// Check existing entity
	if !fs.EntityExists("test_entity") {
		t.Error("EntityExists returned false for existing entity")
	}
}

func TestListEntities(t *testing.T) {
	fs, tempDir := setupTestFileStore(t)
	defer cleanup(tempDir)

	ctx := context.Background()

	// Create multiple entities
	entity1 := models.NewEntity("standards", "guideline")
	entity2 := models.NewEntity("patterns", "pattern")
	entity3 := models.NewEntity("decisions", "guideline")

	err := fs.CreateEntity(ctx, entity1)
	if err != nil {
		t.Fatalf("Failed to create entity1: %v", err)
	}
	err = fs.CreateEntity(ctx, entity2)
	if err != nil {
		t.Fatalf("Failed to create entity2: %v", err)
	}
	err = fs.CreateEntity(ctx, entity3)
	if err != nil {
		t.Fatalf("Failed to create entity3: %v", err)
	}

	// List all entities
	allEntities, err := fs.ListEntities(ctx, "")
	if err != nil {
		t.Fatalf("Failed to list entities: %v", err)
	}

	if len(allEntities) != 3 {
		t.Errorf("Expected 3 entities, got %d", len(allEntities))
	}

	// List entities by type
	guidelines, err := fs.ListEntities(ctx, "guideline")
	if err != nil {
		t.Fatalf("Failed to list guideline entities: %v", err)
	}

	if len(guidelines) != 2 {
		t.Errorf("Expected 2 guideline entities, got %d", len(guidelines))
	}
}

func TestSearchObservations(t *testing.T) {
	fs, tempDir := setupTestFileStore(t)
	defer cleanup(tempDir)

	ctx := context.Background()

	// Create entities with observations
	entity1 := models.NewEntity("standards", "guideline")
	entity1.AddObservation("use conventional commits")
	entity1.AddObservation("format: type(scope): description")

	entity2 := models.NewEntity("patterns", "pattern")
	entity2.AddObservation("use REST API endpoints")
	entity2.AddObservation("implement commit hooks")

	err := fs.CreateEntity(ctx, entity1)
	if err != nil {
		t.Fatalf("Failed to create entity1: %v", err)
	}
	err = fs.CreateEntity(ctx, entity2)
	if err != nil {
		t.Fatalf("Failed to create entity2: %v", err)
	}

	// Search for "commit"
	results, err := fs.SearchObservations(ctx, "commit", "")
	if err != nil {
		t.Fatalf("Failed to search observations: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 search results for 'commit', got %d", len(results))
	}

	// Search with type filter
	guidelineResults, err := fs.SearchObservations(ctx, "commit", "guideline")
	if err != nil {
		t.Fatalf("Failed to search guideline observations: %v", err)
	}

	if len(guidelineResults) != 1 {
		t.Errorf("Expected 1 guideline result for 'commit', got %d", len(guidelineResults))
	}
}

func TestRelations(t *testing.T) {
	fs, tempDir := setupTestFileStore(t)
	defer cleanup(tempDir)

	ctx := context.Background()

	// Get initial empty relations
	relations, err := fs.GetRelations(ctx)
	if err != nil {
		t.Fatalf("Failed to get relations: %v", err)
	}

	if len(relations.Relations) != 0 {
		t.Errorf("Expected 0 initial relations, got %d", len(relations.Relations))
	}

	// Add relations
	relations.AddRelation("project", "standards", "follows")
	relations.AddRelation("project", "patterns", "implements")

	err = fs.SaveRelations(ctx, relations)
	if err != nil {
		t.Fatalf("Failed to save relations: %v", err)
	}

	// Get relations again
	savedRelations, err := fs.GetRelations(ctx)
	if err != nil {
		t.Fatalf("Failed to get saved relations: %v", err)
	}

	if len(savedRelations.Relations) != 2 {
		t.Errorf("Expected 2 saved relations, got %d", len(savedRelations.Relations))
	}
}

func TestCaching(t *testing.T) {
	fs, tempDir := setupTestFileStore(t)
	defer cleanup(tempDir)

	ctx := context.Background()

	// Create entity
	entity := models.NewEntity("test_cache", "test")
	err := fs.CreateEntity(ctx, entity)
	if err != nil {
		t.Fatalf("Failed to create entity: %v", err)
	}

	// First read (from file)
	entity1, err := fs.GetEntity(ctx, "test_cache")
	if err != nil {
		t.Fatalf("Failed to get entity: %v", err)
	}

	// Second read (from cache)
	entity2, err := fs.GetEntity(ctx, "test_cache")
	if err != nil {
		t.Fatalf("Failed to get cached entity: %v", err)
	}

	// Both should have same data
	if entity1.Name != entity2.Name {
		t.Error("Cached entity data differs from original")
	}

	// Clear cache and read again
	fs.ClearCache()
	entity3, err := fs.GetEntity(ctx, "test_cache")
	if err != nil {
		t.Fatalf("Failed to get entity after cache clear: %v", err)
	}

	if entity1.Name != entity3.Name {
		t.Error("Entity data differs after cache clear")
	}
}

func TestCreateDuplicateEntity(t *testing.T) {
	fs, tempDir := setupTestFileStore(t)
	defer cleanup(tempDir)

	ctx := context.Background()

	// Create entity
	entity := models.NewEntity("duplicate_test", "test")
	err := fs.CreateEntity(ctx, entity)
	if err != nil {
		t.Fatalf("Failed to create first entity: %v", err)
	}

	// Try to create duplicate
	duplicateEntity := models.NewEntity("duplicate_test", "test")
	err = fs.CreateEntity(ctx, duplicateEntity)
	if err == nil {
		t.Error("Expected error when creating duplicate entity")
	}
}
