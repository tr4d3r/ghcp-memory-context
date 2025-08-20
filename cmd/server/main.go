package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/tr4d3r/ghcp-memory-context/internal/api"
	"github.com/tr4d3r/ghcp-memory-context/internal/storage/filestore"
)

// Server represents the MCP memory context server
type Server struct {
	port      string
	store     *filestore.FileStore
	apiRouter *api.Router
}

// NewServer creates a new server instance
func NewServer(port string, dataDir string) *Server {
	if port == "" {
		port = "8080"
	}

	if dataDir == "" {
		dataDir = "./data"
	}

	// Initialize file store
	store := filestore.NewFileStore(dataDir)
	if err := store.Initialize(); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Create API router with the store
	apiRouter := api.NewRouter(store)

	return &Server{
		port:      port,
		store:     store,
		apiRouter: apiRouter,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Setup routes
	handler := s.apiRouter.SetupRoutes()

	server := &http.Server{
		Addr:         ":" + s.port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Starting memory context server on port %s", s.port)
	log.Printf("Data directory initialized")
	log.Printf("API endpoints available:")
	log.Printf("  - GET  /health")
	log.Printf("  - GET  /entities")
	log.Printf("  - POST /entities")
	log.Printf("  - GET  /entities/{name}")
	log.Printf("  - POST /memory/remember")
	log.Printf("  - GET  /memory/recall")
	log.Printf("  - GET  /memory/search")
	log.Printf("  - GET  /relations")
	log.Printf("  - GET  /mcp/resources")
	log.Printf("  - POST /mcp/tools/remember_fact")
	log.Printf("  - POST /mcp/tools/recall_facts")
	log.Printf("  - POST /mcp/tools/search_memory")

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	log.Println("Server exited")
	return nil
}

func main() {
	// Get configuration from environment variables
	port := os.Getenv("PORT")
	dataDir := os.Getenv("DATA_DIR")

	// Allow custom data directory via command line
	if len(os.Args) > 1 {
		dataDir = os.Args[1]
	}

	// Ensure data directory is absolute
	if dataDir != "" {
		absPath, err := filepath.Abs(dataDir)
		if err != nil {
			log.Fatalf("Invalid data directory path: %v", err)
		}
		dataDir = absPath
	}

	// Create and start server
	server := NewServer(port, dataDir)
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
