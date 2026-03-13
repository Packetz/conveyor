# YAML Pipeline Loader Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Parse YAML pipeline definitions into core.Pipeline structs with validation, directory scanning, and an API import endpoint.

**Architecture:** New `core/loader/` package with separate YAML types, a validator, and a converter. The loader bridges YAML files to the existing PipelineEngine. Standalone Parse/Validate/Convert functions enable dry-run usage.

**Tech Stack:** Go, `gopkg.in/yaml.v3`, existing `core` and `api` packages.

**Spec:** `docs/superpowers/specs/2026-03-13-yaml-pipeline-loader-design.md`

---

## Chunk 1: Foundation — Types, Slugify, and Parse

### Task 1: Add yaml.v3 dependency

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: Add the dependency**

Run: `cd /Users/chipsteen/Code/conveyor && go get gopkg.in/yaml.v3`

- [ ] **Step 2: Verify it was added**

Run: `grep yaml.v3 go.mod`
Expected: `gopkg.in/yaml.v3 v3.x.x`

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add gopkg.in/yaml.v3 dependency for pipeline loader"
```

### Task 2: Create YAML types

**Files:**
- Create: `core/loader/types.go`

- [ ] **Step 1: Create the types file**

```go
package loader

// YAMLPipeline is the top-level YAML pipeline representation.
type YAMLPipeline struct {
	Name          string            `yaml:"name"`
	Description   string            `yaml:"description"`
	Version       string            `yaml:"version"`
	Triggers      []YAMLTrigger     `yaml:"triggers"`
	Environment   *YAMLEnvironment  `yaml:"environment"`
	Cache         *YAMLCache        `yaml:"cache"`
	Stages        []YAMLStage       `yaml:"stages"`
	Notifications interface{}       `yaml:"notifications"`
	Artifacts     interface{}       `yaml:"artifacts"`
}

// YAMLEnvironment holds environment variable configuration.
type YAMLEnvironment struct {
	Variables map[string]string `yaml:"variables"`
}

// YAMLTrigger represents a pipeline trigger.
type YAMLTrigger struct {
	Type     string   `yaml:"type"`
	Branches []string `yaml:"branches"`
	Events   []string `yaml:"events"`
	Paths    []string `yaml:"paths"`
}

// YAMLCache represents cache configuration.
type YAMLCache struct {
	Key    string   `yaml:"key"`
	Paths  []string `yaml:"paths"`
	Policy string   `yaml:"policy"`
}

// YAMLStage represents a pipeline stage.
type YAMLStage struct {
	Name  string     `yaml:"name"`
	Needs []string   `yaml:"needs"`
	When  *YAMLWhen  `yaml:"when"`
	Steps []YAMLStep `yaml:"steps"`
}

// YAMLStep represents a step within a stage.
type YAMLStep struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Type        string                 `yaml:"type"`
	Run         string                 `yaml:"run"`
	Plugin      string                 `yaml:"plugin"`
	Image       string                 `yaml:"image"`
	Environment map[string]string      `yaml:"environment"`
	Config      map[string]interface{} `yaml:"config"`
	When        *YAMLWhen              `yaml:"when"`
	Retry       *YAMLRetry             `yaml:"retry"`
	Timeout     string                 `yaml:"timeout"`
	Cache       *YAMLCache             `yaml:"cache"`
	DependsOn   []string               `yaml:"depends_on"`
	Outputs     map[string]string      `yaml:"outputs"`
}

// YAMLWhen represents conditional execution configuration.
type YAMLWhen struct {
	Branch  string `yaml:"branch"`
	Status  string `yaml:"status"`
	Custom  string `yaml:"custom"`
	Pattern string `yaml:"pattern"`
}

// YAMLRetry represents retry configuration.
type YAMLRetry struct {
	MaxAttempts        int    `yaml:"max_attempts"`
	Interval           string `yaml:"interval"`
	ExponentialBackoff bool   `yaml:"exponential_backoff"`
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/chipsteen/Code/conveyor && go build ./core/loader/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add core/loader/types.go
git commit -m "feat(loader): add YAML pipeline type definitions"
```

### Task 3: Implement Slugify with tests (TDD)

**Files:**
- Create: `core/loader/slugify.go`
- Create: `core/loader/slugify_test.go`

- [ ] **Step 1: Write the failing tests**

```go
package loader

import "testing"

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "Hello World", "hello-world"},
		{"underscores", "pre_build_step", "pre-build-step"},
		{"special chars", "Build & Test!", "build-test"},
		{"consecutive hyphens", "build---test", "build-test"},
		{"leading/trailing hyphens", "-build-test-", "build-test"},
		{"already slugified", "secure-build", "secure-build"},
		{"mixed case with numbers", "Stage 2 Build", "stage-2-build"},
		{"empty string", "", ""},
		{"only special chars", "!@#$%", ""},
		{"spaces and tabs", "  hello   world  ", "hello-world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slugify(tt.input)
			if result != tt.expected {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/chipsteen/Code/conveyor && go test ./core/loader/ -run TestSlugify -v`
Expected: FAIL — `Slugify` not defined

- [ ] **Step 3: Implement Slugify**

```go
package loader

import (
	"regexp"
	"strings"
)

var (
	nonAlphanumHyphen = regexp.MustCompile(`[^a-z0-9-]`)
	multipleHyphens   = regexp.MustCompile(`-{2,}`)
)

// Slugify converts a name to a URL-safe lowercase ID.
// Algorithm: lowercase, replace spaces/underscores with hyphens,
// strip non-[a-z0-9-], collapse consecutive hyphens, trim hyphens.
func Slugify(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	s = nonAlphanumHyphen.ReplaceAllString(s, "")
	s = multipleHyphens.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/chipsteen/Code/conveyor && go test ./core/loader/ -run TestSlugify -v`
Expected: PASS — all 10 cases

- [ ] **Step 5: Commit**

```bash
git add core/loader/slugify.go core/loader/slugify_test.go
git commit -m "feat(loader): implement Slugify function with tests"
```

### Task 4: Implement Parse with tests (TDD)

**Files:**
- Create: `core/loader/parse.go`
- Create: `core/loader/parse_test.go`
- Create: `core/loader/testdata/valid/minimal.yaml`
- Create: `core/loader/testdata/invalid/bad-syntax.yaml`

- [ ] **Step 1: Create test fixture files**

`core/loader/testdata/valid/minimal.yaml`:
```yaml
name: minimal-pipeline
description: A minimal valid pipeline

stages:
  - name: build
    steps:
      - name: compile
        run: go build ./...
```

`core/loader/testdata/invalid/bad-syntax.yaml`:
```yaml
name: broken
stages:
  - name: [invalid yaml
    steps:
      - this is not valid::: yaml: {{
```

- [ ] **Step 2: Write the failing tests**

```go
package loader

import (
	"os"
	"testing"
)

func TestParse_ValidMinimal(t *testing.T) {
	data, err := os.ReadFile("testdata/valid/minimal.yaml")
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	p, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if p.Name != "minimal-pipeline" {
		t.Errorf("Name = %q, want %q", p.Name, "minimal-pipeline")
	}
	if len(p.Stages) != 1 {
		t.Fatalf("len(Stages) = %d, want 1", len(p.Stages))
	}
	if p.Stages[0].Name != "build" {
		t.Errorf("Stages[0].Name = %q, want %q", p.Stages[0].Name, "build")
	}
	if len(p.Stages[0].Steps) != 1 {
		t.Fatalf("len(Steps) = %d, want 1", len(p.Stages[0].Steps))
	}
	if p.Stages[0].Steps[0].Run != "go build ./..." {
		t.Errorf("Steps[0].Run = %q, want %q", p.Stages[0].Steps[0].Run, "go build ./...")
	}
}

func TestParse_BadSyntax(t *testing.T) {
	data, err := os.ReadFile("testdata/invalid/bad-syntax.yaml")
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	_, err = Parse(data)
	if err == nil {
		t.Error("Parse() expected error for bad YAML syntax, got nil")
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd /Users/chipsteen/Code/conveyor && go test ./core/loader/ -run TestParse -v`
Expected: FAIL — `Parse` not defined

- [ ] **Step 4: Implement Parse**

```go
package loader

import "gopkg.in/yaml.v3"

// Parse unmarshals YAML bytes into a YAMLPipeline struct.
func Parse(data []byte) (*YAMLPipeline, error) {
	var p YAMLPipeline
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /Users/chipsteen/Code/conveyor && go test ./core/loader/ -run TestParse -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add core/loader/parse.go core/loader/parse_test.go core/loader/testdata/
git commit -m "feat(loader): implement Parse function with test fixtures"
```

## Chunk 2: Validator

### Task 5: Implement Validate with tests (TDD)

**Files:**
- Create: `core/loader/validator.go`
- Create: `core/loader/validator_test.go`
- Create: `core/loader/testdata/invalid/missing-name.yaml`
- Create: `core/loader/testdata/invalid/empty-stages.yaml`
- Create: `core/loader/testdata/invalid/no-run-or-plugin.yaml`
- Create: `core/loader/testdata/invalid/circular-deps.yaml`

- [ ] **Step 1: Create test fixture files**

`core/loader/testdata/invalid/missing-name.yaml`:
```yaml
stages:
  - name: build
    steps:
      - name: compile
        run: go build ./...
```

`core/loader/testdata/invalid/empty-stages.yaml`:
```yaml
name: empty-stages
stages: []
```

`core/loader/testdata/invalid/no-run-or-plugin.yaml`:
```yaml
name: no-run-or-plugin
stages:
  - name: build
    steps:
      - name: broken-step
```

`core/loader/testdata/invalid/circular-deps.yaml`:
```yaml
name: circular
stages:
  - name: stage-a
    needs: [stage-b]
    steps:
      - name: step-a
        run: echo a
  - name: stage-b
    needs: [stage-a]
    steps:
      - name: step-b
        run: echo b
```

- [ ] **Step 2: Write the failing tests**

```go
package loader

import (
	"os"
	"strings"
	"testing"
)

func mustParse(t *testing.T, path string) *YAMLPipeline {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	p, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse(%s) error = %v", path, err)
	}
	return p
}

func TestValidate_ValidMinimal(t *testing.T) {
	p := mustParse(t, "testdata/valid/minimal.yaml")
	warnings, err := Validate(p)
	if err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
	if len(warnings) != 0 {
		t.Errorf("Validate() warnings = %v, want none", warnings)
	}
}

func TestValidate_MissingName(t *testing.T) {
	p := mustParse(t, "testdata/invalid/missing-name.yaml")
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error = %q, want it to mention 'name'", err.Error())
	}
}

func TestValidate_EmptyStages(t *testing.T) {
	p := mustParse(t, "testdata/invalid/empty-stages.yaml")
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "stage") {
		t.Errorf("error = %q, want it to mention 'stage'", err.Error())
	}
}

func TestValidate_NoRunOrPlugin(t *testing.T) {
	p := mustParse(t, "testdata/invalid/no-run-or-plugin.yaml")
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "run") || !strings.Contains(err.Error(), "plugin") {
		t.Errorf("error = %q, want it to mention 'run' and 'plugin'", err.Error())
	}
}

func TestValidate_BothRunAndPlugin(t *testing.T) {
	p := &YAMLPipeline{
		Name: "both",
		Stages: []YAMLStage{
			{
				Name: "build",
				Steps: []YAMLStep{
					{Name: "step", Run: "echo hi", Plugin: "security"},
				},
			},
		},
	}
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}
}

func TestValidate_CircularDeps(t *testing.T) {
	p := mustParse(t, "testdata/invalid/circular-deps.yaml")
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("error = %q, want it to mention 'circular'", err.Error())
	}
}

func TestValidate_InvalidNeedsRef(t *testing.T) {
	p := &YAMLPipeline{
		Name: "bad-ref",
		Stages: []YAMLStage{
			{
				Name:  "build",
				Needs: []string{"nonexistent"},
				Steps: []YAMLStep{
					{Name: "step", Run: "echo hi"},
				},
			},
		},
	}
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error = %q, want it to mention 'nonexistent'", err.Error())
	}
}

func TestValidate_DiamondDeps(t *testing.T) {
	p := &YAMLPipeline{
		Name: "diamond",
		Stages: []YAMLStage{
			{Name: "a", Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
			{Name: "b", Needs: []string{"a"}, Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
			{Name: "c", Needs: []string{"a"}, Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
			{Name: "d", Needs: []string{"b", "c"}, Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
		},
	}
	_, err := Validate(p)
	if err != nil {
		t.Errorf("Validate() error = %v, want nil for diamond deps", err)
	}
}

func TestValidate_UnsupportedFieldsWarnings(t *testing.T) {
	p := &YAMLPipeline{
		Name:          "with-extras",
		Version:       "1.0.0",
		Notifications: map[string]interface{}{"type": "slack"},
		Artifacts:     []interface{}{"report.html"},
		Stages: []YAMLStage{
			{Name: "build", Steps: []YAMLStep{{Name: "step", Run: "echo"}}},
		},
	}
	warnings, err := Validate(p)
	if err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
	if len(warnings) != 3 {
		t.Errorf("len(warnings) = %d, want 3", len(warnings))
	}
}

func TestValidate_MissingStageName(t *testing.T) {
	p := &YAMLPipeline{
		Name: "missing-stage-name",
		Stages: []YAMLStage{
			{Steps: []YAMLStep{{Name: "step", Run: "echo"}}},
		},
	}
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}
}

func TestValidate_MissingStepName(t *testing.T) {
	p := &YAMLPipeline{
		Name: "missing-step-name",
		Stages: []YAMLStage{
			{Name: "build", Steps: []YAMLStep{{Run: "echo"}}},
		},
	}
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error, got nil")
	}
}

func TestValidate_DuplicateStageSlug(t *testing.T) {
	p := &YAMLPipeline{
		Name: "dup-slugs",
		Stages: []YAMLStage{
			{Name: "Build Test", Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
			{Name: "build-test", Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
		},
	}
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error for duplicate slugified stage IDs, got nil")
	}
}

func TestValidate_EmptyStage(t *testing.T) {
	p := &YAMLPipeline{
		Name: "empty-stage",
		Stages: []YAMLStage{
			{Name: "build", Steps: []YAMLStep{}},
		},
	}
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error for stage with zero steps, got nil")
	}
	if !strings.Contains(err.Error(), "at least one step") {
		t.Errorf("error = %q, want it to mention 'at least one step'", err.Error())
	}
}

func TestValidate_ThreeNodeCycle(t *testing.T) {
	p := &YAMLPipeline{
		Name: "three-node-cycle",
		Stages: []YAMLStage{
			{Name: "a", Needs: []string{"c"}, Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
			{Name: "b", Needs: []string{"a"}, Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
			{Name: "c", Needs: []string{"b"}, Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
		},
	}
	_, err := Validate(p)
	if err == nil {
		t.Fatal("Validate() expected error for 3-node circular dep, got nil")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("error = %q, want it to mention 'circular'", err.Error())
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd /Users/chipsteen/Code/conveyor && go test ./core/loader/ -run TestValidate -v`
Expected: FAIL — `Validate` not defined

- [ ] **Step 4: Implement Validate**

```go
package loader

import (
	"fmt"
	"strings"
)

// Validate checks a YAMLPipeline for errors and returns warnings for unsupported fields.
// Returns a non-nil error if the pipeline is invalid (hard error).
// Returns warnings for fields that are recognized but not yet supported.
func Validate(p *YAMLPipeline) ([]string, error) {
	var errs []string
	var warnings []string

	// Required: name
	if strings.TrimSpace(p.Name) == "" {
		errs = append(errs, "pipeline name is required")
	}

	// Required: at least one stage
	if len(p.Stages) == 0 {
		errs = append(errs, "pipeline must have at least one stage")
	}

	// Build set of stage names for needs validation
	stageNames := make(map[string]bool)
	slugSeen := make(map[string]string) // slug -> original name

	for i, stage := range p.Stages {
		// Required: stage name
		if strings.TrimSpace(stage.Name) == "" {
			errs = append(errs, fmt.Sprintf("stage %d: name is required", i+1))
			continue
		}

		slug := Slugify(stage.Name)
		if prevName, exists := slugSeen[slug]; exists {
			errs = append(errs, fmt.Sprintf("stage %q: duplicate slugified ID %q (conflicts with stage %q)", stage.Name, slug, prevName))
		}
		slugSeen[slug] = stage.Name
		stageNames[stage.Name] = true

		// Required: at least one step per stage
		if len(stage.Steps) == 0 {
			errs = append(errs, fmt.Sprintf("stage %q: must have at least one step", stage.Name))
		}

		for j, step := range stage.Steps {
			// Required: step name
			if strings.TrimSpace(step.Name) == "" {
				errs = append(errs, fmt.Sprintf("stage %q, step %d: name is required", stage.Name, j+1))
				continue
			}

			// Required: exactly one of run or plugin
			hasRun := strings.TrimSpace(step.Run) != ""
			hasPlugin := strings.TrimSpace(step.Plugin) != ""
			if !hasRun && !hasPlugin {
				errs = append(errs, fmt.Sprintf("stage %q, step %q: must have either 'run' or 'plugin'", stage.Name, step.Name))
			}
			if hasRun && hasPlugin {
				errs = append(errs, fmt.Sprintf("stage %q, step %q: cannot have both 'run' and 'plugin'", stage.Name, step.Name))
			}
		}
	}

	// Validate needs references
	for _, stage := range p.Stages {
		for _, need := range stage.Needs {
			if !stageNames[need] {
				errs = append(errs, fmt.Sprintf("stage %q: needs references unknown stage %q", stage.Name, need))
			}
		}
	}

	// Check for circular dependencies
	if err := detectCycles(p.Stages); err != nil {
		errs = append(errs, err.Error())
	}

	// Warnings for unsupported fields
	if strings.TrimSpace(p.Version) != "" {
		warnings = append(warnings, "field 'version' is not yet supported and will be ignored")
	}
	if p.Notifications != nil {
		warnings = append(warnings, "field 'notifications' is not yet supported and will be ignored")
	}
	if p.Artifacts != nil {
		warnings = append(warnings, "field 'artifacts' is not yet supported and will be ignored")
	}

	if len(errs) > 0 {
		return warnings, fmt.Errorf("validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return warnings, nil
}

// detectCycles checks for circular dependencies in stage needs using DFS.
func detectCycles(stages []YAMLStage) error {
	// Build adjacency map: stage name -> stages it depends on
	adj := make(map[string][]string)
	for _, s := range stages {
		adj[s.Name] = s.Needs
	}

	const (
		unvisited = 0
		visiting  = 1
		visited   = 2
	)

	state := make(map[string]int)

	var visit func(name string, path []string) error
	visit = func(name string, path []string) error {
		if state[name] == visited {
			return nil
		}
		if state[name] == visiting {
			return fmt.Errorf("circular dependency detected: %s -> %s", strings.Join(path, " -> "), name)
		}

		state[name] = visiting
		path = append(path, name)

		for _, dep := range adj[name] {
			if err := visit(dep, path); err != nil {
				return err
			}
		}

		state[name] = visited
		return nil
	}

	for _, s := range stages {
		if state[s.Name] == unvisited {
			if err := visit(s.Name, nil); err != nil {
				return err
			}
		}
	}

	return nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /Users/chipsteen/Code/conveyor && go test ./core/loader/ -run TestValidate -v`
Expected: PASS — all 14 test cases

- [ ] **Step 6: Commit**

```bash
git add core/loader/validator.go core/loader/validator_test.go core/loader/testdata/invalid/
git commit -m "feat(loader): implement Validate function with comprehensive tests"
```

## Chunk 3: Convert and Loader

### Task 6: Implement Convert with tests (TDD)

**Files:**
- Create: `core/loader/convert.go`
- Create: `core/loader/convert_test.go`

- [ ] **Step 1: Write the failing tests**

```go
package loader

import (
	"os"
	"testing"
)

func TestConvert_MinimalPipeline(t *testing.T) {
	p := &YAMLPipeline{
		Name:        "My Pipeline",
		Description: "A test pipeline",
		Stages: []YAMLStage{
			{
				Name: "build",
				Steps: []YAMLStep{
					{Name: "compile", Run: "go build ./..."},
				},
			},
		},
	}

	pipeline, err := Convert(p, "my-pipeline")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if pipeline.ID != "my-pipeline" {
		t.Errorf("ID = %q, want %q", pipeline.ID, "my-pipeline")
	}
	if pipeline.Name != "My Pipeline" {
		t.Errorf("Name = %q, want %q", pipeline.Name, "My Pipeline")
	}
	if len(pipeline.Stages) != 1 {
		t.Fatalf("len(Stages) = %d, want 1", len(pipeline.Stages))
	}

	stage := pipeline.Stages[0]
	if stage.ID != "build" {
		t.Errorf("Stage.ID = %q, want %q", stage.ID, "build")
	}
	if len(stage.Steps) != 1 {
		t.Fatalf("len(Steps) = %d, want 1", len(stage.Steps))
	}

	step := stage.Steps[0]
	if step.ID != "build-compile" {
		t.Errorf("Step.ID = %q, want %q", step.ID, "build-compile")
	}
	if step.Command != "go build ./..." {
		t.Errorf("Step.Command = %q, want %q", step.Command, "go build ./...")
	}
	if step.Type != "script" {
		t.Errorf("Step.Type = %q, want %q", step.Type, "script")
	}
}

func TestConvert_PluginStep(t *testing.T) {
	p := &YAMLPipeline{
		Name: "plugin-test",
		Stages: []YAMLStage{
			{
				Name: "scan",
				Steps: []YAMLStep{
					{
						Name:   "secret-scan",
						Plugin: "security-scanner",
						Config: map[string]interface{}{
							"scanTypes":         []interface{}{"secret"},
							"severityThreshold": "HIGH",
						},
					},
				},
			},
		},
	}

	pipeline, err := Convert(p, "plugin-test")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	step := pipeline.Stages[0].Steps[0]
	if step.Type != "plugin" {
		t.Errorf("Step.Type = %q, want %q", step.Type, "plugin")
	}
	if step.Plugin != "security-scanner" {
		t.Errorf("Step.Plugin = %q, want %q", step.Plugin, "security-scanner")
	}
}

func TestConvert_ExplicitType(t *testing.T) {
	p := &YAMLPipeline{
		Name: "explicit-type",
		Stages: []YAMLStage{
			{
				Name: "build",
				Steps: []YAMLStep{
					{Name: "compile", Run: "go build", Type: "build"},
				},
			},
		},
	}

	pipeline, err := Convert(p, "explicit-type")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}
	if pipeline.Stages[0].Steps[0].Type != "build" {
		t.Errorf("Step.Type = %q, want %q", pipeline.Stages[0].Steps[0].Type, "build")
	}
}

func TestConvert_DescriptionInMetadata(t *testing.T) {
	p := &YAMLPipeline{
		Name: "desc-test",
		Stages: []YAMLStage{
			{
				Name: "build",
				Steps: []YAMLStep{
					{Name: "compile", Run: "go build", Description: "Compile the project"},
				},
			},
		},
	}

	pipeline, err := Convert(p, "desc-test")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	step := pipeline.Stages[0].Steps[0]
	if step.Metadata == nil {
		t.Fatal("Step.Metadata is nil, want description key")
	}
	desc, ok := step.Metadata["description"]
	if !ok {
		t.Fatal("Step.Metadata missing 'description' key")
	}
	if desc != "Compile the project" {
		t.Errorf("description = %q, want %q", desc, "Compile the project")
	}
}

func TestConvert_NeedsResolution(t *testing.T) {
	p := &YAMLPipeline{
		Name: "needs-test",
		Stages: []YAMLStage{
			{Name: "Pre Build", Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
			{Name: "Build", Needs: []string{"Pre Build"}, Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
		},
	}

	pipeline, err := Convert(p, "needs-test")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if len(pipeline.Stages[1].Needs) != 1 || pipeline.Stages[1].Needs[0] != "pre-build" {
		t.Errorf("Needs = %v, want [pre-build]", pipeline.Stages[1].Needs)
	}
}

func TestConvert_Environment(t *testing.T) {
	p := &YAMLPipeline{
		Name: "env-test",
		Environment: &YAMLEnvironment{
			Variables: map[string]string{"GO_VERSION": "1.21", "NODE_ENV": "test"},
		},
		Stages: []YAMLStage{
			{Name: "build", Steps: []YAMLStep{{Name: "s", Run: "echo"}}},
		},
	}

	pipeline, err := Convert(p, "env-test")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if pipeline.Environment["GO_VERSION"] != "1.21" {
		t.Errorf("Environment[GO_VERSION] = %q, want %q", pipeline.Environment["GO_VERSION"], "1.21")
	}
}

func TestConvert_SecureBuildFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/valid/secure-build.yaml")
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	yp, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	pipeline, err := Convert(yp, "secure-build")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if pipeline.ID != "secure-build" {
		t.Errorf("ID = %q, want %q", pipeline.ID, "secure-build")
	}
	if pipeline.Name != "secure-build" {
		t.Errorf("Name = %q, want %q", pipeline.Name, "secure-build")
	}
	if len(pipeline.Stages) != 6 {
		t.Errorf("len(Stages) = %d, want 6", len(pipeline.Stages))
	}
	if len(pipeline.Triggers) != 2 {
		t.Errorf("len(Triggers) = %d, want 2", len(pipeline.Triggers))
	}
	if pipeline.Cache == nil {
		t.Error("Cache is nil, want non-nil")
	}
}
```

- [ ] **Step 2: Copy the secure-build fixture**

Run: `cp /Users/chipsteen/Code/conveyor/samples/pipelines/secure-build.yaml /Users/chipsteen/Code/conveyor/core/loader/testdata/valid/secure-build.yaml`

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd /Users/chipsteen/Code/conveyor && go test ./core/loader/ -run TestConvert -v`
Expected: FAIL — `Convert` not defined

- [ ] **Step 4: Implement Convert**

```go
package loader

import (
	"time"

	"github.com/chip/conveyor/core"
)

// Convert transforms a YAMLPipeline into a core.Pipeline with the given ID.
// The ID is typically derived from the filename.
func Convert(p *YAMLPipeline, id string) (*core.Pipeline, error) {
	// Build name-to-slug map for needs resolution
	nameToSlug := make(map[string]string)
	for _, s := range p.Stages {
		nameToSlug[s.Name] = Slugify(s.Name)
	}

	now := time.Now()
	pipeline := &core.Pipeline{
		ID:          id,
		Name:        p.Name,
		Description: p.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Convert triggers
	for _, t := range p.Triggers {
		pipeline.Triggers = append(pipeline.Triggers, core.Trigger{
			Type:     t.Type,
			Branches: t.Branches,
			Events:   t.Events,
			Paths:    t.Paths,
		})
	}

	// Convert cache
	if p.Cache != nil {
		pipeline.Cache = &core.CacheConfig{
			Key:    p.Cache.Key,
			Paths:  p.Cache.Paths,
			Policy: p.Cache.Policy,
		}
	}

	// Convert environment
	if p.Environment != nil {
		pipeline.Environment = p.Environment.Variables
	}

	// Convert stages
	for _, ys := range p.Stages {
		stageID := Slugify(ys.Name)

		stage := core.Stage{
			ID:   stageID,
			Name: ys.Name,
		}

		// Convert needs to slugified IDs
		for _, need := range ys.Needs {
			if slug, ok := nameToSlug[need]; ok {
				stage.Needs = append(stage.Needs, slug)
			}
		}

		// Convert when
		if ys.When != nil {
			stage.When = &core.ConditionalExecution{
				Branch:  ys.When.Branch,
				Status:  ys.When.Status,
				Custom:  ys.When.Custom,
				Pattern: ys.When.Pattern,
			}
		}

		// Convert steps
		for _, yst := range ys.Steps {
			stepID := Slugify(stageID + "-" + yst.Name)

			step := core.Step{
				ID:          stepID,
				Name:        yst.Name,
				Command:     yst.Run,
				Plugin:      yst.Plugin,
				Image:       yst.Image,
				Environment: yst.Environment,
				Config:      yst.Config,
				Timeout:     yst.Timeout,
				DependsOn:   yst.DependsOn,
				Outputs:     yst.Outputs,
			}

			// Type inference
			if yst.Type != "" {
				step.Type = yst.Type
			} else if yst.Plugin != "" {
				step.Type = "plugin"
			} else {
				step.Type = "script"
			}

			// Description into metadata
			if yst.Description != "" {
				if step.Metadata == nil {
					step.Metadata = make(map[string]interface{})
				}
				step.Metadata["description"] = yst.Description
			}

			// Convert when
			if yst.When != nil {
				step.When = &core.ConditionalExecution{
					Branch:  yst.When.Branch,
					Status:  yst.When.Status,
					Custom:  yst.When.Custom,
					Pattern: yst.When.Pattern,
				}
			}

			// Convert retry
			if yst.Retry != nil {
				step.Retry = &core.RetryConfig{
					MaxAttempts:        yst.Retry.MaxAttempts,
					Interval:           yst.Retry.Interval,
					ExponentialBackoff: yst.Retry.ExponentialBackoff,
				}
			}

			// Convert cache
			if yst.Cache != nil {
				step.Cache = &core.CacheConfig{
					Key:    yst.Cache.Key,
					Paths:  yst.Cache.Paths,
					Policy: yst.Cache.Policy,
				}
			}

			stage.Steps = append(stage.Steps, step)
		}

		pipeline.Stages = append(pipeline.Stages, stage)
	}

	return pipeline, nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /Users/chipsteen/Code/conveyor && go test ./core/loader/ -run TestConvert -v`
Expected: PASS — all 7 test cases

- [ ] **Step 6: Commit**

```bash
git add core/loader/convert.go core/loader/convert_test.go core/loader/testdata/valid/secure-build.yaml
git commit -m "feat(loader): implement Convert function mapping YAML to core types"
```

### Task 7: Implement PipelineLoader with tests (TDD)

**Files:**
- Create: `core/loader/loader.go`
- Create: `core/loader/loader_test.go`

- [ ] **Step 1: Write the failing tests**

```go
package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chip/conveyor/core"
)

func newTestEngine() *core.PipelineEngine {
	return core.NewPipelineEngine()
}

func TestLoadFile_ValidMinimal(t *testing.T) {
	engine := newTestEngine()
	l := NewPipelineLoader(engine, "testdata/valid")

	absPath, _ := filepath.Abs("testdata/valid/minimal.yaml")
	pipeline, warnings, err := l.LoadFile(absPath)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if pipeline.ID != "minimal" {
		t.Errorf("ID = %q, want %q", pipeline.ID, "minimal")
	}

	// Verify it was registered in the engine
	got, err := engine.GetPipeline("minimal")
	if err != nil {
		t.Fatalf("engine.GetPipeline() error = %v", err)
	}
	if got.Name != "minimal-pipeline" {
		t.Errorf("engine pipeline Name = %q, want %q", got.Name, "minimal-pipeline")
	}
}

func TestLoadFile_SecureBuild(t *testing.T) {
	engine := newTestEngine()
	l := NewPipelineLoader(engine, "testdata/valid")

	absPath, _ := filepath.Abs("testdata/valid/secure-build.yaml")
	pipeline, warnings, err := l.LoadFile(absPath)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	// Should warn about version, notifications, artifacts
	if len(warnings) != 3 {
		t.Errorf("len(warnings) = %d, want 3, got %v", len(warnings), warnings)
	}
	if len(pipeline.Stages) != 6 {
		t.Errorf("len(Stages) = %d, want 6", len(pipeline.Stages))
	}
}

func TestLoadFile_InvalidYAML(t *testing.T) {
	engine := newTestEngine()
	l := NewPipelineLoader(engine, "testdata/invalid")

	absPath, _ := filepath.Abs("testdata/invalid/bad-syntax.yaml")
	_, _, err := l.LoadFile(absPath)
	if err == nil {
		t.Fatal("LoadFile() expected error for bad YAML, got nil")
	}
}

func TestLoadFromBytes_UsesProvidedID(t *testing.T) {
	engine := newTestEngine()
	l := NewPipelineLoader(engine, "")

	yamlData := []byte(`
name: Some Pipeline
stages:
  - name: build
    steps:
      - name: compile
        run: go build ./...
`)

	pipeline, _, err := l.LoadFromBytes(yamlData, "custom-id")
	if err != nil {
		t.Fatalf("LoadFromBytes() error = %v", err)
	}
	if pipeline.ID != "custom-id" {
		t.Errorf("ID = %q, want %q", pipeline.ID, "custom-id")
	}
}

func TestLoadDirectory_MixedFiles(t *testing.T) {
	// Create a temp directory with valid and invalid files
	tmpDir, err := os.MkdirTemp("", "loader-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write a valid file
	validYAML := []byte("name: valid\nstages:\n  - name: build\n    steps:\n      - name: step\n        run: echo hi\n")
	os.WriteFile(filepath.Join(tmpDir, "valid.yaml"), validYAML, 0644)

	// Write an invalid file
	invalidYAML := []byte("name: invalid\nstages: []\n")
	os.WriteFile(filepath.Join(tmpDir, "invalid.yaml"), invalidYAML, 0644)

	// Write a non-YAML file (should be ignored)
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("not yaml"), 0644)

	engine := newTestEngine()
	l := NewPipelineLoader(engine, tmpDir)

	result, err := l.LoadDirectory()
	if err != nil {
		t.Fatalf("LoadDirectory() error = %v", err)
	}

	if len(result.Loaded) != 1 {
		t.Errorf("len(Loaded) = %d, want 1", len(result.Loaded))
	}
	if len(result.Errors) != 1 {
		t.Errorf("len(Errors) = %d, want 1", len(result.Errors))
	}
	if _, ok := result.Errors["invalid.yaml"]; !ok {
		t.Errorf("expected error for invalid.yaml, got errors for: %v", result.Errors)
	}
}

func TestLoadDirectory_Nonexistent(t *testing.T) {
	engine := newTestEngine()
	l := NewPipelineLoader(engine, "/nonexistent/path")

	result, err := l.LoadDirectory()
	if err != nil {
		t.Fatalf("LoadDirectory() error = %v", err)
	}
	if len(result.Loaded) != 0 {
		t.Errorf("len(Loaded) = %d, want 0", len(result.Loaded))
	}
}

func TestLoadFromBytes_DuplicatePipelineID(t *testing.T) {
	engine := newTestEngine()
	l := NewPipelineLoader(engine, "")

	yamlData := []byte("name: Dupe\nstages:\n  - name: build\n    steps:\n      - name: step\n        run: echo\n")

	_, _, err := l.LoadFromBytes(yamlData, "dupe-id")
	if err != nil {
		t.Fatalf("first LoadFromBytes() error = %v", err)
	}

	_, _, err = l.LoadFromBytes(yamlData, "dupe-id")
	if err == nil {
		t.Fatal("second LoadFromBytes() expected error for duplicate pipeline ID, got nil")
	}
}

func TestLoadDirectory_DuplicateExtensions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loader-ext-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	yamlData := []byte("name: Same Pipeline\nstages:\n  - name: build\n    steps:\n      - name: step\n        run: echo\n")
	os.WriteFile(filepath.Join(tmpDir, "build.yaml"), yamlData, 0644)
	os.WriteFile(filepath.Join(tmpDir, "build.yml"), yamlData, 0644)

	engine := newTestEngine()
	l := NewPipelineLoader(engine, tmpDir)

	result, err := l.LoadDirectory()
	if err != nil {
		t.Fatalf("LoadDirectory() error = %v", err)
	}

	// First file (build.yaml) should succeed, second (build.yml) should fail as duplicate
	if len(result.Loaded) != 1 {
		t.Errorf("len(Loaded) = %d, want 1", len(result.Loaded))
	}
	if len(result.Errors) != 1 {
		t.Errorf("len(Errors) = %d, want 1", len(result.Errors))
	}
}

func TestLoadFromBytes_NameFromYAML(t *testing.T) {
	engine := newTestEngine()
	l := NewPipelineLoader(engine, "")

	yamlData := []byte("name: My Actual Name\nstages:\n  - name: build\n    steps:\n      - name: step\n        run: echo\n")

	pipeline, _, err := l.LoadFromBytes(yamlData, "custom-id")
	if err != nil {
		t.Fatalf("LoadFromBytes() error = %v", err)
	}
	if pipeline.ID != "custom-id" {
		t.Errorf("ID = %q, want %q", pipeline.ID, "custom-id")
	}
	if pipeline.Name != "My Actual Name" {
		t.Errorf("Name = %q, want %q", pipeline.Name, "My Actual Name")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/chipsteen/Code/conveyor && go test ./core/loader/ -run "TestLoad" -v`
Expected: FAIL — `NewPipelineLoader`, `LoadFile`, etc. not defined

- [ ] **Step 3: Implement PipelineLoader**

```go
package loader

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chip/conveyor/core"
)

// PipelineLoader loads pipeline definitions from YAML files.
type PipelineLoader struct {
	engine       *core.PipelineEngine
	pipelinesDir string
}

// LoadResult contains the results of loading pipelines from a directory.
type LoadResult struct {
	Loaded   []*core.Pipeline
	Warnings map[string][]string // filename -> warnings
	Errors   map[string]error    // filename -> error
}

// NewPipelineLoader creates a new PipelineLoader.
func NewPipelineLoader(engine *core.PipelineEngine, pipelinesDir string) *PipelineLoader {
	return &PipelineLoader{
		engine:       engine,
		pipelinesDir: pipelinesDir,
	}
}

// LoadDirectory scans the configured directory and loads all YAML pipeline files.
// Returns a LoadResult with loaded pipelines, warnings, and errors per file.
// If the directory does not exist, returns an empty result (not an error).
func (l *PipelineLoader) LoadDirectory() (*LoadResult, error) {
	result := &LoadResult{
		Warnings: make(map[string][]string),
		Errors:   make(map[string]error),
	}

	// Check if directory exists
	if _, err := os.Stat(l.pipelinesDir); os.IsNotExist(err) {
		log.Printf("Pipeline directory %q does not exist, skipping", l.pipelinesDir)
		return result, nil
	}

	// Glob for YAML files
	var files []string
	for _, pattern := range []string{"*.yaml", "*.yml"} {
		matches, err := filepath.Glob(filepath.Join(l.pipelinesDir, pattern))
		if err != nil {
			return nil, fmt.Errorf("failed to glob %s: %w", pattern, err)
		}
		files = append(files, matches...)
	}

	// Sort alphabetically for deterministic processing order
	sort.Strings(files)

	for _, path := range files {
		filename := filepath.Base(path)
		pipeline, warnings, err := l.LoadFile(path)
		if err != nil {
			result.Errors[filename] = err
			continue
		}
		if len(warnings) > 0 {
			result.Warnings[filename] = warnings
		}
		result.Loaded = append(result.Loaded, pipeline)
	}

	return result, nil
}

// LoadFile loads a single pipeline from a YAML file.
// The pipeline ID is derived from the filename (without extension).
// Returns the loaded pipeline, any warnings, and an error if loading failed.
func (l *PipelineLoader) LoadFile(path string) (*core.Pipeline, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Derive ID from filename
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	id := strings.TrimSuffix(base, ext)

	return l.loadPipeline(data, id)
}

// LoadFromBytes loads a pipeline from raw YAML bytes with a given ID.
// Returns the loaded pipeline, any warnings, and an error if loading failed.
func (l *PipelineLoader) LoadFromBytes(data []byte, id string) (*core.Pipeline, []string, error) {
	return l.loadPipeline(data, id)
}

// loadPipeline is the shared implementation for LoadFile and LoadFromBytes.
func (l *PipelineLoader) loadPipeline(data []byte, id string) (*core.Pipeline, []string, error) {
	// Step 1: Parse
	yp, err := Parse(data)
	if err != nil {
		return nil, nil, fmt.Errorf("YAML parse error: %w", err)
	}

	// Step 2: Validate
	warnings, err := Validate(yp)
	if err != nil {
		return nil, warnings, err
	}

	// Step 3: Convert
	pipeline, err := Convert(yp, id)
	if err != nil {
		return nil, warnings, fmt.Errorf("conversion error: %w", err)
	}

	// Step 4: Register with engine
	if err := l.engine.CreatePipeline(pipeline); err != nil {
		return nil, warnings, fmt.Errorf("failed to register pipeline: %w", err)
	}

	return pipeline, warnings, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/chipsteen/Code/conveyor && go test ./core/loader/ -run "TestLoad" -v`
Expected: PASS — all 10 test cases

- [ ] **Step 5: Run full test suite**

Run: `cd /Users/chipsteen/Code/conveyor && go test ./core/loader/ -v`
Expected: PASS — all tests in the package

- [ ] **Step 6: Commit**

```bash
git add core/loader/loader.go core/loader/loader_test.go
git commit -m "feat(loader): implement PipelineLoader with directory scanning and file loading"
```

## Chunk 4: Integration — API Route and main.go

### Task 8: Add import API endpoint

**Files:**
- Modify: `api/routes/pipeline.go` (add import handler)
- Modify: `api/routes.go` (register import route — note: the `SetupRoutes` function in `api/routes.go` is what `main.go` calls)

- [ ] **Step 1: Add ImportPipeline handler to pipeline.go**

Add the following function at the end of `api/routes/pipeline.go`:

```go
// RegisterPipelineImportRoute registers the pipeline import route.
// This is separate because it needs a *loader.PipelineLoader dependency.
func RegisterPipelineImportRoute(router *gin.RouterGroup, pipelineLoader interface{ LoadFromBytes([]byte, string) (*core.Pipeline, []string, error) }) {
	router.POST("/import", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'name' is required"})
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
			return
		}

		pipeline, warnings, err := pipelineLoader.LoadFromBytes(body, name)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "warnings": warnings})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"pipeline": pipeline,
			"warnings": warnings,
		})
	})
}
```

Add `"io"` to the imports in `pipeline.go`.

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/chipsteen/Code/conveyor && go build ./api/...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add api/routes/pipeline.go
git commit -m "feat(api): add POST /api/pipelines/import endpoint"
```

### Task 9: Integrate loader into main.go

**Files:**
- Modify: `cli/main.go`
- Modify: `api/routes.go`

- [ ] **Step 1: Update api/routes.go to accept and wire loader**

**Note:** This changes the `SetupRoutes` signature from 2 to 3 parameters. The only caller is `cli/main.go`, updated in step 2. No existing tests call `SetupRoutes` (there are currently zero test files in the project).

In `api/routes.go`, change the `SetupRoutes` function signature to:

```go
func SetupRoutes(r *gin.Engine, engine *core.PipelineEngine, pipelineLoader interface{ LoadFromBytes([]byte, string) (*core.Pipeline, []string, error) })
```

Add after the pipeline routes registration:

```go
	// Pipeline import route (needs loader)
	if pipelineLoader != nil {
		routes.RegisterPipelineImportRoute(pipelineRoutes, pipelineLoader)
	}
```

- [ ] **Step 2: Update main.go to use loader instead of sample data**

Replace the imports, `createSampleData` call, `simulateJobProgress` call, and remove the `createSampleData` and `simulateJobProgress` functions. Replace with:

```go
import (
	// ... existing imports ...
	"github.com/chip/conveyor/core/loader"
)
```

In `main()`, after plugin registration and before route setup:

```go
	// Load pipelines from YAML directory
	pipelineLoader := loader.NewPipelineLoader(engine, "pipelines")
	result, err := pipelineLoader.LoadDirectory()
	if err != nil {
		log.Fatalf("Failed to scan pipeline directory: %v", err)
	}
	for file, warnings := range result.Warnings {
		for _, w := range warnings {
			log.Printf("WARN [%s]: %s", file, w)
		}
	}
	for file, loadErr := range result.Errors {
		log.Printf("ERROR [%s]: %s", file, loadErr)
	}
	log.Printf("Loaded %d pipelines from YAML", len(result.Loaded))
```

Update the `api.SetupRoutes` call:

```go
	api.SetupRoutes(router, engine, pipelineLoader)
```

Remove the `createSampleData(engine)` call, the `go simulateJobProgress(engine)` call, and the `createSampleData`, `createSampleJobs`, and `simulateJobProgress` functions entirely.

- [ ] **Step 3: Copy secure-build.yaml to pipelines/ directory**

Run:
```bash
mkdir -p /Users/chipsteen/Code/conveyor/pipelines
cp /Users/chipsteen/Code/conveyor/samples/pipelines/secure-build.yaml /Users/chipsteen/Code/conveyor/pipelines/
```

- [ ] **Step 4: Verify it compiles**

Run: `cd /Users/chipsteen/Code/conveyor && go build ./...`
Expected: No errors

- [ ] **Step 5: Run all tests**

Run: `cd /Users/chipsteen/Code/conveyor && go test ./... -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add cli/main.go api/routes.go pipelines/secure-build.yaml
git commit -m "feat: integrate YAML pipeline loader into startup and API

Replace hardcoded sample data with YAML-based pipeline loading.
Pipelines are now loaded from the pipelines/ directory at startup.
The POST /api/pipelines/import endpoint allows ad-hoc YAML imports."
```

### Task 10: Verify end-to-end

- [ ] **Step 1: Start the server**

Run: `cd /Users/chipsteen/Code/conveyor && go run ./cli/main.go`
Expected: Log output showing "Loaded 1 pipelines from YAML" and warnings about unsupported fields

- [ ] **Step 2: Test the pipelines API**

Run: `curl -s http://localhost:8080/api/pipelines | python3 -m json.tool`
Expected: JSON array containing the secure-build pipeline with 6 stages

- [ ] **Step 3: Test the import API**

Run:
```bash
curl -s -X POST "http://localhost:8080/api/pipelines/import?name=test-import" \
  -H "Content-Type: text/yaml" \
  -d '
name: Test Import
stages:
  - name: build
    steps:
      - name: compile
        run: echo hello
' | python3 -m json.tool
```
Expected: 201 response with the imported pipeline

- [ ] **Step 4: Verify the imported pipeline appears in the list**

Run: `curl -s http://localhost:8080/api/pipelines | python3 -m json.tool`
Expected: JSON array with 2 pipelines (secure-build + test-import)

- [ ] **Step 5: Stop the server (Ctrl+C) and make final commit if any adjustments needed**
