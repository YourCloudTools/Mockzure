package routes

import (
	"net/http"
)

// RouteHandler is a function that handles an HTTP request
type RouteHandler func(w http.ResponseWriter, r *http.Request, params map[string]string)

// Route represents a generated route from a spec
type Route struct {
	Method      string
	Path        string
	Handler     RouteHandler
	OperationID string
	Tags        []string
}

// RouteGenerator generates routes from API specifications
type RouteGenerator struct {
	store interface{} // Store interface for data access
}

// NewRouteGenerator creates a new route generator
func NewRouteGenerator(store interface{}) *RouteGenerator {
	return &RouteGenerator{
		store: store,
	}
}

