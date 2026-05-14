package core

import "fmt"

// Registry holds the registered verifiers for different Chaos Kinds.
type Registry struct {
	verifiers map[string]Verifier
}

// NewRegistry creates a new empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		verifiers: make(map[string]Verifier),
	}
}

// Register adds a Verifier to the registry.
func (r *Registry) Register(v Verifier) {
	r.verifiers[v.Kind()] = v
}

// Get retrieves a Verifier by Chaos Kind.
func (r *Registry) Get(kind string) (Verifier, error) {
	if v, exists := r.verifiers[kind]; exists {
		return v, nil
	}
	return nil, fmt.Errorf("no verifier registered for Chaos Kind: %s", kind)
}
