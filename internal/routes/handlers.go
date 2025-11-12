package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/yourcloudtools/mockzure/internal/mappers"
	"github.com/yourcloudtools/mockzure/internal/specs"
)

// handleRequest is the generic request handler that routes to appropriate mappers
func handleRequest(w http.ResponseWriter, r *http.Request, pathParams map[string]string, operationID, pathPattern, method string, apiType specs.APIType, store interface{}) {
	// Extract query parameters
	queryParams := make(map[string]string)
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			queryParams[k] = v[0]
		}
	}

	// Merge path and query parameters
	allParams := make(map[string]string)
	for k, v := range pathParams {
		allParams[k] = v
	}
	for k, v := range queryParams {
		allParams[k] = v
	}

	// Route to appropriate mapper based on API type
	switch apiType {
	case specs.APITypeARM:
		handleARMRequest(w, r, allParams, operationID, pathPattern, method, store)
	case specs.APITypeGraph:
		handleGraphRequest(w, r, allParams, operationID, pathPattern, method, store)
	case specs.APITypeIdentity:
		handleIdentityRequest(w, r, allParams, operationID, pathPattern, method, store)
	default:
		http.Error(w, fmt.Sprintf("Unknown API type: %s", apiType), http.StatusInternalServerError)
	}
}

// handleARMRequest handles ARM API requests
func handleARMRequest(w http.ResponseWriter, r *http.Request, params map[string]string, operationID, pathPattern, method string, store interface{}) {
	// Type assert store to access Store methods
	storeTyped, ok := store.(mappers.StoreInterface)
	if !ok {
		http.Error(w, "Invalid store type", http.StatusInternalServerError)
		return
	}

	// Check if this is an operation status check (LRO pattern)
	if strings.Contains(pathPattern, "/operations/") && method == "GET" {
		response, err := mappers.MapARMOperationStatus(operationID, params)
		if err != nil {
			log.Printf("Error mapping ARM operation status: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode JSON response: %v", err)
		}
		return
	}

	// Use ARM mapper to generate response
	response, err := mappers.MapARMResponse(operationID, pathPattern, method, params, storeTyped)
	if err != nil {
		log.Printf("Error mapping ARM response: %v", err)
		// Return spec-compliant error response
		errorResponse := map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "ResourceNotFound",
				"message": err.Error(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(err.Error(), "not found") {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
			log.Printf("Failed to encode error response: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

// handleGraphRequest handles Microsoft Graph API requests
func handleGraphRequest(w http.ResponseWriter, r *http.Request, params map[string]string, operationID, pathPattern, method string, store interface{}) {
	// Type assert store to access Store methods
	storeTyped, ok := store.(mappers.StoreInterface)
	if !ok {
		http.Error(w, "Invalid store type", http.StatusInternalServerError)
		return
	}

	// Use Graph mapper to generate response
	response, err := mappers.MapGraphResponse(operationID, pathPattern, method, params, storeTyped)
	if err != nil {
		log.Printf("Error mapping Graph response: %v", err)
		// Return Graph API-compliant error response
		errorResponse := map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "ItemNotFound",
				"message": err.Error(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(err.Error(), "not found") {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
			log.Printf("Failed to encode error response: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

// handleIdentityRequest handles Identity/OIDC API requests
func handleIdentityRequest(w http.ResponseWriter, r *http.Request, params map[string]string, operationID, pathPattern, method string, store interface{}) {
	// Identity endpoints are handled separately in main.go (OIDC discovery, token, etc.)
	// This is a fallback for any spec-defined identity endpoints
	http.Error(w, "Identity endpoint not implemented", http.StatusNotImplemented)
}

// RegisterRoutes registers generated routes with an HTTP mux
// Uses a single catch-all handler that matches routes dynamically to handle overlapping paths
func RegisterRoutes(mux *http.ServeMux, routes []Route) {
	registeredCount := 0
	byMethod := make(map[string]int)

	// Group routes by their base path prefix to optimize matching
	// For routes with parameters, we need a catch-all handler
	routeMap := make(map[string][]Route)
	var exactRoutes []Route

	for _, route := range routes {
		if strings.Contains(route.Path, "{") {
			// Parameterized route - group by base path
			basePath := extractBasePath(route.Path)
			if basePath == "" {
				basePath = "/"
			}
			routeMap[basePath] = append(routeMap[basePath], route)
		} else {
			// Exact path route
			exactRoutes = append(exactRoutes, route)
		}
		registeredCount++
		byMethod[route.Method]++
	}

	// Register exact path routes first (they take precedence)
	for _, route := range exactRoutes {
		handler := createRouteHandler(route)
		mux.HandleFunc(route.Path, handler)
		if !strings.HasSuffix(route.Path, "/") {
			mux.HandleFunc(route.Path+"/", handler)
		}
	}

	// Register parameterized routes with catch-all handlers
	for basePath, routeGroup := range routeMap {
		// Create a handler that checks all routes in this group
		handler := func(routes []Route) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// Try to match against all routes in this group
				for _, route := range routes {
					if r.Method != route.Method {
						continue
					}

					matched, params := MatchPath(route.Path, r.URL.Path)
					if matched {
						// Found a match - call the handler
						route.Handler(w, r, params)
						return
					}
				}

				// No match found
				http.NotFound(w, r)
			}
		}(routeGroup)

		// Register with the base path
		// In http.ServeMux, patterns ending with '/' match all paths with that prefix
		if basePath == "/" {
			// Root path - register as catch-all (but only if no exact route registered)
			// Check if we already registered exact "/" route
			hasExactRoot := false
			for _, route := range exactRoutes {
				if route.Path == "/" {
					hasExactRoot = true
					break
				}
			}
			if !hasExactRoot {
				mux.HandleFunc("/", handler)
			}
		} else {
			// Ensure base path ends with '/' for prefix matching
			prefixPath := basePath
			if !strings.HasSuffix(prefixPath, "/") {
				prefixPath = prefixPath + "/"
			}
			mux.HandleFunc(prefixPath, handler)
		}
	}

	log.Printf("Registered %d route(s) with HTTP mux", registeredCount)
	for method, count := range byMethod {
		log.Printf("  - %s: %d route(s)", method, count)
	}
}

// createRouteHandler creates a handler function for a single route
func createRouteHandler(route Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only handle the correct HTTP method
		if r.Method != route.Method {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Match the path and extract parameters
		matched, params := MatchPath(route.Path, r.URL.Path)
		if !matched {
			http.NotFound(w, r)
			return
		}

		// Call the route handler
		route.Handler(w, r, params)
	}
}

// extractBasePath extracts the base path before the first parameter
func extractBasePath(pattern string) string {
	idx := strings.Index(pattern, "{")
	if idx == -1 {
		return pattern
	}
	return pattern[:idx]
}
