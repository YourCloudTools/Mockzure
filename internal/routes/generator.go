package routes

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/yourcloudtools/mockzure/internal/specs"
)

// GenerateRoutes generates HTTP routes from loaded specifications
func (rg *RouteGenerator) GenerateRoutes(registry *specs.Registry) ([]Route, error) {
	var routes []Route

	allSpecs := registry.GetAll()
	for apiType, specList := range allSpecs {
		for _, spec := range specList {
			var specRoutes []Route
			var err error

			if spec.IsOpenAPI3() {
				specRoutes, err = rg.generateFromOpenAPI3(spec, apiType)
				if err != nil {
					return nil, fmt.Errorf("failed to generate routes from %s: %w", spec.Path, err)
				}
			} else if spec.IsSwagger2() {
				specRoutes, err = rg.generateFromSwagger2(spec, apiType)
				if err != nil {
					return nil, fmt.Errorf("failed to generate routes from %s: %w", spec.Path, err)
				}
			}

			if len(specRoutes) > 0 {
				log.Printf("Generated %d route(s) from spec '%s' (%s)", len(specRoutes), spec.Name, apiType)
				// Log first few routes as examples
				for i, route := range specRoutes {
					if i < 3 {
						log.Printf("  - %s %s (operation: %s)", route.Method, route.Path, route.OperationID)
					} else if i == 3 {
						log.Printf("  ... and %d more route(s)", len(specRoutes)-3)
						break
					}
				}
			}
			routes = append(routes, specRoutes...)
		}
	}

	return routes, nil
}

// generateFromOpenAPI3 generates routes from an OpenAPI 3.0 specification
func (rg *RouteGenerator) generateFromOpenAPI3(spec *specs.Spec, apiType specs.APIType) ([]Route, error) {
	var routes []Route

	if spec.OpenAPI3 == nil || spec.OpenAPI3.Paths == nil {
		return routes, nil
	}

	// Use Map() to get the paths map
	pathsMap := spec.OpenAPI3.Paths.Map()
	for pathPattern, pathItem := range pathsMap {
		if pathItem == nil {
			continue
		}

		// Handle each HTTP method
		operations := map[string]*openapi3.Operation{
			http.MethodGet:    pathItem.Get,
			http.MethodPost:   pathItem.Post,
			http.MethodPut:    pathItem.Put,
			http.MethodDelete: pathItem.Delete,
			http.MethodPatch:  pathItem.Patch,
		}

		for method, operation := range operations {
			if operation == nil {
				continue
			}

			handler := rg.createHandler(operation.OperationID, pathPattern, method, apiType, spec.OpenAPI3)
			routes = append(routes, Route{
				Method:      method,
				Path:        pathPattern,
				Handler:     handler,
				OperationID: operation.OperationID,
				Tags:        operation.Tags,
			})
		}
	}

	return routes, nil
}

// generateFromSwagger2 generates routes from a Swagger 2.0 specification
func (rg *RouteGenerator) generateFromSwagger2(spec *specs.Spec, apiType specs.APIType) ([]Route, error) {
	var routes []Route

	if spec.Swagger2 == nil || spec.Swagger2.Paths == nil {
		return routes, nil
	}

	for pathPattern, pathItem := range spec.Swagger2.Paths.Paths {
		// Handle each HTTP method
		// Note: pathItem.Get, Post, etc. return *spec.Operation
		operations := map[string]interface{}{
			http.MethodGet:    pathItem.Get,
			http.MethodPost:   pathItem.Post,
			http.MethodPut:    pathItem.Put,
			http.MethodDelete: pathItem.Delete,
			http.MethodPatch:  pathItem.Patch,
		}

		for method, op := range operations {
			if op == nil {
				continue
			}
			
			// Access operation fields directly using type assertion
			// We know op is *spec.Operation, so we can use an interface to access fields
			var operationID string
			var tags []string
			
			// Create an interface that matches Operation's structure
			type operationInterface struct {
				ID   string
				Tags []string
			}
			
			// Use unsafe pointer conversion or reflection - simpler: use the actual operation
			// Since we can't directly type assert, we'll get the operation from pathItem directly
			switch method {
			case http.MethodGet:
				if pathItem.Get != nil {
					operationID = pathItem.Get.ID
					tags = pathItem.Get.Tags
				}
			case http.MethodPost:
				if pathItem.Post != nil {
					operationID = pathItem.Post.ID
					tags = pathItem.Post.Tags
				}
			case http.MethodPut:
				if pathItem.Put != nil {
					operationID = pathItem.Put.ID
					tags = pathItem.Put.Tags
				}
			case http.MethodDelete:
				if pathItem.Delete != nil {
					operationID = pathItem.Delete.ID
					tags = pathItem.Delete.Tags
				}
			case http.MethodPatch:
				if pathItem.Patch != nil {
					operationID = pathItem.Patch.ID
					tags = pathItem.Patch.Tags
				}
			}
			
			if operationID == "" {
				operationID = method + "_" + strings.ReplaceAll(pathPattern, "/", "_")
			}

			handler := rg.createSwagger2Handler(operationID, pathPattern, method, apiType)
			routes = append(routes, Route{
				Method:      method,
				Path:        pathPattern,
				Handler:     handler,
				OperationID: operationID,
				Tags:        tags,
			})
		}
	}

	return routes, nil
}

// createHandler creates a handler function for an OpenAPI 3.0 operation
func (rg *RouteGenerator) createHandler(operationID, pathPattern, method string, apiType specs.APIType, doc *openapi3.T) RouteHandler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		// Delegate to the generic handler
		handleRequest(w, r, params, operationID, pathPattern, method, apiType, rg.store)
	}
}

// createSwagger2Handler creates a handler function for a Swagger 2.0 operation
func (rg *RouteGenerator) createSwagger2Handler(operationID, pathPattern, method string, apiType specs.APIType) RouteHandler {
	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		// Delegate to the generic handler
		handleRequest(w, r, params, operationID, pathPattern, method, apiType, rg.store)
	}
}

// MatchPath matches a request path against a route pattern and extracts parameters
func MatchPath(pattern, requestPath string) (bool, map[string]string) {
	params := make(map[string]string)

	// Convert OpenAPI path pattern to regex
	// e.g., /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}
	// becomes /subscriptions/([^/]+)/resourceGroups/([^/]+)
	regexPattern := pattern
	paramNames := []string{}

	// Find all parameter placeholders
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(pattern, -1)
	for _, match := range matches {
		paramName := match[1]
		paramNames = append(paramNames, paramName)
		regexPattern = strings.ReplaceAll(regexPattern, match[0], `([^/]+)`)
	}

	// Match the pattern
	regex := regexp.MustCompile("^" + regexPattern + "$")
	match := regex.FindStringSubmatch(requestPath)
	if match == nil {
		return false, nil
	}

	// Extract parameter values
	for i, paramName := range paramNames {
		if i+1 < len(match) {
			params[paramName] = match[i+1]
		}
	}

	return true, params
}

