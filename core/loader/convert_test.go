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
