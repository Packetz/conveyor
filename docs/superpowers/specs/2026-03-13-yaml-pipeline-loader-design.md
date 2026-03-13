# YAML Pipeline Loader Design

**Date:** 2026-03-13
**Issue:** https://github.com/Packetz/conveyor/issues/2
**Status:** Approved

## Overview

Add the ability to load pipeline definitions from YAML files, parse them into internal `core.Pipeline` structs, validate them, and register them with the pipeline engine. Pipelines load from a configurable directory at startup and can be imported ad-hoc via an API endpoint.

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| `run` vs `command` in YAML | YAML uses `run`, maps to `Step.Command` internally | Matches CI/CD conventions (GitHub Actions, GitLab CI) without changing existing API |
| Unsupported YAML fields | Warn on `notifications`, `artifacts`, `version`; don't fail | Transparent about what's supported; avoids dead code for unimplemented features |
| Loading source | Directory scan at startup + `POST /api/pipelines/import` | "Pipelines as code" from directory; API for testing and frontend |
| Pipeline ID derivation | From filename (e.g. `secure-build.yaml` -> `secure-build`) | Intuitive convention; avoids name/ID conflicts |
| Architecture | Separate YAML structs with explicit mapping to core types | Clean separation; YAML schema evolves independently of internal/API types |

## Package Structure

New package: `core/loader/`

```
core/loader/
  types.go       # YAML-specific structs with yaml tags
  loader.go      # PipelineLoader: directory scanning, file loading, conversion
  validator.go   # Validation logic: required fields, needs refs, cycle detection
  loader_test.go # Tests for loading and conversion
  validator_test.go # Tests for each validation rule
  testdata/
    valid/
      secure-build.yaml    # Copy of samples/pipelines/secure-build.yaml
      minimal.yaml         # Minimal valid pipeline
    invalid/
      missing-name.yaml
      empty-stages.yaml
      circular-deps.yaml
      no-run-or-plugin.yaml
      bad-syntax.yaml
```

## YAML Types

```go
// YAMLPipeline is the top-level YAML representation
type YAMLPipeline struct {
    Name        string            `yaml:"name"`
    Description string            `yaml:"description"`
    Version     string            `yaml:"version"`     // warned as unsupported
    Triggers    []YAMLTrigger     `yaml:"triggers"`
    Environment *YAMLEnvironment  `yaml:"environment"`
    Cache       *YAMLCache        `yaml:"cache"`
    Stages      []YAMLStage       `yaml:"stages"`
    Notifications interface{}     `yaml:"notifications"` // warned as unsupported
    Artifacts     interface{}     `yaml:"artifacts"`     // warned as unsupported
}

// YAMLEnvironment holds environment variable config
type YAMLEnvironment struct {
    Variables map[string]string `yaml:"variables"`
}

// YAMLTrigger represents a pipeline trigger in YAML
type YAMLTrigger struct {
    Type     string   `yaml:"type"`
    Branches []string `yaml:"branches"`
    Events   []string `yaml:"events"`
    Paths    []string `yaml:"paths"`
}

// YAMLCache represents cache config in YAML
type YAMLCache struct {
    Key    string   `yaml:"key"`
    Paths  []string `yaml:"paths"`
    Policy string   `yaml:"policy"`
}

// YAMLStage represents a stage in YAML
type YAMLStage struct {
    Name  string     `yaml:"name"`
    Needs []string   `yaml:"needs"`
    When  *YAMLWhen  `yaml:"when"`
    Steps []YAMLStep `yaml:"steps"`
}

// YAMLStep represents a step in YAML
type YAMLStep struct {
    Name        string                 `yaml:"name"`
    Description string                 `yaml:"description"`
    Type        string                 `yaml:"type"`    // optional; inferred if omitted
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

// YAMLWhen represents conditional execution
type YAMLWhen struct {
    Branch  string `yaml:"branch"`
    Status  string `yaml:"status"`
    Custom  string `yaml:"custom"`
    Pattern string `yaml:"pattern"`
}

// YAMLRetry represents retry config
type YAMLRetry struct {
    MaxAttempts        int    `yaml:"max_attempts"`
    Interval           string `yaml:"interval"`
    ExponentialBackoff bool   `yaml:"exponential_backoff"`
}
```

## Field Mapping Clarifications

### Step.Type inference

If a YAML step has an explicit `type` field, use it. Otherwise:
- If `plugin` is set -> `Type = "plugin"`
- If `run` is set -> `Type = "script"`

### Step.Description

`core.Step` has no `Description` field. Map `YAMLStep.Description` into `Step.Metadata["description"]` so it's preserved but doesn't require a struct change.

### Stage.Parallel and Stage.DependsOn

These `core.Stage` fields are not represented in the YAML schema. They are left as zero values (`false` and `nil`). The `needs` field serves as the dependency mechanism in YAML. `Parallel` and `DependsOn` are reserved for future use by the job scheduler (roadmap issue #5).

### needs resolution

YAML `needs` references stage **names** as written in the YAML. During conversion, these are resolved to slugified stage IDs (since stage IDs = slugified names). The validator checks `needs` against the set of stage names *before* slugification to match the YAML author's intent. After conversion, `core.Stage.Needs` contains the slugified IDs.

Example: stage name `"Security Checks"` -> stage ID `"security-checks"`. A `needs: [Security Checks]` reference is valid and converts to `needs: ["security-checks"]`.

### Cache.Policy

The `Policy` field on `CacheConfig` is free-form and not validated by the loader. It is passed through as-is. Validation of policy values (e.g. `pull`, `push`, `pull-push`) is deferred to the cache manager when it is implemented.

### Slugification

Slugification uses a simple algorithm: lowercase, replace spaces and underscores with hyphens, strip any character that is not `[a-z0-9-]`, collapse consecutive hyphens, and trim leading/trailing hyphens. If two stages produce the same slugified ID, it is a validation error (duplicate stage ID).

### Duplicate file extensions

When `LoadDirectory` encounters both `build.yaml` and `build.yml`, all matching files are sorted together alphabetically by full filename. The first file processed wins (so `build.yaml` beats `build.yml` since `.yaml` < `.yml` lexicographically). The second file produces a duplicate ID hard error and is skipped. Unreadable files (e.g. permission denied) are recorded in `LoadResult.Errors` and do not cause `LoadDirectory` to return a top-level error.

### Validate-only usage

`LoadFile` and `LoadFromBytes` always call `engine.CreatePipeline` as the final step. For dry-run validation (e.g. a future `POST /api/pipelines/validate` endpoint), callers can use the exported `Parse`, `Validate`, and `Convert` functions independently without calling the full `Load*` methods.

### go.mod

`gopkg.in/yaml.v3` is not currently in `go.mod` and must be added.

### Sample YAML behavioral note

The existing `secure-build.yaml` does not declare `needs` on the `security-checks` stage, so `pre-build` and `security-checks` will have no ordering dependency when loaded from YAML. This differs from the hardcoded sample data ordering in `main.go`. This is acceptable — the YAML is the source of truth. If ordering is desired, the YAML should be updated to add `needs: [pre-build]` to the `security-checks` stage.

## Data Flow

### File Loading

```
YAML file on disk (or API request body)
  |
  v
PipelineLoader.LoadFile(path)
  |
  +-- 1. Read file bytes
  +-- 2. yaml.Unmarshal -> YAMLPipeline
  +-- 3. Validate(yamlPipeline) -> []ValidationError or nil
  +-- 4. CheckUnsupportedFields(yamlPipeline) -> []Warning
  +-- 5. Convert(yamlPipeline, filenameWithoutExt) -> *core.Pipeline
  |      +-- ID = filename without extension
  |      +-- Stage IDs = slugified stage name
  |      +-- Step IDs = "stageid-stepname" slugified
  |      +-- run -> Step.Command
  |      +-- Step.Type: explicit if set, else "plugin" if plugin set, else "script"
  |      +-- Step.Description stored in Step.Metadata["description"]
  |      +-- Stage.Needs -> Stage.Needs (values are slugified stage names)
  |      +-- Stage.Parallel and Stage.DependsOn left as zero values (not in YAML schema)
  |      +-- environment.variables -> Pipeline.Environment
  +-- 6. engine.CreatePipeline(pipeline)
```

### Directory Scanning (startup)

```
main.go calls loader.LoadDirectory("pipelines/")
  +-- Glob for *.yaml, *.yml
  +-- LoadFile() for each
  +-- Collect warnings and errors per file
  +-- Log summary: "Loaded 3 pipelines, 1 failed, 2 warnings"
  +-- Return LoadResult with successes, failures, warnings
```

### API Import

```
POST /api/pipelines/import?name=my-pipeline
  Body: raw YAML
  |
  +-- loader.LoadFromBytes(yamlBytes, nameParam) -> pipeline, warnings, error
  +-- Return 201 with pipeline + warnings, or 400 with validation errors
```

## PipelineLoader API

```go
type PipelineLoader struct {
    engine       *core.PipelineEngine
    pipelinesDir string
}

type LoadResult struct {
    Loaded   []*core.Pipeline
    Warnings map[string][]string  // filename -> warnings
    Errors   map[string]error     // filename -> error
}

func NewPipelineLoader(engine *core.PipelineEngine, pipelinesDir string) *PipelineLoader

// LoadDirectory scans a directory and loads all YAML pipeline files
func (l *PipelineLoader) LoadDirectory() (*LoadResult, error)

// LoadFile loads a single pipeline from a YAML file
func (l *PipelineLoader) LoadFile(path string) (*core.Pipeline, []string, error)

// LoadFromBytes loads a pipeline from raw YAML bytes with a given ID
func (l *PipelineLoader) LoadFromBytes(data []byte, id string) (*core.Pipeline, []string, error)

// --- Exported standalone functions for dry-run / validate-only usage ---

// Parse unmarshals YAML bytes into a YAMLPipeline struct
func Parse(data []byte) (*YAMLPipeline, error)

// Validate checks a YAMLPipeline for errors and returns warnings
func Validate(p *YAMLPipeline) (warnings []string, err error)

// Convert transforms a YAMLPipeline into a core.Pipeline with the given ID
func Convert(p *YAMLPipeline, id string) (*core.Pipeline, error)

// Slugify converts a name to a URL-safe ID
func Slugify(name string) string
```

## Validation Rules

### Hard Errors (pipeline not loaded)

| Rule | Description |
|------|-------------|
| Invalid YAML syntax | `yaml.Unmarshal` fails |
| Missing `name` | Pipeline must have a name |
| No stages | Pipeline must have at least one stage |
| Empty stage | Each stage must have at least one step |
| Missing stage name | Each stage must have a name |
| Missing step name | Each step must have a name |
| No run or plugin | Each step must have either `run` or `plugin` (not both, not neither) |
| Invalid `needs` reference | `needs` must reference existing stage names |
| Circular dependencies | `needs` graph must be a DAG (no cycles) |
| Duplicate pipeline ID | Cannot load a pipeline with an ID already in the engine |

### Warnings (pipeline still loads)

| Field | Warning message |
|-------|----------------|
| `version` | "Field 'version' is not yet supported and will be ignored" |
| `notifications` | "Field 'notifications' is not yet supported and will be ignored" |
| `artifacts` | "Field 'artifacts' is not yet supported and will be ignored" |

## Integration Points

### main.go Changes

Replace `createSampleData(engine)` and `simulateJobProgress(engine)` with:

```go
loader := loader.NewPipelineLoader(engine, "pipelines/")
result, err := loader.LoadDirectory()
if err != nil {
    log.Fatalf("Failed to scan pipeline directory: %v", err)
}
for file, warnings := range result.Warnings {
    for _, w := range warnings {
        log.Printf("WARN [%s]: %s", file, w)
    }
}
for file, err := range result.Errors {
    log.Printf("ERROR [%s]: %s", file, err)
}
log.Printf("Loaded %d pipelines", len(result.Loaded))
```

The `samples/pipelines/secure-build.yaml` moves (or is symlinked) to `pipelines/` so it loads at startup.

### New API Route

Add to `api/routes/`:

```
POST /api/pipelines/import?name=<id>  ->  routes.ImportPipeline
```

Handler validates the `name` query param, reads the body, calls `loader.LoadFromBytes()`, and returns the pipeline with any warnings.

### go.mod

Add dependency: `gopkg.in/yaml.v3`

## Testing Strategy

### loader_test.go

- Load `testdata/valid/secure-build.yaml` — verify all fields map correctly
- Load `testdata/valid/minimal.yaml` — verify minimal pipeline works
- Load `testdata/invalid/*` — verify each returns the expected error
- Load valid YAML with unsupported fields — verify warnings are returned
- `LoadDirectory` with mixed valid/invalid files — verify partial success
- `LoadFromBytes` — verify ID comes from parameter, not YAML content

### validator_test.go

- Test each validation rule independently
- Test circular dependency detection with 2-node and 3-node cycles
- Test valid `needs` graph with diamond dependencies
- Test step with both `run` and `plugin` — error
- Test step with neither `run` nor `plugin` — error
