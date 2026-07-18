package app

import (
	"fmt"
	"net/http"
	"strings"
)

// Module is a bounded context wired into the monolith. Implementations live
// under internal/modules/<name> and are registered in the composition root.
type Module interface {
	// Name is the module's stable identifier (e.g. "policy").
	Name() string
	// Init receives the platform kernel and any cross-module ports it depends
	// on (injected via the module's constructor before registration). It runs
	// once at startup; returning an error aborts boot (fail closed).
	Init(k *Kernel) error
	// Mount returns the module's URL prefix and HTTP handler. The prefix is
	// trimmed before the handler sees the request.
	Mount() (prefix string, handler http.Handler)
}

// Registry holds the ordered set of modules composing the application.
type Registry struct {
	modules []Module
}

// NewRegistry builds a registry from modules in dependency order.
func NewRegistry(modules ...Module) *Registry {
	return &Registry{modules: modules}
}

// InitAll initialises every module, aborting on the first failure.
func (r *Registry) InitAll(k *Kernel) error {
	seen := make(map[string]bool, len(r.modules))
	for _, m := range r.modules {
		name := m.Name()
		if seen[name] {
			return fmt.Errorf("duplicate module registered: %q", name)
		}
		seen[name] = true

		if err := m.Init(k); err != nil {
			return fmt.Errorf("init module %q: %w", name, err)
		}
		k.Logger.Info("module initialised", "module", name)
	}
	return nil
}

// MountAll registers every module's routes on the mux under its prefix.
func (r *Registry) MountAll(mux *http.ServeMux) {
	for _, m := range r.modules {
		prefix, handler := m.Mount()
		if prefix == "" || handler == nil {
			continue
		}
		prefix = "/" + strings.Trim(prefix, "/") + "/"
		mux.Handle(prefix, http.StripPrefix(strings.TrimSuffix(prefix, "/"), handler))
	}
}

// Names returns the registered module names, in order.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.modules))
	for _, m := range r.modules {
		names = append(names, m.Name())
	}
	return names
}
