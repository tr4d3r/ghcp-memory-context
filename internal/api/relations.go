package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/tr4d3r/ghcp-memory-context/internal/models"
)

// CreateRelationRequest represents the request payload for creating a relation
type CreateRelationRequest struct {
	From         string `json:"from"`
	To           string `json:"to"`
	RelationType string `json:"relationType"`
}

// UpdateRelationRequest represents the request payload for updating a relation
type UpdateRelationRequest struct {
	From         string `json:"from,omitempty"`
	To           string `json:"to,omitempty"`
	RelationType string `json:"relationType,omitempty"`
}

// handleRelations handles requests to /relations
func (r *Router) handleRelations(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()

	switch req.Method {
	case http.MethodGet:
		r.handleListRelations(w, req, ctx)
	case http.MethodPost:
		r.handleCreateRelation(w, req, ctx)
	default:
		r.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleRelationByID handles requests to /relations/{id}
func (r *Router) handleRelationByID(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	relationID := extractPathParam(req, "/relations/")

	if relationID == "" {
		r.writeErrorResponse(w, http.StatusBadRequest, "Relation ID is required")
		return
	}

	switch req.Method {
	case http.MethodGet:
		r.handleGetRelation(w, req, ctx, relationID)
	case http.MethodPut:
		r.handleUpdateRelation(w, req, ctx, relationID)
	case http.MethodDelete:
		r.handleDeleteRelation(w, req, ctx, relationID)
	default:
		r.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleListRelations lists all relations with optional filtering
func (r *Router) handleListRelations(w http.ResponseWriter, req *http.Request, ctx context.Context) {
	fromEntity := parseQueryParam(req, "from")
	toEntity := parseQueryParam(req, "to")
	relationType := parseQueryParam(req, "type")

	relations, err := r.store.GetRelations(ctx)
	if err != nil {
		r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get relations: "+err.Error())
		return
	}

	// Apply filters
	var filteredRelations []models.Relation
	for _, relation := range relations.Relations {
		include := true

		if fromEntity != "" && relation.From != fromEntity {
			include = false
		}
		if toEntity != "" && relation.To != toEntity {
			include = false
		}
		if relationType != "" && relation.RelationType != relationType {
			include = false
		}

		if include {
			filteredRelations = append(filteredRelations, relation)
		}
	}

	r.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":    filteredRelations,
		"message": "Relations retrieved successfully",
		"count":   len(filteredRelations),
	})
}

// handleCreateRelation creates a new relation
func (r *Router) handleCreateRelation(w http.ResponseWriter, req *http.Request, ctx context.Context) {
	if err := validateJSONRequest(req); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var createReq CreateRelationRequest
	if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate required fields
	if err := validateRequiredFields(map[string]string{
		"from":         createReq.From,
		"to":           createReq.To,
		"relationType": createReq.RelationType,
	}); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Verify that both entities exist
	if !r.store.EntityExists(createReq.From) {
		r.writeErrorResponse(w, http.StatusBadRequest, "Source entity '"+createReq.From+"' does not exist")
		return
	}
	if !r.store.EntityExists(createReq.To) {
		r.writeErrorResponse(w, http.StatusBadRequest, "Target entity '"+createReq.To+"' does not exist")
		return
	}

	// Get current relations
	relations, err := r.store.GetRelations(ctx)
	if err != nil {
		r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get relations: "+err.Error())
		return
	}

	// Check if relation already exists
	for _, existing := range relations.Relations {
		if existing.From == createReq.From && existing.To == createReq.To && existing.RelationType == createReq.RelationType {
			r.writeErrorResponse(w, http.StatusConflict, "Relation already exists")
			return
		}
	}

	// Add new relation
	relations.AddRelation(createReq.From, createReq.To, createReq.RelationType)

	// Save relations
	if err := r.store.SaveRelations(ctx, relations); err != nil {
		r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to save relation: "+err.Error())
		return
	}

	// Return the newly created relation
	newRelation := relations.Relations[len(relations.Relations)-1]
	r.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"data":    newRelation,
		"message": "Relation created successfully",
	})
}

// handleGetRelation retrieves a specific relation by ID
func (r *Router) handleGetRelation(w http.ResponseWriter, req *http.Request, ctx context.Context, relationID string) {
	relations, err := r.store.GetRelations(ctx)
	if err != nil {
		r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get relations: "+err.Error())
		return
	}

	// Find relation by ID
	for _, relation := range relations.Relations {
		if relation.ID == relationID {
			r.writeSuccessResponse(w, relation, "Relation retrieved successfully")
			return
		}
	}

	r.writeErrorResponse(w, http.StatusNotFound, "Relation not found")
}

// handleUpdateRelation updates a specific relation
func (r *Router) handleUpdateRelation(w http.ResponseWriter, req *http.Request, ctx context.Context, relationID string) {
	if err := validateJSONRequest(req); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	relations, err := r.store.GetRelations(ctx)
	if err != nil {
		r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get relations: "+err.Error())
		return
	}

	// Find relation by ID
	var relationIndex = -1
	for i, relation := range relations.Relations {
		if relation.ID == relationID {
			relationIndex = i
			break
		}
	}

	if relationIndex == -1 {
		r.writeErrorResponse(w, http.StatusNotFound, "Relation not found")
		return
	}

	var updateReq UpdateRelationRequest
	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Update relation fields
	relation := &relations.Relations[relationIndex]
	if updateReq.From != "" {
		if !r.store.EntityExists(updateReq.From) {
			r.writeErrorResponse(w, http.StatusBadRequest, "Source entity '"+updateReq.From+"' does not exist")
			return
		}
		relation.From = updateReq.From
	}
	if updateReq.To != "" {
		if !r.store.EntityExists(updateReq.To) {
			r.writeErrorResponse(w, http.StatusBadRequest, "Target entity '"+updateReq.To+"' does not exist")
			return
		}
		relation.To = updateReq.To
	}
	if updateReq.RelationType != "" {
		relation.RelationType = updateReq.RelationType
	}

	// Save relations
	if err := r.store.SaveRelations(ctx, relations); err != nil {
		r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to save relation: "+err.Error())
		return
	}

	r.writeSuccessResponse(w, *relation, "Relation updated successfully")
}

// handleDeleteRelation deletes a specific relation
func (r *Router) handleDeleteRelation(w http.ResponseWriter, req *http.Request, ctx context.Context, relationID string) {
	relations, err := r.store.GetRelations(ctx)
	if err != nil {
		r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get relations: "+err.Error())
		return
	}

	// Remove relation by ID
	if !relations.RemoveRelation(relationID) {
		r.writeErrorResponse(w, http.StatusNotFound, "Relation not found")
		return
	}

	// Save relations
	if err := r.store.SaveRelations(ctx, relations); err != nil {
		r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to save relations: "+err.Error())
		return
	}

	r.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Relation deleted successfully",
	})
}
