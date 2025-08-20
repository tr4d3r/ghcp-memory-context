package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/tr4d3r/ghcp-memory-context/internal/models"
)

// CreateEntityRequest represents the request payload for creating an entity
type CreateEntityRequest struct {
	Name         string   `json:"name"`
	EntityType   string   `json:"entityType"`
	Observations []string `json:"observations,omitempty"`
}

// UpdateEntityRequest represents the request payload for updating an entity
type UpdateEntityRequest struct {
	EntityType   string   `json:"entityType,omitempty"`
	Observations []string `json:"observations,omitempty"`
}

// AddObservationRequest represents the request payload for adding an observation
type AddObservationRequest struct {
	Text   string `json:"text"`
	Source string `json:"source,omitempty"`
}

// handleEntities handles requests to /entities
func (r *Router) handleEntities(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()

	switch req.Method {
	case http.MethodGet:
		r.handleListEntities(w, req, ctx)
	case http.MethodPost:
		r.handleCreateEntity(w, req, ctx)
	default:
		r.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleEntityByName handles requests to /entities/{name}
func (r *Router) handleEntityByName(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	entityName := extractPathParam(req, "/entities/")

	if entityName == "" {
		r.writeErrorResponse(w, http.StatusBadRequest, "Entity name is required")
		return
	}

	switch req.Method {
	case http.MethodGet:
		r.handleGetEntity(w, req, ctx, entityName)
	case http.MethodPut:
		r.handleUpdateEntity(w, req, ctx, entityName)
	case http.MethodDelete:
		r.handleDeleteEntity(w, req, ctx, entityName)
	case http.MethodPost:
		// Handle adding observations to existing entity
		if req.URL.Query().Get("action") == "add-observation" {
			r.handleAddObservation(w, req, ctx, entityName)
		} else {
			r.writeErrorResponse(w, http.StatusBadRequest, "Invalid action for POST request")
		}
	default:
		r.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleListEntities lists all entities with optional filtering
func (r *Router) handleListEntities(w http.ResponseWriter, req *http.Request, ctx context.Context) {
	entityType := parseQueryParam(req, "type")

	entities, err := r.store.ListEntities(ctx, entityType)
	if err != nil {
		r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list entities: "+err.Error())
		return
	}

	r.writeSuccessResponse(w, entities, "Entities retrieved successfully")
}

// handleCreateEntity creates a new entity
func (r *Router) handleCreateEntity(w http.ResponseWriter, req *http.Request, ctx context.Context) {
	if err := validateJSONRequest(req); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var createReq CreateEntityRequest
	if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate required fields
	if err := validateRequiredFields(map[string]string{
		"name":       createReq.Name,
		"entityType": createReq.EntityType,
	}); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create entity
	entity := models.NewEntity(createReq.Name, createReq.EntityType)

	// Add observations if provided
	for _, obsText := range createReq.Observations {
		if obsText != "" {
			entity.AddObservation(obsText)
		}
	}

	// Store entity
	if err := r.store.CreateEntity(ctx, entity); err != nil {
		if err.Error() == "entity '"+createReq.Name+"' already exists" {
			r.writeErrorResponse(w, http.StatusConflict, err.Error())
		} else {
			r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create entity: "+err.Error())
		}
		return
	}

	r.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"data":    entity,
		"message": "Entity created successfully",
	})
}

// handleGetEntity retrieves a specific entity
func (r *Router) handleGetEntity(w http.ResponseWriter, req *http.Request, ctx context.Context, entityName string) {
	entity, err := r.store.GetEntity(ctx, entityName)
	if err != nil {
		if err.Error() == "entity '"+entityName+"' not found" {
			r.writeErrorResponse(w, http.StatusNotFound, "Entity not found")
		} else {
			r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get entity: "+err.Error())
		}
		return
	}

	r.writeSuccessResponse(w, entity, "Entity retrieved successfully")
}

// handleUpdateEntity updates an existing entity
func (r *Router) handleUpdateEntity(w http.ResponseWriter, req *http.Request, ctx context.Context, entityName string) {
	if err := validateJSONRequest(req); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get existing entity
	entity, err := r.store.GetEntity(ctx, entityName)
	if err != nil {
		if err.Error() == "entity '"+entityName+"' not found" {
			r.writeErrorResponse(w, http.StatusNotFound, "Entity not found")
		} else {
			r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get entity: "+err.Error())
		}
		return
	}

	var updateReq UpdateEntityRequest
	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Update entity type if provided
	if updateReq.EntityType != "" {
		entity.EntityType = updateReq.EntityType
	}

	// Replace observations if provided
	if updateReq.Observations != nil {
		entity.Observations = make([]models.Observation, 0)
		for _, obsText := range updateReq.Observations {
			if obsText != "" {
				entity.AddObservation(obsText)
			}
		}
	}

	// Save updated entity
	if err := r.store.UpdateEntity(ctx, entity); err != nil {
		r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update entity: "+err.Error())
		return
	}

	r.writeSuccessResponse(w, entity, "Entity updated successfully")
}

// handleDeleteEntity deletes an entity
func (r *Router) handleDeleteEntity(w http.ResponseWriter, req *http.Request, ctx context.Context, entityName string) {
	// Check if entity exists
	if !r.store.EntityExists(entityName) {
		r.writeErrorResponse(w, http.StatusNotFound, "Entity not found")
		return
	}

	if err := r.store.DeleteEntity(ctx, entityName); err != nil {
		r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete entity: "+err.Error())
		return
	}

	r.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Entity deleted successfully",
	})
}

// handleAddObservation adds an observation to an existing entity
func (r *Router) handleAddObservation(w http.ResponseWriter, req *http.Request, ctx context.Context, entityName string) {
	if err := validateJSONRequest(req); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get existing entity
	entity, err := r.store.GetEntity(ctx, entityName)
	if err != nil {
		if err.Error() == "entity '"+entityName+"' not found" {
			r.writeErrorResponse(w, http.StatusNotFound, "Entity not found")
		} else {
			r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get entity: "+err.Error())
		}
		return
	}

	var addObsReq AddObservationRequest
	if err := json.NewDecoder(req.Body).Decode(&addObsReq); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate required fields
	if err := validateRequiredFields(map[string]string{
		"text": addObsReq.Text,
	}); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Add observation
	if addObsReq.Source != "" {
		entity.AddObservationWithSource(addObsReq.Text, addObsReq.Source)
	} else {
		entity.AddObservation(addObsReq.Text)
	}

	// Save updated entity
	if err := r.store.UpdateEntity(ctx, entity); err != nil {
		r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update entity: "+err.Error())
		return
	}

	r.writeSuccessResponse(w, entity, "Observation added successfully")
}
