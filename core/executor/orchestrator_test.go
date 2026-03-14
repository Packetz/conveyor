package executor

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/chip/conveyor/core"
)

// mockExecutor records calls and returns configured results.
type mockExecutor struct {
	results map[string]*StepResult // keyed by step command
	errors  map[string]error
	calls   []core.Step
}

func newMockExecutor() *mockExecutor {
	return &mockExecutor{
		results: make(map[string]*StepResult),
		errors:  make(map[string]error),
	}
}

func (m *mockExecutor) Execute(ctx context.Context, step core.Step, envVars map[string]string, onOutput OutputHandler) (*StepResult, error) {
	m.calls = append(m.calls, step)
	if err, ok := m.errors[step.Command]; ok {
		r := m.results[step.Command]
		if r == nil {
			r = &StepResult{ExitCode: 1}
		}
		return r, err
	}
	if r, ok := m.results[step.Command]; ok {
		return r, nil
	}
	return &StepResult{ExitCode: 0, Output: ""}, nil
}

// mockEngine records emitted events.
type mockEngine struct {
	events  []core.Event
	plugins map[string]core.Plugin
}

func newMockEngine() *mockEngine {
	return &mockEngine{
		plugins: make(map[string]core.Plugin),
	}
}

func (m *mockEngine) GetPlugin(name string) (core.Plugin, bool) {
	p, ok := m.plugins[name]
	return p, ok
}

func (m *mockEngine) EmitStepStartedEvent(pipelineID, jobID, stepID string) {
	m.events = append(m.events, core.Event{Type: "step.started", PipelineID: pipelineID, JobID: jobID, StepID: stepID})
}

func (m *mockEngine) EmitStepCompletedEvent(pipelineID, jobID, stepID, status string) {
	m.events = append(m.events, core.Event{Type: "step.completed", PipelineID: pipelineID, JobID: jobID, StepID: stepID, Data: map[string]interface{}{"status": status}})
}

func (m *mockEngine) EmitJobCompletedEvent(pipelineID, jobID, status string) {
	m.events = append(m.events, core.Event{Type: "job.completed", PipelineID: pipelineID, JobID: jobID, Data: map[string]interface{}{"status": status}})
}

func (m *mockEngine) EmitEvent(event core.Event) {
	m.events = append(m.events, event)
}

func (m *mockEngine) LockJob(jobID string)   {}
func (m *mockEngine) UnlockJob(jobID string) {}

// mockPlugin implements core.Plugin for testing.
type mockPlugin struct {
	result map[string]interface{}
	err    error
}

func (p *mockPlugin) Execute(ctx context.Context, step core.Step) (map[string]interface{}, error) {
	return p.result, p.err
}

func (p *mockPlugin) GetManifest() core.PluginManifest {
	return core.PluginManifest{Name: "test-plugin"}
}

// --- Topological Sort Tests ---

func TestTopologicalSort_NoDependencies(t *testing.T) {
	stages := []core.Stage{
		{ID: "a", Name: "A"},
		{ID: "b", Name: "B"},
		{ID: "c", Name: "C"},
	}

	sorted, err := topologicalSort(stages)
	if err != nil {
		t.Fatalf("topologicalSort() error = %v", err)
	}
	if len(sorted) != 3 {
		t.Fatalf("len(sorted) = %d, want 3", len(sorted))
	}
}

func TestTopologicalSort_WithDependencies(t *testing.T) {
	stages := []core.Stage{
		{ID: "deploy", Name: "Deploy", Needs: []string{"build"}},
		{ID: "build", Name: "Build", Needs: []string{"test"}},
		{ID: "test", Name: "Test"},
	}

	sorted, err := topologicalSort(stages)
	if err != nil {
		t.Fatalf("topologicalSort() error = %v", err)
	}
	if len(sorted) != 3 {
		t.Fatalf("len(sorted) = %d, want 3", len(sorted))
	}

	// test must come before build, build must come before deploy
	idxOf := make(map[string]int)
	for i, s := range sorted {
		idxOf[s.ID] = i
	}
	if idxOf["test"] >= idxOf["build"] {
		t.Errorf("test (idx %d) should come before build (idx %d)", idxOf["test"], idxOf["build"])
	}
	if idxOf["build"] >= idxOf["deploy"] {
		t.Errorf("build (idx %d) should come before deploy (idx %d)", idxOf["build"], idxOf["deploy"])
	}
}

func TestTopologicalSort_CycleDetected(t *testing.T) {
	stages := []core.Stage{
		{ID: "a", Name: "A", Needs: []string{"b"}},
		{ID: "b", Name: "B", Needs: []string{"a"}},
	}

	_, err := topologicalSort(stages)
	if err == nil {
		t.Fatal("topologicalSort() expected cycle error, got nil")
	}
}

// --- RunJob Tests ---

func TestRunJob_SingleStageSuccess(t *testing.T) {
	engine := newMockEngine()
	exec := newMockExecutor()
	orch := NewPipelineOrchestrator(engine, exec)

	pipeline := &core.Pipeline{
		ID:   "p1",
		Name: "Test Pipeline",
		Stages: []core.Stage{
			{
				ID:   "build",
				Name: "Build",
				Steps: []core.Step{
					{ID: "build-compile", Name: "compile", Command: "go build"},
				},
			},
		},
	}
	job := &core.Job{
		ID:         "j1",
		PipelineID: "p1",
		Status:     "running",
		StartedAt:  time.Now(),
	}

	err := orch.RunJob(context.Background(), pipeline, job)
	if err != nil {
		t.Fatalf("RunJob() error = %v", err)
	}
	if job.Status != "success" {
		t.Errorf("job.Status = %q, want %q", job.Status, "success")
	}
	if len(exec.calls) != 1 {
		t.Fatalf("executor called %d times, want 1", len(exec.calls))
	}
	if exec.calls[0].Command != "go build" {
		t.Errorf("executed command = %q, want %q", exec.calls[0].Command, "go build")
	}
}

func TestRunJob_StepFailureStopsStage(t *testing.T) {
	engine := newMockEngine()
	exec := newMockExecutor()
	exec.errors["fail-cmd"] = fmt.Errorf("command exited with code 1")
	exec.results["fail-cmd"] = &StepResult{ExitCode: 1, Output: "error output"}

	orch := NewPipelineOrchestrator(engine, exec)

	pipeline := &core.Pipeline{
		ID: "p1",
		Stages: []core.Stage{
			{
				ID: "build",
				Steps: []core.Step{
					{ID: "s1", Name: "step1", Command: "fail-cmd"},
					{ID: "s2", Name: "step2", Command: "should-not-run"},
				},
			},
		},
	}
	job := &core.Job{ID: "j1", PipelineID: "p1", Status: "running", StartedAt: time.Now()}

	orch.RunJob(context.Background(), pipeline, job)

	if job.Status != "failed" {
		t.Errorf("job.Status = %q, want %q", job.Status, "failed")
	}
	if len(exec.calls) != 1 {
		t.Errorf("executor called %d times, want 1 (second step should not run)", len(exec.calls))
	}
}

func TestRunJob_DependencyValidation(t *testing.T) {
	engine := newMockEngine()
	exec := newMockExecutor()
	orch := NewPipelineOrchestrator(engine, exec)

	pipeline := &core.Pipeline{
		ID: "p1",
		Stages: []core.Stage{
			{
				ID:    "test",
				Name:  "Test",
				Steps: []core.Step{{ID: "s1", Command: "echo test"}},
			},
			{
				ID:    "build",
				Name:  "Build",
				Needs: []string{"test"},
				Steps: []core.Step{{ID: "s2", Command: "echo build"}},
			},
			{
				ID:    "deploy",
				Name:  "Deploy",
				Needs: []string{"build"},
				Steps: []core.Step{{ID: "s3", Command: "echo deploy"}},
			},
		},
	}
	job := &core.Job{ID: "j1", PipelineID: "p1", Status: "running", StartedAt: time.Now()}

	err := orch.RunJob(context.Background(), pipeline, job)
	if err != nil {
		t.Fatalf("RunJob() error = %v", err)
	}
	if job.Status != "success" {
		t.Errorf("job.Status = %q, want %q", job.Status, "success")
	}
	if len(exec.calls) != 3 {
		t.Errorf("executor called %d times, want 3", len(exec.calls))
	}
}

func TestRunJob_FailedDependencySkipsDownstream(t *testing.T) {
	engine := newMockEngine()
	exec := newMockExecutor()
	exec.errors["echo test"] = fmt.Errorf("command exited with code 1")
	exec.results["echo test"] = &StepResult{ExitCode: 1}

	orch := NewPipelineOrchestrator(engine, exec)

	pipeline := &core.Pipeline{
		ID: "p1",
		Stages: []core.Stage{
			{
				ID:    "test",
				Steps: []core.Step{{ID: "s1", Command: "echo test"}},
			},
			{
				ID:    "build",
				Needs: []string{"test"},
				Steps: []core.Step{{ID: "s2", Command: "echo build"}},
			},
		},
	}
	job := &core.Job{ID: "j1", PipelineID: "p1", Status: "running", StartedAt: time.Now()}

	orch.RunJob(context.Background(), pipeline, job)

	if job.Status != "failed" {
		t.Errorf("job.Status = %q, want %q", job.Status, "failed")
	}
	// Only the first step should have run
	if len(exec.calls) != 1 {
		t.Errorf("executor called %d times, want 1", len(exec.calls))
	}
}

func TestRunJob_PluginStep(t *testing.T) {
	engine := newMockEngine()
	engine.plugins["security"] = &mockPlugin{
		result: map[string]interface{}{"status": "ok", "findings": 0},
	}
	exec := newMockExecutor()
	orch := NewPipelineOrchestrator(engine, exec)

	pipeline := &core.Pipeline{
		ID: "p1",
		Stages: []core.Stage{
			{
				ID: "scan",
				Steps: []core.Step{
					{ID: "s1", Name: "security-scan", Plugin: "security", Type: "vulnerability-scan"},
				},
			},
		},
	}
	job := &core.Job{ID: "j1", PipelineID: "p1", Status: "running", StartedAt: time.Now()}

	err := orch.RunJob(context.Background(), pipeline, job)
	if err != nil {
		t.Fatalf("RunJob() error = %v", err)
	}
	if job.Status != "success" {
		t.Errorf("job.Status = %q, want %q", job.Status, "success")
	}
	// Shell executor should NOT have been called
	if len(exec.calls) != 0 {
		t.Errorf("shell executor called %d times, want 0", len(exec.calls))
	}
	// Step output should contain serialized plugin result
	if len(job.Steps) != 1 {
		t.Fatalf("job.Steps len = %d, want 1", len(job.Steps))
	}
	if job.Steps[0].Status != "success" {
		t.Errorf("step status = %q, want %q", job.Steps[0].Status, "success")
	}
}

func TestRunJob_PluginNotFound(t *testing.T) {
	engine := newMockEngine()
	exec := newMockExecutor()
	orch := NewPipelineOrchestrator(engine, exec)

	pipeline := &core.Pipeline{
		ID: "p1",
		Stages: []core.Stage{
			{
				ID: "scan",
				Steps: []core.Step{
					{ID: "s1", Plugin: "nonexistent"},
				},
			},
		},
	}
	job := &core.Job{ID: "j1", PipelineID: "p1", Status: "running", StartedAt: time.Now()}

	orch.RunJob(context.Background(), pipeline, job)

	if job.Status != "failed" {
		t.Errorf("job.Status = %q, want %q", job.Status, "failed")
	}
}

func TestRunJob_SkippedStep(t *testing.T) {
	engine := newMockEngine()
	exec := newMockExecutor()
	orch := NewPipelineOrchestrator(engine, exec)

	pipeline := &core.Pipeline{
		ID: "p1",
		Stages: []core.Stage{
			{
				ID: "build",
				Steps: []core.Step{
					{ID: "s1", Name: "empty-step"},
					{ID: "s2", Name: "real-step", Command: "echo ok"},
				},
			},
		},
	}
	job := &core.Job{ID: "j1", PipelineID: "p1", Status: "running", StartedAt: time.Now()}

	err := orch.RunJob(context.Background(), pipeline, job)
	if err != nil {
		t.Fatalf("RunJob() error = %v", err)
	}
	if job.Status != "success" {
		t.Errorf("job.Status = %q, want %q", job.Status, "success")
	}
	if len(job.Steps) != 2 {
		t.Fatalf("job.Steps len = %d, want 2", len(job.Steps))
	}
	if job.Steps[0].Status != "skipped" {
		t.Errorf("step[0].Status = %q, want %q", job.Steps[0].Status, "skipped")
	}
	if job.Steps[1].Status != "success" {
		t.Errorf("step[1].Status = %q, want %q", job.Steps[1].Status, "success")
	}
}

func TestRunJob_RetrySuccess(t *testing.T) {
	engine := newMockEngine()
	// Custom executor that fails first call, succeeds second
	customExec := &retryTestExecutor{failUntil: 2}
	orch := NewPipelineOrchestrator(engine, customExec)

	pipeline := &core.Pipeline{
		ID: "p1",
		Stages: []core.Stage{
			{
				ID: "build",
				Steps: []core.Step{
					{
						ID:      "s1",
						Name:    "flaky-step",
						Command: "flaky-cmd",
						Retry:   &core.RetryConfig{MaxAttempts: 3, Interval: "1ms"},
					},
				},
			},
		},
	}
	job := &core.Job{ID: "j1", PipelineID: "p1", Status: "running", StartedAt: time.Now()}

	err := orch.RunJob(context.Background(), pipeline, job)
	if err != nil {
		t.Fatalf("RunJob() error = %v", err)
	}
	if job.Status != "success" {
		t.Errorf("job.Status = %q, want %q", job.Status, "success")
	}
	// Verify retry event was emitted
	retryEvents := 0
	for _, e := range engine.events {
		if e.Type == "step.retry" {
			retryEvents++
		}
	}
	if retryEvents != 1 {
		t.Errorf("retry events = %d, want 1", retryEvents)
	}
}

// retryTestExecutor fails the first N-1 calls, then succeeds.
type retryTestExecutor struct {
	callCount int
	failUntil int // succeed on this call number
}

func (e *retryTestExecutor) Execute(ctx context.Context, step core.Step, envVars map[string]string, onOutput OutputHandler) (*StepResult, error) {
	e.callCount++
	if e.callCount < e.failUntil {
		return &StepResult{ExitCode: 1}, fmt.Errorf("command exited with code 1")
	}
	return &StepResult{ExitCode: 0, Output: "success"}, nil
}

func TestRunJob_EventSequence(t *testing.T) {
	engine := newMockEngine()
	exec := newMockExecutor()
	orch := NewPipelineOrchestrator(engine, exec)

	pipeline := &core.Pipeline{
		ID: "p1",
		Stages: []core.Stage{
			{
				ID: "build",
				Steps: []core.Step{
					{ID: "s1", Name: "compile", Command: "echo ok"},
				},
			},
		},
	}
	job := &core.Job{ID: "j1", PipelineID: "p1", Status: "running", StartedAt: time.Now()}

	orch.RunJob(context.Background(), pipeline, job)

	// Expected event sequence: step.started, step.completed, job.completed
	expectedTypes := []string{"step.started", "step.completed", "job.completed"}
	if len(engine.events) < len(expectedTypes) {
		t.Fatalf("got %d events, want at least %d", len(engine.events), len(expectedTypes))
	}

	// Check first event is step.started
	if engine.events[0].Type != "step.started" {
		t.Errorf("event[0].Type = %q, want %q", engine.events[0].Type, "step.started")
	}
	if engine.events[0].StepID != "s1" {
		t.Errorf("event[0].StepID = %q, want %q", engine.events[0].StepID, "s1")
	}

	// Find step.completed event
	found := false
	for _, e := range engine.events {
		if e.Type == "step.completed" && e.StepID == "s1" {
			found = true
			if e.Data["status"] != "success" {
				t.Errorf("step.completed status = %q, want %q", e.Data["status"], "success")
			}
		}
	}
	if !found {
		t.Error("step.completed event not found")
	}

	// Last event should be job.completed
	last := engine.events[len(engine.events)-1]
	if last.Type != "job.completed" {
		t.Errorf("last event type = %q, want %q", last.Type, "job.completed")
	}
}

// --- Integration Test ---

func TestIntegration_FullPipelineExecution(t *testing.T) {
	engine := newMockEngine()
	shellExec := &ShellExecutor{}
	orch := NewPipelineOrchestrator(engine, shellExec)

	pipeline := &core.Pipeline{
		ID:          "integration-test",
		Name:        "Integration Test",
		Environment: map[string]string{"TEST_VAR": "hello"},
		Stages: []core.Stage{
			{
				ID:   "setup",
				Name: "Setup",
				Steps: []core.Step{
					{ID: "setup-echo", Name: "echo-setup", Command: "echo setting up"},
				},
			},
			{
				ID:    "build",
				Name:  "Build",
				Needs: []string{"setup"},
				Steps: []core.Step{
					{ID: "build-compile", Name: "compile", Command: "echo building with $TEST_VAR"},
				},
			},
			{
				ID:    "test",
				Name:  "Test",
				Needs: []string{"build"},
				Steps: []core.Step{
					{ID: "test-run", Name: "run-tests", Command: "echo testing; echo done"},
				},
			},
		},
	}
	job := &core.Job{ID: "j1", PipelineID: "integration-test", Status: "running", StartedAt: time.Now()}

	err := orch.RunJob(context.Background(), pipeline, job)
	if err != nil {
		t.Fatalf("RunJob() error = %v", err)
	}
	if job.Status != "success" {
		t.Errorf("job.Status = %q, want %q", job.Status, "success")
	}
	if len(job.Steps) != 3 {
		t.Fatalf("job.Steps len = %d, want 3", len(job.Steps))
	}
	for i, step := range job.Steps {
		if step.Status != "success" {
			t.Errorf("step[%d].Status = %q, want %q", i, step.Status, "success")
		}
		if step.ExitCode != 0 {
			t.Errorf("step[%d].ExitCode = %d, want 0", i, step.ExitCode)
		}
		if step.Output == "" {
			t.Errorf("step[%d].Output is empty", i)
		}
	}

	// Verify env var was passed
	if !strings.Contains(job.Steps[1].Output, "hello") {
		t.Errorf("step[1].Output = %q, expected to contain %q", job.Steps[1].Output, "hello")
	}

	// Verify events were emitted
	stepStarted := 0
	stepCompleted := 0
	jobCompleted := 0
	for _, e := range engine.events {
		switch e.Type {
		case "step.started":
			stepStarted++
		case "step.completed":
			stepCompleted++
		case "job.completed":
			jobCompleted++
		}
	}
	if stepStarted != 3 {
		t.Errorf("step.started events = %d, want 3", stepStarted)
	}
	if stepCompleted != 3 {
		t.Errorf("step.completed events = %d, want 3", stepCompleted)
	}
	if jobCompleted != 1 {
		t.Errorf("job.completed events = %d, want 1", jobCompleted)
	}
}
