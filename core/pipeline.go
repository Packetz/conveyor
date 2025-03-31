package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Event represents a pipeline event
type Event struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	PipelineID string                `json:"pipelineId,omitempty"`
	JobID     string                 `json:"jobId,omitempty"`
	StepID    string                 `json:"stepId,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// Pipeline represents a CI/CD pipeline
type Pipeline struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Stages      []Stage                `json:"stages"`
	Triggers    []Trigger              `json:"triggers,omitempty"`
	Cache       *CacheConfig           `json:"cache,omitempty"`
	Environment map[string]string      `json:"environment,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

// Stage represents a stage in a pipeline
type Stage struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Steps     []Step                 `json:"steps"`
	Needs     []string               `json:"needs,omitempty"`
	When      *ConditionalExecution  `json:"when,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Parallel  bool                   `json:"parallel"`
	DependsOn []string               `json:"dependsOn,omitempty"`
}

// Step represents a step in a pipeline stage
type Step struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Plugin      string                 `json:"plugin,omitempty"`
	Command     string                 `json:"command,omitempty"`
	Image       string                 `json:"image,omitempty"`
	Environment map[string]string      `json:"environment,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	When        *ConditionalExecution  `json:"when,omitempty"`
	Retry       *RetryConfig           `json:"retry,omitempty"`
	Timeout     string                 `json:"timeout,omitempty"`
	Cache       *CacheConfig           `json:"cache,omitempty"`
	DependsOn   []string               `json:"dependsOn,omitempty"`
	Outputs     map[string]string      `json:"outputs,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Trigger represents a pipeline trigger
type Trigger struct {
	Type     string   `json:"type"`
	Branches []string `json:"branches,omitempty"`
	Events   []string `json:"events,omitempty"`
	Paths    []string `json:"paths,omitempty"`
}

// ConditionalExecution represents a condition for executing a step or stage
type ConditionalExecution struct {
	Branch  string `json:"branch,omitempty"`
	Status  string `json:"status,omitempty"`
	Custom  string `json:"custom,omitempty"`
	Pattern string `json:"pattern,omitempty"`
}

// RetryConfig represents retry configuration for a step
type RetryConfig struct {
	MaxAttempts int    `json:"maxAttempts"`
	Interval    string `json:"interval,omitempty"`
	ExponentialBackoff bool `json:"exponentialBackoff,omitempty"`
}

// CacheConfig represents caching configuration
type CacheConfig struct {
	Key    string   `json:"key"`
	Paths  []string `json:"paths"`
	Policy string   `json:"policy,omitempty"`
}

// Job represents a pipeline execution
type Job struct {
	ID         string                 `json:"id"`
	PipelineID string                 `json:"pipelineId"`
	Status     string                 `json:"status"`
	Steps      []StepStatus           `json:"steps,omitempty"`
	StartedAt  time.Time              `json:"startedAt"`
	EndedAt    time.Time              `json:"endedAt,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Logs       []LogEntry             `json:"logs,omitempty"`
}

// StepStatus represents the status of a step execution
type StepStatus struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"startedAt"`
	EndedAt   time.Time `json:"endedAt,omitempty"`
	ExitCode  int       `json:"exitCode,omitempty"`
	Output    string    `json:"output,omitempty"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	StepID    string    `json:"stepId,omitempty"`
}

// PluginManifest represents a plugin manifest
type PluginManifest struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Type        string   `json:"type"`
	StepTypes   []string `json:"stepTypes"`
}

// PipelineEngine handles pipeline execution
type PipelineEngine struct {
	pipelines       map[string]*Pipeline
	jobs            map[string]*Job
	plugins         map[string]Plugin
	eventListeners  map[string]chan Event
	cacheManager    *CacheManager
	mu              sync.RWMutex
	eventsMu        sync.RWMutex
}

// Plugin interface for pipeline plugins
type Plugin interface {
	Execute(ctx context.Context, step Step) (map[string]interface{}, error)
	GetManifest() PluginManifest
}

// CacheManager handles caching of build artifacts
type CacheManager struct {
	caches map[string][]byte
	mu     sync.RWMutex
}

// NewPipelineEngine creates a new pipeline engine
func NewPipelineEngine() *PipelineEngine {
	return &PipelineEngine{
		pipelines:      make(map[string]*Pipeline),
		jobs:           make(map[string]*Job),
		plugins:        make(map[string]Plugin),
		eventListeners: make(map[string]chan Event),
		cacheManager:   &CacheManager{caches: make(map[string][]byte)},
	}
}

// RegisterPlugin registers a plugin with the engine
func (pe *PipelineEngine) RegisterPlugin(plugin Plugin) {
	manifest := plugin.GetManifest()
	pe.mu.Lock()
	pe.plugins[manifest.Name] = plugin
	pe.mu.Unlock()
}

// RegisterEventListener registers an event listener
func (pe *PipelineEngine) RegisterEventListener(id string, ch chan Event) {
	pe.eventsMu.Lock()
	pe.eventListeners[id] = ch
	pe.eventsMu.Unlock()
}

// UnregisterEventListener unregisters an event listener
func (pe *PipelineEngine) UnregisterEventListener(id string) {
	pe.eventsMu.Lock()
	delete(pe.eventListeners, id)
	pe.eventsMu.Unlock()
}

// emitEvent emits an event to all listeners
func (pe *PipelineEngine) emitEvent(event Event) {
	pe.eventsMu.RLock()
	defer pe.eventsMu.RUnlock()

	for _, ch := range pe.eventListeners {
		select {
		case ch <- event:
			// Event sent successfully
		default:
			// Channel buffer is full, just drop the event
		}
	}
}

// CreatePipeline creates a new pipeline
func (pe *PipelineEngine) CreatePipeline(pipeline *Pipeline) error {
	if pipeline.ID == "" {
		return fmt.Errorf("pipeline ID is required")
	}

	pe.mu.Lock()
	defer pe.mu.Unlock()

	if _, exists := pe.pipelines[pipeline.ID]; exists {
		return fmt.Errorf("pipeline with ID %s already exists", pipeline.ID)
	}

	now := time.Now()
	pipeline.CreatedAt = now
	pipeline.UpdatedAt = now

	pe.pipelines[pipeline.ID] = pipeline

	pe.emitEvent(Event{
		Type:      "pipeline.created",
		Timestamp: time.Now(),
		PipelineID: pipeline.ID,
		Data: map[string]interface{}{
			"name": pipeline.Name,
		},
	})

	return nil
}

// GetPipeline retrieves a pipeline by ID
func (pe *PipelineEngine) GetPipeline(id string) (*Pipeline, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	pipeline, exists := pe.pipelines[id]
	if !exists {
		return nil, fmt.Errorf("pipeline with ID %s not found", id)
	}

	return pipeline, nil
}

// ListPipelines returns all pipelines
func (pe *PipelineEngine) ListPipelines() []*Pipeline {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	pipelines := make([]*Pipeline, 0, len(pe.pipelines))
	for _, p := range pe.pipelines {
		pipelines = append(pipelines, p)
	}

	return pipelines
}

// DeletePipeline deletes a pipeline
func (pe *PipelineEngine) DeletePipeline(id string) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if _, exists := pe.pipelines[id]; !exists {
		return fmt.Errorf("pipeline with ID %s not found", id)
	}

	delete(pe.pipelines, id)

	pe.emitEvent(Event{
		Type:      "pipeline.deleted",
		Timestamp: time.Now(),
		PipelineID: id,
	})

	return nil
}

// ExecutePipeline executes a pipeline
func (pe *PipelineEngine) ExecutePipeline(pipelineID string) error {
	pe.mu.RLock()
	_, exists := pe.pipelines[pipelineID]
	pe.mu.RUnlock()

	if !exists {
		return fmt.Errorf("pipeline with ID %s not found", pipelineID)
	}

	// Create a new job
	job := &Job{
		ID:         fmt.Sprintf("job-%d", time.Now().Unix()),
		PipelineID: pipelineID,
		Status:     "running",
		StartedAt:  time.Now(),
		Steps:      []StepStatus{},
	}

	pe.mu.Lock()
	pe.jobs[job.ID] = job
	pe.mu.Unlock()

	pe.emitEvent(Event{
		Type:      "job.started",
		Timestamp: time.Now(),
		PipelineID: pipelineID,
		JobID:     job.ID,
	})

	// Execute the pipeline in a goroutine
	go func() {
		// Simulate pipeline execution
		// In a real implementation, this would execute stages and steps
		time.Sleep(2 * time.Second)

		pe.mu.Lock()
		job.Status = "success"
		job.EndedAt = time.Now()
		pe.mu.Unlock()

		pe.emitEvent(Event{
			Type:      "job.completed",
			Timestamp: time.Now(),
			PipelineID: pipelineID,
			JobID:     job.ID,
			Data: map[string]interface{}{
				"status": "success",
			},
		})
	}()

	return nil
}

// GetJob retrieves a job by ID
func (pe *PipelineEngine) GetJob(pipelineID, jobID string) (*Job, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	job, exists := pe.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job with ID %s not found", jobID)
	}

	if job.PipelineID != pipelineID {
		return nil, fmt.Errorf("job with ID %s is not associated with pipeline %s", jobID, pipelineID)
	}

	return job, nil
}

// ListJobs returns all jobs for a pipeline
func (pe *PipelineEngine) ListJobs(pipelineID string) ([]*Job, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	if _, exists := pe.pipelines[pipelineID]; !exists {
		return nil, fmt.Errorf("pipeline with ID %s not found", pipelineID)
	}

	jobs := make([]*Job, 0)
	for _, j := range pe.jobs {
		if j.PipelineID == pipelineID {
			jobs = append(jobs, j)
		}
	}

	return jobs, nil
}

// RetryJob retries a job
func (pe *PipelineEngine) RetryJob(pipelineID, jobID string) error {
	pe.mu.RLock()
	job, exists := pe.jobs[jobID]
	pe.mu.RUnlock()

	if !exists {
		return fmt.Errorf("job with ID %s not found", jobID)
	}

	if job.PipelineID != pipelineID {
		return fmt.Errorf("job with ID %s is not associated with pipeline %s", jobID, pipelineID)
	}

	// Create a new job based on the old one
	newJob := &Job{
		ID:         fmt.Sprintf("job-%d", time.Now().Unix()),
		PipelineID: pipelineID,
		Status:     "running",
		StartedAt:  time.Now(),
		Steps:      []StepStatus{},
		Metadata: map[string]interface{}{
			"retryOf": jobID,
		},
	}

	pe.mu.Lock()
	pe.jobs[newJob.ID] = newJob
	pe.mu.Unlock()

	pe.emitEvent(Event{
		Type:      "job.started",
		Timestamp: time.Now(),
		PipelineID: pipelineID,
		JobID:     newJob.ID,
		Data: map[string]interface{}{
			"retryOf": jobID,
		},
	})

	// Execute the job in a goroutine
	go func() {
		// Simulate job execution
		// In a real implementation, this would execute stages and steps
		time.Sleep(2 * time.Second)

		pe.mu.Lock()
		newJob.Status = "success"
		newJob.EndedAt = time.Now()
		pe.mu.Unlock()

		pe.emitEvent(Event{
			Type:      "job.completed",
			Timestamp: time.Now(),
			PipelineID: pipelineID,
			JobID:     newJob.ID,
			Data: map[string]interface{}{
				"status": "success",
				"retryOf": jobID,
			},
		})
	}()

	return nil
}

// AddJob adds a job to the engine
func (pe *PipelineEngine) AddJob(job *Job) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	
	pe.jobs[job.ID] = job
	
	// Emit an event for this job addition
	pe.emitEvent(Event{
		Type:      "job.added",
		Timestamp: time.Now(),
		PipelineID: job.PipelineID,
		JobID:     job.ID,
		Data: map[string]interface{}{
			"status": job.Status,
		},
	})
	
	// If the job is running, emit a job.started event
	if job.Status == "running" {
		pe.emitEvent(Event{
			Type:      "job.started",
			Timestamp: time.Now(),
			PipelineID: job.PipelineID,
			JobID:     job.ID,
		})
	} else if job.Status == "success" || job.Status == "failed" {
		eventType := "job.completed"
		pe.emitEvent(Event{
			Type:      eventType,
			Timestamp: time.Now(),
			PipelineID: job.PipelineID,
			JobID:     job.ID,
			Data: map[string]interface{}{
				"status": job.Status,
			},
		})
	}
}

// UpdateJob updates a job in the engine
func (pe *PipelineEngine) UpdateJob(job *Job) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	
	// Check if the job exists
	_, exists := pe.jobs[job.ID]
	if !exists {
		return fmt.Errorf("job with ID %s not found", job.ID)
	}
	
	// Update the job
	pe.jobs[job.ID] = job
	
	return nil
}

// EmitStepStartedEvent emits a step started event
func (pe *PipelineEngine) EmitStepStartedEvent(pipelineID, jobID, stepID string) {
	pe.emitEvent(Event{
		Type:       "step.started",
		Timestamp:  time.Now(),
		PipelineID: pipelineID,
		JobID:      jobID,
		StepID:     stepID,
	})
}

// EmitStepCompletedEvent emits a step completed event
func (pe *PipelineEngine) EmitStepCompletedEvent(pipelineID, jobID, stepID, status string) {
	pe.emitEvent(Event{
		Type:       "step.completed",
		Timestamp:  time.Now(),
		PipelineID: pipelineID,
		JobID:      jobID,
		StepID:     stepID,
		Data: map[string]interface{}{
			"status": status,
		},
	})
}

// EmitJobCompletedEvent emits a job completed event
func (pe *PipelineEngine) EmitJobCompletedEvent(pipelineID, jobID, status string) {
	pe.emitEvent(Event{
		Type:       "job.completed",
		Timestamp:  time.Now(),
		PipelineID: pipelineID,
		JobID:      jobID,
		Data: map[string]interface{}{
			"status": status,
		},
	})
} 