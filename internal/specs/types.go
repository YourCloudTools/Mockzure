package specs

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-openapi/spec"
)

// APIType represents the type of API (ARM, Graph, Identity)
type APIType string

const (
	APITypeARM      APIType = "arm"
	APITypeGraph    APIType = "graph"
	APITypeIdentity APIType = "identity"
)

// Spec represents a loaded API specification
type Spec struct {
	Type        APIType
	OpenAPI3    *openapi3.T
	Swagger2    *spec.Swagger
	Path        string
	Name        string
}

// IsOpenAPI3 returns true if this is an OpenAPI 3.0 spec
func (s *Spec) IsOpenAPI3() bool {
	return s.OpenAPI3 != nil
}

// IsSwagger2 returns true if this is a Swagger 2.0 spec
func (s *Spec) IsSwagger2() bool {
	return s.Swagger2 != nil
}

