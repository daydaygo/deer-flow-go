package subagents

import (
	"fmt"
	"sync"

	"github.com/user/deer-flow-go/internal/subagentstypes"
)

type Registry struct {
	subagents map[string]*Subagent
	mu        sync.RWMutex
}

var globalRegistry = NewRegistry()

func NewRegistry() *Registry {
	r := &Registry{
		subagents: make(map[string]*Subagent),
	}
	return r
}

func (r *Registry) RegisterBuiltins(subagents []*subagentstypes.Subagent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range subagents {
		r.subagents[s.Name] = s
	}
}

func (r *Registry) Register(subagent *Subagent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.subagents[subagent.Name] = subagent
}

func (r *Registry) Get(name string) (*Subagent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.subagents[name]
	if !ok {
		return nil, fmt.Errorf("subagent %q not found", name)
	}
	return s, nil
}

func (r *Registry) List() []*Subagent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*Subagent, 0, len(r.subagents))
	for _, s := range r.subagents {
		result = append(result, s)
	}
	return result
}

func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.subagents))
	for name := range r.subagents {
		names = append(names, name)
	}
	return names
}

func Register(subagent *Subagent) {
	globalRegistry.Register(subagent)
}

func Get(name string) (*Subagent, error) {
	return globalRegistry.Get(name)
}

func List() []*Subagent {
	return globalRegistry.List()
}

func Names() []string {
	return globalRegistry.Names()
}

func InitBuiltins(builtins []*subagentstypes.Subagent) {
	globalRegistry.RegisterBuiltins(builtins)
}
