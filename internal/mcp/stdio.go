package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/tr4d3r/ghcp-memory-context/internal/storage"
)

// StdioServer represents an MCP server that communicates over stdin/stdout
type StdioServer struct {
	store       storage.Storage
	initialized bool
}

// NewStdioServer creates a new MCP stdio server
func NewStdioServer(store storage.Storage) *StdioServer {
	return &StdioServer{
		store: store,
	}
}

// Run starts the stdio server loop
func (s *StdioServer) Run() error {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var request MCPRequest
		if err := json.Unmarshal(line, &request); err != nil {
			s.sendError(nil, ParseError, "Parse error: "+err.Error())
			continue
		}

		response := s.handleRequest(request)
		if response != nil {
			s.sendResponse(*response)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading from stdin: %w", err)
	}

	return nil
}

// handleRequest processes an MCP request and returns a response
func (s *StdioServer) handleRequest(request MCPRequest) *MCPResponse {
	switch request.Method {
	case "initialize":
		return s.handleInitialize(request)
	case "initialized":
		// Client confirms initialization is complete
		s.initialized = true
		return nil // No response required for notifications
	case "tools/list":
		return s.handleToolsList(request)
	case "tools/call":
		return s.handleToolsCall(request)
	case "resources/list":
		return s.handleResourcesList(request)
	case "resources/read":
		return s.handleResourcesRead(request)
	default:
		return s.createErrorResponse(request.ID, MethodNotFound, "Method not found: "+request.Method)
	}
}

// handleInitialize processes the initialize request
func (s *StdioServer) handleInitialize(request MCPRequest) *MCPResponse {
	var params InitializeParams
	if request.Params != nil {
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return s.createErrorResponse(request.ID, InvalidParams, "Invalid parameters: "+err.Error())
		}
	}

	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		ServerInfo: ServerInfo{
			Name:    "ghcp-memory-context",
			Version: "1.0.0",
		},
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
			Resources: &ResourcesCapability{
				Subscribe:   false,
				ListChanged: false,
			},
		},
		Instructions: `Memory Context Server - Persistent AI Assistant Memory

Natural language patterns:
• "Remember that..." → stores facts via remember_fact
• "What do you know about..." → retrieves via recall_facts
• "Search memory for..." → searches via search_memory

Quick tips:
• Facts persist across all sessions
• Organize with entity types: user, project, guideline, pattern, decision
• Use natural conversational language - the AI will translate to appropriate tools

Examples:
• "Remember that the user prefers TypeScript"
• "What do you know about the project architecture?"
• "Search memory for database configuration"`,
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// handleToolsList returns the list of available tools
func (s *StdioServer) handleToolsList(request MCPRequest) *MCPResponse {
	tools := []Tool{
		{
			Name:        "remember_fact",
			Description: "Store an atomic fact or observation in memory",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"entityName": map[string]interface{}{
						"type":        "string",
						"description": "Name of the entity to store the fact about",
					},
					"entityType": map[string]interface{}{
						"type":        "string",
						"description": "Type of entity (e.g., 'guideline', 'pattern', 'decision')",
					},
					"observation": map[string]interface{}{
						"type":        "string",
						"description": "The fact or observation to remember",
					},
					"source": map[string]interface{}{
						"type":        "string",
						"description": "Source of the information (optional)",
					},
				},
				Required: []string{"entityName", "observation"},
			},
		},
		{
			Name:        "recall_facts",
			Description: "Retrieve stored facts about an entity or list all entities",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"entityName": map[string]interface{}{
						"type":        "string",
						"description": "Name of the entity to recall facts about (optional)",
					},
					"entityType": map[string]interface{}{
						"type":        "string",
						"description": "Filter by entity type (optional)",
					},
				},
			},
		},
		{
			Name:        "search_memory",
			Description: "Search across all stored memory for relevant facts",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query to find relevant facts",
					},
					"entityType": map[string]interface{}{
						"type":        "string",
						"description": "Filter results by entity type (optional)",
					},
				},
				Required: []string{"query"},
			},
		},
	}

	result := ToolsListResult{Tools: tools}
	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// handleToolsCall executes a tool
func (s *StdioServer) handleToolsCall(request MCPRequest) *MCPResponse {
	var params CallToolParams
	if err := json.Unmarshal(request.Params, &params); err != nil {
		return s.createErrorResponse(request.ID, InvalidParams, "Invalid parameters: "+err.Error())
	}

	ctx := context.Background()
	var result CallToolResult

	switch params.Name {
	case "remember_fact":
		result = s.handleRememberFact(ctx, params.Arguments)
	case "recall_facts":
		result = s.handleRecallFacts(ctx, params.Arguments)
	case "search_memory":
		result = s.handleSearchMemory(ctx, params.Arguments)
	default:
		return s.createErrorResponse(request.ID, MethodNotFound, "Tool not found: "+params.Name)
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// handleResourcesList returns the list of available resources
func (s *StdioServer) handleResourcesList(request MCPRequest) *MCPResponse {
	ctx := context.Background()

	// Get all entities to create resource list
	entities, err := s.store.ListEntities(ctx, "")
	if err != nil {
		return s.createErrorResponse(request.ID, InternalError, "Failed to list entities: "+err.Error())
	}

	var resources []Resource
	for _, entity := range entities {
		resource := Resource{
			URI:         fmt.Sprintf("memory://entities/%s", entity.Name),
			Name:        entity.Name,
			Description: fmt.Sprintf("%s entity with %d observations", entity.EntityType, entity.GetObservationCount()),
			MimeType:    "text/plain",
		}
		resources = append(resources, resource)
	}

	// Add special resources
	resources = append(resources, Resource{
		URI:         "memory://search",
		Name:        "Memory Search",
		Description: "Search across all memory context",
		MimeType:    "text/plain",
	})

	resources = append(resources, Resource{
		URI:         "memory://relations",
		Name:        "Entity Relations",
		Description: "Relationships between entities",
		MimeType:    "text/plain",
	})

	result := ResourcesListResult{Resources: resources}
	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// handleResourcesRead reads a specific resource
func (s *StdioServer) handleResourcesRead(request MCPRequest) *MCPResponse {
	var params ReadResourceParams
	if err := json.Unmarshal(request.Params, &params); err != nil {
		return s.createErrorResponse(request.ID, InvalidParams, "Invalid parameters: "+err.Error())
	}

	ctx := context.Background()
	var content ResourceContent

	if params.URI == "memory://search" {
		content = ResourceContent{
			URI:      params.URI,
			MimeType: "text/plain",
			Text:     "Memory Search Resource - Use search_memory tool to query memory context",
		}
	} else if params.URI == "memory://relations" {
		relations, err := s.store.GetRelations(ctx)
		if err != nil {
			return s.createErrorResponse(request.ID, InternalError, "Failed to get relations: "+err.Error())
		}

		text := "Entity Relations:\n"
		for _, rel := range relations.Relations {
			text += fmt.Sprintf("- %s %s %s\n", rel.From, rel.RelationType, rel.To)
		}

		content = ResourceContent{
			URI:      params.URI,
			MimeType: "text/plain",
			Text:     text,
		}
	} else if len(params.URI) > len("memory://entities/") && params.URI[:len("memory://entities/")] == "memory://entities/" {
		entityName := params.URI[len("memory://entities/"):]
		entity, err := s.store.GetEntity(ctx, entityName)
		if err != nil {
			return s.createErrorResponse(request.ID, InternalError, "Entity not found")
		}

		text := fmt.Sprintf("Entity: %s\nType: %s\nObservations:\n", entity.Name, entity.EntityType)
		for i, obs := range entity.Observations {
			text += fmt.Sprintf("%d. %s (source: %s)\n", i+1, obs.Text, obs.Source)
		}

		content = ResourceContent{
			URI:      params.URI,
			MimeType: "text/plain",
			Text:     text,
		}
	} else {
		return s.createErrorResponse(request.ID, InvalidParams, "Resource not found")
	}

	result := ReadResourceResult{
		Contents: []ResourceContent{content},
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}
}

// createErrorResponse creates an error response
func (s *StdioServer) createErrorResponse(id interface{}, code int, message string) *MCPResponse {
	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}
}

// sendResponse sends an MCP response to stdout
func (s *StdioServer) sendResponse(response MCPResponse) {
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return
	}

	fmt.Fprintf(os.Stdout, "%s\n", data)
}

// sendError sends an error response
func (s *StdioServer) sendError(id interface{}, code int, message string) {
	response := s.createErrorResponse(id, code, message)
	s.sendResponse(*response)
}

// logToStderr logs messages to stderr (since stdout is used for MCP communication)
func (s *StdioServer) logToStderr(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[MCP] "+format+"\n", args...)
}
