package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Pipeline represents a CI/CD pipeline
type Pipeline struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Steps       []Step                 `json:"steps"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	Cache       map[string]interface{} `json:"cache,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Step represents a step in a pipeline
type Step struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Config   map[string]interface{} `json:"config"`
	Status   string                 `json:"status"`
	DependsOn []string              `json:"dependsOn,omitempty"`
	Retries  int                    `json:"retries"`
	Cache    map[string]interface{} `json:"cache,omitempty"`
}

// PipelineEngine handles the execution of pipelines
type PipelineEngine struct {
	pipelines      map[string]*Pipeline
	plugins        map[string]Plugin
	mu             sync.RWMutex
	eventListeners []func(Event)
	cacheManager   *CacheManager
}

// CacheManager handles caching of build artifacts and dependencies
type CacheManager struct {
	cache map[string]interface{}
	mu    sync.RWMutex
}

// Event represents a pipeline event
type Event struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Pipeline  string      `json:"pipeline"`
	Step      string      `json:"step,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

// Plugin interface for extending pipeline functionality
type Plugin interface {
	Execute(ctx context.Context, step Step) (map[string]interface{}, error)
	GetManifest() PluginManifest
}

// PluginManifest contains metadata about a plugin
type PluginManifest struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	StepTypes   []string `json:"stepTypes"`
}

// NewPipelineEngine creates a new pipeline engine
func NewPipelineEngine() *PipelineEngine {
	return &PipelineEngine{
		pipelines:    make(map[string]*Pipeline),
		plugins:      make(map[string]Plugin),
		cacheManager: &CacheManager{cache: make(map[string]interface{})},
	}
}

// RegisterPlugin registers a plugin with the engine
func (e *PipelineEngine) RegisterPlugin(p Plugin) {
	manifest := p.GetManifest()
	e.mu.Lock()
	defer e.mu.Unlock()
	e.plugins[manifest.Name] = p
}

// CreatePipeline creates a new pipeline
func (e *PipelineEngine) CreatePipeline(name, description string, steps []Step) (*Pipeline, error) {
	now := time.Now()
	pipeline := &Pipeline{
		ID:          fmt.Sprintf("pl-%d", now.UnixNano()),
		Name:        name,
		Description: description,
		Steps:       steps,
		Status:      "idle",
		CreatedAt:   now,
		UpdatedAt:   now,
		Cache:       make(map[string]interface{}),
		Metadata:    make(map[string]interface{}),
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	e.pipelines[pipeline.ID] = pipeline

	e.emitEvent(Event{
		Type:      "PIPELINE_CREATED",
		Timestamp: time.Now(),
		Pipeline:  pipeline.ID,
		Data:      pipeline,
	})

	return pipeline, nil
}

// GetPipeline retrieves a pipeline by ID
func (e *PipelineEngine) GetPipeline(id string) (*Pipeline, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if pipeline, ok := e.pipelines[id]; ok {
		return pipeline, nil
	}

	return nil, fmt.Errorf("pipeline not found: %s", id)
}

// GetPipelines retrieves all pipelines
func (e *PipelineEngine) GetPipelines() []*Pipeline {
	e.mu.RLock()
	defer e.mu.RUnlock()

	pipelines := make([]*Pipeline, 0, len(e.pipelines))
	for _, p := range e.pipelines {
		pipelines = append(pipelines, p)
	}

	return pipelines
}

// ExecutePipeline executes a pipeline with dynamic parallel execution
func (e *PipelineEngine) ExecutePipeline(ctx context.Context, id string) error {
	pipeline, err := e.GetPipeline(id)
	if err != nil {
		return err
	}

	// Update pipeline status
	e.mu.Lock()
	pipeline.Status = "running"
	pipeline.UpdatedAt = time.Now()
	e.mu.Unlock()

	e.emitEvent(Event{
		Type:      "PIPELINE_STARTED",
		Timestamp: time.Now(),
		Pipeline:  pipeline.ID,
		Data:      pipeline,
	})

	// Build dependency graph for parallel execution
	dependencyGraph := buildDependencyGraph(pipeline.Steps)
	
	// Create execution context with cancellation
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Execute steps with dependency management
	err = e.executeStepsWithDependencies(execCtx, pipeline, dependencyGraph)

	// Update pipeline status based on results
	e.mu.Lock()
	if err != nil {
		pipeline.Status = "failed"
	} else {
		pipeline.Status = "success"
	}
	pipeline.UpdatedAt = time.Now()
	e.mu.Unlock()

	e.emitEvent(Event{
		Type:      "PIPELINE_COMPLETED",
		Timestamp: time.Now(),
		Pipeline:  pipeline.ID,
		Data: map[string]interface{}{
			"status": pipeline.Status,
			"error":  err,
		},
	})

	return err
}

// executeStepsWithDependencies executes pipeline steps respecting dependencies
func (e *PipelineEngine) executeStepsWithDependencies(ctx context.Context, pipeline *Pipeline, dependencyGraph map[string][]string) error {
	var wg sync.WaitGroup
	stepResults := make(map[string]error)
	stepMu := sync.Mutex{}
	
	// Find steps with no dependencies (roots)
	ready := findRootsInGraph(dependencyGraph)
	remaining := len(pipeline.Steps)
	
	for remaining > 0 {
		// Execute all ready steps in parallel
		var execWg sync.WaitGroup
		for _, stepID := range ready {
			execWg.Add(1)
			go func(id string) {
				defer execWg.Done()
				
				// Find the step
				var step *Step
				for i := range pipeline.Steps {
					if pipeline.Steps[i].ID == id {
						step = &pipeline.Steps[i]
						break
					}
				}
				
				if step == nil {
					stepMu.Lock()
					stepResults[id] = fmt.Errorf("step not found: %s", id)
					stepMu.Unlock()
					return
				}
				
				// Check for cached results if enabled
				cacheKey := fmt.Sprintf("%s:%s", pipeline.ID, step.ID)
				e.cacheManager.mu.RLock()
				cachedResult, hasCached := e.cacheManager.cache[cacheKey]
				e.cacheManager.mu.RUnlock()
				
				// Update step status
				e.mu.Lock()
				step.Status = "running"
				e.mu.Unlock()
				
				e.emitEvent(Event{
					Type:      "STEP_STARTED",
					Timestamp: time.Now(),
					Pipeline:  pipeline.ID,
					Step:      step.ID,
					Data:      step,
				})
				
				var err error
				var result map[string]interface{}
				
				// Use cached result if available and cache is enabled
				if step.Cache != nil && hasCached {
					result = cachedResult.(map[string]interface{})
				} else {
					// Execute the step
					result, err = e.executeStep(ctx, *step)
					
					// Cache the result if caching is enabled
					if step.Cache != nil && err == nil {
						e.cacheManager.mu.Lock()
						e.cacheManager.cache[cacheKey] = result
						e.cacheManager.mu.Unlock()
					}
				}
				
				// Handle retries
				retries := step.Retries
				for err != nil && retries > 0 {
					e.emitEvent(Event{
						Type:      "STEP_RETRY",
						Timestamp: time.Now(),
						Pipeline:  pipeline.ID,
						Step:      step.ID,
						Data: map[string]interface{}{
							"error":       err.Error(),
							"retriesLeft": retries,
						},
					})
					
					// Execute the step again
					result, err = e.executeStep(ctx, *step)
					retries--
				}
				
				// Update step status
				e.mu.Lock()
				if err != nil {
					step.Status = "failed"
				} else {
					step.Status = "success"
				}
				e.mu.Unlock()
				
				stepMu.Lock()
				stepResults[id] = err
				stepMu.Unlock()
				
				e.emitEvent(Event{
					Type:      "STEP_COMPLETED",
					Timestamp: time.Now(),
					Pipeline:  pipeline.ID,
					Step:      step.ID,
					Data: map[string]interface{}{
						"status": step.Status,
						"error":  err,
					},
				})
			}(stepID)
		}
		
		// Wait for the current batch to complete
		execWg.Wait()
		
		// Remove completed steps from the graph
		for _, id := range ready {
			delete(dependencyGraph, id)
			remaining--
		}
		
		// Find new ready steps (those whose dependencies are all satisfied)
		ready = findNewReadySteps(dependencyGraph, stepResults)
		
		// If we have steps remaining but none are ready, we have a circular dependency
		if remaining > 0 && len(ready) == 0 {
			return fmt.Errorf("circular dependency detected or failed dependencies")
		}
	}
	
	// Check for any errors
	for _, err := range stepResults {
		if err != nil {
			return err
		}
	}
	
	return nil
}

// executeStep executes a single pipeline step
func (e *PipelineEngine) executeStep(ctx context.Context, step Step) (map[string]interface{}, error) {
	// Find the appropriate plugin
	e.mu.RLock()
	plugin, ok := e.plugins[step.Type]
	e.mu.RUnlock()
	
	if !ok {
		return nil, fmt.Errorf("unknown step type: %s", step.Type)
	}
	
	// Execute the step with the plugin
	return plugin.Execute(ctx, step)
}

// AddEventListener adds a listener for pipeline events
func (e *PipelineEngine) AddEventListener(listener func(Event)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.eventListeners = append(e.eventListeners, listener)
}

// emitEvent emits an event to all listeners
func (e *PipelineEngine) emitEvent(event Event) {
	e.mu.RLock()
	listeners := append([]func(Event){}, e.eventListeners...)
	e.mu.RUnlock()
	
	for _, listener := range listeners {
		go listener(event)
	}
}

// Helper functions for dependency management

// buildDependencyGraph builds a dependency graph from steps
func buildDependencyGraph(steps []Step) map[string][]string {
	graph := make(map[string][]string)
	
	for _, step := range steps {
		if len(step.DependsOn) > 0 {
			graph[step.ID] = step.DependsOn
		} else {
			graph[step.ID] = []string{}
		}
	}
	
	return graph
}

// findRootsInGraph finds steps with no dependencies
func findRootsInGraph(graph map[string][]string) []string {
	var roots []string
	
	for id, deps := range graph {
		if len(deps) == 0 {
			roots = append(roots, id)
		}
	}
	
	return roots
}

// findNewReadySteps finds steps whose dependencies are all satisfied
func findNewReadySteps(graph map[string][]string, results map[string]error) []string {
	var ready []string
	
	for id, deps := range graph {
		allDependenciesMet := true
		for _, dep := range deps {
			// Check if dependency was executed and successful
			if err, ok := results[dep]; !ok || err != nil {
				allDependenciesMet = false
				break
			}
		}
		
		if allDependenciesMet {
			ready = append(ready, id)
		}
	}
	
	return ready
} 