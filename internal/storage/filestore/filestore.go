package filestore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tr4d3r/ghcp-memory-context/internal/models"
	"github.com/tr4d3r/ghcp-memory-context/internal/storage"
)

// FileStore implements file-based storage for entities and relations
type FileStore struct {
	baseDir       string
	entitiesDir   string
	relationsFile string

	// In-memory cache for performance
	entityCache   map[string]*models.Entity
	relationCache *models.RelationSet
	cacheMutex    sync.RWMutex

	// File locking for concurrent access
	fileLocks map[string]*sync.RWMutex
	lockMutex sync.Mutex
}

// NewFileStore creates a new file-based storage instance
func NewFileStore(baseDir string) *FileStore {
	entitiesDir := filepath.Join(baseDir, "entities")
	relationsFile := filepath.Join(baseDir, "relations", "relations.json")

	return &FileStore{
		baseDir:       baseDir,
		entitiesDir:   entitiesDir,
		relationsFile: relationsFile,
		entityCache:   make(map[string]*models.Entity),
		relationCache: &models.RelationSet{Relations: make([]models.Relation, 0)},
		fileLocks:     make(map[string]*sync.RWMutex),
	}
}

// Initialize creates the necessary directory structure
func (fs *FileStore) Initialize() error {
	// Create entities directory
	if err := os.MkdirAll(fs.entitiesDir, 0755); err != nil {
		return fmt.Errorf("failed to create entities directory: %w", err)
	}

	// Create relations directory
	relationsDir := filepath.Dir(fs.relationsFile)
	if err := os.MkdirAll(relationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create relations directory: %w", err)
	}

	// Create empty relations file if it doesn't exist
	if _, err := os.Stat(fs.relationsFile); os.IsNotExist(err) {
		emptyRelations := &models.RelationSet{Relations: make([]models.Relation, 0)}
		if err := fs.saveRelationsFile(emptyRelations); err != nil {
			return fmt.Errorf("failed to create relations file: %w", err)
		}
	}

	return nil
}

// Entity Operations

// CreateEntity creates a new entity and saves it to file
func (fs *FileStore) CreateEntity(ctx context.Context, entity *models.Entity) error {
	if err := entity.Validate(); err != nil {
		return fmt.Errorf("entity validation failed: %w", err)
	}

	// Check if entity already exists
	if fs.EntityExists(entity.Name) {
		return fmt.Errorf("entity '%s' already exists", entity.Name)
	}

	// Save to file
	if err := fs.saveEntityFile(entity); err != nil {
		return fmt.Errorf("failed to save entity: %w", err)
	}

	// Update cache
	fs.cacheMutex.Lock()
	fs.entityCache[entity.Name] = entity
	fs.cacheMutex.Unlock()

	return nil
}

// GetEntity retrieves an entity by name
func (fs *FileStore) GetEntity(ctx context.Context, name string) (*models.Entity, error) {
	// Check cache first
	fs.cacheMutex.RLock()
	if cached, exists := fs.entityCache[name]; exists {
		fs.cacheMutex.RUnlock()
		return cached, nil
	}
	fs.cacheMutex.RUnlock()

	// Load from file
	entity, err := fs.loadEntityFile(name)
	if err != nil {
		return nil, err
	}

	// Update cache
	fs.cacheMutex.Lock()
	fs.entityCache[name] = entity
	fs.cacheMutex.Unlock()

	return entity, nil
}

// UpdateEntity updates an existing entity
func (fs *FileStore) UpdateEntity(ctx context.Context, entity *models.Entity) error {
	if err := entity.Validate(); err != nil {
		return fmt.Errorf("entity validation failed: %w", err)
	}

	// Check if entity exists
	if !fs.EntityExists(entity.Name) {
		return fmt.Errorf("entity '%s' does not exist", entity.Name)
	}

	// Save to file
	if err := fs.saveEntityFile(entity); err != nil {
		return fmt.Errorf("failed to update entity: %w", err)
	}

	// Update cache
	fs.cacheMutex.Lock()
	fs.entityCache[entity.Name] = entity
	fs.cacheMutex.Unlock()

	return nil
}

// DeleteEntity removes an entity
func (fs *FileStore) DeleteEntity(ctx context.Context, name string) error {
	// Remove file
	filePath := fs.getEntityFilePath(name)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete entity file: %w", err)
	}

	// Remove from cache
	fs.cacheMutex.Lock()
	delete(fs.entityCache, name)
	fs.cacheMutex.Unlock()

	return nil
}

// ListEntities returns all entities, optionally filtered by type
func (fs *FileStore) ListEntities(ctx context.Context, entityType string) ([]*models.Entity, error) {
	files, err := os.ReadDir(fs.entitiesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read entities directory: %w", err)
	}

	var entities []*models.Entity
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			name := strings.TrimSuffix(file.Name(), ".json")
			entity, err := fs.GetEntity(ctx, name)
			if err != nil {
				continue // Skip invalid entities
			}

			// Apply type filter if specified
			if entityType == "" || entity.EntityType == entityType {
				entities = append(entities, entity)
			}
		}
	}

	return entities, nil
}

// EntityExists checks if an entity exists
func (fs *FileStore) EntityExists(name string) bool {
	filePath := fs.getEntityFilePath(name)
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// SearchObservations searches for observations across all entities
func (fs *FileStore) SearchObservations(ctx context.Context, query string, entityType string) ([]storage.SearchResult, error) {
	entities, err := fs.ListEntities(ctx, entityType)
	if err != nil {
		return nil, err
	}

	var results []storage.SearchResult
	for _, entity := range entities {
		observations := entity.SearchObservations(query)
		for _, obs := range observations {
			results = append(results, storage.SearchResult{
				EntityName:  entity.Name,
				EntityType:  entity.EntityType,
				Observation: obs,
			})
		}
	}

	return results, nil
}

// Relation Operations

// GetRelations returns all relations
func (fs *FileStore) GetRelations(ctx context.Context) (*models.RelationSet, error) {
	// Check cache first
	fs.cacheMutex.RLock()
	if fs.relationCache != nil && len(fs.relationCache.Relations) > 0 {
		fs.cacheMutex.RUnlock()
		return fs.relationCache, nil
	}
	fs.cacheMutex.RUnlock()

	// Load from file
	relations, err := fs.loadRelationsFile()
	if err != nil {
		return nil, err
	}

	// Update cache
	fs.cacheMutex.Lock()
	fs.relationCache = relations
	fs.cacheMutex.Unlock()

	return relations, nil
}

// SaveRelations saves the relation set
func (fs *FileStore) SaveRelations(ctx context.Context, relations *models.RelationSet) error {
	if err := fs.saveRelationsFile(relations); err != nil {
		return err
	}

	// Update cache
	fs.cacheMutex.Lock()
	fs.relationCache = relations
	fs.cacheMutex.Unlock()

	return nil
}

// File Operations

func (fs *FileStore) getEntityFilePath(name string) string {
	return filepath.Join(fs.entitiesDir, name+".json")
}

func (fs *FileStore) getFileLock(filePath string) *sync.RWMutex {
	fs.lockMutex.Lock()
	defer fs.lockMutex.Unlock()

	if lock, exists := fs.fileLocks[filePath]; exists {
		return lock
	}

	lock := &sync.RWMutex{}
	fs.fileLocks[filePath] = lock
	return lock
}

func (fs *FileStore) saveEntityFile(entity *models.Entity) error {
	filePath := fs.getEntityFilePath(entity.Name)
	lock := fs.getFileLock(filePath)

	lock.Lock()
	defer lock.Unlock()

	data, err := entity.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal entity: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

func (fs *FileStore) loadEntityFile(name string) (*models.Entity, error) {
	filePath := fs.getEntityFilePath(name)
	lock := fs.getFileLock(filePath)

	lock.RLock()
	defer lock.RUnlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("entity '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to read entity file: %w", err)
	}

	var entity models.Entity
	if err := entity.FromJSON(data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entity: %w", err)
	}

	return &entity, nil
}

func (fs *FileStore) saveRelationsFile(relations *models.RelationSet) error {
	lock := fs.getFileLock(fs.relationsFile)

	lock.Lock()
	defer lock.Unlock()

	data, err := relations.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal relations: %w", err)
	}

	return os.WriteFile(fs.relationsFile, data, 0644)
}

func (fs *FileStore) loadRelationsFile() (*models.RelationSet, error) {
	lock := fs.getFileLock(fs.relationsFile)

	lock.RLock()
	defer lock.RUnlock()

	data, err := os.ReadFile(fs.relationsFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty relation set if file doesn't exist
			return &models.RelationSet{Relations: make([]models.Relation, 0)}, nil
		}
		return nil, fmt.Errorf("failed to read relations file: %w", err)
	}

	var relations models.RelationSet
	if err := relations.FromJSON(data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal relations: %w", err)
	}

	return &relations, nil
}

// ClearCache clears the in-memory cache
func (fs *FileStore) ClearCache() {
	fs.cacheMutex.Lock()
	defer fs.cacheMutex.Unlock()

	fs.entityCache = make(map[string]*models.Entity)
	fs.relationCache = &models.RelationSet{Relations: make([]models.Relation, 0)}
}
