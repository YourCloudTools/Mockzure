package specs

import (
	"fmt"
	"sync"
)

// Registry stores loaded API specifications
type Registry struct {
	mu    sync.RWMutex
	specs map[APIType][]*Spec
}

// NewRegistry creates a new spec registry
func NewRegistry() *Registry {
	return &Registry{
		specs: make(map[APIType][]*Spec),
	}
}

// Register adds a spec to the registry
func (r *Registry) Register(spec *Spec) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.specs[spec.Type] = append(r.specs[spec.Type], spec)
}

// Get returns all specs of a given type
func (r *Registry) Get(apiType APIType) []*Spec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.specs[apiType]
}

// GetAll returns all registered specs
func (r *Registry) GetAll() map[APIType][]*Spec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[APIType][]*Spec)
	for k, v := range r.specs {
		result[k] = v
	}
	return result
}

// FindByPath finds a spec by its file path
func (r *Registry) FindByPath(path string) (*Spec, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, specs := range r.specs {
		for _, s := range specs {
			if s.Path == path {
				return s, nil
			}
		}
	}
	return nil, fmt.Errorf("spec not found: %s", path)
}

