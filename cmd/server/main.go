package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/tr4d3r/ghcp-memory-context/internal/api"
	"github.com/tr4d3r/ghcp-memory-context/internal/mcp"
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
	// Parse command line flags
	var mcpStdio bool
	var port string
	var dataDir string
	var showVersion bool
	var showHelp bool

	flag.BoolVar(&mcpStdio, "mcp-stdio", false, "Run in MCP stdio mode for integration with MCP clients")
	flag.StringVar(&port, "port", "", "Server port (default: 8080, env: PORT)")
	flag.StringVar(&dataDir, "data-dir", "", "Data storage directory (default: ./.memory-context, env: DATA_DIR)")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showHelp, "help", false, "Show help information")
	flag.Parse()

	if showVersion {
		log.Println("GHCP Memory Context Server v1.0.0")
		return
	}

	if showHelp {
		log.Println("GHCP Memory Context Server - Persistent memory for AI assistants")
		log.Println("")
		log.Println("Usage:")
		log.Println("  ghcp-memory-context [options]")
		log.Println("")
		log.Println("Options:")
		flag.PrintDefaults()
		log.Println("")
		log.Println("Examples:")
		log.Println("  ghcp-memory-context                    # Start HTTP server")
		log.Println("  ghcp-memory-context --mcp-stdio        # Start MCP stdio server")
		log.Println("  ghcp-memory-context --port 3000        # Custom port")
		log.Println("  ghcp-memory-context --data-dir /path   # Custom data directory")
		return
	}

	// Get configuration from environment variables if not set via flags
	if port == "" {
		port = os.Getenv("PORT")
	}
	if dataDir == "" {
		dataDir = os.Getenv("DATA_DIR")
	}

	// Handle legacy positional argument for data directory
	if dataDir == "" && len(flag.Args()) > 0 {
		dataDir = flag.Args()[0]
	}

	// Ensure data directory is absolute
	if dataDir != "" {
		absPath, err := filepath.Abs(dataDir)
		if err != nil {
			log.Fatalf("Invalid data directory path: %v", err)
		}
		dataDir = absPath
	} else {
		dataDir = "./.memory-context"
	}

	// Initialize storage
	store := filestore.NewFileStore(dataDir)
	if err := store.Initialize(); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	if mcpStdio {
		// Run MCP stdio server
		log.Printf("Starting MCP stdio server (data directory: %s)", dataDir)
		mcpServer := mcp.NewStdioServer(store)
		if err := mcpServer.Run(); err != nil {
			log.Fatalf("MCP stdio server error: %v", err)
		}
	} else {
		// Run HTTP server
		server := NewServer(port, dataDir)
		if err := server.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}
