package core

import (
	"fmt"
	"sort"
	"sync"
)

// Factory builds a fresh, unloaded Core instance.
type Factory func() Core

type registration struct {
	meta    Metadata
	factory Factory
}

var (
	mu       sync.RWMutex
	registry = map[string]registration{}
)

// Register makes a system core available to the rest of the app. Systems
// call this from an init() in their own package, so simply importing a
// system package (even with a blank identifier) is enough to plug it in.
func Register(meta Metadata, factory Factory) {
	mu.Lock()
	defer mu.Unlock()
	registry[meta.ID] = registration{meta: meta, factory: factory}
}

// New instantiates a fresh core by system id.
func New(id string) (Core, error) {
	mu.RLock()
	reg, ok := registry[id]
	mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("core: no system registered with id %q", id)
	}
	return reg.factory(), nil
}

// List returns metadata for every registered system, sorted by name.
func List() []Metadata {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Metadata, 0, len(registry))
	for _, reg := range registry {
		out = append(out, reg.meta)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// ExtensionIndex builds a lowercase-extension -> system-id lookup table
// from every registered core, used by the library scanner to classify files.
func ExtensionIndex() map[string]string {
	mu.RLock()
	defer mu.RUnlock()
	index := make(map[string]string)
	for _, reg := range registry {
		for _, ext := range reg.meta.Extensions {
			index[ext] = reg.meta.ID
		}
	}
	return index
}
