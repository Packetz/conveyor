package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chip/conveyor/core"
)

// PipelineOrchestrator coordinates the execution of a pipeline job.
type PipelineOrchestrator struct {
	engine   EngineAccessor
	executor StepExecutor
}

// NewPipelineOrchestrator creates a new orchestrator.
func NewPipelineOrchestrator(engine EngineAccessor, executor StepExecutor) *PipelineOrchestrator {
	return &PipelineOrchestrator{
		engine:   engine,
		executor: executor,
	}
}

// topologicalSort sorts stages so dependencies come before dependents.
// Returns an error if a cycle is detected.
func topologicalSort(stages []core.Stage) ([]core.Stage, error) {
	stageByID := make(map[string]core.Stage)
	for _, s := range stages {
		stageByID[s.ID] = s
	}

	// Validate all dependencies reference existing stages
	for _, s := range stages {
		for _, need := range s.Needs {
			if _, ok := stageByID[need]; !ok {
				return nil, fmt.Errorf("stage %q depends on unknown stage %q", s.ID, need)
			}
		}
	}

	// Kahn's algorithm
	inDegree := make(map[string]int)
	for _, s := range stages {
		if _, ok := inDegree[s.ID]; !ok {
			inDegree[s.ID] = 0
		}
		for range s.Needs {
			inDegree[s.ID]++
		}
	}

	var queue []string
	for _, s := range stages {
		if inDegree[s.ID] == 0 {
			queue = append(queue, s.ID)
		}
	}

	// Build adjacency list: need -> stages that depend on it
	dependents := make(map[string][]string)
	for _, s := range stages {
		for _, need := range s.Needs {
			dependents[need] = append(dependents[need], s.ID)
		}
	}

	var sorted []core.Stage
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		sorted = append(sorted, stageByID[id])

		for _, dep := range dependents[id] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if len(sorted) != len(stages) {
		return nil, fmt.Errorf("cycle detected in stage dependencies")
	}

	return sorted, nil
}

// RunJob executes all stages and steps of a pipeline for a given job.
func (o *PipelineOrchestrator) RunJob(ctx context.Context, pipeline *core.Pipeline, job *core.Job) error {
	// Topological sort stages
	sortedStages, err := topologicalSort(pipeline.Stages)
	if err != nil {
		job.Status = "failed"
		job.EndedAt = time.Now()
		o.engine.EmitJobCompletedEvent(pipeline.ID, job.ID, "failed")
		return fmt.Errorf("dependency error: %w", err)
	}

	completedStages := make(map[string]string) // stage ID -> status

	for _, stage := range sortedStages {
		// Validate dependencies
		for _, need := range stage.Needs {
			status, ok := completedStages[need]
			if !ok {
				o.engine.LockJob(job.ID)
				job.Status = "failed"
				job.EndedAt = time.Now()
				o.engine.UnlockJob(job.ID)
				o.engine.EmitJobCompletedEvent(pipeline.ID, job.ID, "failed")
				return fmt.Errorf("dependency stage %q not found", need)
			}
			if status != "success" {
				o.engine.LockJob(job.ID)
				job.Status = "failed"
				job.EndedAt = time.Now()
				o.engine.UnlockJob(job.ID)
				o.engine.EmitJobCompletedEvent(pipeline.ID, job.ID, "failed")
				return fmt.Errorf("dependency stage %q has status %q, expected success", need, status)
			}
		}

		stageStatus := "success"

		for _, step := range stage.Steps {
			// Emit step started
			o.engine.EmitStepStartedEvent(pipeline.ID, job.ID, step.ID)

			stepStatus := core.StepStatus{
				ID:        step.ID,
				Name:      step.Name,
				Status:    "running",
				StartedAt: time.Now(),
			}

			var result *StepResult
			var stepErr error

			if step.Plugin != "" {
				result, stepErr = o.executePluginStep(ctx, step)
			} else if step.Command != "" {
				result, stepErr = o.executeShellStep(ctx, step, pipeline.ID, job.ID, pipeline.Environment)
			} else {
				// No command or plugin — skip
				stepStatus.Status = "skipped"
				stepStatus.EndedAt = time.Now()
				o.engine.LockJob(job.ID)
				job.Steps = append(job.Steps, stepStatus)
				o.engine.UnlockJob(job.ID)
				o.engine.EmitStepCompletedEvent(pipeline.ID, job.ID, step.ID, "skipped")
				continue
			}

			// Handle retry logic
			if stepErr != nil && step.Retry != nil && step.Retry.MaxAttempts > 1 {
				for attempt := 2; attempt <= step.Retry.MaxAttempts; attempt++ {
					// Emit retry event
					o.engine.EmitEvent(core.Event{
						Type:       "step.retry",
						Timestamp:  time.Now(),
						PipelineID: pipeline.ID,
						JobID:      job.ID,
						StepID:     step.ID,
						Data: map[string]interface{}{
							"attempt":     attempt,
							"maxAttempts": step.Retry.MaxAttempts,
							"error":       stepErr.Error(),
						},
					})

					// Wait interval between retries
					if step.Retry.Interval != "" {
						interval, parseErr := time.ParseDuration(step.Retry.Interval)
						if parseErr == nil {
							time.Sleep(interval)
						}
					}

					if step.Plugin != "" {
						result, stepErr = o.executePluginStep(ctx, step)
					} else {
						result, stepErr = o.executeShellStep(ctx, step, pipeline.ID, job.ID, pipeline.Environment)
					}

					if stepErr == nil {
						break
					}
				}
			}

			// Record step result
			stepStatus.EndedAt = time.Now()
			if stepErr != nil {
				stepStatus.Status = "failed"
				if result != nil {
					stepStatus.ExitCode = result.ExitCode
					stepStatus.Output = result.Output
				}
				o.engine.LockJob(job.ID)
				job.Steps = append(job.Steps, stepStatus)
				o.engine.UnlockJob(job.ID)
				o.engine.EmitStepCompletedEvent(pipeline.ID, job.ID, step.ID, "failed")
				stageStatus = "failed"
				break // Stop remaining steps in this stage
			}

			stepStatus.Status = "success"
			stepStatus.ExitCode = result.ExitCode
			stepStatus.Output = result.Output
			o.engine.LockJob(job.ID)
			job.Steps = append(job.Steps, stepStatus)
			o.engine.UnlockJob(job.ID)
			o.engine.EmitStepCompletedEvent(pipeline.ID, job.ID, step.ID, "success")
		}

		completedStages[stage.ID] = stageStatus

		if stageStatus == "failed" {
			o.engine.LockJob(job.ID)
			job.Status = "failed"
			job.EndedAt = time.Now()
			o.engine.UnlockJob(job.ID)
			o.engine.EmitJobCompletedEvent(pipeline.ID, job.ID, "failed")
			return nil
		}
	}

	o.engine.LockJob(job.ID)
	job.Status = "success"
	job.EndedAt = time.Now()
	o.engine.UnlockJob(job.ID)
	o.engine.EmitJobCompletedEvent(pipeline.ID, job.ID, "success")
	return nil
}

// executeShellStep runs a step via the shell executor.
func (o *PipelineOrchestrator) executeShellStep(ctx context.Context, step core.Step, pipelineID, jobID string, pipelineEnv map[string]string) (*StepResult, error) {
	onOutput := func(line string, stream string) {
		o.engine.EmitEvent(core.Event{
			Type:       "step.output",
			Timestamp:  time.Now(),
			PipelineID: pipelineID,
			JobID:      jobID,
			StepID:     step.ID,
			Data: map[string]interface{}{
				"line":   line,
				"stream": stream,
			},
		})
	}

	return o.executor.Execute(ctx, step, pipelineEnv, onOutput)
}

// executePluginStep runs a step via the plugin system.
func (o *PipelineOrchestrator) executePluginStep(ctx context.Context, step core.Step) (*StepResult, error) {
	plugin, ok := o.engine.GetPlugin(step.Plugin)
	if !ok {
		return nil, fmt.Errorf("plugin %q not found", step.Plugin)
	}

	pluginResult, err := plugin.Execute(ctx, step)
	if err != nil {
		return &StepResult{ExitCode: 1}, err
	}

	// Serialize plugin result as output
	output, _ := json.Marshal(pluginResult)
	return &StepResult{ExitCode: 0, Output: string(output)}, nil
}
