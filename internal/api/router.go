package api

import (
	"net/http"

	"github.com/tr4d3r/ghcp-memory-context/internal/storage"
)

// Router handles HTTP routing for the memory context API
type Router struct {
	store storage.Storage
}

// NewRouter creates a new API router with the given storage backend
func NewRouter(store storage.Storage) *Router {
	return &Router{
		store: store,
	}
}

// SetupRoutes configures all API routes and returns the HTTP handler
func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Entity endpoints
	mux.HandleFunc("/entities", r.handleEntities)
	mux.HandleFunc("/entities/", r.handleEntityByName)

	// Memory operation endpoints
	mux.HandleFunc("/memory/remember", r.handleMemoryRemember)
	mux.HandleFunc("/memory/recall", r.handleMemoryRecall)
	mux.HandleFunc("/memory/search", r.handleMemorySearch)

	// Relation endpoints
	mux.HandleFunc("/relations", r.handleRelations)
	mux.HandleFunc("/relations/", r.handleRelationByID)

	// MCP-specific endpoints
	mux.HandleFunc("/mcp/resources", r.handleMCPResources)
	mux.HandleFunc("/mcp/resources/", r.handleMCPResourceByURI)
	mux.HandleFunc("/mcp/tools/remember_fact", r.handleMCPRememberFact)
	mux.HandleFunc("/mcp/tools/recall_facts", r.handleMCPRecallFacts)
	mux.HandleFunc("/mcp/tools/search_memory", r.handleMCPSearchMemory)
	mux.HandleFunc("/mcp/tools/forget_fact", r.handleMCPForgetFact)

	// Health check endpoint
	mux.HandleFunc("/health", r.handleHealth)

	// Add CORS and common middleware
	return r.corsMiddleware(r.loggingMiddleware(mux))
}

// corsMiddleware adds CORS headers for cross-origin requests
func (r *Router) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if req.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, req)
	})
}

// loggingMiddleware logs HTTP requests
func (r *Router) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Simple request logging - in production, use structured logging
		// fmt.Printf("[%s] %s %s\n", time.Now().Format("2006-01-02 15:04:05"), req.Method, req.URL.Path)
		next.ServeHTTP(w, req)
	})
}

// Health check endpoint
func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":  "healthy",
		"service": "memory-context-server",
		"version": "1.0.0",
	}

	r.writeJSONResponse(w, http.StatusOK, response)
}
