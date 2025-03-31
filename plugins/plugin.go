package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"plugin"
	"sync"
)

// Plugin represents a Conveyor plugin
type Plugin struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	Config      map[string]interface{} `json:"config"`
	Enabled     bool                   `json:"enabled"`
}

// PluginManager handles plugin loading and management
type PluginManager struct {
	plugins map[string]*Plugin
	mu      sync.RWMutex
}

// NewPluginManager creates a new plugin manager
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]*Plugin),
	}
}

// LoadPlugin loads a plugin from the given path
func (pm *PluginManager) LoadPlugin(path string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Read plugin manifest
	manifest, err := os.Open(filepath.Join(path, "manifest.json"))
	if err != nil {
		return fmt.Errorf("failed to read plugin manifest: %w", err)
	}
	defer manifest.Close()

	var p Plugin
	if err := json.NewDecoder(manifest).Decode(&p); err != nil {
		return fmt.Errorf("failed to decode plugin manifest: %w", err)
	}

	// Load plugin binary
	plug, err := plugin.Open(filepath.Join(path, p.Name+".so"))
	if err != nil {
		return fmt.Errorf("failed to load plugin binary: %w", err)
	}

	// Initialize plugin
	initFunc, err := plug.Lookup("Init")
	if err != nil {
		return fmt.Errorf("plugin does not export Init function: %w", err)
	}

	if init, ok := initFunc.(func(map[string]interface{}) error); ok {
		if err := init(p.Config); err != nil {
			return fmt.Errorf("failed to initialize plugin: %w", err)
		}
	}

	pm.plugins[p.Name] = &p
	return nil
}

// GetPlugin returns a plugin by name
func (pm *PluginManager) GetPlugin(name string) (*Plugin, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	p, exists := pm.plugins[name]
	return p, exists
}

// ListPlugins returns all loaded plugins
func (pm *PluginManager) ListPlugins() []*Plugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	plugins := make([]*Plugin, 0, len(pm.plugins))
	for _, p := range pm.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}

// EnablePlugin enables a plugin
func (pm *PluginManager) EnablePlugin(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	p.Enabled = true
	return nil
}

// DisablePlugin disables a plugin
func (pm *PluginManager) DisablePlugin(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	p.Enabled = false
	return nil
}

// ExecutePlugin executes a plugin's main function
func (pm *PluginManager) ExecutePlugin(ctx context.Context, name string, args ...interface{}) (interface{}, error) {
	pm.mu.RLock()
	p, exists := pm.plugins[name]
	pm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	if !p.Enabled {
		return nil, fmt.Errorf("plugin %s is disabled", name)
	}

	plug, err := plugin.Open(filepath.Join("plugins", name+".so"))
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin binary: %w", err)
	}

	mainFunc, err := plug.Lookup("Main")
	if err != nil {
		return nil, fmt.Errorf("plugin does not export Main function: %w", err)
	}

	if main, ok := mainFunc.(func(context.Context, ...interface{}) (interface{}, error)); ok {
		return main(ctx, args...)
	}

	return nil, fmt.Errorf("plugin Main function has invalid signature")
}

// PluginContext provides context for plugin execution
type PluginContext struct {
	Context context.Context
	Config  map[string]interface{}
	Output  io.Writer
}

// NewPluginContext creates a new plugin context
func NewPluginContext(ctx context.Context, config map[string]interface{}, output io.Writer) *PluginContext {
	return &PluginContext{
		Context: ctx,
		Config:  config,
		Output:  output,
	}
} 