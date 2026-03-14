package executor

import (
	"context"
	"strings"
	"testing"

	"github.com/chip/conveyor/core"
)

func TestShellExecutor_BasicCommand(t *testing.T) {
	exec := &ShellExecutor{}
	step := core.Step{Command: "echo hello"}

	result, err := exec.Execute(context.Background(), step, nil, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.Output != "hello\n" {
		t.Errorf("Output = %q, want %q", result.Output, "hello\n")
	}
}

func TestShellExecutor_NonZeroExit(t *testing.T) {
	exec := &ShellExecutor{}
	step := core.Step{Command: "exit 42"}

	result, err := exec.Execute(context.Background(), step, nil, nil)
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	if result.ExitCode != 42 {
		t.Errorf("ExitCode = %d, want 42", result.ExitCode)
	}
}

func TestShellExecutor_StderrCapture(t *testing.T) {
	exec := &ShellExecutor{}
	step := core.Step{Command: "echo error-output >&2"}

	var lines []struct{ line, stream string }
	handler := func(line string, stream string) {
		lines = append(lines, struct{ line, stream string }{line, stream})
	}

	result, err := exec.Execute(context.Background(), step, nil, handler)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if len(lines) == 0 {
		t.Fatal("expected output lines, got none")
	}
	if lines[0].stream != "stderr" {
		t.Errorf("stream = %q, want %q", lines[0].stream, "stderr")
	}
	if lines[0].line != "error-output" {
		t.Errorf("line = %q, want %q", lines[0].line, "error-output")
	}
}

func TestShellExecutor_OutputHandler(t *testing.T) {
	exec := &ShellExecutor{}
	step := core.Step{Command: "echo line1; echo line2; echo line3"}

	var lines []string
	handler := func(line string, stream string) {
		lines = append(lines, line)
	}

	_, err := exec.Execute(context.Background(), step, nil, handler)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	expected := []string{"line1", "line2", "line3"}
	for i, want := range expected {
		if lines[i] != want {
			t.Errorf("lines[%d] = %q, want %q", i, lines[i], want)
		}
	}
}

func TestShellExecutor_Timeout(t *testing.T) {
	exec := &ShellExecutor{}
	step := core.Step{
		Command: "sleep 30",
		Timeout: "1s",
	}

	result, err := exec.Execute(context.Background(), step, nil, nil)
	if err == nil {
		t.Fatal("Execute() expected timeout error, got nil")
	}
	if result.ExitCode != -1 {
		t.Errorf("ExitCode = %d, want -1", result.ExitCode)
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "timed out")
	}
}

func TestShellExecutor_EnvVars(t *testing.T) {
	exec := &ShellExecutor{}
	step := core.Step{
		Command:     "echo $PIPELINE_VAR $STEP_VAR",
		Environment: map[string]string{"STEP_VAR": "step-value"},
	}
	envVars := map[string]string{"PIPELINE_VAR": "pipeline-value"}

	result, err := exec.Execute(context.Background(), step, envVars, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	expected := "pipeline-value step-value\n"
	if result.Output != expected {
		t.Errorf("Output = %q, want %q", result.Output, expected)
	}
}

func TestShellExecutor_StepEnvOverridesPipelineEnv(t *testing.T) {
	exec := &ShellExecutor{}
	step := core.Step{
		Command:     "echo $MY_VAR",
		Environment: map[string]string{"MY_VAR": "from-step"},
	}
	envVars := map[string]string{"MY_VAR": "from-pipeline"}

	result, err := exec.Execute(context.Background(), step, envVars, nil)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	expected := "from-step\n"
	if result.Output != expected {
		t.Errorf("Output = %q, want %q", result.Output, expected)
	}
}

func TestShellExecutor_EmptyCommand(t *testing.T) {
	exec := &ShellExecutor{}
	step := core.Step{Name: "empty-step"}

	_, err := exec.Execute(context.Background(), step, nil, nil)
	if err == nil {
		t.Fatal("Execute() expected error for empty command, got nil")
	}
	if !strings.Contains(err.Error(), "no command") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "no command")
	}
}
