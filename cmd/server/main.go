package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tr4d3r/ghcp-memory-context/internal/models"
)

// Server represents the MCP memory context server
type Server struct {
	port string
	mux  *http.ServeMux
}

// NewServer creates a new server instance
func NewServer(port string) *Server {
	if port == "" {
		port = "8080"
	}

	s := &Server{
		port: port,
		mux:  http.NewServeMux(),
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.mux.HandleFunc("/health", s.handleHealth)

	// API endpoints
	s.mux.HandleFunc("/api/v1/tasks", s.handleTasks)
	s.mux.HandleFunc("/api/v1/tasks/", s.handleTaskByID)

	// Root endpoint with basic info
	s.mux.HandleFunc("/", s.handleRoot)
}

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"service":   "ghcp-memory-context",
	}

	json.NewEncoder(w).Encode(response)
}

// handleRoot returns basic server information
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"name":        "GHCP Memory Context Server",
		"description": "MCP-compliant memory server for GitHub Copilot Premium integration",
		"version":     "1.0.0",
		"endpoints": []string{
			"/health",
			"/api/v1/tasks",
		},
	}

	json.NewEncoder(w).Encode(response)
}

// handleTasks handles task collection operations
func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		s.handleGetTasks(w, r)
	case http.MethodPost:
		s.handleCreateTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleTaskByID handles individual task operations
func (s *Server) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract task ID from path
	taskID := r.URL.Path[len("/api/v1/tasks/"):]
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetTask(w, r, taskID)
	case http.MethodPut:
		s.handleUpdateTask(w, r, taskID)
	case http.MethodDelete:
		s.handleDeleteTask(w, r, taskID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetTasks returns a list of tasks (placeholder implementation)
func (s *Server) handleGetTasks(w http.ResponseWriter, r *http.Request) {
	// Create a sample task for demonstration
	sampleTask := models.NewTask("Sample Task", "This is a sample task for the MCP memory server")
	sampleTask.ID = "550e8400-e29b-41d4-a716-446655440000"
	sampleTask.Timestamp = time.Now().Unix()

	tasks := []*models.Task{sampleTask}

	response := map[string]interface{}{
		"tasks": tasks,
		"count": len(tasks),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleCreateTask creates a new task (placeholder implementation)
func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var taskData struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&taskData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	task := models.NewTask(taskData.Title, taskData.Description)
	task.ID = fmt.Sprintf("task-%d", time.Now().Unix())
	task.Timestamp = time.Now().Unix()

	if err := task.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// handleGetTask retrieves a specific task (placeholder implementation)
func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request, taskID string) {
	// In a real implementation, this would query the database
	task := models.NewTask("Retrieved Task", "This task was retrieved by ID")
	task.ID = taskID
	task.Timestamp = time.Now().Unix()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

// handleUpdateTask updates a specific task (placeholder implementation)
func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request, taskID string) {
	var updates map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// In a real implementation, this would update the database
	response := map[string]interface{}{
		"id":      taskID,
		"message": "Task updated successfully",
		"updates": updates,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleDeleteTask deletes a specific task (placeholder implementation)
func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request, taskID string) {
	// In a real implementation, this would delete from the database
	response := map[string]interface{}{
		"id":      taskID,
		"message": "Task deleted successfully",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Run starts the HTTP server with graceful shutdown
func (s *Server) Run() error {
	server := &http.Server{
		Addr:    ":" + s.port,
		Handler: s.mux,
	}

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Starting GHCP Memory Context Server on port %s", s.port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
		return err
	}

	log.Println("Server stopped")
	return nil
}

func main() {
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create and run server
	server := NewServer(port)
	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
