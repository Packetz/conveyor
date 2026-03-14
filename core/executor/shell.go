package executor

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/chip/conveyor/core"
)

const defaultTimeout = 10 * time.Minute

// ShellExecutor runs step commands via the host shell.
type ShellExecutor struct{}

// Execute runs a step's command via sh -c and captures output.
func (e *ShellExecutor) Execute(ctx context.Context, step core.Step, envVars map[string]string, onOutput OutputHandler) (*StepResult, error) {
	if step.Command == "" {
		return nil, fmt.Errorf("step %q has no command to execute", step.Name)
	}

	// Parse timeout
	timeout := defaultTimeout
	if step.Timeout != "" {
		parsed, err := time.ParseDuration(step.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout %q: %w", step.Timeout, err)
		}
		timeout = parsed
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, "sh", "-c", step.Command)

	// Build environment: os env + pipeline env + step env
	env := os.Environ()
	for k, v := range envVars {
		env = append(env, k+"="+v)
	}
	for k, v := range step.Environment {
		env = append(env, k+"="+v)
	}
	cmd.Env = env

	// Set up pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start command: %w", err)
	}

	// Read output from both pipes concurrently
	var output strings.Builder
	var mu sync.Mutex
	var wg sync.WaitGroup

	scanPipe := func(scanner *bufio.Scanner, stream string) {
		defer wg.Done()
		for scanner.Scan() {
			line := scanner.Text()
			mu.Lock()
			output.WriteString(line)
			output.WriteString("\n")
			mu.Unlock()
			if onOutput != nil {
				onOutput(line, stream)
			}
		}
	}

	wg.Add(2)
	go scanPipe(bufio.NewScanner(stdout), "stdout")
	go scanPipe(bufio.NewScanner(stderr), "stderr")

	// Wait for output to be fully read
	wg.Wait()

	// Wait for command to finish
	waitErr := cmd.Wait()

	result := &StepResult{
		Output: output.String(),
	}

	if waitErr != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			result.ExitCode = -1
			return result, fmt.Errorf("timed out after %s", timeout)
		}

		// Extract exit code
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			return result, fmt.Errorf("command exited with code %d", result.ExitCode)
		}

		return result, fmt.Errorf("command failed: %w", waitErr)
	}

	result.ExitCode = 0
	return result, nil
}
