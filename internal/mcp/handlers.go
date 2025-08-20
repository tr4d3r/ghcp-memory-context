package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/tr4d3r/ghcp-memory-context/internal/models"
)

// handleRememberFact implements the remember_fact tool
func (s *StdioServer) handleRememberFact(ctx context.Context, args map[string]interface{}) CallToolResult {
	entityName, ok := args["entityName"].(string)
	if !ok || entityName == "" {
		return CallToolResult{
			Content: []ToolContent{{Type: "text", Text: "Error: entityName is required"}},
			IsError: true,
		}
	}

	observation, ok := args["observation"].(string)
	if !ok || observation == "" {
		return CallToolResult{
			Content: []ToolContent{{Type: "text", Text: "Error: observation is required"}},
			IsError: true,
		}
	}

	entityType, _ := args["entityType"].(string)
	source, _ := args["source"].(string)

	// Get or create entity
	var entity *models.Entity
	var err error

	if s.store.EntityExists(entityName) {
		entity, err = s.store.GetEntity(ctx, entityName)
		if err != nil {
			return CallToolResult{
				Content: []ToolContent{{Type: "text", Text: "Error: Failed to get entity"}},
				IsError: true,
			}
		}
	} else {
		if entityType == "" {
			entityType = "memory"
		}
		entity = models.NewEntity(entityName, entityType)
		if err := s.store.CreateEntity(ctx, entity); err != nil {
			return CallToolResult{
				Content: []ToolContent{{Type: "text", Text: "Error: Failed to create entity"}},
				IsError: true,
			}
		}
	}

	// Add observation
	if source != "" {
		entity.AddObservationWithSource(observation, source)
	} else {
		entity.AddObservation(observation)
	}

	if err := s.store.UpdateEntity(ctx, entity); err != nil {
		return CallToolResult{
			Content: []ToolContent{{Type: "text", Text: "Error: Failed to update entity"}},
			IsError: true,
		}
	}

	return CallToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("✓ Remembered: %s", observation),
		}},
	}
}

// handleRecallFacts implements the recall_facts tool
func (s *StdioServer) handleRecallFacts(ctx context.Context, args map[string]interface{}) CallToolResult {
	entityName, _ := args["entityName"].(string)
	entityType, _ := args["entityType"].(string)

	if entityName != "" {
		// Recall specific entity
		entity, err := s.store.GetEntity(ctx, entityName)
		if err != nil {
			return CallToolResult{
				Content: []ToolContent{{Type: "text", Text: "Entity not found"}},
			}
		}

		var text strings.Builder
		text.WriteString(fmt.Sprintf("Entity: %s (%s)\n", entity.Name, entity.EntityType))
		text.WriteString("Observations:\n")
		for i, obs := range entity.Observations {
			text.WriteString(fmt.Sprintf("%d. %s\n", i+1, obs.Text))
		}

		return CallToolResult{
			Content: []ToolContent{{Type: "text", Text: text.String()}},
		}
	} else {
		// List entities by type
		entities, err := s.store.ListEntities(ctx, entityType)
		if err != nil {
			return CallToolResult{
				Content: []ToolContent{{Type: "text", Text: "Error retrieving entities"}},
				IsError: true,
			}
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

		return CallToolResult{
			Content: []ToolContent{{Type: "text", Text: text.String()}},
		}
	}
}

// handleSearchMemory implements the search_memory tool
func (s *StdioServer) handleSearchMemory(ctx context.Context, args map[string]interface{}) CallToolResult {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return CallToolResult{
			Content: []ToolContent{{Type: "text", Text: "Error: query is required"}},
			IsError: true,
		}
	}

	entityType, _ := args["entityType"].(string)

	results, err := s.store.SearchObservations(ctx, query, entityType)
	if err != nil {
		return CallToolResult{
			Content: []ToolContent{{Type: "text", Text: "Error searching memory"}},
			IsError: true,
		}
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

	return CallToolResult{
		Content: []ToolContent{{Type: "text", Text: text.String()}},
	}
}
