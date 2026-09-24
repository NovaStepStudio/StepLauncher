package modloader

import (
	"fmt"
	"net/http"
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

// SetHTTPClientAll propaga el cliente HTTP a todos los providers
// registrados (cambios de proxy/límite en caliente).
func (r *Registry) SetHTTPClientAll(client *http.Client) {
	if client == nil {
		return
	}
	for _, p := range r.providers {
		p.SetHTTPClient(client)
	}
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
