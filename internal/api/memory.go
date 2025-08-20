package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/tr4d3r/ghcp-memory-context/internal/models"
)

// RememberRequest represents the request payload for remembering a fact
type RememberRequest struct {
	EntityName  string `json:"entityName"`
	EntityType  string `json:"entityType,omitempty"`
	Observation string `json:"observation"`
	Source      string `json:"source,omitempty"`
}

// RecallRequest represents the request payload for recalling facts
type RecallRequest struct {
	EntityName string `json:"entityName,omitempty"`
	EntityType string `json:"entityType,omitempty"`
	Query      string `json:"query,omitempty"`
}

// SearchRequest represents the request payload for searching memory
type SearchRequest struct {
	Query      string `json:"query"`
	EntityType string `json:"entityType,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

// handleMemoryRemember handles the /memory/remember endpoint
// This is the core "remember X" operation for storing atomic facts
func (r *Router) handleMemoryRemember(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		r.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	ctx := context.Background()

	if err := validateJSONRequest(req); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var rememberReq RememberRequest
	if err := json.NewDecoder(req.Body).Decode(&rememberReq); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate required fields
	if err := validateRequiredFields(map[string]string{
		"entityName":  rememberReq.EntityName,
		"observation": rememberReq.Observation,
	}); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if entity exists, create if it doesn't
	var entity *models.Entity
	var err error

	if r.store.EntityExists(rememberReq.EntityName) {
		// Get existing entity
		entity, err = r.store.GetEntity(ctx, rememberReq.EntityName)
		if err != nil {
			r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get entity: "+err.Error())
			return
		}
	} else {
		// Create new entity
		entityType := rememberReq.EntityType
		if entityType == "" {
			entityType = "memory" // Default type
		}

		entity = models.NewEntity(rememberReq.EntityName, entityType)
		if err := r.store.CreateEntity(ctx, entity); err != nil {
			r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create entity: "+err.Error())
			return
		}
	}

	// Add the observation
	if rememberReq.Source != "" {
		entity.AddObservationWithSource(rememberReq.Observation, rememberReq.Source)
	} else {
		entity.AddObservation(rememberReq.Observation)
	}

	// Update entity in storage
	if err := r.store.UpdateEntity(ctx, entity); err != nil {
		r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update entity: "+err.Error())
		return
	}

	r.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":     "Memory stored successfully",
		"entity":      entity.Name,
		"observation": entity.Observations[len(entity.Observations)-1], // Return the newly added observation
	})
}

// handleMemoryRecall handles the /memory/recall endpoint
// This retrieves specific facts or entities from memory
func (r *Router) handleMemoryRecall(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()

	switch req.Method {
	case http.MethodGet:
		r.handleMemoryRecallGET(w, req, ctx)
	case http.MethodPost:
		r.handleMemoryRecallPOST(w, req, ctx)
	default:
		r.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleMemoryRecallGET handles GET requests to /memory/recall with query parameters
func (r *Router) handleMemoryRecallGET(w http.ResponseWriter, req *http.Request, ctx context.Context) {
	entityName := parseQueryParam(req, "entity")
	entityType := parseQueryParam(req, "type")
	query := parseQueryParam(req, "query")

	if entityName != "" {
		// Recall specific entity
		entity, err := r.store.GetEntity(ctx, entityName)
		if err != nil {
			if err.Error() == "entity '"+entityName+"' not found" {
				r.writeErrorResponse(w, http.StatusNotFound, "Entity not found")
			} else {
				r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get entity: "+err.Error())
			}
			return
		}

		r.writeSuccessResponse(w, entity, "Entity recalled successfully")
	} else if query != "" {
		// Search across observations
		results, err := r.store.SearchObservations(ctx, query, entityType)
		if err != nil {
			r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to search observations: "+err.Error())
			return
		}

		r.writeSuccessResponse(w, results, "Memory search completed")
	} else {
		// List entities by type
		entities, err := r.store.ListEntities(ctx, entityType)
		if err != nil {
			r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list entities: "+err.Error())
			return
		}

		r.writeSuccessResponse(w, entities, "Entities recalled successfully")
	}
}

// handleMemoryRecallPOST handles POST requests to /memory/recall with JSON payload
func (r *Router) handleMemoryRecallPOST(w http.ResponseWriter, req *http.Request, ctx context.Context) {
	if err := validateJSONRequest(req); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var recallReq RecallRequest
	if err := json.NewDecoder(req.Body).Decode(&recallReq); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if recallReq.EntityName != "" {
		// Recall specific entity
		entity, err := r.store.GetEntity(ctx, recallReq.EntityName)
		if err != nil {
			if err.Error() == "entity '"+recallReq.EntityName+"' not found" {
				r.writeErrorResponse(w, http.StatusNotFound, "Entity not found")
			} else {
				r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get entity: "+err.Error())
			}
			return
		}

		r.writeSuccessResponse(w, entity, "Entity recalled successfully")
	} else if recallReq.Query != "" {
		// Search across observations
		results, err := r.store.SearchObservations(ctx, recallReq.Query, recallReq.EntityType)
		if err != nil {
			r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to search observations: "+err.Error())
			return
		}

		r.writeSuccessResponse(w, results, "Memory search completed")
	} else {
		// List entities by type
		entities, err := r.store.ListEntities(ctx, recallReq.EntityType)
		if err != nil {
			r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list entities: "+err.Error())
			return
		}

		r.writeSuccessResponse(w, entities, "Entities recalled successfully")
	}
}

// handleMemorySearch handles the /memory/search endpoint
// This provides advanced search capabilities across all memory
func (r *Router) handleMemorySearch(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()

	switch req.Method {
	case http.MethodGet:
		r.handleMemorySearchGET(w, req, ctx)
	case http.MethodPost:
		r.handleMemorySearchPOST(w, req, ctx)
	default:
		r.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleMemorySearchGET handles GET requests to /memory/search with query parameters
func (r *Router) handleMemorySearchGET(w http.ResponseWriter, req *http.Request, ctx context.Context) {
	query := parseQueryParam(req, "q")
	entityType := parseQueryParam(req, "type")
	limit := parseIntQueryParam(req, "limit", 50) // Default limit of 50

	if query == "" {
		r.writeErrorResponse(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	results, err := r.store.SearchObservations(ctx, query, entityType)
	if err != nil {
		r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to search observations: "+err.Error())
		return
	}

	// Apply limit
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	r.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":    results,
		"message": "Memory search completed",
		"count":   len(results),
		"query":   query,
	})
}

// handleMemorySearchPOST handles POST requests to /memory/search with JSON payload
func (r *Router) handleMemorySearchPOST(w http.ResponseWriter, req *http.Request, ctx context.Context) {
	if err := validateJSONRequest(req); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var searchReq SearchRequest
	if err := json.NewDecoder(req.Body).Decode(&searchReq); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if searchReq.Query == "" {
		r.writeErrorResponse(w, http.StatusBadRequest, "Query is required")
		return
	}

	results, err := r.store.SearchObservations(ctx, searchReq.Query, searchReq.EntityType)
	if err != nil {
		r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to search observations: "+err.Error())
		return
	}

	// Apply limit
	if searchReq.Limit > 0 && len(results) > searchReq.Limit {
		results = results[:searchReq.Limit]
	}

	r.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":    results,
		"message": "Memory search completed",
		"count":   len(results),
		"query":   searchReq.Query,
	})
}
