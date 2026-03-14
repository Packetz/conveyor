package executor

import (
	"context"

	"github.com/chip/conveyor/core"
)

// StepResult holds the outcome of a step execution.
type StepResult struct {
	ExitCode int
	Output   string
}

// OutputHandler is called for each line of output during execution.
// stream is "stdout" or "stderr".
type OutputHandler func(line string, stream string)

// StepExecutor runs a single step's command.
type StepExecutor interface {
	Execute(ctx context.Context, step core.Step, envVars map[string]string, onOutput OutputHandler) (*StepResult, error)
}

// EngineAccessor provides the orchestrator access to engine capabilities
// without creating an import cycle with the core package.
type EngineAccessor interface {
	GetPlugin(name string) (core.Plugin, bool)
	EmitStepStartedEvent(pipelineID, jobID, stepID string)
	EmitStepCompletedEvent(pipelineID, jobID, stepID, status string)
	EmitJobCompletedEvent(pipelineID, jobID, status string)
	EmitEvent(event core.Event)
	LockJob(jobID string)
	UnlockJob(jobID string)
}
