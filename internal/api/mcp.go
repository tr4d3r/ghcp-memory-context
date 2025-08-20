package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tr4d3r/ghcp-memory-context/internal/models"
)

// MCPResource represents an MCP resource
type MCPResource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType"`
}

// MCPToolCall represents an MCP tool call request
type MCPToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// MCPToolResult represents an MCP tool call result
type MCPToolResult struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// MCPContent represents content in MCP responses
type MCPContent struct {
	Type string      `json:"type"`
	Text string      `json:"text,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

// handleMCPResources handles the /mcp/resources endpoint
// Lists available memory resources for MCP clients
func (r *Router) handleMCPResources(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		r.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	ctx := context.Background()

	// Get all entities to create resource list
	entities, err := r.store.ListEntities(ctx, "")
	if err != nil {
		r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list entities: "+err.Error())
		return
	}

	var resources []MCPResource
	for _, entity := range entities {
		resource := MCPResource{
			URI:         fmt.Sprintf("memory://entities/%s", entity.Name),
			Name:        entity.Name,
			Description: fmt.Sprintf("%s entity with %d observations", entity.EntityType, entity.GetObservationCount()),
			MimeType:    "application/json",
		}
		resources = append(resources, resource)
	}

	// Add special resources for search and relations
	resources = append(resources, MCPResource{
		URI:         "memory://search",
		Name:        "Memory Search",
		Description: "Search across all memory context",
		MimeType:    "application/json",
	})

	resources = append(resources, MCPResource{
		URI:         "memory://relations",
		Name:        "Entity Relations",
		Description: "Relationships between entities",
		MimeType:    "application/json",
	})

	r.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"resources": resources,
	})
}

// handleMCPResourceByURI handles the /mcp/resources/{uri} endpoint
// Retrieves content for a specific MCP resource
func (r *Router) handleMCPResourceByURI(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		r.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	ctx := context.Background()
	resourceURI := extractPathParam(req, "/mcp/resources/")

	if resourceURI == "" {
		r.writeErrorResponse(w, http.StatusBadRequest, "Resource URI is required")
		return
	}

	// Decode the URI (it might be URL encoded)
	if strings.HasPrefix(resourceURI, "memory://entities/") {
		entityName := strings.TrimPrefix(resourceURI, "memory://entities/")
		entity, err := r.store.GetEntity(ctx, entityName)
		if err != nil {
			r.writeErrorResponse(w, http.StatusNotFound, "Entity not found")
			return
		}

		content := MCPContent{
			Type: "text",
			Text: fmt.Sprintf("Entity: %s\nType: %s\nObservations:\n", entity.Name, entity.EntityType),
		}

		for i, obs := range entity.Observations {
			content.Text += fmt.Sprintf("%d. %s (source: %s)\n", i+1, obs.Text, obs.Source)
		}

		r.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
			"contents": []MCPContent{content},
		})
	} else if resourceURI == "memory://search" {
		content := MCPContent{
			Type: "text",
			Text: "Memory Search Resource - Use search tools to query memory context",
		}
		r.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
			"contents": []MCPContent{content},
		})
	} else if resourceURI == "memory://relations" {
		relations, err := r.store.GetRelations(ctx)
		if err != nil {
			r.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get relations: "+err.Error())
			return
		}

		content := MCPContent{
			Type: "text",
			Text: "Entity Relations:\n",
		}

		for _, rel := range relations.Relations {
			content.Text += fmt.Sprintf("- %s %s %s\n", rel.From, rel.RelationType, rel.To)
		}

		r.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
			"contents": []MCPContent{content},
		})
	} else {
		r.writeErrorResponse(w, http.StatusNotFound, "Resource not found")
	}
}

// handleMCPRememberFact handles the /mcp/tools/remember_fact endpoint
// MCP tool for storing atomic facts
func (r *Router) handleMCPRememberFact(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		r.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	ctx := context.Background()

	var toolCall MCPToolCall
	if err := json.NewDecoder(req.Body).Decode(&toolCall); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Extract arguments
	entityName, ok := toolCall.Arguments["entityName"].(string)
	if !ok || entityName == "" {
		r.writeErrorResponse(w, http.StatusBadRequest, "entityName argument is required")
		return
	}

	observation, ok := toolCall.Arguments["observation"].(string)
	if !ok || observation == "" {
		r.writeErrorResponse(w, http.StatusBadRequest, "observation argument is required")
		return
	}

	entityType, _ := toolCall.Arguments["entityType"].(string)
	source, _ := toolCall.Arguments["source"].(string)

	// Use the same logic as /memory/remember
	var entity *models.Entity
	var err error

	if r.store.EntityExists(entityName) {
		entity, err = r.store.GetEntity(ctx, entityName)
		if err != nil {
			result := MCPToolResult{
				Content: []MCPContent{{Type: "text", Text: "Error: Failed to get entity"}},
				IsError: true,
			}
			r.writeJSONResponse(w, http.StatusInternalServerError, result)
			return
		}
	} else {
		if entityType == "" {
			entityType = "memory"
		}
		entity = models.NewEntity(entityName, entityType)
		if err := r.store.CreateEntity(ctx, entity); err != nil {
			result := MCPToolResult{
				Content: []MCPContent{{Type: "text", Text: "Error: Failed to create entity"}},
				IsError: true,
			}
			r.writeJSONResponse(w, http.StatusInternalServerError, result)
			return
		}
	}

	// Add observation
	if source != "" {
		entity.AddObservationWithSource(observation, source)
	} else {
		entity.AddObservation(observation)
	}

	if err := r.store.UpdateEntity(ctx, entity); err != nil {
		result := MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "Error: Failed to update entity"}},
			IsError: true,
		}
		r.writeJSONResponse(w, http.StatusInternalServerError, result)
		return
	}

	result := MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: fmt.Sprintf("Successfully remembered: %s", observation),
		}},
	}
	r.writeJSONResponse(w, http.StatusOK, result)
}

// handleMCPRecallFacts handles the /mcp/tools/recall_facts endpoint
// MCP tool for retrieving stored facts
func (r *Router) handleMCPRecallFacts(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		r.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	ctx := context.Background()

	var toolCall MCPToolCall
	if err := json.NewDecoder(req.Body).Decode(&toolCall); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	entityName, _ := toolCall.Arguments["entityName"].(string)
	entityType, _ := toolCall.Arguments["entityType"].(string)

	if entityName != "" {
		// Recall specific entity
		entity, err := r.store.GetEntity(ctx, entityName)
		if err != nil {
			result := MCPToolResult{
				Content: []MCPContent{{Type: "text", Text: "Entity not found"}},
			}
			r.writeJSONResponse(w, http.StatusOK, result)
			return
		}

		var text strings.Builder
		text.WriteString(fmt.Sprintf("Entity: %s (%s)\n", entity.Name, entity.EntityType))
		text.WriteString("Observations:\n")
		for i, obs := range entity.Observations {
			text.WriteString(fmt.Sprintf("%d. %s\n", i+1, obs.Text))
		}

		result := MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: text.String()}},
		}
		r.writeJSONResponse(w, http.StatusOK, result)
	} else {
		// List entities by type
		entities, err := r.store.ListEntities(ctx, entityType)
		if err != nil {
			result := MCPToolResult{
				Content: []MCPContent{{Type: "text", Text: "Error retrieving entities"}},
				IsError: true,
			}
			r.writeJSONResponse(w, http.StatusInternalServerError, result)
			return
		}

		var text strings.Builder
		if entityType != "" {
			text.WriteString(fmt.Sprintf("Entities of type '%s':\n", entityType))
		} else {
			text.WriteString("All entities:\n")
		}

		for _, entity := range entities {
			text.WriteString(fmt.Sprintf("- %s (%s): %d observations\n",
				entity.Name, entity.EntityType, entity.GetObservationCount()))
		}

		result := MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: text.String()}},
		}
		r.writeJSONResponse(w, http.StatusOK, result)
	}
}

// handleMCPSearchMemory handles the /mcp/tools/search_memory endpoint
// MCP tool for searching across all memory
func (r *Router) handleMCPSearchMemory(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		r.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	ctx := context.Background()

	var toolCall MCPToolCall
	if err := json.NewDecoder(req.Body).Decode(&toolCall); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	query, ok := toolCall.Arguments["query"].(string)
	if !ok || query == "" {
		r.writeErrorResponse(w, http.StatusBadRequest, "query argument is required")
		return
	}

	entityType, _ := toolCall.Arguments["entityType"].(string)

	results, err := r.store.SearchObservations(ctx, query, entityType)
	if err != nil {
		result := MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "Error searching memory"}},
			IsError: true,
		}
		r.writeJSONResponse(w, http.StatusInternalServerError, result)
		return
	}

	var text strings.Builder
	text.WriteString(fmt.Sprintf("Search results for '%s':\n", query))

	if len(results) == 0 {
		text.WriteString("No results found.\n")
	} else {
		for i, result := range results {
			text.WriteString(fmt.Sprintf("%d. [%s] %s: %s\n",
				i+1, result.EntityType, result.EntityName, result.Observation.Text))
		}
	}

	mcpResult := MCPToolResult{
		Content: []MCPContent{{Type: "text", Text: text.String()}},
	}
	r.writeJSONResponse(w, http.StatusOK, mcpResult)
}

// handleMCPForgetFact handles the /mcp/tools/forget_fact endpoint
// MCP tool for removing specific observations
func (r *Router) handleMCPForgetFact(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		r.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	ctx := context.Background()

	var toolCall MCPToolCall
	if err := json.NewDecoder(req.Body).Decode(&toolCall); err != nil {
		r.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	entityName, ok := toolCall.Arguments["entityName"].(string)
	if !ok || entityName == "" {
		r.writeErrorResponse(w, http.StatusBadRequest, "entityName argument is required")
		return
	}

	observationID, ok := toolCall.Arguments["observationId"].(string)
	if !ok || observationID == "" {
		r.writeErrorResponse(w, http.StatusBadRequest, "observationId argument is required")
		return
	}

	entity, err := r.store.GetEntity(ctx, entityName)
	if err != nil {
		result := MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "Entity not found"}},
		}
		r.writeJSONResponse(w, http.StatusOK, result)
		return
	}

	if entity.RemoveObservation(observationID) {
		if err := r.store.UpdateEntity(ctx, entity); err != nil {
			result := MCPToolResult{
				Content: []MCPContent{{Type: "text", Text: "Error updating entity"}},
				IsError: true,
			}
			r.writeJSONResponse(w, http.StatusInternalServerError, result)
			return
		}

		result := MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "Observation removed successfully"}},
		}
		r.writeJSONResponse(w, http.StatusOK, result)
	} else {
		result := MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "Observation not found"}},
		}
		r.writeJSONResponse(w, http.StatusOK, result)
	}
}
