package routes

import (
	"net/http"
)

// ValidationMiddleware validates requests against spec schemas
func ValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Add schema validation
		// For now, just pass through
		// In the future, validate:
		// - Request body against request schema
		// - Query parameters against parameter definitions
		// - Required parameters are present
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware enforces authentication requirements
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authentication is handled in main.go's existing logic
		// This middleware can be extended for spec-based auth validation
		next.ServeHTTP(w, r)
	})
}

