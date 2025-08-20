package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// ErrorResponse represents a standardized API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

// SuccessResponse represents a standardized API success response
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Count   int         `json:"count,omitempty"`
}

// writeJSONResponse writes a JSON response with the given status code
func (r *Router) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Fallback error response if JSON encoding fails
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// writeErrorResponse writes a standardized error response
func (r *Router) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
		Code:    statusCode,
	}
	r.writeJSONResponse(w, statusCode, response)
}

// writeSuccessResponse writes a standardized success response
func (r *Router) writeSuccessResponse(w http.ResponseWriter, data interface{}, message string) {
	response := SuccessResponse{
		Data:    data,
		Message: message,
	}

	// Add count for slices
	if slice, ok := data.([]interface{}); ok {
		response.Count = len(slice)
	}

	r.writeJSONResponse(w, http.StatusOK, response)
}

// extractPathParam extracts a parameter from the URL path
// Example: /entities/project_standards -> extractPathParam(r, "/entities/") -> "project_standards"
func extractPathParam(r *http.Request, prefix string) string {
	path := r.URL.Path
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	param := strings.TrimPrefix(path, prefix)
	return strings.Trim(param, "/")
}

// parseQueryParam extracts a query parameter and returns its value
func parseQueryParam(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

// parseIntQueryParam extracts a query parameter and converts it to int
func parseIntQueryParam(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// validateJSONRequest validates that the request has JSON content type
func validateJSONRequest(r *http.Request) error {
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return fmt.Errorf("content-Type must be application/json")
	}
	return nil
}

// validateRequiredFields checks that required fields are not empty
func validateRequiredFields(fields map[string]string) error {
	for field, value := range fields {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("field '%s' is required", field)
		}
	}
	return nil
}
