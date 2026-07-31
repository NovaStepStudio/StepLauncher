package modloader

import (
	"fmt"
	"sort"
)

type Registry struct {
	providers map[string]ModLoaderProvider
}

func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]ModLoaderProvider)}
}

func (r *Registry) Register(p ModLoaderProvider) {
	r.providers[p.Name()] = p
}

func (r *Registry) Get(name string) (ModLoaderProvider, error) {
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("modloader %q not found", name)
	}
	return p, nil
}

func (r *Registry) List() []string {
	names := make([]string, 0, len(r.providers))
	for n := range r.providers {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
