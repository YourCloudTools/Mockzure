package specs

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	yaml "gopkg.in/yaml.v3"
)

// Loader loads and parses API specifications
type Loader struct {
	specsDir string
}

// NewLoader creates a new spec loader
func NewLoader(specsDir string) *Loader {
	return &Loader{
		specsDir: specsDir,
	}
}

// LoadAll loads all specifications from the specs directory
func (l *Loader) LoadAll(registry *Registry) error {
	var loadedCount int
	var skippedCount int

	// Load ARM specs
	armLoaded, armSkipped, err := l.loadARMSpecs(registry)
	if err != nil {
		return fmt.Errorf("failed to load ARM specs: %w", err)
	}
	loadedCount += armLoaded
	skippedCount += armSkipped
	log.Printf("Loaded %d ARM spec(s), skipped %d placeholder(s)", armLoaded, armSkipped)

	// Load Graph specs
	graphLoaded, graphSkipped, err := l.loadGraphSpecs(registry)
	if err != nil {
		return fmt.Errorf("failed to load Graph specs: %w", err)
	}
	loadedCount += graphLoaded
	skippedCount += graphSkipped
	log.Printf("Loaded %d Graph spec(s), skipped %d placeholder(s)", graphLoaded, graphSkipped)

	// Load Identity specs
	identityLoaded, identitySkipped, err := l.loadIdentitySpecs(registry)
	if err != nil {
		return fmt.Errorf("failed to load Identity specs: %w", err)
	}
	loadedCount += identityLoaded
	skippedCount += identitySkipped
	log.Printf("Loaded %d Identity spec(s), skipped %d placeholder(s)", identityLoaded, identitySkipped)

	log.Printf("Total: Loaded %d spec(s), skipped %d placeholder(s)", loadedCount, skippedCount)
	return nil
}

// loadARMSpecs loads ARM API specifications
// Returns: (loadedCount, skippedCount, error)
func (l *Loader) loadARMSpecs(registry *Registry) (int, int, error) {
	armDir := filepath.Join(l.specsDir, "arm")
	files, err := os.ReadDir(armDir)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read ARM directory: %w", err)
	}

	loadedCount := 0
	skippedCount := 0

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(armDir, file.Name())
		spec, err := l.loadSpecFile(filePath, APITypeARM)
		if err != nil {
			// Skip placeholder files (like "404: Not Found")
			if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "placeholder") {
				log.Printf("Skipping placeholder file: %s", file.Name())
				skippedCount++
				continue
			}
			return loadedCount, skippedCount, fmt.Errorf("failed to load %s: %w", filePath, err)
		}
		if spec != nil {
			spec.Name = strings.TrimSuffix(file.Name(), ".json")
			registry.Register(spec)
			log.Printf("Loaded ARM spec: %s (%s)", spec.Name, getSpecFormat(spec))
			loadedCount++
		}
	}

	return loadedCount, skippedCount, nil
}

// loadGraphSpecs loads Graph API specifications
// Returns: (loadedCount, skippedCount, error)
func (l *Loader) loadGraphSpecs(registry *Registry) (int, int, error) {
	graphDir := filepath.Join(l.specsDir, "graph")
	files, err := os.ReadDir(graphDir)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read Graph directory: %w", err)
	}

	loadedCount := 0
	skippedCount := 0

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
			continue
		}

		filePath := filepath.Join(graphDir, file.Name())
		spec, err := l.loadSpecFile(filePath, APITypeGraph)
		if err != nil {
			// Skip placeholder files
			if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "placeholder") {
				log.Printf("Skipping placeholder file: %s", file.Name())
				skippedCount++
				continue
			}
			return loadedCount, skippedCount, fmt.Errorf("failed to load %s: %w", filePath, err)
		}
		if spec != nil {
			spec.Name = strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			registry.Register(spec)
			log.Printf("Loaded Graph spec: %s (%s)", spec.Name, getSpecFormat(spec))
			loadedCount++
		}
	}

	return loadedCount, skippedCount, nil
}

// loadIdentitySpecs loads Identity/OIDC specifications
// Returns: (loadedCount, skippedCount, error)
func (l *Loader) loadIdentitySpecs(registry *Registry) (int, int, error) {
	identityDir := filepath.Join(l.specsDir, "identity")
	files, err := os.ReadDir(identityDir)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read Identity directory: %w", err)
	}

	loadedCount := 0
	skippedCount := 0

	for _, file := range files {
		filePath := filepath.Join(identityDir, file.Name())
		
		// Skip non-spec files (like oidc-configuration.json, oidc-jwks.json)
		if file.Name() == "oidc-configuration.json" || file.Name() == "oidc-jwks.json" {
			skippedCount++
			continue
		}

		if strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml") {
			spec, err := l.loadSpecFile(filePath, APITypeIdentity)
			if err != nil {
				// Skip placeholder files
				if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "placeholder") {
					log.Printf("Skipping placeholder file: %s", file.Name())
					skippedCount++
					continue
				}
				return loadedCount, skippedCount, fmt.Errorf("failed to load %s: %w", filePath, err)
			}
			if spec != nil {
				spec.Name = strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
				registry.Register(spec)
				log.Printf("Loaded Identity spec: %s (%s)", spec.Name, getSpecFormat(spec))
				loadedCount++
			}
		}
	}

	return loadedCount, skippedCount, nil
}

// loadSpecFile loads a single spec file and determines its format
func (l *Loader) loadSpecFile(filePath string, apiType APIType) (*Spec, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Check for placeholder files
	content := strings.TrimSpace(string(data))
	if content == "404: Not Found" || strings.HasPrefix(content, "404") {
		return nil, fmt.Errorf("placeholder file: %s", filePath)
	}

	// Try to detect format
	ext := strings.ToLower(filepath.Ext(filePath))
	
	// Try OpenAPI 3.0 first (YAML or JSON)
	if ext == ".yaml" || ext == ".yml" {
		return l.loadOpenAPI3(data, filePath, apiType)
	}

	// Try Swagger 2.0 (JSON)
	if ext == ".json" {
		// First try as Swagger 2.0
		if spec, err := l.loadSwagger2(data, filePath, apiType); err == nil {
			return spec, nil
		}
		// Fallback to OpenAPI 3.0
		return l.loadOpenAPI3(data, filePath, apiType)
	}

	return nil, fmt.Errorf("unsupported file format: %s", ext)
}

// loadOpenAPI3 loads an OpenAPI 3.0 specification
func (l *Loader) loadOpenAPI3(data []byte, filePath string, apiType APIType) (*Spec, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	var doc *openapi3.T
	var err error

	// Check if it's YAML
	if strings.HasSuffix(strings.ToLower(filePath), ".yaml") || strings.HasSuffix(strings.ToLower(filePath), ".yml") {
		// Parse YAML first, then convert to JSON for kin-openapi
		var yamlDoc interface{}
		if err := yaml.Unmarshal(data, &yamlDoc); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
		
		// Convert to JSON
		jsonData, err := json.Marshal(yamlDoc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert YAML to JSON: %w", err)
		}
		
		doc, err = loader.LoadFromData(jsonData)
	} else {
		// Try JSON directly
		doc, err = loader.LoadFromData(data)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI 3.0: %w", err)
	}

	return &Spec{
		Type:     apiType,
		OpenAPI3: doc,
		Path:     filePath,
	}, nil
}

// loadSwagger2 loads a Swagger 2.0 specification
func (l *Loader) loadSwagger2(data []byte, filePath string, apiType APIType) (*Spec, error) {
	// Parse JSON
	var swagger spec.Swagger
	if err := json.Unmarshal(data, &swagger); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate it's actually Swagger 2.0
	if swagger.Swagger == "" || !strings.HasPrefix(swagger.Swagger, "2.") {
		return nil, fmt.Errorf("not a Swagger 2.0 spec")
	}

	// Use go-openapi/loads for full validation
	doc, err := loads.Analyzed(json.RawMessage(data), "")
	if err != nil {
		return nil, fmt.Errorf("failed to analyze Swagger spec: %w", err)
	}

	return &Spec{
		Type:     apiType,
		Swagger2: doc.Spec(),
		Path:     filePath,
	}, nil
}

// getSpecFormat returns a string describing the spec format
func getSpecFormat(spec *Spec) string {
	if spec.IsOpenAPI3() {
		return "OpenAPI 3.0"
	}
	if spec.IsSwagger2() {
		return "Swagger 2.0"
	}
	return "Unknown"
}

